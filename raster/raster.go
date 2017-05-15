package raster

import (
	"errors"
	"io"

	"github.com/geodatalake/lambdas/geotiff"
	"github.com/geodatalake/lambdas/lidar"
)

type RasterStream interface {
	io.Reader
	io.ReaderAt
}

func IsRaster(src RasterStream) (interface{}, error) {
	if g, err := geotiff.NewDecoder(src); err == nil {
		return g, nil
	} else {
		if l, err := lidar.NewStreamReader(src, nil); err == nil {
			return l, nil
		}
	}
	return nil, errors.New("Not a known raster type")
}
