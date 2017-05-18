// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geotiff

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/geo/s2"
	"golang.org/x/image/tiff"
	"golang.org/x/image/tiff/lzw"
)

type Tiff interface {
	TagFor(int) (*Ifd, error)
	KeyFor(int) (*GeoKey, error)
	IsGeotiff() bool
	DescribeGeokeys() ([]string, error)
	DescribeGeoKey(*GeoKey) string
	GetGeoKeyValue(int) (interface{}, error)
	DescribeTiffTags() []string
	Bounds() (*Bounds, error)
	Points() (*Raster, float32, float32, error)
	IsImage() bool
	GetImage() (image.Image, error)
	GetValueByLonLat(float64, float64, *Raster) (float32, error)
	ZoomLevel() (int64, error)
	Resolution() (float64, error)
	DateTime() (time.Time, error)
	DimensionMeters() (w float64, h float64, err error)
	PixelDimensions() (widthPixels float64, heightPixels float64, err error)
	Close() bool

	// Image ops must first be enabled with StoreImage()
	StoreImage() error
	PixelMax() (maxx, maxy int)
	AtPixelCoord(x, y int) color.Color
	RawPixel(x, y int) []uint8
	BaseZoom() int
	SourceBounds() *Bounds
	GetId() uint32
}

type NotaTiffFile []byte
type GeneralIssue string
type InsufficientBytes int32
type UnsupportedError string
type TagNotFound uint16

func (s NotaTiffFile) Error() string {
	b1 := s[0]
	b2 := s[1]
	b3 := s[2]
	b4 := s[3] // have to offload the bytes out of the type as Sprintf will attempt to format it and lead to a panic
	return fmt.Sprintf("Not a Tiff file, signature bytes are [%v, %v, %v, %v]", b1, b2, b3, b4)
}

func (i InsufficientBytes) Error() string {
	sz := int32(i)
	return "Passed byte array is too small, the first IFD Offset is at length: " + string(sz)
}

func (i UnsupportedError) Error() string {
	return fmt.Sprintf("UnsupportedError: %s\n", string(i))
}

func (i GeneralIssue) Error() string {
	return string(i)
}

func (i TagNotFound) Error() string {
	msg := uint16(i)
	if name, ok := TiffTags[int(msg)]; ok {
		return fmt.Sprintf("Geotiff: Tag %s is not in the IFD map", name)
	}
	return fmt.Sprintf("Geotiff: Tag [%v] is not in the IFD map", msg)
}

type Rational struct {
	numerator   uint32
	denominator uint32
}

type Srational struct {
	numerator   int32
	denominator int32
}

type Ifd struct {
	tag       uint16
	fieldType uint16
	count     uint32
	value     interface{}

	intValue    uint
	floatValue  float64
	intArray    []uint
	floatArray  []float64
	stringValue string
}

func (v *Ifd) TagName() string {
	name, ok := TiffTags[int(v.tag)]
	if ok {
		return name
	}
	return fmt.Sprintf("%v", v.tag)
}

func (v *Ifd) DataTypeName() string {
	name, ok := DataTypes[int(v.fieldType)]
	if ok {
		return name
	}
	return fmt.Sprintf("%v", v.fieldType)
}

func (v *Ifd) ValueName() (string, bool) {
	switch v.tag {
	case tCompression:
		name, ok := CompressionTypes[int(v.intValue)]
		return name, ok
	case tPhotometricInterpretation:
		name, ok := PhotoMetricInterpretation[int(v.intValue)]
		return name, ok
	case tResolutionUnit:
		name, ok := Resolutions[int(v.intValue)]
		return name, ok
	case tSampleFormat:
		name, ok := SampleFormatTypes[int(v.intValue)]
		return name, ok
	}
	return "", false
}

func (v *Ifd) setValue() {
	switch v.fieldType {
	case dtByte:
		data := v.value.([]byte)
		if v.count == 1 {
			v.intValue = uint(data[0])
		} else {
			a := make([]uint, len(data))
			for i, v := range data {
				a[i] = uint(v)
			}
			v.intArray = a
		}
	case dtShort:
		data := v.value.([]uint16)
		if v.count == 1 {
			v.intValue = uint(data[0])
		} else {
			a := make([]uint, len(data))
			for i, v := range data {
				a[i] = uint(v)
			}
			v.intArray = a
		}
	case dtLong:
		data := v.value.([]uint32)
		if v.count == 1 {
			v.intValue = uint(data[0])
		} else {
			a := make([]uint, len(data))
			for i, v := range data {
				a[i] = uint(v)
			}
			v.intArray = a
		}
	case dtDouble:
		data := v.value.([]float64)
		if v.count == 1 {
			v.floatValue = data[0]
		} else {
			a := make([]float64, len(data))
			for i, v := range data {
				a[i] = float64(v)
			}
			v.floatArray = a
		}
	case dtFloat:
		data := v.value.([]float32)
		if v.count == 1 {
			v.floatValue = float64(data[0])
		} else {
			a := make([]float64, len(data))
			for i, v := range data {
				a[i] = float64(v)
			}
			v.floatArray = a
		}
	case dtASCII:
		data := v.value.([]byte)
		v.stringValue = string(data)
	default:
	}
}

