package vector

import (
	"fmt"
	"testing"
	"os"
)


func TestInvalidShapeFile( t *testing.T ) {

	hndle, err := os.Open( "testfiles/Bad.shp" )

	if err == nil {

		result, _ := IsVector(hndle)

		if result.isVector() != false {
			t.Errorf ("File should have been invalid \n")
		} else {
			fmt.Println("ShapeFile Invalid - Good Test ")
		}

	} else {
		t.Errorf ("File should exist.\n")
	}

}

func TestNonExistentFile(t *testing.T) {

	// Test reading Non existent file
	_, err := os.Open( "testfiles/NumChuck.shp" )

	if  err == nil  {
		t.Errorf ("File should not exist \n")
	} else {
		fmt.Println("ShapeFile Non Existant - Good Test")
	}

}

func TestValidShapeFile( t *testing.T ) {

	// Test reading Valid Shape File
	hndle, err := os.Open( "testfiles/Good.shp" )

	if hndle != nil {

		vsInterface , _ := IsVectorType( hndle )

		if vsInterface != nil {

			if  vsInterface.isVector() != true  {
				t.Errorf ("Error reading header %v\n", err)
			} else {
				fmt.Println("ShapeFile Validated")
			}
		} else {
			t.Errorf ("Return interface is nil \n")
		}

	}

}


func TestValidKMLFile( t *testing.T ) {

	// Test reading Valid Shape File
	hndle, err := os.Open( "testfiles/NC_Counties.kml" )

	if hndle != nil {

		vsInterface, _ := IsVectorType( hndle )

		if vsInterface != nil {

			if  vsInterface.isVector() != true  {
				t.Errorf ("Error reading header %v\n", err)
			} else {
				fmt.Println("KML File Validated")
			}

		} else {
			t.Errorf ("Return interface is nil \n")
		}

	}

}

func TestValidLinearRingKML( t *testing.T ) {

	// Test reading Valid Shape File
	hndle, err := os.Open( "testfiles/cb_2016_us_state_20m.kml" )

	if hndle != nil {

		vsInterface, _ := IsVectorType( hndle )

		if vsInterface != nil {

			if  vsInterface.isVector() != true  {
				t.Errorf ("Error reading header %v\n", err)
			} else {
				fmt.Println("KML File Validated")
			}

		} else {
			t.Errorf ("Return interface is nil \n")
		}

	}

}

func TestPointsKML( t *testing.T ) {

	// Test reading Valid Shape File
	hndle, err := os.Open( "testfiles/cities.kml" )

	if hndle != nil {

		vsInterface, _ := IsVectorType( hndle )

		if vsInterface != nil {

			if  vsInterface.isVector() != true  {
				t.Errorf ("Error reading header %v\n", err)
			} else {
				fmt.Println("KML File Validated")
			}

		} else {
			t.Errorf ("Return interface is nil \n")
		}

	}

}

func TestLineStringKML( t *testing.T ) {

	// Test reading Valid Shape File
	hndle, err := os.Open( "testfiles/cta.kml" )

	if hndle != nil {

		vsInterface, _ := IsVectorType( hndle )

		if vsInterface != nil {

			if  vsInterface.isVector() != true  {
				t.Errorf ("Error reading header %v\n", err)
			} else {
				fmt.Println("KML File Validated")
			}

		} else {
			t.Errorf ("Return interface is nil \n")
		}

	}

}


