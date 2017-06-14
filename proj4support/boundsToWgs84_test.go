package proj4support

import (
	"fmt"
	"testing"

	"github.com/geodatalake/geom"
)

func check(vpt, opt geom.Point, t *testing.T) {
	vptX := fmt.Sprintf("%.6f", vpt.X)
	optX := fmt.Sprintf("%.6f", opt.X)
	vptY := fmt.Sprintf("%.6f", vpt.Y)
	optY := fmt.Sprintf("%.6f", opt.Y)

	if vptX == optX && vptY == optY {
		// passed
	} else {
		t.Errorf(fmt.Sprintf("Expected %f,%f received %f,%f", vpt.X, vpt.Y, opt.X, opt.Y))
	}
}

func TestConversionEPSG(t *testing.T) {
	s := fmt.Sprintf("EPSG:%d", 32618)

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{X: -76.0223341130, Y: 36.9337347972})
	validPoints = append(validPoints, geom.Point{X: -76.0216111717, Y: 36.8795388230})
	validPoints = append(validPoints, geom.Point{X: -75.9494219647, Y: 36.9343400453})
	validPoints = append(validPoints, geom.Point{X: -75.9487505619, Y: 36.8801428874})

	var testPoints []geom.Point

	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4081996.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4081996.500000})

	outpoints, _ := ConvertPoints(s, testPoints)

	iCount := len(outpoints)

	for i := 0; i < iCount; i++ {

		check(validPoints[i], outpoints[i], t)
	}
}

func TestConversionEPSGNoColon(t *testing.T) {
	s := fmt.Sprintf("EPSG%d", 32618)

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{X: -76.0223341130, Y: 36.9337347972})
	validPoints = append(validPoints, geom.Point{X: -76.0216111717, Y: 36.8795388230})
	validPoints = append(validPoints, geom.Point{X: -75.9494219647, Y: 36.9343400453})
	validPoints = append(validPoints, geom.Point{X: -75.9487505619, Y: 36.8801428874})

	var testPoints []geom.Point

	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4081996.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4081996.500000})

	outpoints, _ := ConvertPoints(s, testPoints)

	iCount := len(outpoints)

	for i := 0; i < iCount; i++ {
		check(validPoints[i], outpoints[i], t)
	}
}

func TestConversionEPSGBackToBack(t *testing.T) {
	var s = fmt.Sprintf("EPSG %d", 32618)

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{X: -76.0223341130, Y: 36.9337347972})
	validPoints = append(validPoints, geom.Point{X: -76.0216111700, Y: 36.8795388230})
	validPoints = append(validPoints, geom.Point{X: -75.9494219647, Y: 36.9343400453})
	validPoints = append(validPoints, geom.Point{X: -75.9487505635, Y: 36.8801428874})

	var testPoints []geom.Point
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4081996.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4081996.500000})

	var outpoints, _ = ConvertPoints(s, testPoints)
	var iCount = len(outpoints)
	for i := 0; i < iCount; i++ {
		check(validPoints[i], outpoints[i], t)
	}

	s = fmt.Sprintf("EPSG %d", 32618)

	var validPoints1 []geom.Point
	validPoints1 = append(validPoints, geom.Point{X: -76.0223341130, Y: 36.9337347972})
	validPoints1 = append(validPoints, geom.Point{X: -76.0216111700, Y: 36.8795388230})
	validPoints1 = append(validPoints, geom.Point{X: -75.9494219647, Y: 36.9343400453})
	validPoints1 = append(validPoints, geom.Point{X: -75.9487505635, Y: 36.8801428874})

	var testPoints2 []geom.Point
	testPoints2 = append(testPoints, geom.Point{X: 408956.500000, Y: 4088009.500000})
	testPoints2 = append(testPoints, geom.Point{X: 408956.500000, Y: 4081996.500000})
	testPoints2 = append(testPoints, geom.Point{X: 415450.500000, Y: 4088009.500000})
	testPoints2 = append(testPoints, geom.Point{X: 415450.500000, Y: 4081996.500000})

	outpoints, _ = ConvertPoints(s, testPoints2)
	iCount = len(outpoints)
	for i := 0; i < iCount; i++ {
		check(validPoints1[i], outpoints[i], t)
	}
}

func TestConversionTitle(t *testing.T) {
	s := fmt.Sprintf("NAD27(76) / UTM zone 16N")

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{X: -88.0223336890, Y: 36.9355647985})
	validPoints = append(validPoints, geom.Point{X: -88.0216107159, Y: 36.8813675603})
	validPoints = append(validPoints, geom.Point{X: -87.9494215714, Y: 36.9361700862})
	validPoints = append(validPoints, geom.Point{X: -87.9487501406, Y: 36.8819716642})

	var testPoints []geom.Point

	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4081996.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4081996.500000})

	outpoints, _ := ConvertPoints(s, testPoints)

	iCount := len(outpoints)

	for i := 0; i < iCount; i++ {
		check(validPoints[i], outpoints[i], t)
	}
}

func TestConversionTitleBogusTitle(t *testing.T) {
	s := fmt.Sprintf("NAD64(76) UTM zone 16N")

	var testPoints []geom.Point

	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 408956.500000, Y: 4081996.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4088009.500000})
	testPoints = append(testPoints, geom.Point{X: 415450.500000, Y: 4081996.500000})

	_, result := ConvertPoints(s, testPoints)

	if result == false {
		// passed
	} else {
		t.Errorf(fmt.Sprintf("Expected no output points"))
	}
}

func TestGCSDatumCode(t *testing.T) {
	s := fmt.Sprintf("GCS_WGS_84")

	var testPoints []geom.Point
	testPoints = append(testPoints, geom.Point{X: -88.0223336890, Y: 36.9355647985})
	testPoints = append(testPoints, geom.Point{X: -88.0216107159, Y: 36.8813675603})
	testPoints = append(testPoints, geom.Point{X: -87.9494215714, Y: 36.9361700862})
	testPoints = append(testPoints, geom.Point{X: -87.9487501406, Y: 36.8819716642})

	var validPoints []geom.Point
	validPoints = append(validPoints, geom.Point{X: -88.0223336890, Y: 36.9355647985})
	validPoints = append(validPoints, geom.Point{X: -88.0216107159, Y: 36.8813675603})
	validPoints = append(validPoints, geom.Point{X: -87.9494215714, Y: 36.9361700862})
	validPoints = append(validPoints, geom.Point{X: -87.9487501406, Y: 36.8819716642})

	outpoints, _ := ConvertPoints(s, testPoints)

	iCount := len(outpoints)

	if outpoints == nil {
		t.Errorf("No Points Generated")
	} else {
		for i := 0; i < iCount; i++ {
			check(validPoints[i], outpoints[i], t)
		}
	}
}
