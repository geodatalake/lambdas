// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elastichelper

func FormatCoords(lat, lon float64) []float64 {
	return []float64{lon, lat}
}

func ComposeCoords(coords ...[]float64) [][]float64 {
	retval := make([][]float64, 0, len(coords))
	for _, c := range coords {
		retval = append(retval, c)
	}
	return retval
}

func MakeBboxClockwisePolygon(topLeftLat, topLeftLon, bottomRightLat, bottomRightLon float64) [][][]float64 {
	retval := make([][][]float64, 0, 1)
	top, bottom := topLeftLat, bottomRightLat
	left, right := topLeftLon, bottomRightLon
	topLeft := FormatCoords(top, left)
	topRight := FormatCoords(top, right)
	bottomRight := FormatCoords(bottom, right)
	bottomLeft := FormatCoords(bottom, left)
	ring := ComposeCoords(topLeft, topRight, bottomRight, bottomLeft, topLeft)
	retval = append(retval, ring)
	return retval
}

func MakeEnvelope(topLeftLat, topLeftLon, bottomRightLat, bottomRightLon float64) [][]float64 {
	top, bottom := topLeftLat, bottomRightLat
	left, right := topLeftLon, bottomRightLon
	topLeft := FormatCoords(top, left)
	bottomRight := FormatCoords(bottom, right)
	return ComposeCoords(topLeft, bottomRight)
}

type Builder interface {
	Build() map[string]interface{}
}

type Document struct {
	props map[string]interface{}
}

func NewDoc() *Document {
	return &Document{props: make(map[string]interface{})}
}

func (o *Document) AddKV(name string, value interface{}) *Document {
	o.props[name] = value
	return o
}

func (o *Document) Append(name string, object Builder) *Document {
	o.props[name] = object.Build()
	return o
}

func (o *Document) Build() map[string]interface{} {
	return o.props
}
