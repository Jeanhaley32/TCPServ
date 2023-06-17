// Description: This file contains the color enums and colorWrap function
package main

// defining Color Enums
type Color int64

const (
	Red Color = iota
	Green
	Yellow
	Blue
	Purple
	Cyan
	Gray
	White
	Black
	LightGray
	DarkGray
	LightRed
	LightGreen
	LightYellow
	LightBlue
	LightMagenta
	LightCyan
)

// Returns color as a string
func (c Color) Color() string {
	switch c {
	case Red:
		return "\033[31m"
	case Green:
		return "\033[32m"
	case Yellow:
		return "\033[33m"
	case Blue:
		return "\033[34m"
	case Purple:
		return "\033[35m"
	case Cyan:
		return "\033[36m"
	case Gray:
		return "\033[37m"
	case White:
		return "\033[97m"
	case Black:
		return "\033[30m"
	case LightGray:
		return "\033[37m"
	case DarkGray:
		return "\033[90m"
	case LightRed:
		return "\033[91m"
	case LightGreen:
		return "\033[92m"
	case LightYellow:
		return "\033[93m"
	case LightBlue:
		return "\033[94m"
	case LightMagenta:
		return "\033[95m"
	case LightCyan:
		return "\033[96m"
	}
	return ""
}

type timestamp string

// colorWrap wraps a string in a color
func colorWrap(c Color, m string) string {
	const Reset = "\033[0m"
	return c.Color() + m + Reset
}
