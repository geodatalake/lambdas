package scale

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

func FormatManifest(output []*OutputData, parsed []*ParseResult) *ResultsManifest {
	if parsed != nil {
		return &ResultsManifest{
			Version:      "1.1",
			OutputData:   output,
			ParseResults: parsed,
		}
	} else {
		return &ResultsManifest{
			Version:    "1.1",
			OutputData: output,
		}
	}
}

func RegisterJobType(url, token string, data []byte) {
	client := http.Client{}
	urlStr := fmt.Sprintf("%s/service/scale/api/v5/job-types/", url)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(data))
	req.Header.Add("Authorization", fmt.Sprintf("token=%s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		WriteStderr(fmt.Sprintf("Error registering job type: %v", err))
		os.Exit(-1)
	}
	if resp.StatusCode != 201 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			WriteStderr(err.Error())
			os.Exit(-1)
		}
		resp.Body.Close()
		fmt.Println(resp.Status, string(b))
	} else {
		resp.Body.Close()
		fmt.Println("Create Job Type Response:", resp.Status)
	}
}

func GetAwsSession() (*session.Session, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Unable to create initial session: %v", err)
	}
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		})
	sess, errSess := session.NewSession(&aws.Config{
		Credentials: creds,
	})
	if errSess != nil {
		return nil, fmt.Errorf("Unable to create a session with credentials: %v", errSess)
	}
	return sess, nil
}