func (v *Ifd) Value() string {
	switch v.fieldType {
	case dtByte:
		if v.count < 16 {
			if v.count == 1 {
				return fmt.Sprintf("%v", v.intValue)
			}
			return fmt.Sprintf("%v", v.intArray)
		} else {
			return fmt.Sprintf("[%v]%v...", v.count, v.intArray[:16])
		}
	case dtShort:
		if v.count < 16 {
			if v.count == 1 {
				return fmt.Sprintf("%v", v.intValue)
			}
			return fmt.Sprintf("%v", v.intArray)
		} else {
			return fmt.Sprintf("[%v]%v...", v.count, v.intArray[:16])
		}
	case dtLong:
		if v.count < 10 {
			if v.count == 1 {
				return fmt.Sprintf("%v", v.intValue)
			}
			return fmt.Sprintf("%v", v.intArray)
		} else {
			return fmt.Sprintf("[%v]%v...", v.count, v.intArray[:10])
		}
	case dtDouble:
		if v.count < 10 {
			if v.count == 1 {
				return fmt.Sprintf("%f", v.floatValue)
			}
			return fmt.Sprintf("%f", v.floatArray)
		} else {
			return fmt.Sprintf("[%v]%f...", v.count, v.floatArray[:10])
		}
	case dtFloat:
		if v.count < 10 {
			if v.count == 1 {
				return fmt.Sprintf("%f", v.floatValue)
			} else {
				return fmt.Sprintf("%f", v.floatArray)
			}
		} else {
			return fmt.Sprintf("[%v]%f...", v.count, v.floatArray[:10])
		}
	case dtASCII:
		if v.count < 256 {
			return fmt.Sprintf("[%v]%s", v.count, v.stringValue)
		} else {
			return fmt.Sprintf("[%v]%s...", v.count, v.stringValue[:256])
		}
	default:
	}
	return ""
}

func (v *Ifd) PutData(byteOrder binary.ByteOrder, buf []byte) {
	switch v.fieldType {
	case dtByte, dtASCII:
		data := v.value.([]byte)
		copy(buf, data)
		buf = buf[len(data):]
	case dtShort:
		data := v.value.([]uint16)
		for _, e := range data {
			byteOrder.PutUint16(buf, e)
			buf = buf[2:]
		}
	case dtLong:
		data := v.value.([]uint32)
		for _, e := range data {
			byteOrder.PutUint32(buf, e)
			buf = buf[4:]
		}
	case dtRational:
		data := v.value.([]Rational)
		for _, e := range data {
			byteOrder.PutUint32(buf, e.numerator)
			byteOrder.PutUint32(buf, e.denominator)
			buf = buf[8:]
		}
	case dtDouble:
		data := v.value.([]float64)
		for _, e := range data {
			byteOrder.PutUint64(buf, math.Float64bits(e))
			buf = buf[8:]
		}
	case dtFloat:
		data := v.value.([]float32)
		for _, e := range data {
			byteOrder.PutUint32(buf, math.Float32bits(e))
			buf = buf[4:]
		}
	case dtSbyte:
		data := v.value.([]int8)
		for _, e := range data {
			buf[0] = byte(e)
			buf = buf[1:]
		}
	case dtSshort:
		data := v.value.([]int16)
		for _, e := range data {
			byteOrder.PutUint16(buf, uint16(e))
			buf = buf[2:]
		}
	case dtSlong:
		data := v.value.([]int32)
		for _, e := range data {
			byteOrder.PutUint32(buf, uint32(e))
			buf = buf[4:]
		}
	case dtSrational:
		data := v.value.([]Srational)
		for _, e := range data {
			byteOrder.PutUint32(buf, uint32(e.numerator))
			byteOrder.PutUint32(buf, uint32(e.denominator))
			buf = buf[8:]
		}
	}
}

type ByTag []*Ifd

func (a ByTag) Len() int           { return len(a) }
func (a ByTag) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTag) Less(i, j int) bool { return a[i].tag < a[j].tag }

type GeoKey struct {
	KeyId, Location, Count, Value uint16
}

type ByKey []*GeoKey

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].KeyId < a[j].KeyId }

// io.Reader is required for tiff.Decode() (Go x lib)
type TiffReader interface {
	io.Reader
	io.ReaderAt
}

type decoder struct {
	reader         TiffReader
	byteOrder      binary.ByteOrder
	ifdOffset      int64
	ifd            map[uint16]*Ifd
	geokeys        map[uint16]*GeoKey
	geoAsciis      []byte
	geoDoubles     []float64
	bounds         *Bounds
	minZ, maxZ     float64
	myImage        image.Image
	isImage        bool
	imageStored    bool
	latchPixLength sync.Once
	pixLength      int
	id             uint32
}

func GroundResolutionatZoom(lat float64, zoom int64) float64 {
	return (math.Cos(lat*(math.Pi/180.0)) * 2 * math.Pi * Wgs84SemiMajorAxis) / (256.0 * math.Pow(2.0, float64(zoom)))
}

func (d *decoder) Close() bool {
	if c, ok := d.reader.(io.Closer); ok {
		c.Close()
	}
	return true
}

func (d *decoder) PixelDimensions() (widthPixels, heightPixels float64, err error) {
	var w, h float64
	if t, err := d.TagFor(tImageWidth); err != nil {
		return 0, 0, err
	} else {
		w = float64(t.intValue)
	}
	if t, err := d.TagFor(tImageLength); err != nil {
		return 0, 0, err
	} else {
		h = float64(t.intValue)
	}
	return w, h, nil
}

