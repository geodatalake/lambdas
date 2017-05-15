// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lidar

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/geodatalake/lambdas/geotiff"
)

type NotaLasFile []byte

func (s NotaLasFile) Error() string {
	b := make([]byte, 4)
	copy(b, s) // have to offload the bytes out of the type as Sprintf will attempt to format it and lead to a panic
	return fmt.Sprintf("Not a LAS file, signature bytes are %v", b)
}

type Las interface {
	Build() (*geotiff.Raster, error)
	VariableLengthRecords() []*Vlr
	Summarize(*Vlr) string
	GeotiffCrs() *CrsRecordGeoTiff
	SummarizeGeokey(*geotiff.GeoKey) string
	WktCrs() *CrsRecordWkt
	IsWktCrs() bool
	KeyFor(key int) (*geotiff.GeoKey, error)
	DumpHeader() []string
	Bounds() *geotiff.Bounds
	IsLaszip() bool
	DateTime() (time.Time, error)
	Close() bool
}

type CrsRecordGeoTiff struct {
	Asciis              []byte
	Doubles             []float64
	Geokeys             map[uint16]*geotiff.GeoKey
	KeyDirectoryVersion uint16
	KeyRevision         uint16
	MinorRevision       uint16
	NumberOfKeys        uint16
}

func (geocrs *CrsRecordGeoTiff) String() string {
	lines := make([]string, 0, 32)
	for _, v := range geocrs.Geokeys {
		lines = append(lines, fmt.Sprintf("%s: %v",
			geotiff.NameForKey(int(v.KeyId)),
			geotiff.ValueForKey(int(v.KeyId), int(v.Location), 0, int(v.Count), geocrs.Doubles, geocrs.Asciis)))
	}
	return strings.Join(lines, "\n")
}

type CrsRecordWkt struct {
	Wkt string
}

type decoder struct {
	reader     io.ReaderAt
	byteOrder  binary.ByteOrder
	header     HeaderFormat
	vlrs       []*Vlr
	evlrs      []*Evlr
	crsGeotiff *CrsRecordGeoTiff
	crsWkt     *CrsRecordWkt
	opt        *ReadOptions
}

func (d *decoder) Close() bool {
	if c, ok := d.reader.(io.Closer); ok {
		c.Close()
	}
	return true
}

func (d *decoder) DateTime() (time.Time, error) {
	return d.header.DateTime()
}

func (d *decoder) SummarizeGeokey(geokey *geotiff.GeoKey) string {
	name := geotiff.NameForKey(int(geokey.KeyId))
	value := geotiff.ValueForKey(int(geokey.KeyId), int(geokey.Location), int(geokey.Value), int(geokey.Count), d.crsGeotiff.Doubles, d.crsGeotiff.Asciis)
	return fmt.Sprintf("%s: %v", name, value)
}

func (d *decoder) Summarize(vlr *Vlr) string {
	return fmt.Sprintf("reserved: %v\nuserID [%v]: %s\nrecordID: %v\nlengthAfterHeader: %v\ndescription [%v]: %s\n",
		vlr.reserved, len(vlr.userID), vlr.userID, vlr.recordID, vlr.lengthAfterHeader, len(vlr.description), vlr.description)
}

func (d *decoder) IsLaszip() bool {
	for _, vlr := range d.vlrs {
		if vlr.userID == laszipSignature {
			return true
		}
	}
	return false
}

func (d *decoder) Bounds() *geotiff.Bounds {
	return d.header.Bounds()
}

