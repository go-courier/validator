package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRule(t *testing.T) {
	cases := [][]string{
		// simple
		{`@email`, `@email`},

		// with parameters
		{`@map<@email,         @url>`, `@map<@email,@url>`},
		{`@map<@string,>`, `@map<@string,>`},
		{`@map<,@string>`, `@map<,@string>`},
		{`@float32<10,6>`, `@float32<10,6>`},
		{`@float32<10,-1>`, `@float32<10,-1>`},
		{`@slice<@string>`, `@slice<@string>`},

		// with range
		{`@slice[0,   10]`, `@slice[0,10]`},
		{`@array[10]`, `@array[10]`},
		{`@string[0,)`, `@string[0,)`},
		{`@string[0,)`, `@string[0,)`},
		{`@int(0,)`, `@int(0,)`},
		{`@int(,1)`, `@int(,1)`},
		{`@float32(1.10,)`, `@float32(1.10,)`},

		// with values
		{`@string{A, B,    C}`, `@string{A,B,C}`},
		{`@string{, B,    C}`, `@string{,B,C}`},
		{`@uint{%2}`, `@uint{%2}`},

		// with regexp
		{`@string/\w+/`, `@string/\w+/`},
		{`@string/\w+     $/`, `@string/\w+     $/`},
		{`@string/\w+\/abc/`, `@string/\w+\/abc/`},

		// composes
		{`@map<,@string[1,]>`, `@map<,@string[1,]>`},
		{`@map<@string,>[1,2]`, `@map<@string,>[1,2]`},
	}

	for i := range cases {
		c := cases[i]
		t.Run("rule:"+c[0], func(t *testing.T) {
			r, err := ParseRuleString(c[0])
			assert.NoError(t, err)
			assert.Equal(t, c[1], string(r.Bytes()))
		})
	}
}

func TestParseRuleFailed(t *testing.T) {
	cases := []string{
		`@`,
		`@unsupportted-name`,
		`@name<`,
		`@name[`,
		`@name(`,
		`@name{`,
		`@name/`,
		`@name)`,
		`@name<@sub[>`,
		`@name</>`,
		`@/`,
	}

	for _, c := range cases {
		_, err := ParseRuleString(c)
		t.Logf("%s %s", c, err)
	}
}

func TestSlashUnslash(t *testing.T) {
	cases := [][]string{
		{`/\w+\/test/`, `\w+/test`},
		{`/a/`, `a`},
		{`/abc/`, `abc`},
		{`/☺/`, `☺`},
		{`/\xFF/`, `\xFF`},
		{`/\377/`, `\377`},
		{`/\u1234/`, `\u1234`},
		{`/\U00010111/`, `\U00010111`},
		{`/\U0001011111/`, `\U0001011111`},
		{`/\a\b\f\n\r\t\v\\\"/`, `\a\b\f\n\r\t\v\\\"`},
		{`/\//`, `/`},
	}

	for i := range cases {
		c := cases[i]
		t.Run("unslash:"+c[0], func(t *testing.T) {
			r, err := Unslash([]byte(c[0]))
			assert.NoError(t, err)
			assert.Equal(t, string(r), c[1])
		})
		t.Run("slash:"+c[1], func(t *testing.T) {
			assert.Equal(t, string(Slash([]byte(c[1]))), c[0])
		})
	}

	casesForFailed := [][]string{
		{`/`, ``},
		{`/adfadf`, ``},
	}

	for i := range casesForFailed {
		c := casesForFailed[i]
		t.Run("unslash:"+c[0], func(t *testing.T) {
			_, err := Unslash([]byte(c[0]))
			assert.Error(t, err)
		})
	}

}
