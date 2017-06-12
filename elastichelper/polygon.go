// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elastichelper

import "strings"

// "POLYGON ((%.7f %7f, %.7f %7f, %.7f %7f, %.7f %7f, %.7f %7f))"
func ExtractPolygonPoints(s string) ([]string, bool) {
	if strings.HasPrefix(s, "POLYGON") {
		if start := strings.Index(s, "(("); start != -1 {
			if end := strings.Index(s, "))"); end != -1 {
				pairs := strings.Split(s[start+2:end], ",")
				points := make([]string, 0, len(pairs)*2)
				for _, pair := range pairs {
					p := strings.Split(strings.TrimSpace(pair), " ")
					for _, pt := range p {
						points = append(points, strings.TrimSpace(pt))
					}
				}
				return points, true
			}
		}
	}
	return []string{}, false
}
