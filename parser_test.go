package toml

import (
	"github.com/achun/testing-want"
	"io/ioutil"
	"testing"
)

func assertBadParse(t *testing.T, src string, msg string) {
	scan := NewScanner([]byte(src))
	p := &parse{Scanner: scan}
	wt := want.Want{t, 8}
	got := 0
	p.Handler(func(token Token, str string) (err error) {
		if token == tokenError {
			got++
			wt.Equal(str, msg)
		}
		return
	})
	p.Run()
	wt.Skip = 3
	wt.True(got == 1, "want an error: ", msg, " but got: ", got)
}

// outs =["token","value"]
func assertParse(t *testing.T, str string, outs ...string) {
	scan := NewScanner([]byte(str))
	p := &parse{Scanner: scan}
	i := 0
	l := len(outs)
	wt := want.Want{t, 8}
	p.Handler(func(token Token, str string) (err error) {
		if tokenWhitespace == token || token == tokenNewLine || token == tokenEOF {
			return
		}
		if i >= l {
			wt.True(false, "more outs:", i, ">=", l)
		}
		wt.Equal(token.String()+" "+str, outs[i])
		i++
		return
	})
	p.Run()
}

func TestScanner(t *testing.T) {
	wt := want.T(t)
	s := NewScanner([]byte("0123456789"))
	wt.Equal(s.Fetch(true), "0")

	r := s.Rune()
	for i := 0; i < 10; i++ {
		wt.Equal(r, rune('0'+i))
		r = s.Next()
	}

	wt.Equal(s.Next(), rune(EOF))
	wt.True(s.Eof())
	wt.Equal(s.Fetch(true), "123456789")
	wt.Equal(s.Next(), rune(EOF))
	wt.Equal(s.Next(), rune(EOF))
	wt.Equal(s.Fetch(true), "")
}

func TestEmpty(t *testing.T) {
	assertParse(t, ``)
	assertParse(t, ` `)
	assertParse(t, ` `)
	assertParse(t, `
	   	

		 `)
}

func TestToken(t *testing.T) {
	const (
		al = `ArrayLeftBrack [`
		ar = `ArrayRightBrack ]`
		eq = `Equal =`
		i  = `Integer`
		ca = `Comma ,`
	)

	assertParse(t, `string = "is string \n newline"`, `Key string`, eq, `String "is string \n newline"`)
	assertParse(t, `#`, `Comment #`)
	assertParse(t, `#

		# 1  
		`, `Comment #`, `Comment # 1`)
	assertParse(t, `key = 1`, `Key key`, eq, `Integer 1`)
	assertParse(t, `key = []`, `Key key`, eq, al, ar)
	assertParse(t, `ia = [1 , 2]`, `Key ia`, eq, al, `Integer 1`, ca, `Integer 2`, ar)

	assertBadParse(t, `key`, "invalid Key")
	assertBadParse(t, `key 1`, "incomplete Equal")
	assertBadParse(t, `ke y = name`, `incomplete Equal`)
	assertBadParse(t, `key # comment`, "incomplete Equal")
	assertBadParse(t, `key = name`, `incomplete Value`)
	assertBadParse(t, `key = [1,ent]`, "incomplete Array")
	assertBadParse(t, `key = [1,"ent"]`, "incomplete Array")
	assertBadParse(t, `key = [`, "incomplete Array")
	assertBadParse(t, `key = # comment`, "incomplete Value")

	assertBadParse(t, `[]`, "invalid TableName")
	assertBadParse(t, `[table ]`, "invalid TableName")
	assertBadParse(t, `[tab le]`, "invalid TableName")
	assertBadParse(t, `[ table]`, "invalid TableName")
	assertBadParse(t, `[	table]`, "invalid TableName")
	assertBadParse(t, `[[tables]`, "invalid ArrayOfTables")
	assertBadParse(t, `[[ tables]]`, "invalid ArrayOfTables")
	assertBadParse(t, `[[tab les]]`, "invalid ArrayOfTables")
	assertBadParse(t, `[[tables ]]`, "invalid ArrayOfTables")
	assertBadParse(t, `[[	tables]]`, "invalid ArrayOfTables")
	assertBadParse(t, `[[tab	les]]`, "invalid ArrayOfTables")

	assertParse(t, `[name]`, "TableName [name]")
	assertParse(t, `[[name]]`, "ArrayOfTables [[name]]")

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
		`Comment # comment 1`,
		`Comment # comment 2`,
		`TableName [table#]`, `Comment # comment 3`,
		`ArrayOfTables [[arrayoftable]]`,
		`Comment # comment 4`,
		`Key key`, `Equal =`, `Integer -111`, `Comment # comment 5`,
		`Key float`, `Equal =`, `Float 123.22`,
		`Key int`, `Equal =`, `Integer 11234`,
		`Key datetime`, `Equal =`, `Datetime 2012-01-02T13:11:14Z`,
		`Key string`, `Equal =`, `String "is string \n newline"`,
		`Key intger`, `Equal =`,
		`ArrayLeftBrack [`,
		`Integer 1`, `Comma ,`, `Comment # comment 6`, `Integer 2`, `Comment # comment 7`,
		`ArrayRightBrack ]`,
		`Key strings`, `Equal =`,
		`ArrayLeftBrack [`, `Comment # comment 8`,
		`String "a"`, `Comma ,`, `String "b"`, `Comma ,`,
		`ArrayRightBrack ]`, `Comment # comment 9`, `Comment # comment 10`,
	)
}