func (d *decoder) DimensionMeters() (widthMeters float64, heightMeters float64, err error) {
	if bounds, err := d.Bounds(); err != nil {
		return 0.0, 0.0, err
	} else {
		centerX, centerY := bounds.Center()
		p1 := s2.PointFromLatLng(s2.LatLngFromDegrees(centerY, bounds.MinX))
		p2 := s2.PointFromLatLng(s2.LatLngFromDegrees(centerY, bounds.MaxX))
		angl := p1.Distance(p2)
		distanceX := angl.Radians() * Wgs84SemiMajorAxis // ground distance in meters
		p3 := s2.PointFromLatLng(s2.LatLngFromDegrees(centerX, bounds.MinY))
		p4 := s2.PointFromLatLng(s2.LatLngFromDegrees(centerX, bounds.MaxY))
		angl2 := p3.Distance(p4)
		distanceY := angl2.Radians() * Wgs84SemiMajorAxis // ground distance in meters
		return distanceX, distanceY, nil
	}
}

func (d *decoder) DateTime() (time.Time, error) {
	if dt, err := d.TagFor(tDateTime); err == nil {
		if tm, err1 := time.Parse("2006:01:02 15:04:05", dt.stringValue); err1 == nil {
			return tm, nil
		} else {
			return time.Now(), err1
		}
	} else {
		return time.Now(), err
	}
}

// meters/pixel
func (d *decoder) Resolution() (float64, error) {
	bounds, _ := d.Bounds()
	var w float64
	if t, err := d.TagFor(tImageWidth); err != nil {
		return 0, err
	} else {
		w = float64(t.intValue)
	}
	_, centerY := bounds.Center()
	distance := ApproxDistance(s2.LatLngFromDegrees(centerY, bounds.MinX), s2.LatLngFromDegrees(centerY, bounds.MaxX)) * 1000.0
	fmt.Println("Geotiff Distance", distance, "m Width", float64(w), "pixels")
	return distance / float64(w), nil
}

func (d *decoder) ZoomLevel() (int64, error) {
	bounds, _ := d.Bounds()
	var w float64
	if t, err := d.TagFor(tImageWidth); err != nil {
		return 0, err
	} else {
		w = float64(t.intValue)
	}
	_, centerY := bounds.Center()
	distance := ApproxDistance(s2.LatLngFromDegrees(centerY, bounds.MinX), s2.LatLngFromDegrees(centerY, bounds.MaxX)) * 1000.0
	res := distance / float64(w) // meters/pixel
	for i := int64(1); i <= 30; i++ {
		groundResolution := GroundResolutionatZoom(centerY, i)
		if res >= groundResolution {
			return int64(i), nil
		}
	}
	return 30, nil
}

func (d *decoder) GetGeoKeyValue(k int) (interface{}, error) {
	if geokey, err := d.KeyFor(k); err == nil {
		return ValueForKey(int(geokey.KeyId), int(geokey.Location), int(geokey.Value), int(geokey.Count), d.geoDoubles, d.geoAsciis), nil
	} else {
		return nil, err
	}
}

func (d *decoder) DescribeGeoKey(geokey *GeoKey) string {
	name := NameForKey(int(geokey.KeyId))
	var location string
	switch int(geokey.Location) {
	case tGeoAscii:
		location = "AsciiValues"
	case tGeoDoubles:
		location = "DoubleValues"
	default:
		location = "0"
	}

	value := ValueForKey(int(geokey.KeyId), int(geokey.Location), int(geokey.Value), int(geokey.Count), d.geoDoubles, d.geoAsciis)
	return fmt.Sprintf("%s: location: %s count: %v  %v", name, location, geokey.Count, value)
}

func clip(n, minValue, maxValue float64) float64 {
	return math.Min(math.Max(n, minValue), maxValue)
}

func tileXYToLatLng(z, x, y, px, py uint) (float64, float64) {
	mapSize := float64(uint(256) << z)
	pixelX := float64((x * 256) + px)
	pixelY := float64((y * 256) + py)
	x1 := (clip(pixelX, 0.0, mapSize-1) / mapSize) - 0.5
	y1 := 0.5 - (clip(pixelY, 0.0, mapSize-1) / mapSize)

	latitude := 90 - 360*math.Atan(math.Exp(-y1*math.Pi*2))/math.Pi
	longitude := 360 * x1
	return latitude, longitude
}

func (d *decoder) GetValueByLonLat(lon float64, lat float64, raster *Raster) (float32, error) {
	var xIndex int
	var yIndex int
	bounds, _ := d.Bounds()
	xDiff := bounds.MaxX - bounds.MinX
	yDiff := bounds.MaxY - bounds.MinY

	xInc := xDiff / float64(raster.Width())
	yInc := yDiff / float64(raster.Height())

	xD := lon - bounds.MinX
	yD := bounds.MaxY - lat
	xIndex = int(xD / xInc)
	yIndex = int(yD / yInc)
	if xIndex >= 0 && xIndex < raster.Width() && yIndex >= 0 && yIndex < raster.Height() {
		return raster.ValueAt(yIndex, xIndex), nil
	} else {
		return 0.0, errors.New("Point not inside bounds")
	}
}

