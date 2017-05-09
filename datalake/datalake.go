// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datalake

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type DataLakeConn struct {
	accessKey, secretKey, apiEndPoint string
}

func NewDataLakeConn(accessKey, secretKey, apiEndPoint string) *DataLakeConn {
	return &DataLakeConn{accessKey, secretKey, apiEndPoint}
}

func makeManifest(s3Uris []string) ([]byte, error) {
	manifest := make(map[string]interface{})
	locs := make([]map[string]interface{}, len(s3Uris))
	for index, uri := range s3Uris {
		locs[index] = make(map[string]interface{})
		locs[index]["url"] = uri
	}
	manifest["fileLocations"] = locs
	return json.Marshal(manifest)
}

func (dlc *DataLakeConn) CreatePackage(name, description string) (*PackageInfo, error) {
	data := make(map[string]interface{})
	data["package_name"] = name
	data["package_description"] = description
	url := fmt.Sprintf("https://%s/prod/packages/new", dlc.apiEndPoint)
	client := http.Client{}
	body, err1 := json.Marshal(data)
	if err1 != nil {
		return nil, err1
	}
	req, err2 := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err2 != nil {
		return nil, err2
	}
	req.Header.Add("Auth", dlc.Auth())
	resp, err3 := client.Do(req)
	if err3 != nil {
		return nil, err3
	}
	defer resp.Body.Close()
	b, err4 := ioutil.ReadAll(resp.Body)
	if err4 != nil {
		return nil, err4
	}
	retval := new(PackageInfo)
	if err5 := json.Unmarshal(b, retval); err5 != nil {
		return nil, err5
	}
	return retval, nil
}

func (dlc *DataLakeConn) CreatePackageWithManifest(name, description, datasetName string, s3Uris []string) (*StatusResponse, error) {
	info, err := dlc.CreatePackage(name, description)
	if err != nil {
		return nil, err
	}
	status, err1 := dlc.NewDataset(info.PackageId, datasetName)
	if err1 != nil {
		return nil, err1
	}
	manifest, e := makeManifest(s3Uris)
	if e != nil {
		return nil, e
	}
	err2 := dlc.UploadManifest(status, manifest)
	if err2 != nil {
		return nil, err2
	}
	return dlc.Process(status)
}

func (dlc *DataLakeConn) UploadManifest(status *StatusResponse, manifest []byte) error {
	client := http.Client{}
	b, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	req, err1 := http.NewRequest("POST", status.UploadUrl, bytes.NewBuffer(b))
	if err1 != nil {
		return err1
	}
	resp, err2 := client.Do(req)
	if err2 != nil {
		return err2
	}
	if resp.Body != nil {
		resp.Body.Close()
	}
	return nil
}

func (dlc *DataLakeConn) NewDataset(packageId, name string) (*StatusResponse, error) {
	url := fmt.Sprintf("https://%s/prod/packages/%s/datasets/new", dlc.apiEndPoint, packageId)
	body, err := json.Marshal(NewDatasetNewRequest(name))
	if err != nil {
		return nil, err
	}
	return dlc.statusStep("POST", url, string(body))
}

func (dlc *DataLakeConn) Process(resp *StatusResponse) (*StatusResponse, error) {
	url := fmt.Sprintf("https://%s/prod/packages/%s/datasets/%s/process", dlc.apiEndPoint, resp.PackageId, resp.DatasetId)
	return dlc.statusStep("POST", url, "{}")
}

func (dlc *DataLakeConn) statusStep(method, url, body string) (*StatusResponse, error) {
	client := http.Client{}
	var req *http.Request
	var err error
	if len(body) > 0 {
		req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Add("Auth", dlc.Auth())
	if resp, err2 := client.Do(req); err2 != nil {
		return nil, err2
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, errors.New(fmt.Sprintf("Expected status response 200, received %d", resp.StatusCode))
		}
		b, err3 := ioutil.ReadAll(resp.Body)
		if err3 != nil {
			return nil, err3
		}
		retval := new(StatusResponse)
		if e := json.Unmarshal(b, retval); e != nil {
			return nil, e
		}
		return retval, nil
	}
}

func (dlc *DataLakeConn) kDate(datestamp string) string {
	hash := hmac.New(sha256.New, []byte("DATALAKE4"+dlc.secretKey))
	hash.Write([]byte(datestamp))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func (dlc *DataLakeConn) kEndPoint(kDate string) string {
	hash := hmac.New(sha256.New, []byte(kDate))
	hash.Write([]byte(dlc.apiEndPoint))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func (dlc *DataLakeConn) kService(kEndpoint string) string {
	hash := hmac.New(sha256.New, []byte(kEndpoint))
	hash.Write([]byte("datalake"))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func (dlc *DataLakeConn) kSigning(kService string) string {
	hash := hmac.New(sha256.New, []byte(kService))
	hash.Write([]byte("datalake4_request"))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func (dlc *DataLakeConn) signature(dateString string) string {
	var currentValue string
	for i := 0; i < 4; i++ {
		switch i {
		case 0:
			currentValue = dlc.kDate(dateString)
		case 1:
			currentValue = dlc.kEndPoint(currentValue)
		case 2:
			currentValue = dlc.kService(currentValue)
		case 3:
			currentValue = dlc.kSigning(currentValue)
		}
	}
	return currentValue
}

func joinColon(s ...string) string {
	return strings.Join(s, ":")
}

func (dlc *DataLakeConn) authKey(dateString string) string {
	return base64.StdEncoding.EncodeToString([]byte(joinColon(dlc.accessKey, dlc.signature(dateString))))
}

func (dlc *DataLakeConn) Auth() string {
	return dlc.AuthWDate(time.Now().UTC().Format("20060102"))
}

func (dlc *DataLakeConn) AuthWDate(dateString string) string {
	return joinColon("ak", dlc.authKey(dateString))
}
