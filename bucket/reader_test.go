package bucket

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

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
		bucket, prefix := ParsePath(test.Src)
		if bucket != test.Bucket {
			t.Errorf("Expected bucket name %s, received %v", test.Bucket, bucket)
		} else if prefix != test.Prefix {
			t.Errorf("Expected prefix %s, received %v", test.Prefix, prefix)
		}
	}
}

type testState int

const (
	SearchRoot testState = iota
	Root
	Disk1
	Disk1_foo1
	Disk1_foo2
	Disk1_bar1
	Disk1_bar2
	Disk2
	Disk2_foo1
	Disk2_foo2
	Disk2_bar1
	Disk2_bar2
)

func makeTestDir() *DirItem {
	paths := [][]string{
		{"disk1", "foo1", "bar1"},
		{"disk2", "foo1", "bar1"},
		{"disk1", "foo2", "bar2"},
		{"disk2", "foo2", "bar2"},
	}
	root := NewDirItem("/")
	for _, p := range paths {
		current := root
		for _, path := range p {
			if nd, ok := current.Contains(path); !ok {
				nd := NewDirItem(path)
				current.Add(nd)
				current = nd
			} else {
				current = nd
			}
		}
	}
	return root
}

func TestDirIterate(t *testing.T) {
	state := SearchRoot
	count := 0
	d1_count := 0
	d2_count := 0
	var next testState
	root := makeTestDir()
	iter := root.Iterate()
	for {
		nd, ok := iter.Next()
		if !ok {
			break
		} else {
			count++
		}
		switch state {
		case SearchRoot:
			if nd.Name == "/" {
				state = Root
			} else {
				t.Errorf("Expected /, received %s", nd.Name)
				break
			}
		case Root:
			switch nd.Name {
			case "disk1":
				state = Disk1
			case "disk2":
				state = Disk2
			default:
				t.Errorf("Expected disk1 or disk2, received %s", nd.Name)
				break
			}
		case Disk1:
			switch nd.Name {
			case "foo1":
				state = Disk1_bar1
				switch d1_count {
				case 0:
					d1_count++
					next = Disk1
				case 1:
					next = Root
				}
			case "foo2":
				state = Disk1_bar2
				switch d1_count {
				case 0:
					d1_count++
					next = Disk1
				case 1:
					next = Root
				}
			default:
				t.Errorf("Disk1: Expected foo1 or foo2, received %s", nd.Name)
				break
			}
		case Disk2:
			switch nd.Name {
			case "foo1":
				state = Disk2_bar1
				switch d2_count {
				case 0:
					d2_count++
					next = Disk2
				case 1:
					next = Root
				}
			case "foo2":
				state = Disk2_bar2
				switch d2_count {
				case 0:
					d2_count++
					next = Disk2
				case 1:
					next = Root
				}
			default:
				t.Errorf("Disk2: Expected foo1 or foo2, received %s", nd.Name)
				break
			}
		case Disk1_bar1:
			if nd.Name != "bar1" {
				t.Errorf("Disk1: Expected bar1, recieved %s", nd.Name)
				break
			}
			state = next
		case Disk1_bar2:
			if nd.Name != "bar2" {
				t.Errorf("Disk1: Expected bar2, recieved %s", nd.Name)
				break
			}
			state = next
		case Disk2_bar1:
			if nd.Name != "bar1" {
				t.Errorf("Disk2: Expected bar1, recieved %s", nd.Name)
				break
			}
			state = next
		case Disk2_bar2:
			if nd.Name != "bar2" {
				t.Errorf("Disk2: Expected bar2, recieved %s", nd.Name)
				break
			}
			state = next
		default:
			t.Errorf("Unexpected state %d, precessing %s", state, nd.Name)
			break
		}
	}
	if count != 11 {
		t.Errorf("Expected count 11, received %d", count)
	}
}

