// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vector

import (
	"fmt"
	"io"
)

type VectorStream interface {
	io.Reader
	io.ReaderAt
}

func IsVector(stream VectorStream) (interface{}, error) {
	return nil, fmt.Errorf("Not yet implemented")
}
