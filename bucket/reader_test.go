package bucket

import "testing"

type TestCase struct {
	Src, Bucket, Prefix string
}

func TestParsePath(t *testing.T) {
	tests := []TestCase{
		{"s3://foo/bar", "foo", "bar"},
		{"s3://foo/bar/bah/fu", "foo", "bar/bah/fu"},
		{"s3://foo/", "foo", "/"},
		{"s3://foo", "foo", "/"},
		{"foo/bar", "foo", "bar"},
		{"foo/bar/bah/fu", "foo", "bar/bah/fu"},
		{"foo/", "foo", "/"},
		{"foo", "foo", "/"},
	}
	for _, test := range tests {
		bucket, prefix := parsePath(test.Src)
		if bucket != test.Bucket {
			t.Errorf("Expected bucket name %s, received %v", test.Bucket, bucket)
		} else if prefix != test.Prefix {
			t.Errorf("Expected prefix %s, received %v", test.Prefix, prefix)
		}
	}
}
