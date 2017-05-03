// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geotiff

import "testing"

func TestNameForKey(t *testing.T) {

	keys := []int{1024, 1025, 1026, 1027, 1028, 1029}
	answers := []string{"GTModelTypeGeoKey", "GTRasterTypeGeoKey", "GTCitationGeoKey", "1027", "1028", "1029"}

	for i, v := range keys {
		answer := NameForKey(v)
		if answer != answers[i] {
			t.Errorf("NameForKey(%v) yielded %s, expected %s", v, answer, answers[i])
		}
	}
}

func TestValueForKey(t *testing.T) {
	keys := []int{1024, 1025, 1026, 1027, 1028, 1029, 3077, 2048, 2050, 2051, 2052, 2054, 2056, 2060, 3074, 3075, 3076, 4096, 4099}
	offsets := []int{1, 2, 6, 43, 56, 63, 0, 4012, 6023, 8901, 9006, 9102, 7001, 9104, 11531, 3, 9010, 5105, 9001}
	answers := []interface{}{"ModelTypeProjected", "RasterPixelIsPoint", "Test6", 43, 56, 63, 123.456, "GCSE_Clarke1880_RGS", "DatumE_International1967",
		"PM_Greenwich", "Linear_Foot_Indian", "Angular_Degree", "Ellipse_Airy_1830", "Angular_Arc_Second", "Proj_Kansas_CS83_North", "CT_ObliqueMercator",
		"Linear_Chain_Benoit", "VertCS_Baltic_Sea", "Linear_Meter"}
	locations := map[int]int{1026: 34737, 3077: 34736}
	counts := map[int]int{1026: 5}

	doubles := []float64{123.456, 789.012}
	asciis := "Test1,Test6"

	for i, key := range keys {
		var count = 1
		var location = 0
		if c, ok := counts[key]; ok {
			count = c
		}
		if l, ok := locations[key]; ok {
			location = l
		}

		answer := ValueForKey(key, location, offsets[i], count, doubles, []byte(asciis))
		if answer != answers[i] {
			t.Errorf("ValueForKey(%v) yielded %v, expected %v", key, answer, answers[i])
		}
	}
}
