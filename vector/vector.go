// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vector

import (
	"errors"
	"io"
	"github.com/geodatalake/lambdas/geotiff"
)

type VectorStream interface {
	io.Reader
	io.ReaderAt
}

type VectorIntfc interface {

	IsVector()     	bool
	IsKML()        	bool
	IsShape()   	bool
	Bounds()	(*geotiff.Bounds, error)
	GetFileLength()	uint32
}

func IsVector(stream VectorStream) (VectorIntfc, error) {

	vInterface, err := IsVectorType( stream )

	if err == nil {
		return vInterface, nil
	}

	return vInterface, errors.New("Not a known vector type")
}
