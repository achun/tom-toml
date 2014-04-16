package toml

import (
	"fmt"
	"github.com/achun/testing-want"
	"testing"
	"time"
)

func TestTomlFile(t *testing.T) {
	wt := want.T(t)
	tm, err := LoadFile("tests/example.toml")

	wt.Nil(err)
	wt.Equal(tm["servers.alpha.ip"].String(), "10.0.0.1")
	wt.Equal(tm["servers.beta.ip"].String(), "10.0.0.2")

	tm2, err := Parse([]byte(tm.String()))
	wt.Nil(err)

	wt.Equal(tm["servers.alpha.ip"].String(), tm2["servers.alpha.ip"].String())
	wt.Equal(tm["servers.alpha.dc"].String(), tm2["servers.alpha.dc"].String())
}

func TestTomlEmpty(t *testing.T) {
	wt := want.T(t)
	tm, err := Parse([]byte(``))

	wt.Nil(err)
	wt.Equal(len(tm), 0)
}

func TestToml(t *testing.T) {
	tm := testToml(t, testData[0], testData[1:])
	testToml(t, tm.String(), testData[1:])

}
func testToml(t *testing.T, source string, as []string) Toml {
	wt := want.T(t)
	tm := assertToml(t, source, as)

	ia := tm[""]
	wt.True(ia != nil, "LastComments invalid")
	wt.Equal(ia.MultiComments, "# LastComments", "LastComments")

	ia = tm["multiline.ia"]
	wt.True(ia != nil, "multiline.ia.Value invalid")
	wt.Equal(ia.EolComment, "# 10.2", "multiline.ia.EolComment")

	wt.Equal(ia.Index(0).MultiComments, "# 10.1", "multiline.ia·Index[0].MultiComments")
	wt.Equal(ia.Index(1).EolComment, "# 10", "multiline.ia·Index[1].EolComment")

	ia = tm["arrayarray.2.ia"]
	wt.True(ia != nil, "arrayarray.2.ia invalid")

	wt.Equal(ia.Index(0).String(), "[1,2]", "arrayarray.2.ia·Index(0)")
	wt.Equal(ia.Index(0).EolComment, "# 11", "arrayarray.2.ia·Index(0).EolComment")

	aot := tm["array.tables"]
	wt.True(aot != nil, "array.tables invalid")

	ts := aot.Table(0)
	wt.True(ts != nil, "array.tables.Index(0) invalid")
	wt.Equal(ts["ia"].Int(), int64(1), "array.tables.Index(0).ia")
	wt.Equal(ts["aa"].String(), "name", "array.tables.Index(0).aa")
	ts = aot.Table(1)
	wt.Equal(ts["ia"].Int(), int64(2), "array.tables.Index(1).ia")
	wt.Equal(ts["aa"].String(), "jack", "array.tables.Index(1).aa")
	return tm
}

func assertToml(t *testing.T, source string, as []string) (tm Toml) {
	wt := want.T(t)
	tm, err := Parse([]byte(source))
	wt.Nil(err)

	for i := 0; i < len(as); i += 3 {
		path := as[i]
		fn := as[i+1]
		re := as[i+2]
		it := tm[path]

		if it != nil && it.kind == ArrayOfTables {
			ts := it.Table(0)
			if ts == nil {
				t.Fatal(path, it)
			}

			it = &Item{*ts[as[i+3]]}
			i++
		}
		s := ""
		if !it.IsValid() {
			t.Fatalf("\n%v %v :\n%#v\n%#v\n", path, fn, re, it)
		}
		switch fn {
		case "fc":
			s = it.MultiComments
		case "ec":
			s = it.EolComment
		case "i":
			s = fmt.Sprint(it.Int())
		case "f":
			s = fmt.Sprint(it.Float())
		case "b":
			s = fmt.Sprint(it.Boolean())
		case "d":
			s = fmt.Sprint(it.Datetime())
		case "s":
			s = fmt.Sprint(it.String())
		case "k":
			s = fmt.Sprint(kindsName[it.kind])
		case "ts":
			s = fmt.Sprint(it.TomlString())
		default:
			t.Fatal("invalid want:", fn)
		}
		if s != re {
			t.Fatalf("\n%v %v :\n%#v\n%#v\n", path, fn, re, s)
		}
	}
	return
}

