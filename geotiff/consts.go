// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geotiff

// The TIFF specification is at http://partners.adobe.com/public/developer/en/tiff/TIFF6.pdf

const (
	leHeader = "II\x2A\x00" // Header for little-endian files.
	beHeader = "MM\x00\x2A" // Header for big-endian files.

	ifdLen = 12 // Length of an IFD entry in bytes.
)

const (
	Wgs84SemiMajorAxis    = 6378137.0
	Wgs84FlatteningFactor = 298.257223563
	KeyDirectoryVersion   = 1
	KeyRevision           = 1
	MinorRevision         = 0
	DE2RA                 = 0.01745329252
	RA2DE                 = 57.2957795129
	FLATTENING            = 1.000000 / 298.257223563 // Earth flattening (WGS84)
	ERAD                  = 6378.137                 // Km
	ERADM                 = 6378137.0                // meters
)

// Data types (p. 14-16 of the spec).
const (
	dtByte      = 1
	dtASCII     = 2
	dtShort     = 3
	dtLong      = 4
	dtRational  = 5
	dtSbyte     = 6
	dtUndefined = 7
	dtSshort    = 8
	dtSlong     = 9
	dtSrational = 10
	dtFloat     = 11
	dtDouble    = 12
)

const (
	DtByte      = 1
	DtASCII     = 2
	DtShort     = 3
	DtLong      = 4
	DtRational  = 5
	DtSbyte     = 6
	DtUndefined = 7
	DtSshort    = 8
	DtSlong     = 9
	DtSrational = 10
	DtFloat     = 11
	DtDouble    = 12
)

// The length of one instance of each data type in bytes.
var lengths = [...]uint32{0, 1, 1, 2, 4, 8, 1, 1, 2, 4, 8, 4, 8}

// Tags (see p. 28-41 of the spec).
const (
	tImageWidth                = 256
	tImageLength               = 257
	tBitsPerSample             = 258
	tCompression               = 259
	tPhotometricInterpretation = 262

	tStripOffsets    = 273
	tSamplesPerPixel = 277
	tRowsPerStrip    = 278
	tStripByteCounts = 279

	tTileWidth      = 322
	tTileLength     = 323
	tTileOffsets    = 324
	tTileByteCounts = 325

	tNewSubfileType      = 254
	tSubfileType         = 255
	tThreshholding       = 263
	tCellWidth           = 264
	tCellLength          = 265
	tFillOrder           = 266
	tDocumentName        = 269
	tImageDescription    = 270
	tModel               = 272
	tMake                = 271
	tOrientation         = 274
	tMinSampleValue      = 280
	tMaxSampleValue      = 281
	tPlanarConfiguration = 284

	tXResolution    = 282
	tYResolution    = 283
	tResolutionUnit = 296

	tFreeOffsets       = 288
	tFreeByteCounts    = 289
	tGrayResponseUnit  = 290
	tGrayResponseCurve = 291

	tPredictor    = 317
	tColorMap     = 320
	tExtraSamples = 338
	tSampleFormat = 339

	tSoftware     = 305
	tDateTime     = 306
	TDateTime     = 306
	tArtist       = 315
	tHostComputer = 316

	tCopyright = 33432

	tGeoKeys    = 34735
	tGeoDoubles = 34736
	tGeoAscii   = 34737

	tModelTiepointTag       = 33922
	tModelPixelScaleTag     = 33550
	tModelTransformationTag = 34264

	tGDALMetadata = 42112
	tGDALNodata   = 42113
)

// Compression types (defined in various places in the spec and supplements).
const (
	cNone       = 1
	cCCITT      = 2
	cG3         = 3 // Group 3 Fax.
	cG4         = 4 // Group 4 Fax.
	cLZW        = 5
	cJPEGOld    = 6 // Superseded by cJPEG.
	cJPEG       = 7
	cDeflate    = 8 // zlib compression.
	cPackBits   = 32773
	cDeflateOld = 32946 // Superseded by cDeflate.
)

// CompressionType describes the type of compression used in Options for writer.
type CompressionType int

const (
	Uncompressed CompressionType = iota
	Deflate
)

// optToCompression returns the compression type constant from the TIFF spec that
// is equivalent to c.
func (c CompressionType) optToCompression() uint32 {
	switch c {
	case Deflate:
		return cDeflate
	}
	return cNone
}

// Photometric interpretation values (see p. 37 of the spec).
const (
	pWhiteIsZero = 0
	pBlackIsZero = 1
	pRGB         = 2
	pPaletted    = 3
	pTransMask   = 4 // transparency mask
	pCMYK        = 5
	pYCbCr       = 6
	pCIELab      = 8
)

// Values for the tPredictor tag (page 64-65 of the spec).
const (
	prNone       = 1
	prHorizontal = 2
)

// Values for the tResolutionUnit tag (page 18).
const (
	resNone    = 1
	resPerInch = 2 // Dots per inch.
	resPerCM   = 3 // Dots per centimeter.
)

// Values for tSampleFormat
const (
	smplUint      = 1
	smplSint      = 2
	smplFloat     = 3
	smplUndefined = 4
)

// imageMode represents the mode of the image.
type imageMode int

const (
	mBilevel imageMode = iota
	mPaletted
	mGray
	mGrayInvert
	mRGB
	mRGBA
	mNRGBA
)

// geokeys
const (
	gkGTModelTypeGeoKey              = 1024
	gkGTRasterTypeGeoKey             = 1025
	gkGTCitationGeoKey               = 1026
	gkGeographicTypeGeoKey           = 2048
	gkGeogCitationGeoKey             = 2049
	gkGeogGeodeticDatumGeoKey        = 2050
	gkGeogPrimeMeridianGeoKey        = 2051
	gkGeogLinearUnitsGeoKey          = 2052
	gkGeogLinearUnitSizeGeoKey       = 2053
	gkGeogAngularUnitsGeoKey         = 2054
	gkGeogAngularUnitSizeGeoKey      = 2055
	gkGeogEllipsoidGeoKey            = 2056
	gkGeogSemiMajorAxisGeoKey        = 2057
	gkGeogSemiMinorAxisGeoKey        = 2058
	gkGeogInvFlatteningGeoKey        = 2059
	gkGeogAzimuthUnitsGeoKey         = 2060
	gkGeogPrimeMeridianLongGeoKey    = 2061
	gkProjectedCSTypeGeoKey          = 3072
	gkPCSCitationGeoKey              = 3073
	gkProjectionGeoKey               = 3074
	gkProjCoordTransGeoKey           = 3075
	gkProjLinearUnitsGeoKey          = 3076
	gkProjLinearUnitSizeGeoKey       = 3077
	gkProjStdParallel1GeoKey         = 3078
	gkProjStdParallel2GeoKey         = 3079
	gkProjNatOriginLongGeoKey        = 3080
	gkProjNatOriginLatGeoKey         = 3081
	gkProjFalseEastingGeoKey         = 3082
	gkProjFalseNorthingGeoKey        = 3083
	gkProjFalseOriginLongGeoKey      = 3084
	gkProjFalseOriginLatGeoKey       = 3085
	gkProjFalseOriginEastingGeoKey   = 3086
	gkProjFalseOriginNorthingGeoKey  = 3087
	gkProjCenterLongGeoKey           = 3088
	gkProjCenterLatGeoKey            = 3089
	gkProjCenterEastingGeoKey        = 3090
	gkProjCenterNorthingGeoKey       = 3091
	gkProjScaleAtNatOriginGeoKey     = 3092
	gkProjScaleAtCenterGeoKey        = 3093
	gkProjAzimuthAngleGeoKey         = 3094
	gkProjStraightVertPoleLongGeoKey = 3095
	gkVerticalCSTypeGeoKey           = 4096
	gkVerticalCitationGeoKey         = 4097
	gkVerticalDatumGeoKey            = 4098
	gkVerticalUnitsGeoKey            = 4099
)

var CompressionTypes map[int]string = map[int]string{
	cNone:       "None",
	cCCITT:      "CCITT",
	cG3:         "Group 3 Fax",
	cG4:         "Group 4 Fax",
	cLZW:        "LZW",
	cJPEGOld:    "Old JPEG",
	cJPEG:       "JPEG",
	cDeflate:    "Deflate",
	cPackBits:   "Packbits",
	cDeflateOld: "Old Deflate"}

var PhotoMetricInterpretation map[int]string = map[int]string{
	pWhiteIsZero: "White Is Zero",
	pBlackIsZero: "Black is Zero",
	pRGB:         "RGB",
	pPaletted:    "Paletted",
	pTransMask:   "Transparency Mask",
	pCMYK:        "CMYK",
	pCIELab:      "CIELab"}

var Resolutions map[int]string = map[int]string{
	resNone:    "None",
	resPerInch: "PerInch",
	resPerCM:   "PerCM"}

var SampleFormatTypes map[int]string = map[int]string{
	smplUint:      "Unsigned Integer",
	smplSint:      "Twos Compliment Signed Integer",
	smplFloat:     "IEEE Floating Point",
	smplUndefined: "Undefined"}

var TiffTags map[int]string = map[int]string{
	tImageWidth:                "ImageWidth",
	tImageLength:               "ImageLength",
	tBitsPerSample:             "BitsPerSample",
	tCompression:               "Compression",
	tPhotometricInterpretation: "PhotometricInterpretation",

	tStripOffsets:    "StripOffsets",
	tSamplesPerPixel: "SamplesPerPixel",
	tRowsPerStrip:    "RowsPerStrip",
	tStripByteCounts: "StripByteCounts",

	tTileWidth:      "TileWidth",
	tTileLength:     "TileLength",
	tTileOffsets:    "TileOffsets",
	tTileByteCounts: "TileByteCounts",

	tNewSubfileType:      "NewSubfileType",
	tSubfileType:         "SubfileType",
	tThreshholding:       "Threshholding",
	tCellWidth:           "CellWidth",
	tCellLength:          "CellLength",
	tFillOrder:           "FillOrder",
	tDocumentName:        "DocumentName",
	tImageDescription:    "ImageDescription",
	tModel:               "Model",
	tMake:                "Make",
	tOrientation:         "Orientation",
	tMinSampleValue:      "MinSampleValue",
	tMaxSampleValue:      "MaxSampleValue",
	tPlanarConfiguration: "PlanarConfiguration",

	tXResolution:    "XResolution",
	tYResolution:    "YResolution",
	tResolutionUnit: "ResolutionUnit",

	tFreeOffsets:       "FreeOffsets",
	tFreeByteCounts:    "FreeByteCounts",
	tGrayResponseUnit:  "GrayResponseUnit",
	tGrayResponseCurve: "GrayResponseCurve",

	tPredictor:    "Predictor",
	tColorMap:     "ColorMap",
	tExtraSamples: "ExtraSamples",
	tSampleFormat: "SampleFormat",

	tSoftware:     "Software",
	tDateTime:     "DateTime",
	tArtist:       "Artist",
	tHostComputer: "HostComputer",

	tCopyright: "Copyright",

	tGeoKeys:    "GeoKeys",
	tGeoDoubles: "GeoDoubles",
	tGeoAscii:   "GeoAscii",

	tModelTiepointTag:       "ModelTiepointTag",
	tModelPixelScaleTag:     "ModelPixelScaleTag",
	tModelTransformationTag: "ModelTransformationTag",

	tGDALMetadata: "GDALMetadata",
	tGDALNodata:   "GDALNodata"}

var DataTypes map[int]string = map[int]string{
	dtByte:      "Byte",
	dtASCII:     "Ascii",
	dtShort:     "Short",
	dtLong:      "Long",
	dtRational:  "Rational",
	dtSbyte:     "SignedByte",
	dtUndefined: "Undefined",
	dtSshort:    "SignedShort",
	dtSlong:     "SignedLong",
	dtSrational: "SignedRational",
	dtFloat:     "Float",
	dtDouble:    "Double"}

var GeoKeys map[int]string = map[int]string{
	1024: "GTModelTypeGeoKey",
	1025: "GTRasterTypeGeoKey",
	1026: "GTCitationGeoKey",
	2048: "GeographicTypeGeoKey",
	2049: "GeogCitationGeoKey",
	2050: "GeogGeodeticDatumGeoKey",
	2051: "GeogPrimeMeridianGeoKey",
	2052: "GeogLinearUnitsGeoKey",
	2053: "GeogLinearUnitSizeGeoKey",
	2054: "GeogAngularUnitsGeoKey",
	2055: "GeogAngularUnitSizeGeoKey",
	2056: "GeogEllipsoidGeoKey",
	2057: "GeogSemiMajorAxisGeoKey",
	2058: "GeogSemiMinorAxisGeoKey",
	2059: "GeogInvFlatteningGeoKey",
	2060: "GeogAzimuthUnitsGeoKey",
	2061: "GeogPrimeMeridianLongGeoKey",
	3072: "ProjectedCSTypeGeoKey",
	3073: "PCSCitationGeoKey",
	3074: "ProjectionGeoKey",
	3075: "ProjCoordTransGeoKey",
	3076: "ProjLinearUnitsGeoKey",
	3077: "ProjLinearUnitSizeGeoKey",
	3078: "ProjStdParallel1GeoKey",
	3079: "ProjStdParallel2GeoKey",
	3080: "ProjNatOriginLongGeoKey",
	3081: "ProjNatOriginLatGeoKey",
	3082: "ProjFalseEastingGeoKey",
	3083: "ProjFalseNorthingGeoKey",
	3084: "ProjFalseOriginLongGeoKey",
	3085: "ProjFalseOriginLatGeoKey",
	3086: "ProjFalseOriginEastingGeoKey",
	3087: "ProjFalseOriginNorthingGeoKey",
	3088: "ProjCenterLongGeoKey",
	3089: "ProjCenterLatGeoKey",
	3090: "ProjCenterEastingGeoKey",
	3091: "ProjCenterNorthingGeoKey",
	3092: "ProjScaleAtNatOriginGeoKey",
	3093: "ProjScaleAtCenterGeoKey",
	3094: "ProjAzimuthAngleGeoKey",
	3095: "ProjStraightVertPoleLongGeoKey",
	4096: "VerticalCSTypeGeoKey",
	4097: "VerticalCitationGeoKey",
	4098: "VerticalDatumGeoKey",
	4099: "VerticalUnitsGeoKey"}

var GTModelTypeGeoKey map[int]string = map[int]string{
	0: "ModelTypeUndefined",
	1: "ModelTypeProjected",
	2: "ModelTypeGeographic",
	3: "ModelTypeGeocentric"}

var GTRasterTypeGeoKey map[int]string = map[int]string{
	1: "RasterPixelIsArea",
	2: "RasterPixelIsPoint"}

var GeographicTypeGeoKey map[int]string = map[int]string{
	4014:  "GCSE_Clarke1880_SGA1922",
	4017:  "GCSE_Everest1830_1975Definition",
	4002:  "GCSE_AiryModified1849",
	4003:  "GCSE_AustralianNationalSpheroid",
	4012:  "GCSE_Clarke1880_RGS",
	4004:  "GCSE_Bessel1841",
	4022:  "GCSE_International1924",
	4030:  "GCSE_WGS84",
	4006:  "GCSE_BesselNamibia",
	4011:  "GCSE_Clarke1880_IGN",
	4283:  "GCS_GDA94",
	4010:  "GCSE_Clarke1880_Benoit",
	4326:  "GCS_WGS_84",
	4267:  "GCS_NAD27",
	4013:  "GCSE_Clarke1880_Arc",
	4019:  "GCSE_GRS1980",
	4023:  "GCSE_International1967",
	4024:  "GCSE_Krassowsky1940",
	4001:  "GCSE_Airy1830",
	4008:  "GCSE_Clarke1866",
	4005:  "GCSE_Bessel1841Modified",
	32767: "user-defined",
	4269:  "GCS_NAD83",
	4016:  "GCSE_Everest1830_1967Definition",
	4020:  "GCSE_Helmert1906",
	4009:  "GCSE_Clarke1866Michigan",
	4015:  "GCSE_Everest1830_1937Adjustment",
	4289:  "GCS_Amersfoort",
	4018:  "GCSE_Everest1830Modified",
	4034:  "GCSE_Clarke1880",
	4322:  "GCS_WGS_72"}

var GeogGeodeticDatumGeoKey map[int]string = map[int]string{
	6012:  "DatumE_Clarke1880_RGS",
	6283:  "Datum_Geocentric_Datum_of_Australia_1994",
	6034:  "DatumE_Clarke1880",
	6009:  "DatumE_Clarke1866Michigan",
	6010:  "DatumE_Clarke1880_Benoit",
	6024:  "DatumE_Krassowsky1940",
	6269:  "Datum_North_American_Datum_1983",
	6001:  "DatumE_Airy1830",
	6018:  "DatumE_Everest1830Modified",
	6089:  "Datum_Amersfoort",
	6023:  "DatumE_International1967",
	6322:  "Datum_WGS72",
	6002:  "DatumE_AiryModified1849",
	6203:  "Datum_Australian_Geodetic_Datum_1984",
	6003:  "DatumE_AustralianNationalSpheroid",
	6013:  "DatumE_Clarke1880_Arc",
	6030:  "DatumE_WGS84",
	6017:  "DatumE_Everest1830_1975Definition",
	6015:  "DatumE_Everest1830_1937Adjustment",
	6019:  "DatumE_GRS1980",
	6020:  "DatumE_Helmert1906",
	6202:  "Datum_Australian_Geodetic_Datum_1966",
	6005:  "DatumE_BesselModified",
	32767: "user-defined",
	6016:  "DatumE_Everest1830_1967Definition",
	6004:  "DatumE_Bessel1841",
	6011:  "DatumE_Clarke1880_IGN",
	6014:  "DatumE_Clarke1880_SGA1922",
	6006:  "DatumE_BesselNamibia",
	6326:  "Datum_WGS84",
	6022:  "DatumE_International1924",
	6267:  "Datum_North_American_Datum_1927",
	6008:  "DatumE_Clarke1866"}

var GeogPrimeMeridianGeoKey map[int]string = map[int]string{
	8901: "PM_Greenwich",
	8902: "PM_Lisbon"}

var GeogLinearUnitsGeoKey map[int]string = map[int]string{
	9013: "Linear_Yard_Indian",
	9015: "Linear_Mile_International_Nautical",
	9007: "Linear_Link",
	9003: "Linear_Foot_US_Survey",
	9008: "Linear_Link_Benoit",
	9006: "Linear_Foot_Indian",
	9010: "Linear_Chain_Benoit",
	9002: "Linear_Foot",
	9014: "Linear_Fathom",
	9012: "Linear_Yard_Sears",
	9011: "Linear_Chain_Sears",
	9005: "Linear_Foot_Clarke",
	9004: "Linear_Foot_Modified_American",
	9001: "Linear_Meter",
	9009: "Linear_Link_Sears"}

var GeogAngularUnitsGeoKey map[int]string = map[int]string{
	9101: "Angular_Radian",
	9102: "Angular_Degree",
	9103: "Angular_Arc_Minute",
	9104: "Angular_Arc_Second",
	9105: "Angular_Grad",
	9106: "Angular_Gon",
	9107: "Angular_DMS",
	9108: "Angular_DMS_Hemisphere"}

