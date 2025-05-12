package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
)

const (
	imgPath = "DSCN0012.jpg"
	// imgPath = "image.jpeg"
	// imgPath = "Crémieux11.tiff"
)

const (
	// magic
	MarkerAPP1 = 0xFF

	MarkerEXIF uint64 = 0xFFE1

	MarkerZERO = 0x00

	// A JPEG file is a sequence of Type-Length-Value chunks called segments: 0xFF; marker:c; length(value+2):>u2; value;length*c;
	// Magic Signature: at offset `0`, called Start of Image (SOI): 0xFFD8
	MarkerSOI = 0xD8

	// terminator -> at, the end of the file, called End of Image (EOI): D9
	MarkerEOI = 0xD9

	// ECS -> Entropy-Coded Segment
	// restart markers, FF D0 - D7, which are just optional indicators in the middle of the ECS data.
	MarkerRestart = 0xD0 // 0xD0 - 0xD7
)

const (
	SegmentCodeSOI  = 0xD8
	SegmentCodeSOS  = 0xDA
	SegmentCodeEOI  = 0xD9
	SegmentCodeAPP0 = 0xE0
)

// JPEG files start with a Start of Image (SOI) marker (0xFFD8)
// and contain various segments, each beginning with a marker (0xFF followed by a segment code).
// Metadata is typically stored in segments like APP0, APP1, APP2,
// APP13, and COM before the Start of Scan (SOS, 0xFFDA) marker.

// The Identify Segments can look by `0xFF` followed by a segment code

// Its length is unknown in advance, nor defined in the file.
// The only way to get its length is to either decode it or to
// fast-forward over it: just scan forward for a FF byte.
// If it's a restart marker (followed by D0 - D7)
// or a data FF (followed by 00), continue.

// It is read as:
//   -  a Start of Image marker, FF D8. This is the signature, enforced at offset 0.
//   -  a segment: with an Application 0 marker (encoded FF E0) and a length of 16 (encoded 00 10)
//   -  its data:
//        -  a JFIF\0 signature.
//        -  then the rest of the APP0 chunk, of little interest here..

// isStandalone checks if a marker code corresponds to a standalone marker (no length field).
/* func isStandalone(code byte) bool {
	return code == 0xD8 || code == 0xD9 || code == 0x01 || (code >= 0xD0 && code <= 0xD7)
} */

// splitJPEGSegments splits a JPEG file into chunks based on its segments.
/* func splitJPEGSegments(data []byte) [][]byte {
	var chunks [][]byte
	pos := 0

	log.Printf("%X\n", data)
	for pos < len(data) {
		if v := data[pos]; v != 0xFF {
			log.Printf("not a marker: %X\n", data[pos])
			// Not a marker; skip or handle as part of entropy-coded data
			pos++
			continue
		}

		if pos+1 >= len(data) {
			break
		}

		nextByte := data[pos+1]
		if nextByte == 0x00 {
			// Byte stuffing; skip both bytes
			pos += 2
			continue
		}

		// Potential marker
		code := nextByte
		if isStandalone(code) {
			// Standalone marker (e.g., SOI, EOI, TEM, RSTm)
			if pos+1 < len(data) {
				chunk := make([]byte, 2)
				copy(chunk, data[pos:pos+2])
				chunks = append(chunks, chunk)
				pos += 2
			} else {
				break
			}
		} else {
			// Data-containing segment (e.g., APP0, APP1, DQT, SOF, DHT, SOS)
			if pos+3 >= len(data) {
				break
			}
			L := binary.BigEndian.Uint16(data[pos+2 : pos+4])
			totalSize := 2 + int(L) // Marker (2) + Parameter segment (L)
			if pos+totalSize > len(data) {
				break // Incomplete segment
			}
			chunk := make([]byte, totalSize)
			copy(chunk, data[pos:pos+totalSize])
			chunks = append(chunks, chunk)
			pos += totalSize
		}
	}

	return chunks
} */

