package toml

import (
	"github.com/achun/testing-want"
	"testing"
	"time"
)

func TestTomlFile(t *testing.T) {
	wt := want.T(t)
	tm, err := LoadFile("tests/example.toml")
	wt.Nil(err)
	wantExample(wt, tm)
	source := tm.String()
	testBuilder(wt, []byte(source), "TestTomlFile tm.String()")
	println(source)
}

func wantExample(wt want.Want, tm Toml) {
	it := tm["title"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.MultiComments, "# This is a TOML document. Boom.")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.String(), "TOML Example")

	it = tm["owner"]
	wt.Equal(it.Kind(), TableName)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "# owner information")

	it = tm["owner.name"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.String(), "Tom Preston-Werner")

	it = tm["owner.organization"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.String(), "GitHub")

	it = tm["owner.bio"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.String(), "GitHub Cofounder & CEO\nLikes tater tots and beer.")

	it = tm["owner.dob"]
	wt.Equal(it.Kind(), Datetime)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "# First class dates? Why not?")
	wt.Equal(it.String(), "1979-05-27T07:32:00Z")
	wt.Equal(it.Datetime().Format("2006-01-02T15:04:05Z00:00"), "1979-05-27T07:32:00Z00:00")

	it = tm["database"]
	wt.Equal(it.Kind(), TableName)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")

	it = tm["database.server"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.String(), "192.168.1.1")

	it = tm["database.ports"]
	wt.Equal(it.Kind(), IntegerArray, it.kind)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.Len(), 3)
	wt.Equal(it.Index(0).Interger(), 8001)
	wt.Equal(it.Index(1).Interger(), 8001)
	wt.Equal(it.Index(2).Interger(), 8002)

	it = tm["database.connection_max"]
	wt.Equal(it.Kind(), Integer)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.Interger(), 5000)

	it = tm["database.enabled"]
	wt.Equal(it.Kind(), Boolean, it.kind)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.Boolean(), true)

	it = tm["servers"]
	wt.Equal(it.Kind(), TableName, it.kind)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")

	it = tm["servers.alpha"]
	wt.Equal(it.Kind(), TableName, it.kind)
	wt.Equal(it.MultiComments, "# You can indent as you please. Tabs or spaces. TOML don't care.")
	wt.Equal(it.EolComment, "")

	it = tm["servers.alpha.ip"]
	wt.Equal(it.Kind(), String, it.kind)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.String(), "10.0.0.1")

	it = tm["servers.alpha.dc"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.String(), "eqdc10")

	it = tm["servers.beta"]
	wt.Equal(it.Kind(), TableName)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")

	it = tm["servers.beta.ip"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.String(), "10.0.0.2")

	it = tm["servers.beta.dc"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.String(), "eqdc10")

	it = tm["servers.beta.country"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "# This should be parsed as UTF-8")
	wt.Equal(it.String(), "中国")

	it = tm["clients"]
	wt.Equal(it.Kind(), TableName)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "")

	it = tm["clients.data"]
	wt.Equal(it.Kind(), Array)
	wt.Equal(it.MultiComments, "")
	wt.Equal(it.EolComment, "# just an update to make sure parsers support it")
	wt.Equal(it.Len(), 2)

	iv := it.Index(0)
	wt.Equal(iv.Kind(), StringArray)
	wt.Equal(iv.Len(), 2)
	wt.Equal(iv.Index(0).Kind(), String)
	wt.Equal(iv.Index(1).Kind(), String)
	wt.Equal(iv.Index(0).String(), "gamma")
	wt.Equal(iv.Index(1).String(), "delta")

	iv = it.Index(1)
	wt.Equal(iv.Kind(), IntegerArray, iv.kind)
	wt.Equal(iv.Len(), 2)
	wt.Equal(iv.Index(0).Kind(), Integer)
	wt.Equal(iv.Index(1).Kind(), Integer)
	wt.Equal(iv.Index(0).Interger(), 1)
	wt.Equal(iv.Index(1).Interger(), 2)

	it = tm["clients.hosts"]
	wt.Equal(it.Kind(), StringArray)
	wt.Equal(it.MultiComments, "# Line breaks are OK when inside arrays")
	wt.Equal(it.EolComment, "")
	wt.Equal(it.Len(), 2)
	wt.Equal(it.Index(0).Kind(), String)
	wt.Equal(it.Index(1).Kind(), String)
	wt.Equal(it.Index(0).String(), "alpha")
	wt.Equal(it.Index(1).String(), "omega")

	it = tm["products"]
	wt.Equal(it.Kind(), ArrayOfTables)
	wt.Equal(it.Len(), 2)

	ts := it.Tables()
	wt.Equal(ts.Len(), 2)

}

func TestTomlEmpty(t *testing.T) {
	wt := want.T(t)
	tm, err := Parse([]byte(``))

	wt.Nil(err)
	wt.Equal(len(tm), 1) // was iD
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
