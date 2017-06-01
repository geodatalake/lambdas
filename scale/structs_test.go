package scale

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWorkspace(t *testing.T) {
	w := NewS3Workspace("us-east-1", "foo")
	if w.Version != "1.0" {
		t.Errorf("Expected version 1.0, received %v", w.Version)
	}
	b, err := json.Marshal(w)
	if err != nil {
		t.Errorf("Error Marshalling workspace: %v", err)
	} else {
		w2 := new(S3Workspace)
		err2 := json.Unmarshal(b, w2)
		if err2 != nil {
			t.Errorf("Error Unmarshalling workspace: %v", err2)
		}
		if w.Version != w2.Version {
			t.Errorf("Expected version %v, received %v", w.Version, w2.Version)
		}
		if *w.Broker != *w2.Broker {
			t.Errorf("Expected %+v, received %+v", *w, *w2)
		}
	}
}

func TestManifestMarshal(t *testing.T) {
	output := &OutputFile{Path: "bar"}
	manifest := FormatManifestFile("foo", output, nil)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(manifest); err != nil {
		t.Error(err)
	}
}
