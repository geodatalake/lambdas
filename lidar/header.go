// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lidar

import (
	"fmt"
	"time"

	"github.com/geodatalake/lambdas/geotiff"
)

type Vlr struct {
	reserved          uint16
	userID            string
	recordID          uint16
	lengthAfterHeader uint16
	description       string
	data              []byte
}

type Evlr struct {
	reserved          uint16
	userID            string
	recordID          uint16
	lengthAfterHeader uint64
	description       string
	data              []byte
}

type ImageInfo struct {
	width, height                      float64
	scalex, scaley                     float64
	offsetx, offsety                   float64
	minx, miny, maxx, maxy, minz, maxz float64
}

type HeaderFormat interface {
	GetNumberOfPoints() uint64
	ScalePoints(int32, int32, int32) (float64, float64, float64)
	GetPointFormat() byte
	GetPointsOffset() uint64
	GetPointLength() uint16
	GetNumberOfVLR() uint32
	GetNumberOfEVLR() uint32
	GetOffsetOfEVLR() uint64
	GpsTime() GpsTimeType
	DateTime() (time.Time, error)
	isWkt() bool
	Imageinfo() *ImageInfo
	DumpHeader() []string
	Bounds() *geotiff.Bounds
	Version() (int, int)
	VersionString() string
}

type LasHeaderLegacy struct {
	fileSignature              string
	fileSourceId               uint16
	globalEncoding             uint16
	projectIdGuid              string
	versionMajor               byte
	versionMinor               byte
	systemIdentifier           string
	generatingSoftware         string
	fileCreationDayOfYear      uint16
	fileCreationYear           uint16
	headerSize                 uint16
	offsetDataPoint            uint32
	numVarLengthRecords        uint32
	pointDataRecordFormat      byte
	pointDataRecordLength      uint16
	legacyNumberPointRecords   uint32
	legacyNumberPointsByReturn []uint32
	xScaleFactor               float64
	yScaleFactor               float64
	zScaleFactor               float64
	xOffset                    float64
	yOffset                    float64
	zOffset                    float64
	maxX                       float64
	minX                       float64
	maxY                       float64
	minY                       float64
	maxZ                       float64
	minZ                       float64
}

func (h *LasHeaderLegacy) DateTime() (time.Time, error) {
	day := int(h.fileCreationDayOfYear)
	year := int(h.fileCreationYear)
	if tm, err := time.Parse("1-2-2006", fmt.Sprintf("1-1-%d", year)); err == nil {
		return tm.Add(time.Duration(time.Hour * time.Duration(24*(day-1)))), nil
	} else {
		return time.Now(), err
	}
}

func (h *LasHeaderLegacy) Bounds() *geotiff.Bounds {
	// assume orientation upper left origin
	return &geotiff.Bounds{MaxX: h.maxX, MinX: h.minX, MaxY: h.maxY, MinY: h.minY, OriginX: h.minX, OriginY: h.maxY}
}

func (h *LasHeaderLegacy) GetNumberOfVLR() uint32 {
	return h.numVarLengthRecords
}

func (h *LasHeaderLegacy) GetNumberOfEVLR() uint32 {
	return 0
}

func (h *LasHeaderLegacy) GetOffsetOfEVLR() uint64 {
	return 0
}

func (h *LasHeaderLegacy) GetNumberOfPoints() uint64 {
	return uint64(h.legacyNumberPointRecords)
}

func (h *LasHeaderLegacy) GetPointFormat() byte {
	return h.pointDataRecordFormat
}

func (h *LasHeaderLegacy) ScalePoints(x int32, y int32, z int32) (float64, float64, float64) {
	return float64((float64(x) * h.xScaleFactor) + h.xOffset), float64((float64(y) * h.yScaleFactor) + h.yOffset), float64((float64(z) * h.zScaleFactor) + h.zOffset)
}

func (h *LasHeaderLegacy) GetPointsOffset() uint64 {
	return uint64(h.offsetDataPoint)
}

func (h *LasHeaderLegacy) GetPointLength() uint16 {
	return h.pointDataRecordLength
}

func (h *LasHeaderLegacy) GpsTime() GpsTimeType {
	if (int(h.globalEncoding) & geGpsMask) == 0 {
		return geGpsWeekTime
	} else {
		return geGpsStandardOffset
	}
}

func (h *LasHeaderLegacy) isWkt() bool {
	return int(h.globalEncoding)&geWktMask == geWkt
}

func (h *LasHeaderLegacy) Imageinfo() *ImageInfo {
	return &ImageInfo{
		width:   h.maxX - h.minX,
		height:  h.maxY - h.minY,
		scalex:  h.xScaleFactor,
		offsetx: h.xOffset,
		scaley:  h.yScaleFactor,
		offsety: h.yOffset,
		minx:    h.minX,
		maxx:    h.maxX,
		miny:    h.minY,
		maxy:    h.maxY,
		minz:    h.minZ,
		maxz:    h.maxZ}
}

