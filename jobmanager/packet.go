package jobmanager

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type JobPacket struct {
	mutex    sync.RWMutex
	Jobs     []interface{} `json:"jobs"`
	MyJob    *ClusterJob   `json:"job"`
	MySubJob *ClusterJob   `json:"subJob"`
	MaxNext  int           `json:"maxNext"`
	Next     string        `json:"next"`
	PacketId int           `json:"id"`
}

var packetId int = -1
var packetCreateLock sync.RWMutex

func NewJobPacket() *JobPacket {
	packetCreateLock.Lock()
	defer packetCreateLock.Unlock()
	packetId++
	return &JobPacket{PacketId: packetId}
}

func (jp *JobPacket) String() string {
	jp.mutex.RLock()
	defer jp.mutex.RUnlock()
	return fmt.Sprintf("JobPacket{id: %d, jobs: %d, next: %s, maxNext: %d}", jp.PacketId, jp.NumJobs(), jp.Next, jp.MaxNext)
}

func (jp *JobPacket) WithNexts(next string, maxNexts int) *JobPacket {
	jp.Next = next
	jp.MaxNext = maxNexts
	return jp
}

func (jp *JobPacket) WithClusterJobs(job, subJob *ClusterJob) *JobPacket {
	jp.MyJob = job
	jp.MySubJob = subJob
	return jp
}

func (jp *JobPacket) NumJobs() int {
	jp.mutex.RLock()
	defer jp.mutex.RUnlock()
	return len(jp.Jobs)
}

func (jp *JobPacket) AddJob(job interface{}) {
	jp.mutex.Lock()
	defer jp.mutex.Unlock()
	jp.Jobs = append(jp.Jobs, job)
}

func (jp *JobPacket) AddJobs(jobs []interface{}) {
	jp.mutex.Lock()
	defer jp.mutex.Unlock()
	log.Println("JobPacket id:", jp.PacketId, "Adding", len(jobs), "jobs")
	jp.Jobs = append(jp.Jobs, jobs...)
}

func (jp *JobPacket) ExportJobs(start, end int) []interface{} {
	jp.mutex.Lock()
	defer jp.mutex.Unlock()
	retval := make([]interface{}, end-start)
	copy(retval, jp.Jobs[start:end])
	for i := start; i < end; i++ {
		jp.Jobs[i] = nil
	}
	return retval
}

func (jp *JobPacket) Clone() *JobPacket {
	return NewJobPacket().
		WithNexts(jp.Next, jp.MaxNext).
		WithClusterJobs(jp.MyJob, jp.MySubJob)
}

func (jp *JobPacket) Unmarshal(jh JobHelper, m map[string]interface{}) error {
	var err error
	if arr, ok := m["jobs"]; ok {
		if jobs, good := arr.([]interface{}); good {
			jp.Jobs, err = jh.DecodeJobs(jobs)
			if err != nil {
				return err
			}
		}
	}
	if j, ok := SubpropFor("job", m); ok {
		jp.MyJob = new(ClusterJob)
		if err = jp.MyJob.Unmarshal(j); err != nil {
			return err
		}
	}
	if j, ok := SubpropFor("subJob", m); ok {
		jp.MyJob = new(ClusterJob)
		if err = jp.MyJob.Unmarshal(j); err != nil {
			return err
		}
	}
	jp.Next = StringFor("next", "", m)
	if jp.Next == "" {
		return fmt.Errorf("next is missing")
	}
	jp.MaxNext = IntFor("maxNext", -1, m)
	if jp.MaxNext == -1 {
		return fmt.Errorf("maxNext is missing")
	}
	return nil
}

func (jp *JobPacket) JobCharacteristics(jh JobHelper, start, end int) (time.Duration, map[string]int) {
	jp.mutex.RLock()
	defer jp.mutex.RUnlock()

	retval := time.Duration(0)
	occurences := make(map[string]int)
	for _, j := range jp.Jobs[start:end] {
		typ := jh.GetType(j)
		occurences[typ] = occurences[typ] + 1
		timeout := jh.GetTimeout(j)
		if timeout > retval {
			retval = timeout
		}
	}
	return retval, occurences
}
