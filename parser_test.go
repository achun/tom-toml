package toml

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func scannerPanic(t *testing.T, s Scanner, fn string, msg string) {
	defer func() {
		err := recover()
		str := fmt.Sprint(err)
		if msg != str {
			t.Fatal("expected panic:", msg, ",but got:", str)
		}
	}()
	if fn == "get" {
		s.Get()
	} else {
		s.Next()
	}
}

type parseTest struct {
	parse
	outs           []string
	skipWhitespace bool
	nopanic        bool
}

func (p *parseTest) Out(token Token, str string) (err error) {
	if p.skipWhitespace && tokenWhitespace == token || token == tokenNewLine || token == tokenEOF {
		return
	}
	if token == tokenError {
		if p.nopanic {
			return
		}
		panic(str)
	}
	p.outs = append(p.outs, tokensName[token], strings.TrimSpace(str))
	return
}
func assertBadParse(t *testing.T, str string, err string) {
	buf := NewScanner([]byte(str))
	p := &parseTest{parse{buf, 0, nil, nil}, []string{}, true, true}
	p.Handler(p.Out)
	p.run()
	if p.err == nil || p.err.Error() != err {
		t.Fatal("expected error:", err, ",but got:", p.err)
	}
}
func assertParse(t *testing.T, str string, outs ...string) {
	buf := NewScanner([]byte(str))
	p := &parseTest{parse{buf, 0, nil, nil}, []string{}, true, false}
	p.Handler(p.Out)
	p.run()
	i := 0
	for i < len(outs) && i < len(p.outs) {
		if outs[i] != p.outs[i] || outs[i+1] != p.outs[i+1] {
			fmt.Printf("%#v", p.outs)
			t.Fatal("expected:", outs[i], outs[i+1], ",but got:", p.outs[i], p.outs[i+1])
		}
		i += 2
	}
	if i == 0 {
		if len(p.outs) == len(outs) {
			return
		}
		if len(p.outs) == 0 {
			t.Fatal("expected not found:", outs[0], outs[1])
		}
		t.Fatal("unexpected:", p.outs[0], p.outs[1])
	}
	if len(outs) < len(p.outs) {
		fmt.Printf("%#v", p.outs)
		t.Fatal("unexpected:", p.outs[i], p.outs[i+1])
	}
	if len(outs) > len(p.outs) {
		fmt.Printf("%#v", p.outs)
		t.Fatal("expected not found:", outs[i], outs[i+1])
	}
}

func TestScanner(t *testing.T) {
	s := NewScanner([]byte("0123456789"))
	scannerPanic(t, s, "get", "expected Next() first")
	scannerPanic(t, s, "next", "<nil>")
	scannerPanic(t, s, "get", "<nil>")
	scannerPanic(t, s, "get", "expected Next() first")
	for i := 0; i < 10; i++ {
		scannerPanic(t, s, "next", "<nil>")
	}
	scannerPanic(t, s, "next", "EOF")
	scannerPanic(t, s, "get", "<nil>")
	scannerPanic(t, s, "get", "buffer null")
	scannerPanic(t, s, "next", "EOF")

	s = NewScanner([]byte("0123456789"))
	s.Next()
	s.Next()
	s.Next()
	assertEqual(t, s.Get(), "01")
	assertEqual(t, s.Get(), "2")

	scannerPanic(t, s, "get", "expected Next() first")
	str := string(s.Next())
	assertEqual(t, str, "3")
	assertEqual(t, s.Get(), "3")
	s.Next()
	str = string(s.Next())
	assertEqual(t, str, "5")
	assertEqual(t, s.Get(), "4")
	str = string(s.Next())
	assertEqual(t, str, "5")

	s = NewScanner([]byte(""))
	str = string(s.Next())
	assertEqual(t, str, string(EOF))

	s = NewScanner([]byte(" "))
	str = string(s.Next())
	assertEqual(t, str, " ")
	assertEqual(t, s.Get(), " ")

	str = string(s.Next())
	assertEqual(t, str, string(EOF))
	assertEqual(t, s.Get(), "")
}

func TestEmpty(t *testing.T) {
	assertParse(t, ``)
	assertParse(t, ` `)
	assertParse(t, `
		#`, "Comment", `#`)
}

func TestBadParse(t *testing.T) {
	assertBadParse(t, `key = [1,"ent"]`, "incomplete Array")
	assertBadParse(t, `key = [`, "incomplete Array")
	assertBadParse(t, `key`, "invalid Key")
	assertBadParse(t, `key # comment`, "incomplete Equal")
	assertBadParse(t, `[table name]`, "invalid Table")
	assertBadParse(t, `key = # comment`, "incomplete Value")
}

