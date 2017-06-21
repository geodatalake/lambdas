package raster

import (
	"fmt"
	"io"
	"log"

	"github.com/geodatalake/lambdas/geotiff"
	"github.com/geodatalake/lambdas/lidar"
)

type RasterStream interface {
	io.Reader
	io.ReaderAt
}

type NotaRasterFile string

func (s NotaRasterFile) Error() string {
	return fmt.Sprintf("%v", s)
}

func IsRaster(src RasterStream) (interface{}, error) {
	if g, err := geotiff.NewDecoder(src); err == nil {
		return g, nil
	} else {
		if _, ok := err.(geotiff.NotaTiffFile); !ok {
			log.Println(err)
			return nil, err
		}
		if l, err := lidar.NewStreamReader(src, nil); err == nil {
			log.Println("Found lidar", l)
			return l, nil
		} else {
			if _, ok := err.(lidar.NotaLasFile); !ok {
				log.Println(err)
				return nil, err
			}
		}
	}
	return nil, NotaRasterFile("Not a known raster type")
}
