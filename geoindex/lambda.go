package geoindex

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/geodatalake/lambdas/scale"
)

type MonitorConn struct {
	provider *lambda.Lambda
	sess     *session.Session
	name     string
	region   string
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

func (mc *MonitorConn) InvokeIndexer(typ IndexerRequestType, name string, num int) (*IndexerResponse, error) {
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

	req := new(IndexerRequest)
	req.RequestType = typ
	req.Name = name
	req.Num = num
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
