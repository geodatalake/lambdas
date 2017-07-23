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

func mineAllVersions(bucket, path string, svc *s3.S3) ([]*s3.ObjectVersion, error) {
	input := &s3.ListObjectVersionsInput{Bucket: aws.String(bucket), Prefix: aws.String(path), MaxKeys: aws.Int64(1000)}
	if path == "/" {
		input = &s3.ListObjectVersionsInput{Bucket: aws.String(bucket), MaxKeys: aws.Int64(1000)}
	}
	log.Println("Attempting to read all versions", aws.StringValue(input.Bucket), aws.StringValue(input.Prefix))
	output, err := svc.ListObjectVersions(input)
	if err != nil {
		log.Println("Error reading bucket", err)

		return nil, err
	}
	retval := make([]*s3.ObjectVersion, len(output.Versions))
	copy(retval, output.Versions)
	if aws.BoolValue(output.IsTruncated) == true {
		for {
			input.KeyMarker = output.NextKeyMarker
			input.VersionIdMarker = output.NextVersionIdMarker
			output, err = svc.ListObjectVersions(input)
			if err != nil {
				return nil, err
			}
			fmt.Print("\r", len(retval), "           ")
			newslice := make([]*s3.ObjectVersion, len(retval)+len(output.Versions))
			copy(newslice, retval)
			l := len(retval)
			for i, o := range output.Versions {
				newslice[l+i] = o
			}
			retval = newslice
			if !aws.BoolValue(output.IsTruncated) {
				fmt.Println("complete")
				break
			}
		}
	} else {
		return output.Versions, nil
	}
	return retval, nil
}

func mineAllObjects(bucket, path string, svc *s3.S3, versioned bool) ([]*s3.Object, []*s3.ObjectVersion, error) {
	input := &s3.ListObjectsInput{Bucket: aws.String(bucket), Prefix: aws.String(path)}
	if path == "/" {
		input = &s3.ListObjectsInput{Bucket: aws.String(bucket)}
	}
	if versioned {
		log.Println("including versions")
		versions, err := mineAllVersions(bucket, path, svc)
		return []*s3.Object{}, versions, err
	}

	log.Println("Attempting to read", aws.StringValue(input.Bucket), aws.StringValue(input.Prefix))
	output, err := svc.ListObjects(input)
	if err != nil {
		log.Println("Error reading bucket", err)

		return nil, nil, err
	}
	retval := make([]*s3.Object, len(output.Contents))
	copy(retval, output.Contents)
	if aws.BoolValue(output.IsTruncated) == true {
		for {
			input.Marker = output.Contents[len(output.Contents)-1].Key
			output, err = svc.ListObjects(input)
			if err != nil {
				return nil, nil, err
			}
			fmt.Print("\r", len(retval), "  ")
			newslice := make([]*s3.Object, len(retval)+len(output.Contents))
			copy(newslice, retval)
			l := len(retval)
			for i, o := range output.Contents {
				newslice[l+i] = o
			}
			retval = newslice
			if !aws.BoolValue(output.IsTruncated) {
				fmt.Println("complete")
				break
			}
		}
	} else {
		return output.Contents, []*s3.ObjectVersion{}, nil
	}
	return retval, []*s3.ObjectVersion{}, nil
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
	Version      bool   `json:"version"`
}

func NewBucketFile(region, bucket, key, lastModified string, size int64) *BucketFile {
	return &BucketFile{
		Region:       region,
		Bucket:       bucket,
		Key:          key,
		LastModified: lastModified,
		Size:         size,
		Version:      false}
}