func (d *decoder) readLegacyHeader(rawHeader []byte) *LasHeaderLegacy {
	lh := &LasHeaderLegacy{}
	lh.fileSignature = "LASF"
	lh.fileSourceId = d.byteOrder.Uint16(rawHeader[4:6])
	lh.globalEncoding = d.byteOrder.Uint16(rawHeader[6:8])
	lh.projectIdGuid = fmt.Sprintf("%X-%X-%X-%X", d.byteOrder.Uint32(rawHeader[8:12]), d.byteOrder.Uint16(rawHeader[12:14]), d.byteOrder.Uint16(rawHeader[14:16]), rawHeader[16:24])
	lh.versionMajor = rawHeader[24:25][0]
	lh.versionMinor = rawHeader[25:26][0]
	lh.generatingSoftware = string(rawHeader[58:90])
	lh.fileCreationDayOfYear = d.byteOrder.Uint16(rawHeader[90:92])
	lh.fileCreationYear = d.byteOrder.Uint16(rawHeader[92:94])
	lh.headerSize = uint16(len(rawHeader))
	lh.offsetDataPoint = d.byteOrder.Uint32(rawHeader[96:100])
	lh.numVarLengthRecords = d.byteOrder.Uint32(rawHeader[100:104])
	lh.pointDataRecordFormat = rawHeader[104:105][0]
	lh.pointDataRecordLength = d.byteOrder.Uint16(rawHeader[105:107])
	lh.legacyNumberPointRecords = d.byteOrder.Uint32(rawHeader[107:111])
	lh.legacyNumberPointsByReturn = make([]uint32, 5)
	for i := 0; i < 5; i++ {
		lh.legacyNumberPointsByReturn[i] = d.byteOrder.Uint32(rawHeader[111+(4*i) : 115+(4*i)])
	}
	lh.xScaleFactor = math.Float64frombits(d.byteOrder.Uint64(rawHeader[131:139]))
	lh.yScaleFactor = math.Float64frombits(d.byteOrder.Uint64(rawHeader[139:147]))
	lh.zScaleFactor = math.Float64frombits(d.byteOrder.Uint64(rawHeader[147:155]))
	lh.xOffset = math.Float64frombits(d.byteOrder.Uint64(rawHeader[155:163]))
	lh.yOffset = math.Float64frombits(d.byteOrder.Uint64(rawHeader[163:171]))
	lh.zOffset = math.Float64frombits(d.byteOrder.Uint64(rawHeader[171:179]))
	lh.maxX = math.Float64frombits(d.byteOrder.Uint64(rawHeader[179:187]))
	lh.minX = math.Float64frombits(d.byteOrder.Uint64(rawHeader[187:195]))
	lh.maxY = math.Float64frombits(d.byteOrder.Uint64(rawHeader[195:203]))
	lh.minY = math.Float64frombits(d.byteOrder.Uint64(rawHeader[203:211]))
	lh.maxZ = math.Float64frombits(d.byteOrder.Uint64(rawHeader[211:219]))
	lh.minZ = math.Float64frombits(d.byteOrder.Uint64(rawHeader[219:227]))
	return lh
}

func (d *decoder) readLas14Header(rawHeader []byte) *LasHeader14 {
	lh := &LasHeader14{}
	lh.fileSignature = "LASF"
	lh.fileSourceId = d.byteOrder.Uint16(rawHeader[4:6])
	lh.globalEncoding = d.byteOrder.Uint16(rawHeader[6:8])
	lh.projectIdGuid = fmt.Sprintf("%x-%x-%x-%x", d.byteOrder.Uint32(rawHeader[8:12]), d.byteOrder.Uint16(rawHeader[12:14]), d.byteOrder.Uint16(rawHeader[14:16]), rawHeader[16:24])
	lh.versionMajor = rawHeader[24:25][0]
	lh.versionMinor = rawHeader[25:26][0]
	lh.generatingSoftware = string(rawHeader[58:90])
	lh.fileCreationDayOfYear = d.byteOrder.Uint16(rawHeader[90:92])
	lh.fileCreationYear = d.byteOrder.Uint16(rawHeader[92:94])
	lh.headerSize = uint16(len(rawHeader))
	lh.offsetDataPoint = d.byteOrder.Uint32(rawHeader[96:100])
	lh.numVarLengthRecords = d.byteOrder.Uint32(rawHeader[100:104])
	lh.pointDataRecordFormat = rawHeader[104:105][0]
	lh.pointDataRecordLength = d.byteOrder.Uint16(rawHeader[105:107])
	lh.legacyNumberPointRecords = d.byteOrder.Uint32(rawHeader[107:111])
	lh.legacyNumberPointsByReturn = make([]uint32, 5)
	for i := 0; i < 5; i++ {
		lh.legacyNumberPointsByReturn[i] = d.byteOrder.Uint32(rawHeader[111+(4*i) : 115+(4*i)])
	}
	lh.xScaleFactor = math.Float64frombits(d.byteOrder.Uint64(rawHeader[131:139]))
	lh.yScaleFactor = math.Float64frombits(d.byteOrder.Uint64(rawHeader[139:147]))
	lh.zScaleFactor = math.Float64frombits(d.byteOrder.Uint64(rawHeader[147:155]))
	lh.xOffset = math.Float64frombits(d.byteOrder.Uint64(rawHeader[155:163]))
	lh.yOffset = math.Float64frombits(d.byteOrder.Uint64(rawHeader[163:171]))
	lh.zOffset = math.Float64frombits(d.byteOrder.Uint64(rawHeader[171:179]))
	lh.maxX = math.Float64frombits(d.byteOrder.Uint64(rawHeader[179:187]))
	lh.minX = math.Float64frombits(d.byteOrder.Uint64(rawHeader[187:195]))
	lh.maxY = math.Float64frombits(d.byteOrder.Uint64(rawHeader[195:203]))
	lh.minY = math.Float64frombits(d.byteOrder.Uint64(rawHeader[203:211]))
	lh.maxZ = math.Float64frombits(d.byteOrder.Uint64(rawHeader[211:219]))
	lh.minZ = math.Float64frombits(d.byteOrder.Uint64(rawHeader[219:227]))
	//	lh.startWaveformRecord = d.byteOrder.Uint64(rawHeader[227:235])
	if len(rawHeader) > 254 {
		lh.startFirstExtendedVlr = d.byteOrder.Uint64(rawHeader[235:243])
		lh.numExtendedVlr = d.byteOrder.Uint32(rawHeader[243:247])
		lh.numberPointRecords = d.byteOrder.Uint64(rawHeader[247:255])
	}
	//	lh.numberPointsByReturn = make([]uint64, 15)
	//	for i := 0; i < 15; i++ {
	//		lh.numberPointsByReturn[i] = d.byteOrder.Uint64(rawHeader[255+(i*8):263+(i*8)])
	//	}
	return lh
}

