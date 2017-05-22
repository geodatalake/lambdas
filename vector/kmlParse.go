package vector

import (
)



type Placemark struct {
	Name      string `xml:"name"`
	LineString struct {
			  Extrude string `xml:"extrude"`
			  Coordinates []string `xml:"coordinates"`
		  }
}

type LinearRing struct {
	Coordinates []string `xml:"coordinates"`
}

type LineString struct {
	Coordinates []string `xml:"coordinates"`
}

type Point struct {
	Coordinates []string `xml:"coordinates"`
}


type GroundOverlay struct {
	Icon      string `xml:"Icon>href"`
	LatLonBox struct {
		North string `xml:"north"`
		South string `xml:"south"`
		East string  `xml:"east"`
		West string  `xml:"west"`
	  } `xml:"LatLonBox"`
}
