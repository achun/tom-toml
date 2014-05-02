package toml

import (
	"github.com/achun/testing-want"
	"testing"
	"time"
)

func init() {
	want.LocalFileLine = true
}

func TestTomlFile(t *testing.T) {
	if skipTest {
		return
	}
	wt := want.T(t)
	tm, err := LoadFile("tests/example.toml")
	wt.Nil(err)
	wantExample(wt, tm)
	source := tm.String()
	//println(source)
	ok := false
	defer func() {
		if !ok {
			println(source)
		}
	}()
	testBuilder(wt, []byte(source), "toml_test.go TestTomlFile()")
	tm, err = Parse([]byte(source))
	wt.Nil(err)
	wantExample(wt, tm)
	ok = true
}

func wantExample(wt want.Want, tm Toml) {
	id := tm.Id()
	wt.Equal(id.Kind(), InvalidKind)
	wt.Equal(id.multiComments, aString{"# last comments for", "# TOML document"})
	wt.Equal(id.eolComment, "")

	it := tm["title"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.multiComments, "# This is a TOML document. Boom.")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.String(), "TOML Example")

	it = tm["owner"]
	wt.Equal(it.Kind(), TableName)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "# owner information")

	it = tm["owner.name"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.String(), "Tom Preston-Werner")

	it = tm["owner.organization"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.String(), "GitHub")

	it = tm["owner.bio"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.String(), "GitHub Cofounder & CEO\nLikes tater tots and beer.")

	it = tm["owner.dob"]
	wt.Equal(it.Kind(), Datetime)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "# First class dates? Why not?")
	wt.Equal(it.String(), "1979-05-27T07:32:00Z")
	wt.Equal(it.Datetime().Format("2006-01-02T15:04:05Z00:00"), "1979-05-27T07:32:00Z00:00")

	it = tm["database"]
	wt.Equal(it.Kind(), TableName)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = tm["database.server"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.String(), "192.168.1.1")

	it = tm["database.ports"]
	wt.Equal(it.Kind(), IntegerArray, it.kind)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.Len(), 3)
	wt.Equal(it.Index(0).Integer(), 8001)
	wt.Equal(it.Index(1).Integer(), 8001)
	wt.Equal(it.Index(2).Integer(), 8002)

	it = tm["database.connection_max"]
	wt.Equal(it.Kind(), Integer)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.Integer(), 5000)

	it = tm["database.enabled"]
	wt.Equal(it.Kind(), Boolean, it.kind)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.Boolean(), true)

	it = tm["servers"]
	wt.Equal(it.Kind(), TableName, it.kind)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = tm["servers.alpha"]
	wt.Equal(it.Kind(), TableName, it.kind)
	wt.Equal(it.multiComments, "# You can indent as you please. Tabs or spaces. TOML don't care.")
	wt.Equal(it.eolComment, "")

	it = tm["servers.alpha.ip"]
	wt.Equal(it.Kind(), String, it.kind)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.String(), "10.0.0.1")

	it = tm["servers.alpha.dc"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.String(), "eqdc10")

	it = tm["servers.beta"]
	wt.Equal(it.Kind(), TableName)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = tm["servers.beta.ip"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.String(), "10.0.0.2")

	it = tm["servers.beta.dc"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.String(), "eqdc10")

	it = tm["servers.beta.country"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "# This should be parsed as UTF-8")
	wt.Equal(it.String(), "中国")

	it = tm["clients"]
	wt.Equal(it.Kind(), TableName)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = tm["clients.data"]
	wt.Equal(it.Kind(), Array)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "# just an update to make sure parsers support it")
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
	wt.Equal(iv.Index(0).Integer(), 1)
	wt.Equal(iv.Index(1).Integer(), 2)

	it = tm["clients.hosts"]
	wt.Equal(it.Kind(), StringArray)
	wt.Equal(it.multiComments, "# Line breaks are OK when inside arrays")
	wt.Equal(it.eolComment, "")
	wt.Equal(it.Len(), 2)
	wt.Equal(it.Index(0).Kind(), String)
	wt.Equal(it.Index(1).Kind(), String)
	wt.Equal(it.Index(0).String(), "alpha")
	wt.Equal(it.Index(1).String(), "omega")

	/** ArrayOfTables **/

	// products ==============
	it = tm["products"]
	wt.Equal(it.Kind(), ArrayOfTables)
	wt.Equal(it.Len(), 2)

	ts := it.TomlArray()
	wt.Equal(ts.Len(), 2)

	wt.Equal(ts[0].Id().multiComments, "# Products")
	wt.Equal(ts[0].Id().eolComment, "")
	wt.Equal(ts[1].Id().multiComments, "")
	wt.Equal(ts[1].Id().eolComment, "")

	it = ts[0]["name"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "Hammer")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = ts[0]["sku"]
	wt.Equal(it.Kind(), Integer)
	wt.Equal(it.Integer(), 738594937)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = ts[1]["name"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "Nail")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = ts[1]["sku"]
	wt.Equal(it.Kind(), Integer)
	wt.Equal(it.Integer(), 284758393)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = ts[1]["color"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "gray")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	// fruit ===============
	it = tm["fruit"]
	wt.Equal(it.Kind(), ArrayOfTables)
	wt.Equal(it.Len(), 2)

	ts = it.TomlArray()
	wt.Equal(ts.Len(), 2)

	tns, aots := ts[0].TableNames()
	wt.Equal(tns, []string{"physical"})
	wt.Equal(aots, []string{"variety"})

	id = ts[0].Id()
	wt.Equal(id.Kind(), InvalidKind)
	wt.Equal(id.multiComments, "# nested")
	wt.Equal(id.eolComment, "")

	id = ts[1].Id()
	wt.Equal(id.Kind(), InvalidKind)
	wt.Equal(id.multiComments, "")
	wt.Equal(id.eolComment, "")

	// fruit[0]
	it = ts[0]["name"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "apple")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = ts[0]["physical"]
	wt.Equal(it.Kind(), TableName)
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = ts[0]["physical.color"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "red")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = ts[0]["physical.shape"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "round")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = ts[0]["variety"]
	wt.Equal(it.Kind(), ArrayOfTables)
	wt.Equal(it.Len(), 2)

	// nested again, fruit[0]variety[0]
	tm = it.TomlArray()[0]
	wt.Equal(len(tm), 2) // with iD

	it = tm["name"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "red delicious")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	// nested nested, fruit[0]variety[1]
	tm = ts[0]["variety"].TomlArray()[1]
	wt.Equal(len(tm), 2) // with iD

	it = tm["name"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "granny smith")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	// fruit[1]
	it = ts[1]["name"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "banana")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")

	it = ts[1]["variety"]
	wt.Equal(it.Kind(), ArrayOfTables)
	wt.Equal(it.Len(), 1)

	tm = it.TomlArray()[0]
	wt.Equal(len(tm), 2) // with iD

	it = tm["name"]
	wt.Equal(it.Kind(), String)
	wt.Equal(it.String(), "plantain")
	wt.Equal(it.multiComments, "")
	wt.Equal(it.eolComment, "")
}

func TestTomlEmpty(t *testing.T) {
	if skipTest {
		return
	}
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
	if skipTest {
		return
	}
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