var GeogEllipsoidGeoKey map[int]string = map[int]string{
	7003: "Ellipse_Australian_National_Spheroid",
	7018: "Ellipse_Everest1830Modified",
	7034: "Ellipse_Clarke_1880",
	7001: "Ellipse_Airy_1830",
	7011: "Ellipse_Clarke1880_IGN",
	7017: "Ellipse_Everest1830_1975Definition",
	7002: "Ellipse_Airy_Modified_1849",
	7013: "Ellipse_Clarke1880_Arc",
	7019: "Ellipse_GRS_1980",
	7004: "Ellipse_Bessel_1841",
	7012: "Ellipse_Clarke1880_RGS",
	7016: "Ellipse_Everest1830_1967Definition",
	7023: "Ellipse_International1967",
	7030: "Ellipse_WGS_84",
	7009: "Ellipse_Clarke_1866_Michigan",
	7020: "Ellipse_Helmert1906",
	7010: "Ellipse_Clarke1880_Benoit",
	7005: "Ellipse_Bessel_Modified",
	7014: "Ellipse_Clarke1880_SGA1922",
	7022: "Ellipse_International1924",
	7008: "Ellipse_Clarke_1866",
	7024: "Ellipse_Krassowsky1940",
	7015: "Ellipse_Everest1830_1937Adjustment",
	7006: "Ellipse_Bessel_Namibia"}

var GeogAzimuthUnitsGeoKey map[int]string = map[int]string{
	9101: "Angular_Radian",
	9102: "Angular_Degree",
	9103: "Angular_Arc_Minute",
	9104: "Angular_Arc_Second",
	9105: "Angular_Grad",
	9106: "Angular_Gon",
	9107: "Angular_DMS",
	9108: "Angular_DMS_Hemisphere"}

var ProjCoordTransGeoKey map[int]string = map[int]string{
	7:  "CT_Mercator",
	20: "CT_MillerCylindrical",
	27: "CT_TransvMercator_SouthOriented",
	1:  "CT_TransverseMercator",
	24: "CT_Sinusoidal",
	4:  "CT_ObliqueMercator_Laborde",
	15: "CT_PolarStereographic",
	21: "CT_Orthographic",
	13: "CT_EquidistantConic",
	22: "CT_Polyconic",
	6:  "CT_ObliqueMercator_Spherical",
	25: "CT_VanDerGrinten",
	17: "CT_Equirectangular",
	3:  "CT_ObliqueMercator",
	12: "CT_AzimuthalEquidistant",
	2:  "CT_TransvMercator_Modified_Alaska",
	23: "CT_Robinson",
	19: "CT_Gnomonic",
	11: "CT_AlbersEqualArea",
	9:  "CT_LambertConfConic_Helmert",
	5:  "CT_ObliqueMercator_Rosenmund",
	14: "CT_Stereographic",
	26: "CT_NewZealandMapGrid",
	16: "CT_ObliqueStereographic",
	10: "CT_LambertAzimEqualArea",
	18: "CT_CassiniSoldner",
	8:  "CT_LambertConfConic_2SP"}

var ProjectionGeoKey map[int]string = map[int]string{
	14803: "Proj_Wisconsin_CS27_South",
	15133: "Proj_Hawaii_CS83_3",
	12701: "Proj_Nevada_CS27_East",
	14832: "Proj_Wisconsin_CS83_Central",
	12401: "Proj_Missouri_CS27_East",
	14831: "Proj_Wisconsin_CS83_North",
	14931: "Proj_Wyoming_CS83_East",
	11132: "Proj_Idaho_CS83_Central",
	14701: "Proj_West_Virginia_CS27_North",
	14904: "Proj_Wyoming_CS27_West",
	11531: "Proj_Kansas_CS83_North",
	12630: "Proj_Nebraska_CS83",
	13701: "Proj_Pennsylvania_CS27_North",
	12403: "Proj_Missouri_CS27_West",
	12233: "Proj_Minnesota_CS83_South",
	12332: "Proj_Mississippi_CS83_West",
	11131: "Proj_Idaho_CS83_East",
	10132: "Proj_Alabama_CS83_West",
	13501: "Proj_Oklahoma_CS27_North",
	15039: "Proj_Alaska_CS83_9",
	13101: "Proj_New_York_CS27_East",
	11103: "Proj_Idaho_CS27_West",
	10402: "Proj_California_CS27_II",
	19912: "Proj_RSO_Borneo",
	12501: "Proj_Montana_CS27_North",
	13502: "Proj_Oklahoma_CS27_South",
	11601: "Proj_Kentucky_CS27_North",
	14631: "Proj_Washington_CS83_North",
	12101: "Proj_Michigan_State_Plane_East",
	10630: "Proj_Connecticut_CS83",
	11431: "Proj_Iowa_CS83_North",
	14333: "Proj_Utah_CS83_South",
	10901: "Proj_Florida_CS27_East",
	11232: "Proj_Illinois_CS83_West",
	17454: "Proj_Australian_Map_Grid_54",
	18052: "Proj_Colombia_Bogota",
	11900: "Proj_Maryland_CS27",
	10502: "Proj_Colorado_CS27_Central",
	14332: "Proj_Utah_CS83_Central",
	18054: "Proj_Colombia_6E",
	17348: "Proj_Map_Grid_of_Australia_48",
	14302: "Proj_Utah_CS27_Central",
	15005: "Proj_Alaska_CS27_5",
	17450: "Proj_Australian_Map_Grid_50",
	10403: "Proj_California_CS27_III",
	13632: "Proj_Oregon_CS83_South",
	12111: "Proj_Michigan_CS27_North",
	15104: "Proj_Hawaii_CS27_4",
	17351: "Proj_Map_Grid_of_Australia_51",
	13200: "Proj_North_Carolina_CS27",
	17449: "Proj_Australian_Map_Grid_49",
	11101: "Proj_Idaho_CS27_East",
	10432: "Proj_California_CS83_2",
	11802: "Proj_Maine_CS27_West",
	14231: "Proj_Texas_CS83_North",
	13032: "Proj_New_Mexico_CS83_Central",
	10932: "Proj_Florida_CS83_West",
	18035: "Proj_Argentina_5",
	14203: "Proj_Texas_CS27_Central",
	13102: "Proj_New_York_CS27_Central",
	10232: "Proj_Arizona_CS83_Central",
	17456: "Proj_Australian_Map_Grid_56",
	14204: "Proj_Texas_CS27_South_Central",
	11832: "Proj_Maine_CS83_West",
	17349: "Proj_Map_Grid_of_Australia_49",
	15034: "Proj_Alaska_CS83_4",
	10434: "Proj_California_CS83_4",
	17350: "Proj_Map_Grid_of_Australia_50",
	14233: "Proj_Texas_CS83_Central",
	12733: "Proj_Nevada_CS83_West",
	14934: "Proj_Wyoming_CS83_West",
	13001: "Proj_New_Mexico_CS27_East",
	14602: "Proj_Washington_CS27_South",
	17354: "Proj_Map_Grid_of_Australia_54",
	14301: "Proj_Utah_CS27_North",
	11702: "Proj_Louisiana_CS27_South",
	15037: "Proj_Alaska_CS83_7",
	12032: "Proj_Massachusetts_CS83_Island",
	12142: "Proj_Michigan_CS83_Central",
	10302: "Proj_Arkansas_CS27_South",
	13230: "Proj_North_Carolina_CS83",
	12113: "Proj_Michigan_CS27_South",
	10406: "Proj_California_CS27_VI",
	11401: "Proj_Iowa_CS27_North",
	10233: "Proj_Arizona_CS83_west",
	14903: "Proj_Wyoming_CS27_West_Central",
	12830: "Proj_New_Hampshire_CS83",
	18141: "Proj_New_Zealand_North_Island_Nat_Grid",
	12702: "Proj_Nevada_CS27_Central",
	13302: "Proj_North_Dakota_CS27_South",
	13602: "Proj_Oregon_CS27_South",
	13431: "Proj_Ohio_CS83_North",
	15914: "Proj_BLM_14N_feet",
	10203: "Proj_Arizona_Coordinate_System_west",
	12201: "Proj_Minnesota_CS27_North",
	10533: "Proj_Colorado_CS83_South",
	13800: "Proj_Rhode_Island_CS27",
	14731: "Proj_West_Virginia_CS83_North",
	14400: "Proj_Vermont_CS27",
	12301: "Proj_Mississippi_CS27_East",
	15008: "Proj_Alaska_CS27_8",
	12002: "Proj_Massachusetts_CS27_Island",
	10933: "Proj_Florida_CS83_North",
	11032: "Proj_Georgia_CS83_West",
	14002: "Proj_South_Dakota_CS27_South",
	13104: "Proj_New_York_CS27_Long_Island",
	17357: "Proj_Map_Grid_of_Australia_57",
	18031: "Proj_Argentina_1",
	18036: "Proj_Argentina_6",
	17355: "Proj_Map_Grid_of_Australia_55",
	14902: "Proj_Wyoming_CS27_East_Central",
	15132: "Proj_Hawaii_CS83_2",
	14732: "Proj_West_Virginia_CS83_South",
	12112: "Proj_Michigan_CS27_Central",
	14601: "Proj_Washington_CS27_North",
	15004: "Proj_Alaska_CS27_4",
	17455: "Proj_Australian_Map_Grid_55",
	11231: "Proj_Illinois_CS83_East",
	14205: "Proj_Texas_CS27_South",
	10102: "Proj_Alabama_CS27_West",
	10401: "Proj_California_CS27_I",
	10101: "Proj_Alabama_CS27_East",
	17457: "Proj_Australian_Map_Grid_57",
	15917: "Proj_BLM_17N_feet",
	10501: "Proj_Colorado_CS27_North",
	14501: "Proj_Virginia_CS27_North",
	14100: "Proj_Tennessee_CS27",
	12103: "Proj_Michigan_State_Plane_West",
	12402: "Proj_Missouri_CS27_Central",
	14303: "Proj_Utah_CS27_South",
	13830: "Proj_Rhode_Island_CS83",
	14702: "Proj_West_Virginia_CS27_South",
	18037: "Proj_Argentina_7",
	15036: "Proj_Alaska_CS83_6",
	10700: "Proj_Delaware_CS27",
	12732: "Proj_Nevada_CS83_Central",
	17448: "Proj_Australian_Map_Grid_48",
	15002: "Proj_Alaska_CS27_2",
	14235: "Proj_Texas_CS83_South",
	11631: "Proj_Kentucky_CS83_North",
	15103: "Proj_Hawaii_CS27_3",
	10231: "Proj_Arizona_CS83_east",
	11731: "Proj_Louisiana_CS83_North",
	13401: "Proj_Ohio_CS27_North",
	13133: "Proj_New_York_CS83_West",
	13132: "Proj_New_York_CS83_Central",
	14933: "Proj_Wyoming_CS83_West_Central",
	10503: "Proj_Colorado_CS27_South",
	13131: "Proj_New_York_CS83_East",
	10404: "Proj_California_CS27_IV",
	10331: "Proj_Arkansas_CS83_North",
	11632: "Proj_Kentucky_CS83_South",
	13702: "Proj_Pennsylvania_CS27_South",
	14833: "Proj_Wisconsin_CS83_South",
	15038: "Proj_Alaska_CS83_8",
	12203: "Proj_Minnesota_CS27_South",
	15010: "Proj_Alaska_CS27_10",
	11102: "Proj_Idaho_CS27_Central",
	10730: "Proj_Delaware_CS83",
	13930: "Proj_South_Carolina_CS83",
	18072: "Proj_Egypt_Red_Belt",
	12232: "Proj_Minnesota_CS83_Central",
	14201: "Proj_Texas_CS27_North",
	17458: "Proj_Australian_Map_Grid_58",
	12800: "Proj_New_Hampshire_CS27",
	13902: "Proj_South_Carolina_CS27_South",
	12731: "Proj_Nevada_CS83_East",
	12001: "Proj_Massachusetts_CS27_Mainland",
	15102: "Proj_Hawaii_CS27_2",
	13402: "Proj_Ohio_CS27_South",
	14502: "Proj_Virginia_CS27_South",
	14802: "Proj_Wisconsin_CS27_Central",
	13003: "Proj_New_Mexico_CS27_West",
	18053: "Proj_Colombia_3E",
	13103: "Proj_New_York_CS27_West",
	12703: "Proj_Nevada_CS27_West",
	14031: "Proj_South_Dakota_CS83_North",
	13033: "Proj_New_Mexico_CS83_West",
	12231: "Proj_Minnesota_CS83_North",
	15033: "Proj_Alaska_CS83_3",
	12141: "Proj_Michigan_CS83_North",
	13732: "Proj_Pennsylvania_CS83_South",
	14531: "Proj_Virginia_CS83_North",
	11202: "Proj_Illinois_CS27_West",
	13901: "Proj_South_Carolina_CS27_North",
	14234: "Proj_Texas_CS83_South_Central",
	14430: "Proj_Vermont_CS83",
	13432: "Proj_Ohio_CS83_South",
	12202: "Proj_Minnesota_CS27_Central",
	10131: "Proj_Alabama_CS83_East",
	15201: "Proj_Puerto_Rico_CS27",
	11332: "Proj_Indiana_CS83_West",
	11133: "Proj_Idaho_CS83_West",
	10405: "Proj_California_CS27_V",
	14801: "Proj_Wisconsin_CS27_North",
	12601: "Proj_Nebraska_CS27_North",
	10903: "Proj_Florida_CS27_North",
	32767: "user_defined",
	13031: "Proj_New_Mexico_CS83_East",
	12930: "Proj_New_Jersey_CS83",
	15230: "Proj_Puerto_Rico_Virgin_Is",
	17358: "Proj_Map_Grid_of_Australia_58",
	10902: "Proj_Florida_CS27_West",
	13332: "Proj_North_Dakota_CS83_South",
	13301: "Proj_North_Dakota_CS27_North",
	11831: "Proj_Maine_CS83_East",
	15006: "Proj_Alaska_CS27_6",
	10531: "Proj_Colorado_CS83_North",
	15032: "Proj_Alaska_CS83_2",
	18033: "Proj_Argentina_3",
	10301: "Proj_Arkansas_CS27_North",
	11432: "Proj_Iowa_CS83_South",
	15009: "Proj_Alaska_CS27_9",
	15131: "Proj_Hawaii_CS83_1",
	12900: "Proj_New_Jersey_CS27",
	18142: "Proj_New_Zealand_South_Island_Nat_Grid",
	11501: "Proj_Kansas_CS27_North",
	17452: "Proj_Australian_Map_Grid_52",
	13631: "Proj_Oregon_CS83_North",
	10931: "Proj_Florida_CS83_East",
	11701: "Proj_Louisiana_CS27_North",
	17356: "Proj_Map_Grid_of_Australia_56",
	12433: "Proj_Missouri_CS83_West",
	10436: "Proj_California_CS83_6",
	15007: "Proj_Alaska_CS27_7",
	14331: "Proj_Utah_CS83_North",
	18032: "Proj_Argentina_2",
	14532: "Proj_Virginia_CS83_South",
	15101: "Proj_Hawaii_CS27_1",
	13331: "Proj_North_Dakota_CS83_North",
	10201: "Proj_Arizona_Coordinate_System_east",
	12602: "Proj_Nebraska_CS27_South",
	15134: "Proj_Hawaii_CS83_4",
	14202: "Proj_Texas_CS27_North_Central",
	15202: "Proj_St_Croix",
	18074: "Proj_Extended_Purple_Belt",
	13531: "Proj_Oklahoma_CS83_North",
	15915: "Proj_BLM_15N_feet",
	12031: "Proj_Massachusetts_CS83_Mainland",
	15135: "Proj_Hawaii_CS83_5",
	14632: "Proj_Washington_CS83_South",
	17453: "Proj_Australian_Map_Grid_53",
	12530: "Proj_Montana_CS83",
	11302: "Proj_Indiana_CS27_West",
	13532: "Proj_Oklahoma_CS83_South",
	15001: "Proj_Alaska_CS27_1",
	13601: "Proj_Oregon_CS27_North",
	15916: "Proj_BLM_16N_feet",
	18073: "Proj_Egypt_Purple_Belt",
	11002: "Proj_Georgia_CS27_West",
	18034: "Proj_Argentina_4",
	13002: "Proj_New_Mexico_CS27_Central",
	12143: "Proj_Michigan_CS83_South",
	11602: "Proj_Kentucky_CS27_South",
	11732: "Proj_Louisiana_CS83_South",
	10600: "Proj_Connecticut_CS27",
	13134: "Proj_New_York_CS83_Long_Island",
	15003: "Proj_Alaska_CS27_3",
	11402: "Proj_Iowa_CS27_South",
	12302: "Proj_Mississippi_CS27_West",
	14232: "Proj_Texas_CS83_North_Central",
	18051: "Proj_Colombia_3W",
	17353: "Proj_Map_Grid_of_Australia_53",
	10332: "Proj_Arkansas_CS83_South",
	11532: "Proj_Kansas_CS83_South",
	12432: "Proj_Missouri_CS83_Central",
	17352: "Proj_Map_Grid_of_Australia_52",
	14901: "Proj_Wyoming_CS27_East",
	11801: "Proj_Maine_CS27_East",
	17451: "Proj_Australian_Map_Grid_51",
	10407: "Proj_California_CS27_VII",
	12502: "Proj_Montana_CS27_Central",
	12503: "Proj_Montana_CS27_South",
	19900: "Proj_Bahrain_Grid",
	12331: "Proj_Mississippi_CS83_East",
	11301: "Proj_Indiana_CS27_East",
	15031: "Proj_Alaska_CS83_1",
	10532: "Proj_Colorado_CS83_Central",
	14001: "Proj_South_Dakota_CS27_North",
	11502: "Proj_Kansas_CS27_South",
	10431: "Proj_California_CS83_1",
	10202: "Proj_Arizona_Coordinate_System_Central",
	11331: "Proj_Indiana_CS83_East",
	10435: "Proj_California_CS83_5",
	14032: "Proj_South_Dakota_CS83_South",
	11001: "Proj_Georgia_CS27_East",
	19905: "Proj_Netherlands_E_Indies_Equatorial",
	12431: "Proj_Missouri_CS83_East",
	10433: "Proj_California_CS83_3",
	15035: "Proj_Alaska_CS83_5",
	13731: "Proj_Pennsylvania_CS83_North",
	12102: "Proj_Michigan_State_Plane_Old_Central",
	14932: "Proj_Wyoming_CS83_East_Central",
	15040: "Proj_Alaska_CS83_10",
	15105: "Proj_Hawaii_CS27_5",
	11201: "Proj_Illinois_CS27_East",
	14130: "Proj_Tennessee_CS83",
	11031: "Proj_Georgia_CS83_East",
	11930: "Proj_Maryland_CS83"}

