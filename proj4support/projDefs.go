package proj4support

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"github.com/geodatalake/geom/proj"
)

const EPSGTITLEFILENAME = "EPSGTITLE.bin"
const EPSGVALFILENAME  = "EPSGVAL.bin"

type EPSGForm struct {
	Projection *proj.SR
	Datum      proj.DatumExport
}


var defsByEPSGValue map[string] *EPSGForm
var defsByEPSGTitle map[string] *EPSGForm

func addDef(name, def string) error {

	if defsByEPSGValue == nil {
		defsByEPSGValue = make(map[string] *EPSGForm)
	}

	parseDef, err := proj.Parse( def)

	if err != nil {
		return err
	}

	var myProj *proj.SR = parseDef

	//encodedDatum,_ := proj.GetEncodedDatum( myProj)
	//fmt.Println( encodedDatum )

	var toSave = new(EPSGForm)

	toSave.Projection = myProj
	toSave.Datum = proj.DatumExposed( myProj)


	defsByEPSGValue[name] = toSave
	//fmt.Println( "print addDef")
	return nil
}

func AddDef(name, def string) error {

	return addDef(name, def )

}

func addTitleDef(name, def string) error {

	if defsByEPSGTitle == nil {
		defsByEPSGTitle = make(map[string]*EPSGForm)
	}
	theProj, err := proj.Parse(def)
	if err != nil {
		return err
	}

	var toSave = new(EPSGForm)

	toSave.Projection = theProj
	toSave.Datum = proj.DatumExposed( theProj)


	defsByEPSGTitle[name] = toSave

	//fmt.Println( "print addDef")
	return nil
}

func AddTitleDef(name, def string) error {
	return addTitleDef(name, def )
}

func GetDefByEPSG( name string )(*proj.SR, error){

	fmt.Println( "GetDefByEPSG " + name)
	fmt.Println(len(defsByEPSGValue))
	fmt.Println( defsByEPSGValue[name] )

	var epsgForm = defsByEPSGValue[name]

	proj.RestoreDatumExposed( epsgForm.Projection, epsgForm.Datum)

	if ( epsgForm != nil ) {
		proj.RestoreDatumExposed(epsgForm.Projection, epsgForm.Datum)

		return epsgForm.Projection, nil
	} else {
		return nil, fmt.Errorf("No Valid Projection found by EPSG Value Title]")
	}
}



func GetDefByTitle( name string ) (*proj.SR, error) {

	fmt.Println( "GetDefByTitle " + name)
	fmt.Println(len(defsByEPSGTitle))

	var epsgForm = defsByEPSGTitle[name]

	if ( epsgForm != nil ) {
		proj.RestoreDatumExposed(epsgForm.Projection, epsgForm.Datum)

		return epsgForm.Projection, nil
	} else {
		return nil, fmt.Errorf("No Valid Projection found for Title]")
	}
}

func SerializeMaps( ) {

	var b1 = new(bytes.Buffer)

	var e = gob.NewEncoder(b1)

	// Encode the maps
	var err = e.Encode( defsByEPSGValue )
	if err != nil {
		panic(err)
	}

	var encodedStr = base64.StdEncoding.EncodeToString( b1.Bytes() )
	var encodedBytes = []byte(encodedStr)

	err = ioutil.WriteFile( EPSGVALFILENAME, encodedBytes, 0644)
	if err != nil {
		panic(e)
	}

	var b2 = new(bytes.Buffer)

	e = gob.NewEncoder(b2)

	// Encode the maps
	err = e.Encode( defsByEPSGTitle )
	if err != nil {
		panic(err)
	}

	encodedStr = base64.StdEncoding.EncodeToString( b2.Bytes() )
	encodedBytes = []byte(encodedStr)

	err = ioutil.WriteFile( EPSGTITLEFILENAME, encodedBytes, 0644)
	if err != nil {
		panic(e)
	}

}

func CheckAndLoadMaps() ( bool ) {

	var loaded = false

	var dataRead, err1 = ioutil.ReadFile(EPSGVALFILENAME)
	if ( err1 == nil ) {

		//var n = bytes.Index( dataRead, []byte{0} )
		var s = string( dataRead[:len(dataRead)] )

		var by, err = base64.StdEncoding.DecodeString(s)
		if ( err != nil ) {
			return loaded
		}

		b := bytes.Buffer{}
		b.Write(by)

		d := gob.NewDecoder(&b)
		err = d.Decode( &defsByEPSGValue)

		if err == nil {

			dataRead, _ = ioutil.ReadFile(EPSGTITLEFILENAME)
			//var n = bytes.Index(dataRead, []byte{0})
			var s = string(dataRead[:len(dataRead)])

			by, err = base64.StdEncoding.DecodeString(s)
			if ( err != nil ) {
				return loaded
			}

			b := bytes.Buffer{}
			b.Write(by)

			d := gob.NewDecoder(&b)
			err = d.Decode(&defsByEPSGTitle)

			if err != nil {
				fmt.Println("EPSG Title file not available")
			} else {
				loaded = true
			}
		}

	} else {
		fmt.Println("EPSG Value file not available")
	}

	return loaded

}

