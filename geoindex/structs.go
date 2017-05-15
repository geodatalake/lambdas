package geoindex

import (
	"github.com/geodatalake/lambdas/bucket"
)

type BucketFileInfo struct {
	Region string `json:"region"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Size   int64  `json:"size"`
}

func NewBucketFileInfo(bf *bucket.BucketFile) *BucketFileInfo {
	return &BucketFileInfo{
		Region: bf.Region,
		Bucket: bf.Bucket,
		Key:    bf.Key,
		Size:   bf.Size,
	}
}

type ExtractFile struct {
	File *BucketFileInfo `json:"file"`
	Prj  *BucketFileInfo `json:"prj,omitempty"`
}

func NewExtractFile(base, prj *bucket.BucketFile) *ExtractFile {
	if prj != nil {
		return &ExtractFile{
			File: NewBucketFileInfo(base),
			Prj:  NewBucketFileInfo(prj),
		}
	}
	return &ExtractFile{
		File: NewBucketFileInfo(base),
	}
}

type DcosRequest int

const (
	ScanBucket DcosRequest = iota
	ExtractFileType
)

type ClusterRequest struct {
	RequestType DcosRequest  `json:"type"`
	File        *ExtractFile `json:"file"`
}
