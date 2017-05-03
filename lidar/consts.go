// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lidar

const (
	headerSizePosition = 94
	geotiffSignature   = "LASF_Projection\x00"
	laszipSignature    = "laszip encoded\x00\x00"

	RGeoKeys    = 34735
	RGeoDoubles = 34736
	RGeoAscii   = 34737

	MathTransformWKT    = 2111
	CoordinateSystemWKT = 2112
)

type GpsTimeType int

// GlobalEncoding Bits
const (
	geGpsMask             = 0x01
	geWaveformMask        = 0x06
	geSyntheticReturnMask = 0x08
	geWktMask             = 0x10

	geGpsWeekTime       GpsTimeType = 0x00
	geGpsStandardOffset GpsTimeType = 0x01

	geWfInline   = 0x02
	geWfExternal = 0x04

	geSrFalse = 0x00
	geSrTrue  = 0x08

	geGeotiff = 0x00
	geWkt     = 0x10
)

// Pre version 1.4 Classifications
type classificationLegacy int

const (
	clCreatedNeverClassified classificationLegacy = iota
	clUnclassified
	clGround
	clLowVegitation
	clMediumVegitation
	clHighVegititation
	clBuilding
	clLowPointNoise
	clModelKeyPointMassPoint
	clWater
	clReserved10
	clReserved11
	clOverlapPoints
)

var legacyLookup = map[classificationLegacy]string{
	clCreatedNeverClassified: "Never Classified",
	clUnclassified:           "Unclassified",
	clGround:                 "Ground",
	clLowVegitation:          "Low Vegitation",
	clMediumVegitation:       "Medium Vegitation",
	clHighVegititation:       "High Vegitiation",
	clBuilding:               "Building",
	clLowPointNoise:          "Low Point (Noise)",
	clModelKeyPointMassPoint: "Model Key Point",
	clWater:                  "Water",
	clReserved10:             "Reserved10",
	clReserved11:             "Reserved11",
	clOverlapPoints:          "Overlap Points",
}

// Version 1.4 Classifications
type classification14 int

const (
	cCreatedNeverClassified classification14 = iota
	cUnclassified
	cGround
	cLowVegitation
	cMediumVegitation
	cHighVegititation
	cBuilding
	cLowPointNoise
	cReserved8
	cWater
	cRail
	cRoadSurface
	cReserved12
	cWireGuard
	cWireConductor
	cTransmissionTower
	cWireStructure
	cBridgeDeck
	cHighNoise
)

var classificationLookup = map[classification14]string{
	cCreatedNeverClassified: "Never Classified",
	cUnclassified:           "Unclassified",
	cGround:                 "Ground",
	cLowVegitation:          "Low Vegitation",
	cMediumVegitation:       "Medium Vegitation",
	cHighVegititation:       "High Vegitiation",
	cBuilding:               "Building",
	cLowPointNoise:          "Low Point (Noise)",
	cReserved8:              "Reserved8",
	cWater:                  "Water",
	cRail:                   "Rail",
	cRoadSurface:            "Road",
	cReserved12:             "Reserved12",
	cWireGuard:              "Wire - Guard (Shield)",
	cWireConductor:          "Wire – Conductor (Phase)",
	cTransmissionTower:      "Transmission Tower",
	cWireStructure:          "Wire-structure Connector (e.g. Insulator)",
	cBridgeDeck:             "Bridge Deck",
	cHighNoise:              "High Noise",
}
