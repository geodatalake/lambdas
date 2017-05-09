// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elastichelper

type ElasticPoint struct {
	lon, lat float64
}

func MakePoint(lat, lon float64) *ElasticPoint {
	return &ElasticPoint{lon: lon, lat: lat}
}

func (ep *ElasticPoint) Build() map[string]interface{} {
	return NewDoc().
		AddKV("type", "point").
		AddKV("coordinates", FormatCoords(ep.lat, ep.lon)).Build()
}
