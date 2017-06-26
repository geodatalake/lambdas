package geoindex

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/geodatalake/lambdas/jobmanager"
)

func TestUnmarshal(t *testing.T) {
	testString := `{"type":1,"underContract":true,"file":{"file":{"region":"us-east-1","bucket":"test-data-lake-009","key":"sources/18STJ6251.tif","size":9137084,"lastModified":"2017-06-14T02:42:31Z"}}}`

	var req map[string]interface{}
	if err := json.Unmarshal([]byte(testString), &req); err != nil {
		t.Errorf("Error unmarshaing JSON string %v", err)
	} else {
		cr := new(ClusterRequest)
		if err1 := cr.Unmarshal(req); err1 != nil {
			t.Errorf("Error unmarshling %v", err1)
		} else {
			if cr.RequestType != ExtractFileType {
				t.Errorf("Expected ExtractFileType, but was %v", cr.RequestType)
			}
		}
	}
	// test Bucket requests
	testString = `{"type": 0, "underContract": false, "bucket": {"bucket": "test-data-lake-010","region": "us-west-2"}}`
	if err := json.Unmarshal([]byte(testString), &req); err != nil {
		t.Errorf("Error unmarshaing JSON string %v", err)
	} else {
		cr := new(ClusterRequest)
		if err1 := cr.Unmarshal(req); err1 != nil {
			t.Errorf("Error unmarshling %v", err1)
		} else {
			if cr.RequestType != ScanBucket {
				t.Errorf("Expected ScanBucket, but was %v", cr.RequestType)
			}
		}
	}

	// Test ClusterMaster request
	j := jobmanager.NewClusterJob("id", time.Now().UTC().String(), 0, 0)
	j1 := new(ClusterRequest)
	j1.RequestType = GroupFiles
	j1.DirFiles = &DirRequest{Region: "us-east-1", Bucket: "test-data-lake-004", Prefix: "fort-story"}
	j2 := new(ClusterRequest)
	j2.RequestType = GroupFiles
	j2.DirFiles = &DirRequest{Region: "us-east-1", Bucket: "test-data-lake-005", Prefix: "fort-story"}

	jp := new(jobmanager.JobPacket).
		WithNexts("next", 30).
		WithClusterJobs(j, nil)
	jp.AddJobs([]interface{}{NewClusterResponse(j1, "foo1"), NewClusterResponse(j2, "foo2")})

	test := new(ClusterRequest)
	test.RequestType = ClusterMaster
	test.Packet = jp

	b, err := json.Marshal(test)
	if err != nil {
		t.Errorf("Error marshaling test: %v", err)
	} else {
		if err := json.Unmarshal(b, &req); err != nil {
			t.Errorf("Error unmarshaing JSON string %v", err)
		} else {
			cr := new(ClusterRequest)
			if err1 := cr.Unmarshal(req); err1 != nil {
				t.Errorf("Error unmarshling %v", err1)
			} else {
				if cr.RequestType != ClusterMaster {
					t.Errorf("Expected ClusterMaster request, but was %v", cr.RequestType)
				}
				if cr.Packet == nil {
					t.Errorf("Packet is nil")
				} else {
					if cr.Packet.NumJobs() != 2 {
						t.Errorf("Expected 2 items, received %d", cr.Packet.NumJobs())
					}
					if cr.Packet.MyJob.Id != "id" {
						t.Errorf("Expected jobId to be id, it was %s", cr.Packet.MyJob.Id)
					}
				}
			}
		}
	}
}
