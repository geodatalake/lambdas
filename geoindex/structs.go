package geoindex

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/jobmanager"
)

func Float64For(name string, dflt float64, m map[string]interface{}) float64 {
	if val, ok := m[name]; ok {
		if v, good := val.(float64); good {
			return v
		}
	}
	return dflt
}

func IntFor(name string, dflt int, m map[string]interface{}) int {
	return int(Float64For(name, float64(dflt), m))
}

func StringFor(name, dflt string, m map[string]interface{}) string {
	if val, ok := m[name]; ok {
		return val.(string)
	}
	return dflt
}

func SubpropFor(name string, m map[string]interface{}) (map[string]interface{}, bool) {
	if j, ok := m[name]; ok {
		if data, good := j.(map[string]interface{}); good {
			return data, good
		}
	}
	return nil, false
}

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
	bfi.Region = StringFor("region", "", m)
	bfi.Bucket = StringFor("bucket", "", m)
	bfi.Key = StringFor("key", "", m)
	bfi.LastModified = StringFor("lastModified", "", m)
	if s, ok := m["size"]; ok {
		bfi.Size = int64(s.(float64))
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
	if f, ok := SubpropFor("file", m); ok {
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
	br.Bucket = StringFor("bucket", "", m)
	br.Region = StringFor("region", "", m)
	return nil
}

type DirRequest struct {
	Bucket string `json:"bucket"`
	Region string `json:"region"`
	Prefix string `json:"prefix"`
}

func (dr *DirRequest) String() string {
	return fmt.Sprintf("{DirRequest: Region: %s, Bucket: %s, Prefix: %s}", dr.Region, dr.Bucket, dr.Prefix)
}

func (dr *DirRequest) Unmarshal(m map[string]interface{}) error {
	dr.Bucket = StringFor("bucket", "", m)
	dr.Region = StringFor("region", "", m)
	dr.Prefix = StringFor("prefix", "", m)
	return nil
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
	RequestType DcosRequest           `json:"type"`
	Bucket      *BucketRequest        `json:"bucket,omitempty"`
	File        *ExtractFile          `json:"file,omitempty"`
	DirFiles    *DirRequest           `json:"dir,omitempty"`
	Packet      *jobmanager.JobPacket `json:"packet,omitempty"`
	Id          string                `json:"id"`
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
		return fmt.Sprintf("Request: MineSQS, Id: %s, %+v", cr.Id, cr.Packet)
	case ClusterMaster:
		return fmt.Sprintf("Request: ClusterMaster, Id: %s, %+v", cr.Id, cr.Packet)
	default:
		return "Unknown request type"
	}
}

func (cr *ClusterRequest) Unmarshal(m map[string]interface{}) error {
	cr.Id = StringFor("id", "", m)
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
				var err error
				if packet, ok := m["packet"]; ok {
					if data, good := packet.(map[string]interface{}); good {
						manager := jobmanager.NewJobManager().
							WithJobHelper(&JobHelper{})
						cr.Packet, err = manager.UnmarshalJobPacket(data)
						if err != nil {
							return err
						}
					}
				}
			case 4:
				cr.RequestType = ClusterMaster
				var err error
				if packet, ok := m["packet"]; ok {
					if data, good := packet.(map[string]interface{}); good {
						manager := jobmanager.NewJobManager().
							WithJobHelper(&JobHelper{})
						cr.Packet, err = manager.UnmarshalJobPacket(data)
						if err != nil {
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
	to := Float64For("timeout", 0.0, m)
	if to != 0.0 {
		cr.Timeout = time.Duration(int64(to))
	} else {
		cr.Timeout = time.Minute
	}
	cr.ParentId = StringFor("parentId", "", m)
	if cItem, ok := SubpropFor("item", m); ok {
		cr.Item = new(ClusterRequest)
		if e := cr.Item.Unmarshal(cItem); e != nil {
			return e
		}
	} else {
		return fmt.Errorf("No item found")
	}
	return nil
}

type ClusterJob struct {
	Part      int    `json:"part"`
	Last      int    `json:"last"`
	Id        string `json:"id"`
	StartTime string `json:"startTime"`
}

func (cj *ClusterJob) Unmarshal(m map[string]interface{}) error {
	cj.Id = StringFor("id", "", m)
	cj.StartTime = StringFor("startTime", "", m)
	cj.Part = IntFor("part", -1, m)
	cj.Last = IntFor("last", -1, m)
	if cj.Part == -1 {
		return errors.New("part is missing")
	}
	if cj.Last == -1 {
		return errors.New("last is missing")
	}
	return nil
}

func (cj *ClusterJob) IsLastPart() bool {
	return cj.Part == cj.Last
}

func (cj *ClusterJob) CalcDuration() time.Duration {
	st, err := time.Parse("2006-01-02 15:04:05.000000000 -0700 MST", cj.StartTime)
	if err != nil {
		log.Println("Job", cj.Id, "completed but time", cj.StartTime, "could not be parsed", err)
		return time.Duration(0)
	} else {
		return time.Now().UTC().Sub(st)
	}
}

type ClusterQueue struct {
	Items   []*ClusterResponse `json:"items"`
	Next    string             `json:"next"`
	MaxNext int                `json:"maxNext"`
	Job     *ClusterJob        `json:"job"`
	SubJob  *ClusterJob        `json:"subjob,omitempty"`
}

func (cq *ClusterQueue) CloneWith(items []*ClusterResponse) *ClusterQueue {
	c := new(ClusterQueue)
	c.Next = cq.Next
	c.MaxNext = cq.MaxNext
	c.Items = items
	c.Job = cq.Job
	c.SubJob = cq.SubJob
	return c
}

func (cq *ClusterQueue) IsLastPart() bool {
	return cq.Job.IsLastPart()
}

func (cq *ClusterQueue) IsSubJob() bool {
	return cq.SubJob != nil
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
	cq.Next = StringFor("next", "", m)
	cq.MaxNext = IntFor("maxNext", 10, m)

	if job, ok := SubpropFor("job", m); ok {
		cq.Job = new(ClusterJob)
		if e := cq.Job.Unmarshal(job); e != nil {
			return e
		}
	}
	if j, ok := m["subJob"]; ok {
		if job, good := j.(map[string]interface{}); good {
			cq.SubJob = new(ClusterJob)
			if e := cq.SubJob.Unmarshal(job); e != nil {
				return e
			}
		}
	}
	return nil
}

type IndexerRequestType int

const (
	Enter IndexerRequestType = iota
	Leave
	Reset
	JobComplete
	ActivePart
	PartComplete
)

type IndexerRequest struct {
	RequestType IndexerRequestType `json:"type"`
	Name        string             `json:"name"`
	Num         int                `json:"num"`
	Duration    time.Duration      `json:"duration"`
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
	case JobComplete:
		rtype = "JobComplete"
	case ActivePart:
		rtype = "ActivePart"
	case PartComplete:
		rtype = "PartComplete"
	}
	return fmt.Sprintf("%s: Name: %s, Num: %d, Duration: %s", rtype, ir.Name, ir.Num, ir.Duration.String())
}

func (ir *IndexerRequest) Unmarshal(m map[string]interface{}) error {
	if v, ok := m["type"]; ok {
		if val, good := v.(float64); good {
			ir.RequestType = IndexerRequestType(int64(val))
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
	if d, ok := m["duration"]; ok {
		if dur, good := d.(float64); good {
			ir.Duration = time.Duration(int64(dur))
		}
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
