package vector

import (
	"fmt"
	"github.com/geodatalake/lambdas/geotiff"
	"encoding/binary"
	"math"
	"strings"
	"encoding/xml"
	"strconv"
)


type decoder struct {
	reader         VectorStream
	bounds         	*geotiff.Bounds
	bVector       	bool
	bKML        	bool
	bShape   	bool
	fileName 	string
	fileLength	uint32
}


func (d *decoder) Bounds() ( *geotiff.Bounds, error ) {
	return d.bounds, nil
}

func (d *decoder) GetFileLength() ( uint32 ) {
	return d.fileLength
}

func (d *decoder) IsVector() ( bool ) {
	return d.bVector
}

func (d *decoder) IsKML() ( bool ) {
	return d.bKML
}

func (d *decoder) IsShape() ( bool ) {
	return d.bShape
}

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func (r *decoder) getShapeInfo( vs VectorStream) {

	fileLength := make([]byte, 4)

	// file length
	vs.ReadAt( fileLength, 24 )
	r.fileLength = binary.BigEndian.Uint32(fileLength[0:4])

	r.bounds = &geotiff.Bounds{MinX: 0.0, MaxY: 0.0, MaxX: 0.0, MinY:0.0, OriginX: 0.0, OriginY: 0.0}

	var temp = make([]byte, 8)

	vs.ReadAt( temp, 36 )
	var tempFloat = binary.LittleEndian.Uint64( temp[0:8])
	fmt.Println( fmt.Sprintf("MinX in Float Form %f", math.Float64frombits(tempFloat)))
	r.bounds.MinX = math.Float64frombits(tempFloat)

	vs.ReadAt( temp, 44 )
	tempFloat = binary.LittleEndian.Uint64( temp[0:8])
	fmt.Println( fmt.Sprintf("MinY in Float Form %f", math.Float64frombits(tempFloat)))
	r.bounds.MinY = math.Float64frombits(tempFloat)

	vs.ReadAt( temp, 52 )
	tempFloat = binary.LittleEndian.Uint64( temp[0:8])
	fmt.Println( fmt.Sprintf("MaxX in Float Form %f", math.Float64frombits(tempFloat)))
	r.bounds.MaxX = math.Float64frombits(tempFloat)

	vs.ReadAt( temp, 60 )
	tempFloat = binary.LittleEndian.Uint64( temp[0:8])
	fmt.Println( fmt.Sprintf("MaxY in Float Form %f", math.Float64frombits(tempFloat)))
	r.bounds.MaxY = math.Float64frombits(tempFloat)

}

func TryShapeDecoder ( vs VectorStream ) ( VectorIntfc, error) {


	fileCodeBytes := make([]byte, 4)

	d := &decoder{
		reader: vs,
		bounds: nil,
		bVector: false,
		bKML: false,
		bShape: false,
	}

	fmt.Println( fmt.Sprintf("Initialized Reading Code %d", fileCodeBytes))

	vs.ReadAt( fileCodeBytes, 0 )

	fmt.Println( fmt.Sprintf("Data read code %d", fileCodeBytes))

	fileCode := binary.BigEndian.Uint32( fileCodeBytes[0:4] )

	fmt.Println( fmt.Sprintf("Reading Code %d", fileCode))

	if  fileCode !=  9994 {
		fmt.Println( "Not valid file")
		return d, fmt.Errorf("Error reading header reading shape file code, read %d", fileCode)
	}

	fmt.Println( "Done Reading Code")

	// We have validated file so time to build the vector and move ahead
        d.bShape = true
	d.bVector = true

	d.getShapeInfo( vs )

	return d, nil

}

func updateBounds ( candidate  *geotiff.Bounds,   masterBounds *geotiff.Bounds)  {


	if candidate.MinX < masterBounds.MinX {
		masterBounds.MinX = candidate.MinX
	}
	if candidate.MaxX > masterBounds.MaxX {
		masterBounds.MaxX = candidate.MaxX
	}

	if candidate.MinY < masterBounds.MinY {
		masterBounds.MinY = candidate.MinY
	}
	if candidate.MaxY > masterBounds.MaxY {
		masterBounds.MaxY = candidate.MaxY
	}



}



func (d *decoder) findMaxBounds(  coordList []string,  masterBounds *geotiff.Bounds ) {

	for _, coord := range coordList {

		coordsArray := strings.Split( coord, "\n" )

		for _, row := range coordsArray {

			//fmt.Println (row)

			elems := strings.Split( row, ",")

			if len( elems ) >= 2 {

				//fmt.Print("Lat: " + elems[0])
				//fmt.Print(" Lon: " + elems[1])

				newX, errorX := strconv.ParseFloat(strings.Trim(elems[0], " "), 64)
				newY, errorY := strconv.ParseFloat(strings.Trim(elems[1], " "), 64)

				if errorX == nil  && errorY == nil {
					if newX < masterBounds.MinX {
						masterBounds.MinX = newX
					}
					if newX > masterBounds.MaxX {
						masterBounds.MaxX = newX
					}

					if newY < masterBounds.MinY {
						masterBounds.MinY = newY
					}
					if newY > masterBounds.MaxY {
						masterBounds.MaxY = newY
					}

				} else {
					fmt.Println( "Errors")
					fmt.Println( errorX)
					fmt.Println( errorY)
				}
			}


		}


	}

}

