package proj4support


import (
	"fmt"
	"testing"
	"github.com/geodatalake/geom"
	"runtime"
	"strings"
)

func check( vpt, opt geom.Point, t *testing.T) {

	vptX := fmt.Sprintf("%.6f", vpt.X)
	optX := fmt.Sprintf("%.6f", opt.X)
	vptY := fmt.Sprintf("%.6f", vpt.Y)
	optY := fmt.Sprintf("%.6f", opt.Y)

	if vptX == optX && vptY == optY {
		fmt.Println( "Passed")
	} else {
		t.Errorf(fmt.Sprintf("Expected %f,%f received %f,%f", vpt.X, vpt.Y, opt.X, opt.Y))
	}
}

func TestBuildMaps (t *testing.T) {

	fmt.Println("Initializing tables")
	_, fileName, _, _ := runtime.Caller(0)
	eosIndex := strings.LastIndex(fileName, "/")
	rootStr := fileName[:eosIndex]
	fullPath := rootStr + "/config/epsg"
	fmt.Println(fullPath)

	BuildMaps( fullPath, "" )

}

func TestConversionEPSG (t *testing.T) {


	fmt.Println("Calculating Bounding box")
	s := fmt.Sprintf("EPSG:%d", 32618)
	fmt.Println("Code is " + s)

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{ X: -76.0223341130, Y: 36.9337347972})
	validPoints = append(validPoints, geom.Point{ X: -76.0216111717, Y: 36.8795388230})
	validPoints = append(validPoints, geom.Point{ X: -75.9494219647, Y: 36.9343400453})
	validPoints = append(validPoints, geom.Point{ X: -75.9487505619, Y: 36.8801428874})

	var testPoints []geom.Point

	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4081996.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4081996.500000 })

	outpoints,_ := ConvertPoints( s, testPoints, "")

	iCount := len(outpoints)

	fmt.Println( fmt.Sprintf( "Count is %d", iCount) )

	for i := 0; i < iCount; i++ {


		var sResult1 string = fmt.Sprintf("%.10f", outpoints[i].X )
		var sResult2 string = fmt.Sprintf("%.10f", outpoints[i].Y )

		fmt.Println ( sResult1  + "," +  sResult2 )

		check( validPoints[i], outpoints[i], t)


	}


}

func TestConversionEPSGNoColon (t *testing.T) {


	fmt.Println("Calculating Bounding box")
	s := fmt.Sprintf("EPSG%d", 32618)
	fmt.Println("Code is " + s)

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{ X: -76.0223341130, Y: 36.9337347972})
	validPoints = append(validPoints, geom.Point{ X: -76.0216111717, Y: 36.8795388230})
	validPoints = append(validPoints, geom.Point{ X: -75.9494219647, Y: 36.9343400453})
	validPoints = append(validPoints, geom.Point{ X: -75.9487505619, Y: 36.8801428874})

	var testPoints []geom.Point

	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4081996.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4081996.500000 })

	outpoints,_ := ConvertPoints( s, testPoints, "")

	iCount := len(outpoints)

	fmt.Println( fmt.Sprintf( "Count is %d", iCount) )

	for i := 0; i < iCount; i++ {


		var sResult1 string = fmt.Sprintf("%.10f", outpoints[i].X )
		var sResult2 string = fmt.Sprintf("%.10f", outpoints[i].Y )

		fmt.Println ( sResult1  + "," +  sResult2 )

		check( validPoints[i], outpoints[i], t)


	}


}

