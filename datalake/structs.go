// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datalake

import "github.com/satori/go.uuid"

// "{\"package_id\":\"HkCFxU9y-\",\"created_at\":\"2017-05-05T19:20:37Z\",\"updated_at\":\"2017-05-05T19:20:37Z\",\"owner\":\"stevei_spaceflightindustries_com\",\"name\":\"test-package-009\",\"description\":\"009 description goes here\",\"deleted\":false}"
type PackageInfo struct {
	PackageId   string `json:"package_id"`
	Created     string `json:"created_at"`
	Updated     string `json:"updated_at"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Deleted     bool   `json:"deleted"`
}

// Dataset New Response
//  "{\"name\":\"mosul_dam.manifest.json\",\"type\":\"manifest\",\"content_type\":\"application/json\",\"owner\":\"stevei_spaceflightindustries_com\",\"package_id\":\"HJqL1UwyZ\",\"dataset_id\":\"ryKxUUP1b\",\"created_at\":\"2017-05-03T13:06:57Z\",\"updated_at\":\"2017-05-03T13:06:57Z\",\"created_by\":\"stevei_spaceflightindustries_com\",\"s3_bucket\":\"data-lake-us-west-2-414519249282\",\"s3_key\":\"HJqL1UwyZ/1493816817145/mosul_dam.manifest.json\",\"state_desc\":\"Pending Upload\",\"uploadUrl\":\"https://data-lake-us-west-2-414519249282.s3-us-west-2.amazonaws.com/HJqL1UwyZ/1493816817145/mosul_dam.manifest.json?Content-Type=application%2Fjson&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAJ2LJA54FFRGT6NDA%2F20170503%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20170503T130657Z&X-Amz-Expires=900&X-Amz-Security-Token=FQoDYXdzEL3%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDKPa2iQMpKRu1drlHyLyAVS2%2BEjKbnKGkCG8lr%2FnZp848G5O02lwHxNmCadGMMzJenlwBc%2FMetYhBwBbYj8ZnUt3yU4WA6NVardnG4WtgtfqxzwQickSKjtutZRXOxsy7g0PFFe5iYm88VwOfy5Z7SKQ9P%2BFgEQ2gWdWf%2FUcY%2B9rHyC%2Bige10p5Pw03VfZKFSPs%2F4Qu4EBFPYtjHytBnN%2Bg%2BuP%2FgZ9asTRL6rXpPIhfM%2FJf2ef5N1Qn20sTZ1L04CNBtluky67yyzAbrNqt6Wh3tTQd7S5m3Jau6%2BvO5lkZypSJyVgs7Z%2BpxOdU3iehQCa9GNIxI%2BzcWfihx24tE5w72KJeUp8gF&X-Amz-Signature=c5ece9660067c99d533c487f85ac4f19ed358568ff057590c50f1dbf284e8bdb&X-Amz-SignedHeaders=host%3Bx-amz-server-side-encryption%3Bx-amz-server-side-encryption-aws-kms-key-id&x-amz-server-side-encryption=aws%3Akms&x-amz-server-side-encryption-aws-kms-key-id=alias%2Fdatalake-us-west-2\"}"
// Process Response
// "{\"updated_at\":\"2017-05-03T13:06:57Z\",\"package_id\":\"HJqL1UwyZ\",\"created_at\":\"2017-05-03T13:06:57Z\",\"s3_bucket\":\"data-lake-us-west-2-414519249282\",\"content_type\":\"application/json\",\"created_by\":\"stevei_spaceflightindustries_com\",\"dataset_id\":\"ryKxUUP1b\",\"owner\":\"stevei_spaceflightindustries_com\",\"state_desc\":\"Processing\",\"name\":\"mosul_dam.manifest.json\",\"s3_key\":\"HJqL1UwyZ/1493816817145/mosul_dam.manifest.json\",\"type\":\"manifest\"}"
type StatusResponse struct {
	PackageId        string `json:"package_id"`
	DatasetId        string `json:"dataset_id"`
	Created          string `json:"created_at"`
	Updated          string `json:"updated_at"`
	S3Bucket         string `json:"s3_bucket"`
	S3Key            string `json:"s3_key"`
	ContentType      string `json:"content_type"`
	CreatedBy        string `json:"created_by"`
	Owner            string `json:"owner"`
	StateDescription string `json:"state_desc"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	UploadUrl        string `json:"uploadUrl"`
}

//  "{\"name\":\"mosul_dam.manifest.json\",\"type\":\"manifest\",\"content_type\":\"application/json\",\"owner\":\"S3 Import\"}"
type DatasetNewRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	ContentType string `json:"content_type"`
	Owner       string `json:"owner"`
}

// This creates a dataset new request with appropriate defaults (hardcodes)
func NewDatasetNewRequest(name string) *DatasetNewRequest {
	return &DatasetNewRequest{
		Name:        name + "_" + uuid.NewV4().String() + ".json",
		Type:        "manifest",
		ContentType: "application/json",
		Owner:       "S3 Import",
	}
}
