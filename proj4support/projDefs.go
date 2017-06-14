package proj4support

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/geodatalake/geom/proj"
)

const EPSGTITLEFILENAME = "EPSGTITLE.bin"
const EPSGVALFILENAME = "EPSGVAL.bin"

type EPSGForm struct {
	Projection *proj.SR
	Datum      proj.DatumExport
}

var defsByEPSGValue map[string]*EPSGForm

var DefsByEPSGGCS = map[string]string{
	"GCS_Adindan":             "EPSG:4201",
	"GCS_AGD66":               "EPSG:4202",
	"GCS_AGD84":               "EPSG:4203",
	"GCS_Ain_el_Abd":          "EPSG:4204",
	"GCS_Afgooye":             "EPSG:4205",
	"GCS_Agadez":              "EPSG:4206",
	"GCS_Lisbon":              "EPSG:4207",
	"GCS_Aratu":               "EPSG:4208",
	"GCS_Arc_1950":            "EPSG:4209",
	"GCS_Arc_1960":            "EPSG:4210",
	"GCS_Batavia":             "EPSG:4211",
	"GCS_Barbados":            "EPSG:4212",
	"GCS_Beduaram":            "EPSG:  4213",
	"GCS_Beijing_1954":        "EPSG:4214",
	"GCS_Belge_1950":          "EPSG:4215",
	"GCS_Bermuda_1957":        "EPSG:4216",
	"GCS_Bern_1898":           "EPSG:4217",
	"GCS_Bogota":              "EPSG:4218",
	"GCS_Bukit_Rimpah":        "EPSG:4219",
	"GCS_Camacupa":            "EPSG:4220",
	"GCS_Campo_Inchauspe":     "EPSG:4221",
	"GCS_Cape":                "EPSG:4222",
	"GCS_Carthage":            "EPSG:4223",
	"GCS_Chua":                "EPSG:4224",
	"GCS_Corrego_Alegre":      "EPSG:4225",
	"GCS_Cote_d_Ivoire":       "EPSG:4226",
	"GCS_Deir_ez_Zor":         "EPSG:4227",
	"GCS_Douala":              "EPSG:4228",
	"GCS_Egypt_1907":          "EPSG:4229",
	"GCS_ED50":                "EPSG:4230",
	"GCS_ED87":                "EPSG:4231",
	"GCS_Fahud":               "EPSG:4232",
	"GCS_Gandajika_1970":      "EPSG:4233",
	"GCS_Garoua":              "EPSG:4234",
	"GCS_Guyane_Francaise":    "EPSG:4235",
	"GCS_Hu_Tzu_Shan":         "EPSG:4236",
	"GCS_HD72":                "EPSG:4237",
	"GCS_ID74":                "EPSG:4238",
	"GCS_Indian_1954":         "EPSG:4239",
	"GCS_Indian_1975":         "EPSG:4240",
	"GCS_Jamaica_1875":        "EPSG:4241",
	"GCS_JAD69":               "EPSG:4242",
	"GCS_Kalianpur":           "EPSG:4243",
	"GCS_Kandawala":           "EPSG:4244",
	"GCS_Kertau":              "EPSG:4245",
	"GCS_KOC":                 "EPSG:4246",
	"GCS_La_Canoa":            "EPSG:4247",
	"GCS_PSAD56":              "EPSG:4248",
	"GCS_Lake":                "EPSG:4249",
	"GCS_Leigon":              "EPSG:4250",
	"GCS_Liberia_1964":        "EPSG:4251",
	"GCS_Lome":                "EPSG:4252",
	"GCS_Luzon_1911":          "EPSG:4253",
	"GCS_Hito_XVIII_1963":     "EPSG:4254",
	"GCS_Herat_North":         "EPSG:4255",
	"GCS_Mahe_1971":           "EPSG:4256",
	"GCS_Makassar":            "EPSG:4257",
	"GCS_EUREF89":             "EPSG:4258",
	"GCS_Malongo_1987":        "EPSG:4259",
	"GCS_Manoca":              "EPSG:4260",
	"GCS_Merchich":            "EPSG:4261",
	"GCS_Massawa":             "EPSG:4262",
	"GCS_Minna":               "EPSG:4263",
	"GCS_Mhast":               "EPSG:4264",
	"GCS_Monte_Mario":         "EPSG:4265",
	"GCS_M_poraloko":          "EPSG:4266",
	"GCS_NAD27":               "EPSG:4267",
	"GCS_NAD_Michigan":        "EPSG:4268",
	"GCS_NAD83":               "EPSG:4269",
	"GCS_Nahrwan_1967":        "EPSG:4270",
	"GCS_Naparima_1972":       "EPSG:4271",
	"GCS_GD49":                "EPSG:4272",
	"GCS_NGO_1948":            "EPSG:4273",
	"GCS_Datum_73":            "EPSG:4274",
	"GCS_NTF":                 "EPSG:4275",
	"GCS_NSWC_9Z_2":           "EPSG:4276",
	"GCS_OSGB_1936":           "EPSG:4277",
	"GCS_OSGB70":              "EPSG:4278",
	"GCS_OS_SN80":             "EPSG:4279",
	"GCS_Padang":              "EPSG:4280",
	"GCS_Palestine_1923":      "EPSG:4281",
	"GCS_Pointe_Noire":        "EPSG:4282",
	"GCS_GDA94":               "EPSG:4283",
	"GCS_Pulkovo_1942":        "EPSG:4284",
	"GCS_Qatar":               "EPSG:4285",
	"GCS_Qatar_1948":          "EPSG:4286",
	"GCS_Qornoq":              "EPSG:4287",
	"GCS_Loma_Quintana":       "EPSG:4288",
	"GCS_Amersfoort":          "EPSG:4289",
	"GCS_RT38":                "EPSG:4290",
	"GCS_SAD69":               "EPSG:4291",
	"GCS_Sapper_Hill_1943":    "EPSG:4292",
	"GCS_Schwarzeck":          "EPSG:4293",
	"GCS_Segora":              "EPSG:4294",
	"GCS_Serindung":           "EPSG:4295",
	"GCS_Sudan":               "EPSG:4296",
	"GCS_Tananarive":          "EPSG:4297",
	"GCS_Timbalai_1948":       "EPSG:4298",
	"GCS_TM65":                "EPSG:4299",
	"GCS_TM75":                "EPSG:4300",
	"GCS_Tokyo":               "EPSG:4301",
	"GCS_Trinidad_1903":       "EPSG:4302",
	"GCS_TC_1948":             "EPSG:4303",
	"GCS_Voirol_1875":         "EPSG:4304",
	"GCS_Voirol_Unifie":       "EPSG:4305",
	"GCS_Bern_1938":           "EPSG:4306",
	"GCS_Nord_Sahara_1959":    "EPSG:4307",
	"GCS_Stockholm_1938":      "EPSG: 4308",
	"GCS_Yacare":              "EPSG:4309",
	"GCS_Yoff":                "EPSG:4310",
	"GCS_Zanderij":            "EPSG:4311",
	"GCS_MGI":                 "EPSG:4312",
	"GCS_Belge_1972":          "EPSG:4313",
	"GCS_DHDN":                "EPSG:4314",
	"GCS_Conakry_1905":        "EPSG:4315",
	"GCS_WGS_72":              "EPSG:4322",
	"GCS_WGS_72BE":            "EPSG:4324",
	"GCS_WGS_84":              "EPSG:4326",
	"GCS_Bern_1898_Bern":      "EPSG:4801",
	"GCS_Bogota_Bogota":       "EPSG:4802",
	"GCS_Lisbon_Lisbon":       "EPSG:4803",
	"GCS_Makassar_Jakarta":    "EPSG:4804",
	"GCS_MGI_Ferro":           "EPSG:4805",
	"GCS_Monte_Mario_Rome":    "EPSG:4806",
	"GCS_NTF_Paris":           "EPSG:4807",
	"GCS_Padang_Jakarta":      "EPSG:4808",
	"GCS_Belge_1950_Brussels": "EPSG:4809",
	"GCS_Tananarive_Paris":    "EPSG:4810",
	"GCS_Voirol_1875_Paris":   "EPSG:4811",
	"GCS_Voirol_Unifie_Paris": "EPSG:4812",
	"GCS_Batavia_Jakarta":     "EPSG:4813",
	"GCS_ATF_Paris":           "EPSG:4901",
	"GCS_NDG_Paris":           "EPSG: 4902",
}

