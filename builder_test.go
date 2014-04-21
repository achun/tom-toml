package toml

import (
	"github.com/achun/testing-want"
	"io/ioutil"
	"testing"
)

func TestBuilder(t *testing.T) {
	source, err := ioutil.ReadFile("tests/example.toml")
	wt := want.T(t)
	wt.Nil(err)
	testBuilder(wt, source, "TestBuilder")
}

func testBuilder(wt want.Want, source []byte, from string) {
	var err error

	p := &parse{Scanner: NewScanner(source)}

	tb := NewBuilder(nil)
	p.Handler(
		func(token Token, str string) error {
			tb, err = tb.Token(token, str)
			return err
		})

	p.Run()

	if err != nil {
		line, col, str := p.Scanner.LastLine()
		str = want.String("line:", line, " col:", col, "\n", str)
		wt.Nil(err, "from: ", from, str, fetchBuilderInfo(tb))
	}
}

func fetchBuilderInfo(tb TomlBuilder) string {
	str := "\nToken: " + tb.root.token.String() + " ," + tb.token.String()
	if tb.it != nil {
		str += want.String("Item: ", tb.it.kind.String(), " Key:", tb.it.key, " Len:", tb.it.Len())
	}

	if tb.iv != nil {
		str += want.String("Value: ", tb.iv.kind.String(), " Key: ", tb.iv.key, " Len: ", tb.iv.Len())
	}

	for key, it := range tb.Toml() {
		if it.kind != InvalidKind || key == iD {
			continue
		}
		str += want.String("Id: ", it.idx, " ,Path: ", key, " ,IsNil: ", it.v == nil)
	}
	return str
}