func (d *decoder) readLasHeader(rawHeader []byte) HeaderFormat {
	var retval HeaderFormat
	if len(rawHeader) == 227 {
		retval = d.readLegacyHeader(rawHeader)
	} else {
		retval = d.readLas14Header(rawHeader)
	}
	return retval
}

func (d *decoder) parseGeotiffCrs() {
	geoCrs := &CrsRecordGeoTiff{Geokeys: make(map[uint16]*geotiff.GeoKey)}
	for _, vlr := range d.vlrs {
		if vlr.userID == geotiffSignature {
			switch vlr.recordID {
			case RGeoKeys:
				geoKeys := make([]uint16, len(vlr.data)/2)
				for i := 0; i < len(vlr.data); i += 2 {
					geoKeys[i/2] = d.byteOrder.Uint16(vlr.data[i : i+2])
				}
				geoCrs.KeyDirectoryVersion = geoKeys[0]
				geoCrs.KeyRevision = geoKeys[1]
				geoCrs.MinorRevision = geoKeys[2]
				geoCrs.NumberOfKeys = geoKeys[3]
				for i := 0; i < int(geoKeys[3]); i++ {
					pos := 4 + (i * 4)
					keyId := geoKeys[pos]
					location := geoKeys[pos+1]
					count := geoKeys[pos+2]
					offset := geoKeys[pos+3]
					if d.opt != nil && d.opt.FilterCrs {
						if _, ok := d.opt.AcceptableGeoKeys[int(keyId)]; ok {
							geoCrs.Geokeys[keyId] = &geotiff.GeoKey{KeyId: keyId, Location: location, Count: count, Value: offset}
						}
					} else {
						geoCrs.Geokeys[keyId] = &geotiff.GeoKey{KeyId: keyId, Location: location, Count: count, Value: offset}
					}
				}
				if d.opt != nil && d.opt.FilterCrs {
					geoCrs.NumberOfKeys = uint16(len(geoCrs.Geokeys))
				}
			case RGeoDoubles:
				geodoubles := make([]float64, len(vlr.data)/8)
				for i := 0; i < len(vlr.data); i += 8 {
					geodoubles[i/8] = math.Float64frombits(d.byteOrder.Uint64(vlr.data[i : i+8]))
				}
				geoCrs.Doubles = geodoubles
			case RGeoAscii:
				geoCrs.Asciis = vlr.data
			}
		}
	}
	d.crsGeotiff = geoCrs
}

func (d *decoder) KeyFor(key int) (*geotiff.GeoKey, error) {
	ret, ok := d.crsGeotiff.Geokeys[uint16(key)]
	if ok {
		return ret, nil
	}
	return nil, geotiff.GeneralIssue(fmt.Sprintf("GeoKey for %v is not present", key))
}

