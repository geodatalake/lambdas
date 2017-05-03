// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geotiff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/golang/geo/s2"
)

// Bounding box in degrees lat, lng

type Bounds struct {
	MinX    float64
	MinY    float64
	MaxX    float64
	MaxY    float64
	OriginX float64
	OriginY float64
}

func (b *Bounds) Center() (centerX float64, centerY float64) {
	return (b.MinX + b.MaxX) / 2.0, (b.MinY + b.MaxY) / 2.0
}

func (b *Bounds) Contains(lon, lat float64) bool {
	return lat >= b.MinY && lat <= b.MaxY && lon >= b.MinX && lon <= b.MaxX
}

func (b *Bounds) UpperLeft() s2.LatLng {
	return s2.LatLngFromDegrees(b.MaxY, b.MinX)
}

func (b *Bounds) UpperRight() s2.LatLng {
	return s2.LatLngFromDegrees(b.MaxY, b.MaxX)
}

func (b *Bounds) LowerLeft() s2.LatLng {
	return s2.LatLngFromDegrees(b.MinY, b.MinX)
}

func (b *Bounds) LowerRight() s2.LatLng {
	return s2.LatLngFromDegrees(b.MinY, b.MaxX)
}

func (b *Bounds) CenterLocation() s2.LatLng {
	return s2.LatLngFromDegrees((b.MinY+b.MaxY)/2.0, (b.MinX+b.MaxX)/2.0)
}

const (
	UpperLeftQuadrant = iota
	UpperRightQuadrant
	LowerRightQuadrant
	LowerLeftQuadrant
)

// Break the Bounds into 4 quadrants
//
// 0 = Upper Left   /--------|---------\
// 1 = Upper Right  |   0    |    1    |
// 2 = Lower Right  |--------+---------|
// 3 = Lower Left   |   3    |    2    |
//                  \--------|---------/
func (b *Bounds) Quadrant(i int) *Bounds {
	centerX, centerY := b.Center()
	switch i {
	case UpperLeftQuadrant:
		return &Bounds{MinX: b.MinX, MaxX: centerX, MinY: centerY, MaxY: b.MaxY, OriginX: b.MinX, OriginY: b.MaxY}
	case UpperRightQuadrant:
		return &Bounds{MinX: centerX, MaxX: b.MaxX, MinY: centerY, MaxY: b.MaxY, OriginX: centerX, OriginY: b.MaxY}
	case LowerRightQuadrant:
		return &Bounds{MinX: centerX, MaxX: b.MaxX, MinY: b.MinY, MaxY: centerY, OriginX: centerX, OriginY: centerY}
	case LowerLeftQuadrant:
		return &Bounds{MinX: b.MinX, MaxX: centerX, MinY: b.MinY, MaxY: centerY, OriginX: b.MinX, OriginY: centerY}
	default:
		panic(fmt.Sprintf("Requested quadrant [%d] is out of range [0-3]", i))
	}
}

func (b *Bounds) Xspan() float64 {
	return b.MaxX - b.MinX
}

func (b *Bounds) Yspan() float64 {
	return b.MaxY - b.MinY
}

func (b *Bounds) AsWkt() string {
	return fmt.Sprintf("POLYGON ((%.7f %7f, %.7f %7f, %.7f %7f, %.7f %7f, %.7f %7f))",
		b.MinX, b.MaxY,
		b.MaxX, b.MaxY,
		b.MaxX, b.MinY,
		b.MinX, b.MinY,
		b.MinX, b.MaxY)
}

func MarshalBool(writer io.Writer, value bool) {
	var v uint16
	if value {
		v = uint16(1)
	}
	binary.Write(writer, binary.BigEndian, v)
}

func UnmarshalBool(reader io.Reader, value *bool) error {
	var bVal uint16
	if err := binary.Read(reader, binary.BigEndian, &bVal); err != nil {
		return err
	}
	if bVal == uint16(1) {
		*value = true
	} else {
		*value = false
	}
	return nil
}

func (b *Bounds) WriteTo(writer io.Writer) (int64, error) {
	binary.Write(writer, binary.BigEndian, uint16(0xbd))
	binary.Write(writer, binary.BigEndian, b.MaxX)
	binary.Write(writer, binary.BigEndian, b.MaxX)
	binary.Write(writer, binary.BigEndian, b.MinY)
	binary.Write(writer, binary.BigEndian, b.MaxY)
	binary.Write(writer, binary.BigEndian, b.OriginX)
	binary.Write(writer, binary.BigEndian, b.OriginY)
	return 52, nil
}

func (b *Bounds) ReadFrom(reader io.Reader) (int64, error) {
	var signature uint16
	var err error
	if err = binary.Read(reader, binary.BigEndian, &signature); err != nil {
		return -1, err
	}
	if signature != uint16(0xbd) {
		return -1, fmt.Errorf("Error reading Bounds from binary, signature: %x", signature)
	}
	if err = binary.Read(reader, binary.BigEndian, &b.MinX); err != nil {
		return -1, err
	}
	if err = binary.Read(reader, binary.BigEndian, &b.MaxX); err != nil {
		return -1, err
	}
	if err := binary.Read(reader, binary.BigEndian, &b.MinY); err != nil {
		return -1, err
	}
	if err = binary.Read(reader, binary.BigEndian, &b.MaxY); err != nil {
		return -1, err
	}
	if err := binary.Read(reader, binary.BigEndian, &b.OriginX); err != nil {
		return -1, err
	}
	if err = binary.Read(reader, binary.BigEndian, &b.OriginY); err != nil {
		return -1, err
	}
	return 52, nil
}

func (b *Bounds) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	if _, err := b.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *Bounds) UnmarshalBinary(data []byte) error {
	if len(data) < 52 {
		return fmt.Errorf("Not enough data [%d bytes] to read Bounds [52 bytes required]", len(data))
	}
	_, err := b.ReadFrom(bytes.NewReader(data[:58]))
	return err
}

func (b *Bounds) String() string {
	return fmt.Sprintf("Bounds{MinX: %f, MaxY: %f, MaxX: %f, MinY: %f}", b.MinX, b.MaxY, b.MaxX, b.MinY)
}

func (b *Bounds) Intersects(other *Bounds) bool {
	if other.MinX > b.MaxX || other.MinY > b.MaxY || other.MaxX < b.MinX || other.MaxY < b.MinY {
		return false
	}
	return true
}

func (b *Bounds) IntersectPercentage(other *Bounds) float64 {
	xOverlap := math.Max(0.0, math.Min(other.MaxX, b.MaxX)-math.Max(other.MinX, b.MinX))
	yOverlap := math.Max(0.0, math.Min(other.MaxY, b.MaxY)-math.Max(other.MinY, b.MinY))
	si := xOverlap * yOverlap
	s := b.Xspan() * b.Yspan()
	return math.Min(1.0, si/s)
}
