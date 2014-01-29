package toml

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Kind uint

// don't change order
const (
	InvalidKind Kind = iota
	String
	Integer
	Float
	Boolean
	Datetime
	StringArray
	IntegerArray
	FloatArray
	BooleanArray
	DatetimeArray
	Array
	Table
	ArrayOfTables
)

func (k Kind) String() string {
	return kindsName[k]
}

var kindsName = [...]string{
	"InvalidKind",
	"String",
	"Integer",
	"Float",
	"Boolean",
	"Datetime",
	"StringArray",
	"IntegerArray",
	"FloatArray",
	"BooleanArray",
	"DatetimeArray",
	"Array",
	"Table",
	"ArrayOfTables",
}

var (
	NotSupported  = errors.New("not supported")
	OutOfRange    = errors.New("out of range")
	InternalError = errors.New("internal error")
	InvalidItem   = errors.New("invalid Item")
)
var _counter = 0

func counter(idx int) int {
	if idx > 0 {
		return idx
	}
	_counter++
	return _counter
}
func NewItem(kind Kind) *Item {
	if kind < 0 || kind > ArrayOfTables {
		return nil
	}

	it := &Item{}
	it.kind = kind

	it.idx = counter(0)

	if kind == ArrayOfTables {
		it.v = []Tables{}
	}

	return it
}

type Value struct {
	kind          Kind
	v             interface{}
	MultiComments string // Multi-line comments
	EolComment    string // end of line comment
	key           string // for TOML formatter
	idx           int
}

func NewValue(kind Kind) *Value {
	if kind < 0 || kind > Table {
		return nil
	}

	it := &Value{}
	it.kind = kind

	it.idx = counter(0)

	return it
}

func (p *Value) Kind() Kind {
	return p.kind
}

func (p *Value) IsValid() bool {
	return p.kind != InvalidKind && (p.v != nil || p.kind == Table)
}

func (p *Value) canNotSet(k Kind) bool {
	return p.kind != InvalidKind && p.kind != k
}

func (p *Value) Set(x interface{}) error {
	switch v := x.(type) {
	case string:
		if p.canNotSet(String) {
			return NotSupported
		}
		p.v = v
		p.kind = String
	case bool:
		if p.canNotSet(Boolean) {
			return NotSupported
		}
		p.v = v
		p.kind = Boolean
	case float64:
		if p.canNotSet(Float) {
			return NotSupported
		}
		p.v = v
		p.kind = Float
	case time.Time:
		if p.canNotSet(Datetime) {
			return NotSupported
		}
		p.v = v.UTC()
		p.kind = Datetime
	case int64:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		p.v = int64(v)
		p.kind = Integer
	case int:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		p.v = int64(v)
		p.kind = Integer
	case uint:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		p.v = int64(v)
		p.kind = Integer
	case int8:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		p.v = int64(v)
		p.kind = Integer
	case int16:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		p.v = int64(v)
		p.kind = Integer
	case int32:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		p.v = int64(v)
		p.kind = Integer
	case uint8:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		p.v = int64(v)
		p.kind = Integer
	case uint16:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		p.v = int64(v)
		p.kind = Integer
	case uint32:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		p.v = int64(v)
		p.kind = Integer
	case uint64:
		if p.canNotSet(Integer) {
			return NotSupported
		}
		if v > 9223372036854775807 {
			return OutOfRange
		}
		p.kind = Integer
		p.v = int64(v)
	default:
		return NotSupported
	}
	p.idx = counter(p.idx)
	return nil
}

func (p *Value) SetAs(s string, kind Kind) (err error) {
	if p.canNotSet(kind) {
		return NotSupported
	}
	switch kind {
	case String:
		p.v = s
	case Integer:
		var v int64
		v, err = strconv.ParseInt(s, 10, 64)
		if err == nil {
			p.v = v
		}
	case Float:
		var v float64
		v, err = strconv.ParseFloat(s, 64)
		if err == nil {
			p.v = v
		}
	case Boolean:
		var v bool
		v, err = strconv.ParseBool(s)
		if err == nil {
			p.v = v
		}
	case Datetime:
		var v time.Time
		v, err = time.Parse(time.RFC3339, s) // time zone +00:00
		if err == nil {
			p.v = v.UTC()
		}
	default:
		return NotSupported
	}
	if err == nil {
		p.kind = kind
		p.idx = counter(p.idx)
	}
	return
}

func asValue(i interface{}) (v *Value, ok bool) {
	it, ok := i.(*Item)
	if ok {
		v = &it.Value
	} else {
		v, ok = i.(*Value)
	}
	return
}