func splitter(v []byte) (segments [][]byte, err error) {
	const (
		sof0 = 0xC0 // SOF0 marker
	)

	log.Printf("%X\n", v)
	for i := 0; i < len(v)-1; {
		if v[i] != MarkerAPP1 {
			i++
			continue
		}

		// skip any padding FF's (0xFF 0xFF ..)
		j := i + 1
		for j < len(v) && v[j] == MarkerAPP1 {
			log.Println("skip the padding 0xFF marker")
			j++
		}

		if j >= len(v) {
			return nil, fmt.Errorf("truncated marker at %d", i)
		}

		marker := v[j]
		// 0x00 is a "stuffed" zero byte (not a real marker),
		// it's fine to skip it

		if marker == 0x00 {
			i = j + 1
			continue
		}

		log.Printf("current marker: at %d: %X", j, marker)
		// now v[i] == 0xFF, v[j] == marker
		// some markers (SOI 0xD8, EOI 0xD9, RSTn 0xD0-D7)
		// have no length field mean it's standalone marker
		if marker == 0xD8 || marker == 0xD9 || (marker >= 0xD0 && marker <= 0xD7) {
			log.Printf("found standalone marker: at %d: %X", i, marker)
			segments = append(segments, v[i:j+1])
			i = j + 1
			continue
		}

		// otherwise the two bytes after the marker are a big-endian length:
		if j+2 >= len(v) {
			return nil, fmt.Errorf("truncated length at %d", j)
		}

		length := int(binary.BigEndian.Uint16(v[j+1 : j+3]))
		end := j + 1 + length

		if end > len(v) {
			return nil, fmt.Errorf("segment at %d overruns buffer: want %d, have %d", i, end, len(v))
		}
		segments = append(segments, v[i:end])
		i = end
	}

	return
}

/* func splitAPP1EXIF(data []byte) ([][]byte, error) {
	// Check if data is long enough for EXIF identifier
	if len(data) < 6 {
		return nil, fmt.Errorf("data too short for EXIF identifier")
	}

	// Verify EXIF identifier ("Exif\0\0")
	if string(data[0:6]) != "Exif\x00\x00" {
		return nil, fmt.Errorf("invalid EXIF identifier")
	}
	chunks := [][]byte{data[0:6]} // Add EXIF identifier as first chunk

	// Check if data is long enough for TIFF header
	if len(data) < 14 {
		return nil, fmt.Errorf("data too short for TIFF header")
	}

	// Determine byte order
	var byteOrder binary.ByteOrder
	switch string(data[6:8]) {
	case "II":
		byteOrder = binary.LittleEndian
	case "MM":
		byteOrder = binary.BigEndian
	default:
		return nil, fmt.Errorf("unsupported byte order")
	}

	// Verify TIFF identifier (42)
	tiffID := byteOrder.Uint16(data[8:10])
	if tiffID != 42 {
		return nil, fmt.Errorf("invalid TIFF identifier")
	}

	// Read offset to 0th IFD
	offset0th := byteOrder.Uint32(data[10:14])
	start0th := int(offset0th) + 6 // Offset is relative to start of TIFF header

	// Validate 0th IFD offset
	if start0th >= len(data) || start0th+2 > len(data) {
		return nil, fmt.Errorf("0th IFD offset out of bounds")
	}

	// Read number of entries in 0th IFD
	numEntries0th := byteOrder.Uint16(data[start0th : start0th+2])
	ifdSize0th := 2 + 12*int(numEntries0th) + 4 // 2 bytes for count, 12 bytes per entry, 4 bytes for next offset
	end0th := start0th + ifdSize0th

	// Validate 0th IFD size
	if end0th > len(data) {
		return nil, fmt.Errorf("0th IFD size exceeds data length")
	}

	// Add TIFF header and 0th IFD to chunks
	chunks = append(chunks, data[6:14])            // TIFF header
	chunks = append(chunks, data[start0th:end0th]) // 0th IFD

	// Check for 1st IFD
	nextOffset := byteOrder.Uint32(data[end0th-4 : end0th])
	if nextOffset != 0 {
		start1st := int(nextOffset) + 6 // Offset relative to start of TIFF header

		// Validate 1st IFD offset
		if start1st >= len(data) || start1st+2 > len(data) {
			return nil, fmt.Errorf("1st IFD offset out of bounds")
		}

		// Read number of entries in 1st IFD
		numEntries1st := byteOrder.Uint16(data[start1st : start1st+2])
		ifdSize1st := 2 + 12*int(numEntries1st) + 4
		end1st := start1st + ifdSize1st

		// Validate 1st IFD size
		if end1st > len(data) {
			return nil, fmt.Errorf("1st IFD size exceeds data length")
		}

		// Add 1st IFD to chunks
		chunks = append(chunks, data[start1st:end1st])
	}

	return chunks, nil
} */

// readAPP1Segment reads the APP1 segment from a JPEG file
/* func readAPP1Segment(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the entire file into memory (for simplicity; use buffering for large files)
	var data []byte
	buf := make([]byte, 1)
	for {
		_, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		data = append(data, buf[0])
	}

	// Locate APP1 marker (0xFFE1)
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xE1 {
			length := binary.BigEndian.Uint16(data[i+2 : i+4])
			app1Data := data[i+4 : i+4+int(length)-2] // Exclude marker and length fields
			return app1Data, nil
		}
	}
	return nil, fmt.Errorf("APP1 segment not found")
} */

