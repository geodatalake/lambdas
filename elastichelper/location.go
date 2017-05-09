// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elastichelper

import (
	"fmt"
)

type LocationMapping struct {
	tree, precision string
}

type PrecisionUnit int

const (
	Inch PrecisionUnit = iota
	Yard
	Miles
	Kilometers
	Meters
	Centimeters
	Millimeters
)

type LocationTree int

const (
	GeoHash LocationTree = iota
	QuadTree
)

func NewLocationMapping() *LocationMapping {
	return new(LocationMapping).Tree(GeoHash).Precision(10, Meters)
}

func (lm *LocationMapping) Tree(tree LocationTree) *LocationMapping {
	switch tree {
	case GeoHash:
		lm.tree = "geohash"
	case QuadTree:
		lm.tree = "quadtree"
	}
	return lm
}

func (lm *LocationMapping) Precision(num int, unit PrecisionUnit) *LocationMapping {
	switch unit {
	case Inch:
		lm.precision = fmt.Sprintf("%din", num)
	case Yard:
		lm.precision = fmt.Sprintf("%dyd", num)
	case Miles:
		lm.precision = fmt.Sprintf("%dmi", num)
	case Kilometers:
		lm.precision = fmt.Sprintf("%dkm", num)
	case Meters:
		lm.precision = fmt.Sprintf("%dm", num)
	case Centimeters:
		lm.precision = fmt.Sprintf("%dcm", num)
	case Millimeters:
		lm.precision = fmt.Sprintf("%dmm", num)
	}
	return lm
}

func (lm *LocationMapping) Build() map[string]interface{} {
	return NewDoc().
		AddKV("type", "geo_shape").
		AddKV("tree", lm.tree).
		AddKV("precision", lm.precision).Build()
}
