package term

import (
	"bitbucket.org/johnsto/go-httpud/highlight"
	"github.com/fatih/color"
)

type Output struct {
	Colors map[highlight.TokenType]*color.Color
}

func NewOutput() *Output {
	return &Output{
		Colors: map[highlight.TokenType]*color.Color{
			highlight.Error:       color.New(color.FgRed, color.Bold),
			highlight.Comment:     color.New(color.FgWhite, color.Faint),
			highlight.Text:        color.New(color.FgHiWhite),
			highlight.Number:      color.New(color.FgHiMagenta),
			highlight.String:      color.New(color.FgHiGreen),
			highlight.Attribute:   color.New(color.FgGreen, color.Bold),
			highlight.Assignment:  color.New(color.FgYellow, color.Faint),
			highlight.Operator:    color.New(color.FgGreen),
			highlight.Punctuation: color.New(color.FgYellow),
			highlight.Constant:    color.New(color.FgBlue, color.Bold),
			highlight.Entity:      color.New(color.FgHiYellow),
			highlight.Whitespace:  color.New(color.FgWhite),
		},
	}
}

func (o *Output) Emit(t highlight.Token) (int, error) {
	c := o.Colors[t.Type]
	if c == nil {
		return o.Colors[highlight.Error].Printf("%s", t.Value)
	} else {
		return c.Printf("%s", t.Value)
	}
}