func TestFile(t *testing.T) {
	buf, err := ioutil.ReadFile("tests/example.toml")
	if err != nil {
		t.Fatal(err)
	}
	assertParse(t, string(buf),
		`Comment # This is a TOML document. Boom.`,
		`Key title`, `Equal =`, `String "TOML Example"`,
		`TableName [owner]`,
		`Comment # owner information`,
		`Key name`, `Equal =`, `String "Tom Preston-Werner"`,
		`Key organization`, `Equal =`, `String "GitHub"`,
		`Key bio`, `Equal =`, `String "GitHub Cofounder & CEO\nLikes tater tots and beer."`,
		`Key dob`, `Equal =`, `Datetime 1979-05-27T07:32:00Z`, `Comment # First class dates? Why not?`,
		`TableName [database]`,
		`Key server`, `Equal =`, `String "192.168.1.1"`,
		`Key ports`, `Equal =`,

		`ArrayLeftBrack [`,
		`Integer 8001`, `Comma ,`, `Integer 8001`, `Comma ,`, `Integer 8002`,
		`ArrayRightBrack ]`,

		`Key connection_max`, `Equal =`, `Integer 5000`,
		`Key enabled`, `Equal =`, `Boolean true`,
		`TableName [servers]`,
		`Comment # You can indent as you please. Tabs or spaces. TOML don't care.`,
		`TableName [servers.alpha]`,
		`Key ip`, `Equal =`, `String "10.0.0.1"`,
		`Key dc`, `Equal =`, `String "eqdc10"`,
		`TableName [servers.beta]`,
		`Key ip`, `Equal =`, `String "10.0.0.2"`,
		`Key dc`, `Equal =`, `String "eqdc10"`,
		`Key country`, `Equal =`, `String "中国"`, `Comment # This should be parsed as UTF-8`,
		`TableName [clients]`,
		`Key data`, `Equal =`,

		`ArrayLeftBrack [`,

		`ArrayLeftBrack [`,
		`String "gamma"`, `Comma ,`, `String "delta"`,
		`ArrayRightBrack ]`,
		`Comma ,`,
		`ArrayLeftBrack [`,
		`Integer 1`, `Comma ,`, `Integer 2`,
		`ArrayRightBrack ]`,

		`ArrayRightBrack ]`,

		`Comment # just an update to make sure parsers support it`,
		`Comment # Line breaks are OK when inside arrays`,
		`Key hosts`, `Equal =`,

		`ArrayLeftBrack [`,
		`String "alpha"`, `Comma ,`, `String "omega"`,
		`ArrayRightBrack ]`,

		`Comment # Products`,
		`ArrayOfTables [[products]]`,
		`Key name`, `Equal =`, `String "Hammer"`,
		`Key sku`, `Equal =`, `Integer 738594937`,
		`ArrayOfTables [[products]]`,
		`Key name`, `Equal =`, `String "Nail"`,
		`Key sku`, `Equal =`, `Integer 284758393`,
		`Key color`, `Equal =`, `String "gray"`,

		`Comment # nested`,
		`ArrayOfTables [[fruit]]`,
		`Key name`, `Equal =`, `String "apple"`,

		`TableName [fruit.physical]`,
		`Key color`, `Equal =`, `String "red"`,
		`Key shape`, `Equal =`, `String "round"`,

		`ArrayOfTables [[fruit.variety]]`,
		`Key name`, `Equal =`, `String "red delicious"`,

		`ArrayOfTables [[fruit.variety]]`,
		`Key name`, `Equal =`, `String "granny smith"`,

		`ArrayOfTables [[fruit]]`,
		`Key name`, `Equal =`, `String "banana"`,

		`ArrayOfTables [[fruit.variety]]`,
		`Key name`, `Equal =`, `String "plantain"`,
	)
}
