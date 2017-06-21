package geoindex

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/geodatalake/lambdas/scale"
)

func (ct *ContractTracker) InvokeIndexer(typ IndexerRequestType, name string, num int) (*IndexerResponse, error) {
	if ct.provider == nil {
		if ct.endpoint == "" {
			return nil, fmt.Errorf("Missing monitor in environment")
		}
		if sess, err := scale.GetAwsSession(); err == nil {
			ct.provider = lambda.New(sess, aws.NewConfig().WithRegion("us-west-2"))
		} else {
			return nil, err
		}
	}

	req := new(IndexerRequest)
	req.RequestType = typ
	req.Name = name
	req.Num = num
	b, _ := json.Marshal(req)
	out, err := ct.provider.Invoke(new(lambda.InvokeInput).
		SetFunctionName(ct.endpoint).
		SetInvocationType(lambda.InvocationTypeRequestResponse).
		SetPayload(b))
	if err != nil {
		log.Println("Error Invoking indexer", err)
		return new(IndexerResponse), err
	}
	retval := new(IndexerResponse)
	if jsonErr := json.Unmarshal(out.Payload, retval); jsonErr != nil {
		return retval, jsonErr
	}
	return retval, nil
}

func callNext(cr *ClusterRequest, name, invokeType string) (*lambda.InvokeOutput, error) {
	if sess, err := scale.GetAwsSession(); err == nil {
		l := lambda.New(sess, aws.NewConfig().WithRegion("us-west-2"))
		if data, errJson := json.Marshal(cr); errJson == nil {
			invoke := new(lambda.InvokeInput).
				SetFunctionName(name).
				SetInvocationType(invokeType).
				SetPayload(data)
			return l.Invoke(invoke)
		} else {
			return nil, errJson
		}
	} else {
		return nil, err
	}
}

func AwaitCallNext(cr *ClusterRequest, name string) (*lambda.InvokeOutput, error) {
	return callNext(cr, name, lambda.InvocationTypeRequestResponse)
}

func AsyncCallNext(cr *ClusterRequest, name string) (*lambda.InvokeOutput, error) {
	return callNext(cr, name, lambda.InvocationTypeEvent)
}

func PoolCallNexts(calls []*ClusterRequest, name string, contractFor *ContractFor) ([]*lambda.InvokeOutput, error) {
	retval := make([]*lambda.InvokeOutput, 0, len(calls))
	log.Println("Making", len(calls), "Lambda Invocations")
	errc := make(chan error, len(calls))
	done := make(chan *lambda.InvokeOutput)
	if contractFor != nil {
		timeout := time.After(40 * time.Second)
		total := len(calls)
	GIVEUP:
		for {
			if cnt := contractFor.ReserveManyWait(total); cnt == total {
				break
			} else {
				total = total - cnt
				log.Println("Timeout waiting to send", len(calls), "Reserved so far", len(calls)-total)
			}
			select {
			case <-timeout:
				log.Println("Giving up and sending")
				break GIVEUP
			default:
				log.Println("Retry")
			}
		}
	}
	for _, cr := range calls {
		go func(c *ClusterRequest) {
			out, err := AsyncCallNext(c, name)
			if err != nil {
				errc <- err
			} else {
				done <- out
			}
		}(cr)
	}
	total := 0
	timeExpired := false
	var lastErr error
	timeout := time.After(20 * time.Second)
DONE:
	for {
		select {
		case out := <-done:
			retval = append(retval, out)
			total++
			if total >= len(calls) {
				break DONE
			}
		case e := <-errc:
			lastErr = e
			total++
			if total >= len(calls) {
				break DONE
			}
		case <-timeout:
			timeExpired = true
			break DONE
		}
	}
	if timeExpired {
		return nil, fmt.Errorf("Timeout expired after %d executions", total)
	}
	return retval, lastErr
}

type ChunkHandler struct {
	chunk []*ClusterRequest
}

func NewChunkHandler(sz int) *ChunkHandler {
	if sz <= 0 {
		sz = 10
	}
	return &ChunkHandler{chunk: make([]*ClusterRequest, 0, sz)}
}

func (ch *ChunkHandler) String() string {
	return fmt.Sprintf("%d requests", len(ch.chunk))
}

func (ch *ChunkHandler) Add(cr *ClusterRequest) bool {
	ch.chunk = append(ch.chunk, cr)
	return ch.Full()
}

func (ch *ChunkHandler) Full() bool {
	return len(ch.chunk) == cap(ch.chunk)
}

func (ch *ChunkHandler) Empty() bool {
	return len(ch.chunk) == 0
}

func (ch *ChunkHandler) Clear() {
	// to avoid memory leaks, we have to clear the slice
	// simply resetting the pointers to zero does not free
	// the pointers
	for i := 0; i < len(ch.chunk); i++ {
		ch.chunk[i] = nil
	}
	ch.chunk = ch.chunk[:0]
}

func (ch *ChunkHandler) Send(name string, contractFor *ContractFor) ([]*lambda.InvokeOutput, error) {
	retval, err := PoolCallNexts(ch.chunk, name, contractFor)
	ch.Clear()
	return retval, err
}
