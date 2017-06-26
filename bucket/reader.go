// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bucket

import (
	"fmt"
	"io"
	"log"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	ISO8601FORMAT = "2006-01-02T15:04:05Z"
)

func mineAllObjects(bucket, path string, svc *s3.S3) ([]*s3.Object, error) {
	input := &s3.ListObjectsInput{Bucket: aws.String(bucket), Prefix: aws.String(path)}
	if path == "/" {
		input = &s3.ListObjectsInput{Bucket: aws.String(bucket)}
	}
	log.Println("Attempting to read", aws.StringValue(input.Bucket), aws.StringValue(input.Prefix))
	output, err := svc.ListObjects(input)
	if err != nil {
		log.Println("Error reading bucket", err)

		return nil, err
	}
	retval := make([]*s3.Object, int(*output.MaxKeys))
	n := copy(retval, output.Contents)
	fmt.Printf("Copied %d objects, truncated=%v, maxkeys=%d\n", n, *output.IsTruncated, *output.MaxKeys)
	if *output.IsTruncated == true {
		for {
			input.Marker = aws.String(aws.StringValue(output.Contents[len(output.Contents)-1].Key))
			output, err = svc.ListObjects(input)
			if err != nil {
				return nil, err
			}
			fmt.Printf("Copied %d objects, truncated=%v\n", len(output.Contents), *output.IsTruncated)
			newslice := make([]*s3.Object, len(retval)+len(output.Contents))
			copy(newslice, retval)
			l := len(retval)
			for i, o := range output.Contents {
				newslice[l+i] = o
			}
			retval = newslice
			if !*output.IsTruncated {
				break
			}
		}
	} else {
		return output.Contents, nil
	}
	return retval, nil
}

func ParsePath(dir string) (bucket, prefix string) {
	var parts []string
	start := 0
	if strings.HasPrefix(dir, "s3://") {
		start = 5
	}
	end := len(dir)
	if strings.HasSuffix(dir, "/") {
		end = end - 1
	}
	parts = strings.Split(dir[start:end], "/")
	bucket = parts[0]
	if len(parts) > 1 {
		prefix = strings.Join(parts[1:], "/")
	} else {
		prefix = "/"
	}
	return
}

type BucketFile struct {
	Bucket       string `json:"bucket"`
	Key          string `json:"key"`
	LastModified string `json:"lastModified"`
	Region       string `json:"region"`
	Size         int64  `json:"size"`
}

func NewBucketFile(region, bucket, key, lastModified string, size int64) *BucketFile {
	return &BucketFile{
		Region:       region,
		Bucket:       bucket,
		Key:          key,
		LastModified: lastModified,
		Size:         size}
}

func (bf *BucketFile) Unmarshal(m map[string]interface{}) error {
	if r, ok := m["region"]; ok {
		bf.Region = r.(string)
	}
	if b, ok := m["bucket"]; ok {
		bf.Bucket = b.(string)
	}
	if k, ok := m["key"]; ok {
		bf.Key = k.(string)
	}
	if s, ok := m["size"]; ok {
		bf.Size = int64(s.(float64))
	}
	if lm, ok := m["lastModified"]; ok {
		bf.LastModified = lm.(string)
	}
	return nil
}

func (bf *BucketFile) Paths() ([]string, string) {
	dir, file := path.Split(bf.Key)
	dirs := strings.Split(dir, "/")
	return dirs, file
}

func readBucket(region, dir string, svc *s3.S3) ([]*BucketFile, error) {
	bucket, prefix := ParsePath(dir)
	objects, err := mineAllObjects(bucket, prefix, svc)
	if err != nil {
		return []*BucketFile{}, err
	}

	retval := make([]*BucketFile, 0, len(objects))
	dirName := fmt.Sprintf("%s/", prefix)
	for _, object := range objects {
		filename := aws.StringValue(object.Key)
		// The directory will match, so filter it out
		if filename != dirName {
			dateTime := fmt.Sprintf("%s", object.LastModified.UTC().Format(ISO8601FORMAT))
			retval = append(retval, NewBucketFile(region, bucket, aws.StringValue(object.Key), dateTime, *object.Size))
		}
	}
	log.Println("Read", len(retval), "BucketFiles")
	return retval, nil
}

func ReadBucketDir(region, bucket, prefix string, svc *s3.S3) ([]*BucketFile, error) {
	objects, err := mineAllObjects(bucket, prefix, svc)
	if err != nil {
		return []*BucketFile{}, err
	}

	retval := make([]*BucketFile, 0, len(objects))
	dirName := fmt.Sprintf("%s/", prefix)
	for _, object := range objects {
		filename := aws.StringValue(object.Key)
		// The directory will match, so filter it out
		if filename != dirName {
			dateTime := fmt.Sprintf("%s", object.LastModified.UTC().Format(ISO8601FORMAT))
			retval = append(retval, NewBucketFile(region, bucket, aws.StringValue(object.Key), dateTime, *object.Size))
		}
	}
	log.Println("Read", len(retval), "BucketFiles")
	return retval, nil
}