func (d *decoder) parseWktCrs() {
	for _, evlr := range d.evlrs {
		if evlr.userID == geotiffSignature {
			switch evlr.recordID {
			case MathTransformWKT:
				d.crsWkt = &CrsRecordWkt{Wkt: string(evlr.data)}
			case CoordinateSystemWKT:
				d.crsWkt = &CrsRecordWkt{Wkt: string(evlr.data)}
			}
		}
	}
}

func (d *decoder) parseCrsRecord() {
	if d.header.isWkt() {
		d.parseWktCrs()
	} else {
		d.parseGeotiffCrs()
	}
}

func (d *decoder) GeotiffCrs() *CrsRecordGeoTiff {
	return d.crsGeotiff
}

func (d *decoder) WktCrs() *CrsRecordWkt {
	return d.crsWkt
}

func (d *decoder) IsWktCrs() bool {
	return d.header.isWkt()
}

func (d *decoder) VariableLengthRecords() []*Vlr {
	return d.vlrs
}

func (d *decoder) DumpHeader() []string {
	return d.header.DumpHeader()
}

type PointPacket struct {
	cancel                      bool
	num                         int64
	points                      []byte
	onlyFirstReturns            bool
	onlyBareEarthClassification bool
	onlyClassifications         bool
	onlyLastReturns             bool
	onlyIntensity               bool
}

type PointReturn struct {
	points []float64
	totalX float64
	totalY float64
	totalZ float64
	c      []int64
	r      []int64
	cancel bool
}

func (p *PointReturn) combine(other *PointReturn) {
	p.totalX += other.totalX
	p.totalY += other.totalY
	p.totalZ += other.totalZ
	for i, v := range other.c {
		p.c[i] += v
	}
	for i, v := range other.r {
		p.r[i] += v
	}
	p.points = append(p.points, other.points...)
}

func readPoints(chIn chan *PointPacket, chOut chan *PointReturn, header HeaderFormat, waiter *sync.WaitGroup) {
	var point PointFormat
	switch header.GetPointFormat() {
	case 0:
		point = &Point0{}
	case 1:
		point = &Point1{}
	case 2:
		point = &Point2{}
	case 3:
		point = &Point3{}
	case 4:
		point = &Point4{}
	case 5:
		point = &Point5{}
	case 6:
		point = &Point6{}
	case 7:
		point = &Point7{}
	case 8:
		point = &Point8{}
	case 9:
		point = &Point9{}
	case 10:
		point = &Point10{}
	default:
		panic(fmt.Sprintf("No point format for %v", header.GetPointFormat()))
	}
	pointLength := int64(header.GetPointLength())
	for {
		retval := &PointReturn{
			c: make([]int64, 257),
			r: make([]int64, 11)}
		packet := <-chIn
		if packet.cancel {
			waiter.Done()
			return
		}
		filtering := packet.onlyFirstReturns || packet.onlyBareEarthClassification || packet.onlyClassifications || packet.onlyLastReturns || packet.onlyIntensity
		retval.points = make([]float64, 0, packet.num*3)

		for i := int64(0); i < packet.num; i++ {
			point.ReadPoint(packet.points[i*pointLength : (i+1)*pointLength])
			c := int(point.GetClassification())
			if c >= 256 {
				c = 256
			}
			r := int(point.GetReturnNumber())

			if r >= 10 {
				r = 10
			}
			retval.c[c]++
			retval.r[r]++
			fx, fy, fz := header.ScalePoints(point.GetX(), point.GetY(), point.GetZ())
			retval.totalX += fx
			retval.totalY += fy
			retval.totalZ += fz
			if !filtering || (packet.onlyFirstReturns && point.GetReturnNumber() == 1) || (packet.onlyBareEarthClassification && point.GetClassification() == 2) ||
				(packet.onlyLastReturns) || (packet.onlyIntensity) {

				if packet.onlyClassifications {
					retval.points = append(retval.points, fx, fy, float64(point.GetClassification()))
				} else if packet.onlyIntensity {
					ptIntensity := point.GetIntensity()
					if ptIntensity > 7.0 {
						retval.points = append(retval.points, fx, fy, float64(ptIntensity))
					} else {
						retval.points = append(retval.points, fx, fy, float64(-9999))
					}
				} else {

					if packet.onlyLastReturns {

						if point.GetReturnNumber() == point.GetTotalReturns() {
							retval.points = append(retval.points, fx, fy, fz)
						}

					} else {
						retval.points = append(retval.points, fx, fy, fz)
					}
				}

			} else if packet.onlyClassifications {
				retval.points = append(retval.points, fx, fy, float64(point.GetClassification()))
			}
		}
		chOut <- retval
	}
}

