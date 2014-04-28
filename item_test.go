package toml

import (
	"github.com/achun/testing-want"
	"testing"
)

func TestItemAdd(t *testing.T) {
	wt := want.T(t)

	wt.Error(GenItem(InvalidKind).Add(1))
	wt.Error(GenItem(String).Add(1))
	wt.Error(GenItem(Integer).Add(1))
	wt.Error(GenItem(Float).Add(1))
	wt.Error(GenItem(Boolean).Add(1))
	wt.Error(GenItem(Datetime).Add(1))

	a := GenItem(Array)

	wt.Nil(a.Add(1), "emptyArray.Add(int)")
	wt.Nil(a.Add(2, 3), "IntegerArray.Add(int,int)")

	wt.Equal(a.kind, IntegerArray)
	wt.Equal(a.Len(), 3)
	wt.Equal(a.String(), "[1, 2, 3]")
	wt.Equal(a.Index(0).Int(), int64(1))
	wt.Equal(a.Index(1).Int(), int64(2))
	wt.Equal(a.Index(2).Int(), int64(3))
	// 负数下标
	wt.Equal(a.Index(-1).Int(), int64(3))
	// 超出下标
	wt.Nil(a.Index(3))
	wt.Equal(a.Index(3).Int(), int64(0))

	wt.Error(a.Add("string"))

	aa := GenItem(Array)
	wt.Nil(aa.Add(a), "Array.Add(IntegerArray)")
	wt.Equal(aa.kind, Array)

	b := GenItem(Array)
	wt.Nil(b.Add("hello"))
	wt.Nil(b.Add("world"))
	wt.Equal(b.kind, StringArray)
	wt.Equal(b.Len(), 2)

	wt.Nil(aa.Add(b), "Array.Add(StringArray)")
	wt.Equal(aa.kind, Array)
	wt.Equal(aa.Len(), 2)

	wt.Nil(aa.Add(a, b), "Array.Add(IntegerArray,StringArray)")
	wt.Equal(aa.kind, Array)

	wt.Equal(b.Add(1), NotSupported, "StringArray.Add(int)")

	wt.Equal(aa.kind, Array)
	wt.Equal(aa.Add(1), NotSupported, aa.kind.String(), ".Add(int)")

}

func TestItemPlain(t *testing.T) {
	wt := want.T(t)
	a := GenItem(Datetime)

	wt.Nil(a.SetAs("2012-01-02T13:11:14Z", Datetime))

	wt.Equal(a.String(), "2012-01-02T13:11:14Z")
	wt.Equal(a.String(), "2012-01-02T13:11:14Z")
}