// Add element for Array or typeArray
func (p *Value) Add(ai ...interface{}) error {
	if p.kind != InvalidKind && (p.kind < StringArray || p.kind > Array) {
		return NotSupported
	}
	if len(ai) == 0 {
		return nil
	}
	vs := make([]*Value, len(ai))
	k := 0
	for i, s := range ai {
		v, ok := asValue(s)
		if !ok {
			v = &Value{}
		}
		if ok && (v.kind == InvalidKind || v.kind > DatetimeArray) || !ok && nil != v.Set(s) {
			return NotSupported
		}
		if v.kind < StringArray {
			k = k | 1<<v.kind
		} else {
			k = k | 1
		}
		if k > 2 && k != 1<<v.kind {
			return NotSupported
		}
		v.idx = counter(v.idx)
		vs[i] = v
	}
	if k == 1 { // typeArray
		if p.canNotSet(Array) {
			return NotSupported
		}
		if p.v == nil {
			p.v = []*Value{}
		}
		v := p.v.([]*Value)
		p.v = append(v, vs...)
	} else { // plain value

		if p.v == nil {
			p.v = []*Value{}
		}
		v := p.v.([]*Value)

		if len(v) != 0 &&
			(v[0].kind != vs[0].kind || p.kind != vs[0].kind+StringArray-String) {
			return NotSupported
		}
		if len(v) == 0 {
			p.kind = vs[0].kind + StringArray - String
		}
		p.v = append(v, vs...)
	}
	return nil
}

func (p *Value) String() string {
	return p.string(0, -1)
}

func (p *Value) TomlString() string {
	return p.string(1, -1)
}

func (p *Value) Int() int64 {
	if p.kind != Integer {
		return 0
	}
	return p.v.(int64)
}
func (p *Value) UInt() uint64 {
	if p.kind != Integer {
		return 0
	}
	return uint64(p.v.(int64))
}
func (p *Value) Float() float64 {
	if p.kind != Float {
		return 0
	}
	return p.v.(float64)
}
func (p *Value) Boolean() bool {
	if p.kind != Boolean {
		return false
	}
	return p.v.(bool)
}
func (p *Value) Datetime() time.Time {
	if p.kind != Datetime {
		return time.Unix(978307200-63113904000, 0).UTC()
	}
	return p.v.(time.Time)
}

// Len returns length for Array , typeArray.
// Otherwise Kind return -1.
func (p *Value) Len() int {
	if p.kind >= StringArray && p.kind <= Array {
		a, ok := p.v.([]*Value)
		if ok {
			return len(a)
		}
	}
	return -1
}

// Index returns *Value for Array , typeArray.
// Otherwise Kind return nil.
func (p *Value) Index(idx int) *Value {
	if p.kind < StringArray && p.kind > Array {
		return nil
	}
	a, ok := p.v.([]*Value)
	if !ok {
		return nil
	}
	size := len(a)
	if idx < 0 {
		idx = size + idx
	}
	if idx < 0 || idx >= size {
		return nil
	}
	return a[idx]
}

// Tables is an map container for Kind() < Table
type Tables map[string]*Value

func (p Tables) String() (fmt string) {
	if len(p) == 0 {
		return
	}
	var keyidx []int
	keys := map[int]string{}
	for key, it := range p {
		if it != nil && it.IsValid() {
			if it.kind < Table {
				keys[it.idx] = key
				keyidx = append(keyidx, it.idx)
			} else {
				// panic ?
			}
		}
	}

	sort.Sort(sort.IntSlice(keyidx))

	for _, i := range keyidx {
		key := keys[i]
		it := p[key]
		if it == nil {
			// panic(InternalError) ?
			continue
		}
		it.key = key
		fmt += it.string(1, 1)
	}
	return
}

type Item struct {
	Value
}

func (p *Item) AddTables(ts Tables) error {
	if ts == nil || p.kind != ArrayOfTables && p.kind != InvalidKind {
		return NotSupported
	}
	if p.kind == InvalidKind {
		p.v = []Tables{}
		p.kind = ArrayOfTables
	}
	aot, ok := p.v.([]Tables)
	if !ok {
		return InternalError
	}
	p.v = append(aot, ts)
	return nil
}

func (p *Item) DelTables(idx int) error {
	if p.kind != ArrayOfTables {
		return NotSupported
	}
	aot, ok := p.v.([]Tables)
	if !ok {
		return InternalError
	}
	size := len(aot)
	if idx < 0 {
		idx = size + idx
	}
	if idx < 0 || idx >= size {
		return OutOfRange
	}
	p.v = append(aot[:idx], aot[idx+1:]...)
	return nil
}