func makePointPacket(f io.ReaderAt, opt *ReadOptions, sz uint16, num int64, pointFormat byte, offset int64, pointIndex uint64) *PointPacket {
	data := make([]byte, int64(sz)*num)
	if _, err := f.ReadAt(data, offset); err != nil {
		fmt.Println(err)
		fmt.Printf("Error at pointIndex %v, offset %v, num=%v and size=%v\n", pointIndex, offset, num, sz)
		panic(err)
	}
	var oft = false
	var obe = false
	var oc = false
	var olt = false
	var intensity = false
	if opt != nil {
		if opt.BareEarthClass {
			obe = true
		}
		if opt.FirstReturns {
			oft = true
		}
		if opt.GatherClassifications {
			oc = true
		}
		if opt.LastReturns {
			olt = true
		}
		if opt.Intensity {
			intensity = true
		}
	}
	return &PointPacket{
		cancel:                      false,
		num:                         num,
		points:                      data,
		onlyFirstReturns:            oft,
		onlyBareEarthClassification: obe,
		onlyClassifications:         oc,
		onlyLastReturns:             olt,
		onlyIntensity:               intensity}
}

func MergeValues(initial *PointReturn, output chan *PointReturn, waiter *sync.WaitGroup) {
	for {
		p := <-output
		if p.cancel {
			waiter.Done()
			break
		} else {
			initial.combine(p)
		}
	}
}

func filterLegacyClassifications(classification uint16) bool {
	switch classification {
	case 0: // Never Classified
		return false
	case 1: // Unclassified
		return false
	case 7: // Low Point (noise)
		return false
	case 8: // Reserved
		return false
	default:
		if classification > 9 {
			return false
		}
	}
	return true
}

func filterClassifications(classification uint16) bool {
	switch classification {
	case 0: // Never Classified
		return false
	case 1: // Unclassified
		return false
	case 7: // Low Point (noise)
		return false
	case 8: // Reserved
		return false
	case 18: // High Noise
		return false
	default:
		if classification > 18 {
			return false
		}
	}
	return true
}

func countNoDataCells(ras *geotiff.Raster) int {

	rasterWidth := ras.Width()
	rasterLength := ras.Height()

	var nodatacnt = 0
	for row := 0; row < rasterLength; row++ {
		for col := 0; col < rasterWidth; col++ {
			if ras.ValueAt(row, col) == -9999 {
				nodatacnt++
			}
		}
	}

	fmt.Println("Number Data Points ", nodatacnt)

	return nodatacnt
}

func findMinLimit(min int, value int, neighborhood int) int {

	var returnVal = 0
	if value-neighborhood < min {
		if value > min {
			returnVal = min
		} else {
			returnVal = value + 1
		}

	} else {
		returnVal = value - neighborhood
	}

	return returnVal
}

func findMaxLimit(max int, value int, neighborhood int) int {

	var returnVal = 0
	if value+neighborhood > (max - 1) {
		if value < (max - 1) {
			returnVal = max
		} else {
			returnVal = value - 1
		}

	} else {
		returnVal = value + neighborhood
	}

	return returnVal
}

func interpolateCell(ras *geotiff.Raster, cellrow int, cellcol int) float32 {

	neighborhoodSize := 2
	neighborhoodWidth := (neighborhoodSize + neighborhoodSize + 1)
	neighborhoodArea := neighborhoodWidth * neighborhoodWidth

	rasterWidth := ras.Width()
	rasterLength := ras.Height()

	var startRow = findMinLimit(0, cellrow, neighborhoodSize)
	var startCol = findMinLimit(0, cellcol, neighborhoodSize)
	var endRow = findMaxLimit(rasterLength, cellrow, neighborhoodSize)
	var endCol = findMaxLimit(rasterWidth, cellcol, neighborhoodSize)

	var cellsCounted = 0
	var valueCount float32 = 0
	var returnVal = ras.ValueAt(cellrow, cellcol)

	for row := startRow; row < endRow; row++ {
		for col := startCol; col < endCol; col++ {
			data := ras.ValueAt(row, col)
			if data != -9999 && (row != cellrow && col != cellcol) {
				cellsCounted++
				valueCount += data
			}
		}
	}

	if cellsCounted >= (neighborhoodArea / 10) {
		returnVal = valueCount / float32(cellsCounted)
	}

	return returnVal
}

