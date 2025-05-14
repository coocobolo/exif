package tag

type Exif uint16

const (
	ExifOffset       Exif = 0x8769
	ExposureTime     Exif = 0x829A
	FNumber          Exif = 0x829D
	GPSInfo          Exif = 0x8825
	ImageDescription Exif = 0x010E
	Make             Exif = 0x010F
	Model            Exif = 0x0110
	ModifyDate       Exif = 0x0132
	Orientation      Exif = 0x0112
	ResolutionUnit   Exif = 0x0128
	Software         Exif = 0x0131
	SubfileType      Exif = 0x00FE
	Unknown          Exif = 0x00
	XResolution      Exif = 0x011A
	YCbCrPositioning Exif = 0x0213
	YResolution      Exif = 0x011B
)

var (
	ExifTags = map[uint16]Exif{
		0x8769: ExifOffset,
		0x829A: ExposureTime,
		0x829D: FNumber,
		0x8825: GPSInfo,
		0x010E: ImageDescription,
		0x010F: Make,
		0x0110: Model,
		0x0132: ModifyDate,
		0x0112: Orientation,
		0x0128: ResolutionUnit,
		0x0131: Software,
		0x00FE: SubfileType,
		0x00:   Unknown,
		0x011A: XResolution,
		0x0213: YCbCrPositioning,
		0x011B: YResolution,
	}
	ExifTagStr = map[Exif]string{
		ExifOffset:       "ExifOffset",
		ExposureTime:     "ExposureTime",
		FNumber:          "FNumber",
		GPSInfo:          "GPSInfo",
		ImageDescription: "ImageDescription",
		Make:             "Make",
		Model:            "Model",
		ModifyDate:       "ModifyDate",
		Orientation:      "Orientation",
		ResolutionUnit:   "ResolutionUnit",
		Software:         "Software",
		SubfileType:      "SubfileType",
		Unknown:          "Unknown",
		XResolution:      "XResolution",
		YCbCrPositioning: "YCbCrPositioning",
		YResolution:      "YResolution",
	}
)

func (e *Exif) String() string {
	if e == nil {
		*e = Unknown
		return e.String()
	}
	return ExifTagStr[*e]
}

type GPS uint16

const (
	GPSAltitude     GPS = 0x0006
	GPSAltitudeRef  GPS = 0x0005
	GPSLatitude     GPS = 0x0002
	GPSLatitudeRef  GPS = 0x0001
	GPSLongitude    GPS = 0x0004
	GPSLongitudeRef GPS = 0x0003
	GPSVersionID    GPS = 0x0000
)

var GPSTags = map[uint16]GPS{
	0x0006: GPSAltitude,
	0x0005: GPSAltitudeRef,
	0x0002: GPSLatitude,
	0x0001: GPSLatitudeRef,
	0x0004: GPSLongitude,
	0x0003: GPSLongitudeRef,
	0x0000: GPSVersionID,
}
