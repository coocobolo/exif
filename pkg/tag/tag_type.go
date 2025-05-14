package tag

type Type uint16

const (
	BYTE      Type = 0x01
	ASCII     Type = 0x02
	SHORT     Type = 0x03
	LONG      Type = 0x04
	RATIONAL  Type = 0x05
	SBYTE     Type = 0x06
	UNDEFINED Type = 0x07
	SSHORT    Type = 0x08
	SLONG     Type = 0x09
	SRATIONAL Type = 0x0A
	FLOAT     Type = 0x0B
	DOUBLE    Type = 0x0C
)

var TypeSizes = map[Type]uint16{
	BYTE:      1,
	ASCII:     1,
	SHORT:     2,
	LONG:      4,
	RATIONAL:  8, // two LONGs (2 x 4 bytes)
	SBYTE:     1,
	UNDEFINED: 1,
	SSHORT:    2,
	SLONG:     4,
	SRATIONAL: 8, // two SLONGs (2 x 4 bytes)
	FLOAT:     4,
	DOUBLE:    8,
}