func (d *decoder) Build() (*geotiff.Raster, error) {
	imageInfo := d.header.Imageinfo()

	format := d.header.GetPointFormat()
	numPoints := d.header.GetNumberOfPoints()
	fmt.Println("Reading", humanize.Comma(int64(numPoints)), "Points", "Format", format, "Point Length", d.header.GetPointLength(), "LAS Version", d.header.VersionString())

	input := make(chan *PointPacket, 15)
	output := make(chan *PointReturn, 15)
	values := &PointReturn{
		points: make([]float64, 0, numPoints*3),
		c:      make([]int64, 257),
		r:      make([]int64, 11)}
	var waiter sync.WaitGroup
	waiter.Add(4)
	go readPoints(input, output, d.header, &waiter)
	go readPoints(input, output, d.header, &waiter)
	go readPoints(input, output, d.header, &waiter)
	go readPoints(input, output, d.header, &waiter)
	go MergeValues(values, output, &waiter)
	t0 := time.Now()
	pointsOffset := int64(d.header.GetPointsOffset())
	chunkSize := uint64(10000)
	for pt := uint64(0); pt < numPoints; pt += chunkSize {
		var numPacketPoints int64
		if (pt + chunkSize) < numPoints {
			numPacketPoints = int64(chunkSize)
		} else {
			numPacketPoints = int64(numPoints - pt)
		}
		packet := makePointPacket(d.reader, d.opt, d.header.GetPointLength(), numPacketPoints, format, pointsOffset, pt)
		pointsOffset += numPacketPoints * int64(d.header.GetPointLength())
		input <- packet
	}
	cancelPacket := &PointPacket{cancel: true}
	for i := 0; i < 4; i++ {
		input <- cancelPacket
	}
	waiter.Wait()

	waiter.Add(1)
	output <- &PointReturn{cancel: true}
	waiter.Wait()
	t1 := time.Now()
	fmt.Printf("Completed points in %v, avg z: %f\n", t1.Sub(t0), values.totalZ/float64(numPoints))
	rows := 1 + int(math.Floor(imageInfo.height))
	cols := 1 + int(math.Floor(imageInfo.width))
	grid := make([][]*[]float64, 0, rows)
	var raster *geotiff.Raster
	raster = geotiff.NewRaster(cols, rows)
	for i := 0; i < rows; i++ {
		grid = append(grid, make([]*[]float64, cols))
	}
	t2 := time.Now()
	totalZ := 0.0
	fmt.Printf("Formatted grid %v x %v in %v\n", rows, cols, t2.Sub(t1))
	bounds := d.header.Bounds()
	xinc := bounds.Xspan() / imageInfo.width
	yinc := bounds.Yspan() / imageInfo.height
	fmt.Println("XSpan:", bounds.Xspan(), "YSpan:", bounds.Yspan(), "xinc", xinc, "yinc", yinc)
	errPoints := 0
	for i := 0; i+3 < len(values.points); i += 3 {
		col := int(math.Floor((values.points[i] - imageInfo.minx) / xinc))
		row := int(math.Floor((imageInfo.maxy - values.points[i+1]) / yinc)) // origin is minX, maxY
		zval := values.points[i+2]
		if row >= 0 && row < rows && col >= 0 && col < cols && zval >= imageInfo.minz && zval <= imageInfo.maxz {
			if grid[row][col] == nil {
				k := make([]float64, 0, 8)
				grid[row][col] = &k
			}
			*(grid[row][col]) = append(*(grid[row][col]), zval)
			totalZ += zval
		} else {
			errPoints++
		}
	}
	if errPoints > 0 {
		fmt.Printf("Warning: %d points were outside the bounding box defined in the header\n", errPoints)
	}
	t3 := time.Now()
	fmt.Printf("Placed points on grid in %v, avg z: %v\n", t3.Sub(t2), totalZ/float64(numPoints))

	totalAvgZ := float32(0.0)
	totalAvgZCount := float32(0.0)
	countNodata := 0
	major, minor := d.header.Version()
	legacy := false
	if major == 1 && minor < 3 {
		legacy = true
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if grid[r][c] != nil {
				sum := float32(0.0)
				cnt := float32(0.0)
				if d.opt != nil && d.opt.GatherClassifications {
					cv := 0.0
					for _, v := range *(grid[r][c]) {
						if (legacy && filterLegacyClassifications(uint16(math.Floor(v)))) ||
							(!legacy && filterClassifications(uint16(math.Floor(v)))) {
							if cv == 0.0 {
								cv = v
							} else {
								if cv != v {
									panic(fmt.Sprintf("Different classifications [%.2f, %.2f] exists for the same cell {%d, %d} all for cell: %v", cv, v, r, c, *(grid[r][c])))
								}
							}
						}
					}
					if cv == 0.0 {
						raster.SetValue(r, c, float32(-9999.0))
						countNodata += 1
					} else {
						raster.SetValue(r, c, float32(cv))
					}
				} else {
					// Average Z values over the grid cell
					for _, v := range *(grid[r][c]) {
						sum += float32(v)
						cnt += float32(1.0)
					}
					raster.SetValue(r, c, float32(sum/cnt))
					totalAvgZ += float32(sum / cnt)
					totalAvgZCount += float32(1.0)
				}
			} else {
				// Handle no data cells
				if d.opt != nil && d.opt.GatherClassifications {
					raster.SetValue(r, c, float32(9.0)) // nodata is water when pulling classifications
				} else {
					raster.SetValue(r, c, float32(-9999.0))
					countNodata += 1
				}
			}
		}
	}

	// Whether DEM or classification generation use sliding window algorithm to interpolate NODATA values

	countOfNoDataCells := countNoDataCells(raster)
	for countOfNoDataCells != 0 {

		rasterWidth := raster.Width()
		rasterLength := raster.Height()

		for row := 0; row < rasterLength; row++ {
			for col := 0; col < rasterWidth; col++ {
				if raster.ValueAt(row, col) == -9999 {
					raster.SetValue(row, col, interpolateCell(raster, row, col))
				}
			}
		}

		newNodataCellCount := countNoDataCells(raster)

		if countOfNoDataCells == countOfNoDataCells {
			break
		} else {
			countOfNoDataCells = newNodataCellCount
		}
	}

	fmt.Printf("raster avg z: %f, number of nodata points: %v\n", totalAvgZ/totalAvgZCount, countNodata)
	rStrings := make([]string, 0, len(values.r))
	for ind, v := range values.r {
		if v > 0 {
			rStrings = append(rStrings, fmt.Sprintf("%d: %s", ind, humanize.Comma(v)))
		}
	}
	fmt.Println("Returns:", strings.Join(rStrings, ", "))
	cStrings := make([]string, 0, len(values.c))
	for ind, v := range values.c {
		if v > 0 {
			if legacy {
				index := classificationLegacy(ind)
				cStrings = append(cStrings, fmt.Sprintf("%s: %s", legacyLookup[index], humanize.Comma(v)))
			} else {
				index := classification14(ind)
				cStrings = append(cStrings, fmt.Sprintf("%s: %s", classificationLookup[index], humanize.Comma(v)))
			}
		}
	}
	fmt.Println("Classifications:", strings.Join(cStrings, ", "))
	return raster, nil
}

