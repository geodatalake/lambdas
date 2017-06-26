// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jobmanager

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type HarnessJob struct {
	Amount time.Duration `json:"timeout"`
}

func (job *HarnessJob) Timeout() time.Duration {
	return job.Amount
}

func newJob(timeout time.Duration) *HarnessJob {
	retval := new(HarnessJob)
	retval.Amount = timeout
	return retval
}

type Harness struct {
	jobs         []*HarnessJob
	next         string
	job, subJob  *ClusterJob
	timeout      time.Duration
	name         string
	t            *testing.T
	jobNum       func()
	done         func()
	pending      int
	totalPending int
	sleeper      *Setter
	bloomer      *Setter
	jobLock      sync.RWMutex
}

func (h *Harness) DecodeJobs(data []interface{}) ([]interface{}, error) {
	retval := make([]interface{}, 0, len(data))
	for _, arr := range data {
		if m, ok := arr.(map[string]interface{}); ok {
			job := new(HarnessJob)
			if v := Float64For("timeout", -1.0, m); v != -1.0 {
				job.Amount = time.Duration(int64(v))
			}
		} else {
			return nil, fmt.Errorf("Job was not map[string]interface{}, it was %s", reflect.TypeOf(arr).String())
		}
	}
	return retval, nil
}

func (h *Harness) UnmarshalJobs(io []byte) ([]interface{}, error) {
	if io != nil && len(io) > 0 {
		jobs := []*HarnessJob{}
		if err := json.Unmarshal(io, &jobs); err != nil {
			h.t.Errorf("Error unmarshaling jobs %v", err)
			return nil, err
		}
		retval := make([]interface{}, 0, len(jobs))
		for _, j := range jobs {
			retval = append(retval, j)
		}
		return retval, nil
	}
	return []interface{}{}, nil
}

func (h *Harness) NumJobs() int {
	h.jobLock.RLock()
	defer h.jobLock.RUnlock()
	return len(h.jobs)
}

func (h *Harness) Timeout() time.Duration {
	return h.timeout
}
func (h *Harness) GetTimeout(job interface{}) time.Duration {
	if hj, ok := job.(*HarnessJob); ok {
		return hj.Amount
	} else {
		h.t.Errorf("Expected a HArnessJob pointer, but got %s", reflect.TypeOf(job).String())
	}
	return 2 * time.Second
}

func (h *Harness) GetType(job interface{}) string {
	if _, ok := job.(*HarnessJob); ok {
		return "foo"
	} else {
		h.t.Errorf("Expected a HArnessJob pointer, but got %s", reflect.TypeOf(job).String())
	}
	return "bah"
}

func (h *Harness) GetActualJob(job interface{}) interface{} {
	return job
}

func (h *Harness) GetPartFor(name string) int {
	return 0
}

func (h *Harness) RegisterStart(name string, num int) {
	//	log.Println("RegisterStart(", name, ")", num)
	h.jobLock.Lock()
	defer h.jobLock.Unlock()
	h.totalPending -= num
}

func (h *Harness) RegisterStop(name string, num int) {
	//	log.Println("RegisterStop(", name, ")", num)
}

func (h *Harness) RegisterJobComplete(name string, dur time.Duration) {
	h.done()
}

func (h *Harness) FinishPart(name string, num int) {

}

func (h *Harness) AddPending(num int) {
	h.jobLock.Lock()
	defer h.jobLock.Unlock()
	h.pending += num
}

func (h *Harness) FlushPending() {
	h.jobLock.Lock()
	defer h.jobLock.Unlock()
	h.totalPending += h.pending
	h.pending = 0
}

type HarnessResponse struct {
	payload []byte
}

func (resp *HarnessResponse) Payload() []byte {
	return resp.payload
}

func newHarnessResponse(payload []byte) *HarnessResponse {
	return &HarnessResponse{payload: payload}
}

func (h *Harness) InvokeRequestResponse(job interface{}, next string) (LambdaInvokeResponse, error) {
	if next != "foo1" {
		h.t.Errorf("Expected next to be foo1, but was %s", next)
	}
	if _, ok := job.(*HarnessJob); !ok {
		h.t.Errorf("Expected a HarnessJob pointer, received %s", reflect.TypeOf(job).String())
		return nil, fmt.Errorf("job not a HarnessJob, it is %s", reflect.TypeOf(job).String())
	}
	h.jobNum()
	if h.sleeper.GetNext() {
		time.Sleep(3 * time.Second)
	}
	if h.bloomer.GetNext() {
		jobs := makeJobs(5, 2*time.Second)
		if b, err := json.Marshal(jobs); err != nil {
			h.t.Errorf("Error marshaling HarnessJobs %v", err)
			return nil, err
		} else {
			return newHarnessResponse(b), nil
		}
	}
	return newHarnessResponse([]byte{}), nil
}
func (h *Harness) InvokeAsync(packet *JobPacket, next string) error {
	if next != "JobManagerARN" {
		h.t.Errorf("Expected next to be JobManagerARN, but was %s", next)
	}
	return nil
}

func makeJobs(num int, timeout time.Duration) []interface{} {
	retval := make([]interface{}, 0, num)
	for i := 0; i < num; i++ {
		retval = append(retval, newJob(timeout))
	}
	return retval
}

type Setter struct {
	mutex   sync.RWMutex
	pattern []bool
	index   int
	dflt    bool
}

func (s *Setter) WithTrue(p ...int) *Setter {
	sz := 0
	for _, v := range p {
		if v > sz {
			sz = v
		}
	}
	s.pattern = make([]bool, sz+1)
	for _, v := range p {
		s.pattern[v] = true
	}
	return s
}

func (s *Setter) GetNext() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	i := s.index
	if i >= len(s.pattern) {
		return s.dflt
	} else {
		s.index++
		return s.pattern[i]
	}
}

func TestMasterRun(t *testing.T) {
	bunch := new(Harness)
	bunch.timeout = 30 * time.Second
	bunch.t = t

	if bunch.Timeout() != 30*time.Second {
		t.Errorf("Expected bunch timeout to be 30 seconds, but was %s", bunch.Timeout().String())
	}

	cj := NewClusterJob("fu", time.Now().UTC().String(), 0, 0)
	jp := new(JobPacket).WithNexts("foo1", 5).WithClusterJobs(cj, nil)
	jp.AddJobs(makeJobs(20, 2*time.Second))
	bunch.AddPending(jp.NumJobs())

	test := NewJobManager().
		WithDriver(bunch).
		WithJobTimeout(bunch).
		WithJobHelper(bunch).
		WithSyncHelper(bunch)

	numRun := 0
	var runLock sync.RWMutex
	bunch.jobNum = func() {
		runLock.Lock()
		numRun++
		runLock.Unlock()
	}

	bunch.sleeper = new(Setter).WithTrue(9)  // one failure (job timeout)
	bunch.bloomer = new(Setter).WithTrue(14) // one bloomer (5 new jobs)
	isDone := false
	bunch.done = func() {
		isDone = true
	}

	if err := test.RunMaster(jp, "JobManagerARN"); err != nil {
		t.Errorf("RunMaster returned %v", err)
	}
	runLock.RLock()
	if numRun != 26 {
		t.Errorf("Expected 26 jobs, but received %d", numRun)
	}
	runLock.RUnlock()
	if !isDone {
		t.Errorf("Not done")
	}
	if bunch.totalPending != 0 {
		t.Errorf("Expected pending to be zero, but was %d", bunch.totalPending)
	}
}
