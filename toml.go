package toml

import (
	"errors"
	"io/ioutil"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Toml 是一个 maps, 不是 tree 实现.
type Toml map[string]Item

func New() Toml {
	tm := Toml{}
	tm[iD] = MakeItem(0)
	return tm
}

func (tm Toml) safeId() int {
	id, ok := tm[iD]
	if !ok || id.idx == 0 {
		tm[iD] = MakeItem(0)
	}
	return tm[iD].idx
}

/**
IdValue 返回用于管理的 ".ID..." 对象副本.
*/
func (tm Toml) IdValue() Value {
	it, ok := tm[iD]
	if !ok {
		return Value{}
	}
	return *it.Value
}

// String returns TOML layout string
// 格式化输出带缩进的 TOML 格式.
func (p Toml) String() string {
	return p.string(-1, 1)
}
func (p Toml) string(indent, offset int) (fmt string) {
	l := len(p)
	if l == 0 {
		return
	}
	var keys, values []string
	var keyidx, tabidx []int
	tables := map[int]string{}
	lastcomments := ""
	for key, it := range p {
		if len(key) == 0 {
			lastcomments = FixComments(it.MultiComments)
			continue
		}
		if !it.IsValid() {
			continue
		}
		pos := strings.LastIndex(key, ".")

		if pos == -1 && it.kind < TableName {
			keys = append(keys, key)
			keyidx = append(keyidx, it.idx)
			it.key = key
		} else {
			if it.kind < TableName {
				values = append(values, key)
				it.key = key[pos+1:]
			} else {
				tables[it.idx] = key
				tabidx = append(tabidx, it.idx)
				it.key = key
			}
		}
	}
	sort.Sort(sort.IntSlice(keyidx))
	sort.Sort(sort.IntSlice(tabidx))
	sort.Sort(sort.StringSlice(values))

	for i, _ := range keyidx {
		key := keys[i]
		it := p[key]
		fmt += it.string(1, indent)
	}

	for _, idx := range tabidx {
		key := tables[idx]
		it := p[key]
		if len(fmt) != 0 {
			fmt += "\n"
		}
		count := indent
		ks := ""
		for _, k := range strings.Split(key, ".") {
			if ks == "" {
				ks = k
			} else {
				ks += "." + k
			}
			_, ok := p[ks]
			if ok {
				count += offset
			}
		}
		fmt += it.string(1, count)
		var i, l int
		var vk string
		key = key + "."
		l = len(key)
		f := false
		dots := strings.Count(key, ".")
		for i, vk = range values {
			if len(vk) < l || vk[:l] != key {
				if f {
					break
				}
				continue
			}
			if strings.Count(vk, ".") != dots {
				continue
			}
			f = true
			values[i] = ""
			fmt += p[vk].string(1, count+1)
		}
	}
	if len(lastcomments) != 0 {
		fmt += "\n" + lastcomments
	}
	return
}

// Fetch returns Sub Toml of p, and reset name. not clone.

/**
such as:
	p.Fetch("")       // returns all valid elements in p
	p.Fetch("prefix") // same as p.Fetch("prefix.")
从 Toml 中提取出 prefix 开头的所有 Table 元素, 返回值也是一个 Toml.
注意:
	返回值是原 Toml 的子集.
	返回子集中不包括 [prefix] TableName.
	对返回子集添加 *Item 不会增加到原 Toml 中.
	对返回子集中的 *Item 进行更新, 原 Toml 也会更新.
	子集中不会含有 ArrayOfTables 类型数据.
*/
func (p Toml) Fetch(prefix string) Toml {
	nt := Toml{}
	ln := len(prefix)
	if ln != 0 {
		if prefix[ln-1] != '.' {
			prefix += "."
			ln++
		}
	}

	for key, it := range p {
		if !it.IsValid() || strings.Index(key, prefix) != 0 {
			continue
		}
		newkey := key[ln:]
		if newkey == "" {
			continue
		}
		nt[newkey] = it
	}
	return nt
}

// TableNames returns all name of TableName.
// 返回所有 TableName 的名字和 ArrayOfTables 的名字.
func (p Toml) TableNames() (tableNames []string, arrayOfTablesNames []string) {
	for key, it := range p {
		if it.IsValid() {
			if it.kind == TableName {
				tableNames = append(tableNames, key)
			} else if it.kind == ArrayOfTables {
				arrayOfTablesNames = append(arrayOfTablesNames, key)
			}
		}
	}
	return
}

// Apply to each field in the struct, case sensitive.
/**
Apply 把 p 存储的值赋给 dst , TypeOf(dst).Kind() 为 reflect.Struct, 返回赋值成功的次数.
*/
func (p Toml) Apply(dst interface{}) (count int) {
	var (
		vv reflect.Value
		ok bool
	)

	vv, ok = dst.(reflect.Value)
	if ok {
		vv = reflect.Indirect(vv)
	} else {
		vv = reflect.Indirect(reflect.ValueOf(dst))
	}
	return p.apply(vv)
}

func (p Toml) apply(vv reflect.Value) (count int) {

	var it Item
	vt := vv.Type()
	if !vv.IsValid() || !vv.CanSet() || vt.Kind() != reflect.Struct || vt.String() == "time.Time" {
		return
	}

	for i := 0; i < vv.NumField(); i++ {
		name := vt.Field(i).Name
		it = p[name]

		if !it.IsValid() {
			continue
		}

		if it.kind == TableName {
			count += p.Fetch(name).Apply(vv.Field(i))
		} else {
			count += it.apply(vv.Field(i))
		}
	}
	return
}

var (
	InValidFormat = errors.New("invalid TOML format")
	Redeclared    = errors.New("duplicate definitionin")
)

// 从 TOML 格式 source 解析出 Toml 对象.
func Parse(source []byte) (tm Toml, err error) {
	p := &parse{Scanner: NewScanner(source)}

	tb := NewBuilder(nil)

	p.Handler(
		func(token Token, str string) error {
			tb, err = tb.Token(token, str)
			return err
		})

	p.Run()
	tm = tb.Toml()
	return
}

// 如果 p!=nil 表示是子集模式, tablename 必须有相同的 prefix
type TomlBuilder struct {
	tm     Toml
	root   *TomlBuilder
	p      *TomlBuilder
	it     *Item
	iv     *Value
	c      string // comment
	prefix string // with "."
	token  Token
}

func NewBuilder(root *TomlBuilder) TomlBuilder {
	tb := TomlBuilder{}

	tb.tm = Toml{}
	tb.tm.safeId()

	if root == nil {
		tb.root = &tb
	} else {
		tb.root = root
		tb.token = tb.root.token
	}
	return tb
}

func (t TomlBuilder) Toml() Toml {
	return t.tm
}

func (t TomlBuilder) Token(token Token, str string) (TomlBuilder, error) {
	defer func() {
		// 缓存上一个 token, EolComment 等需要用
		if token == tokenWhitespace {
			return
		}

		t.root.token = token

		if token != tokenComment && token != tokenNewLine {
			t.token = token
		}
	}()
	switch token {
	case tokenError:
		return t.Error(str)
	case tokenRuneError:
		return t.RuneError(str)
	case tokenEOF:
		return t.EOF(str)
	case tokenWhitespace:
		return t.Whitespace(str)
	case tokenEqual:
		return t.Equal(str)
	case tokenNewLine:
		return t.NewLine(str)
	case tokenComment:
		return t.Comment(str)
	case tokenString:
		return t.String(str)
	case tokenInteger:
		return t.Integer(str)
	case tokenFloat:
		return t.Float(str)
	case tokenBoolean:
		return t.Boolean(str)
	case tokenDatetime:
		return t.Datetime(str)
	case tokenTableName:
		return t.TableName(str)
	case tokenArrayOfTables:
		return t.ArrayOfTables(str)
	case tokenKey:
		return t.Key(str)
	case tokenArrayLeftBrack: // [
		return t.ArrayLeftBrack(str)
	case tokenArrayRightBrack: // ]
		return t.ArrayRightBrack(str)
	case tokenComma:
		return t.Comma(str)
	}
	return t, NotSupported
}

func (t TomlBuilder) Error(str string) (TomlBuilder, error) {
	return t, errors.New(str)
}

func (t TomlBuilder) RuneError(str string) (TomlBuilder, error) {
	return t, errors.New(str)
}

func (t TomlBuilder) EOF(str string) (TomlBuilder, error) {
	return *t.root, nil
}

func (t TomlBuilder) Whitespace(str string) (TomlBuilder, error) {
	return t, nil
}

func (t TomlBuilder) NewLine(str string) (TomlBuilder, error) {
	return t, nil
}

func (t TomlBuilder) Comment(str string) (TomlBuilder, error) {

	// EolComment
	if t.root.token != tokenEOF && t.root.token != tokenNewLine {

		if t.c != "" {
			return t, InternalError
		}

		if t.iv == nil {
			if t.it.EolComment == "" {
				t.it.EolComment, t.c = str, ""
			}
		} else if t.iv.EolComment == "" {
			t.iv.EolComment, t.c = str, ""
		} else {
			return t, InternalError
		}

		return t, nil
	}

	// 给多行注释添加换行
	l := len(t.c)
	if l > 0 && t.c[l-1] != '\n' {
		t.c += "\n" + str
	} else {
		t.c = str
	}
	return t, nil
}

func (t TomlBuilder) String(str string) (TomlBuilder, error) {
	if t.iv == nil {
		return t, InternalError
	}

	str, err := strconv.Unquote(str)
	if err != nil {
		return t, err
	}

	if t.iv.kind != Array && t.iv.kind != StringArray {
		return t, t.iv.SetAs(str, String)
	}
	return t, t.iv.Add(str)
}

func (t TomlBuilder) Integer(str string) (TomlBuilder, error) {
	if t.iv == nil {
		return t, InternalError
	}

	if t.iv.kind != Array && t.iv.kind != IntegerArray {
		return t, t.iv.SetAs(str, Integer)
	}
	v, err := conv(str, Integer)
	if err != nil {
		return t, err
	}
	return t, t.iv.Add(v)
}
func (t TomlBuilder) Float(str string) (TomlBuilder, error) {
	if t.iv == nil {
		return t, InternalError
	}

	if t.iv.kind != Array && t.iv.kind != FloatArray {
		return t, t.iv.SetAs(str, Float)
	}
	v, err := conv(str, Float)
	if err != nil {
		return t, err
	}
	return t, t.iv.Add(v)
}
func (t TomlBuilder) Boolean(str string) (TomlBuilder, error) {
	if t.iv == nil {
		return t, InternalError
	}

	if t.iv.kind != Array && t.iv.kind != BooleanArray {
		return t, t.iv.SetAs(str, Boolean)
	}
	v, err := conv(str, Boolean)
	if err != nil {
		return t, err
	}
	return t, t.iv.Add(v)
}
func (t TomlBuilder) Datetime(str string) (TomlBuilder, error) {
	if t.iv == nil {
		return t, InternalError
	}

	if t.iv.kind != Array && t.iv.kind != DatetimeArray {
		return t, t.iv.SetAs(str, Datetime)
	}
	v, err := conv(str, Datetime)
	if err != nil {
		return t, err
	}
	return t, t.iv.Add(v)
}

func (t TomlBuilder) TableName(str string) (TomlBuilder, error) {
	path := str[1 : len(str)-1]

	it, ok := t.tm[path]
	if ok {
		return t, Redeclared
	}

	comment := t.c
	t.c = ""

	if t.prefix != "" {
		if t.p == nil {
			return t, InternalError
		}

		if path == t.prefix {
			return t, Redeclared
		}

		if !strings.HasPrefix(path, t.prefix+".") {
			t = *t.p
			t.c = comment
			return t.TableName(str)
		}
	}

	it = MakeItem(TableName)
	it.MultiComments = comment
	it.key = path

	t.tm[path] = it
	t.it = &it
	t.iv = nil
	return t, nil
}

func (t TomlBuilder) Key(str string) (TomlBuilder, error) {

	it := MakeItem(0)
	it.key = str
	it.MultiComments, t.c = t.c, ""

	if t.it != nil {
		str = t.it.key + "." + str
	}

	t.tm[str] = it
	t.iv = it.Value

	return t, nil
}

func (t TomlBuilder) Equal(str string) (TomlBuilder, error) {

	if t.root.token != tokenKey {
		return t, InValidFormat
	}

	if t.iv == nil {
		return t, InternalError
	}
	return t, nil
}

func (t TomlBuilder) ArrayOfTables(str string) (nt TomlBuilder, err error) {
	path := str[2 : len(str)-2]

	if t.prefix != "" {
		if t.p == nil {
			return t, InternalError
		}

		comment := t.c

		// 增加兄弟 table
		if t.prefix == path {
			t = *t.p
			t.c = comment
			return t.ArrayOfTables(str)
		} else if !strings.HasPrefix(path, t.prefix+".") {
			// 递归向上
			t = *t.p
			t.c = comment
			return t.ArrayOfTables(str)
		}
	}

	nt, err = t.nestToml(path)
	if err != nil {
		return t, err
	}

	return nt, err
}

// 嵌套 TOML , prefix 就是 [[arrayOftablesName]]
func (t TomlBuilder) nestToml(prefix string) (TomlBuilder, error) {

	it, ok := t.tm[prefix]

	// [[foo.bar]] 合法性检查不够完全??????
	if ok && it.kind != ArrayOfTables {
		return t, Redeclared
	}

	tb := NewBuilder(t.root)

	tb.p = &t
	tb.tm = Toml{}
	tb.tm.safeId()

	// first [[...]]
	if !ok {
		it = MakeItem(ArrayOfTables)
		it.v = Tables{tb.tm}
		t.tm[prefix] = it
	} else {
		// again [[...]]
		ts := it.v.(Tables)
		it.v = append(ts, tb.tm)
	}
	it.MultiComments, t.c = t.c, ""

	tb.prefix = prefix
	return tb, nil
}

func (t TomlBuilder) ArrayLeftBrack(str string) (TomlBuilder, error) {
	if t.iv == nil {
		return t, NotSupported
	}

	if t.iv.kind == InvalidKind {
		t.iv.kind = Array
		return t, nil
	}
	if t.iv.kind != Array {
		return t, NotSupported
	}

	nt := t
	nt.iv = NewValue(Array)
	nt.p = &t
	t.iv.Add(nt.iv)
	return nt, nil
}

func (t TomlBuilder) ArrayRightBrack(str string) (TomlBuilder, error) {

	if t.iv == nil || t.iv.kind < StringArray || t.iv.kind > Array {
		return t, InValidFormat
	}

	if t.p == nil {
		return t, nil
	}
	return *t.p, nil
}

func (t TomlBuilder) Comma(str string) (TomlBuilder, error) {
	if t.iv == nil || t.iv.kind < StringArray || t.iv.kind > Array {
		return t, InValidFormat
	}

	return t, nil
}

// Create a Toml from a file.
// 便捷方法, 从 TOML 文件解析出 Toml 对象.
func LoadFile(path string) (toml Toml, err error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	toml, err = Parse(source)
	return
}