func addDef(name, def string) error {

	if defsByEPSGValue == nil {
		defsByEPSGValue = make(map[string]*EPSGForm)
	}

	parseDef, err := proj.Parse(def)

	if err != nil {
		return err
	}

	var myProj *proj.SR = parseDef

	var toSave = new(EPSGForm)

	toSave.Projection = myProj
	toSave.Datum = proj.DatumExposed(myProj)

	defsByEPSGValue[name] = toSave

	return nil
}

func AddDef(name, def string) error {
	return addDef(name, def)
}

func GetDefByEPSG(name string) (*proj.SR, error) {
	if epsgForm, ok := defsByEPSGValue[name]; ok {
		proj.RestoreDatumExposed(epsgForm.Projection, epsgForm.Datum)
		return epsgForm.Projection, nil
	} else {
		if projString, found := EpsgToProjection[name]; found {
			if err := addDef(name, projString); err == nil {
				return GetDefByEPSG(name)
			}
		}
		return nil, fmt.Errorf("No Valid Projection found by EPSG Value: %s", name)
	}
}

func GetDefByTitle(name string) (*proj.SR, error) {
	if code, found := TitleToEpsg[name]; found {
		return GetDefByEPSG(code)
	}
	return nil, fmt.Errorf("No Valid Projection found for Title: %s", name)
}

