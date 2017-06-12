package proj4support

import (

	"fmt"
	"math"
	"strings"
	"github.com/geodatalake/geom"
	"github.com/geodatalake/geom/proj"
)

func ConvertPoints( epsg  string , points []geom.Point, pathToReadFrom string )  (outpoints []geom.Point) {

	const deg2Rad = math.Pi / 180.0
	const rad2Deg = 180.0 / math.Pi

	var srcProjection *proj.SR

	// Check if serialization versions of lookups exist
	if CheckAndLoadMaps( pathToReadFrom ) == false {

		fmt.Println("Failed to load binary map data")
		return nil

	} else {
		fmt.Println("Maps already loaded")
	}


	if  strings.Contains( strings.ToUpper(epsg), "EPSG") {

		// First lets get epsg string down to its base
		var epsgPurified = strings.Replace( epsg, "EPSG","",1)
		epsgPurified = strings.Replace( epsgPurified, ":","",1)
		epsgPurified = strings.TrimSpace( epsgPurified)

		tempProjection, err := GetDefByEPSG( "EPSG:"+  epsgPurified)

		if err != nil {
			fmt.Println("srcProjection is nil")
			return nil
		}

		srcProjection = tempProjection
	} else if strings.Contains( strings.ToUpper(epsg), "GCS") {
		epsgcode := DefsByEPSGGCS[ epsg ]

		tempProjection, err := GetDefByEPSG( epsgcode)

		if err != nil {
			fmt.Println("srcProjection is nil")
			return nil
		}

		srcProjection = tempProjection
	} else
	{

		fmt.Println("Assuming Title")
		var titleStr = strings.TrimSpace( epsg )
		titleStr = strings.Replace( titleStr, "/","",-1)
		titleStr = strings.Replace( titleStr, " ","",-1)
		titleStr = strings.ToUpper( titleStr )

		fmt.Println("Searching with: " + titleStr)
		tempProjection, err  := GetDefByTitle( titleStr )

		if err != nil {
			fmt.Println("srcProjection is nil")
			return nil
		}

		srcProjection = tempProjection
	}

	fmt.Println(srcProjection)
	tgtProjection, err := proj.Parse("+proj=longlat +datum=WGS84 +no_defs")

	if err != nil {
		fmt.Println( "Fatal error tgtProjection")
	}

	if tgtProjection == nil {
		fmt.Println("tgtProjection is nil")
	}

	fmt.Println( "Units " + tgtProjection.DatumCode )
	trans, err := srcProjection.NewTransform( tgtProjection )
	if err != nil {
		fmt.Println("Bad new Transform")
	}

	iCount := len(points)

	for i := 0; i < iCount; i++ {


		rsltx, rslty, err := trans(points[i].X, points[i].Y)

		if err != nil {
			fmt.Println( "Error on translation")
		} else {
			var sResult1 string = fmt.Sprintf("%.10f", rsltx )
			var sResult2 string = fmt.Sprintf("%.10f", rslty )

			fmt.Println(sResult1 + "," + sResult2)

			outpoints = append(outpoints, geom.Point{X: rsltx, Y: rslty})
		}

	}


	return outpoints

}