// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geoindex

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/geotiff"
	"github.com/geodatalake/lambdas/lidar"
	"github.com/geodatalake/lambdas/raster"
	"github.com/geodatalake/lambdas/vector"
)

func handleRaster(r interface{}) (string, string, error) {
	switch rtype := r.(type) {
	case geotiff.Tiff:
		if rtype.IsGeotiff() {
			if value, err := rtype.GetGeoKeyValue(geotiff.GeoKeyProjectedCSTypeGeoKey); err != nil {
				return "", "", err
			} else if bounds, err1 := rtype.Bounds(); err1 != nil {
				return "", "", err1
			} else {
				return bounds.AsWkt(), value.(string), nil
			}
		}
	case lidar.Las:
		bounds := rtype.Bounds()
		log.Println("Found bounds", bounds)
		if rtype.IsWktCrs() {
			log.Println("WKT CRS")
			return bounds.AsWkt(), rtype.WktCrs().Wkt, nil
		} else {
			log.Println("GeoKeys CRS")
			prj, ok := rtype.GeotiffCrs().GetProjectedCSType()
			if !ok {
				return bounds.AsWkt(), "", nil
			} else {
				return bounds.AsWkt(), prj, nil
			}
		}
	}
	log.Println("Unknown raster interface")
	return "", "", fmt.Errorf("Unknown raster type %v", r)
}

func handleVector(v interface{}) (string, string, error) {
	// TODO: Add vector types, bounds are returned as WKT
	// bounds.AsWkt()
	// The 2nd string is Projection
	return "", "", fmt.Errorf("TODO: Implement vector methods")
}

func DetectType(stream raster.RasterStream) (string, string, error) {
	if r, err := raster.IsRaster(stream); err != nil {
		if v, err1 := vector.IsVector(stream); err1 != nil {
			return "", "", fmt.Errorf("Unknown file type %v", err1)
		} else {
			return handleVector(v)
		}
	} else {
		return handleRaster(r)
	}
}

func getExtension(name string) (string, string, bool) {
	idex := strings.LastIndex(name, ".")
	if idex != -1 {
		base, ext := name[0:idex], name[idex+1:]
		return base, ext, true
	}
	return "", "", false
}

type BucketData struct {
	Base, Ext string
	Bucket    *bucket.BucketFile
}

func NewBucketData(base, ext string, b *bucket.BucketFile) *BucketData {
	return &BucketData{Base: base, Ext: ext, Bucket: b}
}

type StringBucketSet struct {
	set map[string][]*BucketData
}

func NewStringBucketSet() *StringBucketSet {
	return &StringBucketSet{set: make(map[string][]*BucketData)}
}

func (b *StringBucketSet) Add(s string, newB *BucketData) {
	b.set[s] = append(b.set[s], newB)
}

func isRootExt(ext string) bool {
	switch ext {
	case "shp":
		return true
	case "tif":
		return true
	case "nitf":
		return true
	case "nif":
		return true
	case "asc":
		return false
	case "dbf":
		return false
	case "prj":
		return false
	case "xml":
		return false
	default:
		log.Println("Unknown extension", ext)
		return false
	}
}

func Extract(di *bucket.DirItem) ([]*ExtractFile, bool) {
	dirFiles := make(map[string]*bucket.BucketFile)
	baseNames := NewStringBucketSet()
	retval := make([]*ExtractFile, 0, len(di.Keys))
	// First split out files with extensions, add others
	for _, f := range di.Keys {
		_, name := path.Split(f.Key)
		dirFiles[name] = f
		if base, ext, ok := getExtension(name); ok {
			baseNames.Add(base, NewBucketData(base, ext, f))
		} else {
			ef := new(ExtractFile)
			ef.File = NewBucketFileInfo(f)
			retval = append(retval, ef)
		}
	}
	// Now go thru the files with extensions and create ExtractFiles for them
	for _, bSet := range baseNames.set {
		ef := new(ExtractFile)
		for _, v := range bSet {
			if isRootExt(v.Ext) {
				ef.File = NewBucketFileInfo(v.Bucket)
			} else {
				ef.Aux = append(ef.Aux, NewBucketFileInfo(v.Bucket))
			}
		}
		retval = append(retval, ef)
	}

	return retval, true
}