var VerticalCSTypeGeoKey map[int]string = map[int]string{
	5004: "VertCS_Bessel_1841_ellipsoid",
	5007: "VertCS_Clarke_1858_ellipsoid",
	5019: "VertCS_GRS_1980_ellipsoid",
	5712: "AHD (Tasmania) height (Reserved EPSG)",
	5701: "ODN height (Reserved EPSG)",
	5024: "VertCS_Krassowsky_1940_ellipsoid",
	5703: "NAVD88 height (Reserved EPSG)",
	5104: "VertCS_Yellow_Sea_1956",
	5020: "VertCS_Helmert_1906_ellipsoid",
	5101: "VertCS_Newlyn",
	5012: "VertCS_Clarke_1880_RGS_ellipsoid",
	5022: "VertCS_International_1924_ellipsoid",
	5711: "AHD height (Reserved EPSG)",
	5014: "VertCS_Clarke_1880_SGA_1922_ellipsoid",
	5003: "VertCS_ANS_ellipsoid",
	5027: "VertCS_Plessis_1817_ellipsoid",
	5021: "VertCS_INS_ellipsoid",
	5017: "VertCS_Everest_1830_1975_Definition_ellipsoid",
	5103: "VertCS_North_American_Vertical_Datum_1988",
	5105: "VertCS_Baltic_Sea",
	5702: "NGVD29 height (Reserved EPSG)",
	5705: "Baltic height (Reserved EPSG)",
	5018: "VertCS_Everest_1830_Modified_ellipsoid",
	5029: "VertCS_War_Office_ellipsoid",
	5030: "VertCS_WGS_84_ellipsoid",
	5031: "VertCS_GEM_10C_ellipsoid",
	5032: "VertCS_OSU86F_ellipsoid",
	5028: "VertCS_Struve_1860_ellipsoid",
	5001: "VertCS_Airy_1830_ellipsoid",
	5033: "VertCS_OSU91A_ellipsoid",
	5011: "VertCS_Clarke_1880_IGN_ellipsoid",
	5015: "VertCS_Everest_1830_1937_Adjustment_ellipsoid",
	5707: "NAP height (Reserved EPSG)",
	5005: "VertCS_Bessel_Modified_ellipsoid",
	5013: "VertCS_Clarke_1880_Arc_ellipsoid",
	5710: "Oostende height (Reserved EPSG)",
	5010: "VertCS_Clarke_1880_Benoit_ellipsoid",
	5016: "VertCS_Everest_1830_1967_Definition_ellipsoid",
	5106: "VertCS_Caspian_Sea",
	5025: "VertCS_NWL_9D_ellipsoid",
	5002: "VertCS_Airy_Modified_1849_ellipsoid",
	5102: "VertCS_North_American_Vertical_Datum_1929",
	5008: "VertCS_Clarke_1866_ellipsoid",
	5023: "VertCS_International_1967_ellipsoid",
	5026: "VertCS_NWL_10D_ellipsoid",
	5706: "Caspian depth (Reserved EPSG)",
	5006: "VertCS_Bessel_Namibia_ellipsoid",
	5704: "Yellow Sea (Reserved EPSG)"}

