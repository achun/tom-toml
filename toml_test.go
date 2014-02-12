package toml

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestTomlFile(t *testing.T) {
	tm, err := LoadFile("tests/example.toml")
	assertError(t, err)
	assertEqual(t, tm["servers.alpha.ip"].String(), "10.0.0.1")
	assertEqual(t, tm["servers.beta.ip"].String(), "10.0.0.2")

	tm2, err := Parse([]byte(tm.String()))
	assertError(t, err)

	assertEqual(t, tm["servers.alpha.ip"].String(), tm2["servers.alpha.ip"].String())
	assertEqual(t, tm["servers.alpha.dc"].String(), tm2["servers.alpha.dc"].String())
}

func TestTomlEmpty(t *testing.T) {
	tm, err := Parse([]byte(``))
	assertError(t, err)
	assertEqual(t, len(tm), 0)
}

func TestToml(t *testing.T) {
	tm := testToml(t, testData[0], testData[1:])
	testToml(t, tm.String(), testData[1:])

}
func testToml(t *testing.T, source string, as []string) Toml {
	tm := assertToml(t, source, as)
	ia := tm[""]
	assertFalse(t, ia != nil, "LastComments invalid")
	assertEqual(t, ia.MultiComments, "# LastComments", "LastComments")

	ia = tm["multiline.ia"]
	assertFalse(t, ia != nil, "multiline.ia.Value invalid")
	assertEqual(t, ia.EolComment, "# 10.2", "multiline.ia.EolComment")

	assertEqual(t, ia.Index(0).MultiComments, "# 10.1", "multiline.ia·Index[0].MultiComments")
	assertEqual(t, ia.Index(1).EolComment, "# 10", "multiline.ia·Index[1].EolComment")

	ia = tm["arrayarray.2.ia"]
	assertFalse(t, ia != nil, "arrayarray.2.ia invalid")

	assertEqual(t, ia.Index(0).String(), "[1,2]", "arrayarray.2.ia·Index(0)")
	assertEqual(t, ia.Index(0).EolComment, "# 11", "arrayarray.2.ia·Index(0).EolComment")

	aot := tm["array.tables"]
	assertFalse(t, aot != nil, "array.tables invalid")

	ts := aot.Tables(0)
	assertFalse(t, ts != nil, "array.tables.Index(0) invalid")
	assertEqual(t, ts["ia"].Int(), int64(1), "array.tables.Index(0).ia")
	assertEqual(t, ts["aa"].String(), "name", "array.tables.Index(0).aa")
	ts = aot.Tables(1)
	assertEqual(t, ts["ia"].Int(), int64(2), "array.tables.Index(1).ia")
	assertEqual(t, ts["aa"].String(), "jack", "array.tables.Index(1).aa")
	return tm
}

func iS(is []interface{}) string {
	if len(is) != 0 {
		return fmt.Sprint(is...)
	}
	return ""
}
func getCaller(skip int) string {
	pc, _, line, _ := runtime.Caller(skip)
	return runtime.FuncForPC(pc).Name() + ":" + fmt.Sprint(line) + "\n"
}
func assertFalse(t *testing.T, ok bool, is ...interface{}) {
	if !ok {
		t.Fatal(getCaller(2), iS(is))
	}
}
func assertEqual(t *testing.T, got, want interface{}, is ...interface{}) {
	if want != got {
		t.Fatal(getCaller(2), "expected:", want, ",but got:", got, iS(is))
	}
}

func assertError(t *testing.T, err error, is ...interface{}) {
	if err != nil {
		t.Fatal(getCaller(2), err, iS(is))
	}
}

func wantPainc(t *testing.T, msg string, fn func()) {
	defer func() {
		str := fmt.Sprint(recover())
		if msg != str {
			t.Fatal(getCaller(3), "expected panic:", msg, ",but got:", str)
		}
	}()
	fn()
}

func wantError(t *testing.T, err error, is ...interface{}) {
	if err == nil {
		t.Fatal(getCaller(2), "want an error but not got.", iS(is))
	}
}

func assertToml(t *testing.T, source string, as []string) (tm Toml) {

	tm, err := Parse([]byte(source))
	assertError(t, err)

	for i := 0; i < len(as); i += 3 {
		path := as[i]
		fn := as[i+1]
		re := as[i+2]
		it := tm[path]

		if it != nil && it.kind == ArrayOfTables {
			ts := it.Tables(0)
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

func TestToml_Apply(t *testing.T) {
	tm, err := Parse([]byte(testDat))
	assertError(t, err)
	v := testStruct{
		"", 0, time.Time{},
		testTable{[]int{0, 0, 0}},
	}
	assertEqual(t, tm.Apply(&v), 6)
	assertEqual(t, v.Key, "string")
	assertEqual(t, v.Int, 123456)
	assertEqual(t, v.Time.String(), "2014-01-02 15:04:05 +0000 UTC")
	//assertFalse(t, v.Table != nil)
	assertEqual(t, len(v.Table.IntArray), 3)
	assertEqual(t, v.Table.IntArray[0], 1)
	assertEqual(t, v.Table.IntArray[1], 2)
	assertEqual(t, v.Table.IntArray[2], 3)

	m := map[string]interface{}{
		"Key":   "",
		"Int":   0,
		"Time":  time.Time{},
		"Table": &testTable{},
	}
	assertEqual(t, tm.Apply(m), 0)
}
