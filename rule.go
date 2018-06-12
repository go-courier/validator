package validator

import (
	"bytes"
	"regexp"
	"text/scanner"
)

func MustParseRuleString(rule string) *Rule {
	r, err := ParseRuleString(rule)
	if err != nil {
		panic(err)
	}
	return r
}

func ParseRuleString(rule string) (*Rule, error) {
	return ParseRule([]byte(rule))
}

func ParseRule(b []byte) (*Rule, error) {
	return newRuleScanner(b).rootRule()
}

func newRuleScanner(b []byte) *ruleScanner {
	s := &scanner.Scanner{}
	s.Init(bytes.NewReader(b))

	return &ruleScanner{
		data:    b,
		Scanner: s,
	}
}

type ruleScanner struct {
	data []byte
	*scanner.Scanner
}

func (s *ruleScanner) rootRule() (*Rule, error) {
	rule, err := s.rule()
	if err != nil {
		return nil, err
	}
	if tok := s.Scan(); tok != scanner.EOF {
		return nil, NewSyntaxErrorf("%s | rule should be end but got `%s`", s.data[0:s.Pos().Offset], string(tok))
	}
	return rule, nil
}

var keychars = func() map[rune]bool {
	m := map[rune]bool{}
	for _, r := range []rune("@[](){}/<>,:") {
		m[r] = true
	}
	return m
}()

func (s *ruleScanner) scanLit() (string, error) {
	tok := s.Scan()
	if keychars[tok] {
		return "", NewSyntaxErrorf("%s | invalid literal token `%s`", s.data[0:s.Pos().Offset], string(tok))
	}
	return s.TokenText(), nil
}

func (s *ruleScanner) rule() (*Rule, error) {
	if firstToken := s.Next(); firstToken != '@' {
		return nil, NewSyntaxErrorf("%s | rule should start with `@` but got `%s`", s.data[0:s.Pos().Offset], string(firstToken))
	}
	startAt := s.Pos().Offset - 1

	name, err := s.scanLit()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, NewSyntaxErrorf("%s | rule missing name", s.data[0:s.Pos().Offset])
	}
	rule := NewRule(name)

LOOP:
	for tok := s.Peek(); ; tok = s.Peek() {
		switch tok {
		case '<':
			params, err := s.params()
			if err != nil {
				return nil, err
			}
			rule.Params = params
		case '[', '(':
			ranges, endTok, err := s.ranges()
			if err != nil {
				return nil, err
			}
			rule.Range = ranges
			rule.ExclusiveLeft = tok == '('
			rule.ExclusiveRight = endTok == ')'
		case '{':
			values, err := s.values()
			if err != nil {
				return nil, err
			}
			rule.Values = values
		case '/':
			pattern, err := s.pattern()
			if err != nil {
				return nil, err
			}
			rule.Pattern = pattern
		default:
			break LOOP
		}
	}

	endAt := s.Pos().Offset
	rule.RAW = s.data[startAt:endAt]
	return rule, nil
}

func (s *ruleScanner) params() ([]RuleNode, error) {
	if firstToken := s.Next(); firstToken != '<' {
		return nil, NewSyntaxErrorf("%s | parameters of rule should start with `<` but got `%s`", s.data[0:s.Pos().Offset], string(firstToken))
	}

	params := map[int]RuleNode{}
	paramCount := 1

	for tok := s.Peek(); tok != '>'; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, NewSyntaxErrorf("%s | parameters of rule should end with `>` but got `%s`", s.data[0:s.Pos().Offset], string(tok))
		}
		switch tok {
		case ' ':
			tok = s.Next()
		case ',':
			tok = s.Next()
			paramCount++
		case '@':
			rule, err := s.rule()
			if err != nil {
				return nil, err
			}
			params[paramCount] = rule
		default:
			lit, err := s.scanLit()
			if err != nil {
				return nil, err
			}
			if ruleNode, ok := params[paramCount]; !ok {
				params[paramCount] = NewRuleLit([]byte(lit))
			} else if ruleLit, ok := ruleNode.(*RuleLit); ok {
				ruleLit.Append([]byte(lit))
			} else {
				return nil, NewSyntaxErrorf("%s | rule should be end but got `%s`", s.data[0:s.Pos().Offset], string(tok))
			}
		}
	}
	paramList := make([]RuleNode, paramCount)

	for i := range paramList {
		if p, ok := params[i+1]; ok {
			paramList[i] = p
		} else {
			paramList[i] = NewRuleLit([]byte(""))
		}
	}

	s.Next()
	return paramList, nil
}

