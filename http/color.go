package main

import "github.com/fatih/color"

var (
	ColorNormal      = NewColor(color.Reset)
	ColorReserved    = NewColor(color.FgHiCyan)
	ColorOperator    = NewColor(color.FgHiRed)
	ColorAttribute   = NewColor(color.FgHiGreen)
	ColorNumber      = NewColor(color.FgHiBlue)
	ColorText        = NewColor(color.FgWhite)
	ColorPunctuation = NewColor(color.FgWhite, color.Faint)
	ColorComment     = NewColor(color.FgHiBlack)
	ColorStatus      = NewColor(color.FgHiGreen)
	ColorString      = NewColor(color.FgHiYellow)
	ColorError       = NewColor(color.FgWhite, color.BgRed)
)

type Color []color.Attribute

// NewColor creates a new Color with a combination of the given attributes.
func NewColor(as ...color.Attribute) Color {
	return Color(as)
}

// Color returns a color.Color value.
func (cs Color) Color() *color.Color {
	cc := color.New()
	for _, c := range cs {
		cc = cc.Add(c)
	}
	return cc
}

// Print emits the given arguments to the output, returning the number of
// characters written and any error.
func (cs Color) Print(a ...interface{}) (int, error) {
	return cs.Color().Print(a...)
}

// Print formats and emits the given arguments to the output, returning the
// number of characters written and any error.
func (cs Color) Printf(fmt string, a ...interface{}) (int, error) {
	return cs.Color().Printf(fmt, a...)
}