var testData = []string{
	`
		# comment 0
		# comment 00

		key = "first key"
			# 1
		# 2
		[table] # 3
		int = 123456 # 4
		str = "string"
		
		[table2] # 5
			true = true
			false = false
			float = 123.45
			datetime = 2012-01-02T13:11:14Z
		
		# 6
			ia = [ 1 , 2] # 7
		# 8 overwrite 6
			ia = [ 1, 2,3] # 9 overwrite 7
		
			ba=[true,false]
		
		[multiline]
			ia= [ # 10.2
			# 10.1
			1,
				2, # 10
			3,
			] # 10.2
		[arrayarray]
			ia=[[1,2]]
		[arrayarray.2]
			ia = [
				[1,2], # 11
				[true],
			]
		[[array.tables]]
			ia=1
			aa="name"
		[[array.tables]]
			ia=2
			aa="jack"
	# LastComments`,
	"key", "fc", "# comment 0\n# comment 00",
	"key", "ts", "\n# comment 0\n# comment 00\nkey = \"first key\"",

	"table", "fc", "# 1\n# 2",
	"table", "ec", "# 3",

	"table.int", "i", "123456",
	"table.int", "ec", "# 4",
	"table.str", "s", "string",

	"table2", "ec", "# 5",
	"table2.true", "b", "true",
	"table2.false", "b", "false",
	"table2.float", "f", "123.45",
	"table2.datetime", "ts", "\ndatetime = 2012-01-02T13:11:14Z",
	"table2.datetime", "s", "2012-01-02 13:11:14",

	"table2.ia", "s", "[1,2,3]",
	"table2.ia", "fc", "# 8 overwrite 6",
	"table2.ia", "ec", "# 9 overwrite 7",

	"table2.ba", "s", "[true,false]",

	"multiline.ia", "s", "[1,2,3]",
	"multiline.ia", "ts", "\nia = [\n\t# 10.1\n\t1,2, # 10\n\t3] # 10.2",

	"arrayarray.ia", "s", "[[1,2]]",
	"arrayarray.2.ia", "s", "[[1,2],[true]]",
	"arrayarray.2.ia", "ts", "\nia = [[1,2], # 11\n\t[true]]",
	"array.tables", "s", "1", "ia",
}

var testDat = `
	Key = "string"
	Int = 123456
	Time = 2014-01-02T15:04:05Z
	[Table]
		IntArray = [1,2,3]
	`

type testStruct struct {
	Key   string
	Int   int
	Time  time.Time
	Table testTable
}
type testTable struct {
	IntArray []int
}

func TestTomlApply(t *testing.T) {
	wt := want.T(t)
	tm, err := Parse([]byte(testDat))

	wt.Nil(err)

	ts := testTable{[]int{0, 0, 0}}
	v := testStruct{
		"", 0, time.Time{},
		ts,
	}

	wt.Equal(tm.Apply(&v), 6)
	wt.Equal(v.Key, "string")
	wt.Equal(v.Int, 123456)
	wt.Equal(v.Time.String(), "2014-01-02 15:04:05 +0000 UTC")

	wt.Equal(len(v.Table.IntArray), 3)
	wt.Equal(v.Table.IntArray[0], 1)
	wt.Equal(v.Table.IntArray[1], 2)
	wt.Equal(v.Table.IntArray[2], 3)

	s := ""
	wt.Equal(tm["key"].Apply(&s), 0) // case sensitive
	wt.Equal(s, "")
	wt.Equal(tm["Key"].Apply(&s), 1)
	wt.Equal(s, "string")

	wt.Equal(tm["Table"].Apply(&ts), 0) // Table kind is TableName
	wt.Equal(tm.Fetch("Table").Apply(&ts), 3)

	// not support maps
	m := map[string]interface{}{
		"Key":   "",
		"Int":   0,
		"Time":  time.Time{},
		"Table": &testTable{},
	}
	wt.Equal(tm.Apply(m), 0)
}
