// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package jobmanager

import (
	"errors"
	"time"
)

type JobHelper interface {
	DecodeJobs([]interface{}) ([]interface{}, error)
	UnmarshalJobs([]byte) ([]interface{}, error)
	GetTimeout(interface{}) time.Duration
	GetType(interface{}) string
	GetActualJob(interface{}) interface{}
}

type SyncHelper interface {
	GetPartFor(name string) int
	RegisterStart(string, int)
	RegisterStop(string, int)
	RegisterJobComplete(string, time.Duration)
	FinishPart(string, int)
	AddPending(int)
	FlushPending()
}

type LambdaInvokeResponse interface {
	Payload() []byte
}

type LambdaHelper interface {
	InvokeRequestResponse(interface{}, string) (LambdaInvokeResponse, error)
	InvokeAsync(*JobPacket, string) error
}

type JobTimeout interface {
	Timeout() time.Duration
}

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

type JobManager struct {
	driver     LambdaHelper
	jobHelper  JobHelper
	syncHelper SyncHelper
	jobTimeout JobTimeout
}

func NewJobManager() *JobManager {
	return new(JobManager)
}

func (jm *JobManager) WithDriver(driver LambdaHelper) *JobManager {
	jm.driver = driver
	return jm
}

func (jm *JobManager) WithJobHelper(jh JobHelper) *JobManager {
	jm.jobHelper = jh
	return jm
}

func (jm *JobManager) WithSyncHelper(sh SyncHelper) *JobManager {
	jm.syncHelper = sh
	return jm
}

func (jm *JobManager) WithJobTimeout(driver JobTimeout) *JobManager {
	jm.jobTimeout = driver
	return jm
}

func (jm *JobManager) UnmarshalJobPacket(m map[string]interface{}) (*JobPacket, error) {
	jp := new(JobPacket)
	if err := jp.Unmarshal(jm.checkJobHelper(), m); err != nil {
		return nil, err
	}
	return jp, nil
}

type ClusterJob struct {
	Part      int    `json:"part"`
	Last      int    `json:"last"`
	Id        string `json:"id"`
	StartTime string `json:"startTime"`
}

func NewClusterJob(id, startTime string, part, last int) *ClusterJob {
	return &ClusterJob{Id: id, StartTime: startTime, Part: part, Last: last}
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

var timeTemplates = []string{
	"2006-01-02 15:04:05.000000000 -0700 MST",
	"2006-01-02 15:04:05.00000000 -0700 MST",
	"2006-01-02 15:04:05.0000000 -0700 MST"}

func (cj *ClusterJob) CalcDuration() time.Duration {
	for _, tt := range timeTemplates {
		if st, err := time.Parse(tt, cj.StartTime); err == nil {
			return time.Now().UTC().Sub(st)
		}
	}
	return time.Duration(0)
}