const (
	PCS_Adindan_UTM_zone_37N              = 20137
	PCS_Adindan_UTM_zone_38N              = 20138
	PCS_AGD66_AMG_zone_48                 = 20248
	PCS_AGD66_AMG_zone_49                 = 20249
	PCS_AGD66_AMG_zone_50                 = 20250
	PCS_AGD66_AMG_zone_51                 = 20251
	PCS_AGD66_AMG_zone_52                 = 20252
	PCS_AGD66_AMG_zone_53                 = 20253
	PCS_AGD66_AMG_zone_54                 = 20254
	PCS_AGD66_AMG_zone_55                 = 20255
	PCS_AGD66_AMG_zone_56                 = 20256
	PCS_AGD66_AMG_zone_57                 = 20257
	PCS_AGD66_AMG_zone_58                 = 20258
	PCS_AGD84_AMG_zone_48                 = 20348
	PCS_AGD84_AMG_zone_49                 = 20349
	PCS_AGD84_AMG_zone_50                 = 20350
	PCS_AGD84_AMG_zone_51                 = 20351
	PCS_AGD84_AMG_zone_52                 = 20352
	PCS_AGD84_AMG_zone_53                 = 20353
	PCS_AGD84_AMG_zone_54                 = 20354
	PCS_AGD84_AMG_zone_55                 = 20355
	PCS_AGD84_AMG_zone_56                 = 20356
	PCS_AGD84_AMG_zone_57                 = 20357
	PCS_AGD84_AMG_zone_58                 = 20358
	PCS_Ain_el_Abd_UTM_zone_37N           = 20437
	PCS_Ain_el_Abd_UTM_zone_38N           = 20438
	PCS_Ain_el_Abd_UTM_zone_39N           = 20439
	PCS_Ain_el_Abd_Bahrain_Grid           = 20499
	PCS_Afgooye_UTM_zone_38N              = 20538
	PCS_Afgooye_UTM_zone_39N              = 20539
	PCS_Lisbon_Portugese_Grid             = 20700
	PCS_Aratu_UTM_zone_22S                = 20822
	PCS_Aratu_UTM_zone_23S                = 20823
	PCS_Aratu_UTM_zone_24S                = 20824
	PCS_Arc_1950_Lo13                     = 20973
	PCS_Arc_1950_Lo15                     = 20975
	PCS_Arc_1950_Lo17                     = 20977
	PCS_Arc_1950_Lo19                     = 20979
	PCS_Arc_1950_Lo21                     = 20981
	PCS_Arc_1950_Lo23                     = 20983
	PCS_Arc_1950_Lo25                     = 20985
	PCS_Arc_1950_Lo27                     = 20987
	PCS_Arc_1950_Lo29                     = 20989
	PCS_Arc_1950_Lo31                     = 20991
	PCS_Arc_1950_Lo33                     = 20993
	PCS_Arc_1950_Lo35                     = 20995
	PCS_Batavia_NEIEZ                     = 21100
	PCS_Batavia_UTM_zone_48S              = 21148
	PCS_Batavia_UTM_zone_49S              = 21149
	PCS_Batavia_UTM_zone_50S              = 21150
	PCS_Beijing_Gauss_zone_13             = 21413
	PCS_Beijing_Gauss_zone_14             = 21414
	PCS_Beijing_Gauss_zone_15             = 21415
	PCS_Beijing_Gauss_zone_16             = 21416
	PCS_Beijing_Gauss_zone_17             = 21417
	PCS_Beijing_Gauss_zone_18             = 21418
	PCS_Beijing_Gauss_zone_19             = 21419
	PCS_Beijing_Gauss_zone_20             = 21420
	PCS_Beijing_Gauss_zone_21             = 21421
	PCS_Beijing_Gauss_zone_22             = 21422
	PCS_Beijing_Gauss_zone_23             = 21423
	PCS_Beijing_Gauss_13N                 = 21473
	PCS_Beijing_Gauss_14N                 = 21474
	PCS_Beijing_Gauss_15N                 = 21475
	PCS_Beijing_Gauss_16N                 = 21476
	PCS_Beijing_Gauss_17N                 = 21477
	PCS_Beijing_Gauss_18N                 = 21478
	PCS_Beijing_Gauss_19N                 = 21479
	PCS_Beijing_Gauss_20N                 = 21480
	PCS_Beijing_Gauss_21N                 = 21481
	PCS_Beijing_Gauss_22N                 = 21482
	PCS_Beijing_Gauss_23N                 = 21483
	PCS_Belge_Lambert_50                  = 21500
	PCS_Bern_1898_Swiss_Old               = 21790
	PCS_Bogota_UTM_zone_17N               = 21817
	PCS_Bogota_UTM_zone_18N               = 21818
	PCS_Bogota_Colombia_3W                = 21891
	PCS_Bogota_Colombia_Bogota            = 21892
	PCS_Bogota_Colombia_3E                = 21893
	PCS_Bogota_Colombia_6E                = 21894
	PCS_Camacupa_UTM_32S                  = 22032
	PCS_Camacupa_UTM_33S                  = 22033
	PCS_C_Inchauspe_Argentina_1           = 22191
	PCS_C_Inchauspe_Argentina_2           = 22192
	PCS_C_Inchauspe_Argentina_3           = 22193
	PCS_C_Inchauspe_Argentina_4           = 22194
	PCS_C_Inchauspe_Argentina_5           = 22195
	PCS_C_Inchauspe_Argentina_6           = 22196
	PCS_C_Inchauspe_Argentina_7           = 22197
	PCS_Carthage_UTM_zone_32N             = 22332
	PCS_Carthage_Nord_Tunisie             = 22391
	PCS_Carthage_Sud_Tunisie              = 22392
	PCS_Corrego_Alegre_UTM_23S            = 22523
	PCS_Corrego_Alegre_UTM_24S            = 22524
	PCS_Douala_UTM_zone_32N               = 22832
	PCS_Egypt_1907_Red_Belt               = 22992
	PCS_Egypt_1907_Purple_Belt            = 22993
	PCS_Egypt_1907_Ext_Purple             = 22994
	PCS_ED50_UTM_zone_28N                 = 23028
	PCS_ED50_UTM_zone_29N                 = 23029
	PCS_ED50_UTM_zone_30N                 = 23030
	PCS_ED50_UTM_zone_31N                 = 23031
	PCS_ED50_UTM_zone_32N                 = 23032
	PCS_ED50_UTM_zone_33N                 = 23033
	PCS_ED50_UTM_zone_34N                 = 23034
	PCS_ED50_UTM_zone_35N                 = 23035
	PCS_ED50_UTM_zone_36N                 = 23036
	PCS_ED50_UTM_zone_37N                 = 23037
	PCS_ED50_UTM_zone_38N                 = 23038
	PCS_Fahud_UTM_zone_39N                = 23239
	PCS_Fahud_UTM_zone_40N                = 23240
	PCS_Garoua_UTM_zone_33N               = 23433
	PCS_ID74_UTM_zone_46N                 = 23846
	PCS_ID74_UTM_zone_47N                 = 23847
	PCS_ID74_UTM_zone_48N                 = 23848
	PCS_ID74_UTM_zone_49N                 = 23849
	PCS_ID74_UTM_zone_50N                 = 23850
	PCS_ID74_UTM_zone_51N                 = 23851
	PCS_ID74_UTM_zone_52N                 = 23852
	PCS_ID74_UTM_zone_53N                 = 23853
	PCS_ID74_UTM_zone_46S                 = 23886
	PCS_ID74_UTM_zone_47S                 = 23887
	PCS_ID74_UTM_zone_48S                 = 23888
	PCS_ID74_UTM_zone_49S                 = 23889
	PCS_ID74_UTM_zone_50S                 = 23890
	PCS_ID74_UTM_zone_51S                 = 23891
	PCS_ID74_UTM_zone_52S                 = 23892
	PCS_ID74_UTM_zone_53S                 = 23893
	PCS_ID74_UTM_zone_54S                 = 23894
	PCS_Indian_1954_UTM_47N               = 23947
	PCS_Indian_1954_UTM_48N               = 23948
	PCS_Indian_1975_UTM_47N               = 24047
	PCS_Indian_1975_UTM_48N               = 24048
	PCS_Jamaica_1875_Old_Grid             = 24100
	PCS_JAD69_Jamaica_Grid                = 24200
	PCS_Kalianpur_India_0                 = 24370
	PCS_Kalianpur_India_I                 = 24371
	PCS_Kalianpur_India_IIa               = 24372
	PCS_Kalianpur_India_IIIa              = 24373
	PCS_Kalianpur_India_IVa               = 24374
	PCS_Kalianpur_India_IIb               = 24382
	PCS_Kalianpur_India_IIIb              = 24383
	PCS_Kalianpur_India_IVb               = 24384
	PCS_Kertau_Singapore_Grid             = 24500
	PCS_Kertau_UTM_zone_47N               = 24547
	PCS_Kertau_UTM_zone_48N               = 24548
	PCS_La_Canoa_UTM_zone_20N             = 24720
	PCS_La_Canoa_UTM_zone_21N             = 24721
	PCS_PSAD56_UTM_zone_18N               = 24818
	PCS_PSAD56_UTM_zone_19N               = 24819
	PCS_PSAD56_UTM_zone_20N               = 24820
	PCS_PSAD56_UTM_zone_21N               = 24821
	PCS_PSAD56_UTM_zone_17S               = 24877
	PCS_PSAD56_UTM_zone_18S               = 24878
	PCS_PSAD56_UTM_zone_19S               = 24879
	PCS_PSAD56_UTM_zone_20S               = 24880
	PCS_PSAD56_Peru_west_zone             = 24891
	PCS_PSAD56_Peru_central               = 24892
	PCS_PSAD56_Peru_east_zone             = 24893
	PCS_Leigon_Ghana_Grid                 = 25000
	PCS_Lome_UTM_zone_31N                 = 25231
	PCS_Luzon_Philippines_I               = 25391
	PCS_Luzon_Philippines_II              = 25392
	PCS_Luzon_Philippines_III             = 25393
	PCS_Luzon_Philippines_IV              = 25394
	PCS_Luzon_Philippines_V               = 25395
	PCS_Makassar_NEIEZ                    = 25700
	PCS_Malongo_1987_UTM_32S              = 25932
	PCS_Merchich_Nord_Maroc               = 26191
	PCS_Merchich_Sud_Maroc                = 26192
	PCS_Merchich_Sahara                   = 26193
	PCS_Massawa_UTM_zone_37N              = 26237
	PCS_Minna_UTM_zone_31N                = 26331
	PCS_Minna_UTM_zone_32N                = 26332
	PCS_Minna_Nigeria_West                = 26391
	PCS_Minna_Nigeria_Mid_Belt            = 26392
	PCS_Minna_Nigeria_East                = 26393
	PCS_Mhast_UTM_zone_32S                = 26432
	PCS_Monte_Mario_Italy_1               = 26591
	PCS_Monte_Mario_Italy_2               = 26592
	PCS_M_poraloko_UTM_32N                = 26632
	PCS_M_poraloko_UTM_32S                = 26692
	PCS_NAD27_UTM_zone_3N                 = 26703
	PCS_NAD27_UTM_zone_4N                 = 26704
	PCS_NAD27_UTM_zone_5N                 = 26705
	PCS_NAD27_UTM_zone_6N                 = 26706
	PCS_NAD27_UTM_zone_7N                 = 26707
	PCS_NAD27_UTM_zone_8N                 = 26708
	PCS_NAD27_UTM_zone_9N                 = 26709
	PCS_NAD27_UTM_zone_10N                = 26710
	PCS_NAD27_UTM_zone_11N                = 26711
	PCS_NAD27_UTM_zone_12N                = 26712
	PCS_NAD27_UTM_zone_13N                = 26713
	PCS_NAD27_UTM_zone_14N                = 26714
	PCS_NAD27_UTM_zone_15N                = 26715
	PCS_NAD27_UTM_zone_16N                = 26716
	PCS_NAD27_UTM_zone_17N                = 26717
	PCS_NAD27_UTM_zone_18N                = 26718
	PCS_NAD27_UTM_zone_19N                = 26719
	PCS_NAD27_UTM_zone_20N                = 26720
	PCS_NAD27_UTM_zone_21N                = 26721
	PCS_NAD27_UTM_zone_22N                = 26722
	PCS_NAD27_Alabama_East                = 26729
	PCS_NAD27_Alabama_West                = 26730
	PCS_NAD27_Alaska_zone_1               = 26731
	PCS_NAD27_Alaska_zone_2               = 26732
	PCS_NAD27_Alaska_zone_3               = 26733
	PCS_NAD27_Alaska_zone_4               = 26734
	PCS_NAD27_Alaska_zone_5               = 26735
	PCS_NAD27_Alaska_zone_6               = 26736
	PCS_NAD27_Alaska_zone_7               = 26737
	PCS_NAD27_Alaska_zone_8               = 26738
	PCS_NAD27_Alaska_zone_9               = 26739
	PCS_NAD27_Alaska_zone_10              = 26740
	PCS_NAD27_California_I                = 26741
	PCS_NAD27_California_II               = 26742
	PCS_NAD27_California_III              = 26743
	PCS_NAD27_California_IV               = 26744
	PCS_NAD27_California_V                = 26745
	PCS_NAD27_California_VI               = 26746
	PCS_NAD27_California_VII              = 26747
	PCS_NAD27_Arizona_East                = 26748
	PCS_NAD27_Arizona_Central             = 26749
	PCS_NAD27_Arizona_West                = 26750
	PCS_NAD27_Arkansas_North              = 26751
	PCS_NAD27_Arkansas_South              = 26752
	PCS_NAD27_Colorado_North              = 26753
	PCS_NAD27_Colorado_Central            = 26754
	PCS_NAD27_Colorado_South              = 26755
	PCS_NAD27_Connecticut                 = 26756
	PCS_NAD27_Delaware                    = 26757
	PCS_NAD27_Florida_East                = 26758
	PCS_NAD27_Florida_West                = 26759
	PCS_NAD27_Florida_North               = 26760
	PCS_NAD27_Hawaii_zone_1               = 26761
	PCS_NAD27_Hawaii_zone_2               = 26762
	PCS_NAD27_Hawaii_zone_3               = 26763
	PCS_NAD27_Hawaii_zone_4               = 26764
	PCS_NAD27_Hawaii_zone_5               = 26765
	PCS_NAD27_Georgia_East                = 26766
	PCS_NAD27_Georgia_West                = 26767
	PCS_NAD27_Idaho_East                  = 26768
	PCS_NAD27_Idaho_Central               = 26769
	PCS_NAD27_Idaho_West                  = 26770
	PCS_NAD27_Illinois_East               = 26771
	PCS_NAD27_Illinois_West               = 26772
	PCS_NAD27_Indiana_East                = 26773
	PCS_NAD27_BLM_14N_feet                = 26774
	PCS_NAD27_Indiana_West                = 26774
	PCS_NAD27_BLM_15N_feet                = 26775
	PCS_NAD27_Iowa_North                  = 26775
	PCS_NAD27_BLM_16N_feet                = 26776
	PCS_NAD27_Iowa_South                  = 26776
	PCS_NAD27_BLM_17N_feet                = 26777
	PCS_NAD27_Kansas_North                = 26777
	PCS_NAD27_Kansas_South                = 26778
	PCS_NAD27_Kentucky_North              = 26779
	PCS_NAD27_Kentucky_South              = 26780
	PCS_NAD27_Louisiana_North             = 26781
	PCS_NAD27_Louisiana_South             = 26782
	PCS_NAD27_Maine_East                  = 26783
	PCS_NAD27_Maine_West                  = 26784
	PCS_NAD27_Maryland                    = 26785
	PCS_NAD27_Massachusetts               = 26786
	PCS_NAD27_Massachusetts_Is            = 26787
	PCS_NAD27_Michigan_North              = 26788
	PCS_NAD27_Michigan_Central            = 26789
	PCS_NAD27_Michigan_South              = 26790
	PCS_NAD27_Minnesota_North             = 26791
	PCS_NAD27_Minnesota_Cent              = 26792
	PCS_NAD27_Minnesota_South             = 26793
	PCS_NAD27_Mississippi_East            = 26794
	PCS_NAD27_Mississippi_West            = 26795
	PCS_NAD27_Missouri_East               = 26796
	PCS_NAD27_Missouri_Central            = 26797
	PCS_NAD27_Missouri_West               = 26798
	PCS_NAD_Michigan_Michigan_East        = 26801
	PCS_NAD_Michigan_Michigan_Old_Central = 26802
	PCS_NAD_Michigan_Michigan_West        = 26803
	PCS_NAD83_UTM_zone_3N                 = 26903
	PCS_NAD83_UTM_zone_4N                 = 26904
	PCS_NAD83_UTM_zone_5N                 = 26905
	PCS_NAD83_UTM_zone_6N                 = 26906
	PCS_NAD83_UTM_zone_7N                 = 26907
	PCS_NAD83_UTM_zone_8N                 = 26908
	PCS_NAD83_UTM_zone_9N                 = 26909
	PCS_NAD83_UTM_zone_10N                = 26910
	PCS_NAD83_UTM_zone_11N                = 26911
	PCS_NAD83_UTM_zone_12N                = 26912
	PCS_NAD83_UTM_zone_13N                = 26913
	PCS_NAD83_UTM_zone_14N                = 26914
	PCS_NAD83_UTM_zone_15N                = 26915
	PCS_NAD83_UTM_zone_16N                = 26916
	PCS_NAD83_UTM_zone_17N                = 26917
	PCS_NAD83_UTM_zone_18N                = 26918
	PCS_NAD83_UTM_zone_19N                = 26919
	PCS_NAD83_UTM_zone_20N                = 26920
	PCS_NAD83_UTM_zone_21N                = 26921
	PCS_NAD83_UTM_zone_22N                = 26922
	PCS_NAD83_UTM_zone_23N                = 26923
	PCS_NAD83_Alabama_East                = 26929
	PCS_NAD83_Alabama_West                = 26930
	PCS_NAD83_Alaska_zone_1               = 26931
	PCS_NAD83_Alaska_zone_2               = 26932
	PCS_NAD83_Alaska_zone_3               = 26933
	PCS_NAD83_Alaska_zone_4               = 26934
	PCS_NAD83_Alaska_zone_5               = 26935
	PCS_NAD83_Alaska_zone_6               = 26936
	PCS_NAD83_Alaska_zone_7               = 26937
	PCS_NAD83_Alaska_zone_8               = 26938
	PCS_NAD83_Alaska_zone_9               = 26939
	PCS_NAD83_Alaska_zone_10              = 26940
	PCS_NAD83_California_1                = 26941
	PCS_NAD83_California_2                = 26942
	PCS_NAD83_California_3                = 26943
	PCS_NAD83_California_4                = 26944
	PCS_NAD83_California_5                = 26945
	PCS_NAD83_California_6                = 26946
	PCS_NAD83_Arizona_East                = 26948
	PCS_NAD83_Arizona_Central             = 26949
	PCS_NAD83_Arizona_West                = 26950
	PCS_NAD83_Arkansas_North              = 26951
	PCS_NAD83_Arkansas_South              = 26952
	PCS_NAD83_Colorado_North              = 26953
	PCS_NAD83_Colorado_Central            = 26954
	PCS_NAD83_Colorado_South              = 26955
	PCS_NAD83_Connecticut                 = 26956
	PCS_NAD83_Delaware                    = 26957
	PCS_NAD83_Florida_East                = 26958
	PCS_NAD83_Florida_West                = 26959
	PCS_NAD83_Florida_North               = 26960
	PCS_NAD83_Hawaii_zone_1               = 26961
	PCS_NAD83_Hawaii_zone_2               = 26962
	PCS_NAD83_Hawaii_zone_3               = 26963
	PCS_NAD83_Hawaii_zone_4               = 26964
	PCS_NAD83_Hawaii_zone_5               = 26965
	PCS_NAD83_Georgia_East                = 26966
	PCS_NAD83_Georgia_West                = 26967
	PCS_NAD83_Idaho_East                  = 26968
	PCS_NAD83_Idaho_Central               = 26969
	PCS_NAD83_Idaho_West                  = 26970
	PCS_NAD83_Illinois_East               = 26971
	PCS_NAD83_Illinois_West               = 26972
	PCS_NAD83_Indiana_East                = 26973
	PCS_NAD83_Indiana_West                = 26974
	PCS_NAD83_Iowa_North                  = 26975
	PCS_NAD83_Iowa_South                  = 26976
	PCS_NAD83_Kansas_North                = 26977
	PCS_NAD83_Kansas_South                = 26978
	PCS_NAD83_Kentucky_North              = 26979
	PCS_NAD83_Kentucky_South              = 26980
	PCS_NAD83_Louisiana_North             = 26981
	PCS_NAD83_Louisiana_South             = 26982
	PCS_NAD83_Maine_East                  = 26983
	PCS_NAD83_Maine_West                  = 26984
	PCS_NAD83_Maryland                    = 26985
	PCS_NAD83_Massachusetts               = 26986
	PCS_NAD83_Massachusetts_Is            = 26987
	PCS_NAD83_Michigan_North              = 26988
	PCS_NAD83_Michigan_Central            = 26989
	PCS_NAD83_Michigan_South              = 26990
	PCS_NAD83_Minnesota_North             = 26991
	PCS_NAD83_Minnesota_Cent              = 26992
	PCS_NAD83_Minnesota_South             = 26993
	PCS_NAD83_Mississippi_East            = 26994
	PCS_NAD83_Mississippi_West            = 26995
	PCS_NAD83_Missouri_East               = 26996
	PCS_NAD83_Missouri_Central            = 26997
	PCS_NAD83_Missouri_West               = 26998
	PCS_Nahrwan_1967_UTM_38N              = 27038
	PCS_Nahrwan_1967_UTM_39N              = 27039
	PCS_Nahrwan_1967_UTM_40N              = 27040
	PCS_Naparima_UTM_20N                  = 27120
	PCS_GD49_NZ_Map_Grid                  = 27200
	PCS_GD49_North_Island_Grid            = 27291
	PCS_GD49_South_Island_Grid            = 27292
	PCS_Datum_73_UTM_zone_29N             = 27429
	PCS_ATF_Nord_de_Guerre                = 27500
	PCS_NTF_France_I                      = 27581
	PCS_NTF_France_II                     = 27582
	PCS_NTF_France_III                    = 27583
	PCS_NTF_Nord_France                   = 27591
	PCS_NTF_Centre_France                 = 27592
	PCS_NTF_Sud_France                    = 27593
	PCS_British_National_Grid             = 27700
	PCS_Point_Noire_UTM_32S               = 28232
	PCS_GDA94_MGA_zone_48                 = 28348
	PCS_GDA94_MGA_zone_49                 = 28349
	PCS_GDA94_MGA_zone_50                 = 28350
	PCS_GDA94_MGA_zone_51                 = 28351
	PCS_GDA94_MGA_zone_52                 = 28352
	PCS_GDA94_MGA_zone_53                 = 28353
	PCS_GDA94_MGA_zone_54                 = 28354
	PCS_GDA94_MGA_zone_55                 = 28355
	PCS_GDA94_MGA_zone_56                 = 28356
	PCS_GDA94_MGA_zone_57                 = 28357
	PCS_GDA94_MGA_zone_58                 = 28358
	PCS_Pulkovo_Gauss_zone_4              = 28404
	PCS_Pulkovo_Gauss_zone_5              = 28405
	PCS_Pulkovo_Gauss_zone_6              = 28406
	PCS_Pulkovo_Gauss_zone_7              = 28407
	PCS_Pulkovo_Gauss_zone_8              = 28408
	PCS_Pulkovo_Gauss_zone_9              = 28409
	PCS_Pulkovo_Gauss_zone_10             = 28410
	PCS_Pulkovo_Gauss_zone_11             = 28411
	PCS_Pulkovo_Gauss_zone_12             = 28412
	PCS_Pulkovo_Gauss_zone_13             = 28413
	PCS_Pulkovo_Gauss_zone_14             = 28414
	PCS_Pulkovo_Gauss_zone_15             = 28415
	PCS_Pulkovo_Gauss_zone_16             = 28416
	PCS_Pulkovo_Gauss_zone_17             = 28417
	PCS_Pulkovo_Gauss_zone_18             = 28418
	PCS_Pulkovo_Gauss_zone_19             = 28419
	PCS_Pulkovo_Gauss_zone_20             = 28420
	PCS_Pulkovo_Gauss_zone_21             = 28421
	PCS_Pulkovo_Gauss_zone_22             = 28422
	PCS_Pulkovo_Gauss_zone_23             = 28423
	PCS_Pulkovo_Gauss_zone_24             = 28424
	PCS_Pulkovo_Gauss_zone_25             = 28425
	PCS_Pulkovo_Gauss_zone_26             = 28426
	PCS_Pulkovo_Gauss_zone_27             = 28427
	PCS_Pulkovo_Gauss_zone_28             = 28428
	PCS_Pulkovo_Gauss_zone_29             = 28429
	PCS_Pulkovo_Gauss_zone_30             = 28430
	PCS_Pulkovo_Gauss_zone_31             = 28431
	PCS_Pulkovo_Gauss_zone_32             = 28432
	PCS_Pulkovo_Gauss_4N                  = 28464
	PCS_Pulkovo_Gauss_5N                  = 28465
	PCS_Pulkovo_Gauss_6N                  = 28466
	PCS_Pulkovo_Gauss_7N                  = 28467
	PCS_Pulkovo_Gauss_8N                  = 28468
	PCS_Pulkovo_Gauss_9N                  = 28469
	PCS_Pulkovo_Gauss_10N                 = 28470
	PCS_Pulkovo_Gauss_11N                 = 28471
	PCS_Pulkovo_Gauss_12N                 = 28472
	PCS_Pulkovo_Gauss_13N                 = 28473
	PCS_Pulkovo_Gauss_14N                 = 28474
	PCS_Pulkovo_Gauss_15N                 = 28475
	PCS_Pulkovo_Gauss_16N                 = 28476
	PCS_Pulkovo_Gauss_17N                 = 28477
	PCS_Pulkovo_Gauss_18N                 = 28478
	PCS_Pulkovo_Gauss_19N                 = 28479
	PCS_Pulkovo_Gauss_20N                 = 28480
	PCS_Pulkovo_Gauss_21N                 = 28481
	PCS_Pulkovo_Gauss_22N                 = 28482
	PCS_Pulkovo_Gauss_23N                 = 28483
	PCS_Pulkovo_Gauss_24N                 = 28484
	PCS_Pulkovo_Gauss_25N                 = 28485
	PCS_Pulkovo_Gauss_26N                 = 28486
	PCS_Pulkovo_Gauss_27N                 = 28487
	PCS_Pulkovo_Gauss_28N                 = 28488
	PCS_Pulkovo_Gauss_29N                 = 28489
	PCS_Pulkovo_Gauss_30N                 = 28490
	PCS_Pulkovo_Gauss_31N                 = 28491
	PCS_Pulkovo_Gauss_32N                 = 28492
	PCS_Qatar_National_Grid               = 28600
	PCS_RD_Netherlands_Old                = 28991
	PCS_RD_Netherlands_New                = 28992
	PCS_SAD69_UTM_zone_18N                = 29118
	PCS_SAD69_UTM_zone_19N                = 29119
	PCS_SAD69_UTM_zone_20N                = 29120
	PCS_SAD69_UTM_zone_21N                = 29121
	PCS_SAD69_UTM_zone_22N                = 29122
	PCS_SAD69_UTM_zone_17S                = 29177
	PCS_SAD69_UTM_zone_18S                = 29178
	PCS_SAD69_UTM_zone_19S                = 29179
	PCS_SAD69_UTM_zone_20S                = 29180
	PCS_SAD69_UTM_zone_21S                = 29181
	PCS_SAD69_UTM_zone_22S                = 29182
	PCS_SAD69_UTM_zone_23S                = 29183
	PCS_SAD69_UTM_zone_24S                = 29184
	PCS_SAD69_UTM_zone_25S                = 29185
	PCS_Sapper_Hill_UTM_20S               = 29220
	PCS_Sapper_Hill_UTM_21S               = 29221
	PCS_Schwarzeck_UTM_33S                = 29333
	PCS_Sudan_UTM_zone_35N                = 29635
	PCS_Sudan_UTM_zone_36N                = 29636
	PCS_Tananarive_Laborde                = 29700
	PCS_Tananarive_UTM_38S                = 29738
	PCS_Tananarive_UTM_39S                = 29739
	PCS_Timbalai_1948_Borneo              = 29800
	PCS_Timbalai_1948_UTM_49N             = 29849
	PCS_Timbalai_1948_UTM_50N             = 29850
	PCS_TM65_Irish_Nat_Grid               = 29900
	PCS_Trinidad_1903_Trinidad            = 30200
	PCS_TC_1948_UTM_zone_39N              = 30339
	PCS_TC_1948_UTM_zone_40N              = 30340
	PCS_Voirol_N_Algerie_ancien           = 30491
	PCS_Voirol_S_Algerie_ancien           = 30492
	PCS_Voirol_Unifie_N_Algerie           = 30591
	PCS_Voirol_Unifie_S_Algerie           = 30592
	PCS_Bern_1938_Swiss_New               = 30600
	PCS_Nord_Sahara_UTM_29N               = 30729
	PCS_Nord_Sahara_UTM_30N               = 30730
	PCS_Nord_Sahara_UTM_31N               = 30731
	PCS_Nord_Sahara_UTM_32N               = 30732
	PCS_Yoff_UTM_zone_28N                 = 31028
	PCS_Zanderij_UTM_zone_21N             = 31121
	PCS_MGI_Austria_West                  = 31291
	PCS_MGI_Austria_Central               = 31292
	PCS_MGI_Austria_East                  = 31293
	PCS_Belge_Lambert_72                  = 31300
	PCS_DHDN_Germany_zone_1               = 31491
	PCS_DHDN_Germany_zone_2               = 31492
	PCS_DHDN_Germany_zone_3               = 31493
	PCS_DHDN_Germany_zone_4               = 31494
	PCS_DHDN_Germany_zone_5               = 31495
	PCS_NAD27_Montana_North               = 32001
	PCS_NAD27_Montana_Central             = 32002
	PCS_NAD27_Montana_South               = 32003
	PCS_NAD27_Nebraska_North              = 32005
	PCS_NAD27_Nebraska_South              = 32006
	PCS_NAD27_Nevada_East                 = 32007
	PCS_NAD27_Nevada_Central              = 32008
	PCS_NAD27_Nevada_West                 = 32009
	PCS_NAD27_New_Hampshire               = 32010
	PCS_NAD27_New_Jersey                  = 32011
	PCS_NAD27_New_Mexico_East             = 32012
	PCS_NAD27_New_Mexico_Cent             = 32013
	PCS_NAD27_New_Mexico_West             = 32014
	PCS_NAD27_New_York_East               = 32015
	PCS_NAD27_New_York_Central            = 32016
	PCS_NAD27_New_York_West               = 32017
	PCS_NAD27_New_York_Long_Is            = 32018
	PCS_NAD27_North_Carolina              = 32019
	PCS_NAD27_North_Dakota_N              = 32020
	PCS_NAD27_North_Dakota_S              = 32021
	PCS_NAD27_Ohio_North                  = 32022
	PCS_NAD27_Ohio_South                  = 32023
	PCS_NAD27_Oklahoma_North              = 32024
	PCS_NAD27_Oklahoma_South              = 32025
	PCS_NAD27_Oregon_North                = 32026
	PCS_NAD27_Oregon_South                = 32027
	PCS_NAD27_Pennsylvania_N              = 32028
	PCS_NAD27_Pennsylvania_S              = 32029
	PCS_NAD27_Rhode_Island                = 32030
	PCS_NAD27_South_Carolina_N            = 32031
	PCS_NAD27_South_Carolina_S            = 32033
	PCS_NAD27_South_Dakota_N              = 32034
	PCS_NAD27_South_Dakota_S              = 32035
	PCS_NAD27_Tennessee                   = 32036
	PCS_NAD27_Texas_North                 = 32037
	PCS_NAD27_Texas_North_Cen             = 32038
	PCS_NAD27_Texas_Central               = 32039
	PCS_NAD27_Texas_South_Cen             = 32040
	PCS_NAD27_Texas_South                 = 32041
	PCS_NAD27_Utah_North                  = 32042
	PCS_NAD27_Utah_Central                = 32043
	PCS_NAD27_Utah_South                  = 32044
	PCS_NAD27_Vermont                     = 32045
	PCS_NAD27_Virginia_North              = 32046
	PCS_NAD27_Virginia_South              = 32047
	PCS_NAD27_Washington_North            = 32048
	PCS_NAD27_Washington_South            = 32049
	PCS_NAD27_West_Virginia_N             = 32050
	PCS_NAD27_West_Virginia_S             = 32051
	PCS_NAD27_Wisconsin_North             = 32052
	PCS_NAD27_Wisconsin_Cen               = 32053
	PCS_NAD27_Wisconsin_South             = 32054
	PCS_NAD27_Wyoming_East                = 32055
	PCS_NAD27_Wyoming_E_Cen               = 32056
	PCS_NAD27_Wyoming_W_Cen               = 32057
	PCS_NAD27_Wyoming_West                = 32058
	PCS_NAD27_Puerto_Rico                 = 32059
	PCS_NAD27_St_Croix                    = 32060
	PCS_NAD83_Montana                     = 32100
	PCS_NAD83_Nebraska                    = 32104
	PCS_NAD83_Nevada_East                 = 32107
	PCS_NAD83_Nevada_Central              = 32108
	PCS_NAD83_Nevada_West                 = 32109
	PCS_NAD83_New_Hampshire               = 32110
	PCS_NAD83_New_Jersey                  = 32111
	PCS_NAD83_New_Mexico_East             = 32112
	PCS_NAD83_New_Mexico_Cent             = 32113
	PCS_NAD83_New_Mexico_West             = 32114
	PCS_NAD83_New_York_East               = 32115
	PCS_NAD83_New_York_Central            = 32116
	PCS_NAD83_New_York_West               = 32117
	PCS_NAD83_New_York_Long_Is            = 32118
	PCS_NAD83_North_Carolina              = 32119
	PCS_NAD83_North_Dakota_N              = 32120
	PCS_NAD83_North_Dakota_S              = 32121
	PCS_NAD83_Ohio_North                  = 32122
	PCS_NAD83_Ohio_South                  = 32123
	PCS_NAD83_Oklahoma_North              = 32124
	PCS_NAD83_Oklahoma_South              = 32125
	PCS_NAD83_Oregon_North                = 32126
	PCS_NAD83_Oregon_South                = 32127
	PCS_NAD83_Pennsylvania_N              = 32128
	PCS_NAD83_Pennsylvania_S              = 32129
	PCS_NAD83_Rhode_Island                = 32130
	PCS_NAD83_South_Carolina              = 32133
	PCS_NAD83_South_Dakota_N              = 32134
	PCS_NAD83_South_Dakota_S              = 32135
	PCS_NAD83_Tennessee                   = 32136
	PCS_NAD83_Texas_North                 = 32137
	PCS_NAD83_Texas_North_Cen             = 32138
	PCS_NAD83_Texas_Central               = 32139
	PCS_NAD83_Texas_South_Cen             = 32140
	PCS_NAD83_Texas_South                 = 32141
	PCS_NAD83_Utah_North                  = 32142
	PCS_NAD83_Utah_Central                = 32143
	PCS_NAD83_Utah_South                  = 32144
	PCS_NAD83_Vermont                     = 32145
	PCS_NAD83_Virginia_North              = 32146
	PCS_NAD83_Virginia_South              = 32147
	PCS_NAD83_Washington_North            = 32148
	PCS_NAD83_Washington_South            = 32149
	PCS_NAD83_West_Virginia_N             = 32150
	PCS_NAD83_West_Virginia_S             = 32151
	PCS_NAD83_Wisconsin_North             = 32152
	PCS_NAD83_Wisconsin_Cen               = 32153
	PCS_NAD83_Wisconsin_South             = 32154
	PCS_NAD83_Wyoming_East                = 32155
	PCS_NAD83_Wyoming_E_Cen               = 32156
	PCS_NAD83_Wyoming_W_Cen               = 32157
	PCS_NAD83_Wyoming_West                = 32158
	PCS_NAD83_Puerto_Rico_Virgin_Is       = 32161
	PCS_WGS72_UTM_zone_1N                 = 32201
	PCS_WGS72_UTM_zone_2N                 = 32202
	PCS_WGS72_UTM_zone_3N                 = 32203
	PCS_WGS72_UTM_zone_4N                 = 32204
	PCS_WGS72_UTM_zone_5N                 = 32205
	PCS_WGS72_UTM_zone_6N                 = 32206
	PCS_WGS72_UTM_zone_7N                 = 32207
	PCS_WGS72_UTM_zone_8N                 = 32208
	PCS_WGS72_UTM_zone_9N                 = 32209
	PCS_WGS72_UTM_zone_10N                = 32210
	PCS_WGS72_UTM_zone_11N                = 32211
	PCS_WGS72_UTM_zone_12N                = 32212
	PCS_WGS72_UTM_zone_13N                = 32213
	PCS_WGS72_UTM_zone_14N                = 32214
	PCS_WGS72_UTM_zone_15N                = 32215
	PCS_WGS72_UTM_zone_16N                = 32216
	PCS_WGS72_UTM_zone_17N                = 32217
	PCS_WGS72_UTM_zone_18N                = 32218
	PCS_WGS72_UTM_zone_19N                = 32219
	PCS_WGS72_UTM_zone_20N                = 32220
	PCS_WGS72_UTM_zone_21N                = 32221
	PCS_WGS72_UTM_zone_22N                = 32222
	PCS_WGS72_UTM_zone_23N                = 32223
	PCS_WGS72_UTM_zone_24N                = 32224
	PCS_WGS72_UTM_zone_25N                = 32225
	PCS_WGS72_UTM_zone_26N                = 32226
	PCS_WGS72_UTM_zone_27N                = 32227
	PCS_WGS72_UTM_zone_28N                = 32228
	PCS_WGS72_UTM_zone_29N                = 32229
	PCS_WGS72_UTM_zone_30N                = 32230
	PCS_WGS72_UTM_zone_31N                = 32231
	PCS_WGS72_UTM_zone_32N                = 32232
	PCS_WGS72_UTM_zone_33N                = 32233
	PCS_WGS72_UTM_zone_34N                = 32234
	PCS_WGS72_UTM_zone_35N                = 32235
	PCS_WGS72_UTM_zone_36N                = 32236
	PCS_WGS72_UTM_zone_37N                = 32237
	PCS_WGS72_UTM_zone_38N                = 32238
	PCS_WGS72_UTM_zone_39N                = 32239
	PCS_WGS72_UTM_zone_40N                = 32240
	PCS_WGS72_UTM_zone_41N                = 32241
	PCS_WGS72_UTM_zone_42N                = 32242
	PCS_WGS72_UTM_zone_43N                = 32243
	PCS_WGS72_UTM_zone_44N                = 32244
	PCS_WGS72_UTM_zone_45N                = 32245
	PCS_WGS72_UTM_zone_46N                = 32246
	PCS_WGS72_UTM_zone_47N                = 32247
	PCS_WGS72_UTM_zone_48N                = 32248
	PCS_WGS72_UTM_zone_49N                = 32249
	PCS_WGS72_UTM_zone_50N                = 32250
	PCS_WGS72_UTM_zone_51N                = 32251
	PCS_WGS72_UTM_zone_52N                = 32252
	PCS_WGS72_UTM_zone_53N                = 32253
	PCS_WGS72_UTM_zone_54N                = 32254
	PCS_WGS72_UTM_zone_55N                = 32255
	PCS_WGS72_UTM_zone_56N                = 32256
	PCS_WGS72_UTM_zone_57N                = 32257
	PCS_WGS72_UTM_zone_58N                = 32258
	PCS_WGS72_UTM_zone_59N                = 32259
	PCS_WGS72_UTM_zone_60N                = 32260
	PCS_WGS72_UTM_zone_1S                 = 32301
	PCS_WGS72_UTM_zone_2S                 = 32302
	PCS_WGS72_UTM_zone_3S                 = 32303
	PCS_WGS72_UTM_zone_4S                 = 32304
	PCS_WGS72_UTM_zone_5S                 = 32305
	PCS_WGS72_UTM_zone_6S                 = 32306
	PCS_WGS72_UTM_zone_7S                 = 32307
	PCS_WGS72_UTM_zone_8S                 = 32308
	PCS_WGS72_UTM_zone_9S                 = 32309
	PCS_WGS72_UTM_zone_10S                = 32310
	PCS_WGS72_UTM_zone_11S                = 32311
	PCS_WGS72_UTM_zone_12S                = 32312
	PCS_WGS72_UTM_zone_13S                = 32313
	PCS_WGS72_UTM_zone_14S                = 32314
	PCS_WGS72_UTM_zone_15S                = 32315
	PCS_WGS72_UTM_zone_16S                = 32316
	PCS_WGS72_UTM_zone_17S                = 32317
	PCS_WGS72_UTM_zone_18S                = 32318
	PCS_WGS72_UTM_zone_19S                = 32319
	PCS_WGS72_UTM_zone_20S                = 32320
	PCS_WGS72_UTM_zone_21S                = 32321
	PCS_WGS72_UTM_zone_22S                = 32322
	PCS_WGS72_UTM_zone_23S                = 32323
	PCS_WGS72_UTM_zone_24S                = 32324
	PCS_WGS72_UTM_zone_25S                = 32325
	PCS_WGS72_UTM_zone_26S                = 32326
	PCS_WGS72_UTM_zone_27S                = 32327
	PCS_WGS72_UTM_zone_28S                = 32328
	PCS_WGS72_UTM_zone_29S                = 32329
	PCS_WGS72_UTM_zone_30S                = 32330
	PCS_WGS72_UTM_zone_31S                = 32331
	PCS_WGS72_UTM_zone_32S                = 32332
	PCS_WGS72_UTM_zone_33S                = 32333
	PCS_WGS72_UTM_zone_34S                = 32334
	PCS_WGS72_UTM_zone_35S                = 32335
	PCS_WGS72_UTM_zone_36S                = 32336
	PCS_WGS72_UTM_zone_37S                = 32337
	PCS_WGS72_UTM_zone_38S                = 32338
	PCS_WGS72_UTM_zone_39S                = 32339
	PCS_WGS72_UTM_zone_40S                = 32340
	PCS_WGS72_UTM_zone_41S                = 32341
	PCS_WGS72_UTM_zone_42S                = 32342
	PCS_WGS72_UTM_zone_43S                = 32343
	PCS_WGS72_UTM_zone_44S                = 32344
	PCS_WGS72_UTM_zone_45S                = 32345
	PCS_WGS72_UTM_zone_46S                = 32346
	PCS_WGS72_UTM_zone_47S                = 32347
	PCS_WGS72_UTM_zone_48S                = 32348
	PCS_WGS72_UTM_zone_49S                = 32349
	PCS_WGS72_UTM_zone_50S                = 32350
	PCS_WGS72_UTM_zone_51S                = 32351
	PCS_WGS72_UTM_zone_52S                = 32352
	PCS_WGS72_UTM_zone_53S                = 32353
	PCS_WGS72_UTM_zone_54S                = 32354
	PCS_WGS72_UTM_zone_55S                = 32355
	PCS_WGS72_UTM_zone_56S                = 32356
	PCS_WGS72_UTM_zone_57S                = 32357
	PCS_WGS72_UTM_zone_58S                = 32358
	PCS_WGS72_UTM_zone_59S                = 32359
	PCS_WGS72_UTM_zone_60S                = 32360
	PCS_WGS72BE_UTM_zone_1N               = 32401
	PCS_WGS72BE_UTM_zone_2N               = 32402
	PCS_WGS72BE_UTM_zone_3N               = 32403
	PCS_WGS72BE_UTM_zone_4N               = 32404
	PCS_WGS72BE_UTM_zone_5N               = 32405
	PCS_WGS72BE_UTM_zone_6N               = 32406
	PCS_WGS72BE_UTM_zone_7N               = 32407
	PCS_WGS72BE_UTM_zone_8N               = 32408
	PCS_WGS72BE_UTM_zone_9N               = 32409
	PCS_WGS72BE_UTM_zone_10N              = 32410
	PCS_WGS72BE_UTM_zone_11N              = 32411
	PCS_WGS72BE_UTM_zone_12N              = 32412
	PCS_WGS72BE_UTM_zone_13N              = 32413
	PCS_WGS72BE_UTM_zone_14N              = 32414
	PCS_WGS72BE_UTM_zone_15N              = 32415
	PCS_WGS72BE_UTM_zone_16N              = 32416
	PCS_WGS72BE_UTM_zone_17N              = 32417
	PCS_WGS72BE_UTM_zone_18N              = 32418
	PCS_WGS72BE_UTM_zone_19N              = 32419
	PCS_WGS72BE_UTM_zone_20N              = 32420
	PCS_WGS72BE_UTM_zone_21N              = 32421
	PCS_WGS72BE_UTM_zone_22N              = 32422
	PCS_WGS72BE_UTM_zone_23N              = 32423
	PCS_WGS72BE_UTM_zone_24N              = 32424
	PCS_WGS72BE_UTM_zone_25N              = 32425
	PCS_WGS72BE_UTM_zone_26N              = 32426
	PCS_WGS72BE_UTM_zone_27N              = 32427
	PCS_WGS72BE_UTM_zone_28N              = 32428
	PCS_WGS72BE_UTM_zone_29N              = 32429
	PCS_WGS72BE_UTM_zone_30N              = 32430
	PCS_WGS72BE_UTM_zone_31N              = 32431
	PCS_WGS72BE_UTM_zone_32N              = 32432
	PCS_WGS72BE_UTM_zone_33N              = 32433
	PCS_WGS72BE_UTM_zone_34N              = 32434
	PCS_WGS72BE_UTM_zone_35N              = 32435
	PCS_WGS72BE_UTM_zone_36N              = 32436
	PCS_WGS72BE_UTM_zone_37N              = 32437
	PCS_WGS72BE_UTM_zone_38N              = 32438
	PCS_WGS72BE_UTM_zone_39N              = 32439
	PCS_WGS72BE_UTM_zone_40N              = 32440
	PCS_WGS72BE_UTM_zone_41N              = 32441
	PCS_WGS72BE_UTM_zone_42N              = 32442
	PCS_WGS72BE_UTM_zone_43N              = 32443
	PCS_WGS72BE_UTM_zone_44N              = 32444
	PCS_WGS72BE_UTM_zone_45N              = 32445
	PCS_WGS72BE_UTM_zone_46N              = 32446
	PCS_WGS72BE_UTM_zone_47N              = 32447
	PCS_WGS72BE_UTM_zone_48N              = 32448
	PCS_WGS72BE_UTM_zone_49N              = 32449
	PCS_WGS72BE_UTM_zone_50N              = 32450
	PCS_WGS72BE_UTM_zone_51N              = 32451
	PCS_WGS72BE_UTM_zone_52N              = 32452
	PCS_WGS72BE_UTM_zone_53N              = 32453
	PCS_WGS72BE_UTM_zone_54N              = 32454
	PCS_WGS72BE_UTM_zone_55N              = 32455
	PCS_WGS72BE_UTM_zone_56N              = 32456
	PCS_WGS72BE_UTM_zone_57N              = 32457
	PCS_WGS72BE_UTM_zone_58N              = 32458
	PCS_WGS72BE_UTM_zone_59N              = 32459
	PCS_WGS72BE_UTM_zone_60N              = 32460
	PCS_WGS72BE_UTM_zone_1S               = 32501
	PCS_WGS72BE_UTM_zone_2S               = 32502
	PCS_WGS72BE_UTM_zone_3S               = 32503
	PCS_WGS72BE_UTM_zone_4S               = 32504
	PCS_WGS72BE_UTM_zone_5S               = 32505
	PCS_WGS72BE_UTM_zone_6S               = 32506
	PCS_WGS72BE_UTM_zone_7S               = 32507
	PCS_WGS72BE_UTM_zone_8S               = 32508
	PCS_WGS72BE_UTM_zone_9S               = 32509
	PCS_WGS72BE_UTM_zone_10S              = 32510
	PCS_WGS72BE_UTM_zone_11S              = 32511
	PCS_WGS72BE_UTM_zone_12S              = 32512
	PCS_WGS72BE_UTM_zone_13S              = 32513
	PCS_WGS72BE_UTM_zone_14S              = 32514
	PCS_WGS72BE_UTM_zone_15S              = 32515
	PCS_WGS72BE_UTM_zone_16S              = 32516
	PCS_WGS72BE_UTM_zone_17S              = 32517
	PCS_WGS72BE_UTM_zone_18S              = 32518
	PCS_WGS72BE_UTM_zone_19S              = 32519
	PCS_WGS72BE_UTM_zone_20S              = 32520
	PCS_WGS72BE_UTM_zone_21S              = 32521
	PCS_WGS72BE_UTM_zone_22S              = 32522
	PCS_WGS72BE_UTM_zone_23S              = 32523
	PCS_WGS72BE_UTM_zone_24S              = 32524
	PCS_WGS72BE_UTM_zone_25S              = 32525
	PCS_WGS72BE_UTM_zone_26S              = 32526
	PCS_WGS72BE_UTM_zone_27S              = 32527
	PCS_WGS72BE_UTM_zone_28S              = 32528
	PCS_WGS72BE_UTM_zone_29S              = 32529
	PCS_WGS72BE_UTM_zone_30S              = 32530
	PCS_WGS72BE_UTM_zone_31S              = 32531
	PCS_WGS72BE_UTM_zone_32S              = 32532
	PCS_WGS72BE_UTM_zone_33S              = 32533
	PCS_WGS72BE_UTM_zone_34S              = 32534
	PCS_WGS72BE_UTM_zone_35S              = 32535
	PCS_WGS72BE_UTM_zone_36S              = 32536
	PCS_WGS72BE_UTM_zone_37S              = 32537
	PCS_WGS72BE_UTM_zone_38S              = 32538
	PCS_WGS72BE_UTM_zone_39S              = 32539
	PCS_WGS72BE_UTM_zone_40S              = 32540
	PCS_WGS72BE_UTM_zone_41S              = 32541
	PCS_WGS72BE_UTM_zone_42S              = 32542
	PCS_WGS72BE_UTM_zone_43S              = 32543
	PCS_WGS72BE_UTM_zone_44S              = 32544
	PCS_WGS72BE_UTM_zone_45S              = 32545
	PCS_WGS72BE_UTM_zone_46S              = 32546
	PCS_WGS72BE_UTM_zone_47S              = 32547
	PCS_WGS72BE_UTM_zone_48S              = 32548
	PCS_WGS72BE_UTM_zone_49S              = 32549
	PCS_WGS72BE_UTM_zone_50S              = 32550
	PCS_WGS72BE_UTM_zone_51S              = 32551
	PCS_WGS72BE_UTM_zone_52S              = 32552
	PCS_WGS72BE_UTM_zone_53S              = 32553
	PCS_WGS72BE_UTM_zone_54S              = 32554
	PCS_WGS72BE_UTM_zone_55S              = 32555
	PCS_WGS72BE_UTM_zone_56S              = 32556
	PCS_WGS72BE_UTM_zone_57S              = 32557
	PCS_WGS72BE_UTM_zone_58S              = 32558
	PCS_WGS72BE_UTM_zone_59S              = 32559
	PCS_WGS72BE_UTM_zone_60S              = 32560
	PCS_WGS84_UTM_zone_1N                 = 32601
	PCS_WGS84_UTM_zone_2N                 = 32602
	PCS_WGS84_UTM_zone_3N                 = 32603
	PCS_WGS84_UTM_zone_4N                 = 32604
	PCS_WGS84_UTM_zone_5N                 = 32605
	PCS_WGS84_UTM_zone_6N                 = 32606
	PCS_WGS84_UTM_zone_7N                 = 32607
	PCS_WGS84_UTM_zone_8N                 = 32608
	PCS_WGS84_UTM_zone_9N                 = 32609
	PCS_WGS84_UTM_zone_10N                = 32610
	PCS_WGS84_UTM_zone_11N                = 32611
	PCS_WGS84_UTM_zone_12N                = 32612
	PCS_WGS84_UTM_zone_13N                = 32613
	PCS_WGS84_UTM_zone_14N                = 32614
	PCS_WGS84_UTM_zone_15N                = 32615
	PCS_WGS84_UTM_zone_16N                = 32616
	PCS_WGS84_UTM_zone_17N                = 32617
	PCS_WGS84_UTM_zone_18N                = 32618
	PCS_WGS84_UTM_zone_19N                = 32619
	PCS_WGS84_UTM_zone_20N                = 32620
	PCS_WGS84_UTM_zone_21N                = 32621
	PCS_WGS84_UTM_zone_22N                = 32622
	PCS_WGS84_UTM_zone_23N                = 32623
	PCS_WGS84_UTM_zone_24N                = 32624
	PCS_WGS84_UTM_zone_25N                = 32625
	PCS_WGS84_UTM_zone_26N                = 32626
	PCS_WGS84_UTM_zone_27N                = 32627
	PCS_WGS84_UTM_zone_28N                = 32628
	PCS_WGS84_UTM_zone_29N                = 32629
	PCS_WGS84_UTM_zone_30N                = 32630
	PCS_WGS84_UTM_zone_31N                = 32631
	PCS_WGS84_UTM_zone_32N                = 32632
	PCS_WGS84_UTM_zone_33N                = 32633
	PCS_WGS84_UTM_zone_34N                = 32634
	PCS_WGS84_UTM_zone_35N                = 32635
	PCS_WGS84_UTM_zone_36N                = 32636
	PCS_WGS84_UTM_zone_37N                = 32637
	PCS_WGS84_UTM_zone_38N                = 32638
	PCS_WGS84_UTM_zone_39N                = 32639
	PCS_WGS84_UTM_zone_40N                = 32640
	PCS_WGS84_UTM_zone_41N                = 32641
	PCS_WGS84_UTM_zone_42N                = 32642
	PCS_WGS84_UTM_zone_43N                = 32643
	PCS_WGS84_UTM_zone_44N                = 32644
	PCS_WGS84_UTM_zone_45N                = 32645
	PCS_WGS84_UTM_zone_46N                = 32646
	PCS_WGS84_UTM_zone_47N                = 32647
	PCS_WGS84_UTM_zone_48N                = 32648
	PCS_WGS84_UTM_zone_49N                = 32649
	PCS_WGS84_UTM_zone_50N                = 32650
	PCS_WGS84_UTM_zone_51N                = 32651
	PCS_WGS84_UTM_zone_52N                = 32652
	PCS_WGS84_UTM_zone_53N                = 32653
	PCS_WGS84_UTM_zone_54N                = 32654
	PCS_WGS84_UTM_zone_55N                = 32655
	PCS_WGS84_UTM_zone_56N                = 32656
	PCS_WGS84_UTM_zone_57N                = 32657
	PCS_WGS84_UTM_zone_58N                = 32658
	PCS_WGS84_UTM_zone_59N                = 32659
	PCS_WGS84_UTM_zone_60N                = 32660
	PCS_WGS84_UTM_zone_1S                 = 32701
	PCS_WGS84_UTM_zone_2S                 = 32702
	PCS_WGS84_UTM_zone_3S                 = 32703
	PCS_WGS84_UTM_zone_4S                 = 32704
	PCS_WGS84_UTM_zone_5S                 = 32705
	PCS_WGS84_UTM_zone_6S                 = 32706
	PCS_WGS84_UTM_zone_7S                 = 32707
	PCS_WGS84_UTM_zone_8S                 = 32708
	PCS_WGS84_UTM_zone_9S                 = 32709
	PCS_WGS84_UTM_zone_10S                = 32710
	PCS_WGS84_UTM_zone_11S                = 32711
	PCS_WGS84_UTM_zone_12S                = 32712
	PCS_WGS84_UTM_zone_13S                = 32713
	PCS_WGS84_UTM_zone_14S                = 32714
	PCS_WGS84_UTM_zone_15S                = 32715
	PCS_WGS84_UTM_zone_16S                = 32716
	PCS_WGS84_UTM_zone_17S                = 32717
	PCS_WGS84_UTM_zone_18S                = 32718
	PCS_WGS84_UTM_zone_19S                = 32719
	PCS_WGS84_UTM_zone_20S                = 32720
	PCS_WGS84_UTM_zone_21S                = 32721
	PCS_WGS84_UTM_zone_22S                = 32722
	PCS_WGS84_UTM_zone_23S                = 32723
	PCS_WGS84_UTM_zone_24S                = 32724
	PCS_WGS84_UTM_zone_25S                = 32725
	PCS_WGS84_UTM_zone_26S                = 32726
	PCS_WGS84_UTM_zone_27S                = 32727
	PCS_WGS84_UTM_zone_28S                = 32728
	PCS_WGS84_UTM_zone_29S                = 32729
	PCS_WGS84_UTM_zone_30S                = 32730
	PCS_WGS84_UTM_zone_31S                = 32731
	PCS_WGS84_UTM_zone_32S                = 32732
	PCS_WGS84_UTM_zone_33S                = 32733
	PCS_WGS84_UTM_zone_34S                = 32734
	PCS_WGS84_UTM_zone_35S                = 32735
	PCS_WGS84_UTM_zone_36S                = 32736
	PCS_WGS84_UTM_zone_37S                = 32737
	PCS_WGS84_UTM_zone_38S                = 32738
	PCS_WGS84_UTM_zone_39S                = 32739
	PCS_WGS84_UTM_zone_40S                = 32740
	PCS_WGS84_UTM_zone_41S                = 32741
	PCS_WGS84_UTM_zone_42S                = 32742
	PCS_WGS84_UTM_zone_43S                = 32743
	PCS_WGS84_UTM_zone_44S                = 32744
	PCS_WGS84_UTM_zone_45S                = 32745
	PCS_WGS84_UTM_zone_46S                = 32746
	PCS_WGS84_UTM_zone_47S                = 32747
	PCS_WGS84_UTM_zone_48S                = 32748
	PCS_WGS84_UTM_zone_49S                = 32749
	PCS_WGS84_UTM_zone_50S                = 32750
	PCS_WGS84_UTM_zone_51S                = 32751
	PCS_WGS84_UTM_zone_52S                = 32752
	PCS_WGS84_UTM_zone_53S                = 32753
	PCS_WGS84_UTM_zone_54S                = 32754
	PCS_WGS84_UTM_zone_55S                = 32755
	PCS_WGS84_UTM_zone_56S                = 32756
	PCS_WGS84_UTM_zone_57S                = 32757
	PCS_WGS84_UTM_zone_58S                = 32758
	PCS_WGS84_UTM_zone_59S                = 32759
	PCS_WGS84_UTM_zone_60S                = 32760
)