func TestDirIterateAbort(t *testing.T) {
	state := SearchRoot
	var expected testState
	count := 0
	d1_count := 0
	d2_count := 0
	foo1_count := 0
	foo2_count := 0
	var next testState
	root := makeTestDir()
	iter := root.Iterate()
	for {
		nd, ok := iter.Next()
		if !ok {
			break
		} else {
			count++
		}
		switch state {
		case SearchRoot:
			if nd.Name == "/" {
				state = Root
			} else {
				t.Errorf("Expected /, received %s", nd.Name)
				break
			}
		case Root:
			switch nd.Name {
			case "disk1":
				state = Disk1
				expected = Disk1
			case "disk2":
				state = Disk2
				expected = Disk2
			default:
				t.Errorf("Expected disk1 or disk2, received %s", nd.Name)
				break
			}
		case Disk1:
			if expected != Disk1 {
				t.Errorf("Expected Abort to skip Disk1")
				break
			}
			switch nd.Name {
			case "foo1":
				foo1_count++
				state = Disk1_bar1
				switch d1_count {
				case 0:
					d1_count++
					next = Disk1
				case 1:
					next = Root
				}
			case "foo2":
				foo1_count++
				state = Disk1_bar2
				switch d1_count {
				case 0:
					d1_count++
					next = Disk1
				case 1:
					next = Root
				}
			default:
				t.Errorf("Disk1: Expected foo1 or foo2, received %s", nd.Name)
				break
			}
		case Disk2:
			if expected != Disk2 {
				t.Errorf("Expected Abort to skip Disk2")
				break
			}
			switch nd.Name {
			case "foo1":
				foo2_count++
				state = Disk2_bar1
				switch d2_count {
				case 0:
					d2_count++
					next = Disk2
				case 1:
					next = Root
				}
			case "foo2":
				foo2_count++
				state = Disk2_bar2
				switch d2_count {
				case 0:
					d2_count++
					next = Disk2
				case 1:
					next = Root
				}
			default:
				t.Errorf("Disk2: Expected foo1 or foo2, received %s", nd.Name)
				break
			}
		case Disk1_bar1:
			if expected != Disk1 {
				t.Errorf("Expected Abort to skip Disk1")
				break
			}
			if nd.Name != "bar1" {
				t.Errorf("Disk1: Expected bar1, recieved %s", nd.Name)
				break
			}
			state = next
			if foo1_count > 1 {
				iter.Abort()
			}
		case Disk1_bar2:
			if expected != Disk1 {
				t.Errorf("Expected Abort to skip Disk1")
				break
			}
			if nd.Name != "bar2" {
				t.Errorf("Disk1: Expected bar2, recieved %s", nd.Name)
				break
			}
			state = next
			if foo1_count > 1 {
				iter.Abort()
			}
		case Disk2_bar1:
			if expected != Disk2 {
				t.Errorf("Expected Abort to skip Disk2")
				break
			}
			if nd.Name != "bar1" {
				t.Errorf("Disk2: Expected bar1, recieved %s", nd.Name)
				break
			}
			state = next
			if foo2_count > 1 {
				iter.Abort()
			}
		case Disk2_bar2:
			if expected != Disk2 {
				t.Errorf("Expected Abort to skip Disk2")
				break
			}
			if nd.Name != "bar2" {
				t.Errorf("Disk2: Expected bar2, recieved %s", nd.Name)
				break
			}
			state = next
			if foo2_count > 1 {
				iter.Abort()
			}
		default:
			t.Errorf("Unexpected state %d, precessing %s", state, nd.Name)
			break
		}
	}
	if count != 6 {
		t.Errorf("Expected count 6, received %d", count)
	}
}

// A simple buffer that supports ReadAt
type BufReaderAt struct {
	buf bytes.Buffer
}

func (br *BufReaderAt) ReadAt(p []byte, off int64) (int, error) {
	reqLen := int64(len(p))
	if br.buf.Len() >= int(off)+len(p) {
		copy(p, br.buf.Bytes()[int(off):int(off+reqLen)])
		return len(p), nil
	}
	return 0, fmt.Errorf("Request out of bounds, max: %d, req: %d", br.buf.Len(), int(off)+len(p))
}

func TestChunkReader(t *testing.T) {
	br := &BufReaderAt{}
	mySize := int64(512)

	br.buf.WriteString(strings.Repeat("0", int(mySize)))
	br.buf.WriteString(strings.Repeat("1", int(mySize)))
	br.buf.WriteString(strings.Repeat("2", int(mySize)))
	br.buf.WriteString(strings.Repeat("3", int(mySize)))

	cr := NewChunkReader(4*mySize, br)
	p := make([]byte, 10)
	n, err := cr.ReadAt(p, 0)
	if err != nil {
		t.Error(err)
	}
	if n < 10 {
		t.Errorf("Expected 10 bytes, received %d", n)
	}
	if string(p) != "0000000000" {
		t.Errorf("Expected 10 zero's, received %s", string(p))
	}

	n, err = cr.ReadAt(p, mySize*1)
	if err != nil {
		t.Error(err)
	}
	if n < 10 {
		t.Errorf("Expected 10 bytes, received %d", n)
	}
	if string(p) != "1111111111" {
		t.Errorf("Expected 10 1's, received %s", string(p))
	}

	n, err = cr.ReadAt(p, mySize*3)
	if err != nil {
		t.Error(err)
	}
	if n < 10 {
		t.Errorf("Expected 10 bytes, received %d", n)
	}
	if string(p) != "3333333333" {
		t.Errorf("Expected 10 3's, received %s", string(p))
	}

	n, err = cr.ReadAt(p, mySize*2)
	if err != nil {
		t.Error(err)
	}
	if n < 10 {
		t.Errorf("Expected 10 bytes, received %d", n)
	}
	if string(p) != "2222222222" {
		t.Errorf("Expected 10 2's, received %s", string(p))
	}

	p1 := make([]byte, 1024)
	n, err = cr.ReadAt(p1, mySize*2)
	if err != nil {
		t.Error(err)
	}
	if n < 1024 {
		t.Errorf("Expected 1024 bytes, received %d", n)
	}
	if string(p1)[0:10] != "2222222222" {
		t.Errorf("Expected 10 2's, received %s", string(p1)[0:10])
	}
	if string(p1)[1000:1010] != "3333333333" {
		t.Errorf("Expected 10 3's, received %s", string(p1)[1000:1010])
	}
}
