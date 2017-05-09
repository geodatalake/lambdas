package proj4support

import (
	"fmt"
	"math"
	"os"
	"github.com/xeonx/proj4"
	"github.com/xeonx/geom"
)


func ConvertPoints( epsg  string , points []geom.Point )  (outpoints []geom.Point) {

	const deg2Rad = math.Pi / 180.0
	const rad2Deg = 180.0 / math.Pi

	// Point the project to the local proj4 configuration files.
	os.Setenv("PROJ_LIB", "config")

	srcProjection, err := proj.InitPlus( "+init=epsg:"+ epsg )

	if err != nil {
		fmt.Println( "Fatal error srcProjection")
	}
	if srcProjection == nil {
		fmt.Println("srcProjection is nil")
	}

	tgtProjection, err := proj.InitPlus("+proj=longlat +datum=WGS84 +no_defs")
	if err != nil {
		fmt.Println( "Fatal error tgtProjection")
	}
	if tgtProjection == nil {
		fmt.Println("tgtProjection is nil")
	}


	iCount := len(points)


	if err := proj.TransformPoints(srcProjection, tgtProjection, points); err != nil {
		fmt.Println( "Conversion Error ")
	}


	result := srcProjection.IsGeoCent()

	if result == true {
		fmt.Println( "Is Geo Centric")
	}

	result2 := tgtProjection.IsLatLong()

	if result2 == true {
		fmt.Println( "Is Lat/Long")
	}

	//var outpoints []geom.Point

	for i := 0; i < iCount; i++ {


		var sResult1 string = fmt.Sprintf("%.10f", points[i].X * rad2Deg)
		var sResult2 string = fmt.Sprintf("%.10f", points[i].Y * rad2Deg)

		fmt.Println ( sResult1  + "," +  sResult2 )

		outpoints = append(outpoints, geom.Point{X: points[i].X * rad2Deg, Y: points[i].Y * rad2Deg })


	}


	return outpoints

}