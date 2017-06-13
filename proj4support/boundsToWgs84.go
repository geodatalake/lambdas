package proj4support

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/geodatalake/geom"
	"github.com/geodatalake/geom/proj"
	"github.com/geodatalake/lambdas/geotiff"
)

func PointAsWkt(pt geom.Point) string {
	return fmt.Sprintf("%s %s", strconv.FormatFloat(pt.X, 'f', 6, 64), strconv.FormatFloat(pt.Y, 'f', 6, 64))
}

func MakePoints(data []string) []geom.Point {
	retval := make([]geom.Point, 0, len(data)/2)
	raw := make([]float64, 0, len(data))
	for _, s := range data {
		val, _ := strconv.ParseFloat(s, 64)
		raw = append(raw, val)
	}
	for i := 0; i < len(raw); i += 2 {
		retval = append(retval, geom.Point{X: raw[i], Y: raw[i+1]})
	}
	return retval
}

type ReProject struct {
	LoadPath string
}

func (rp *ReProject) Convert(epsg string, bounds *geotiff.Bounds) *geotiff.Bounds {
	return geotiff.NewBoundsFromGeomPoints(ConvertPoints(epsg, bounds.AsGeomPoints(), rp.LoadPath))
}

const (
	deg2Rad = math.Pi / 180.0
	rad2Deg = 180.0 / math.Pi
)

func ConvertPoints(epsg string, points []geom.Point, pathToReadFrom string) (outpoints []geom.Point) {

	var srcProjection *proj.SR

	// Check if serialization versions of lookups exist
	if CheckAndLoadMaps(pathToReadFrom) == false {

		fmt.Println("Failed to load binary map data")
		return nil

	} else {
		fmt.Println("Maps already loaded")
	}

	if strings.Contains(strings.ToUpper(epsg), "EPSG") {

		// First lets get epsg string down to its base
		var epsgPurified = strings.Replace(epsg, "EPSG", "", 1)
		epsgPurified = strings.Replace(epsgPurified, ":", "", 1)
		epsgPurified = strings.TrimSpace(epsgPurified)

		tempProjection, err := GetDefByEPSG("EPSG:" + epsgPurified)

		if err != nil {
			fmt.Println("srcProjection is nil")
			return nil
		}

		srcProjection = tempProjection
	} else {

		fmt.Println("Assuming Title")
		var titleStr = strings.TrimSpace(epsg)
		titleStr = strings.Replace(titleStr, "/", "", -1)
		titleStr = strings.Replace(titleStr, " ", "", -1)
		titleStr = strings.ToUpper(titleStr)

		fmt.Println("Searching with: " + titleStr)
		tempProjection, err := GetDefByTitle(titleStr)

		if err != nil {
			fmt.Println("srcProjection is nil")
			return nil
		}

		srcProjection = tempProjection
	}

	fmt.Println(srcProjection)
	tgtProjection, err := proj.Parse("+proj=longlat +datum=WGS84 +no_defs")

	if err != nil {
		fmt.Println("Fatal error tgtProjection")
	}

	if tgtProjection == nil {
		fmt.Println("tgtProjection is nil")
	}

	fmt.Println("Units " + tgtProjection.DatumCode)
	trans, err := srcProjection.NewTransform(tgtProjection)
	if err != nil {
		fmt.Println("Bad new Transform")
	}

	iCount := len(points)

	for i := 0; i < iCount; i++ {

		rsltx, rslty, err := trans(points[i].X, points[i].Y)

		if err != nil {
			fmt.Println("Error on translation")
		} else {
			var sResult1 string = fmt.Sprintf("%.10f", rsltx)
			var sResult2 string = fmt.Sprintf("%.10f", rslty)

			fmt.Println(sResult1 + "," + sResult2)

			outpoints = append(outpoints, geom.Point{X: rsltx, Y: rslty})
		}

	}

	return outpoints

}
