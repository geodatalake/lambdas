package proj4support


import (
	"fmt"
	"testing"
	"github.com/ctessum/geom"
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

func TestConversion (t *testing.T) {


	fmt.Println("Calculating Bounding box")
	s := fmt.Sprintf("%d", 32618)
	fmt.Println("Code is " + s)

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{ X: -76.0223341130, Y: 36.9337347972})
	validPoints = append(validPoints, geom.Point{ X: -76.0216111700, Y: 36.8795388230})
	validPoints = append(validPoints, geom.Point{ X: -75.9494219647, Y: 36.9343400453})
	validPoints = append(validPoints, geom.Point{ X: -75.9487505619, Y: 36.8801428874})

	var testPoints []geom.Point

	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y:4081996.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4088009.500000 })
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y:4081996.500000 })

	outpoints := ConvertPoints( s, testPoints)

	iCount := len(outpoints)

	fmt.Println(" Length of return list %d ", iCount)
	for i := 0; i < iCount; i++ {


		var sResult1 string = fmt.Sprintf("%.10f", outpoints[i].X )
		var sResult2 string = fmt.Sprintf("%.10f", outpoints[i].Y )

		fmt.Println ( sResult1  + "," +  sResult2 )

		check( validPoints[i], outpoints[i], t)


	}


}