func ListBucketStructure(region, bucket string, svc *s3.S3) (*DirItem, error) {
	if !strings.HasSuffix(bucket, "/") {
		bucket = bucket + "/"
	}
	bfs, err := readBucket(region, bucket, svc)
	if err != nil {
		return nil, err
	}
	root := NewDirItem("/")
	for _, bf := range bfs {
		rootDir := root
		paths, _ := bf.Paths()
		for _, p := range paths {
			if p != "" {
				if nd, ok := rootDir.Contains(p); !ok {
					newDir := NewDirItem(p)
					rootDir.Add(newDir)
					rootDir = newDir
				} else {
					rootDir = nd
				}
			}
		}
		rootDir.AddKey(bf)
	}
	return root, nil
}

type DirItem struct {
	Name     string
	Children map[string]*DirItem
	Keys     []*BucketFile
	Size     int64
}

func (di *DirItem) GetKeys() []*BucketFile {
	return di.Keys
}

func NewDirItem(name string) *DirItem {
	return &DirItem{
		Name:     name,
		Children: make(map[string]*DirItem),
		Keys:     make([]*BucketFile, 0, 32),
		Size:     0,
	}
}

func (di *DirItem) Contains(name string) (*DirItem, bool) {
	v, ok := di.Children[name]
	return v, ok
}

func (di *DirItem) Add(item *DirItem) *DirItem {
	if nd, ok := di.Contains(item.Name); !ok {
		di.Children[item.Name] = item
		return item
	} else {
		return nd
	}
}

func (di *DirItem) AddKey(item *BucketFile) {
	di.Keys = append(di.Keys, item)
	di.Size += item.Size
}

type DirInfo struct {
	Level int
	Keys  int
	Name  string
	Size  int64
}

func (di *DirItem) Print(lvl int, ch chan *DirInfo) {
	ch <- &DirInfo{Level: lvl, Keys: len(di.Keys), Name: di.Name, Size: di.Size}
	for _, v := range di.Children {
		v.Print(lvl+1, ch)
	}
}

func (di *DirItem) Iterate() *DirIterator {
	return newDirIterator(di)
}

func (di *DirItem) mine(ch chan *DirItem) {
	ch <- di
	for _, v := range di.Children {
		v.mine(ch)
	}
}

type DirIterator struct {
	root *DirItem
	ch   chan *DirItem
}

func newDirIterator(di *DirItem) *DirIterator {
	bi := new(DirIterator)
	bi.ch = make(chan *DirItem)
	bi.root = di
	go bi.traverse()
	return bi
}

func (bi *DirIterator) traverse() {
	bi.root.mine(bi.ch)
	close(bi.ch)
}

func (bi *DirIterator) NextWithKeys() (*DirItem, bool) {
	for {
		v, ok := bi.Next()
		if !ok {
			return nil, false
		}
		if len(v.Keys) > 0 {
			return v, true
		}
	}
}

func (bi *DirIterator) Next() (*DirItem, bool) {
	timeout := time.After(10 * time.Second)
	for {
		select {
		case v, good := <-bi.ch:
			if good {
				return v, true
			} else {
				log.Println("Channel is closed")
				return nil, false
			}
		case <-timeout:
			log.Println("Timeout reading next DirItem")
			return nil, false
		}
	}
}

// Drain the channel
func (bi *DirIterator) Abort() {
	for {
		_, ok := bi.Next()
		if !ok {
			return
		}
	}
}

type S3FileReader struct {
	sectionReader *io.SectionReader
}

func (sfr *S3FileReader) Read(p []byte) (n int, err error) {
	return sfr.sectionReader.Read(p)
}

func (sfr *S3FileReader) ReadAt(p []byte, off int64) (n int, err error) {
	return sfr.sectionReader.ReadAt(p, off)
}

func (sfr *S3FileReader) Size() int64 {
	return sfr.sectionReader.Size()
}

// S3FileReader is completely reusable
// This keeps requests down while allowing multiple readers
// to access it for determining types
func (bf *BucketFile) Stream(sess *session.Session) *S3FileReader {
	svc := s3.New(sess, &aws.Config{Region: aws.String(bf.Region)})
	return &S3FileReader{
		sectionReader: io.NewSectionReader(NewChunkReader(bf.Size, NewS3Reader(bf, svc)), 0, bf.Size)}
}
