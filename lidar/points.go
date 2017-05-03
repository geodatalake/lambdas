// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lidar

import (
	"encoding/binary"
	"math"
)

type PointFormat interface {
	ReadPoint([]byte)
	GetX() int32
	GetY() int32
	GetZ() int32
	GetClassification() int16
	GetReturnNumber() uint8
	GetTotalReturns() uint8
	GetIntensity() uint16
}

type Point0 struct {
	x                 int32
	y                 int32
	z                 int32
	intensity         uint16
	returnNumber      byte
	numberOfReturns   byte
	scanDirectionFlag byte
	edgeOfFlightLine  byte
	classification    byte
	scanAngleRank     int8
	userData          byte
	pointSourceID     uint16
}

func (p *Point0) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x07
	p.numberOfReturns = returnInfo & 0x38 >> 3
	p.scanDirectionFlag = returnInfo & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo & 0x80 >> 7
	p.classification = rawPoint[15:16][0]
	p.scanAngleRank = int8(rawPoint[16:17][0])
	p.userData = rawPoint[17:18][0]
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[18:20])
}

func (p *Point0) GetX() int32 {
	return p.x
}
func (p *Point0) GetY() int32 {
	return p.y
}

func (p *Point0) GetZ() int32 {
	return p.z
}

func (p *Point0) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point0) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point0) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point0) GetIntensity() uint16 {
	return p.intensity
}

type Point1 struct {
	x                 int32
	y                 int32
	z                 int32
	intensity         uint16
	returnNumber      byte
	numberOfReturns   byte
	scanDirectionFlag byte
	edgeOfFlightLine  byte
	classification    byte
	scanAngleRank     int8
	userData          byte
	pointSourceID     uint16
	gpsTime           float64
}

func (p *Point1) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x07
	p.numberOfReturns = returnInfo & 0x38 >> 3
	p.scanDirectionFlag = returnInfo & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo & 0x80 >> 7
	p.classification = rawPoint[15:16][0]
	p.scanAngleRank = int8(rawPoint[16:17][0])
	p.userData = rawPoint[17:18][0]
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[18:20])
	p.gpsTime = math.Float64frombits(binary.LittleEndian.Uint64(rawPoint[20:28]))
}

func (p *Point1) GetX() int32 {
	return p.x
}
func (p *Point1) GetY() int32 {
	return p.y
}

func (p *Point1) GetZ() int32 {
	return p.z
}

func (p *Point1) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point1) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point1) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point1) GetIntensity() uint16 {
	return p.intensity
}

type Point2 struct {
	x                 int32
	y                 int32
	z                 int32
	intensity         uint16
	returnNumber      byte
	numberOfReturns   byte
	scanDirectionFlag byte
	edgeOfFlightLine  byte
	classification    byte
	scanAngleRank     int8
	userData          byte
	pointSourceID     uint16
	red               uint16
	green             uint16
	blue              uint16
}

func (p *Point2) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x07
	p.numberOfReturns = returnInfo & 0x38 >> 3
	p.scanDirectionFlag = returnInfo & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo & 0x80 >> 7
	p.classification = rawPoint[15:16][0]
	p.scanAngleRank = int8(rawPoint[16:17][0])
	p.userData = rawPoint[17:18][0]
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[18:20])
	p.red = binary.LittleEndian.Uint16(rawPoint[20:22])
	p.green = binary.LittleEndian.Uint16(rawPoint[22:24])
	p.blue = binary.LittleEndian.Uint16(rawPoint[24:26])
}

func (p *Point2) GetX() int32 {
	return p.x
}
func (p *Point2) GetY() int32 {
	return p.y
}

func (p *Point2) GetZ() int32 {
	return p.z
}

func (p *Point2) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point2) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point2) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point2) GetIntensity() uint16 {
	return p.intensity
}

type Point3 struct {
	x                 int32
	y                 int32
	z                 int32
	intensity         uint16
	returnNumber      byte
	numberOfReturns   byte
	scanDirectionFlag byte
	edgeOfFlightLine  byte
	classification    byte
	scanAngleRank     int8
	userData          byte
	pointSourceID     uint16
	gpsTime           float64
	red               uint16
	green             uint16
	blue              uint16
}