func (d *decoder) parseIfd(p []byte) error {
	var raw []byte
	tag := d.byteOrder.Uint16(p[0:2])

	datatype := d.byteOrder.Uint16(p[2:4])
	count := d.byteOrder.Uint32(p[4:8])
	if dt := int(datatype); dt <= 0 || dt >= len(lengths) {
		return UnsupportedError(fmt.Sprintf("IFD entry datatype: %v", datatype))
	}
	if datalen := lengths[datatype] * count; datalen > 4 {
		// The IFD contains a pointer to the real value.
		raw = make([]byte, datalen)
		if _, e := d.reader.ReadAt(raw, int64(d.byteOrder.Uint32(p[8:12]))); e != nil {
			return e
		}
	} else {
		raw = p[8 : 8+count]
	}

	switch datatype {
	case dtByte:
		u := make([]byte, count)
		for i := uint32(0); i < count; i++ {
			u[i] = raw[i]
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	case dtShort:
		u := make([]uint16, count)
		for i := uint32(0); i < count; i++ {
			u[i] = d.byteOrder.Uint16(raw[2*i : 2*(i+1)])
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	case dtLong:
		u := make([]uint32, count)
		for i := uint32(0); i < count; i++ {
			u[i] = d.byteOrder.Uint32(raw[4*i : 4*(i+1)])
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	case dtASCII:
		u := make([]byte, count)
		copy(u, raw)
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	case dtSbyte:
		u := make([]int8, count)
		for i := uint32(0); i < count; i++ {
			u[i] = int8(raw[i])
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	case dtSshort:
		u := make([]int16, count)
		for i := uint32(0); i < count; i++ {
			u[i] = int16(d.byteOrder.Uint16(raw[2*i : 2*(i+1)]))
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	case dtSlong:
		u := make([]int32, count)
		for i := uint32(0); i < count; i++ {
			u[i] = int32(d.byteOrder.Uint32(raw[4*i : 4*(i+1)]))
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	case dtFloat:
		u := make([]float32, count)
		scratch := make([]uint32, count)
		for i := uint32(0); i < count; i++ {
			scratch[i] = d.byteOrder.Uint32(raw[4*i : 4*(i+1)])
			u[i] = math.Float32frombits(scratch[i])
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}

	case dtRational:
		u := make([]Rational, count)
		for i := uint32(0); i < count; i++ {
			num := d.byteOrder.Uint32(raw[8*i : (8*i)+4])
			denom := d.byteOrder.Uint32(raw[(8*i)+4 : (8*i)+8])
			u[i] = Rational{num, denom}
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	case dtSrational:
		u := make([]Srational, count)
		for i := uint32(0); i < count; i++ {
			num := int32(d.byteOrder.Uint32(raw[8*i : (8*i)+4]))
			denom := int32(d.byteOrder.Uint32(raw[(8*i)+4 : (8*i)+8]))
			u[i] = Srational{num, denom}
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	case dtDouble:
		u := make([]float64, count)
		scratch := make([]uint64, count)
		for i := uint32(0); i < count; i++ {
			scratch[i] = d.byteOrder.Uint64(raw[8*i : 8*(i+1)])
			u[i] = math.Float64frombits(scratch[i])
		}
		d.ifd[tag] = &Ifd{tag: tag, fieldType: datatype, count: count, value: u}
	default:
		return UnsupportedError(fmt.Sprintf("unsupported data type [%v]", datatype))
	}
	return nil
}

func (d *decoder) TagFor(tag int) (*Ifd, error) {
	ret, ok := d.ifd[uint16(tag)]
	if ok {
		return ret, nil
	}
	return nil, TagNotFound(tag)
}

func (d *decoder) IntegerValue(tag int, defaultValue ...uint) (uint, error) {
	ifd, err := d.TagFor(tag)
	if err == nil {
		return ifd.intValue, nil
	}
	if defaultValue != nil && len(defaultValue) > 0 {
		return defaultValue[0], nil
	}
	return 0, err
}

func (d *decoder) FloatValue(tag int, defaultValue ...float64) (float64, error) {
	ifd, err := d.TagFor(tag)
	if err == nil {
		return ifd.floatValue, nil
	}
	if defaultValue != nil && len(defaultValue) > 0 {
		return defaultValue[0], nil
	}
	return 0.0, err
}

func (d *decoder) IntegerArrayValue(tag int) ([]uint, error) {
	ifd, err := d.TagFor(tag)
	if err == nil {
		return ifd.intArray, nil
	}
	return []uint{}, err
}

func (d *decoder) FloatArrayValue(tag int) ([]float64, error) {
	ifd, err := d.TagFor(tag)
	if err == nil {
		return ifd.floatArray, nil
	}
	return []float64{}, err
}

func (d *decoder) StringValue(tag int) (string, error) {
	ifd, err := d.TagFor(tag)
	if err == nil {
		return ifd.stringValue, nil
	}
	return "", err

}

func (d *decoder) KeyFor(key int) (*GeoKey, error) {
	ret, ok := d.geokeys[uint16(key)]
	if ok {
		return ret, nil
	}
	return nil, GeneralIssue(fmt.Sprintf("GeoKey for %v is not present", key))
}

func (d *decoder) IsGeotiff() bool {
	if _, err := d.TagFor(tGeoKeys); err == nil {
		return true
	}
	return false
}

func (d *decoder) parseGeoKeys() error {
	geoKeys, err := d.IntegerArrayValue(tGeoKeys)
	if err != nil {
		return err
	}
	numberOfKeys := geoKeys[3]
	for i := 0; i < int(numberOfKeys); i++ {
		pos := 4 + (i * 4)
		keyId := geoKeys[pos]
		location := geoKeys[pos+1]
		count := geoKeys[pos+2]
		offset := geoKeys[pos+3]
		d.geokeys[uint16(keyId)] = &GeoKey{uint16(keyId), uint16(location), uint16(count), uint16(offset)}
	}
	// Not an error to be missing either of these optional tags
	// though they should not be referenced within geokeys if omitted
	geoDoubles, err := d.FloatArrayValue(tGeoDoubles)
	if err == nil {
		d.geoDoubles = geoDoubles
	}
	tg, err := d.TagFor(tGeoAscii)
	if err == nil {
		data := tg.value.([]byte)
		d.geoAsciis = data
	}
	return nil
}

func (d *decoder) DescribeGeokeys() ([]string, error) {
	if d.IsGeotiff() {
		retval := make([]string, 0, len(d.geokeys))
		keys := make([]*GeoKey, 0, len(d.geokeys))
		for _, v := range d.geokeys {
			keys = append(keys, v)
		}
		sort.Sort(ByKey(keys))
		for _, v := range keys {
			retval = append(retval, d.DescribeGeoKey(v))
		}
		return retval, nil
	}
	return nil, GeneralIssue("Not a geotiff")
}

func (d *decoder) DescribeTiffTags() []string {
	retval := make([]string, 0, len(d.ifd))
	tags := make([]*Ifd, 0, len(d.ifd))
	for _, v := range d.ifd {
		tags = append(tags, v)
	}
	sort.Sort(ByTag(tags))
	for _, v := range tags {
		vn, ok := v.ValueName()
		if ok {
			retval = append(retval, fmt.Sprintf("%s [%s]: %s", v.TagName(), v.DataTypeName(), vn))
		} else {
			retval = append(retval, fmt.Sprintf("%s [%s]: %v", v.TagName(), v.DataTypeName(), v.Value()))
		}
	}
	return retval
}

type Raster struct {
	Data []float32
	w, h int
}

func (r *Raster) ValueAt(row, col int) float32 {
	return r.Data[row*r.w+col]
}

func (r *Raster) SetValue(row, col int, value float32) {
	r.Data[row*r.w+col] = value
}

func (r *Raster) Size() int {
	return binary.Size(r.Data) + binary.Size(r.w) + binary.Size(r.h)
}

func (r *Raster) Width() int {
	return r.w
}

func (r *Raster) Height() int {
	return r.h
}

func NewRaster(width, height int) *Raster {
	return &Raster{
		w:    width,
		h:    height,
		Data: make([]float32, width*height),
	}
}

type ImageInfo struct {
	width, height, rowsPerStrip, bitsPerSample, sampleFormat, samplesPerPixel, compression uint
	nodata                                                                                 float32
	writeRow                                                                               int
}

type TileImageInfo struct {
	width, height, tilesAcross, tilesDown, tileWidth, tileLength, bitsPerSample, sampleFormat, samplesPerPixel, compression uint
	nodata                                                                                                                  float32
	writeRow                                                                                                                int
}

func (d *decoder) readBlock(blockData []byte, row uint, imageInfo *ImageInfo, raster *Raster) *Raster {
	w := imageInfo.width
	h := imageInfo.height
	writeRow := int(row)
	rowsPerStrip := imageInfo.rowsPerStrip
	block := rowsPerStrip
	if (h - row) < rowsPerStrip {
		block = h - row
	}
	writeCol := 0
	for rowi := uint(0); rowi < block; rowi++ {
		for i := rowi * w * 4; i < (rowi+1)*w*4; i += 4 {
			zval := math.Float32frombits(d.byteOrder.Uint32(blockData[i : i+4]))
			if zval <= imageInfo.nodata {
				raster.SetValue(writeRow, writeCol, -9999.0)
				writeCol++
			} else {
				d.minZ = math.Min(d.minZ, float64(zval))
				d.maxZ = math.Max(d.maxZ, float64(zval))
				raster.SetValue(writeRow, writeCol, zval)
				writeCol++
			}
		}
		writeRow++
		writeCol = 0
	}
	return raster
}

func uintMin(m1, m2 uint) uint {
	if m1 <= m2 {
		return m1
	}
	return m2
}

func uintMax(m1, m2 uint) uint {
	if m1 >= m2 {
		return m1
	}
	return m2
}

func (d *decoder) readTileBlock(blockData []byte, tileRow, tileCol uint, imageInfo *TileImageInfo, raster *Raster) *Raster {
	w := imageInfo.width
	h := imageInfo.height
	var rmax = int(uintMin((imageInfo.tilesDown * imageInfo.tileLength), h))
	var cmax = int(uintMin((imageInfo.tilesAcross * imageInfo.tileWidth), w))

	x, y := (tileCol * imageInfo.tileWidth), (tileRow * imageInfo.tileLength)

	for tRow := 0; tRow < int(imageInfo.tileLength); tRow++ {
		py := int(y) + tRow
		px := int(x)
		for tCol := 0; tCol < int(imageInfo.tileWidth*4); tCol += 4 {
			i := tRow*int(imageInfo.tileWidth*4) + tCol
			zval := math.Float32frombits(d.byteOrder.Uint32(blockData[i : i+4]))

			// Tiles can exceed the image width and height, they can be padded
			// in either direction, we test to make sure we are writing within the raster
			if px < cmax && py < rmax {
				if zval <= imageInfo.nodata {
					raster.SetValue(py, px, float32(-9999.0))
				} else {
					d.minZ = math.Min(d.minZ, float64(zval))
					d.maxZ = math.Max(d.maxZ, float64(zval))
					raster.SetValue(py, px, zval)
				}
			}
			px++
		}
	}
	return raster
}

func (d *decoder) TilePoints() (*Raster, float32, float32, error) {
	imageWidth, e1 := d.IntegerValue(tImageWidth)
	imageLength, e2 := d.IntegerValue(tImageLength)
	samplesPerPixel, e3 := d.IntegerValue(tSamplesPerPixel)
	tileWidth, e4 := d.IntegerValue(tTileWidth)
	tileLength, e5 := d.IntegerValue(tTileLength)
	tileByteCounts, e6 := d.IntegerArrayValue(tTileByteCounts)
	tileOffsets, e7 := d.IntegerArrayValue(tTileOffsets)
	bitsPerSample, e8 := d.IntegerValue(tBitsPerSample)
	sampleFormat, e9 := d.IntegerValue(tSampleFormat, uint(1))
	compression, e10 := d.IntegerValue(tCompression, uint(1))
	nodata, e11 := d.FloatValue(tGDALNodata, float64(-9999.0))
	if e := checkFailure(e1, e2, e3, e4, e5, e6, e7, e8, e9, e10, e11); e != nil {
		return nil, 0, 0, e
	}
	tilesAcross := (imageWidth + tileWidth - 1) / tileWidth
	tilesDown := (imageLength + tileLength - 1) / tileLength
	fmt.Printf("Tiles are arranged %d x %d (acrossxdown)\n", tilesAcross, tilesDown)
	imageInfo := &TileImageInfo{imageWidth, imageLength, tilesAcross, tilesDown, tileWidth, tileLength, bitsPerSample, sampleFormat, samplesPerPixel, compression, float32(nodata), 0}
	d.minZ = 9999.0
	d.maxZ = -9999.0
	// TODO: Add other sample formats, such as RGB, return type of Raster would need to change
	// or we bit pack RGBA into float32
	if samplesPerPixel == uint(1) && bitsPerSample == uint(32) && sampleFormat == uint(3) {
		raster := NewRaster(int(imageWidth), int(imageLength))
		for td := uint(0); td < tilesDown; td++ {
			for ta := uint(0); ta < tilesAcross; ta++ {
				tile := (td * tilesAcross) + ta
				p := make([]byte, int(tileByteCounts[tile]))

				if _, err := d.reader.ReadAt(p, int64(tileOffsets[tile])); err != nil {
					return nil, 0, 0, err
				}

				switch int(compression) {
				case cNone:
					d.readTileBlock(p, td, ta, imageInfo, raster)
				case cLZW:
					dataReader := lzw.NewReader(bytes.NewReader(p), lzw.MSB, 8)
					if p1, err := ioutil.ReadAll(dataReader); err == nil {
						d.readTileBlock(p1, td, ta, imageInfo, raster)
					} else {
						return nil, 0, 0, err
					}
				case cDeflate:
					if dataReader, e := zlib.NewReader(bytes.NewReader(p)); e == nil {
						if p1, err := ioutil.ReadAll(dataReader); err == nil {
							d.readTileBlock(p1, td, ta, imageInfo, raster)
						} else {
							return nil, 0, 0, err
						}
					} else {
						return nil, 0, 0, e
					}
				}
			}
		}
		return raster, float32(d.minZ), float32(d.maxZ), nil
	}
	return nil, float32(0), float32(0), fmt.Errorf("Error reading tile")
}

func (d *decoder) Points() (*Raster, float32, float32, error) {
	_, e0 := d.IntegerValue(tTileWidth)
	if e := checkFailure(e0); e == nil {
		return d.TilePoints()
	}
	imageWidth, e1 := d.IntegerValue(tImageWidth)
	imageLength, e2 := d.IntegerValue(tImageLength)
	samplesPerPixel, e3 := d.IntegerValue(tSamplesPerPixel)
	stripOffsets, e4 := d.IntegerArrayValue(tStripOffsets)
	stripByteCounts, e5 := d.IntegerArrayValue(tStripByteCounts)
	bitsPerSample, e6 := d.IntegerValue(tBitsPerSample)
	rowsPerStrip, e7 := d.IntegerValue(tRowsPerStrip)
	sampleFormat, e8 := d.IntegerValue(tSampleFormat, uint(1))
	compression, e9 := d.IntegerValue(tCompression, uint(1))
	nodata, e10 := d.FloatValue(tGDALNodata, float64(-9999.0))
	if e := checkFailure(e1, e2, e3, e4, e5, e6, e7, e8, e9, e10); e != nil {
		return nil, 0, 0, e
	}
	imageInfo := &ImageInfo{imageWidth, imageLength, rowsPerStrip, bitsPerSample, sampleFormat, samplesPerPixel, compression, float32(nodata), 0}
	d.minZ = 9999.0
	d.maxZ = -9999.0
	if samplesPerPixel == uint(1) && bitsPerSample == uint(32) && sampleFormat == uint(3) {
		raster := NewRaster(int(imageWidth), int(imageLength))
		switch int(compression) {
		case cNone:
			for row := uint(0); row < imageLength; row += rowsPerStrip {
				p := make([]byte, int(stripByteCounts[row/rowsPerStrip]))
				if _, err := d.reader.ReadAt(p, int64(stripOffsets[row/rowsPerStrip])); err != nil {
					return nil, 0, 0, err
				}
				d.readBlock(p, row, imageInfo, raster)
			}
		case cLZW:
			for row := uint(0); row < imageLength; row += rowsPerStrip {
				p := make([]byte, int(stripByteCounts[row/rowsPerStrip]))
				d.reader.ReadAt(p, int64(stripOffsets[row/rowsPerStrip]))
				dataReader := lzw.NewReader(bytes.NewReader(p), lzw.MSB, 8)
				if p1, err := ioutil.ReadAll(dataReader); err == nil {
					d.readBlock(p1, row, imageInfo, raster)
				} else {
					return nil, 0, 0, err
				}
			}
		case cDeflate:
			for row := uint(0); row < imageLength; row += rowsPerStrip {
				p := make([]byte, int(stripByteCounts[row/rowsPerStrip]))
				d.reader.ReadAt(p, int64(stripOffsets[row/rowsPerStrip]))
				if dataReader, e := zlib.NewReader(bytes.NewReader(p)); e == nil {
					if p1, err := ioutil.ReadAll(dataReader); err == nil {
						d.readBlock(p1, row, imageInfo, raster)
					} else {
						return nil, 0, 0, err
					}
				} else {
					return nil, 0, 0, e
				}
			}
		}
		return raster, float32(d.minZ), float32(d.maxZ), nil
	}
	return nil, 0, 0, GeneralIssue(fmt.Sprintf("data is not in float32 format: samplesPerPixel=%v, bitsPerSample=%v, sampleFormat=%v\n", samplesPerPixel, bitsPerSample, sampleFormat))
}
func (d *decoder) IsImage() bool {
	sampleFormat, e8 := d.IntegerValue(tSampleFormat, uint(1))
	photoMetricInterpretation, e9 := d.IntegerValue(tPhotometricInterpretation)
	if e := checkFailure(e8, e9); e != nil {
		return false
	} else {
		if sampleFormat == smplUint && photoMetricInterpretation == pRGB {
			return true
		}
	}
	return false
}

func (d *decoder) GetImage() (image.Image, error) {
	if d.imageStored {
		return d.myImage, nil
	} else {
		if err := d.StoreImage(); err != nil {
			return nil, err
		} else {
			return d.myImage, nil
		}
	}
}

func (d *decoder) StoreImage() error {
	if d.isImage {
		if !d.imageStored {
			if img, err := tiff.Decode(d.reader); err != nil {
				return err
			} else {
				d.myImage = img
				d.imageStored = true
				return nil
			}
		}
		return nil
	}
	return fmt.Errorf("Not an image")
}

// PixelSource interface methods (ingester.PixelSource)
func (d *decoder) PixelMax() (maxx, maxy int) {
	if d.imageStored {
		bounds := d.myImage.Bounds()
		return bounds.Dx(), bounds.Dy()
	}
	panic("Image not yet stored, call StoreImage() first")
}

func (d *decoder) RawPixel(x, y int) []uint8 {
	if d.imageStored {
		if nrgba, ok := d.myImage.(*image.NRGBA); ok {
			index := nrgba.PixOffset(x, y)
			return nrgba.Pix[index : index+4]
		}
		if rgba, ok := d.myImage.(*image.RGBA); ok {
			index := rgba.PixOffset(x, y)
			return rgba.Pix[index : index+4]
		}
		panic("Unsupported image format for RawPixel")
	}
	panic("Image not yet stored, call StoreImage() first")
}

func (d *decoder) AtPixelCoord(x, y int) color.Color {
	if d.imageStored {
		return d.myImage.At(x, y)
	}
	panic("Image not yet stored, call StoreImage() first")
}

func (d *decoder) BaseZoom() int {
	zoom, err := d.ZoomLevel()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return int(zoom)
}

func (d *decoder) SourceBounds() *Bounds {
	b, err := d.Bounds()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}

func checkFailure(errors ...error) error {
	for _, e := range errors {
		if e != nil {
			return e
		}
	}
	return nil
}

func (d *decoder) IsMercator() bool {
	// TODO
	return false
}

func (d *decoder) Bounds() (*Bounds, error) {
	if d.bounds != nil {
		return d.bounds, nil
	}
	modelTiepoint, e1 := d.FloatArrayValue(tModelTiepointTag)
	modelScale, e2 := d.FloatArrayValue(tModelPixelScaleTag)
	imageWidth, e3 := d.IntegerValue(tImageWidth)
	imageLength, e4 := d.IntegerValue(tImageLength)
	orientation, e5 := d.IntegerValue(tOrientation, uint(1)) // provide a default value
	if e := checkFailure(e1, e2, e3, e4, e5); e != nil {
		return nil, e
	}

	x := modelTiepoint[3]
	y := modelTiepoint[4]
	xPixels := modelScale[0] * float64(imageWidth)
	yPixels := modelScale[1] * float64(imageLength)
	switch orientation {
	case 1: // y=top, x=left
		d.bounds = &Bounds{MinX: x, MaxY: y, MaxX: x + xPixels, MinY: y - yPixels, OriginX: x, OriginY: y}
	case 2: // y=top, x=right
		d.bounds = &Bounds{MaxX: x, MaxY: y, MinX: x - xPixels, MinY: y - yPixels, OriginX: x, OriginY: y}
	case 3: // y=bottom x=right
		d.bounds = &Bounds{MaxX: x, MinY: y, MinX: x - xPixels, MaxY: y + yPixels, OriginX: x, OriginY: y}
	case 4: // y=bottom x=left
		d.bounds = &Bounds{MinX: x, MinY: y, MaxX: x + xPixels, MaxY: y + yPixels, OriginX: x, OriginY: y}
	case 5: // y=left x=top
		d.bounds = &Bounds{MinX: y, MaxY: x, MaxX: y + yPixels, MinY: x - xPixels, OriginX: x, OriginY: y}
	case 6: // y=right x=top
		d.bounds = &Bounds{MaxX: y, MaxY: x, MinX: y - yPixels, MinY: x - xPixels, OriginX: x, OriginY: y}
	case 7: // y=right x=bottom
		d.bounds = &Bounds{MaxX: y, MinY: x, MinX: y - yPixels, MaxY: x + xPixels, OriginX: x, OriginY: y}
	case 8: // y=left x=bottom
		d.bounds = &Bounds{MinX: y, MinY: x, MaxX: y + yPixels, MaxY: x + xPixels, OriginX: x, OriginY: y}
	default: // y=top, x=left
		d.bounds = &Bounds{MinX: x, MaxY: y, MaxX: x + xPixels, MinY: y - yPixels, OriginX: x, OriginY: y}
	}
	return d.bounds, nil
}

func (d *decoder) GetId() uint32 {
	return d.id
}

var LittleSignature []byte = []byte{0x49, 0x49, 0x2a, 0x00}
var BigSignature []byte = []byte{0x4d, 0x4d, 0x0, 0x2a}

func isTiff(b []byte) (binary.ByteOrder, error) {
	if bytes.Equal(b, LittleSignature) {
		return binary.LittleEndian, nil
	} else if bytes.Equal(b, BigSignature) {
		return binary.BigEndian, nil
	}
	return nil, NotaTiffFile(b)
}

var counter uint32

// NewDecoder sets up a file Tiff to supply all tif tags, geo keys and Points as *Raster.
// The points Raster is not created until Tiff.Points() is executed.
// All tags and geo keys are loaded when this method returns Tiff (with error == nil)
func NewDecoder(f TiffReader) (Tiff, error) {
	header := make([]byte, 8)
	if _, err := f.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("Error reading header %v", err)
	}
	endian, err := isTiff(header[:4])
	if err != nil {
		return nil, err
	}

	d := &decoder{
		reader:    f,
		byteOrder: endian,
		ifdOffset: int64(endian.Uint32(header[4:])),
		ifd:       make(map[uint16]*Ifd),
		geokeys:   make(map[uint16]*GeoKey),
		minZ:      9999.0,
		maxZ:      -9999.0,
		id:        atomic.AddUint32(&counter, 1),
	}
	if _, err := d.reader.ReadAt(header[0:2], d.ifdOffset); err != nil {
		return nil, fmt.Errorf("Error reading offset %d: %v", d.ifdOffset, err)
	}
	numItems := d.byteOrder.Uint16(header[0:2])
	p := make([]byte, numItems*ifdLen)
	if _, err := d.reader.ReadAt(p, d.ifdOffset+2); err != nil {
		return nil, fmt.Errorf("Error reading tags %v", err)
	}
	for i := 0; i < len(p); i += ifdLen {
		if err := d.parseIfd(p[i : i+ifdLen]); err != nil {
			return nil, fmt.Errorf("Error parsing IFD tag %d, %v", d.byteOrder.Uint16(p[i:i+2]), err)
		}
	}
	for _, v := range d.ifd {
		v.setValue()
	}
	if err := d.parseGeoKeys(); err != nil {
		return nil, fmt.Errorf("Error parsing GeoKeys %v", err)
	}
	d.isImage = d.IsImage()
	return d, nil
}

func NewFileDecoder(f *os.File) (Tiff, error) {
	return NewDecoder(f)
}

// Quick approximate distance on WGS84 Ellipsoid, Note: Approx is very very good for most usecases
// Usually more accurate than great circle calcs
// Returns in kilometers
// If you desire miles, copy this method and alter D to miles
func ApproxDistance(p1, p2 s2.LatLng) float64 {
	p1.Lng *= -1.0
	p2.Lng *= -1.0

	F := (p1.Lat.Degrees() + p2.Lat.Degrees()) / 2.0
	G := (p1.Lat.Degrees() - p2.Lat.Degrees()) / 2.0
	L := (p1.Lng.Degrees() - p2.Lng.Degrees()) / 2.0

	cosg := math.Cos(G * DE2RA)
	sing := math.Sin(G * DE2RA)
	cosl := math.Cos(L * DE2RA)
	sinl := math.Sin(L * DE2RA)
	cosf := math.Cos(F * DE2RA)
	sinf := math.Sin(F * DE2RA)

	S := sing*sing*cosl*cosl + cosf*cosf*sinl*sinl
	C := cosg*cosg*cosl*cosl + sinf*sinf*sinl*sinl
	W := math.Atan2(math.Sqrt(S), math.Sqrt(C))
	R := math.Sqrt((S * C)) / W
	H1 := (3*R - 1.0) / (2.0 * C)
	H2 := (3*R + 1.0) / (2.0 * S)
	D := 2 * W * ERAD // ERAD is in Km
	return D * (1 + FLATTENING*H1*sinf*sinf*cosg*cosg - FLATTENING*H2*cosf*cosf*sing*sing)
}