func main() {
	log.SetOutput(os.Stdout)

	// Open and Read the file
	f, err := os.Open(imgPath)
	if err != nil {
		log.Fatal("failed to open file", err)
	}

	fs, _ := f.Stat()
	log.Println("file size:", fs.Size())

	// TODO Identify segments
	// Look for `0xFF` follow by a segment code
	// (e.g., 0xE0 -> APP0, 0xE1 APP1)
	// TODO get magic signature
	buf := make([]byte, fs.Size())
	if n, xerr := f.Read(buf); xerr != nil || int64(n) != fs.Size() {
		log.Fatal("failed to read file:", xerr)
	}

	chunks, err := splitter(buf)
	if err != nil {
		log.Fatal("failed to split buffer file to segments:", err)
	}

	for i, v := range chunks {
		log.Printf("segment-%d: %X | %d\n", i, v, len(v))
	}

	/* app1Seg, err := splitAPP1EXIF(chunks[1][4:])
	if err != nil {
		log.Fatal("failed to parse app1 exif:", err)
	}

	for i, v := range app1Seg {
		log.Printf("APP1 segment-%d: %X\n", i, v)
	} */

	if err = ParseAPP1(chunks[1]); err != nil {
		log.Fatal("failed to parse APP1 Segment:", err)
	}

	return
	/* r := bufio.NewReader(f)

	// TODO Read SOI (Start of Image)
	magicSign := make([]byte, 2)
	n, err := r.Read(magicSign)
	if err != nil {
		err = errors.Wrap(err, "no bytes available in second one SOI")
		log.Fatal(err)
	}

	if !slices.Equal(magicSign, []byte{MarkerAPP1, MarkerSOI}) {
		log.Println("enforced at offset 0 is not SOI marker:", string(magicSign))
		return
	}

	log.Println("SOI", n, "-", string(magicSign))

	for {
		b, err := r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("EOF, done")
				break
			}

			log.Println("failed to reading metadata bytes:", err)
			break
		}

		if b != MarkerAPP1 {
			// log.Printf("expect to be 0xFF flag, got: %X\n", string(b))
			continue
		}

		code, err := r.ReadByte()
		if err != nil {
			log.Println("failed to reading metadata after flag 0xFF:", err)
			break
		}

		log.Printf("found marker: %X %X\n", b, code)

		switch code {
		// EOI
		case 0xD9:
			log.Println("EOI reached!")

		// SOS
		case 0xDA:
			log.Println("SOS found, start of image data")

		default:
			// count length of current segment
			lengthBytes := make([]byte, 2)
			_, err := r.Read(lengthBytes)
			if err != nil {
				log.Println("failed to read with 2 bytes len:", err)
				break
			}

			L := binary.BigEndian.Uint16(lengthBytes)
			if L < 2 {
				log.Println("invalid length:", L)
				continue
			}

			dataSize := int(L - 2)
			data := make([]byte, dataSize)

			_, err = r.Read(data)
			if err != nil {
				log.Println("failed to read metadata:", err)
				break
			}

			log.Printf("Segment 0xFF%02X, length %d, data size %d\n", code, L, dataSize)
			switch code {
			case 0xFE:
				log.Println("Comment:", string(data))
			case 0xE0:
				log.Println("FF E0 APP0 marker found")
				log.Printf("data: %X", data)
				if len(data) >= 4 && string(data[0:4]) == "JFIF" {
					log.Println("JFIF found")
					if len(data) >= 7 {
						versionMajor := data[5]
						versionMinor := data[6]
						log.Printf("JFIF version: %d.%d\n", versionMajor, versionMinor)
					}

					units := data[7]
					log.Printf("units: %x\n", units)

				}
			case 0xE1:
				log.Println("EXIF Found")
				if len(data) >= 6 && string(data[0:6]) == "Exif\\0\\0" {
					if len(data) >= 10 {
						byteOrder := string(data[6:8])
						log.Println("Byte order:", byteOrder)
						if len(data) >= 14 {
							tiffID := binary.BigEndian.Uint32(data[8:12])
							if tiffID == 0x002A {
								log.Println("TIFF ID correct")
								ifdOffset := binary.BigEndian.Uint32(data[12:16])
								log.Printf("First IFD offset: %d\n", ifdOffset)
							}
						}
					} else {
						log.Printf("data: %v\n", string(data))
					}
				}

			default:
				// log.Printf("Unknown segment or not parsed yet: segment 0xFF%02X\n", code)
			}
		}

	} */

	// TODO Read Segment Length
	// After the marker, read a 2-byte length field (big-endian) to determine the segment's data size.
}

