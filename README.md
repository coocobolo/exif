JPEG files start with a Start of Image (SOI) marker (0xFFD8)
and contain various segments, each beginning with a marker (0xFF followed by a segment code).
Metadata is typically stored in segments like APP0, APP1, APP2,
APP13, and COM before the Start of Scan (SOS, 0xFFDA) marker.

The Identify Segments can look by `0xFF` followed by a segment code

Its length is unknown in advance, nor defined in the file.
The only way to get its length is to either decode it or to
fast-forward over it: just scan forward for a FF byte.
If it's a restart marker (followed by D0 - D7)
or a data FF (followed by 00), continue.

It is read as:

- a Start of Image marker, FF D8. This is the signature, enforced at offset 0.
- a segment: with an Application 0 marker (encoded FF E0) and a length of 16 (encoded 00 10)
- its data:
  - a JFIF\0 signature.
  - then the rest of the APP0 chunk, of little interest here..

FFE1 [2] -> APP1 Marker
SSSS [2] -> APP1 Data Size (bytes). NOTE the size `SSSS` includes the size of descriptor as well
45 78 69 66 00 00 [6] -> Exif Header. this is a special data to identify whether EXIF or not, ASCII chars "Exif" and 2 bytes of 0x00 used. After the APP1 Marker area, the other JPEG Markers follows.
4949 2A00 0800 0000 [8] -> TIFF Header. NOTE if tiff header this is

Roughly structure of Exif data (APP1) is shown as below.
This is a case of "Intel" byte align and it contains JPEG format thumbnail.
As described above, Exif data is starts from ASCII character "Exif" and 2bytes of 0x00,
then Exif data follows. Exif uses TIFF format to store data.
For more datails of TIFF format, please refer to "TIFF6.0 specification".

param (v) is slice byte that refer to the current APP1 segment

REFERENCES:

- https://www.loc.gov/preservation/digital/formats/fdd/fdd000022.shtml
- https://web.archive.org/web/20190624045241if_/http://www.cipa.jp:80/std/documents/e/DC-008-Translation-2019-E.pdf
- https://www.media.mit.edu/pia/Research/deepview/exif.html
- https://en.wikipedia.org/wiki/Specials_(Unicode_block)
- https://exiftool.org/TagNames/EXIF.html
- https://exiftool.org/TagNames/GPS.html
