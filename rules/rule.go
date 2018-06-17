package rules

import (
	"bytes"
	"regexp"
	"text/scanner"

	"github.com/go-courier/validator/errors"
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
		return nil, errors.NewSyntaxError("%s | rule should be end but got `%s`", s.data[0:s.Pos().Offset], string(tok))
	}
	return rule, nil
}

var keychars = func() map[rune]bool {
	m := map[rune]bool{}
	for _, r := range []rune("@?=[](){}/<>,:") {
		m[r] = true
	}
	return m
}()

func (s *ruleScanner) scanLit() (string, error) {
	tok := s.Scan()
	if keychars[tok] {
		return "", errors.NewSyntaxError("%s | invalid literal token `%s`", s.data[0:s.Pos().Offset], string(tok))
	}
	return s.TokenText(), nil
}

func (s *ruleScanner) rule() (*Rule, error) {
	if firstToken := s.Next(); firstToken != '@' {
		return nil, errors.NewSyntaxError("%s | rule should start with `@` but got `%s`", s.data[0:s.Pos().Offset], string(firstToken))
	}
	startAt := s.Pos().Offset - 1

	name, err := s.scanLit()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.NewSyntaxError("%s | rule missing name", s.data[0:s.Pos().Offset])
	}
	rule := NewRule(name)

LOOP:
	for tok := s.Peek(); ; tok = s.Peek() {
		switch tok {
		case '?', '=':
			optional, defaultValue, err := s.optionalAndDefaultValue()
			if err != nil {
				return nil, err
			}
			rule.Optional = optional
			rule.DefaultValue = defaultValue
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
		case ' ':
			tok = s.Next()
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
		return nil, errors.NewSyntaxError("%s | parameters of rule should start with `<` but got `%s`", s.data[0:s.Pos().Offset], string(firstToken))
	}

	params := map[int]RuleNode{}
	paramCount := 1

	for tok := s.Peek(); tok != '>'; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, errors.NewSyntaxError("%s | parameters of rule should end with `>` but got `%s`", s.data[0:s.Pos().Offset], string(tok))
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
				return nil, errors.NewSyntaxError("%s | rule should be end but got `%s`", s.data[0:s.Pos().Offset], string(tok))
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
		return nil, firstToken, errors.NewSyntaxError("%s range of rule should start with `[` or `(` but got `%s`", s.data[0:s.Pos().Offset], string(firstToken))
	}

	ruleLits := map[int]*RuleLit{}
	litCount := 1

	for tok := s.Peek(); !(tok == ']' || tok == ')'); tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, tok, errors.NewSyntaxError("%s range of rule should end with `]` `)` but got `%s`", s.data[0:s.Pos().Offset], string(tok))
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
		return nil, errors.NewSyntaxError("%s | values of rule should start with `{` but got `%s`", s.data[0:s.Pos().Offset], string(firstToken))
	}

	ruleValues := map[int]*RuleLit{}
	valueCount := 1

	for tok := s.Peek(); tok != '}'; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, errors.NewSyntaxError("%s values of rule should end with `}`", s.data[0:s.Pos().Offset])
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
	s.Next()

	valueList := make([]*RuleLit, valueCount)
	for i := range valueList {
		if p, ok := ruleValues[i+1]; ok {
			valueList[i] = p
		} else {
			valueList[i] = NewRuleLit([]byte(""))
		}
	}
	return valueList, nil
}

func (s *ruleScanner) pattern() (*regexp.Regexp, error) {
	firstTok := s.Next()
	if firstTok != '/' {
		return nil, errors.NewSyntaxError("%s | pattern of rule should start with `/`", s.data[0:s.Pos().Offset])
	}

	b := &bytes.Buffer{}

	for tok := s.Peek(); tok != '/'; tok = s.Peek() {
		if tok == scanner.EOF {
			return nil, errors.NewSyntaxError("%s | pattern of rule should end with `/`", s.data[0:s.Pos().Offset])
		}
		if tok == '\\' {
			tok = s.Next()
			next := s.Next()
			// \/ -> /
			if next != '/' {
				b.WriteRune(tok)
			}
			b.WriteRune(next)
			continue
		}
		b.WriteRune(tok)
		tok = s.Next()
	}
	s.Next()

	return regexp.Compile(b.String())
}

func (s *ruleScanner) optionalAndDefaultValue() (bool, []byte, error) {
	firstTok := s.Next()
	if !(firstTok == '=' || firstTok == '?') {
		return false, nil, errors.NewSyntaxError("%s | optional or default value of rule should start with `?` or `=`", s.data[0:s.Pos().Offset])
	}

	b := &bytes.Buffer{}

	tok := s.Peek()
	for tok == ' ' {
		tok = s.Next()
	}

	if tok == '\'' {
		for tok = s.Peek(); tok != '\''; tok = s.Peek() {
			if tok == scanner.EOF {
				return true, nil, errors.NewSyntaxError("%s | default value of of rule should end with `'`", s.data[0:s.Pos().Offset])
			}
			if tok == '\\' {
				tok = s.Next()
				next := s.Next()
				// \' -> '
				if next != '\'' {
					b.WriteRune(tok)
				}
				b.WriteRune(next)
				continue
			}
			b.WriteRune(tok)
			tok = s.Next()
		}
		s.Next()
	} else if tok != scanner.EOF {
		b.WriteRune(tok)
		lit, err := s.scanLit()
		if err != nil {
			return false, nil, err
		}
		b.WriteString(lit)
	}

	defaultValue := b.Bytes()

	if firstTok == '=' && defaultValue == nil {
		return true, []byte{}, nil
	}

	return true, defaultValue, nil
}
