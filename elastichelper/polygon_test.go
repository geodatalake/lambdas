// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elastichelper

import "testing"

func TestExtractPolygonPoints(t *testing.T) {
	test := "POLYGON ((a b,c d , e f, g h,i j))"
	test2 := "POLYGON(( a b, c d , e f, g h,i j ))"
	for _, testString := range []string{test, test2} {
		results, ok := ExtractPolygonPoints(testString)
		if !ok {
			t.Errorf("Not a polygon string %s", testString)
		} else {
			if len(results) != 10 {
				t.Errorf("Expected 10 points, received %d", len(results))
			} else {
				expects := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
				for i, expect := range expects {
					if results[i] != expect {
						t.Errorf("Expected position %d '%s', received '%s'", i, expect, results[i])
					}
				}
			}
		}
	}
}
