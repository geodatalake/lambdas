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

type HandleReturn struct {
	Bounds       string
	Prj          string
	Typ          string
	LastModified string
}

func getProjection(rtype geotiff.Tiff) string {
	var value interface{}
	var err error
	if value, err = rtype.GetGeoKeyValue(geotiff.GeoKeyProjectedCSTypeGeoKey); err == nil {
		return value.(string)
	}
	if value, err = rtype.GetGeoKeyValue(geotiff.GeoKeyGeographicType); err == nil {
		return value.(string)
	}
	return "GCS_WGS_84"
}

func handleRaster(r interface{}, proj Projector) (*HandleReturn, error) {
	switch rtype := r.(type) {
	case geotiff.Tiff:
		if rtype.IsGeotiff() {
			value := getProjection(rtype)
			if bounds, err1 := rtype.Bounds(); err1 != nil {
				return nil, err1
			} else {
				lastModified := ""
				tm, err := rtype.DateTime()
				if err == nil {
					lastModified = tm.UTC().Format(bucket.ISO8601FORMAT)
				}
				if proj != nil {
					if value != "GCS_WGS_84" && value != "EPSG 4326" && value != "" {
						bounds = proj.Convert(value, bounds)
					}
					return &HandleReturn{Bounds: bounds.AsWkt(), Prj: "EPSG 4326", Typ: "geotiff", LastModified: lastModified}, nil
				} else {
					return &HandleReturn{Bounds: bounds.AsWkt(), Prj: value, Typ: "geotiff", LastModified: lastModified}, nil
				}
			}
		}
	case lidar.Las:
		bounds := rtype.Bounds()
		log.Println("Found bounds", bounds)
		tm, err := rtype.DateTime()
		lastModified := ""
		if err == nil {
			lastModified = tm.UTC().Format(bucket.ISO8601FORMAT)
		}
		if rtype.IsWktCrs() {
			log.Println("WKT CRS")
			return &HandleReturn{Bounds: bounds.AsWkt(), Prj: rtype.WktCrs().Wkt, Typ: "lidar", LastModified: lastModified}, nil
		} else {
			log.Println("GeoKeys CRS")
			prj, ok := rtype.GeotiffCrs().GetProjectedCSType()
			if !ok {
				return &HandleReturn{Bounds: bounds.AsWkt(), Prj: "EPSG 4326", Typ: "lidar", LastModified: lastModified}, nil
			} else {
				if proj != nil {
					if prj != "GCS_WGS_84" && prj != "EPSG 4326" && prj != "" {
						bounds = proj.Convert(prj, bounds)
					}
					return &HandleReturn{Bounds: bounds.AsWkt(), Prj: "EPSG 4326", Typ: "lidar", LastModified: lastModified}, nil
				} else {
					return &HandleReturn{Bounds: bounds.AsWkt(), Prj: prj, Typ: "lidar", LastModified: lastModified}, nil
				}
			}
		}
	}
	log.Println("Unknown raster interface")
	return nil, fmt.Errorf("Unknown raster type %v", r)
}

func handleVector(v vector.VectorIntfc, proj Projector) (*HandleReturn, error) {

	// bounds.AsWkt()
	// The 2nd string is Projection
	// The 3rd is the geo type (kml, shapefile, etc)
	vBounds, err := v.Bounds()

	if err != nil {
		return nil, fmt.Errorf("TODO: Implement vector methods")
	}

	var geotype = "kml"
	if v.IsShape() {
		geotype = "shapefile"
	}

	projection := "EPSG 4326"
	boundswkt := vBounds.AsWkt()

	return &HandleReturn{Bounds: boundswkt, Prj: projection, Typ: geotype, LastModified: ""}, nil
}

type Projector interface {
	Convert(epsg string, bounds *geotiff.Bounds) *geotiff.Bounds
}

func DetectType(stream raster.RasterStream, proj Projector) (*HandleReturn, error) {
	if r, err := raster.IsRaster(stream); err != nil {
		if v, err1 := vector.IsVector(stream); err1 != nil {
			return nil, fmt.Errorf("Unknown file type %v", err1)
		} else {
			return handleVector(v, proj)
		}
	} else {
		return handleRaster(r, proj)
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
		return false
	}
}

type Extractable interface {
	GetKeys() []*bucket.BucketFile
}

func (ef *ExtractFile) IsShapeFile() bool {
	if ef.File.IsShapeFileRoot() {
		return true
	}
	for _, f := range ef.Aux {
		if f.IsShapeFileRoot() {
			return true
		}
	}
	return false
}

func (ef *ExtractFile) EjectNonShapeFile() ([]*ExtractFile, bool) {
	if ef.IsShapeFile() {
		// So we have a .shp somewhere in File or Aux
		// Eject non shapefile extensions, put .shp in File
		retval := make([]*ExtractFile, 0, 32)
		if !ef.File.IsShapeFileRoot() {
			// root files are either .shp or
			// other extension not part of shape files
			// If not .shp, eject it and replace with .shp
			eject := new(ExtractFile)
			eject.File = ef.File
			retval = append(retval, eject)
			for _, f := range ef.Aux {
				if f.IsShapeFileRoot() {
					ef.File = f
				}
			}
		}
		// Now eject any non shape file extensions
		// This will include remocing the shape file root (.shp) (if present)
		updatedAux := make([]*BucketFileInfo, 0, 32)
		for _, f := range ef.Aux {
			if !f.IsShapeFileAux() {
				if !f.IsShapeFileRoot() {
					eject := new(ExtractFile)
					eject.File = f
					retval = append(retval, eject)
				}
			} else {
				updatedAux = append(updatedAux, f)
			}
		}
		ef.Aux = make([]*BucketFileInfo, len(updatedAux))
		copy(ef.Aux, updatedAux)
		return retval, len(retval) > 0
	}
	return nil, false
}

func Extract(di Extractable) ([]*ExtractFile, bool) {
	dirFiles := make(map[string]*bucket.BucketFile)
	baseNames := NewStringBucketSet()
	keys := di.GetKeys()
	retval := make([]*ExtractFile, 0, len(keys))
	// First split out files with extensions, add others
	for _, f := range keys {
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
		if ef.File == nil {
			// No root file found, just use the first Aux file
			if len(ef.Aux) > 1 {
				ef.File = ef.Aux[0]
				ef.Aux = ef.Aux[1:]
			} else if len(ef.Aux) == 1 {
				ef.File = ef.Aux[0]
				ef.Aux = nil
			}
		}
		// Ejects files that are not part of a shapefile array
		// If the ef does not contain a .shp, then nothing is done
		if files, ok := ef.EjectNonShapeFile(); ok {
			// Add the ejected files as individuals
			retval = append(retval, files...)
		}
		// Now add the ef (potentially altered to have the .shp in root File)
		retval = append(retval, ef)
	}

	return retval, true
}
