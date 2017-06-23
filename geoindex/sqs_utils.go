// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geoindex

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/geodatalake/lambdas/scale"
)

type SqsInstance struct {
	region string
	queue  string
	sess   *session.Session
	client *sqs.SQS
}

func NewSqsInstance() *SqsInstance {
	return new(SqsInstance)
}

func (c *SqsInstance) WithSession(sess *session.Session) *SqsInstance {
	c.sess = sess
	return c
}

func (c *SqsInstance) WithRegion(region string) *SqsInstance {
	c.region = region
	return c
}

func (c *SqsInstance) WithQueue(queue string) *SqsInstance {
	c.queue = queue
	return c
}

func (c *SqsInstance) create() error {
	if c.queue == "" {
		return fmt.Errorf("queue not specified")
	}
	if c.sess == nil {
		var err error
		c.sess, err = scale.GetAwsSession()
		if err != nil {
			return err
		} else {
			if c.region != "" {
				c.client = sqs.New(c.sess, aws.NewConfig().WithRegion(c.region))
			} else {
				c.client = sqs.New(c.sess)
			}
		}
	}
	return nil
}

func (c *SqsInstance) init() error {
	if c.client == nil {
		return c.create()
	}
	return nil
}

func (c *SqsInstance) SendAsJson(message string) (string, error) {
	data := make(map[string]interface{})
	data["Message"] = message
	b, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Error converting %s to json: %v", message, err)
	}
	return c.Send(string(b))
}

func (c *SqsInstance) Send(message string) (string, error) {
	if err := c.init(); err != nil {
		return "", fmt.Errorf("Unable to create client")
	} else {
		resp, err := c.client.SendMessage(new(sqs.SendMessageInput).
			SetQueueUrl(c.queue).
			SetMessageBody(message))
		if err != nil {
			return "", err
		}
		return aws.StringValue(resp.MessageId), nil
	}
}

func (c *SqsInstance) SendClusterRequest(cr *ClusterRequest) (string, error) {
	if b, err := json.Marshal(cr); err == nil {
		return c.Send(string(b))
	} else {
		return "", err
	}
}

func (c *SqsInstance) SendClusterResponse(cr *ClusterResponse) (string, error) {
	if b, err := json.Marshal(cr); err == nil {
		return c.Send(string(b))
	} else {
		return "", err
	}
}

func (c *SqsInstance) Delete(receiptHandle string) error {
	if err := c.init(); err != nil {
		return fmt.Errorf("Unable to create client")
	} else {
		_, err = c.client.DeleteMessage(new(sqs.DeleteMessageInput).
			SetQueueUrl(c.queue).
			SetReceiptHandle(receiptHandle))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *SqsInstance) Receive(num int64) (*sqs.ReceiveMessageOutput, error) {
	if err := c.init(); err != nil {
		return nil, fmt.Errorf("Unable to create client")
	} else {
		out, err := c.client.ReceiveMessage(new(sqs.ReceiveMessageInput).SetMaxNumberOfMessages(num).SetQueueUrl(c.queue).SetWaitTimeSeconds(3))
		if err != nil {
			return nil, err
		}
		return out, nil
	}
}