func (h *LasHeaderLegacy) Version() (int, int) {
	return int(h.versionMajor), int(h.versionMinor)
}

func (h *LasHeaderLegacy) VersionString() string {
	return fmt.Sprintf("%v.%v", h.versionMajor, h.versionMinor)
}

func (h *LasHeaderLegacy) DumpHeader() []string {
	retval := make([]string, 0, 28)
	retval = append(retval, fmt.Sprintf("fileSignature:                %s", h.fileSignature))
	retval = append(retval, fmt.Sprintf("fileSourceId:                 %v", h.fileSourceId))
	retval = append(retval, fmt.Sprintf("globalEncoding:               %v", h.globalEncoding))
	retval = append(retval, fmt.Sprintf("projectIdGuid:                %s", h.projectIdGuid))
	retval = append(retval, fmt.Sprintf("version:                      %v.%v", h.versionMajor, h.versionMinor))
	retval = append(retval, fmt.Sprintf("systemIdentifier:             %s", h.systemIdentifier))
	retval = append(retval, fmt.Sprintf("generatingSoftware:           %s", h.generatingSoftware))
	retval = append(retval, fmt.Sprintf("fileCreationDay/Yr:           %d/%d", h.fileCreationDayOfYear, h.fileCreationYear))
	retval = append(retval, fmt.Sprintf("headerSize:                   %v", h.headerSize))
	retval = append(retval, fmt.Sprintf("offsetDataPoint:              %v", h.offsetDataPoint))
	retval = append(retval, fmt.Sprintf("numVarLengthRecords:          %v", h.numVarLengthRecords))
	retval = append(retval, fmt.Sprintf("pointDataRecordFormat:        %v", h.pointDataRecordFormat))
	retval = append(retval, fmt.Sprintf("pointDataRecordLength:        %v", h.pointDataRecordLength))
	retval = append(retval, fmt.Sprintf("legacyNumberPointRecords:     %v", h.legacyNumberPointRecords))
	retval = append(retval, fmt.Sprintf("legacyNumberPointsByReturn:   %v", h.legacyNumberPointsByReturn))
	retval = append(retval, fmt.Sprintf("xScaleFactor:                 %f", h.xScaleFactor))
	retval = append(retval, fmt.Sprintf("yScaleFactor:                 %f", h.yScaleFactor))
	retval = append(retval, fmt.Sprintf("zScaleFactor:                 %f", h.zScaleFactor))
	retval = append(retval, fmt.Sprintf("xOffset:                      %f", h.xOffset))
	retval = append(retval, fmt.Sprintf("yOffset:                      %f", h.yOffset))
	retval = append(retval, fmt.Sprintf("zOffset:                      %f", h.zOffset))
	retval = append(retval, fmt.Sprintf("maxX:                         %f", h.maxX))
	retval = append(retval, fmt.Sprintf("minX:                         %f", h.minX))
	retval = append(retval, fmt.Sprintf("maxY:                         %f", h.maxY))
	retval = append(retval, fmt.Sprintf("minY:                         %f", h.minY))
	retval = append(retval, fmt.Sprintf("maxZ:                         %f", h.maxZ))
	retval = append(retval, fmt.Sprintf("minZ:                         %f", h.minZ))
	return retval
}

type LasHeader14 struct {
	LasHeaderLegacy
	startWaveformRecord   uint64
	startFirstExtendedVlr uint64
	numExtendedVlr        uint32
	numberPointRecords    uint64
	numberPointsByReturn  []uint64
}

func (h *LasHeader14) Bounds() *geotiff.Bounds {
	// assume orientation upper left origin
	return &geotiff.Bounds{MaxX: h.maxX, MinX: h.minX, MaxY: h.maxY, MinY: h.minY, OriginX: h.minX, OriginY: h.maxY}
}

func (h *LasHeader14) GetNumberOfPoints() uint64 {
	if h.numberPointRecords != 0 {
		return h.numberPointRecords
	} else {
		return uint64(h.legacyNumberPointRecords)
	}
}

func (h *LasHeader14) ScalePoints(x int32, y int32, z int32) (float64, float64, float64) {
	return float64((float64(x) * h.xScaleFactor) + h.xOffset), float64((float64(y) * h.yScaleFactor) + h.yOffset), float64((float64(z) * h.zScaleFactor) + h.zOffset)
}

func (h *LasHeader14) GetPointFormat() byte {
	return h.pointDataRecordFormat
}

func (h *LasHeader14) GetPointsOffset() uint64 {
	return uint64(h.offsetDataPoint)
}

func (h *LasHeader14) GetPointLength() uint16 {
	return h.pointDataRecordLength
}

func (h *LasHeader14) GetNumberOfVLR() uint32 {
	return h.numVarLengthRecords
}

