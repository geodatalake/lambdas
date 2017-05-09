// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bucket

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	ISO8601FORMAT = "20060102T150405Z"
	VERSION       = "0.1"
)

func mineAllObjects(region, bucket, path string) ([]*s3.Object, error) {
	svc := s3.New(session.New(), &aws.Config{Region: aws.String(region)})
	input := &s3.ListObjectsInput{Bucket: aws.String(bucket), Prefix: aws.String(path)}
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

func parsePath(dir string) (bucket, prefix string) {
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
	Name         string
	LastModified string
}

func ReadBucket(region, dir string) ([]*BucketFile, error) {
	bucket, prefix := parsePath(dir)
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
			retval = append(retval, &BucketFile{Name: aws.StringValue(object.Key), LastModified: dateTime})
		}
	}
	return retval, nil
}
