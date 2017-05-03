// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geotiff

import "testing"

func TestOptToCompression(t *testing.T) {
	vals := []CompressionType{Deflate, Uncompressed}
	answers := []uint32{cDeflate, cNone}

	for i, v := range vals {
		answer := CompressionType.optToCompression(v)
		if answer != answers[i] {
			t.Errorf("compression type %v, returned %v, expected %v", Deflate, answer, answers[i])
		}
	}
}
