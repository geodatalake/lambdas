// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geotiff

import "fmt"

const (
	GeoKeyModelType      = 1024
	GeoKeyRasterType     = 1025
	GeoKeyGeographicType = 2048
)

func NameForKey(key int) string {
	name, ok := GeoKeys[key]
	if ok {
		return name
	}
	return fmt.Sprintf("%v", key)
}

func ValueForKey(key int, location int, valueOffset int, count int, doubles []float64, asciis []byte) interface{} {
	if location == tGeoDoubles {
		if valueOffset < len(doubles) {
			return doubles[valueOffset]
		} else {
			return valueOffset
		}
	} else if location == tGeoAscii {
		return string(asciis[valueOffset : valueOffset+count])
	}
	switch key {
	case 1024:
		return GTModelTypeGeoKey[valueOffset]
	case 1025:
		return GTRasterTypeGeoKey[valueOffset]
	case 2048:
		return GeographicTypeGeoKey[valueOffset]
	case 2050:
		return GeogGeodeticDatumGeoKey[valueOffset]
	case 2051:
		return GeogPrimeMeridianGeoKey[valueOffset]
	case 2052:
		return GeogLinearUnitsGeoKey[valueOffset]
	case 2054:
		return GeogAngularUnitsGeoKey[valueOffset]
	case 2056:
		return GeogEllipsoidGeoKey[valueOffset]
	case 2060:
		return GeogAzimuthUnitsGeoKey[valueOffset]
	case 3072:
		return ProjectionCSTypeGeoKey[valueOffset]
	case 3074:
		return ProjectionGeoKey[valueOffset]
	case 3075:
		return ProjCoordTransGeoKey[valueOffset]
	case 3076:
		return GeogLinearUnitsGeoKey[valueOffset]
	case 4096:
		return VerticalCSTypeGeoKey[valueOffset]
	case 4099:
		return GeogLinearUnitsGeoKey[valueOffset]
	default:
		return valueOffset
	}
}