/*
func (d *decoder) updateBoundaries( newBounds *geotiff.Bounds )  {


	return d.bounds, nil
}
*/

func TryKMLDecoder ( vs VectorStream ) ( VectorIntfc, error) {



	d := &decoder{
		reader: vs,
		bounds: nil,
		bVector: false,
		bKML: false,
		bShape: false,
	}

	kmlNameSpace := make([]byte, 500)
	vs.ReadAt(kmlNameSpace, 0)

	s := string(kmlNameSpace[:500])

	if  !strings.Contains( s, "kml" )   {
		return d, fmt.Errorf("Error reading header reading shape file code, read %d", kmlNameSpace)
	}


	// Setup a Decoder and start parsing the xml file
	decoder := xml.NewDecoder( vs )

	var inElement string

        // Initialize initial boundaries
	d.bounds = &geotiff.Bounds{MinX: 180.0, MaxY: -180.0, MaxX: -90.0, MinY: 90.0, OriginX: 0.0, OriginY: 0.0}


	// Loop through the document and get all the elements
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		elementBounds := &geotiff.Bounds{MinX: 180.0, MaxY: -180.0, MaxX: -90.0, MinY: 90.0, OriginX: 0.0, OriginY: 0.0}

		// Inspect the type of the token just read.
		switch se := t.(type) {

		case xml.StartElement:

			// If we just read a StartElement token
			inElement = se.Name.Local

			if inElement == "GroundOverlay" {

				var p GroundOverlay

				fmt.Println( "In Overlay")

				// decode a whole chunk of following XML into the
				// variable p which is a Overlay (se above)
				decoder.DecodeElement( &p, &se )

				//fmt.Println( p.Icon )
				//fmt.Println( p.LatLonBox )

				minx, _ := strconv.ParseFloat( p.LatLonBox.West, 64)
				maxx, _ := strconv.ParseFloat(p.LatLonBox.East,64)
				miny, _ := strconv.ParseFloat(p.LatLonBox.North,64)
				maxy, _  :=strconv.ParseFloat(p.LatLonBox.South, 64)
				lclCoords := &geotiff.Bounds{MinX: minx,
					MinY: miny,
					MaxX: maxx,
					MaxY: maxy,
					OriginX: 0.0, OriginY: 0.0}


				updateBounds( lclCoords, elementBounds)

			} else if inElement == "LinearRing" {
				var p LinearRing

				//fmt.Println( "Reading LinearRing")

				// decode a whole chunk of following XML into the
				// variable p which is a Overlay (se above)
				decoder.DecodeElement( &p, &se )
				d.findMaxBounds( p.Coordinates, elementBounds)
			}  else if inElement == "Point" {

				var p Point

				//fmt.Println( "Reading  Point")
				// decode a whole chunk of following XML into the
				// variable p which is a Overlay (se above)
				decoder.DecodeElement( &p, &se )
				d.findMaxBounds( p.Coordinates, elementBounds)

			} else if inElement == "LineString" {

				var p LineString

				//fmt.Println( "Reading LineString")
				// decode a whole chunk of following XML into the
				// variable p which is a Overlay (se above)
				decoder.DecodeElement( &p, &se )

				d.findMaxBounds( p.Coordinates, elementBounds)
			}

		default:
		}

		updateBounds( elementBounds, d.bounds)
	}


	// We have validated file so time to build the vector and move ahead
	d.bKML = true ;
	d.bVector = true ;


	fmt.Println ( fmt.Sprintf( "The Box %f, %f, %f, %f ", d.bounds.MinX, d.bounds.MinY, d.bounds.MaxX,d.bounds.MaxY))

	return d, nil


}

func IsVectorType( vs VectorStream ) (  VectorIntfc, error ) {

	// Test if vector data in Shapefile format
	var vInterface, err = TryShapeDecoder ( vs  )

	if err == nil {

		// We have a shape file hence a vector file so return the interface
		return vInterface, nil

	}

	fmt.Println( "Try KML Test")
	vInterface, err = TryKMLDecoder ( vs  )

	if err == nil {

		// We have a KML file hence a vector file so return the interface
		return vInterface, nil

	}

	return vInterface, fmt.Errorf("No Vector data types supported.")


}




