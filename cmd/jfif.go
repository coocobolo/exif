package main

type UnitType byte

const (
	Unit_UNKNOWN UnitType = iota // 0
	Unit_INCH                    // 1
	Unit_CM                      // 2
)

type ExtensionCode byte

const (
	// 0x10 Thumbnail coded using JPEG
	ExtensionCode_JPEG ExtensionCode = 0x10

	// 0x11 Thumbnail using 1 byte/pixel
	ExtensionCode_BYTE_PER_PIXEL ExtensionCode = 0x11

	// 0x13 Thumbnail using 3 bytes/pixel
	ExtensionCode_3_BYTES_PER_PIXEL ExtensionCode = 0x11
)

// JFIF APP0 Segments are used in the old JFIF standard
// to store information about the picture dimensions and
// an optional thumbnail.
// The format of a JFIF APP0 segment is as follows
// (not that size of thumbnail data is 3n, where n = Xthumbnal * Ythumbnail, it's present if n is not zero; only the first 8 records are mandatory):
// There is also an extended JFIF
// (only possible for JFIF versions 1.02 and above).
// In this case the identifier is not JFIF but JFXX.
// This extension allows for the inclusion of differently
// encoded thumbnails. The syntax in this case is modified as follows:
type APP0 struct {
	// (JFIF\000 = 0x4a46494600) or (JFXX\000 = 0x4a46585800; on JFIF v1.02+)
	Identifier string `json:"identifier"`

	// thumbnail image data
	ThumbnailData []byte `json:"thumbnailData"`

	// horizontal pixel density
	XDensity int `json:"xDensity"`

	// vertical pixel density
	YDensity int `json:"yDensity"`

	MajorVersion byte `json:"majorVersion"`

	MinorVersion byte `json:"minorVersion"`

	Units UnitType `json:"units"`

	// thumbnail horizontal pixel count
	XThumbnail byte `json:"xThumbnail"`

	// thumbnail vertical pixel count
	YThumbnail byte `json:"yThumbnail"`

	// (0x10 Thumbnail coded using JPEG
	ExtensionCode ExtensionCode `json:"extensionCode"`

	// If the thumbnail is coded using a JPEG stream,
	// a binary JPEG stream immediately follows the
	// extension code (the byte count of this file is
	// included in the byte count of the APP0 Segment).
	// This stream conforms to the syntax for a JPEG file
	// (SOI .... SOF ... EOI);
	// however, no 'JFIF' or 'JFXX' marker Segments should be present:
	// a variable length JPEG picture of thumbnail
	JPEGThumbnail []byte `json:"jpegThumbnail"`

	// If the thumbnail is stored using one byte per pixel,
	// after the extension code one should find a palette
	// and an indexed RGB. The records are as follows
	// (remember that n = Xthumbnail * Ythumbnail):

	// 24-bit RGB values for the colour palette
	// (defining the colours represented by each
	// value of an 8-bit binary enconding)
	ColorPalette [768]*byte `json:"collorPalette"`

	// 8-bit indexed values for the thumbnail
	OneByteThumbnail []*byte `json:"oneByteThumbnail"`

	// If the thumbnail is stored using three bytes per pixel,
	// there is no colour palette, so the previous fields simplify into:
	// 24-bit RGB values for the thumbnail
	ThreeBytesThumbnail []*byte `json:"threeBytesThumbnail"`
}
