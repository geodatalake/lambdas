package geoindex

import (
	"fmt"
	"reflect"
	"strings"
	"time"

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

func (bfi *BucketFileInfo) Unmarshal(m map[string]interface{}) error {
	if r, ok := m["region"]; ok {
		bfi.Region = r.(string)
	}
	if b, ok := m["bucket"]; ok {
		bfi.Bucket = b.(string)
	}
	if k, ok := m["key"]; ok {
		bfi.Key = k.(string)
	}
	if s, ok := m["size"]; ok {
		bfi.Size = int64(s.(float64))
	}
	if lm, ok := m["lastModified"]; ok {
		bfi.LastModified = lm.(string)
	}
	return nil
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

func (ef *ExtractFile) Unmarshal(m map[string]interface{}) error {
	ef.File = new(BucketFileInfo)
	if f, ok := m["file"].(map[string]interface{}); ok {
		if err1 := ef.File.Unmarshal(f); err1 != nil {
			return fmt.Errorf("Error unmarshalling bucketFileInfo %v", err1)
		}
	} else {
		return fmt.Errorf("bfi.file is not map[string]interface{} %v", m["file"])
	}
	if a, ok := m["aux"]; ok {
		if arr, ok := a.([]interface{}); ok {
			ef.Aux = make([]*BucketFileInfo, 0, len(arr))
			for _, v := range arr {
				if bucket, ok := v.(map[string]interface{}); ok {
					bfi := new(BucketFileInfo)
					bfi.Unmarshal(bucket)
					ef.Aux = append(ef.Aux, bfi)
				} else {
					return fmt.Errorf("Aux file is not of type map[string]interface{} it is %s", reflect.TypeOf(v).String())
				}
			}
		} else {
			return fmt.Errorf("Aux is not of type []interface{} it is %s", reflect.TypeOf(a).String())
		}
	}
	return nil
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

func (br *BucketRequest) Unmarshal(m map[string]interface{}) error {
	if b, ok := m["bucket"]; ok {
		br.Bucket = b.(string)
	}
	if r, ok := m["region"]; ok {
		br.Region = r.(string)
	}
	return nil
}

type DirRequest struct {
	Files []*bucket.BucketFile `json:"files"`
}

func (dr *DirRequest) String() string {
	if len(dr.Files) != 0 {
		val := make([]string, 0, len(dr.Files))
		for _, f := range dr.Files {
			val = append(val, f.Key)
		}
		return strings.Join(val, ", ")
	} else {
		return "empty"
	}
}

func (dr *DirRequest) Unmarshal(m map[string]interface{}) error {
	if a, ok := m["files"]; ok {
		if arr, ok := a.([]interface{}); ok {
			dr.Files = make([]*bucket.BucketFile, 0, len(arr))
			for _, v := range arr {
				if bf, ok := v.(map[string]interface{}); ok {
					resolved := new(bucket.BucketFile)
					if err := resolved.Unmarshal(bf); err != nil {
						return fmt.Errorf("Error unmarshaing bucket.BucketFile %v", err)
					}
					dr.Files = append(dr.Files, resolved)
				}
			}
		} else {
			return fmt.Errorf("files is not of type []interface{}, it is %s", reflect.TypeOf(a).String())
		}
	} else {
		return fmt.Errorf("No files found in map %v", m)
	}
	return nil
}

func (dr *DirRequest) GetKeys() []*bucket.BucketFile {
	return dr.Files
}

type DcosRequest int

const (
	ScanBucket DcosRequest = iota
	ExtractFileType
	GroupFiles
	MineSQS
	ClusterMaster
)

type ClusterRequest struct {
	RequestType DcosRequest    `json:"type"`
	Bucket      *BucketRequest `json:"bucket,omitempty"`
	File        *ExtractFile   `json:"file,omitempty"`
	DirFiles    *DirRequest    `json:"dir,omitempty"`
	Master      *ClusterQueue  `json:"master,omitempty"`
	Id          string         `json:"id"`
}

func (cr *ClusterRequest) String() string {
	switch cr.RequestType {
	case ScanBucket:
		return fmt.Sprintf("Request: ScanBucket, Id: %s, Bucket: %+v", cr.Id, *cr.Bucket)
	case ExtractFileType:
		return fmt.Sprintf("Request: ExtractFileType, Id: %s, %v", cr.Id, cr.File)
	case GroupFiles:
		return fmt.Sprintf("Request: GroupFiles, Id: %s, %s", cr.Id, cr.DirFiles.String())
	case MineSQS:
		return fmt.Sprintf("Request: MineSQS, Id: %s, %+v", cr.Id, *cr.Master)
	case ClusterMaster:
		return fmt.Sprintf("Request: ClusterMaster, Id: %s, %+v", cr.Id, *cr.Master)
	default:
		return "Unknown request type"
	}
}

func (cr *ClusterRequest) Unmarshal(m map[string]interface{}) error {
	if id, ok := m["id"]; ok {
		cr.Id = id.(string)
	}
	if rtype, ok := m["type"]; ok {
		if val, good := rtype.(float64); good {
			switch int64(val) {
			case 0:
				cr.RequestType = ScanBucket
				if b, ok := m["bucket"]; ok {
					if bf, ok := b.(map[string]interface{}); ok {
						cr.Bucket = new(BucketRequest)
						if err := cr.Bucket.Unmarshal(bf); err != nil {
							return fmt.Errorf("Error Unmarshing BucketRequest %v", err)
						}
					}
				} else {
					return fmt.Errorf("No bucket found in request")
				}
			case 1:
				cr.RequestType = ExtractFileType
				if f, ok := m["file"]; ok {
					if ef, ok := f.(map[string]interface{}); ok {
						cr.File = new(ExtractFile)
						if err1 := cr.File.Unmarshal(ef); err1 != nil {
							return fmt.Errorf("Error extracting file %v", err1)
						}
					} else {
						return fmt.Errorf("file is not of type map[string]interface{}  %s", reflect.TypeOf(f).String())
					}
				} else {
					return fmt.Errorf("No file found in request")
				}
			case 2:
				cr.RequestType = GroupFiles
				if d, ok := m["dir"]; ok {
					if di, ok := d.(map[string]interface{}); ok {
						cr.DirFiles = new(DirRequest)
						if err := cr.DirFiles.Unmarshal(di); err != nil {
							return fmt.Errorf("Error unmarshing DirRequest %v", err)
						}
					} else {
						return fmt.Errorf("dir is not of type map[string]interface{}  %s", reflect.TypeOf(d).String())
					}
				} else {
					return fmt.Errorf("No dir found in request")
				}
			case 3:
				cr.RequestType = MineSQS
				if master, ok := m["master"]; ok {
					if data, good := master.(map[string]interface{}); good {
						cr.Master = new(ClusterQueue)
						if err := cr.Master.Unmarshal(data); err != nil {
							return err
						}
					}
				}
			case 4:
				cr.RequestType = ClusterMaster
				if master, ok := m["master"]; ok {
					if data, good := master.(map[string]interface{}); good {
						cr.Master = new(ClusterQueue)
						if err := cr.Master.Unmarshal(data); err != nil {
							return err
						}
					}
				}
			}
		} else {
			return fmt.Errorf("Request type is not an float64 %s", reflect.TypeOf(rtype).String())
		}
	} else {
		return fmt.Errorf("Unable to determine request type from %v", m["type"])
	}
	return nil
}

type ClusterResponse struct {
	Item     *ClusterRequest `json:"item"`
	Timeout  time.Duration   `json:"timeout"`
	ParentId string          `json:"parentId"`
}

func NewClusterResponse(cr *ClusterRequest, id string) *ClusterResponse {
	return &ClusterResponse{Item: cr, ParentId: id, Timeout: time.Minute}
}

func (cr *ClusterResponse) Unmarshal(m map[string]interface{}) error {
	if to, ok := m["timeout"]; ok {
		if nto, good := to.(float64); good {
			cr.Timeout = time.Duration(int64(nto))
		}
	} else {
		cr.Timeout = time.Minute
	}
	if pi, ok := m["parentId"]; ok {
		cr.ParentId = pi.(string)
	}
	if item, ok := m["item"]; ok {
		if cItem, good := item.(map[string]interface{}); good {
			cr.Item = new(ClusterRequest)
			cr.Item.Unmarshal(cItem)
		} else {
			return fmt.Errorf("item is not a map[string]interface{}, it is a %s", reflect.TypeOf(item).String())
		}
	} else {
		return fmt.Errorf("No item found")
	}
	return nil
}

type ClusterQueue struct {
	Items     []*ClusterResponse `json:"items"`
	Next      string             `json:"next"`
	MaxNext   int                `json:"maxNext"`
	StartTime string             `json:"startTime"`
	ParentId  string             `json:"parentId"`
}

func (cq *ClusterQueue) CloneWith(items []*ClusterResponse) *ClusterQueue {
	c := new(ClusterQueue)
	c.Next = cq.Next
	c.MaxNext = cq.MaxNext
	c.StartTime = cq.StartTime
	c.ParentId = cq.ParentId
	c.Items = items
	return c
}

func (cq *ClusterQueue) Unmarshal(m map[string]interface{}) error {
	if a, ok := m["items"]; ok {
		if all, good := a.([]interface{}); good {
			cq.Items = make([]*ClusterResponse, 0, len(all))
			for _, item := range all {
				cr := new(ClusterResponse)
				if err := cr.Unmarshal(item.(map[string]interface{})); err != nil {
					return err
				}
				cq.Items = append(cq.Items, cr)
			}
		}
	} else {
		return fmt.Errorf("No items found")
	}
	if next, ok := m["next"]; ok {
		cq.Next = next.(string)
	} else {
		return fmt.Errorf("next not found")
	}
	if st, ok := m["startTime"]; ok {
		cq.StartTime = st.(string)
	} else {
		return fmt.Errorf("startTime not found")
	}
	if mn, ok := m["maxNext"]; ok {
		if mx, good := mn.(float64); good {
			cq.MaxNext = int(mx)
		}
	} else {
		return fmt.Errorf("maxNext not found")
	}
	return nil
}

type IndexerRequestType int

const (
	Enter IndexerRequestType = iota
	Leave
	Reset
)

type IndexerRequest struct {
	RequestType IndexerRequestType `json:"type"`
	Name        string             `json:"name"`
	Num         int                `json:"num"`
}

func (ir *IndexerRequest) String() string {
	rtype := "Unknown"
	switch ir.RequestType {
	case Enter:
		rtype = "Enter"
	case Leave:
		rtype = "Leave"
	case Reset:
		rtype = "Reset"
	}
	return fmt.Sprintf("%s: Name: %s, Num: %d", rtype, ir.Name, ir.Num)
}

func (ir *IndexerRequest) Unmarshal(m map[string]interface{}) error {
	if v, ok := m["type"]; ok {
		if val, good := v.(float64); good {
			switch int64(val) {
			case 0:
				ir.RequestType = Enter
			case 1:
				ir.RequestType = Leave
			case 2:
				ir.RequestType = Reset
			default:
				return fmt.Errorf("Unknown IndexerRequestType %v", int64(val))
			}
		}
	} else {
		return fmt.Errorf("No type property found")
	}
	if n, ok := m["name"]; ok {
		ir.Name = n.(string)
	} else {
		return fmt.Errorf("No name property found")
	}
	if a, ok := m["num"]; ok {
		if num, good := a.(float64); good {
			ir.Num = int(num)
		} else {
			return fmt.Errorf("Num is not a float64, it is %s", reflect.TypeOf(a).String())
		}
	} else {
		return fmt.Errorf("No num property found")
	}
	return nil
}

type IndexerResponse struct {
	Success bool `json:"success"`
	Num     int  `json:"num"`
}

func NewIndexerResponse(s bool, num int) *IndexerResponse {
	return &IndexerResponse{Success: s, Num: num}
}
