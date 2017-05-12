// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bucket

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	CHUNKSIZE = 512
)

type ChunkReader struct {
	start, end int64
	buf        bytes.Buffer
	file       *BucketFile
	svc        *s3.S3
}

// ChunkReader never calls Read(), it merely maintains
// a buffer of data and adjusts this buffer to satisfy
// ReadAt() requests. It is completely resuable by many readers
// but it is not threadsafe
func NewChunkReader(file *BucketFile, svc *s3.S3) *ChunkReader {
	return &ChunkReader{file: file, svc: svc}
}

func (cr *ChunkReader) inSpan(off, amnt int64) bool {
	if off < cr.start {
		return false
	}
	if off+amnt > cr.end {
		return false
	}
	return true
}

// Sectionreader assures that the req is inbounds
func (cr *ChunkReader) ReadAt(p []byte, off int64) (n int, err error) {
	reqLen := int64(len(p))
	if !cr.inSpan(off, reqLen) {
		chunckLen := int64(CHUNKSIZE)
		if reqLen > chunckLen {
			chunckLen = reqLen
		}
		if max := cr.file.Size - off; max < chunckLen {
			chunckLen = max
		}
		req := &s3.GetObjectInput{
			Bucket: aws.String(cr.file.Bucket),
			Key:    aws.String(cr.file.Key),
			Range:  aws.String(fmt.Sprintf("%d-%d", off, off+chunckLen-1))} // Range are inclusive
		in, err := cr.svc.GetObject(req)
		if err != nil {
			return 0, err
		}
		defer in.Body.Close()
		if cr.buf.Cap() < int(*in.ContentLength) {
			cr.buf.Grow(int(*in.ContentLength) - cr.buf.Len())
		}
		cr.buf.Reset()

		n, err := io.Copy(&cr.buf, in.Body)
		if err != nil {
			return 0, err
		}
		cr.start = off
		cr.end = off + n - 1
	}
	cStart := off - cr.start
	cEnd := cStart + reqLen // cEnd is exclusive
	copy(p, cr.buf.Bytes()[cStart:cEnd])
	return int(cEnd - cStart), nil
}