func (h *LasHeader14) GetNumberOfEVLR() uint32 {
	return h.numExtendedVlr
}

func (h *LasHeader14) GetOffsetOfEVLR() uint64 {
	return h.startFirstExtendedVlr
}

func (h *LasHeader14) GpsTime() GpsTimeType {
	if (int(h.globalEncoding) & geGpsMask) == 0 {
		return geGpsWeekTime
	} else {
		return geGpsStandardOffset
	}
}

func (h *LasHeader14) isWkt() bool {
	return int(h.globalEncoding)&geWktMask == geWkt
}

func (h *LasHeader14) Version() (int, int) {
	return int(h.versionMajor), int(h.versionMinor)
}

func (h *LasHeader14) VersionString() string {
	return fmt.Sprintf("%v.%v", h.versionMajor, h.versionMinor)
}

func (h *LasHeader14) Imageinfo() *ImageInfo {
	return &ImageInfo{
		width:   h.maxX - h.minX,
		height:  h.maxY - h.minY,
		scalex:  h.xScaleFactor,
		offsetx: h.xOffset,
		scaley:  h.yScaleFactor,
		offsety: h.yOffset,
		minx:    h.minX,
		maxx:    h.maxX,
		miny:    h.minY,
		maxy:    h.maxY,
		minz:    h.maxZ}
}

func (h *LasHeader14) DumpHeader() []string {
	retval := make([]string, 0, 33)
	retval = append(retval, fmt.Sprintf("fileSignature:                %s", h.fileSignature))
	retval = append(retval, fmt.Sprintf("fileSourceId:                 %v", h.fileSourceId))
	retval = append(retval, fmt.Sprintf("globalEncoding:               %v", h.globalEncoding))
	retval = append(retval, fmt.Sprintf("projectIdGuid:                %s", h.projectIdGuid))
	retval = append(retval, fmt.Sprintf("version:                      %v.%v", h.versionMajor, h.versionMinor))
	retval = append(retval, fmt.Sprintf("systemIdentifier:             %s", h.systemIdentifier))
	retval = append(retval, fmt.Sprintf("generatingSoftware:           %s", h.generatingSoftware))
	retval = append(retval, fmt.Sprintf("fileCreationDay/Yr:           %d/%d", h.fileCreationDayOfYear, h.fileCreationYear))
	retval = append(retval, fmt.Sprintf("headerSize:                   %v", h.headerSize))
	retval = append(retval, fmt.Sprintf("offsetDataPoint:              %v", h.offsetDataPoint))
	retval = append(retval, fmt.Sprintf("numVarLengthRecords:          %v", h.numVarLengthRecords))
	retval = append(retval, fmt.Sprintf("pointDataRecordFormat:        %v", h.pointDataRecordFormat))
	retval = append(retval, fmt.Sprintf("pointDataRecordLength:        %v", h.pointDataRecordLength))
	retval = append(retval, fmt.Sprintf("legacyNumberPointRecords:     %v", h.legacyNumberPointRecords))
	retval = append(retval, fmt.Sprintf("legacyNumberPointsByReturn:   %v", h.legacyNumberPointsByReturn))
	retval = append(retval, fmt.Sprintf("xScaleFactor:                 %f", h.xScaleFactor))
	retval = append(retval, fmt.Sprintf("yScaleFactor:                 %f", h.yScaleFactor))
	retval = append(retval, fmt.Sprintf("zScaleFactor:                 %f", h.zScaleFactor))
	retval = append(retval, fmt.Sprintf("xOffset:                      %f", h.xOffset))
	retval = append(retval, fmt.Sprintf("yOffset:                      %f", h.yOffset))
	retval = append(retval, fmt.Sprintf("zOffset:                      %f", h.zOffset))
	retval = append(retval, fmt.Sprintf("maxX:                         %f", h.maxX))
	retval = append(retval, fmt.Sprintf("minX:                         %f", h.minX))
	retval = append(retval, fmt.Sprintf("maxY:                         %f", h.maxY))
	retval = append(retval, fmt.Sprintf("minY:                         %f", h.minY))
	retval = append(retval, fmt.Sprintf("maxZ:                         %f", h.maxZ))
	retval = append(retval, fmt.Sprintf("minZ:                         %f", h.minZ))
	retval = append(retval, fmt.Sprintf("startWaveformRecord:          %v", h.startWaveformRecord))
	retval = append(retval, fmt.Sprintf("startFirstExtendedVlr:        %v", h.startFirstExtendedVlr))
	retval = append(retval, fmt.Sprintf("numExtendedVlr:               %v", h.numExtendedVlr))
	retval = append(retval, fmt.Sprintf("numberPointRecords:           %v", h.numberPointRecords))
	retval = append(retval, fmt.Sprintf("numberPointsByReturn:         %v", h.numberPointsByReturn))
	return retval
}
