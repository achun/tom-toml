package toml

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Kind 用来标识 TOML 所有的规格.
// 对于用于配置的 TOML 定义, Kind 其实是分段定义和值定义的.
// 由于 TOML 官方没有使用 段/值 这样的词汇, tom-toml 选用规格这个词.
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

// 计数器为保持格式化输出次序准备.
var _counter = 0

func counter(idx int) int {
	if idx > 0 {
		return idx
	}
	_counter++
	return _counter
}

// NewItem 函数返回一个新 *Item.
// 为保持格式化输出次序, NewItem 内部使用了一个计数器.
// 使用者应该使用该函数来得到新的 *Item. 而不是用 new(Item) 获得.
// 那样的话就无法保持格式化输出次序.
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

// Value 用来存储除了 ArrayOfTables 规格的数据.
// Table 规格本身也当成一个数据保存, 具体影响见 Toml 的结构.
type Value struct {
	kind          Kind
	v             interface{}
	MultiComments string // Multi-line comments
	EolComment    string // end of line comment
	key           string // for TOML formatter
	idx           int
}

// NewValue 函数返回一个新 *Value.
// 为保持格式化输出次序, NewValue 内部使用了一个计数器.
// 使用者应该使用该函数来得到新的 *Value. 而不是用 new(Value) 获得.
// 那样的话就无法保持格式化输出次序.
func NewValue(kind Kind) *Value {
	if kind < 0 || kind > Table {
		return nil
	}

	it := &Value{}
	it.kind = kind

	it.idx = counter(0)

	return it
}

// Kind 返回数据的风格.
// 注意 Table 只是一个规格标记, 并不保存真正的值.
func (p *Value) Kind() Kind {
	return p.kind
}

// IsValid 返回 *Value 是否是一个有效的规格.
func (p *Value) IsValid() bool {
	return p != nil && p.kind != InvalidKind && (p.v != nil || p.kind == Table)
}

func (p *Value) canNotSet(k Kind) bool {
	return p.kind != InvalidKind && p.kind != k
}

// Set 用来设置 *Value 要保存的具体值. 参数 x 的类型范围可以是
// String,Integer,Float,Boolean,Datetime 之一
// 如果 *Value 的 Kind 是 InvalidKind(也就是没有明确值类型),
// 调用 Set 后, *Value 的 kind 会相应的更改, 否则要求 x 的类型必须符合 *Value 的 kind
// Set 失败会返回 NotSupported 错误.
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

// SetAs是个便捷方法, 通过参数 kind 对 string 参数进行转换并执行 Set.
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

// Add element for Array or typeArray.
// Add 方法为类型数组或者二维数组添加值对象.
// 如果 *Value 是 InvalidKind, Add 会根据参数的类型自动设置 kind. 否则要求参数符合 kind.
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

// String 返回 *Value 保存数据的字符串表示.
// 注意所有的规格定义都是可以字符串化的.
func (p *Value) String() string {
	return p.string(0, 0)
}

// TomlString 返回用于格式化输出的字符串表示.
// 与 String 不同, 输出包括了注释和缩进.
func (p *Value) TomlString() string {
	return p.string(1, 0)
}

// 如果值是 Interger 可以使用 Int 返回其 int64 值.
// 否则返回 0
func (p *Value) Int() int64 {
	if p.kind != Integer {
		return 0
	}
	return p.v.(int64)
}

// 如果值是 Interger 可以使用 UInt 返回其 uint64 值.
// 否则返回 0
func (p *Value) UInt() uint64 {
	if p.kind != Integer {
		return 0
	}
	return uint64(p.v.(int64))
}

// 如果值是 Float 可以使用 Float 返回其 float64 值.
// 否则返回 0
func (p *Value) Float() float64 {
	if p.kind != Float {
		return 0
	}
	return p.v.(float64)
}

