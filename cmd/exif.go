package main

// EXIF APP1 segment
// EXIF (Exchangable Image File Format) JPEG file use APP1 segments
// in order not to conflict with JFIF files (which use APP0).
// Exif APP1 segments store a great amount of information on photographic
// parameters for digital cameras. Also, Exif are the preferred way to
// store thumbnail images nowadays. They can also host an additional
// section with GPS data. The reference document for Exif 2.2 and the
// the Interoperability standards are respectively:
//
// *B<"Exchangeable image file format for digital still cameras:
// Exif Version 2.2", JEITA CP-3451, Apr 2002
// Japan Electronic Industry Development Association (JEIDA)>*
//
// B<"Design rule for Camera File system", (DCF), v1.0
// English Version 1999.1.7, Adopted December 1998
// Japan Electronic Industry Development Association (JEIDA)>
//
// The TIFF (Tagged Image File format) standard documents, as
// well as some updates and corrections, are useful:
// Exif APP1 segments are made up by an identifier,
// a TIFF header and a sequence of IFDs (Image File Directories)
// and subIFDs. The high level IFDs are only two (IFD0, for
// photographic parameters, and IFD1 for thumbnail parameters);
// they can be followed by thumbnail data. The structure is as follows:
// [Record name]    [size]   [description]
// ---------------------------------------
// Identifier       6 bytes   ("Exif\000\000" = 0x457869660000), not stored
// Endianness       2 bytes   'II' (little-endian) or 'MM' (big-endian)
// Signature        2 bytes   a fixed value = 42
// IFD0_Pointer     4 bytes   offset of 0th IFD (usually 8), not stored
// IFD0                ...    main image IFD
// IFD0@SubIFD         ...    Exif private tags (optional, linked by IFD0)
// IFD0@SubIFD@Interop ...    Interoperability IFD (optional,linked by SubIFD)
// IFD0@GPS            ...    GPS IFD (optional, linked by IFD0)
// APP1@IFD1           ...    thumbnail IFD (optional, pointed to by IFD0)
// ThumbnailData       ...    Thumbnail image (optional, 0xffd8.....ffd9)
type APP1 struct {
	// ("Exif\000\000" = 0x457869660000), not stored
	Identifier [6]byte `json:"identifier"`
}
