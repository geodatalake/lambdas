package proj4support

import (
	"github.com/ctessum/geom/proj"
)

var defs map[string]*proj.SR

func addDef(name, def string) error {
	if defs == nil {
		defs = make(map[string]*proj.SR)
	}
	proj, err := proj.Parse(def)
	if err != nil {
		return err
	}
	defs[name] = proj
	return nil
}

func AddDef(name, def string) error {

	return addDef(name, def )

}


func GetDef( name string ) *proj.SR {

	return defs[name]
}
