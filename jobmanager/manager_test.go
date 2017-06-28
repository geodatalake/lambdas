// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jobmanager

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"
)

type HarnessJob struct {
	Amount time.Duration `json:"timeout"`
	Id     int
}

func (job *HarnessJob) String() string {
	return fmt.Sprintf("Job{id: %d}", job.Id)
}

func (job *HarnessJob) Timeout() time.Duration {
	return job.Amount
}

func newJob(timeout time.Duration, id int) *HarnessJob {
	retval := new(HarnessJob)
	retval.Amount = timeout
	retval.Id = id
	return retval
}

type Harness struct {
	jobs         []*HarnessJob
	record       []bool
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
	delay        int64
	forwared     int
	packets      *list.List
	pauseFlush   bool
	part         int
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
		h.t.Errorf("Expected a HarnessJob pointer, but got %s", reflect.TypeOf(job).String())
	}
	return 2 * time.Second
}

func (h *Harness) GetType(job interface{}) string {
	if _, ok := job.(*HarnessJob); ok {
		return "foo"
	} else {
		h.t.Errorf("Expected a HarnessJob pointer, but got %s", reflect.TypeOf(job).String())
	}
	return "bah"
}

func (h *Harness) GetActualJob(job interface{}) interface{} {
	return job
}

func (h *Harness) GetPartFor(name string) int {
	return h.part
}

func (h *Harness) RegisterStart(name string, num int) {
	h.jobLock.Lock()
	defer h.jobLock.Unlock()
	log.Println("Start", name, num)
	h.totalPending -= num
}

func (h *Harness) RegisterStop(name string, num int) {
	log.Println("Stop", name, num)
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
	if h.pauseFlush {
		time.Sleep(2 * time.Second)
	}
}

type HarnessResponse struct {
	payload []byte
	key     int
	id      int
}

func (resp *HarnessResponse) Payload() []byte {
	return resp.payload
}
func (resp *HarnessResponse) GetId() int {
	return resp.id
}

func (resp *HarnessResponse) String() string {
	return fmt.Sprintf("HarnessResponse id: %d has %d bytes", resp.id, resp.key)
}

var nextId int = -1
var hrLock sync.RWMutex

func newHarnessResponse(p []byte) *HarnessResponse {
	hrLock.Lock()
	defer hrLock.Unlock()
	nextId++
	return &HarnessResponse{payload: p, key: len(p), id: nextId}
}

func (h *Harness) InvokeRequestResponse(job interface{}, next string) (LambdaInvokeResponse, error) {
	if next != "foo1" {
		h.t.Errorf("Expected next to be foo1, but was %s", next)
	}
	if j, ok := job.(*HarnessJob); !ok {
		h.t.Errorf("Expected a HarnessJob pointer, received %s", reflect.TypeOf(job).String())
		return nil, fmt.Errorf("job not a HarnessJob, it is %s", reflect.TypeOf(job).String())
	} else {
		if !h.record[j.Id] {
			h.record[j.Id] = true
		}
	}
	if h.delay > 0 {
		time.Sleep(time.Duration(h.delay) * time.Millisecond)
	}
	h.jobNum()
	if h.sleeper.GetNext() {
		h.bloomer.GetNext() // advance the bloomer too
		time.Sleep(3 * time.Second)
	} else {
		if h.bloomer.GetNext() {
			jobs := makeJobs(5, 2*time.Second, 20)
			if b, err := json.Marshal(jobs); err != nil {
				h.t.Errorf("Error marshaling HarnessJobs %v", err)
				return nil, err
			} else {
				return newHarnessResponse(b), nil
			}
		}
	}
	return newHarnessResponse([]byte{}), nil
}
func (h *Harness) InvokeAsync(packet *JobPacket, next string) error {
	if next != "JobManagerARN" {
		h.t.Errorf("Expected next to be JobManagerARN, but was %s", next)
	}
	log.Println("Async", packet)
	h.packets.PushFront(packet)
	return nil
}

func makeJobs(num int, timeout time.Duration, start int) []interface{} {
	retval := make([]interface{}, 0, num)
	for i := 0; i < num; i++ {
		retval = append(retval, newJob(timeout, start+i))
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
	bunch.record = make([]bool, 25)

	if bunch.Timeout() != 30*time.Second {
		t.Errorf("Expected bunch timeout to be 30 seconds, but was %s", bunch.Timeout().String())
	}

	cj := NewClusterJob("fu", time.Now().UTC().String(), 0, 0)
	jp := new(JobPacket).WithNexts("foo1", 5).WithClusterJobs(cj, nil)
	jp.AddJobs(makeJobs(20, 2*time.Second, 0))
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

	for i, v := range bunch.record {
		if v == false {
			t.Errorf("job %d did not run", i)
		}
	}
}

func TestTimeout(t *testing.T) {
	bunch := new(Harness)
	bunch.timeout = 14 * time.Second
	bunch.pauseFlush = true
	bunch.t = t
	bunch.record = make([]bool, 25)
	bunch.packets = list.New()

	cj := NewClusterJob("fu", time.Now().UTC().String(), 0, 0)
	jp := NewJobPacket().WithNexts("foo1", 15).WithClusterJobs(cj, nil)
	jp.AddJobs(makeJobs(20, 2*time.Second, 0))
	bunch.AddPending(jp.NumJobs())

	bunch.packets.PushFront(jp)
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

	for {
		e := bunch.packets.Front()
		if e != nil {
			packet := e.Value.(*JobPacket)
			bunch.part = packet.MyJob.Part
			if err := test.RunMaster(packet, "JobManagerARN"); err != nil {
				t.Errorf("RunMaster returned %v", err)
			}
			bunch.timeout = 30 * time.Second
			bunch.pauseFlush = false
			bunch.packets.Remove(e)
		} else {
			break
		}
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

	for i, v := range bunch.record {
		if v == false {
			t.Errorf("job %d did not run", i)
		}
	}
}