func TestConversionEPSGBackToBack (t *testing.T) {


	fmt.Println("Calculating Bounding box")
	var s = fmt.Sprintf("EPSG %d", 32618)
	fmt.Println("Code is " + s)

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{ X: -76.0223341130, Y: 36.9337347972})
	validPoints = append(validPoints, geom.Point{ X: -76.0216111700, Y: 36.8795388230})
	validPoints = append(validPoints, geom.Point{ X: -75.9494219647, Y: 36.9343400453})
	validPoints = append(validPoints, geom.Point{ X: -75.9487505635, Y: 36.8801428874})

	var testPoints []geom.Point
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4081996.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4081996.500000 })

	var outpoints,_ = ConvertPoints( s, testPoints, "")
	var iCount = len(outpoints)
	for i := 0; i < iCount; i++ {


		var sResult1 string = fmt.Sprintf("%.10f", outpoints[i].X )
		var sResult2 string = fmt.Sprintf("%.10f", outpoints[i].Y )

		fmt.Println ( sResult1  + "," +  sResult2 )

		check( validPoints[i], outpoints[i], t)


	}

	fmt.Println("Calculating Bounding box")
	s = fmt.Sprintf("EPSG %d", 32618)
	fmt.Println("Code is " + s)

	var validPoints1 []geom.Point
	validPoints1 = append(validPoints, geom.Point{ X: -76.0223341130, Y: 36.9337347972})
	validPoints1 = append(validPoints, geom.Point{ X: -76.0216111700, Y: 36.8795388230})
	validPoints1 = append(validPoints, geom.Point{ X: -75.9494219647, Y: 36.9343400453})
	validPoints1 = append(validPoints, geom.Point{ X: -75.9487505635, Y: 36.8801428874})

	var testPoints2 []geom.Point
	testPoints2 = append(testPoints, geom.Point{X: 408956.500000, Y:4088009.500000 })
	testPoints2 = append(testPoints, geom.Point{X: 408956.500000, Y:4081996.500000 })
	testPoints2 = append(testPoints, geom.Point{X: 415450.500000, Y:4088009.500000 })
	testPoints2 = append(testPoints, geom.Point{X: 415450.500000, Y:4081996.500000 })

	outpoints, _ = ConvertPoints( s, testPoints2, "")
	iCount = len(outpoints)
	for i := 0; i < iCount; i++ {


		var sResult1 string = fmt.Sprintf("%.10f", outpoints[i].X )
		var sResult2 string = fmt.Sprintf("%.10f", outpoints[i].Y )

		fmt.Println ( sResult1  + "," +  sResult2 )

		check( validPoints1[i], outpoints[i], t)


	}


}

func TestConversionTitle (t *testing.T) {


	fmt.Println("Calculating Bounding box for Title String")
	s := fmt.Sprintf("NAD27(76) / UTM zone 16N")
	fmt.Println("Code is " + s)

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{ X: -88.0223336890, Y: 36.9355647985})
	validPoints = append(validPoints, geom.Point{ X: -88.0216107159, Y: 36.8813675603})
	validPoints = append(validPoints, geom.Point{ X: -87.9494215714, Y: 36.9361700862})
	validPoints = append(validPoints, geom.Point{ X: -87.9487501406, Y: 36.8819716642})

	var testPoints []geom.Point

	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4081996.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4081996.500000 })

	outpoints, _:= ConvertPoints( s, testPoints, "")

	iCount := len(outpoints)

	for i := 0; i < iCount; i++ {


		var sResult1 string = fmt.Sprintf("%.10f", outpoints[i].X )
		var sResult2 string = fmt.Sprintf("%.10f", outpoints[i].Y )

		fmt.Println ( sResult1  + "," +  sResult2 )

		check( validPoints[i], outpoints[i], t)


	}


}

func TestConversionTitleBogusTitle (t *testing.T) {


	fmt.Println("Calculating Bounding box for Title String")
	s := fmt.Sprintf("NAD64(76) UTM zone 16N")
	fmt.Println("Code is " + s)

	var testPoints []geom.Point

	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4081996.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4081996.500000 })

	_, result := ConvertPoints( s, testPoints, "")

	if result == false {
		fmt.Println( "Passed")
	} else {
		t.Errorf(fmt.Sprintf("Expected no output points"))
	}



}

func TestGCSDatumCode(t *testing.T) {


	fmt.Println("Calculating Bounding box for GCS Code String")
	s := fmt.Sprintf("GCS_WGS_84")
	fmt.Println("Code is " + s)

	var testPoints []geom.Point
	testPoints = append(testPoints, geom.Point{ X: -88.0223336890, Y: 36.9355647985})
	testPoints = append(testPoints, geom.Point{ X: -88.0216107159, Y: 36.8813675603})
	testPoints = append(testPoints, geom.Point{ X: -87.9494215714, Y: 36.9361700862})
	testPoints = append(testPoints, geom.Point{ X: -87.9487501406, Y: 36.8819716642})



	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{ X: -88.0223336890, Y: 36.9355647985})
	validPoints = append(validPoints, geom.Point{ X: -88.0216107159, Y: 36.8813675603})
	validPoints = append(validPoints, geom.Point{ X: -87.9494215714, Y: 36.9361700862})
	validPoints = append(validPoints, geom.Point{ X: -87.9487501406, Y: 36.8819716642})

	outpoints, _:= ConvertPoints( s, testPoints, "")

	iCount := len(outpoints)

	if outpoints == nil {
		fmt.Println( "No Points Generated")
	} else {

		for i := 0; i < iCount; i++ {


			var sResult1 string = fmt.Sprintf("%.10f", outpoints[i].X )
			var sResult2 string = fmt.Sprintf("%.10f", outpoints[i].Y )

			fmt.Println ( sResult1  + "," +  sResult2 )

			check( validPoints[i], outpoints[i], t)


		}
	}



}