type ReadOptions struct {
	Filtering             bool
	FirstReturns          bool
	BareEarthClass        bool
	LastReturns           bool
	GatherClassifications bool
	Intensity             bool
	FilterCrs             bool
	AcceptableGeoKeys     map[int]bool
}

func (opt *ReadOptions) String() string {
	return fmt.Sprintf("ReadOptions: Filtering: %v, FirstReturns: %v, BareEarthClass: %v, LastReturns: %v, Intensity: %v, GatherIngClassifications: %v, FilterCrs: %v, AcceptableGeoKeys: %v",
		opt.Filtering, opt.FirstReturns, opt.BareEarthClass, opt.LastReturns, opt.Intensity, opt.GatherClassifications, opt.FilterCrs, opt.AcceptableGeoKeys)
}

func validateOpt(opt *ReadOptions) error {
	if opt != nil && opt.Filtering {
		if !opt.GatherClassifications && !opt.FirstReturns && !opt.BareEarthClass && !opt.LastReturns && !opt.Intensity {
			return fmt.Errorf("Filtering is enabled without specifying any filter option")
		}
		if opt.BareEarthClass && opt.FirstReturns {
			return fmt.Errorf("FirstReturns and BareEarthClass can not both be TRUE")
		}
		if opt.BareEarthClass && opt.LastReturns {
			return fmt.Errorf("LastReturns and BareEarthClass can not both be TRUE")
		}
		if opt.FirstReturns && opt.LastReturns {
			return fmt.Errorf("LastReturns and FirstReturns can not both be TRUE")
		}
		if opt.FirstReturns && opt.Intensity {
			return fmt.Errorf("Intensity and FirstReturns can not both be TRUE")
		}
		if opt.GatherClassifications && opt.LastReturns {
			return fmt.Errorf("GatherClassifications and LastReturns can not both be TRUE")
		}
		if opt.GatherClassifications && opt.Intensity {
			return fmt.Errorf("GatherClassifications and LastReturns can not both be TRUE")
		}
		if opt.GatherClassifications && opt.BareEarthClass {
			return fmt.Errorf("GatherClassifications and BareEarthClass can not both be TRUE")
		}
		if opt.FilterCrs && (opt.AcceptableGeoKeys == nil || len(opt.AcceptableGeoKeys) == 0) {
			return fmt.Errorf("AcceptableGeoKeys must be specified when FilterCrs is TRUE")
		}
	} else if opt != nil {
		if opt.Intensity {
			return fmt.Errorf("Intensity is on without Filtering being enabled")
		}
		if opt.FirstReturns {
			return fmt.Errorf("FirstReturns is on without Filtering being enabled")
		}
		if opt.BareEarthClass {
			return fmt.Errorf("BareEarthClass is on without Filtering being enabled")
		}
		if opt.LastReturns {
			return fmt.Errorf("LastReturns is on without Filtering being enabled")
		}
		if opt.GatherClassifications {
			return fmt.Errorf("GatherClassifications is on without Filtering being enabled")
		}
		if opt.FilterCrs {
			return fmt.Errorf("FilterCrs is on without Filtering being enabled")
		}
	}
	fmt.Println("Processing using", opt)
	return nil
}

