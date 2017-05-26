// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bucket

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	CHUNKSIZE = 1024 + 512
)

type S3Reader struct {
	svc  *s3.S3
	file *BucketFile
}

func (r *S3Reader) ReadAt(p []byte, off int64) (int, error) {
	reqLen := int64(len(p))
	log.Println("Requesting S3", off, off+reqLen-1, reqLen, "bytes")
	req := &s3.GetObjectInput{
		Bucket: aws.String(r.file.Bucket),
		Key:    aws.String(r.file.Key),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", off, off+reqLen-1))} // Range are inclusive
	in, err := r.svc.GetObject(req)
	if err != nil {
		return -1, err
	}
	defer in.Body.Close()
	b, err := ioutil.ReadAll(in.Body)
	if err != nil {
		return -1, err
	}
	copy(p, b)
	return len(b), nil
}

func NewS3Reader(file *BucketFile, svc *s3.S3) *S3Reader {
	return &S3Reader{file: file, svc: svc}
}

// Implements io.WriteCloser
type S3Writer struct {
	file *BucketFile
	svc  *s3.S3
	buf  bytes.Buffer
}

func NewS3Writer(bf *BucketFile, svc *s3.S3) *S3Writer {
	return &S3Writer{file: bf, svc: svc}
}

func (w *S3Writer) Write(p []byte) (int, error) {
	if w.file == nil {
		return 0, errors.New("Writer is closed")
	}
	return w.buf.Write(p)
}

func (w *S3Writer) Close() error {
	if w.file != nil {
		o := new(s3.PutObjectInput)
		o.Body = NewByteBuffer(w.buf.Bytes())
		o.Bucket = aws.String(w.file.Bucket)
		o.Key = aws.String(w.file.Key)
		_, err := w.svc.PutObject(o)
		w.buf.Reset()
		w.file = nil
		if err != nil {
			return err
		}
	}
	return nil
}

type ChunkReader struct {
	start, end int64
	buf        bytes.Buffer
	size       int64
	reader     io.ReaderAt
}

// ChunkReader never calls Read(), it merely maintains
// a buffer of data and adjusts this buffer to satisfy
// ReadAt() requests. It is completely resuable by many readers
// but it is not threadsafe
func NewChunkReader(sz int64, r io.ReaderAt) *ChunkReader {
	return &ChunkReader{size: sz, reader: r}
}

func (cr *ChunkReader) inSpan(off, amnt int64) bool {
	if off < cr.start {
		return false
	}
	if off+amnt-1 > cr.end {
		return false
	}
	return true
}

// Sectionreader assures that the req is inbounds
func (cr *ChunkReader) ReadAt(p []byte, off int64) (int, error) {
	reqLen := int64(len(p))
	if !cr.inSpan(off, reqLen) {
		chunkLen := int64(CHUNKSIZE)
		if reqLen > chunkLen {
			chunkLen = reqLen
		}
		if max := cr.size - off; max < chunkLen {
			chunkLen = max
		}
		p1 := make([]byte, chunkLen)
		rsize, err := cr.reader.ReadAt(p1, off)
		if err != nil {
			log.Println("Error requesting", err)
			return 0, err
		}
		if cr.buf.Cap() < int(rsize) {
			cr.buf.Grow(int(rsize) - cr.buf.Len())
		}
		cr.buf.Reset()

		n, err := io.Copy(&cr.buf, bytes.NewBuffer(p1))
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