// FFE1 [2] -> APP1 Marker
// SSSS [2] -> APP1 Data Size (bytes). NOTE the size `SSSS` includes the size of descriptor as well
// 45 78 69 66 00 00 [6] -> Exif Header. this is a special data to identify whether EXIF or not, ASCII chars "Exif" and 2 bytes of 0x00 used. After the APP1 Marker area, the other JPEG Markers follows.
// 4949 2A00 0800 0000  [8] -> TIFF Header. NOTE if tiff header this is

// Roughly structure of Exif data (APP1) is shown as below.
// This is a case of "Intel" byte align and it contains JPEG format thumbnail.
// As described above, Exif data is starts from ASCII character "Exif" and 2bytes of 0x00,
// then Exif data follows. Exif uses TIFF format to store data.
// For more datails of TIFF format, please refer to "TIFF6.0 specification".

// param (v) is slice byte that refer to the current APP1 segment
const (
	MOVE_ONE_BYTE = iota + 1
	MOVE_TWO_BYTES
	MOVE_SIX_BYTES = iota * 3
)

func ParseAPP1(v []byte) error {
	log.Printf("v len: %d", len(v))

	log.Println("start parsing APP1 EXIF Marker...")

	data := make(map[string]any)
	pos := 0
	offset := MOVE_TWO_BYTES

	// APP1 Marker
	if !bytes.Equal(v[pos:pos+offset], []byte{0xFF, 0xE1}) {
		return errors.New("invalid marker EXIF APP1!")
	}

	data["marker"] = fmt.Sprintf("0x%X", v[pos:pos+offset])
	pos = pos + offset

	// APP1 Data Size
	data["data_size"] = binary.BigEndian.Uint16(v[pos : pos+offset])
	pos = pos + offset

	offset = MOVE_SIX_BYTES
	exifHeader, ok := bytes.CutSuffix(v[pos:pos+offset], []byte{MarkerZERO, MarkerZERO})
	if !ok {
		return errors.New("failed to get marker exif header on app1")
	}
	data["exif_header"] = string(exifHeader)
	exifHeaderBound := pos
	pos = pos + offset

	offset = MOVE_TWO_BYTES
	endian, ok := EndianTypeFromStr[string(v[pos:pos+offset])]
	if !ok {
		return fmt.Errorf("failed to get endianness: %d", endian.ID())
	}
	data["endianness"] = endian.String()

	bo := endian.ByteOrder()

	tiffBase := uint32(6)
	v2 := v[exifHeaderBound:]
	firstIFDOffset := bo.Uint32(v2[tiffBase+4 : tiffBase+8])
	log.Printf("0th IFD offset: %d", firstIFDOffset)

	log.Printf("%X\n", v[exifHeaderBound:])
	out, next := parseIFD(v2, bo, tiffBase, firstIFDOffset)
	log.Printf("next: %d\n", next)
	for name, e := range out {
		fmt.Printf("%s Tag=0x%04X Type=%d Count=%d Value=%v\n",
			name, e.TagID, e.TypeID, e.Count, e.Value)
	}

	gpsEntry, found := out["GPSInfo"]
	if !found {
		log.Println("No GPSInfo tag in first IFD")
		return nil
	}

	// 2. gpsEntry.TypeID should be 4 (LONG), gpsEntry.Count == 1, and gpsEntry.Value is the offset:
	gpsOffset, ok := gpsEntry.Value.(uint32)
	if !ok {
		log.Fatalf("unexpected GPSInfo value type: %T", gpsEntry.Value)
	}

	// 3. Extend your tagNames map with GPS tags:
	// …add any others you care about (see https://exiftool.org/TagNames/GPS.html)…

	// 4. Parse the GPS IFD at tiffBase + gpsOffset:
	gpsMap, _ := parseIFD(v2, bo, tiffBase, gpsOffset)

	fmt.Println("\nGPS IFD entries:")
	for name, e := range gpsMap {
		fmt.Printf("  %-20s Tag=0x%04X Type=%d Count=%d Value=%v\n",
			name, e.TagID, e.TypeID, e.Count, e.Value)
	}

	for k, v := range data {
		log.Printf("%s: %v\n", k, v)
	}
	return nil
}

type IFDEntry struct {
	TagID  uint16
	TypeID uint16
	Count  uint32
	Value  any
}

