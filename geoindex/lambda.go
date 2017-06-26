package geoindex

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/geodatalake/lambdas/jobmanager"
	"github.com/geodatalake/lambdas/scale"
)

type MonitorConn struct {
	provider *lambda.Lambda
	sess     *session.Session
	name     string
	region   string
	mutex    sync.RWMutex
	pending  int
}

func NewMonitorConn() *MonitorConn {
	return new(MonitorConn).WithRegion("us-east-1")
}

func (mc *MonitorConn) WithRegion(region string) *MonitorConn {
	mc.region = region
	return mc
}

func (mc *MonitorConn) WithSession(session *session.Session) *MonitorConn {
	mc.sess = session
	return mc
}

func (mc *MonitorConn) WithFunctionArn(name string) *MonitorConn {
	mc.name = name
	return mc
}

func (mc *MonitorConn) Invoke(req *IndexerRequest) (*IndexerResponse, error) {
	if mc.provider == nil {
		if mc.name == "" {
			return nil, fmt.Errorf("Missing name in MonitorConn")
		}
		if mc.sess == nil {
			if sess, err := scale.GetAwsSession(); err == nil {
				mc.sess = sess
			} else {
				return nil, err
			}
		}
		mc.provider = lambda.New(mc.sess, aws.NewConfig().WithRegion(mc.region))
	}
	b, _ := json.Marshal(req)
	out, err := mc.provider.Invoke(new(lambda.InvokeInput).
		SetFunctionName(mc.name).
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

func (mc *MonitorConn) SendJobDone(name string, dur time.Duration) (*IndexerResponse, error) {
	req := new(IndexerRequest)
	req.RequestType = JobComplete
	req.Name = name
	req.Duration = dur
	return mc.Invoke(req)
}

func (mc *MonitorConn) GetPartFor(name string) int {
	req := new(IndexerRequest)
	req.RequestType = ActivePart
	req.Name = name
	result, err := mc.Invoke(req)
	if err == nil && result.Success {
		return result.Num
	} else {
		log.Println("Got an error from monitor", err)
	}
	return -1
}

func (mc *MonitorConn) RegisterStart(name string, num int) {
	req := new(IndexerRequest)
	req.RequestType = Enter
	req.Name = name
	req.Num = num
	mc.Invoke(req)
}

func (mc *MonitorConn) RegisterStop(name string, num int) {
	req := new(IndexerRequest)
	req.RequestType = Leave
	req.Name = name
	req.Num = num
	mc.Invoke(req)
}

func (mc *MonitorConn) RegisterJobComplete(name string, dur time.Duration) {
	req := new(IndexerRequest)
	req.RequestType = JobComplete
	req.Name = name
	req.Duration = dur
	mc.Invoke(req)
}

func (mc *MonitorConn) FinishPart(name string, part int) {
	mc.InvokeIndexer(PartComplete, name, part)
}

func (mc *MonitorConn) AddPending(num int) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.pending += num
}

func (mc *MonitorConn) FlushPending() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	if mc.pending > 0 {
		mc.InvokeIndexer(Enter, JPending, mc.pending)
		mc.pending = 0
	}
}
func (mc *MonitorConn) InvokeIndexer(typ IndexerRequestType, name string, num int) (*IndexerResponse, error) {
	req := new(IndexerRequest)
	req.RequestType = typ
	req.Name = name
	req.Num = num
	return mc.Invoke(req)
}

type LambdaHelper struct {
	sess   *session.Session
	region string
	driver *lambda.Lambda
}

func NewLamabdaHelper() *LambdaHelper {
	return new(LambdaHelper).WithRegion("us-west-2")
}

func (lh *LambdaHelper) WithSession(sess *session.Session) *LambdaHelper {
	lh.sess = sess
	return lh
}

func (lh *LambdaHelper) WithRegion(r string) *LambdaHelper {
	lh.region = r
	return lh
}

func (lh *LambdaHelper) build() error {
	var err error
	if lh.sess == nil {
		lh.sess, err = scale.GetAwsSession()
		if err != nil {
			return err
		}
	}
	lh.driver = lambda.New(lh.sess, aws.NewConfig().WithRegion(lh.region))
	return nil
}

type LambdaResponse struct {
	payload []byte
}

func (lr *LambdaResponse) Payload() []byte {
	return lr.payload
}

func (lh *LambdaHelper) InvokeRequestResponse(job interface{}, next string) (jobmanager.LambdaInvokeResponse, error) {
	if lh.driver == nil {
		if err := lh.build(); err != nil {
			return nil, err
		}
	}
	if cr, ok := job.(*ClusterResponse); ok {
		if data, errJson := json.Marshal(cr.Item); errJson == nil {
			invoke := new(lambda.InvokeInput).
				SetFunctionName(next).
				SetInvocationType(lambda.InvocationTypeRequestResponse).
				SetPayload(data)
			if lout, err := lh.driver.Invoke(invoke); err != nil {
				return nil, err
			} else {
				return &LambdaResponse{payload: lout.Payload}, nil
			}
		} else {
			return nil, errJson
		}
	} else if req, good := job.(*ClusterRequest); good {
		if data, errJson := json.Marshal(req); errJson == nil {
			invoke := new(lambda.InvokeInput).
				SetFunctionName(next).
				SetInvocationType(lambda.InvocationTypeRequestResponse).
				SetPayload(data)
			if lout, err := lh.driver.Invoke(invoke); err != nil {
				return nil, err
			} else {
				return &LambdaResponse{payload: lout.Payload}, nil
			}
		} else {
			return nil, errJson
		}
	}
	log.Println("Expected a ClusterResponse or ClusterRequest, received a", reflect.TypeOf(job).String())
	return nil, fmt.Errorf("Unexpected format")
}

func (lh *LambdaHelper) InvokeAsync(packet *jobmanager.JobPacket, next string) error {
	if lh.driver == nil {
		if err := lh.build(); err != nil {
			return err
		}
	}
	req := new(ClusterRequest)
	req.RequestType = ClusterMaster
	req.Packet = packet
	if data, errJson := json.Marshal(req); errJson == nil {
		invoke := new(lambda.InvokeInput).
			SetFunctionName(next).
			SetInvocationType(lambda.InvocationTypeEvent).
			SetPayload(data)
		if _, err := lh.driver.Invoke(invoke); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return errJson
	}
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

func AlertCallNext(cr *ClusterRequest, name string) (chan *lambda.InvokeOutput, chan error) {
	resultChan := make(chan *lambda.InvokeOutput, 1)
	errChan := make(chan error, 1)
	go func() {
		out, err := callNext(cr, name, lambda.InvocationTypeRequestResponse)
		if err != nil {
			errChan <- err
		} else {
			resultChan <- out
		}
	}()
	return resultChan, errChan
}
