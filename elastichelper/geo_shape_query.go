// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elastichelper

import (
	"errors"
)

type GeoShapeQuery struct {
	name        string
	top, bottom *float64
	left, right *float64
}

func NewBboxShapeQuery(name string, top, left, bottom, right float64) *GeoShapeQuery {
	return &GeoShapeQuery{
		name:   name,
		top:    &top,
		left:   &left,
		bottom: &bottom,
		right:  &right,
	}
}

func (gsq *GeoShapeQuery) Source() (interface{}, error) {
	if gsq.top == nil {
		return nil, errors.New("geo_shape requires top latitude to be set")
	}
	if gsq.bottom == nil {
		return nil, errors.New("geo_shape requires bottom latitude to be set")
	}
	if gsq.right == nil {
		return nil, errors.New("geo_shape requires right longitude to be set")
	}
	if gsq.left == nil {
		return nil, errors.New("geo_shape requires left longitude to be set")
	}
	source := NewDoc().
		Append("geo_shape", NewDoc().
			Append(gsq.name, NewDoc().
				Append("shape", NewDoc().
					AddKV("type", "envelope").
					AddKV("coordinates", ComposeCoords(FormatCoords(*gsq.top, *gsq.left), FormatCoords(*gsq.bottom, *gsq.right))))))

	return source.Build(), nil
}