func (p *Point3) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x07
	p.numberOfReturns = returnInfo & 0x38 >> 3
	p.scanDirectionFlag = returnInfo & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo & 0x80 >> 7
	p.classification = rawPoint[15:16][0]
	p.scanAngleRank = int8(rawPoint[16:17][0])
	p.userData = rawPoint[17:18][0]
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[18:20])
	p.gpsTime = math.Float64frombits(binary.LittleEndian.Uint64(rawPoint[20:28]))
	p.red = binary.LittleEndian.Uint16(rawPoint[28:30])
	p.green = binary.LittleEndian.Uint16(rawPoint[30:32])
	p.blue = binary.LittleEndian.Uint16(rawPoint[32:34])
}

func (p *Point3) GetX() int32 {
	return p.x
}
func (p *Point3) GetY() int32 {
	return p.y
}

func (p *Point3) GetZ() int32 {
	return p.z
}

func (p *Point3) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point3) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point3) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point3) GetIntensity() uint16 {
	return p.intensity
}

type Point4 struct {
	x                           int32
	y                           int32
	z                           int32
	intensity                   uint16
	returnNumber                byte
	numberOfReturns             byte
	scanDirectionFlag           byte
	edgeOfFlightLine            byte
	classification              byte
	scanAngleRank               int8
	userData                    byte
	pointSourceID               uint16
	gpsTime                     float64
	wavePacketDescriptorIndex   byte
	byteOffsetToWaveformData    uint64
	WaveformPacketSizeInBytes   uint32
	returnPointWaveformLocation float32
	xt                          float32
	yt                          float32
	zt                          float32
}

func (p *Point4) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x07
	p.numberOfReturns = returnInfo & 0x38 >> 3
	p.scanDirectionFlag = returnInfo & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo & 0x80 >> 7
	p.classification = rawPoint[15:16][0]
	p.scanAngleRank = int8(rawPoint[16:17][0])
	p.userData = rawPoint[17:18][0]
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[18:20])
	p.gpsTime = math.Float64frombits(binary.LittleEndian.Uint64(rawPoint[20:28]))
	p.wavePacketDescriptorIndex = rawPoint[28:29][0]
	p.byteOffsetToWaveformData = binary.LittleEndian.Uint64(rawPoint[29:37])
	p.WaveformPacketSizeInBytes = binary.LittleEndian.Uint32(rawPoint[37:41])
	p.returnPointWaveformLocation = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[41:45]))
	p.xt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[45:49]))
	p.yt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[49:53]))
	p.zt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[53:57]))
}

func (p *Point4) GetX() int32 {
	return p.x
}
func (p *Point4) GetY() int32 {
	return p.y
}

func (p *Point4) GetZ() int32 {
	return p.z
}

func (p *Point4) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point4) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point4) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point4) GetIntensity() uint16 {
	return p.intensity
}

type Point5 struct {
	x                           int32
	y                           int32
	z                           int32
	intensity                   uint16
	returnNumber                byte
	numberOfReturns             byte
	scanDirectionFlag           byte
	edgeOfFlightLine            byte
	classification              byte
	scanAngleRank               int8
	userData                    byte
	pointSourceID               uint16
	gpsTime                     float64
	red                         uint16
	green                       uint16
	blue                        uint16
	wavePacketDescriptorIndex   byte
	byteOffsetToWaveformData    uint64
	WaveformPacketSizeInBytes   uint32
	returnPointWaveformLocation float32
	xt                          float32
	yt                          float32
	zt                          float32
}

func (p *Point5) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x07
	p.numberOfReturns = returnInfo & 0x38 >> 3
	p.scanDirectionFlag = returnInfo & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo & 0x80 >> 7
	p.classification = rawPoint[15:16][0]
	p.scanAngleRank = int8(rawPoint[16:17][0])
	p.userData = rawPoint[17:18][0]
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[18:20])
	p.gpsTime = math.Float64frombits(binary.LittleEndian.Uint64(rawPoint[20:28]))
	p.red = binary.LittleEndian.Uint16(rawPoint[28:30])
	p.green = binary.LittleEndian.Uint16(rawPoint[30:32])
	p.blue = binary.LittleEndian.Uint16(rawPoint[32:34])
	p.wavePacketDescriptorIndex = rawPoint[34:35][0]
	p.byteOffsetToWaveformData = binary.LittleEndian.Uint64(rawPoint[35:43])
	p.WaveformPacketSizeInBytes = binary.LittleEndian.Uint32(rawPoint[43:47])
	p.returnPointWaveformLocation = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[47:51]))
	p.xt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[51:55]))
	p.yt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[55:59]))
	p.zt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[59:63]))
}

func (p *Point5) GetX() int32 {
	return p.x
}
func (p *Point5) GetY() int32 {
	return p.y
}

