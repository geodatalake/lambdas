// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geoindex

import (
	"fmt"
	"log"

	"github.com/geodatalake/lambdas/geotiff"
	"github.com/geodatalake/lambdas/lidar"
	"github.com/geodatalake/lambdas/raster"
	"github.com/geodatalake/lambdas/vector"
)

func handleRaster(r interface{}) (*geotiff.Bounds, error) {
	switch rtype := r.(type) {
	case geotiff.Tiff:
		if rtype.IsGeotiff() {
			// TODO: Turn Bounds into WGS84
			if s, err2 := rtype.DescribeGeokeys(); err2 != nil {
				log.Println(err2)
			} else {
				log.Println("Geokeys:")
				for _, line := range s {
					log.Println(line)
				}
			}
			return rtype.Bounds()
		}
	case lidar.Las:
		bounds := rtype.Bounds()
		// TODO: Turn Bounds into WGS84
		if rtype.IsWktCrs() {
			log.Println("WKTCrs", rtype.WktCrs())
		} else {
			log.Println("Geocrs:", rtype.GeotiffCrs())
		}
		return bounds, nil
	}
	return nil, fmt.Errorf("Unknown raster type %v", r)
}

func handleVector(v interface{}) (*geotiff.Bounds, error) {
	// TODO: Add vector types
	return nil, fmt.Errorf("TODO: Implement vector methods")
}

func DetectType(stream raster.RasterStream) (*geotiff.Bounds, error) {
	if r, err := raster.IsRaster(stream); err != nil {
		if v, err1 := vector.IsVector(stream); err1 != nil {
			return nil, fmt.Errorf("Unknown file type %v", err1)
		} else {
			return handleVector(v)
		}
	} else {
		return handleRaster(r)
	}
}
