package proj4support

import (

	"fmt"
	"math"
	"os"
	"log"
	"bufio"
	"strings"
	"github.com/geodatalake/geom"
	"github.com/geodatalake/geom/proj"
	"runtime"
)

func ConvertPoints( epsg  string , points []geom.Point )  (outpoints []geom.Point) {

	const deg2Rad = math.Pi / 180.0
	const rad2Deg = 180.0 / math.Pi

	var srcProjection *proj.SR

	// Check if serialization versions of lookups exist
	if CheckAndLoadMaps() == false {

		fmt.Println("Initializing tables")
		_, fileName, _, _ := runtime.Caller(0)
		eosIndex :=  strings.LastIndex( fileName, "/")
		rootStr := fileName[:eosIndex]
		fullPath := rootStr + "/config/epsg"
		fmt.Println( fullPath )

		file, err := os.Open( fullPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		var epsgTitle = ""
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {

			// Look for EPSG Name First
			newLine := scanner.Text()
			if strings.IndexAny( newLine, "#" ) == 0 {

				if len( strings.Trim( newLine," ") )> 1 {
					epsgTitle = newLine[2:len(newLine)]
					epsgTitle = strings.Replace(epsgTitle, "/", "", -1)
					epsgTitle = strings.TrimSpace( epsgTitle )
					epsgTitle = strings.ToUpper(  epsgTitle )
				} else {
					epsgTitle = ""
				}
			} else if  strings.IndexAny( newLine, "<" ) == 0 && len( epsgTitle ) > 0 {

				// Handle getting the EPSG code and Proj4 String
				epsgIndex := strings.IndexAny( newLine, ">")
				var epsgCode = newLine[1:epsgIndex]
				epsgCode = "EPSG:" + epsgCode

				// Now get proj string
				var projString = newLine[epsgIndex+1:len(newLine)]
				projString = strings.Replace(projString, "<>", "",1)
				projString = strings.TrimSpace( projString)

				totalString := "+title=" + epsgTitle + " " + projString

				AddDef( epsgCode, totalString )
				AddTitleDef( epsgTitle, totalString)
			}
		}

		SerializeMaps()

	} else {
		fmt.Println("Maps already loaded")
	}

	//if err := scanner.Err(); err != nil {
	//	log.Fatal(err)
	//}

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
	} else {

		var titleStr = strings.TrimSpace( epsg )
		titleStr = strings.Replace( titleStr, "/","",-1)
		titleStr = strings.ToUpper( titleStr )
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
			fmt.Println( "Error on translateion")
		} else {
			var sResult1 string = fmt.Sprintf("%.10f", rsltx )
			var sResult2 string = fmt.Sprintf("%.10f", rslty )

			fmt.Println(sResult1 + "," + sResult2)

			outpoints = append(outpoints, geom.Point{X: rsltx, Y: rslty})
		}

	}


	return outpoints

}