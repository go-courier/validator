package rules

import (
	"bytes"
	"regexp"
)

type RuleNode interface {
	node()
	Bytes() []byte
}

func NewRule(name string) *Rule {
	return &Rule{
		Name:  name,
		Metas: Metas{},
	}
}

type Rule struct {
	RAW []byte

	Name   string
	Params []RuleNode

	Range          []*RuleLit
	ExclusiveLeft  bool
	ExclusiveRight bool

	Values []*RuleLit

	Pattern *regexp.Regexp

	Optional     bool
	DefaultValue []byte

	Metas Metas

	RuleNode
}

func (r *Rule) Bytes() []byte {
	if r == nil {
		return nil
	}

	buf := &bytes.Buffer{}
	buf.WriteByte('@')
	buf.WriteString(r.Name)

	if len(r.Params) > 0 {
		buf.WriteByte('<')
		for i, p := range r.Params {
			if i > 0 {
				buf.WriteByte(',')
			}
			if p != nil {
				buf.Write(p.Bytes())
			}
		}
		buf.WriteByte('>')
	}

	if len(r.Range) > 0 {
		if r.ExclusiveLeft {
			buf.WriteRune('(')
		} else {
			buf.WriteRune('[')
		}
		for i, p := range r.Range {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(p.Bytes())
		}
		if r.ExclusiveRight {
			buf.WriteRune(')')
		} else {
			buf.WriteRune(']')
		}
	}

	if len(r.Values) > 0 {
		buf.WriteByte('{')
		for i, p := range r.Values {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(p.Bytes())
		}
		buf.WriteByte('}')
	}

	if r.Pattern != nil {
		buf.Write(Slash([]byte(r.Pattern.String())))
	}

	if r.Optional {
		if r.DefaultValue != nil {
			buf.WriteByte(' ')
			buf.WriteByte('=')
			buf.WriteByte(' ')

			buf.Write(SingleQuote(r.DefaultValue))
		} else {
			buf.WriteByte('?')
		}
	}

	return buf.Bytes()
}

func NewRuleLit(lit []byte) *RuleLit {
	return &RuleLit{
		Lit: lit,
	}
}

type RuleLit struct {
	Lit []byte
	RuleNode
}

func (lit *RuleLit) Append(b []byte) {
	lit.Lit = append(lit.Lit, b...)
}

func (lit *RuleLit) Bytes() []byte {
	if lit == nil {
		return nil
	}
	return lit.Lit
}
