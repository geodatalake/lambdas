package geoindex

import (
	"fmt"
	"strings"

	"github.com/geodatalake/lambdas/bucket"
)

type BucketFileInfo struct {
	Region       string `json:"region"`
	Bucket       string `json:"bucket"`
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	LastModified string `json:"lastModified"`
}

func NewBucketFileInfo(bf *bucket.BucketFile) *BucketFileInfo {
	return &BucketFileInfo{
		Region:       bf.Region,
		Bucket:       bf.Bucket,
		Key:          bf.Key,
		Size:         bf.Size,
		LastModified: bf.LastModified,
	}
}

func (bfi *BucketFileInfo) AsBucketFile() *bucket.BucketFile {
	return bucket.NewBucketFile(bfi.Region, bfi.Bucket, bfi.Key, bfi.LastModified, bfi.Size)
}

func (bfi *BucketFileInfo) IsShapeFileRoot() bool {
	if _, ext, ok := getExtension(bfi.Key); ok {
		if ext == "shp" {
			return true
		}
	}
	return false
}

// As of ArcGis 9.2 (.shp, .shx, .dbf are required)
// The rest are optional
func (bfi *BucketFileInfo) IsShapeFileAux() bool {
	if _, ext, ok := getExtension(bfi.Key); ok {
		switch ext {
		case "dbf":
			return true
		case "shx":
			return true
		case "prj":
			return true
		case "xml":
			return true
		case "sbn":
			return true
		case "sbx":
			return true
		case "fbn":
			return true
		case "fbx":
			return true
		case "ain":
			return true
		case "aih":
			return true
		case "atx":
			return true
		case "ixs":
			return true
		case "mxs":
			return true
		case "cpg":
			return true
		}
	}
	return false
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

type DirRequest struct {
	Files []*bucket.BucketFile `json:"files"`
}

func (dr *DirRequest) GetKeys() []*bucket.BucketFile {
	return dr.Files
}

type DcosRequest int

const (
	ScanBucket DcosRequest = iota
	ExtractFileType
	GroupFiles
)

type ClusterRequest struct {
	RequestType DcosRequest    `json:"type"`
	Bucket      *BucketRequest `json:"bucket,omitempty"`
	File        *ExtractFile   `json:"file,omitempty"`
	DirFiles    *DirRequest    `json:"dir,omitempty"`
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
