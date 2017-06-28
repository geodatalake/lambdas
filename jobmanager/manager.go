// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jobmanager

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

func (jm *JobManager) CalcPackets(numJobs, maxJobs int) int {
	parts := numJobs / maxJobs
	if numJobs%maxJobs > 0 {
		parts++
	}
	log.Println(parts, "packets are required to send", numJobs, "jobs")
	return parts
}

func (jm *JobManager) createPackets(packet *JobPacket, startTime string) []*JobPacket {
	var retval []*JobPacket
	parts := jm.CalcPackets(packet.NumJobs(), packet.MaxNext)
	if parts == 1 {
		retval = make([]*JobPacket, 0, 1)
		retval = append(retval, packet)
	} else {
		retval = make([]*JobPacket, 0, parts)
		newId := uuid.NewV4().String()
		for part := 0; part < parts; part++ {
			pjob := NewClusterJob(newId, startTime, part, parts-1)
			cq := NewJobPacket().
				WithNexts(packet.Next, packet.MaxNext).
				WithClusterJobs(pjob, packet.MyJob)
			start := part * packet.MaxNext
			end := start + packet.MaxNext
			if end > packet.NumJobs() {
				end = packet.NumJobs()
			}
			cq.AddJobs(packet.ExportJobs(start, end))
			retval = append(retval, cq)
		}
	}
	return retval
}

func (jm *JobManager) queueJob(job interface{}, next string) (LambdaInvokeResponse, error) {
	if jm.driver != nil {
		// Some systems wrap the job with metadata, this call
		// allows that info to be stripped prior to sending
		actualJob := jm.checkJobHelper().GetActualJob(job)
		return jm.driver.InvokeRequestResponse(actualJob, next)
	} else {
		log.Println("Lambda driver has not been set on JobManager")
		return nil, fmt.Errorf("Lambda driver has not been set on JobManager")
	}
}

func (jm *JobManager) queuePacket(packet *JobPacket, next string) error {
	if jm.driver != nil {
		return jm.driver.InvokeAsync(packet, next)
	} else {
		log.Println("Lambda driver has not been set on JobManager")
		return fmt.Errorf("Lambda driver has not been set on JobManager")
	}
}

func (jm *JobManager) checkSync() SyncHelper {
	if jm.syncHelper == nil {
		log.Println("No syncHelper is installed, job will never start")
		panic("No syncHelper is installed, job will never start")
	}
	return jm.syncHelper
}

func (jm *JobManager) checkTimeout() JobTimeout {
	if jm.jobTimeout == nil {
		log.Println("No jobTimeout is installed, job will never start")
		panic("No jobTimeout is installed, job will never start")
	}
	return jm.jobTimeout
}

func (jm *JobManager) checkJobHelper() JobHelper {
	if jm.jobHelper == nil {
		log.Println("No jobHelper is installed, job will never start")
		panic("No jobHelper is installed, job will never start")
	}
	return jm.jobHelper
}

func (jm *JobManager) jobCharacteristics(packet *JobPacket, start, end int) (time.Duration, map[string]int) {
	return packet.JobCharacteristics(jm.checkJobHelper(), start, end)
}

func (jm *JobManager) CanRun(jp *JobPacket) bool {
	pn := jp.MyJob.Part

	return pn == 0 ||
		jm.checkSync().GetPartFor(jp.MyJob.Id) == pn
}

func (jm *JobManager) WaitForRun(jp *JobPacket) bool {
	jobDuration := jm.checkTimeout().Timeout()
	timeout := time.After(jobDuration - (30 * time.Second))
	tick := time.Tick(10 * time.Second)

	timedOut := false
DONE:
	for {
		select {
		case <-tick:
			if jm.CanRun(jp) {
				break DONE
			}
		case <-timeout:
			timedOut = true
			break DONE
		}
	}
	return !timedOut
}

