package proj4support

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"github.com/geodatalake/geom/proj"
	"os"
	"log"
	"bufio"
	"strings"

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

	//fmt.Println( "GetDefByTitle " + name)
	//fmt.Println(len(defsByEPSGTitle))

	//for cnt,src := range defsByEPSGTitle {
	//	fmt.Println( cnt)
	//	fmt.Println( src)
	//}
	var epsgForm = defsByEPSGTitle[name]

	if ( epsgForm != nil ) {
		proj.RestoreDatumExposed(epsgForm.Projection, epsgForm.Datum)

		return epsgForm.Projection, nil
	} else {
		return nil, fmt.Errorf("No Valid Projection found for Title]")
	}
}

func SerializeMaps( outpath string  ) {

	var b1 = new(bytes.Buffer)

	var e = gob.NewEncoder(b1)

	// Encode the maps
	var err = e.Encode( defsByEPSGValue )
	if err != nil {
		panic(err)
	}

	var outModified = ""
	if len(outpath) > 0 {
		outModified = outpath + "/"
	}

	var encodedStr = base64.StdEncoding.EncodeToString( b1.Bytes() )
	var encodedBytes = []byte(encodedStr)

	err = ioutil.WriteFile( outModified + EPSGVALFILENAME, encodedBytes, 0644)
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

	err = ioutil.WriteFile( outModified + EPSGTITLEFILENAME, encodedBytes, 0644)
	if err != nil {
		panic(e)
	}

}

func CheckAndLoadMaps( pathToLoad string ) ( bool ) {

	var loaded = false

	var inModified = ""
	if len(pathToLoad) > 0 {
		inModified = pathToLoad + "/"
	}

	fmt.Println("Passed pathToLoad" + pathToLoad)
	fmt.Println("Passed inModified" + inModified)

	var dataRead, err1 = ioutil.ReadFile(inModified + EPSGVALFILENAME)
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

			dataRead, _ = ioutil.ReadFile(inModified + EPSGTITLEFILENAME)
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

func LoadMaps() ( bool ) {

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

func BuildMaps( inpath string, outpath string ) ( bool ) {


	var built = false
	file, err := os.Open( inpath )

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var epsgTitle = ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		// Look for EPSG Name First
		newLine := scanner.Text()
		if strings.IndexAny(newLine, "#") == 0 {

			if len(strings.Trim(newLine, " ")) > 1 {
				epsgTitle = newLine[2:len(newLine)]
				epsgTitle = strings.Replace(epsgTitle, "/", "", -1)
				epsgTitle = strings.TrimSpace(epsgTitle)
				epsgTitle = strings.ToUpper(epsgTitle)
			} else {
				epsgTitle = ""
			}
		} else if strings.IndexAny(newLine, "<") == 0 && len(epsgTitle) > 0 {

			// Handle getting the EPSG code and Proj4 String
			epsgIndex := strings.IndexAny(newLine, ">")
			var epsgCode = newLine[1:epsgIndex]
			epsgCode = "EPSG:" + epsgCode

			// Now get proj string
			var projString = newLine[epsgIndex + 1:len(newLine)]

			projString = strings.Replace(projString, "<>", "", -1)
			projString = strings.TrimSpace(projString)

			totalString := "+title=" + epsgTitle + " " + projString

			AddDef(epsgCode, totalString)
			//fmt.Println( "Title is " + epsgTitle + " " + totalString)
			var newTitle = strings.Replace(epsgTitle, " ", "", -1)
			AddTitleDef(newTitle, totalString)
		}
	}

	SerializeMaps( outpath )

	built = true

	return  built
}


