package toml

import (
	"testing"
)

func TestItemAdd(t *testing.T) {
	a := NewItem(Array)
	assertError(t, a.Add(1), "emptyArray.Add(int)")
	assertError(t, a.Add(2, 3), "IntergerArray.Add(int,int)")

	assertEqual(t, a.kind, IntegerArray, "Kind != IntergerArray")
	assertEqual(t, a.String(), "[1,2,3]")

	aa := NewItem(Array)
	assertError(t, aa.Add(a), "emptyArray.Add(IntergerArray)")
	assertEqual(t, aa.kind, Array, "Kind != Array")

	b := NewItem(Array)
	b.Add("hello")
	b.Add("world")
	assertError(t, aa.Add(b), "Array.Add(StringArray)")
	assertError(t, aa.Add(a, b), "Array.Add(IntergerArray,StringArray)")

	assertEqual(t, b.Add(1), NotSupported, "StringArray.Add(int)")

	assertEqual(t, aa.Add(1), NotSupported, "Array.Add(int)")

}

func TestItemPlain(t *testing.T) {
	a := NewItem(Datetime)
	assertError(t, a.SetAs("2012-01-02T13:11:14Z", Datetime))

	assertEqual(t, a.TomlString(), "2012-01-02T13:11:14Z")
	assertEqual(t, a.String(), "2012-01-02 13:11:14")
}

func TestItemArrayOfTable(t *testing.T) {
	aot := NewItem(ArrayOfTables)
	v := NewItem(Integer)
	v.Set(1)
	ts := Tables{"A": &v.Value}
	assertError(t, aot.AddTables(ts))
	assertError(t, v.Set(2))
	assertFalse(t, v.Int() == aot.Tables(-1)["A"].Int())

}
