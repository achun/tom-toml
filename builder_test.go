package toml

import (
	"github.com/achun/testing-want"
	"io/ioutil"
	"testing"
)

var skipTest bool = false

func TestComments(t *testing.T) {
	if skipTest {
		return
	}

	wt := want.T(t)
	testBuilder(wt, []byte(`
		[[name]] # 世界,
			id = 1
		[[name]] # Template, 尝试独立的 _layouts 
			id =2
		`),
		`TestComments`)
}

func TestReadFile(t *testing.T) {
	if skipTest {
		return
	}
	source, err := ioutil.ReadFile("tests/example.toml")
	wt := want.T(t)
	wt.Nil(err)
	testBuilder(wt, source, "TestReadFile")
}

func testBuilder(wt want.Want, source []byte, from string) {

	p := &parse{Scanner: NewScanner(source)}

	tb := newBuilder(nil)
	p.Handler(
		func(token Token, str string) (err error) {
			tb, err = tb.Token(token, str)
			return
		})

	p.Run()
	if p.err != nil {
		line, col, str := p.Scanner.LastLine()
		str = want.String("line:", line, " col:", col, "\n", str)
		wt.Nil(p.err, "from: ", from, str, FetchBuilderInfo(tb))
	}
}

func FetchBuilderInfo(tb tomlBuilder) string {
	str := "\nToken: " + tb.root.token.String() + " ," + tb.token.String()
	if tb.it != nil {
		str += want.String("Item: ", tb.it.kind.String(), " TableName: ", tb.tableName, " Len: ", tb.it.Len())
	}

	if tb.iv != nil {
		str += want.String("Value: ", tb.iv.kind.String(), " TableName: ", tb.tableName, " Len: ", tb.iv.Len())
	}

	for key, it := range tb.Toml() {
		if it.kind != InvalidKind || key == iD {
			continue
		}
		str += want.String("Id: ", it.idx, " ,Path: ", key, " ,IsNil: ", it.v == nil)
	}
	return str
}
