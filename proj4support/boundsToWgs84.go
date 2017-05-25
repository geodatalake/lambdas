package proj4support

import (
	"fmt"
	"math"
	"os"
	"log"
	"bufio"
	"strings"
	"github.com/ctessum/geom"
	"github.com/ctessum/geom/proj"
	"runtime"
)


func ConvertPoints( epsg  string , points []geom.Point )  (outpoints []geom.Point) {

	const deg2Rad = math.Pi / 180.0
	const rad2Deg = 180.0 / math.Pi

	_, fileName, _, _ := runtime.Caller(0)

	eosIndex :=  strings.LastIndex( fileName, "/")
	rootStr := fileName[:eosIndex]
	fullPath := rootStr + "/config/epsg"
	//fmt.Println( fullPath )

	file, err := os.Open( fullPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//startEPSG := true
	var epsgTitle = ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		// Look for EPSG Name First
		newLine := scanner.Text()
		if strings.IndexAny( newLine, "#" ) == 0 {

			if len( strings.Trim( newLine," ") )> 1 {
				epsgTitle = newLine[2:len(newLine)]
				epsgTitle = strings.Replace(epsgTitle, "/", "", 1)
				epsgTitle = "+title=" + epsgTitle

				if len(epsgTitle) == len("+title=") {
					epsgTitle = ""
				}
			} else {
				epsgTitle = ""
			}
		} else if  strings.IndexAny( newLine, "<" ) == 0 &&  len(epsgTitle) > 0 {

			// Handle getting the EPSG code and Proj4 String
			epsgIndex := strings.IndexAny( newLine, ">")
			var epsgCode = newLine[1:epsgIndex]
			epsgCode = "EPSG:" + epsgCode

			// Now get proj string
			var projString = newLine[epsgIndex+1:len(newLine)]
			projString = strings.Replace(projString, "<>", "",1)
			projString = strings.Trim( projString, " ")

			totalString := epsgTitle + " " + projString

			AddDef( epsgCode, totalString )
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}


	srcProjection := GetDef("EPSG:"+  epsg)

	if srcProjection == nil {
		fmt.Println("srcProjection is nil")
	}

	tgtProjection, err := proj.Parse("+proj=longlat +datum=WGS84 +no_defs")

	if err != nil {
		fmt.Println( "Fatal error tgtProjection")
	}
	if tgtProjection == nil {
		fmt.Println("tgtProjection is nil")
	}

	trans, err := srcProjection.NewTransform(tgtProjection)
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