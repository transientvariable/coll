package trie

import (
	"fmt"
	"strings"
)

// Digitizer ...
type Digitizer interface {
	// Base returns the base for the Digitizer.
	Base() int

	// IsPrefixFree returns true if and only if the Digitizer guarantees that no node is a prefix of another.
	IsPrefixFree() bool

	// NumDigitsOf returns the number of digits in the provided node.
	NumDigitsOf(value string) int

	// DigitOf returns the element of digit place for the provided node. The returned error will be non-nil if
	// the Digitizer does support the character set of the provided string, or if the place is greater than
	// Digitizer.Base().
	DigitOf(value string, place int) (int, error)

	// FormatDigit returns a string representation of the digit in the place specified for the given node. The returned
	// error will be non-nil if the Digitizer does support the character set of the provided string, or if the place is
	// greater than Digitizer.Base().
	FormatDigit(value string, place int) (string, error)
}

type asciiDigitizer struct {
	base int
}

// NewASCIIDigitizer creates a new Digitizer that uses the ASCII character set for digitizing strings. The base for
// the Digitizer will be the sum of printable ASCII characters (95) plus 1 for end of string character.
func NewASCIIDigitizer() Digitizer {
	return &asciiDigitizer{base: len(asciiTable) + 1}
}

// Base the base of the alphabet used by the ASCII Digitizer that includes the end of string character.
func (d *asciiDigitizer) Base() int {
	return d.base
}

// IsPrefixFree returns true since the ASCII Digitizer is a prefix free.
func (d *asciiDigitizer) IsPrefixFree() bool {
	return true
}

// NumDigitsOf returns the number of digits in the provided string including the end of string character. The returned
// error will be non-nil if the Digitizer does support the character set of the provided string, or if the place is
// greater than Digitizer.Base().
func (d *asciiDigitizer) NumDigitsOf(value string) int {
	return len(value) + 1
}

// DigitOf returns the integer element mapped to by the digit in the given place. The returned error will be non-nil if
// the Digitizer does support the character set of the provided string, or if the place is greater than
// Digitizer.Base().
func (d *asciiDigitizer) DigitOf(value string, place int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" || place >= len(value) {
		return 0, nil
	}

	if place > d.Base() {
		return -1, fmt.Errorf("digitizer_ascii: requested place is greater than the supported alphabet size: %d", d.Base())
	}

	i, ok := asciiTable[rune(value[place])]
	if !ok {
		return -1, fmt.Errorf("digitizer_ascii: character for node is unsupported: node = %s, place = %d, character = %c", value, place, value[place])
	}
	return i, nil
}

// FormatDigit returns a string representation of the digit in the place specified for the given node where '#' is
// used for the end of string character.
func (d *asciiDigitizer) FormatDigit(value string, place int) (string, error) {
	i, err := d.DigitOf(value, place)
	if err != nil {
		return "", err
	}

	if i == 0 {
		return "#", nil
	}
	return string(value[place]), nil
}

var asciiTable = map[rune]int{
	' ':  1,
	'!':  2,
	'"':  3,
	'#':  4,
	'$':  5,
	'%':  6,
	'&':  7,
	'\'': 8,
	'(':  9,
	')':  10,
	'*':  11,
	'+':  12,
	',':  13,
	'-':  14,
	'.':  15,
	'/':  16,
	'0':  17,
	'1':  18,
	'2':  19,
	'3':  20,
	'4':  21,
	'5':  22,
	'6':  23,
	'7':  24,
	'8':  25,
	'9':  26,
	':':  27,
	';':  28,
	'<':  29,
	'=':  30,
	'>':  31,
	'?':  32,
	'@':  33,
	'A':  34,
	'B':  35,
	'C':  36,
	'D':  37,
	'E':  38,
	'F':  39,
	'G':  40,
	'H':  41,
	'I':  42,
	'J':  43,
	'K':  44,
	'L':  45,
	'M':  46,
	'N':  47,
	'O':  48,
	'P':  49,
	'Q':  50,
	'R':  51,
	'S':  52,
	'T':  53,
	'U':  54,
	'V':  55,
	'W':  56,
	'X':  57,
	'Y':  58,
	'Z':  59,
	'[':  60,
	'\\': 61,
	']':  62,
	'^':  63,
	'_':  64,
	'`':  65,
	'a':  66,
	'b':  67,
	'c':  68,
	'd':  69,
	'e':  70,
	'f':  71,
	'g':  72,
	'h':  73,
	'i':  74,
	'j':  75,
	'k':  76,
	'l':  77,
	'm':  78,
	'n':  79,
	'o':  80,
	'p':  81,
	'q':  82,
	'r':  83,
	's':  84,
	't':  85,
	'u':  86,
	'v':  87,
	'w':  88,
	'x':  89,
	'y':  90,
	'z':  91,
	'{':  92,
	'|':  93,
	'}':  94,
	'~':  95,
}
