package scale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/geodatalake/lambdas/elastichelper"
)

func WriteStderr(s string) {
	os.Stderr.Write([]byte(s))
}

func WriteJson(writer io.Writer, objectToWrite interface{}) {
	if jErr := json.NewEncoder(writer).Encode(objectToWrite); jErr != nil {
		WriteStderr(fmt.Sprintf("Error wrinting JSON: %v", jErr))
		os.Exit(30)
	}
}

func doc() *elastichelper.Document {
	return elastichelper.NewDoc()
}

func array() *elastichelper.ArrayBuilder {
	return elastichelper.StartArray()
}

func CreateScaleError(url, token string, data map[string]interface{}) {
	if data == nil {
		return
	}
	client := http.Client{}
	urlStr := fmt.Sprintf("%s/service/scale/api/v5/errors/", url)
	var b []byte
	var err error
	if b, err = json.Marshal(data); err != nil {
		WriteStderr(err.Error())
		os.Exit(-1)
	}
	req, err1 := http.NewRequest("POST", urlStr, bytes.NewBuffer(b))
	if err1 != nil {
		WriteStderr(err1.Error())
		os.Exit(-1)
	}
	req.Header.Add("Authorization", fmt.Sprintf("token=%s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	resp, err2 := client.Do(req)
	if err2 != nil {
		WriteStderr(err2.Error())
		os.Exit(-1)
	}
	resp.Body.Close()
	fmt.Println("Create New Error", data["name"], resp.Status)
}

func ErrorDoc(name, title, description string, existing map[string]int) map[string]interface{} {
	if _, ok := existing[name]; !ok {
		return doc().
			AddKV("name", name).
			AddKV("title", title).
			AddKV("description", description).
			AddKV("category", "ALGORITHM").Build()
	}
	return nil
}

type ExistingError struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	IsBuiltin    bool   `json:"is_builtin"`
	Created      string `json:"created"`
	LastModified string `json:"last_modified"`
}

type AllExistingErrors struct {
	Count   int             `json:"count"`
	Next    string          `json:"next,omitempty"`
	Prev    string          `json:"prev,omitempty"`
	Results []ExistingError `json:"results"`
}

func GatherExistingErrors(url, token string) map[string]int {
	retval := make(map[string]int)
	client := http.Client{}
	urlStr := fmt.Sprintf("%s/service/scale/api/v5/errors/", url)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		WriteStderr(err.Error())
		os.Exit(-1)
	}
	req.Header.Add("Authorization", fmt.Sprintf("token=%s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept", "application/json")
	fmt.Println("Url", urlStr, "Headers", req.Header)
	resp, err1 := client.Do(req)
	if err1 != nil {
		WriteStderr(err1.Error())
		os.Exit(-1)
	}
	b, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		WriteStderr(err2.Error())
		os.Exit(-1)
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		allErrors := new(AllExistingErrors)
		if errJson := json.Unmarshal(b, allErrors); errJson != nil {
			WriteStderr(errJson.Error())
			fmt.Println(" Response:", string(b))
			os.Exit(-1)
		}
		for _, existing := range allErrors.Results {
			retval[existing.Name] = existing.Id
		}
		return retval
	} else {
		fmt.Println(" Response:", string(b))
		os.Exit(-1)
	}
	return nil
}