func FormatMaps(inpath string, outpath string) {
	file, err := os.Open(inpath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var bufEpsg bytes.Buffer
	var bufTitle bytes.Buffer
	allTitles := make(map[string]bool)
	bufEpsg.WriteString("package proj4support\n\n")
	bufEpsg.WriteString("var EpsgToProjection = map[string]string{\n")
	bufTitle.WriteString("package proj4support\n\n")
	bufTitle.WriteString("var TitleToEspg = map[string]string{\n")
	var epsgTitle string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Look for EPSG Name First
		newLine := scanner.Text()
		if strings.IndexAny(newLine, "#") == 0 {
			if len(strings.Trim(newLine, " ")) > 1 {
				epsgTitle = strings.ToUpper(
					strings.TrimSpace(
						strings.Replace(newLine[2:], "/", "", -1)))
			} else {
				epsgTitle = ""
			}
		} else if strings.IndexAny(newLine, "<") == 0 && len(epsgTitle) > 0 {
			// Handle getting the EPSG code and Proj4 String
			epsgIndex := strings.IndexAny(newLine, ">")
			var epsgCode = newLine[1:epsgIndex]
			epsgCode = "EPSG:" + epsgCode

			// Now get proj string
			var projString = newLine[epsgIndex+1:]

			projString = strings.TrimSpace(
				strings.Replace(projString, "<>", "", -1))

			totalString := "+title=" + epsgTitle + " " + projString
			bufEpsg.WriteString(fmt.Sprintf("\t\"%s\": \"%s\",\n", epsgCode, totalString))
			newTitle := strings.Replace(epsgTitle, " ", "", -1)
			if _, ok := allTitles[newTitle]; !ok {
				bufTitle.WriteString(fmt.Sprintf("\t\"%s\": \"%s\",\n", newTitle, epsgCode))
				allTitles[newTitle] = true
			} else {
				log.Println("Found a duplicate title", newTitle, epsgCode)
			}
		}
	}
	bufEpsg.WriteString("}")
	if f, err0 := os.Create(path.Join(outpath, "epsg_map.go")); err0 == nil {
		f.WriteString(bufEpsg.String())
		f.Close()
	} else {
		log.Println(err0)
	}
	bufTitle.WriteString("}")
	if f1, err1 := os.Create(path.Join(outpath, "title_map.go")); err1 == nil {
		f1.WriteString(bufTitle.String())
		f1.Close()
	} else {
		log.Println(err1)
	}
}