func TestParserOK(t *testing.T) {
	assertParse(t,
		`
# comment 1
# comment 2
[table#] # comment 3
[[arrayoftable]]
     # comment 4
key  =  -111# comment 5
float= 123.22
int = 11234
datetime = 2012-01-02T13:11:14Z
string = "is string \n newline"
intger = [
1,# comment 6
2# comment 7
]
strings=[# comment 8
"a",
"b",
]# comment 9
# comment 10`,
		"Comment", `# comment 1`,
		"Comment", `# comment 2`,
		"Table", `[table#]`, "Comment", `# comment 3`,
		"ArrayOfTables", `[[arrayoftable]]`,
		"Comment", `# comment 4`,
		"Key", `key`, "Equal", `=`, "Integer", `-111`, "Comment", `# comment 5`,
		"Key", `float`, "Equal", `=`, "Float", `123.22`,
		"Key", `int`, "Equal", `=`, "Integer", `11234`,
		"Key", `datetime`, "Equal", `=`, "Datetime", `2012-01-02T13:11:14Z`,
		"Key", `string`, "Equal", `=`, "String", `"is string \n newline"`,
		"Key", `intger`, "Equal", `=`,
		"ArrayLeftBrack", `[`,
		"Integer", `1`, "Comma", `,`, "Comment", `# comment 6`, "Integer", `2`, "Comment", `# comment 7`,
		"ArrayRightBrack", `]`,
		"Key", `strings`, "Equal", `=`,
		"ArrayLeftBrack", `[`, "Comment", `# comment 8`,
		"String", `"a"`, "Comma", `,`, "String", `"b"`, "Comma", `,`,
		"ArrayRightBrack", `]`, "Comment", `# comment 9`, "Comment", `# comment 10`,
	)
}

func TestFile(t *testing.T) {
	buf, err := ioutil.ReadFile("tests/example.toml")
	if err != nil {
		t.Fatal(err)
	}
	assertParse(t, string(buf),
		"Comment", `# This is a TOML document. Boom.`,
		"Key", `title`, "Equal", `=`, "String", `"TOML Example"`,
		"Table", `[owner]`,
		"Key", `name`, "Equal", `=`, "String", `"Tom Preston-Werner"`,
		"Key", `organization`, "Equal", `=`, "String", `"GitHub"`,
		"Key", `bio`, "Equal", `=`, "String", `"GitHub Cofounder & CEO\nLikes tater tots and beer."`,
		"Key", `dob`, "Equal", `=`, "Datetime", `1979-05-27T07:32:00Z`, "Comment", `# First class dates? Why not?`,
		"Table", `[database]`,
		"Key", `server`, "Equal", `=`, "String", `"192.168.1.1"`,
		"Key", `ports`, "Equal", `=`,

		"ArrayLeftBrack", `[`,
		"Integer", `8001`, "Comma", `,`, "Integer", `8001`, "Comma", `,`, "Integer", `8002`,
		"ArrayRightBrack", `]`,

		"Key", `connection_max`, "Equal", `=`, "Integer", `5000`,
		"Key", `enabled`, "Equal", `=`, "Boolean", `true`,
		"Table", `[servers]`,
		"Comment", `# You can indent as you please. Tabs or spaces. TOML don't care.`,
		"Table", `[servers.alpha]`,
		"Key", `ip`, "Equal", `=`, "String", `"10.0.0.1"`,
		"Key", `dc`, "Equal", `=`, "String", `"eqdc10"`,
		"Table", `[servers.beta]`,
		"Key", `ip`, "Equal", `=`, "String", `"10.0.0.2"`,
		"Key", `dc`, "Equal", `=`, "String", `"eqdc10"`,
		"Key", `country`, "Equal", `=`, "String", `"中国"`, "Comment", `# This should be parsed as UTF-8`,
		"Table", `[clients]`,
		"Key", `data`, "Equal", `=`,

		"ArrayLeftBrack", `[`,

		"ArrayLeftBrack", `[`,
		"String", `"gamma"`, "Comma", `,`, "String", `"delta"`,
		"ArrayRightBrack", `]`,
		"Comma", `,`,
		"ArrayLeftBrack", `[`,
		"Integer", `1`, "Comma", `,`, "Integer", `2`,
		"ArrayRightBrack", `]`,

		"ArrayRightBrack", `]`,

		"Comment", `# just an update to make sure parsers support it`,
		"Comment", `# Line breaks are OK when inside arrays`,
		"Key", `hosts`, "Equal", `=`,

		"ArrayLeftBrack", `[`,
		"String", `"alpha"`, "Comma", `,`, "String", `"omega"`,
		"ArrayRightBrack", `]`,

		"Comment", `# Products`,
		"ArrayOfTables", `[[products]]`,
		"Key", `name`, "Equal", `=`, "String", `"Hammer"`,
		"Key", `sku`, "Equal", `=`, "Integer", `738594937`,
		"ArrayOfTables", `[[products]]`,
		"Key", `name`, "Equal", `=`, "String", `"Nail"`,
		"Key", `sku`, "Equal", `=`, "Integer", `284758393`,
		"Key", `color`, "Equal", `=`, "String", `"gray"`,
	)
}
