// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elastichelper

import (
	"errors"
)

type Envelope struct {
	top, bottom *float64
	left, right *float64
}

func NewEnvelope(topLeftLat, topLeftLon, bottomRightLat, bottomRightLon float64) *Envelope {
	return &Envelope{top: &topLeftLat, bottom: &bottomRightLat, left: &topLeftLon, right: &bottomRightLon}
}

func (env *Envelope) Build() map[string]interface{} {
	if env.top == nil {
		panic(errors.New("Envelope.top has to be specified"))
	}
	if env.bottom == nil {
		panic(errors.New("Envelope.bottom has to be specified"))
	}
	if env.left == nil {
		panic(errors.New("Envelope.left has to be specified"))
	}
	if env.right == nil {
		panic(errors.New("Envelope.right has to be specified"))
	}
	return NewDoc().
		AddKV("type", "envelope").
		AddKV("coordinates", ComposeCoords(FormatCoords(*env.top, *env.left), FormatCoords(*env.bottom, *env.right))).Build()
}