var ProjectionCSTypeGeoKey map[int]string = map[int]string{
	20137: "Adindan UTM Zone 37N",
	20138: "Adindan UTM Zone 38N",
	20248: "AGD66 AMG Zone 48",
	20249: "AGD66 AMG Zone 49",
	20250: "AGD66 AMG Zone 50",
	20251: "AGD66 AMG Zone 51",
	20252: "AGD66 AMG Zone 52",
	20253: "AGD66 AMG Zone 53",
	20254: "AGD66 AMG Zone 54",
	20255: "AGD66 AMG Zone 55",
	20256: "AGD66 AMG Zone 56",
	20257: "AGD66 AMG Zone 57",
	20258: "AGD66 AMG Zone 58",
	20348: "AGD84 AMG Zone 48",
	20349: "AGD84 AMG Zone 49",
	20350: "AGD84 AMG Zone 50",
	20351: "AGD84 AMG Zone 51",
	20352: "AGD84 AMG Zone 52",
	20353: "AGD84 AMG Zone 53",
	20354: "AGD84 AMG Zone 54",
	20355: "AGD84 AMG Zone 55",
	20356: "AGD84 AMG Zone 56",
	20357: "AGD84 AMG Zone 57",
	20358: "AGD84 AMG Zone 58",
	20437: "Ain el Abd UTM Zone 37N",
	20438: "Ain el Abd UTM Zone 38N",
	20439: "Ain el Abd UTM Zone 39N",
	20499: "Ain el Abd Bahrain Grid",
	20538: "Afgooye UTM Zone 38N",
	20539: "Afgooye UTM Zone 39N",
	20700: "Lisbon Portugese Grid",
	20822: "Aratu UTM Zone 22S",
	20823: "Aratu UTM Zone 23S",
	20824: "Aratu UTM Zone 24S",
	20973: "Arc 1950 Lo13",
	20975: "Arc 1950 Lo15",
	20977: "Arc 1950 Lo17",
	20979: "Arc 1950 Lo19",
	20981: "Arc 1950 Lo21",
	20983: "Arc 1950 Lo23",
	20985: "Arc 1950 Lo25",
	20987: "Arc 1950 Lo27",
	20989: "Arc 1950 Lo29",
	20991: "Arc 1950 Lo31",
	20993: "Arc 1950 Lo33",
	20995: "Arc 1950 Lo35",
	21100: "Batavia NEIEZ",
	21148: "Batavia UTM Zone 48S",
	21149: "Batavia UTM Zone 49S",
	21150: "Batavia UTM Zone 50S",
	21413: "Beijing Gauss Zone 13",
	21414: "Beijing Gauss Zone 14",
	21415: "Beijing Gauss Zone 15",
	21416: "Beijing Gauss Zone 16",
	21417: "Beijing Gauss Zone 17",
	21418: "Beijing Gauss Zone 18",
	21419: "Beijing Gauss Zone 19",
	21420: "Beijing Gauss Zone 20",
	21421: "Beijing Gauss Zone 21",
	21422: "Beijing Gauss Zone 22",
	21423: "Beijing Gauss Zone 23",
	21473: "Beijing Gauss 13N",
	21474: "Beijing Gauss 14N",
	21475: "Beijing Gauss 15N",
	21476: "Beijing Gauss 16N",
	21477: "Beijing Gauss 17N",
	21478: "Beijing Gauss 18N",
	21479: "Beijing Gauss 19N",
	21480: "Beijing Gauss 20N",
	21481: "Beijing Gauss 21N",
	21482: "Beijing Gauss 22N",
	21483: "Beijing Gauss 23N",
	21500: "Belge Lambert 50",
	21790: "Bern 1898 Swiss Old",
	21817: "Bogota UTM Zone 17N",
	21818: "Bogota UTM Zone 18N",
	21891: "Bogota Colombia 3W",
	21892: "Bogota Colombia Bogota",
	21893: "Bogota Colombia 3E",
	21894: "Bogota Colombia 6E",
	22032: "Camacupa UTM 32S",
	22033: "Camacupa UTM 33S",
	22191: "C Inchauspe Argentina 1",
	22192: "C Inchauspe Argentina 2",
	22193: "C Inchauspe Argentina 3",
	22194: "C Inchauspe Argentina 4",
	22195: "C Inchauspe Argentina 5",
	22196: "C Inchauspe Argentina 6",
	22197: "C Inchauspe Argentina 7",
	22332: "Carthage UTM Zone 32N",
	22391: "Carthage Nord Tunisie",
	22392: "Carthage Sud Tunisie",
	22523: "Corrego Alegre UTM 23S",
	22524: "Corrego Alegre UTM 24S",
	22832: "Douala UTM Zone 32N",
	22992: "Egypt 1907 Red Belt",
	22993: "Egypt 1907 Purple Belt",
	22994: "Egypt 1907 Ext Purple",
	23028: "ED50 UTM Zone 28N",
	23029: "ED50 UTM Zone 29N",
	23030: "ED50 UTM Zone 30N",
	23031: "ED50 UTM Zone 31N",
	23032: "ED50 UTM Zone 32N",
	23033: "ED50 UTM Zone 33N",
	23034: "ED50 UTM Zone 34N",
	23035: "ED50 UTM Zone 35N",
	23036: "ED50 UTM Zone 36N",
	23037: "ED50 UTM Zone 37N",
	23038: "ED50 UTM Zone 38N",
	23239: "Fahud UTM Zone 39N",
	23240: "Fahud UTM Zone 40N",
	23433: "Garoua UTM Zone 33N",
	23846: "ID74 UTM Zone 46N",
	23847: "ID74 UTM Zone 47N",
	23848: "ID74 UTM Zone 48N",
	23849: "ID74 UTM Zone 49N",
	23850: "ID74 UTM Zone 50N",
	23851: "ID74 UTM Zone 51N",
	23852: "ID74 UTM Zone 52N",
	23853: "ID74 UTM Zone 53N",
	23886: "ID74 UTM Zone 46S",
	23887: "ID74 UTM Zone 47S",
	23888: "ID74 UTM Zone 48S",
	23889: "ID74 UTM Zone 49S",
	23890: "ID74 UTM Zone 50S",
	23891: "ID74 UTM Zone 51S",
	23892: "ID74 UTM Zone 52S",
	23893: "ID74 UTM Zone 53S",
	23894: "ID74 UTM Zone 54S",
	23947: "Indian 1954 UTM 47N",
	23948: "Indian 1954 UTM 48N",
	24047: "Indian 1975 UTM 47N",
	24048: "Indian 1975 UTM 48N",
	24100: "Jamaica 1875 Old Grid",
	24200: "JAD69 Jamaica Grid",
	24370: "Kalianpur India 0",
	24371: "Kalianpur India I",
	24372: "Kalianpur India IIa",
	24373: "Kalianpur India IIIa",
	24374: "Kalianpur India IVa",
	24382: "Kalianpur India IIb",
	24383: "Kalianpur India IIIb",
	24384: "Kalianpur India IVb",
	24500: "Kertau Singapore Grid",
	24547: "Kertau UTM Zone 47N",
	24548: "Kertau UTM Zone 48N",
	24720: "La Canoa UTM Zone 20N",
	24721: "La Canoa UTM Zone 21N",
	24818: "PSAD56 UTM Zone 18N",
	24819: "PSAD56 UTM Zone 19N",
	24820: "PSAD56 UTM Zone 20N",
	24821: "PSAD56 UTM Zone 21N",
	24877: "PSAD56 UTM Zone 17S",
	24878: "PSAD56 UTM Zone 18S",
	24879: "PSAD56 UTM Zone 19S",
	24880: "PSAD56 UTM Zone 20S",
	24891: "PSAD56 Peru west Zone",
	24892: "PSAD56 Peru central",
	24893: "PSAD56 Peru east Zone",
	25000: "Leigon Ghana Grid",
	25231: "Lome UTM Zone 31N",
	25391: "Luzon Philippines I",
	25392: "Luzon Philippines II",
	25393: "Luzon Philippines III",
	25394: "Luzon Philippines IV",
	25395: "Luzon Philippines V",
	25700: "Makassar NEIEZ",
	25932: "Malongo 1987 UTM 32S",
	26191: "Merchich Nord Maroc",
	26192: "Merchich Sud Maroc",
	26193: "Merchich Sahara",
	26237: "Massawa UTM Zone 37N",
	26331: "Minna UTM Zone 31N",
	26332: "Minna UTM Zone 32N",
	26391: "Minna Nigeria West",
	26392: "Minna Nigeria Mid Belt",
	26393: "Minna Nigeria East",
	26432: "Mhast UTM Zone 32S",
	26591: "Monte Mario Italy 1",
	26592: "Monte Mario Italy 2",
	26632: "M poraloko UTM 32N",
	26692: "M poraloko UTM 32S",
	26703: "NAD27 UTM Zone 3N",
	26704: "NAD27 UTM Zone 4N",
	26705: "NAD27 UTM Zone 5N",
	26706: "NAD27 UTM Zone 6N",
	26707: "NAD27 UTM Zone 7N",
	26708: "NAD27 UTM Zone 8N",
	26709: "NAD27 UTM Zone 9N",
	26710: "NAD27 UTM Zone 10N",
	26711: "NAD27 UTM Zone 11N",
	26712: "NAD27 UTM Zone 12N",
	26713: "NAD27 UTM Zone 13N",
	26714: "NAD27 UTM Zone 14N",
	26715: "NAD27 UTM Zone 15N",
	26716: "NAD27 UTM Zone 16N",
	26717: "NAD27 UTM Zone 17N",
	26718: "NAD27 UTM Zone 18N",
	26719: "NAD27 UTM Zone 19N",
	26720: "NAD27 UTM Zone 20N",
	26721: "NAD27 UTM Zone 21N",
	26722: "NAD27 UTM Zone 22N",
	26729: "NAD27 Alabama East",
	26730: "NAD27 Alabama West",
	26731: "NAD27 Alaska Zone 1",
	26732: "NAD27 Alaska Zone 2",
	26733: "NAD27 Alaska Zone 3",
	26734: "NAD27 Alaska Zone 4",
	26735: "NAD27 Alaska Zone 5",
	26736: "NAD27 Alaska Zone 6",
	26737: "NAD27 Alaska Zone 7",
	26738: "NAD27 Alaska Zone 8",
	26739: "NAD27 Alaska Zone 9",
	26740: "NAD27 Alaska Zone 10",
	26741: "NAD27 California I",
	26742: "NAD27 California II",
	26743: "NAD27 California III",
	26744: "NAD27 California IV",
	26745: "NAD27 California V",
	26746: "NAD27 California VI",
	26747: "NAD27 California VII",
	26748: "NAD27 Arizona East",
	26749: "NAD27 Arizona Central",
	26750: "NAD27 Arizona West",
	26751: "NAD27 Arkansas North",
	26752: "NAD27 Arkansas South",
	26753: "NAD27 Colorado North",
	26754: "NAD27 Colorado Central",
	26755: "NAD27 Colorado South",
	26756: "NAD27 Connecticut",
	26757: "NAD27 Delaware",
	26758: "NAD27 Florida East",
	26759: "NAD27 Florida West",
	26760: "NAD27 Florida North",
	26761: "NAD27 Hawaii Zone 1",
	26762: "NAD27 Hawaii Zone 2",
	26763: "NAD27 Hawaii Zone 3",
	26764: "NAD27 Hawaii Zone 4",
	26765: "NAD27 Hawaii Zone 5",
	26766: "NAD27 Georgia East",
	26767: "NAD27 Georgia West",
	26768: "NAD27 Idaho East",
	26769: "NAD27 Idaho Central",
	26770: "NAD27 Idaho West",
	26771: "NAD27 Illinois East",
	26772: "NAD27 Illinois West",
	26773: "NAD27 Indiana East",
	26774: "NAD27 BLM 14N feet",
	26775: "NAD27 Iowa North",
	26776: "NAD27 Iowa South",
	26777: "NAD27 Kansas North",
	26778: "NAD27 Kansas South",
	26779: "NAD27 Kentucky North",
	26780: "NAD27 Kentucky South",
	26781: "NAD27 Louisiana North",
	26782: "NAD27 Louisiana South",
	26783: "NAD27 Maine East",
	26784: "NAD27 Maine West",
	26785: "NAD27 Maryland",
	26786: "NAD27 Massachusetts",
	26787: "NAD27 Massachusetts Is",
	26788: "NAD27 Michigan North",
	26789: "NAD27 Michigan Central",
	26790: "NAD27 Michigan South",
	26791: "NAD27 Minnesota North",
	26792: "NAD27 Minnesota Cent",
	26793: "NAD27 Minnesota South",
	26794: "NAD27 Mississippi East",
	26795: "NAD27 Mississippi West",
	26796: "NAD27 Missouri East",
	26797: "NAD27 Missouri Central",
	26798: "NAD27 Missouri West",
	26801: "NAD Michigan Michigan East",
	26802: "NAD Michigan Michigan Old Central",
	26803: "NAD Michigan Michigan West",
	26903: "NAD83 UTM Zone 3N",
	26904: "NAD83 UTM Zone 4N",
	26905: "NAD83 UTM Zone 5N",
	26906: "NAD83 UTM Zone 6N",
	26907: "NAD83 UTM Zone 7N",
	26908: "NAD83 UTM Zone 8N",
	26909: "NAD83 UTM Zone 9N",
	26910: "NAD83 UTM Zone 10N",
	26911: "NAD83 UTM Zone 11N",
	26912: "NAD83 UTM Zone 12N",
	26913: "NAD83 UTM Zone 13N",
	26914: "NAD83 UTM Zone 14N",
	26915: "NAD83 UTM Zone 15N",
	26916: "NAD83 UTM Zone 16N",
	26917: "NAD83 UTM Zone 17N",
	26918: "NAD83 UTM Zone 18N",
	26919: "NAD83 UTM Zone 19N",
	26920: "NAD83 UTM Zone 20N",
	26921: "NAD83 UTM Zone 21N",
	26922: "NAD83 UTM Zone 22N",
	26923: "NAD83 UTM Zone 23N",
	26929: "NAD83 Alabama East",
	26930: "NAD83 Alabama West",
	26931: "NAD83 Alaska Zone 1",
	26932: "NAD83 Alaska Zone 2",
	26933: "NAD83 Alaska Zone 3",
	26934: "NAD83 Alaska Zone 4",
	26935: "NAD83 Alaska Zone 5",
	26936: "NAD83 Alaska Zone 6",
	26937: "NAD83 Alaska Zone 7",
	26938: "NAD83 Alaska Zone 8",
	26939: "NAD83 Alaska Zone 9",
	26940: "NAD83 Alaska Zone 10",
	26941: "NAD83 California 1",
	26942: "NAD83 California 2",
	26943: "NAD83 California 3",
	26944: "NAD83 California 4",
	26945: "NAD83 California 5",
	26946: "NAD83 California 6",
	26948: "NAD83 Arizona East",
	26949: "NAD83 Arizona Central",
	26950: "NAD83 Arizona West",
	26951: "NAD83 Arkansas North",
	26952: "NAD83 Arkansas South",
	26953: "NAD83 Colorado North",
	26954: "NAD83 Colorado Central",
	26955: "NAD83 Colorado South",
	26956: "NAD83 Connecticut",
	26957: "NAD83 Delaware",
	26958: "NAD83 Florida East",
	26959: "NAD83 Florida West",
	26960: "NAD83 Florida North",
	26961: "NAD83 Hawaii Zone 1",
	26962: "NAD83 Hawaii Zone 2",
	26963: "NAD83 Hawaii Zone 3",
	26964: "NAD83 Hawaii Zone 4",
	26965: "NAD83 Hawaii Zone 5",
	26966: "NAD83 Georgia East",
	26967: "NAD83 Georgia West",
	26968: "NAD83 Idaho East",
	26969: "NAD83 Idaho Central",
	26970: "NAD83 Idaho West",
	26971: "NAD83 Illinois East",
	26972: "NAD83 Illinois West",
	26973: "NAD83 Indiana East",
	26974: "NAD83 Indiana West",
	26975: "NAD83 Iowa North",
	26976: "NAD83 Iowa South",
	26977: "NAD83 Kansas North",
	26978: "NAD83 Kansas South",
	26979: "NAD83 Kentucky North",
	26980: "NAD83 Kentucky South",
	26981: "NAD83 Louisiana North",
	26982: "NAD83 Louisiana South",
	26983: "NAD83 Maine East",
	26984: "NAD83 Maine West",
	26985: "NAD83 Maryland",
	26986: "NAD83 Massachusetts",
	26987: "NAD83 Massachusetts Is",
	26988: "NAD83 Michigan North",
	26989: "NAD83 Michigan Central",
	26990: "NAD83 Michigan South",
	26991: "NAD83 Minnesota North",
	26992: "NAD83 Minnesota Cent",
	26993: "NAD83 Minnesota South",
	26994: "NAD83 Mississippi East",
	26995: "NAD83 Mississippi West",
	26996: "NAD83 Missouri East",
	26997: "NAD83 Missouri Central",
	26998: "NAD83 Missouri West",
	27038: "Nahrwan 1967 UTM 38N",
	27039: "Nahrwan 1967 UTM 39N",
	27040: "Nahrwan 1967 UTM 40N",
	27120: "Naparima UTM 20N",
	27200: "GD49 NZ Map Grid",
	27291: "GD49 North Island Grid",
	27292: "GD49 South Island Grid",
	27429: "Datum 73 UTM Zone 29N",
	27500: "ATF Nord de Guerre",
	27581: "NTF France I",
	27582: "NTF France II",
	27583: "NTF France III",
	27591: "NTF Nord France",
	27592: "NTF Centre France",
	27593: "NTF Sud France",
	27700: "British National Grid",
	28232: "Point Noire UTM 32S",
	28348: "GDA94 MGA Zone 48",
	28349: "GDA94 MGA Zone 49",
	28350: "GDA94 MGA Zone 50",
	28351: "GDA94 MGA Zone 51",
	28352: "GDA94 MGA Zone 52",
	28353: "GDA94 MGA Zone 53",
	28354: "GDA94 MGA Zone 54",
	28355: "GDA94 MGA Zone 55",
	28356: "GDA94 MGA Zone 56",
	28357: "GDA94 MGA Zone 57",
	28358: "GDA94 MGA Zone 58",
	28404: "Pulkovo Gauss Zone 4",
	28405: "Pulkovo Gauss Zone 5",
	28406: "Pulkovo Gauss Zone 6",
	28407: "Pulkovo Gauss Zone 7",
	28408: "Pulkovo Gauss Zone 8",
	28409: "Pulkovo Gauss Zone 9",
	28410: "Pulkovo Gauss Zone 10",
	28411: "Pulkovo Gauss Zone 11",
	28412: "Pulkovo Gauss Zone 12",
	28413: "Pulkovo Gauss Zone 13",
	28414: "Pulkovo Gauss Zone 14",
	28415: "Pulkovo Gauss Zone 15",
	28416: "Pulkovo Gauss Zone 16",
	28417: "Pulkovo Gauss Zone 17",
	28418: "Pulkovo Gauss Zone 18",
	28419: "Pulkovo Gauss Zone 19",
	28420: "Pulkovo Gauss Zone 20",
	28421: "Pulkovo Gauss Zone 21",
	28422: "Pulkovo Gauss Zone 22",
	28423: "Pulkovo Gauss Zone 23",
	28424: "Pulkovo Gauss Zone 24",
	28425: "Pulkovo Gauss Zone 25",
	28426: "Pulkovo Gauss Zone 26",
	28427: "Pulkovo Gauss Zone 27",
	28428: "Pulkovo Gauss Zone 28",
	28429: "Pulkovo Gauss Zone 29",
	28430: "Pulkovo Gauss Zone 30",
	28431: "Pulkovo Gauss Zone 31",
	28432: "Pulkovo Gauss Zone 32",
	28464: "Pulkovo Gauss 4N",
	28465: "Pulkovo Gauss 5N",
	28466: "Pulkovo Gauss 6N",
	28467: "Pulkovo Gauss 7N",
	28468: "Pulkovo Gauss 8N",
	28469: "Pulkovo Gauss 9N",
	28470: "Pulkovo Gauss 10N",
	28471: "Pulkovo Gauss 11N",
	28472: "Pulkovo Gauss 12N",
	28473: "Pulkovo Gauss 13N",
	28474: "Pulkovo Gauss 14N",
	28475: "Pulkovo Gauss 15N",
	28476: "Pulkovo Gauss 16N",
	28477: "Pulkovo Gauss 17N",
	28478: "Pulkovo Gauss 18N",
	28479: "Pulkovo Gauss 19N",
	28480: "Pulkovo Gauss 20N",
	28481: "Pulkovo Gauss 21N",
	28482: "Pulkovo Gauss 22N",
	28483: "Pulkovo Gauss 23N",
	28484: "Pulkovo Gauss 24N",
	28485: "Pulkovo Gauss 25N",
	28486: "Pulkovo Gauss 26N",
	28487: "Pulkovo Gauss 27N",
	28488: "Pulkovo Gauss 28N",
	28489: "Pulkovo Gauss 29N",
	28490: "Pulkovo Gauss 30N",
	28491: "Pulkovo Gauss 31N",
	28492: "Pulkovo Gauss 32N",
	28600: "Qatar National Grid",
	28991: "RD Netherlands Old",
	28992: "RD Netherlands New",
	29118: "SAD69 UTM Zone 18N",
	29119: "SAD69 UTM Zone 19N",
	29120: "SAD69 UTM Zone 20N",
	29121: "SAD69 UTM Zone 21N",
	29122: "SAD69 UTM Zone 22N",
	29177: "SAD69 UTM Zone 17S",
	29178: "SAD69 UTM Zone 18S",
	29179: "SAD69 UTM Zone 19S",
	29180: "SAD69 UTM Zone 20S",
	29181: "SAD69 UTM Zone 21S",
	29182: "SAD69 UTM Zone 22S",
	29183: "SAD69 UTM Zone 23S",
	29184: "SAD69 UTM Zone 24S",
	29185: "SAD69 UTM Zone 25S",
	29220: "Sapper Hill UTM 20S",
	29221: "Sapper Hill UTM 21S",
	29333: "Schwarzeck UTM 33S",
	29635: "Sudan UTM Zone 35N",
	29636: "Sudan UTM Zone 36N",
	29700: "Tananarive Laborde",
	29738: "Tananarive UTM 38S",
	29739: "Tananarive UTM 39S",
	29800: "Timbalai 1948 Borneo",
	29849: "Timbalai 1948 UTM 49N",
	29850: "Timbalai 1948 UTM 50N",
	29900: "TM65 Irish Nat Grid",
	30200: "Trinidad 1903 Trinidad",
	30339: "TC 1948 UTM Zone 39N",
	30340: "TC 1948 UTM Zone 40N",
	30491: "Voirol N Algerie ancien",
	30492: "Voirol S Algerie ancien",
	30591: "Voirol Unifie N Algerie",
	30592: "Voirol Unifie S Algerie",
	30600: "Bern 1938 Swiss New",
	30729: "Nord Sahara UTM 29N",
	30730: "Nord Sahara UTM 30N",
	30731: "Nord Sahara UTM 31N",
	30732: "Nord Sahara UTM 32N",
	31028: "Yoff UTM Zone 28N",
	31121: "Zanderij UTM Zone 21N",
	31291: "MGI Austria West",
	31292: "MGI Austria Central",
	31293: "MGI Austria East",
	31300: "Belge Lambert 72",
	31491: "DHDN Germany Zone 1",
	31492: "DHDN Germany Zone 2",
	31493: "DHDN Germany Zone 3",
	31494: "DHDN Germany Zone 4",
	31495: "DHDN Germany Zone 5",
	32001: "NAD27 Montana North",
	32002: "NAD27 Montana Central",
	32003: "NAD27 Montana South",
	32005: "NAD27 Nebraska North",
	32006: "NAD27 Nebraska South",
	32007: "NAD27 Nevada East",
	32008: "NAD27 Nevada Central",
	32009: "NAD27 Nevada West",
	32010: "NAD27 New Hampshire",
	32011: "NAD27 New Jersey",
	32012: "NAD27 New Mexico East",
	32013: "NAD27 New Mexico Cent",
	32014: "NAD27 New Mexico West",
	32015: "NAD27 New York East",
	32016: "NAD27 New York Central",
	32017: "NAD27 New York West",
	32018: "NAD27 New York Long Is",
	32019: "NAD27 North Carolina",
	32020: "NAD27 North Dakota N",
	32021: "NAD27 North Dakota S",
	32022: "NAD27 Ohio North",
	32023: "NAD27 Ohio South",
	32024: "NAD27 Oklahoma North",
	32025: "NAD27 Oklahoma South",
	32026: "NAD27 Oregon North",
	32027: "NAD27 Oregon South",
	32028: "NAD27 Pennsylvania N",
	32029: "NAD27 Pennsylvania S",
	32030: "NAD27 Rhode Island",
	32031: "NAD27 South Carolina N",
	32033: "NAD27 South Carolina S",
	32034: "NAD27 South Dakota N",
	32035: "NAD27 South Dakota S",
	32036: "NAD27 Tennessee",
	32037: "NAD27 Texas North",
	32038: "NAD27 Texas North Cen",
	32039: "NAD27 Texas Central",
	32040: "NAD27 Texas South Cen",
	32041: "NAD27 Texas South",
	32042: "NAD27 Utah North",
	32043: "NAD27 Utah Central",
	32044: "NAD27 Utah South",
	32045: "NAD27 Vermont",
	32046: "NAD27 Virginia North",
	32047: "NAD27 Virginia South",
	32048: "NAD27 Washington North",
	32049: "NAD27 Washington South",
	32050: "NAD27 West Virginia N",
	32051: "NAD27 West Virginia S",
	32052: "NAD27 Wisconsin North",
	32053: "NAD27 Wisconsin Cen",
	32054: "NAD27 Wisconsin South",
	32055: "NAD27 Wyoming East",
	32056: "NAD27 Wyoming E Cen",
	32057: "NAD27 Wyoming W Cen",
	32058: "NAD27 Wyoming West",
	32059: "NAD27 Puerto Rico",
	32060: "NAD27 St Croix",
	32100: "NAD83 Montana",
	32104: "NAD83 Nebraska",
	32107: "NAD83 Nevada East",
	32108: "NAD83 Nevada Central",
	32109: "NAD83 Nevada West",
	32110: "NAD83 New Hampshire",
	32111: "NAD83 New Jersey",
	32112: "NAD83 New Mexico East",
	32113: "NAD83 New Mexico Cent",
	32114: "NAD83 New Mexico West",
	32115: "NAD83 New York East",
	32116: "NAD83 New York Central",
	32117: "NAD83 New York West",
	32118: "NAD83 New York Long Is",
	32119: "NAD83 North Carolina",
	32120: "NAD83 North Dakota N",
	32121: "NAD83 North Dakota S",
	32122: "NAD83 Ohio North",
	32123: "NAD83 Ohio South",
	32124: "NAD83 Oklahoma North",
	32125: "NAD83 Oklahoma South",
	32126: "NAD83 Oregon North",
	32127: "NAD83 Oregon South",
	32128: "NAD83 Pennsylvania N",
	32129: "NAD83 Pennsylvania S",
	32130: "NAD83 Rhode Island",
	32133: "NAD83 South Carolina",
	32134: "NAD83 South Dakota N",
	32135: "NAD83 South Dakota S",
	32136: "NAD83 Tennessee",
	32137: "NAD83 Texas North",
	32138: "NAD83 Texas North Cen",
	32139: "NAD83 Texas Central",
	32140: "NAD83 Texas South Cen",
	32141: "NAD83 Texas South",
	32142: "NAD83 Utah North",
	32143: "NAD83 Utah Central",
	32144: "NAD83 Utah South",
	32145: "NAD83 Vermont",
	32146: "NAD83 Virginia North",
	32147: "NAD83 Virginia South",
	32148: "NAD83 Washington North",
	32149: "NAD83 Washington South",
	32150: "NAD83 West Virginia N",
	32151: "NAD83 West Virginia S",
	32152: "NAD83 Wisconsin North",
	32153: "NAD83 Wisconsin Cen",
	32154: "NAD83 Wisconsin South",
	32155: "NAD83 Wyoming East",
	32156: "NAD83 Wyoming E Cen",
	32157: "NAD83 Wyoming W Cen",
	32158: "NAD83 Wyoming West",
	32161: "NAD83 Puerto Rico Virgin Is",
	32201: "WGS72 UTM Zone 1N",
	32202: "WGS72 UTM Zone 2N",
	32203: "WGS72 UTM Zone 3N",
	32204: "WGS72 UTM Zone 4N",
	32205: "WGS72 UTM Zone 5N",
	32206: "WGS72 UTM Zone 6N",
	32207: "WGS72 UTM Zone 7N",
	32208: "WGS72 UTM Zone 8N",
	32209: "WGS72 UTM Zone 9N",
	32210: "WGS72 UTM Zone 10N",
	32211: "WGS72 UTM Zone 11N",
	32212: "WGS72 UTM Zone 12N",
	32213: "WGS72 UTM Zone 13N",
	32214: "WGS72 UTM Zone 14N",
	32215: "WGS72 UTM Zone 15N",
	32216: "WGS72 UTM Zone 16N",
	32217: "WGS72 UTM Zone 17N",
	32218: "WGS72 UTM Zone 18N",
	32219: "WGS72 UTM Zone 19N",
	32220: "WGS72 UTM Zone 20N",
	32221: "WGS72 UTM Zone 21N",
	32222: "WGS72 UTM Zone 22N",
	32223: "WGS72 UTM Zone 23N",
	32224: "WGS72 UTM Zone 24N",
	32225: "WGS72 UTM Zone 25N",
	32226: "WGS72 UTM Zone 26N",
	32227: "WGS72 UTM Zone 27N",
	32228: "WGS72 UTM Zone 28N",
	32229: "WGS72 UTM Zone 29N",
	32230: "WGS72 UTM Zone 30N",
	32231: "WGS72 UTM Zone 31N",
	32232: "WGS72 UTM Zone 32N",
	32233: "WGS72 UTM Zone 33N",
	32234: "WGS72 UTM Zone 34N",
	32235: "WGS72 UTM Zone 35N",
	32236: "WGS72 UTM Zone 36N",
	32237: "WGS72 UTM Zone 37N",
	32238: "WGS72 UTM Zone 38N",
	32239: "WGS72 UTM Zone 39N",
	32240: "WGS72 UTM Zone 40N",
	32241: "WGS72 UTM Zone 41N",
	32242: "WGS72 UTM Zone 42N",
	32243: "WGS72 UTM Zone 43N",
	32244: "WGS72 UTM Zone 44N",
	32245: "WGS72 UTM Zone 45N",
	32246: "WGS72 UTM Zone 46N",
	32247: "WGS72 UTM Zone 47N",
	32248: "WGS72 UTM Zone 48N",
	32249: "WGS72 UTM Zone 49N",
	32250: "WGS72 UTM Zone 50N",
	32251: "WGS72 UTM Zone 51N",
	32252: "WGS72 UTM Zone 52N",
	32253: "WGS72 UTM Zone 53N",
	32254: "WGS72 UTM Zone 54N",
	32255: "WGS72 UTM Zone 55N",
	32256: "WGS72 UTM Zone 56N",
	32257: "WGS72 UTM Zone 57N",
	32258: "WGS72 UTM Zone 58N",
	32259: "WGS72 UTM Zone 59N",
	32260: "WGS72 UTM Zone 60N",
	32301: "WGS72 UTM Zone 1S",
	32302: "WGS72 UTM Zone 2S",
	32303: "WGS72 UTM Zone 3S",
	32304: "WGS72 UTM Zone 4S",
	32305: "WGS72 UTM Zone 5S",
	32306: "WGS72 UTM Zone 6S",
	32307: "WGS72 UTM Zone 7S",
	32308: "WGS72 UTM Zone 8S",
	32309: "WGS72 UTM Zone 9S",
	32310: "WGS72 UTM Zone 10S",
	32311: "WGS72 UTM Zone 11S",
	32312: "WGS72 UTM Zone 12S",
	32313: "WGS72 UTM Zone 13S",
	32314: "WGS72 UTM Zone 14S",
	32315: "WGS72 UTM Zone 15S",
	32316: "WGS72 UTM Zone 16S",
	32317: "WGS72 UTM Zone 17S",
	32318: "WGS72 UTM Zone 18S",
	32319: "WGS72 UTM Zone 19S",
	32320: "WGS72 UTM Zone 20S",
	32321: "WGS72 UTM Zone 21S",
	32322: "WGS72 UTM Zone 22S",
	32323: "WGS72 UTM Zone 23S",
	32324: "WGS72 UTM Zone 24S",
	32325: "WGS72 UTM Zone 25S",
	32326: "WGS72 UTM Zone 26S",
	32327: "WGS72 UTM Zone 27S",
	32328: "WGS72 UTM Zone 28S",
	32329: "WGS72 UTM Zone 29S",
	32330: "WGS72 UTM Zone 30S",
	32331: "WGS72 UTM Zone 31S",
	32332: "WGS72 UTM Zone 32S",
	32333: "WGS72 UTM Zone 33S",
	32334: "WGS72 UTM Zone 34S",
	32335: "WGS72 UTM Zone 35S",
	32336: "WGS72 UTM Zone 36S",
	32337: "WGS72 UTM Zone 37S",
	32338: "WGS72 UTM Zone 38S",
	32339: "WGS72 UTM Zone 39S",
	32340: "WGS72 UTM Zone 40S",
	32341: "WGS72 UTM Zone 41S",
	32342: "WGS72 UTM Zone 42S",
	32343: "WGS72 UTM Zone 43S",
	32344: "WGS72 UTM Zone 44S",
	32345: "WGS72 UTM Zone 45S",
	32346: "WGS72 UTM Zone 46S",
	32347: "WGS72 UTM Zone 47S",
	32348: "WGS72 UTM Zone 48S",
	32349: "WGS72 UTM Zone 49S",
	32350: "WGS72 UTM Zone 50S",
	32351: "WGS72 UTM Zone 51S",
	32352: "WGS72 UTM Zone 52S",
	32353: "WGS72 UTM Zone 53S",
	32354: "WGS72 UTM Zone 54S",
	32355: "WGS72 UTM Zone 55S",
	32356: "WGS72 UTM Zone 56S",
	32357: "WGS72 UTM Zone 57S",
	32358: "WGS72 UTM Zone 58S",
	32359: "WGS72 UTM Zone 59S",
	32360: "WGS72 UTM Zone 60S",
	32401: "WGS72BE UTM Zone 1N",
	32402: "WGS72BE UTM Zone 2N",
	32403: "WGS72BE UTM Zone 3N",
	32404: "WGS72BE UTM Zone 4N",
	32405: "WGS72BE UTM Zone 5N",
	32406: "WGS72BE UTM Zone 6N",
	32407: "WGS72BE UTM Zone 7N",
	32408: "WGS72BE UTM Zone 8N",
	32409: "WGS72BE UTM Zone 9N",
	32410: "WGS72BE UTM Zone 10N",
	32411: "WGS72BE UTM Zone 11N",
	32412: "WGS72BE UTM Zone 12N",
	32413: "WGS72BE UTM Zone 13N",
	32414: "WGS72BE UTM Zone 14N",
	32415: "WGS72BE UTM Zone 15N",
	32416: "WGS72BE UTM Zone 16N",
	32417: "WGS72BE UTM Zone 17N",
	32418: "WGS72BE UTM Zone 18N",
	32419: "WGS72BE UTM Zone 19N",
	32420: "WGS72BE UTM Zone 20N",
	32421: "WGS72BE UTM Zone 21N",
	32422: "WGS72BE UTM Zone 22N",
	32423: "WGS72BE UTM Zone 23N",
	32424: "WGS72BE UTM Zone 24N",
	32425: "WGS72BE UTM Zone 25N",
	32426: "WGS72BE UTM Zone 26N",
	32427: "WGS72BE UTM Zone 27N",
	32428: "WGS72BE UTM Zone 28N",
	32429: "WGS72BE UTM Zone 29N",
	32430: "WGS72BE UTM Zone 30N",
	32431: "WGS72BE UTM Zone 31N",
	32432: "WGS72BE UTM Zone 32N",
	32433: "WGS72BE UTM Zone 33N",
	32434: "WGS72BE UTM Zone 34N",
	32435: "WGS72BE UTM Zone 35N",
	32436: "WGS72BE UTM Zone 36N",
	32437: "WGS72BE UTM Zone 37N",
	32438: "WGS72BE UTM Zone 38N",
	32439: "WGS72BE UTM Zone 39N",
	32440: "WGS72BE UTM Zone 40N",
	32441: "WGS72BE UTM Zone 41N",
	32442: "WGS72BE UTM Zone 42N",
	32443: "WGS72BE UTM Zone 43N",
	32444: "WGS72BE UTM Zone 44N",
	32445: "WGS72BE UTM Zone 45N",
	32446: "WGS72BE UTM Zone 46N",
	32447: "WGS72BE UTM Zone 47N",
	32448: "WGS72BE UTM Zone 48N",
	32449: "WGS72BE UTM Zone 49N",
	32450: "WGS72BE UTM Zone 50N",
	32451: "WGS72BE UTM Zone 51N",
	32452: "WGS72BE UTM Zone 52N",
	32453: "WGS72BE UTM Zone 53N",
	32454: "WGS72BE UTM Zone 54N",
	32455: "WGS72BE UTM Zone 55N",
	32456: "WGS72BE UTM Zone 56N",
	32457: "WGS72BE UTM Zone 57N",
	32458: "WGS72BE UTM Zone 58N",
	32459: "WGS72BE UTM Zone 59N",
	32460: "WGS72BE UTM Zone 60N",
	32501: "WGS72BE UTM Zone 1S",
	32502: "WGS72BE UTM Zone 2S",
	32503: "WGS72BE UTM Zone 3S",
	32504: "WGS72BE UTM Zone 4S",
	32505: "WGS72BE UTM Zone 5S",
	32506: "WGS72BE UTM Zone 6S",
	32507: "WGS72BE UTM Zone 7S",
	32508: "WGS72BE UTM Zone 8S",
	32509: "WGS72BE UTM Zone 9S",
	32510: "WGS72BE UTM Zone 10S",
	32511: "WGS72BE UTM Zone 11S",
	32512: "WGS72BE UTM Zone 12S",
	32513: "WGS72BE UTM Zone 13S",
	32514: "WGS72BE UTM Zone 14S",
	32515: "WGS72BE UTM Zone 15S",
	32516: "WGS72BE UTM Zone 16S",
	32517: "WGS72BE UTM Zone 17S",
	32518: "WGS72BE UTM Zone 18S",
	32519: "WGS72BE UTM Zone 19S",
	32520: "WGS72BE UTM Zone 20S",
	32521: "WGS72BE UTM Zone 21S",
	32522: "WGS72BE UTM Zone 22S",
	32523: "WGS72BE UTM Zone 23S",
	32524: "WGS72BE UTM Zone 24S",
	32525: "WGS72BE UTM Zone 25S",
	32526: "WGS72BE UTM Zone 26S",
	32527: "WGS72BE UTM Zone 27S",
	32528: "WGS72BE UTM Zone 28S",
	32529: "WGS72BE UTM Zone 29S",
	32530: "WGS72BE UTM Zone 30S",
	32531: "WGS72BE UTM Zone 31S",
	32532: "WGS72BE UTM Zone 32S",
	32533: "WGS72BE UTM Zone 33S",
	32534: "WGS72BE UTM Zone 34S",
	32535: "WGS72BE UTM Zone 35S",
	32536: "WGS72BE UTM Zone 36S",
	32537: "WGS72BE UTM Zone 37S",
	32538: "WGS72BE UTM Zone 38S",
	32539: "WGS72BE UTM Zone 39S",
	32540: "WGS72BE UTM Zone 40S",
	32541: "WGS72BE UTM Zone 41S",
	32542: "WGS72BE UTM Zone 42S",
	32543: "WGS72BE UTM Zone 43S",
	32544: "WGS72BE UTM Zone 44S",
	32545: "WGS72BE UTM Zone 45S",
	32546: "WGS72BE UTM Zone 46S",
	32547: "WGS72BE UTM Zone 47S",
	32548: "WGS72BE UTM Zone 48S",
	32549: "WGS72BE UTM Zone 49S",
	32550: "WGS72BE UTM Zone 50S",
	32551: "WGS72BE UTM Zone 51S",
	32552: "WGS72BE UTM Zone 52S",
	32553: "WGS72BE UTM Zone 53S",
	32554: "WGS72BE UTM Zone 54S",
	32555: "WGS72BE UTM Zone 55S",
	32556: "WGS72BE UTM Zone 56S",
	32557: "WGS72BE UTM Zone 57S",
	32558: "WGS72BE UTM Zone 58S",
	32559: "WGS72BE UTM Zone 59S",
	32560: "WGS72BE UTM Zone 60S",
	32601: "WGS84 UTM Zone 1N",
	32602: "WGS84 UTM Zone 2N",
	32603: "WGS84 UTM Zone 3N",
	32604: "WGS84 UTM Zone 4N",
	32605: "WGS84 UTM Zone 5N",
	32606: "WGS84 UTM Zone 6N",
	32607: "WGS84 UTM Zone 7N",
	32608: "WGS84 UTM Zone 8N",
	32609: "WGS84 UTM Zone 9N",
	32610: "WGS84 UTM Zone 10N",
	32611: "WGS84 UTM Zone 11N",
	32612: "WGS84 UTM Zone 12N",
	32613: "WGS84 UTM Zone 13N",
	32614: "WGS84 UTM Zone 14N",
	32615: "WGS84 UTM Zone 15N",
	32616: "WGS84 UTM Zone 16N",
	32617: "WGS84 UTM Zone 17N",
	32618: "WGS84 UTM Zone 18N",
	32619: "WGS84 UTM Zone 19N",
	32620: "WGS84 UTM Zone 20N",
	32621: "WGS84 UTM Zone 21N",
	32622: "WGS84 UTM Zone 22N",
	32623: "WGS84 UTM Zone 23N",
	32624: "WGS84 UTM Zone 24N",
	32625: "WGS84 UTM Zone 25N",
	32626: "WGS84 UTM Zone 26N",
	32627: "WGS84 UTM Zone 27N",
	32628: "WGS84 UTM Zone 28N",
	32629: "WGS84 UTM Zone 29N",
	32630: "WGS84 UTM Zone 30N",
	32631: "WGS84 UTM Zone 31N",
	32632: "WGS84 UTM Zone 32N",
	32633: "WGS84 UTM Zone 33N",
	32634: "WGS84 UTM Zone 34N",
	32635: "WGS84 UTM Zone 35N",
	32636: "WGS84 UTM Zone 36N",
	32637: "WGS84 UTM Zone 37N",
	32638: "WGS84 UTM Zone 38N",
	32639: "WGS84 UTM Zone 39N",
	32640: "WGS84 UTM Zone 40N",
	32641: "WGS84 UTM Zone 41N",
	32642: "WGS84 UTM Zone 42N",
	32643: "WGS84 UTM Zone 43N",
	32644: "WGS84 UTM Zone 44N",
	32645: "WGS84 UTM Zone 45N",
	32646: "WGS84 UTM Zone 46N",
	32647: "WGS84 UTM Zone 47N",
	32648: "WGS84 UTM Zone 48N",
	32649: "WGS84 UTM Zone 49N",
	32650: "WGS84 UTM Zone 50N",
	32651: "WGS84 UTM Zone 51N",
	32652: "WGS84 UTM Zone 52N",
	32653: "WGS84 UTM Zone 53N",
	32654: "WGS84 UTM Zone 54N",
	32655: "WGS84 UTM Zone 55N",
	32656: "WGS84 UTM Zone 56N",
	32657: "WGS84 UTM Zone 57N",
	32658: "WGS84 UTM Zone 58N",
	32659: "WGS84 UTM Zone 59N",
	32660: "WGS84 UTM Zone 60N",
	32701: "WGS84 UTM Zone 1S",
	32702: "WGS84 UTM Zone 2S",
	32703: "WGS84 UTM Zone 3S",
	32704: "WGS84 UTM Zone 4S",
	32705: "WGS84 UTM Zone 5S",
	32706: "WGS84 UTM Zone 6S",
	32707: "WGS84 UTM Zone 7S",
	32708: "WGS84 UTM Zone 8S",
	32709: "WGS84 UTM Zone 9S",
	32710: "WGS84 UTM Zone 10S",
	32711: "WGS84 UTM Zone 11S",
	32712: "WGS84 UTM Zone 12S",
	32713: "WGS84 UTM Zone 13S",
	32714: "WGS84 UTM Zone 14S",
	32715: "WGS84 UTM Zone 15S",
	32716: "WGS84 UTM Zone 16S",
	32717: "WGS84 UTM Zone 17S",
	32718: "WGS84 UTM Zone 18S",
	32719: "WGS84 UTM Zone 19S",
	32720: "WGS84 UTM Zone 20S",
	32721: "WGS84 UTM Zone 21S",
	32722: "WGS84 UTM Zone 22S",
	32723: "WGS84 UTM Zone 23S",
	32724: "WGS84 UTM Zone 24S",
	32725: "WGS84 UTM Zone 25S",
	32726: "WGS84 UTM Zone 26S",
	32727: "WGS84 UTM Zone 27S",
	32728: "WGS84 UTM Zone 28S",
	32729: "WGS84 UTM Zone 29S",
	32730: "WGS84 UTM Zone 30S",
	32731: "WGS84 UTM Zone 31S",
	32732: "WGS84 UTM Zone 32S",
	32733: "WGS84 UTM Zone 33S",
	32734: "WGS84 UTM Zone 34S",
	32735: "WGS84 UTM Zone 35S",
	32736: "WGS84 UTM Zone 36S",
	32737: "WGS84 UTM Zone 37S",
	32738: "WGS84 UTM Zone 38S",
	32739: "WGS84 UTM Zone 39S",
	32740: "WGS84 UTM Zone 40S",
	32741: "WGS84 UTM Zone 41S",
	32742: "WGS84 UTM Zone 42S",
	32743: "WGS84 UTM Zone 43S",
	32744: "WGS84 UTM Zone 44S",
	32745: "WGS84 UTM Zone 45S",
	32746: "WGS84 UTM Zone 46S",
	32747: "WGS84 UTM Zone 47S",
	32748: "WGS84 UTM Zone 48S",
	32749: "WGS84 UTM Zone 49S",
	32750: "WGS84 UTM Zone 50S",
	32751: "WGS84 UTM Zone 51S",
	32752: "WGS84 UTM Zone 52S",
	32753: "WGS84 UTM Zone 53S",
	32754: "WGS84 UTM Zone 54S",
	32755: "WGS84 UTM Zone 55S",
	32756: "WGS84 UTM Zone 56S",
	32757: "WGS84 UTM Zone 57S",
	32758: "WGS84 UTM Zone 58S",
	32759: "WGS84 UTM Zone 59S",
	32760: "WGS84 UTM Zone 60S"}