// Index returns Tables for ArrayOfTables.
// Otherwise Kind return nil.
func (p *Item) Tables(idx int) Tables {
	if p.kind != ArrayOfTables {
		return nil
	}
	aot, ok := p.v.([]Tables)
	if !ok {
		return nil
	}
	size := len(aot)
	if idx < 0 {
		idx = size + idx
	}
	if idx < 0 || idx >= size {
		return nil
	}
	return aot[idx]
}

// Len returns length for Array , typeArray , ArrayOfTables.
// Otherwise Kind return -1.
func (p *Item) Len() int {
	if p.kind == ArrayOfTables {
		a, ok := p.v.([]Tables)
		if ok {
			return len(a)
		}
		return -1
	}
	return p.Value.Len()
}
func (p *Item) String() string {
	return p.string(0, -1)
}

func (p *Item) TomlString() string {
	return p.string(1, -1)
}

// FixComments returns comments with "\n"
func FixComments(str string) string {
	as := strings.Split(str, "#")
	re := ""
	for _, s := range as {
		s = strings.TrimSpace(s)
		if len(s) != 0 {
			if len(re) != 0 {
				re += "\n# " + s
			} else {
				re += "# " + s
			}
		}
	}
	return re
}

func (p *Item) string(layout int, indent int) (fmt string) {
	if p.kind != ArrayOfTables {
		return p.Value.string(layout, indent)
	}
	p.MultiComments = FixComments(p.MultiComments)
	p.EolComment = FixComments(p.EolComment)
	fmt = p.MultiComments
	// Item is comment?
	if p.v == nil && p.kind < Table {
		return
	}
	aot, ok := p.v.([]Tables)
	if !ok {
		panic(InternalError)
	}
	indents := ""
	if indent > 0 {
		indents = strings.Repeat("\t", indent)
	}
	tn := indents + "[[" + p.key + "]]"

	for i, ts := range aot {

		if i == 0 {
			if layout == 0 {
				fmt += tn
			} else {
				fmt += p.Value.comments(tn, layout, indent)
			}
		} else {
			fmt += "\n" + tn
		}
		fmt += ts.String()
	}
	return
}

func (p *Value) string(layout int, indent int) string {
	if !p.IsValid() {
		return ""
	}
	s := ""
	switch p.kind {
	case String:
		s = p.v.(string)
		if layout != 0 {
			s = strconv.Quote(s)
		}
	case Integer:
		s = strconv.FormatInt(p.v.(int64), 10)
	case Float:
		s = strconv.FormatFloat(p.v.(float64), 'f', -1, 64)
	case Boolean:
		s = strconv.FormatBool(p.v.(bool))
	case Datetime:
		if layout == 0 {
			s = p.v.(time.Time).Format("2006-01-02 15:04:05")
		} else {
			s = p.v.(time.Time).Format(time.RFC3339)
		}

	case StringArray, IntegerArray, FloatArray, BooleanArray, DatetimeArray:
		return p.typeArrayString(layout, indent)
	case Array:
		return p.typeArrayString(layout, indent)
	case Table:
		indents := ""
		if indent > 0 {
			indents = strings.Repeat("\t", indent)
		}
		s = indents + "[" + p.key + "]"
	default:
		panic(InternalError)
	}
	if layout != 0 {
		return p.comments(s, layout, indent)
	}
	return s
}

func (p *Value) comments(v string, layout int, indent int) string {
	p.MultiComments = FixComments(p.MultiComments)
	p.EolComment = FixComments(p.EolComment)
	key := p.key
	indents := ""
	if indent > 0 {
		indents = strings.Repeat("\t", indent)
	}
	if len(key) != 0 {
		if p.kind < Table {
			key = indents + key + " = "
		} else {
			key = ""
		}
		if len(p.MultiComments) == 0 {
			v = "\n" + key + v
		} else if key == "" {
			//v = v
		} else {
			v = key + v
		}
	}

	if len(p.MultiComments) != 0 {
		v = "\n" + indents +
			strings.Replace(p.MultiComments, "\n#", "\n"+indents+"#", -1) +
			"\n" + v
	}
	if layout&2 == 2 {
		v += ","
	}
	if len(p.EolComment) != 0 {
		v += " " + p.EolComment
		if layout&2 == 2 {
			v += "\n"
			if indent != -1 {
				v += indents + "\t"
			}
		}
	}

	return v
}

func (p *Value) typeArrayString(layout int, indent int) string {
	a := p.v.([]*Value)
	s := ""
	max := len(a) - 1
	for i, bv := range a {
		if i != max {
			s += bv.string(layout|2, indent) // for "," # comment
		} else {
			s += bv.string(layout&1, indent)
		}
	}
	if layout != 0 {
		return p.comments("["+s+"]", layout, indent)
	}
	return "[" + s + "]"
}