func (jm *JobManager) RunMaster(packet *JobPacket, invokedFunctionARN string) error {
	if !jm.CanRun(packet) {
		log.Println("Can't run right now, waiting")
		if !jm.WaitForRun(packet) {
			log.Println("Timed out waiting for start of", packet.MyJob)
			jm.queuePacket(packet, invokedFunctionARN)
			return nil
		}
	}
	log.Println("Assuming master, starting processing a queue of", packet.NumJobs(), "jobs", packet.MaxNext, "jobs at a time")
	nq, timedout, complete := jm.SendQueue(packet)
	for {
		if complete {
			if nq.MySubJob != nil {
				if nq.MySubJob.IsLastPart() {
					jm.checkSync().RegisterJobComplete(nq.MySubJob.Id, nq.MySubJob.CalcDuration())
				} else {
					jm.checkSync().FinishPart(nq.MySubJob.Id, nq.MySubJob.Part)
				}
			} else {
				if nq.MyJob.IsLastPart() {
					jm.checkSync().RegisterJobComplete(nq.MyJob.Id, nq.MyJob.CalcDuration())
				} else {
					jm.checkSync().FinishPart(nq.MyJob.Id, nq.MyJob.Part)
				}
			}
			return nil
		}
		if timedout {
			if nq != nil {
				if nq.NumJobs() <= nq.MaxNext {
					// we can stay on this job
					log.Println("We can stay on this Packet")
					jm.queuePacket(nq, invokedFunctionARN)
				} else {
					// We need to spawn a sub-job because
					// the inital calcs are wrong and jobs have grown
					log.Println("Spawning new Packets")
					startTime := time.Now().UTC().String()
					parts := jm.createPackets(nq, startTime)
					for _, part := range parts {
						jm.queuePacket(part, invokedFunctionARN)
					}
				}
			}
			return nil
		} else if nq != nil {
			nq, timedout, complete = jm.SendQueue(nq)
		} else {
			return nil
		}
	}
}

func (jm *JobManager) SendQueue(packet *JobPacket) (nb *JobPacket, timeout, complete bool) {
	numToSend := packet.MaxNext
	if numToSend > packet.NumJobs() {
		numToSend = packet.NumJobs()
	}
	numLeft := packet.NumJobs() - numToSend
	if numLeft == 0 {
		numLeft = numToSend
	}

	maxTimeout, occureances := jm.jobCharacteristics(packet, 0, numToSend)

	timeLeft := jm.checkTimeout().Timeout()
	if timeLeft <= maxTimeout+(10*time.Second) {
		log.Println("job does not have enough time left to run", timeLeft, maxTimeout)
		return packet, true, packet.NumJobs() == 0
	}
	jobTimeout := time.After(timeLeft - (10 * time.Second))

	for k, v := range occureances {
		jm.checkSync().RegisterStart(k, v)
	}

	items := packet.ExportJobs(0, numToSend)
	nextPacket := packet.Clone()
	nj := packet.NumJobs()
	if nj > numToSend {
		nextPacket.AddJobs(packet.ExportJobs(numToSend, nj))
	}

	var wait sync.WaitGroup
	wait.Add(len(items))

	for _, item := range items {
		go jm.ExecJob(item, nextPacket, &wait)
	}

	waitDone := make(chan bool, 1)
	go func() {
		wait.Wait()
		waitDone <- true
	}()

	timedOut := false
	select {
	case <-waitDone:
		timedOut = false
		break
	case <-jobTimeout:
		timedOut = true
		break
	}
	jm.checkSync().FlushPending()
	for k, v := range occureances {
		jm.checkSync().RegisterStop(k, v)
	}

	return nextPacket, timedOut, nextPacket.NumJobs() == 0
}

func (jm *JobManager) ExecJob(job interface{}, nextPacket *JobPacket, wait *sync.WaitGroup) {
	if ok, out := jm.sendJob(job, nextPacket.Next, time.After(jm.checkJobHelper().GetTimeout(job))); !ok {
		jm.checkSync().AddPending(1)
		nextPacket.AddJob(job)
	} else if out != nil && len(out) > 0 {
		newJobs, err := jm.checkJobHelper().UnmarshalJobs(out)
		if err != nil {
			log.Println("Error decoding job return", err)
			jm.checkSync().AddPending(len(newJobs))
			nextPacket.AddJobs(newJobs)
		} else {
			if len(newJobs) > 0 {
				jm.checkSync().AddPending(len(newJobs))
				nextPacket.AddJobs(newJobs)
			}
		}
	}
	wait.Done()
}

func (jm *JobManager) sendJob(job interface{}, next string, timeout <-chan time.Time) (bool, []byte) {
	resultChan := make(chan LambdaInvokeResponse, 1)
	errChan := make(chan error, 1)
	go func() {
		resp, err := jm.queueJob(job, next)
		if err != nil {
			errChan <- err
		} else {
			resultChan <- resp
		}
	}()

	for {
		select {
		case io := <-resultChan:
			return true, io.Payload()
		case e := <-errChan:
			log.Println(e)
			return true, nil
		case <-timeout:
			return false, nil
		}
	}
}