func (p *Point5) GetZ() int32 {
	return p.z
}

func (p *Point5) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point5) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point5) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point5) GetIntensity() uint16 {
	return p.intensity
}

type Point6 struct {
	x                   int32
	y                   int32
	z                   int32
	intensity           uint16
	returnNumber        byte
	numberOfReturns     byte
	classificationFlags byte
	scannerChannel      byte
	scanDirectionFlag   byte
	edgeOfFlightLine    byte
	classification      byte
	userData            byte
	scanAngle           int16
	pointSourceID       uint16
	gpsTime             float64
}

func (p *Point6) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x0f
	p.numberOfReturns = returnInfo & 0xf0 >> 4
	returnInfo2 := rawPoint[15:16][0]
	p.classificationFlags = returnInfo2 & 0x0f
	p.scannerChannel = returnInfo2 & 0x30 >> 4
	p.scanDirectionFlag = returnInfo2 & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo2 & 0x80 >> 7
	p.classification = rawPoint[16:17][0]
	p.userData = rawPoint[17:18][0]
	p.scanAngle = int16(binary.LittleEndian.Uint16(rawPoint[18:20]))
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[20:22])
	p.gpsTime = math.Float64frombits(binary.LittleEndian.Uint64(rawPoint[22:30]))
}

func (p *Point6) GetX() int32 {
	return p.x
}
func (p *Point6) GetY() int32 {
	return p.y
}

func (p *Point6) GetZ() int32 {
	return p.z
}

func (p *Point6) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point6) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point6) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point6) GetIntensity() uint16 {
	return p.intensity
}

type Point7 struct {
	Point6
	red   uint16
	green uint16
	blue  uint16
}

func (p *Point7) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x0f
	p.numberOfReturns = returnInfo & 0xf0 >> 4
	returnInfo2 := rawPoint[15:16][0]
	p.classificationFlags = returnInfo2 & 0x0f
	p.scannerChannel = returnInfo2 & 0x30 >> 4
	p.scanDirectionFlag = returnInfo2 & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo2 & 0x80 >> 7
	p.classification = rawPoint[16:17][0]
	p.userData = rawPoint[17:18][0]
	p.scanAngle = int16(binary.LittleEndian.Uint16(rawPoint[18:20]))
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[20:22])
	p.gpsTime = math.Float64frombits(binary.LittleEndian.Uint64(rawPoint[22:30]))
	p.red = binary.LittleEndian.Uint16(rawPoint[30:32])
	p.green = binary.LittleEndian.Uint16(rawPoint[32:34])
	p.blue = binary.LittleEndian.Uint16(rawPoint[34:36])
}

func (p *Point7) GetX() int32 {
	return p.x
}
func (p *Point7) GetY() int32 {
	return p.y
}

func (p *Point7) GetZ() int32 {
	return p.z
}

func (p *Point7) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point7) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point7) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point7) GetIntensity() uint16 {
	return p.intensity
}

type Point8 struct {
	Point7
	nir uint16
}

func (p *Point8) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x0f
	p.numberOfReturns = returnInfo & 0xf0 >> 4
	returnInfo2 := rawPoint[15:16][0]
	p.classificationFlags = returnInfo2 & 0x0f
	p.scannerChannel = returnInfo2 & 0x30 >> 4
	p.scanDirectionFlag = returnInfo2 & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo2 & 0x80 >> 7
	p.classification = rawPoint[16:17][0]
	p.userData = rawPoint[17:18][0]
	p.scanAngle = int16(binary.LittleEndian.Uint16(rawPoint[18:20]))
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[20:22])
	p.gpsTime = math.Float64frombits(binary.LittleEndian.Uint64(rawPoint[22:30]))
	p.red = binary.LittleEndian.Uint16(rawPoint[30:32])
	p.green = binary.LittleEndian.Uint16(rawPoint[32:34])
	p.blue = binary.LittleEndian.Uint16(rawPoint[34:36])
	p.nir = binary.LittleEndian.Uint16(rawPoint[36:38])
}

func (p *Point8) GetX() int32 {
	return p.x
}
func (p *Point8) GetY() int32 {
	return p.y
}

func (p *Point8) GetZ() int32 {
	return p.z
}

func (p *Point8) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point8) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point8) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point8) GetIntensity() uint16 {
	return p.intensity
}

type Point9 struct {
	Point6
	wavePacketDescriptorIndex   byte
	byteOffsetToWaveformData    uint64
	WaveformPacketSizeInBytes   uint32
	returnPointWaveformLocation float32
	xt                          float32
	yt                          float32
	zt                          float32
}