func NewFileReader(f *os.File, opt *ReadOptions) (Las, error) {
	return NewStreamReader(f, opt)
}

func NewStreamReader(r io.ReaderAt, opt *ReadOptions) (Las, error) {
	if err := validateOpt(opt); err != nil {
		return nil, err
	}
	signature := make([]byte, 4)
	if _, err := r.ReadAt(signature, 0); err != nil {
		return nil, err
	}
	if string(signature) == "LASF" {
		if _, err := r.ReadAt(signature[0:2], headerSizePosition); err != nil {
			return nil, err
		}
		headerSize := binary.LittleEndian.Uint16(signature[0:2])
		rawHeader := make([]byte, headerSize)
		r.ReadAt(rawHeader, 0)
		d := &decoder{reader: r, byteOrder: binary.LittleEndian, opt: opt}
		hdr := d.readLasHeader(rawHeader)
		d.header = hdr
		vlrPos := int64(headerSize)
		d.vlrs = make([]*Vlr, 0, hdr.GetNumberOfVLR())
		for i := uint32(0); i < hdr.GetNumberOfVLR(); i++ {
			p := make([]byte, 54)
			d.reader.ReadAt(p, vlrPos)
			v := &Vlr{userID: string(p[2:18]), recordID: d.byteOrder.Uint16(p[18:20]), lengthAfterHeader: d.byteOrder.Uint16(p[20:22]), description: string(p[22:54])}
			v.data = make([]byte, int64(v.lengthAfterHeader))
			d.reader.ReadAt(v.data, vlrPos+54)
			d.vlrs = append(d.vlrs, v)
			vlrPos += 54 + int64(v.lengthAfterHeader)
		}

		d.evlrs = make([]*Evlr, 0, hdr.GetNumberOfEVLR())
		evlrPos := hdr.GetOffsetOfEVLR()
		for i := uint32(0); i < hdr.GetNumberOfEVLR(); i++ {
			p := make([]byte, 60)
			d.reader.ReadAt(p, int64(evlrPos))
			v := &Evlr{userID: string(p[2:18]), recordID: d.byteOrder.Uint16(p[18:20]), lengthAfterHeader: d.byteOrder.Uint64(p[20:28]), description: string(p[28:60])}
			v.data = make([]byte, int64(v.lengthAfterHeader))
			d.reader.ReadAt(v.data, int64(evlrPos+60))
			d.evlrs = append(d.evlrs, v)
			evlrPos += 60 + v.lengthAfterHeader
		}
		d.parseCrsRecord()
		return d, nil
	}
	return nil, NotaLasFile(signature)
}