// 如果值是 Boolean 可以使用 Boolean 返回其 bool 值.
// 否则返回 false
func (p *Value) Boolean() bool {
	if p.kind != Boolean {
		return false
	}
	return p.v.(bool)
}

// 如果值是 Datetime 可以使用 Datetime 返回其 time.Time 值.
// 否则返回 UTC 元年1月1日.
func (p *Value) Datetime() time.Time {
	if p.kind != Datetime {
		return time.Unix(978307200-63113904000, 0).UTC()
	}
	return p.v.(time.Time)
}

// Len returns length for Array , typeArray.
// Otherwise Kind return -1.
// Len 返回数组类型元素个数.
// 否则返回 -1.
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
// Index 根据 idx 下标返回类型数组或二维数组对应的元素.
// 如果非数组或者下标超出范围返回 nil.
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
// Tables 是一个 maps 容器, 用来保存 TOML 规格定义中的 key/value.
// Tables 这个名字是 tom-toml 实现需求定义的, 不在 TOML 定义中.
type Tables map[string]*Value

// String 返回 Tables 的 TOML 格式化字符串
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

// Item 扩展自 Value, 支持 ArrayOfTables.
type Item struct {
	Value
}

// AddTables 为 ArrayOfTables 增加新的 Tables 元素.
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

// DelTables 为 ArrayOfTables 删除下标为 idx 的元素.
// 如果 idx 超出下标范围返回 OutOfRange 错误.
// 如果保存了非法的数据会返回 InternalError 错误.
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
// Tables 为 ArrayOfTables 返回下标为 idx 的 Tables 元素.
// 非 ArrayOfTables 返回 nil.
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
// Len 返回数组类型或者ArrayOfTables的元素个数.
// 否则返回 -1.
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

// String 返回 *Item 保存数据的字符串表示.
// 注意所有的规格定义都是可以字符串化的.
func (p *Item) String() string {
	return p.string(0, 0)
}

// TomlString 返回用于格式化输出的字符串表示.
// 与 String 不同, 输出包括了注释和缩进.
func (p *Item) TomlString() string {
	return p.string(1, 0)
}

// FixComments returns comments, newline equal "\n"
// 对完整注释字符串进行修正, 修正后的多行字符串使用换行符 "\n".
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
	ts := layout&1 == 1 // toml string
	key := p.key
	indents := ""
	if ts && indent >= 0 {
		indents = strings.Repeat("\t", indent)
	}
	p.MultiComments = FixComments(p.MultiComments)
	p.EolComment = FixComments(p.EolComment)
	if ts && len(key) != 0 {
		if p.kind < Table {
			key = indents + key + " = "
		} else {
			key = ""
		}
		if ts && len(p.MultiComments) == 0 {
			v = "\n" + key + v
		} else if key == "" {
			//v = v
		} else {
			v = key + v
		}
	}

	if ts && layout>>1 == 1 { // element of array
		indents += "\t"
	}

	if ts && len(p.MultiComments) != 0 {
		if len(key) == 0 {
			v = "\n" + strings.Replace(p.MultiComments, "#", indents+"#", -1) +
				"\n" + indents + v
		} else {
			v = "\n" + strings.Replace(p.MultiComments, "#", indents+"#", -1) +
				"\n" + v
		}
	}
	if layout>>1 == 1 {
		v += ","
	}

	if ts && len(p.EolComment) != 0 {
		v += " " + p.EolComment
		if layout>>1 > 0 {
			v += "\n" + indents
		}
	}

	return v
}

func (p *Value) typeArrayString(layout int, indent int) string {
	a := p.v.([]*Value)
	s := ""
	max := len(a) - 1
	nlayout := layout & 1
	for i, bv := range a {
		if i != max {
			s += bv.string(nlayout|2, indent) // for "," # comment
		} else {
			s += bv.string(nlayout|6, indent)
		}
	}
	if layout != 0 {
		return p.comments("["+s+"]", layout, indent)
	}
	return "[" + s + "]"
}