func (p *Point9) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x0f
	p.numberOfReturns = returnInfo & 0xf0 >> 4
	returnInfo2 := rawPoint[15:16][0]
	p.classificationFlags = returnInfo2 & 0x0f
	p.scannerChannel = returnInfo2 & 0x30 >> 4
	p.scanDirectionFlag = returnInfo2 & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo2 & 0x80 >> 7
	p.classification = rawPoint[16:17][0]
	p.userData = rawPoint[17:18][0]
	p.scanAngle = int16(binary.LittleEndian.Uint16(rawPoint[18:20]))
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[20:22])
	p.gpsTime = math.Float64frombits(binary.LittleEndian.Uint64(rawPoint[22:30]))
	p.wavePacketDescriptorIndex = rawPoint[30:31][0]
	p.byteOffsetToWaveformData = binary.LittleEndian.Uint64(rawPoint[31:39])
	p.WaveformPacketSizeInBytes = binary.LittleEndian.Uint32(rawPoint[39:43])
	p.returnPointWaveformLocation = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[43:47]))
	p.xt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[47:51]))
	p.yt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[51:55]))
	p.zt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[55:59]))
}

func (p *Point9) GetX() int32 {
	return p.x
}

func (p *Point9) GetY() int32 {
	return p.y
}

func (p *Point9) GetZ() int32 {
	return p.z
}

func (p *Point9) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point9) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point9) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point9) GetIntensity() uint16 {
	return p.intensity
}

type Point10 struct {
	Point7
	wavePacketDescriptorIndex   byte
	byteOffsetToWaveformData    uint64
	WaveformPacketSizeInBytes   uint32
	returnPointWaveformLocation float32
	xt                          float32
	yt                          float32
	zt                          float32
}

func (p *Point10) ReadPoint(rawPoint []byte) {
	p.x = int32(binary.LittleEndian.Uint32(rawPoint[0:4]))
	p.y = int32(binary.LittleEndian.Uint32(rawPoint[4:8]))
	p.z = int32(binary.LittleEndian.Uint32(rawPoint[8:12]))
	p.intensity = binary.LittleEndian.Uint16(rawPoint[12:14])
	returnInfo := rawPoint[14:15][0]
	p.returnNumber = returnInfo & 0x0f
	p.numberOfReturns = returnInfo & 0xf0 >> 4
	returnInfo2 := rawPoint[15:16][0]
	p.classificationFlags = returnInfo2 & 0x0f
	p.scannerChannel = returnInfo2 & 0x30 >> 4
	p.scanDirectionFlag = returnInfo2 & 0x40 >> 6
	p.edgeOfFlightLine = returnInfo2 & 0x80 >> 7
	p.classification = rawPoint[16:17][0]
	p.userData = rawPoint[17:18][0]
	p.scanAngle = int16(binary.LittleEndian.Uint16(rawPoint[18:20]))
	p.pointSourceID = binary.LittleEndian.Uint16(rawPoint[20:22])
	p.gpsTime = math.Float64frombits(binary.LittleEndian.Uint64(rawPoint[22:30]))
	p.red = binary.LittleEndian.Uint16(rawPoint[30:32])
	p.green = binary.LittleEndian.Uint16(rawPoint[32:34])
	p.blue = binary.LittleEndian.Uint16(rawPoint[34:36])
	p.wavePacketDescriptorIndex = rawPoint[36:37][0]
	p.byteOffsetToWaveformData = binary.LittleEndian.Uint64(rawPoint[37:45])
	p.WaveformPacketSizeInBytes = binary.LittleEndian.Uint32(rawPoint[45:49])
	p.returnPointWaveformLocation = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[49:53]))
	p.xt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[53:57]))
	p.yt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[57:61]))
	p.zt = math.Float32frombits(binary.LittleEndian.Uint32(rawPoint[61:65]))
}

func (p *Point10) GetX() int32 {
	return p.x
}
func (p *Point10) GetY() int32 {
	return p.y
}

func (p *Point10) GetZ() int32 {
	return p.z
}

func (p *Point10) GetClassification() int16 {
	return int16(p.classification)
}

func (p *Point10) GetReturnNumber() uint8 {
	return p.returnNumber
}

func (p *Point10) GetTotalReturns() uint8 {
	return p.numberOfReturns
}

func (p *Point10) GetIntensity() uint16 {
	return p.intensity
}
