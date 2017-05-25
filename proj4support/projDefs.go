package proj4support

import (
	"github.com/ctessum/geom/proj"
)

var defsByEPSGValue map[string]*proj.SR
var defsByEPSGTitle map[string]*proj.SR

func addDef(name, def string) error {

	if defsByEPSGValue == nil {
		defsByEPSGValue = make(map[string]*proj.SR)
	}
	proj, err := proj.Parse(def)
	if err != nil {
		return err
	}
	defsByEPSGValue[name] = proj
	return nil
}

func AddDef(name, def string) error {

	return addDef(name, def )

}

func addTitleDef(name, def string) error {

	if defsByEPSGTitle == nil {
		defsByEPSGTitle = make(map[string]*proj.SR)
	}
	proj, err := proj.Parse(def)
	if err != nil {
		return err
	}
	defsByEPSGTitle[name] = proj
	return nil
}

func AddTitleDef(name, def string) error {
	return addTitleDef(name, def )
}

func GetDefByEPSG( name string ) *proj.SR {

	return defsByEPSGValue[name]
}


func GetDefByTitle( name string ) *proj.SR {

	return defsByEPSGTitle[name]
}