func (s *ruleScanner) ranges() ([]*RuleLit, rune, error) {
	if firstToken := s.Next(); !(firstToken == '[' || firstToken == '(') {
		return nil, firstToken, NewSyntaxErrorf("%s range of rule should start with `[` or `(` but got `%s`", s.data[0:s.Pos().Offset], string(firstToken))
	}

	ruleLits := map[int]*RuleLit{}
	litCount := 1

	for tok := s.Peek(); !(tok == ']' || tok == ')'); tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, tok, NewSyntaxErrorf("%s range of rule should end with `]` `)` but got `%s`", s.data[0:s.Pos().Offset], string(tok))
		}
		switch tok {
		case ' ':
			tok = s.Next()
		case ',':
			tok = s.Next()
			litCount++
		default:
			lit, err := s.scanLit()
			if err != nil {
				return nil, tok, err
			}
			if ruleLit, ok := ruleLits[litCount]; !ok {
				ruleLits[litCount] = NewRuleLit([]byte(lit))
			} else {
				ruleLit.Append([]byte(lit))
			}
		}
	}

	litList := make([]*RuleLit, litCount)

	for i := range litList {
		if p, ok := ruleLits[i+1]; ok {
			litList[i] = p
		} else {
			litList[i] = NewRuleLit([]byte(""))
		}
	}

	return litList, s.Next(), nil
}

func (s *ruleScanner) values() ([]*RuleLit, error) {
	if firstToken := s.Next(); firstToken != '{' {
		return nil, NewSyntaxErrorf("%s | values of rule should start with `{` but got `%s`", s.data[0:s.Pos().Offset], string(firstToken))
	}

	ruleValues := map[int]*RuleLit{}
	valueCount := 1

	for tok := s.Peek(); tok != '}'; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, NewSyntaxErrorf("%s values of rule should end with `}`", s.data[0:s.Pos().Offset])
		}
		switch tok {
		case ' ':
			tok = s.Next()
		case ',':
			tok = s.Next()
			valueCount++
		default:
			lit, err := s.scanLit()
			if err != nil {
				return nil, err
			}
			if ruleLit, ok := ruleValues[valueCount]; !ok {
				ruleValues[valueCount] = NewRuleLit([]byte(lit))
			} else {
				ruleLit.Append([]byte(lit))
			}
		}
	}
	valueList := make([]*RuleLit, valueCount)
	for i := range valueList {
		if p, ok := ruleValues[i+1]; ok {
			valueList[i] = p
		} else {
			valueList[i] = NewRuleLit([]byte(""))
		}
	}

	s.Next()
	return valueList, nil
}

func (s *ruleScanner) pattern() (*regexp.Regexp, error) {
	firstTok := s.Next()
	if firstTok != '/' {
		return nil, NewSyntaxErrorf("%s | pattern of rule should start with `/`", s.data[0:s.Pos().Offset])
	}

	b := &bytes.Buffer{}

	for tok := s.Peek(); tok != '/'; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, NewSyntaxErrorf("%s | pattern of rule should end with `/`", s.data[0:s.Pos().Offset])
		}
		switch tok {
		case '\\':
			tok = s.Next()
			next := s.Next()
			if next != '/' {
				b.WriteRune(tok)
			}
			b.WriteRune(next)
		default:
			b.WriteRune(tok)
			tok = s.Next()
		}
	}
	s.Next()

	return regexp.Compile(b.String())
}

type RuleNode interface {
	node()
	Bytes() []byte
}

func NewRule(name string) *Rule {
	return &Rule{
		Name: name,
	}
}

// @name
// @name<param1,param2,...>
// @name[from, to)
// @name<param1,param2,...>[from:to]
// @name<param1,param2,...>[length]
// @name<param1,param3,...>{Value1,Value2,Value3}
// @name<param1,param2,...>/\w+/
type Rule struct {
	RAW []byte

	Name   string
	Params []RuleNode

	Range          []*RuleLit
	ExclusiveLeft  bool
	ExclusiveRight bool

	Values []*RuleLit

	Pattern *regexp.Regexp

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
			buf.Write(p.Bytes())
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

func Unslash(src []byte) ([]byte, error) {
	n := len(src)
	if n < 2 {
		return src, NewSyntaxErrorf("%s", src)
	}
	quote := src[0]
	if quote != '/' || quote != src[n-1] {
		return src, NewSyntaxErrorf("%s", src)
	}

	src = src[1 : n-1]
	n = len(src)

	finalData := make([]byte, 0)
	for i, b := range src {
		if b == '\\' && i != n-1 && src[i+1] == '/' {
			continue
		}
		finalData = append(finalData, b)
	}
	return finalData, nil
}

func Slash(data []byte) []byte {
	buf := &bytes.Buffer{}
	buf.WriteRune('/')
	for _, b := range data {
		if b == '/' {
			buf.WriteRune('\\')
		}
		buf.WriteByte(b)
	}
	buf.WriteRune('/')
	return buf.Bytes()
}
