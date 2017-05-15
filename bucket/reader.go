// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bucket

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	ISO8601FORMAT = "20060102T150405Z"
)

func mineAllObjects(region, bucket, path string) ([]*s3.Object, error) {
	svc := s3.New(session.New(), &aws.Config{Region: aws.String(region)})
	input := &s3.ListObjectsInput{Bucket: aws.String(bucket), Prefix: aws.String(path)}
	if path == "/" {
		input = &s3.ListObjectsInput{Bucket: aws.String(bucket)}
	}
	output, err := svc.ListObjects(input)
	if err != nil {
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
	if strings.HasPrefix(dir, "s3://") {
		if strings.HasSuffix(dir, "/") {
			parts = strings.Split(dir[5:len(dir)-1], "/")
		} else {
			parts = strings.Split(dir[5:], "/")
		}
	} else {
		if strings.HasSuffix(dir, "/") {
			parts = strings.Split(dir[:len(dir)-1], "/")
		} else {
			parts = strings.Split(dir, "/")
		}
	}
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

func (bf *BucketFile) Paths() ([]string, string) {
	dir, file := path.Split(bf.Key)
	dirs := strings.Split(dir, "/")
	return dirs, file
}

func readBucket(region, dir string) ([]*BucketFile, error) {
	bucket, prefix := ParsePath(dir)
	objects, err := mineAllObjects(region, bucket, prefix)
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
	return retval, nil
}

func ListBucketStructure(region, bucket string) (*DirItem, error) {
	if !strings.HasSuffix(bucket, "/") {
		bucket = bucket + "/"
	}
	bfs, err := readBucket(region, bucket)
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
	for {
		v, good := <-bi.ch
		if good {
			return v, true
		} else {
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
func (bf *BucketFile) Stream() *S3FileReader {
	svc := s3.New(session.New(), &aws.Config{Region: aws.String(bf.Region)})
	return &S3FileReader{
		sectionReader: io.NewSectionReader(NewChunkReader(bf, svc), 0, bf.Size)}
}