// IfdEntry represents a parsed EXIF IFD entry.
type IfdEntry struct {
	TagID  uint16
	TypeID uint16
	Count  uint32
	Value  any
}

// tagNames maps common EXIF tag IDs to human-readable names.
var tagNames = map[uint16]string{
	0x010F: "Make",
	0x0110: "Model",
	0x8769: "ExifOffset",
	0x829A: "ExposureTime",
	0x829D: "FNumber",
	0x00FE: "SubfileType",
	0x0131: "Software",
	0x011B: "YResolution",
	0x011A: "XResolution",
	0x0213: "YCbCrPositioning",
	0x0112: "Orientation", // TODO enum
	0x010E: "ImageDescription",
	0x0132: "ModifyDate",
	0x0128: "ResolutionUnit",

	// GPS
	0x8825: "GPSInfo", // marker
	0x0000: "GPSVersionID",
	0x0001: "GPSLatitudeRef",
	0x0002: "GPSLatitude",
	0x0003: "GPSLongitudeRef",
	0x0004: "GPSLongitude",
	0x0005: "GPSAltitudeRef",
	0x0006: "GPSAltitude",
}

// parseIFD parses an IFD at the given offset (relative to tiffBase) and returns
// a map of tag name to IfdEntry, plus the next IFD offset (relative to tiffBase).
func parseIFD(data []byte, bo binary.ByteOrder, tiffBase, offset uint32) (map[string]IfdEntry, uint32) {
	start := int(tiffBase + offset)
	if len(data) < start+2 {
		log.Fatalf("data too short for IFD count at %d", start)
	}
	num := bo.Uint16(data[start : start+2])
	pos := start + 2
	results := make(map[string]IfdEntry, num)

	for i := range int(num) {
		if pos+12 > len(data) {
			log.Fatalf("data too short for IFD entry %d", i)
		}
		tag := bo.Uint16(data[pos : pos+2])
		tp := bo.Uint16(data[pos+2 : pos+4])
		cnt := bo.Uint32(data[pos+4 : pos+8])
		valp := data[pos+8 : pos+12]
		var val any

		switch tp {
		case 1, 2: // BYTE or ASCII
			if cnt <= 4 {
				raw := valp[:cnt]
				if tp == 2 {
					val = string(bytes.TrimRight(raw, "\x00"))
				} else {
					val = raw
				}
			} else {
				off := bo.Uint32(valp)
				s := int(tiffBase + off)
				e := s + int(cnt)
				if e > len(data) {
					log.Fatalf("invalid offset: %d + %d > %d", s, cnt, len(data))
				}
				chunk := data[s:e]
				if tp == 2 {
					val = string(chunk[:len(chunk)-1]) // drop NUL
				} else {
					val = chunk
				}
			}
		case 3: // SHORT
			arr := make([]uint16, cnt)
			if cnt <= 2 {
				for j := uint32(0); j < cnt; j++ {
					arr[j] = bo.Uint16(valp[j*2 : j*2+2])
				}
			} else {
				off := bo.Uint32(valp)
				for j := uint32(0); j < cnt; j++ {
					idx := int(tiffBase + off + j*2)
					arr[j] = bo.Uint16(data[idx : idx+2])
				}
			}
			val = arr
		case 4: // LONG
			if cnt == 1 {
				val = bo.Uint32(valp)
			} else {
				arr := make([]uint32, cnt)
				off := bo.Uint32(valp)
				for j := uint32(0); j < cnt; j++ {
					idx := int(tiffBase + off + j*4)
					arr[j] = bo.Uint32(data[idx : idx+4])
				}
				val = arr
			}
		case 5: // RATIONAL
			rats := make([][2]uint32, cnt)
			off := bo.Uint32(valp)
			for j := uint32(0); j < cnt; j++ {
				idx := int(tiffBase + off + j*8)
				n := bo.Uint32(data[idx : idx+4])
				d := bo.Uint32(data[idx+4 : idx+8])
				rats[j] = [2]uint32{n, d}
			}
			val = rats
		default:
			val = valp // raw fallback
		}

		name, ok := tagNames[tag]
		if !ok {
			name = fmt.Sprintf("UnknownTag(0x%04X)", tag)
		}
		results[name] = IfdEntry{TagID: tag, TypeID: tp, Count: cnt, Value: val}
		pos += 12
	}

	if pos+4 > len(data) {
		log.Fatalf("data too short for next IFD pointer at %d", pos)
	}
	next := bo.Uint32(data[pos : pos+4])
	return results, next
}
