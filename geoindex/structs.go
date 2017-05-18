package geoindex

import (
	"fmt"
	"strings"

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

func (bfi *BucketFileInfo) AsBucketFile() *bucket.BucketFile {
	return bucket.NewBucketFile(bfi.Region, bfi.Bucket, bfi.Key, "", bfi.Size)
}

type ExtractFile struct {
	File *BucketFileInfo   `json:"file"`
	Aux  []*BucketFileInfo `json:"aux,omitempty"`
}

func (ef *ExtractFile) String() string {
	if ef.Aux == nil || len(ef.Aux) == 0 {
		return fmt.Sprintf("File: %+v", *ef.File)
	}
	aux := make([]string, 0, len(ef.Aux))
	for _, bf := range ef.Aux {
		aux = append(aux, fmt.Sprintf("%+v", *bf))
	}
	return fmt.Sprintf("File: %+v, Aux: %s", *ef.File, strings.Join(aux, ", "))
}

func NewExtractFile(base *bucket.BucketFile, aux []*bucket.BucketFile) *ExtractFile {
	if aux != nil && len(aux) > 0 {
		newAux := make([]*BucketFileInfo, 0, len(aux))
		for _, bf := range aux {
			newAux = append(newAux, NewBucketFileInfo(bf))
		}
		return &ExtractFile{
			File: NewBucketFileInfo(base),
			Aux:  newAux,
		}
	}
	return &ExtractFile{
		File: NewBucketFileInfo(base),
	}
}

type BucketRequest struct {
	Bucket string `json:"bucket"`
	Region string `json:"region"`
}

type DcosRequest int

const (
	ScanBucket DcosRequest = iota
	ExtractFileType
)

type ClusterRequest struct {
	RequestType DcosRequest    `json:"type"`
	Bucket      *BucketRequest `json:"bucket,omitempty"`
	File        *ExtractFile   `json:"file"`
}

func (cr *ClusterRequest) String() string {
	switch cr.RequestType {
	case ScanBucket:
		return fmt.Sprintf("Request: ScanBucket, Bucket: %+v", *cr.Bucket)
	case ExtractFileType:
		return fmt.Sprintf("Request: ExtractFileType, %v", cr.File)
	default:
		return "Unknown request type"
	}
}
