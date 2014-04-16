package toml

import (
	"github.com/achun/testing-want"
	"testing"
)

func TestItemAdd(t *testing.T) {
	wt := want.T(t)
	a := NewItem(Array)

	wt.Nil(a.Add(1), "emptyArray.Add(int)")
	wt.Nil(a.Add(2, 3), "IntergerArray.Add(int,int)")

	wt.Equal(a.kind, IntegerArray, "Kind != IntergerArray")
	wt.Equal(a.String(), "[1,2,3]")

	aa := NewItem(Array)
	wt.Nil(aa.Add(a), "emptyArray.Add(IntergerArray)")
	wt.Equal(aa.kind, Array, "Kind != Array")

	b := NewItem(Array)
	b.Add("hello")
	b.Add("world")
	wt.Nil(aa.Add(b), "Array.Add(StringArray)")
	wt.Nil(aa.Add(a, b), "Array.Add(IntergerArray,StringArray)")

	wt.Equal(b.Add(1), NotSupported, "StringArray.Add(int)")

	wt.Equal(aa.Add(1), NotSupported, "Array.Add(int)")

}

func TestItemPlain(t *testing.T) {
	wt := want.T(t)
	a := NewItem(Datetime)

	wt.Nil(a.SetAs("2012-01-02T13:11:14Z", Datetime))

	wt.Equal(a.TomlString(), "2012-01-02T13:11:14Z")
	wt.Equal(a.String(), "2012-01-02 13:11:14")
}

func TestItemArrayOfTable(t *testing.T) {
	wt := want.T(t)

	aot := NewItem(ArrayOfTables)
	v := NewItem(Integer)
	v.Set(1)

	ts := Table{"A": &v.Value}
	wt.Nil(aot.AddTable(ts))
	wt.Nil(v.Set(2))
	wt.Equal(aot.Table(-1)["A"].Int(), v.Int())
}