var PcsZones map[string]uint16 = map[string]uint16{
	"1N":  PCS_WGS84_UTM_zone_1N,
	"2N":  PCS_WGS84_UTM_zone_2N,
	"3N":  PCS_WGS84_UTM_zone_3N,
	"4N":  PCS_WGS84_UTM_zone_4N,
	"5N":  PCS_WGS84_UTM_zone_5N,
	"6N":  PCS_WGS84_UTM_zone_6N,
	"7N":  PCS_WGS84_UTM_zone_7N,
	"8N":  PCS_WGS84_UTM_zone_8N,
	"9N":  PCS_WGS84_UTM_zone_9N,
	"10N": PCS_WGS84_UTM_zone_10N,
	"11N": PCS_WGS84_UTM_zone_11N,
	"12N": PCS_WGS84_UTM_zone_12N,
	"13N": PCS_WGS84_UTM_zone_13N,
	"14N": PCS_WGS84_UTM_zone_14N,
	"15N": PCS_WGS84_UTM_zone_15N,
	"16N": PCS_WGS84_UTM_zone_16N,
	"17N": PCS_WGS84_UTM_zone_17N,
	"18N": PCS_WGS84_UTM_zone_18N,
	"19N": PCS_WGS84_UTM_zone_19N,
	"20N": PCS_WGS84_UTM_zone_20N,
	"21N": PCS_WGS84_UTM_zone_21N,
	"22N": PCS_WGS84_UTM_zone_22N,
	"23N": PCS_WGS84_UTM_zone_23N,
	"24N": PCS_WGS84_UTM_zone_24N,
	"25N": PCS_WGS84_UTM_zone_25N,
	"26N": PCS_WGS84_UTM_zone_26N,
	"27N": PCS_WGS84_UTM_zone_27N,
	"28N": PCS_WGS84_UTM_zone_28N,
	"29N": PCS_WGS84_UTM_zone_29N,
	"30N": PCS_WGS84_UTM_zone_30N,
	"31N": PCS_WGS84_UTM_zone_31N,
	"32N": PCS_WGS84_UTM_zone_32N,
	"33N": PCS_WGS84_UTM_zone_33N,
	"34N": PCS_WGS84_UTM_zone_34N,
	"35N": PCS_WGS84_UTM_zone_35N,
	"36N": PCS_WGS84_UTM_zone_36N,
	"37N": PCS_WGS84_UTM_zone_37N,
	"38N": PCS_WGS84_UTM_zone_38N,
	"39N": PCS_WGS84_UTM_zone_39N,
	"40N": PCS_WGS84_UTM_zone_40N,
	"41N": PCS_WGS84_UTM_zone_41N,
	"42N": PCS_WGS84_UTM_zone_42N,
	"43N": PCS_WGS84_UTM_zone_43N,
	"44N": PCS_WGS84_UTM_zone_44N,
	"45N": PCS_WGS84_UTM_zone_45N,
	"46N": PCS_WGS84_UTM_zone_46N,
	"47N": PCS_WGS84_UTM_zone_47N,
	"48N": PCS_WGS84_UTM_zone_48N,
	"49N": PCS_WGS84_UTM_zone_49N,
	"50N": PCS_WGS84_UTM_zone_50N,
	"51N": PCS_WGS84_UTM_zone_51N,
	"52N": PCS_WGS84_UTM_zone_52N,
	"53N": PCS_WGS84_UTM_zone_53N,
	"54N": PCS_WGS84_UTM_zone_54N,
	"55N": PCS_WGS84_UTM_zone_55N,
	"56N": PCS_WGS84_UTM_zone_56N,
	"57N": PCS_WGS84_UTM_zone_57N,
	"58N": PCS_WGS84_UTM_zone_58N,
	"59N": PCS_WGS84_UTM_zone_59N,
	"60N": PCS_WGS84_UTM_zone_60N,
	"1S":  PCS_WGS84_UTM_zone_1S,
	"2S":  PCS_WGS84_UTM_zone_2S,
	"3S":  PCS_WGS84_UTM_zone_3S,
	"4S":  PCS_WGS84_UTM_zone_4S,
	"5S":  PCS_WGS84_UTM_zone_5S,
	"6S":  PCS_WGS84_UTM_zone_6S,
	"7S":  PCS_WGS84_UTM_zone_7S,
	"8S":  PCS_WGS84_UTM_zone_8S,
	"9S":  PCS_WGS84_UTM_zone_9S,
	"10S": PCS_WGS84_UTM_zone_10S,
	"11S": PCS_WGS84_UTM_zone_11S,
	"12S": PCS_WGS84_UTM_zone_12S,
	"13S": PCS_WGS84_UTM_zone_13S,
	"14S": PCS_WGS84_UTM_zone_14S,
	"15S": PCS_WGS84_UTM_zone_15S,
	"16S": PCS_WGS84_UTM_zone_16S,
	"17S": PCS_WGS84_UTM_zone_17S,
	"18S": PCS_WGS84_UTM_zone_18S,
	"19S": PCS_WGS84_UTM_zone_19S,
	"20S": PCS_WGS84_UTM_zone_20S,
	"21S": PCS_WGS84_UTM_zone_21S,
	"22S": PCS_WGS84_UTM_zone_22S,
	"23S": PCS_WGS84_UTM_zone_23S,
	"24S": PCS_WGS84_UTM_zone_24S,
	"25S": PCS_WGS84_UTM_zone_25S,
	"26S": PCS_WGS84_UTM_zone_26S,
	"27S": PCS_WGS84_UTM_zone_27S,
	"28S": PCS_WGS84_UTM_zone_28S,
	"29S": PCS_WGS84_UTM_zone_29S,
	"30S": PCS_WGS84_UTM_zone_30S,
	"31S": PCS_WGS84_UTM_zone_31S,
	"32S": PCS_WGS84_UTM_zone_32S,
	"33S": PCS_WGS84_UTM_zone_33S,
	"34S": PCS_WGS84_UTM_zone_34S,
	"35S": PCS_WGS84_UTM_zone_35S,
	"36S": PCS_WGS84_UTM_zone_36S,
	"37S": PCS_WGS84_UTM_zone_37S,
	"38S": PCS_WGS84_UTM_zone_38S,
	"39S": PCS_WGS84_UTM_zone_39S,
	"40S": PCS_WGS84_UTM_zone_40S,
	"41S": PCS_WGS84_UTM_zone_41S,
	"42S": PCS_WGS84_UTM_zone_42S,
	"43S": PCS_WGS84_UTM_zone_43S,
	"44S": PCS_WGS84_UTM_zone_44S,
	"45S": PCS_WGS84_UTM_zone_45S,
	"46S": PCS_WGS84_UTM_zone_46S,
	"47S": PCS_WGS84_UTM_zone_47S,
	"48S": PCS_WGS84_UTM_zone_48S,
	"49S": PCS_WGS84_UTM_zone_49S,
	"50S": PCS_WGS84_UTM_zone_50S,
	"51S": PCS_WGS84_UTM_zone_51S,
	"52S": PCS_WGS84_UTM_zone_52S,
	"53S": PCS_WGS84_UTM_zone_53S,
	"54S": PCS_WGS84_UTM_zone_54S,
	"55S": PCS_WGS84_UTM_zone_55S,
	"56S": PCS_WGS84_UTM_zone_56S,
	"57S": PCS_WGS84_UTM_zone_57S,
	"58S": PCS_WGS84_UTM_zone_58S,
	"59S": PCS_WGS84_UTM_zone_59S,
	"60S": PCS_WGS84_UTM_zone_60S}