func NewBucketFileVersion(region, bucket, key, lastModified string, size int64) *BucketFile {
	return &BucketFile{
		Region:       region,
		Bucket:       bucket,
		Key:          key,
		LastModified: lastModified,
		Size:         size,
		Version:      true}
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

func isVersioned(bucket string, svc *s3.S3) bool {
	versioning, err := svc.GetBucketVersioning(&s3.GetBucketVersioningInput{Bucket: aws.String(bucket)})
	if err == nil {
		versionStatus := aws.StringValue(versioning.Status)
		if versionStatus == "Enabled" {
			return true
		}
	}
	return false
}

func readBucket(region, dir string, svc *s3.S3, wantVersions bool, filterBefore time.Time) ([]*BucketFile, error) {
	bucket, prefix := ParsePath(dir)
	versioned := isVersioned(bucket, svc)
	if versioned {
		log.Println("bucket is versioned")
	}
	objects, versions, err := mineAllObjects(bucket, prefix, svc, versioned && wantVersions)
	if err != nil {
		return []*BucketFile{}, err
	}

	retval := make([]*BucketFile, 0, len(objects)+len(versions))
	for _, object := range objects {
		if !strings.HasSuffix(aws.StringValue(object.Key), "/") && object.LastModified.Before(filterBefore) {
			dateTime := fmt.Sprintf("%s", object.LastModified.UTC().Format(ISO8601FORMAT))
			retval = append(retval, NewBucketFile(region, bucket, aws.StringValue(object.Key), dateTime, *object.Size))
		}
	}
	if versioned && wantVersions {
		for _, v := range versions {
			if !strings.HasSuffix(aws.StringValue(v.Key), "/") && v.LastModified.Before(filterBefore) {
				dateTime := fmt.Sprintf("%s", v.LastModified.UTC().Format(ISO8601FORMAT))
				retval = append(retval, NewBucketFileVersion(region, bucket, aws.StringValue(v.Key), dateTime, aws.Int64Value(v.Size)))
			}
		}
	}
	log.Println("Read", len(retval), "BucketFiles")
	return retval, nil
}

func ReadBucketDir(region, bucket, prefix string, svc *s3.S3) ([]*BucketFile, error) {
	versioned := isVersioned(bucket, svc)
	objects, versions, err := mineAllObjects(bucket, prefix, svc, versioned)
	if err != nil {
		return []*BucketFile{}, err
	}

	retval := make([]*BucketFile, 0, len(objects)+len(versions))
	for _, object := range objects {
		if !strings.HasSuffix(aws.StringValue(object.Key), "/") {
			dateTime := fmt.Sprintf("%s", object.LastModified.UTC().Format(ISO8601FORMAT))
			retval = append(retval, NewBucketFile(region, bucket, aws.StringValue(object.Key), dateTime, *object.Size))
		}
	}
	if versioned {
		for _, v := range versions {
			if !strings.HasSuffix(aws.StringValue(v.Key), "/") {
				dateTime := fmt.Sprintf("%s", v.LastModified.UTC().Format(ISO8601FORMAT))
				retval = append(retval, NewBucketFileVersion(region, bucket, aws.StringValue(v.Key), dateTime, aws.Int64Value(v.Size)))
			}
		}
	}
	log.Println("Read", len(retval), "BucketFiles")
	return retval, nil
}

func ListBucketStructure(region, bucket string, svc *s3.S3, wantVersions bool, filterBefore time.Time) (*DirItem, error) {
	if !strings.HasSuffix(bucket, "/") {
		bucket = bucket + "/"
	}
	bfs, err := readBucket(region, bucket, svc, wantVersions, filterBefore)
	if err != nil {
		return nil, err
	}
	root := NewDirItem("/")
	for _, bf := range bfs {
		currentDir := root
		paths, _ := bf.Paths()
		for _, p := range paths {
			if p != "" {
				if nd, ok := currentDir.Contains(p); !ok {
					newDir := NewDirItem(p)
					currentDir.Add(newDir)
					currentDir = newDir
				} else {
					currentDir = nd
				}
			}
		}
		if !bf.Version || wantVersions {
			currentDir.AddKey(bf)
		}
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
	if item.Size != 0 {
		di.Keys = append(di.Keys, item)
		di.Size += item.Size
	}
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
