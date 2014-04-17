package toml

import (
	"errors"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
)

// Toml 是一个 maps, 不是 tree 实现.
type Toml map[string]*Item

// String returns TOML layout string
// 格式化输出带缩进的 TOML 格式.
func (p Toml) String() (fmt string) {
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
		fmt += it.string(1, -1)
	}

	for _, idx := range tabidx {
		key := tables[idx]
		it := p[key]
		if len(fmt) != 0 {
			fmt += "\n"
		}
		count := -1
		ks := ""
		for _, k := range strings.Split(key, ".") {
			if ks == "" {
				ks = k
			} else {
				ks += "." + k
			}
			_, ok := p[ks]
			if ok {
				count++
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

// Fetch returns subset of p, and reset name. not clone.
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
		if it == nil || !it.IsValid() || strings.Index(key, prefix) != 0 {
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

	var it *Item
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

func (p Toml) newItem(k Kind) *Item {
	return NewItem(k)
}

func (p Toml) newValue(k Kind) *Value {
	return NewValue(k)
}

var (
	InValidFormat = errors.New("invalid TOML format")
	Redeclared    = errors.New("duplicate definitionin TOML")
)

// 从 TOML 格式 source 解析出 Toml 对象.
func Parse(source []byte) (tm Toml, err error) {
	var (
		it            *Item
		iv            *Value
		path, key, fc string // fc is Multi-line comments cache
		handler       TokenHandler
		flag          Token
		lastKey       string
	)

	p := &parse{Scanner: NewScanner(source)}
	tm = Toml{}

	handler = func(token Token, s string) error {
		switch token {
		case tokenError, tokenRuneError:
			err = errors.New(s)
		case tokenEOF:
			return nil
		case tokenWhitespace, tokenEqual:
			return nil
		case tokenNewLine:

		case tokenComment:
			token = flag
			// front comments
			if flag == tokenNewLine || (it == nil && iv == nil) {
				if fc == "" {
					fc = s
				} else {
					fc += "\n" + s
				}
				break
			}
			if iv == nil {
				it.EolComment = s
				break
			}
			// plain value or
			/*
				[ # comment
				1,
				2,
				] # comment
			*/
			if iv.kind < StringArray || flag == tokenArrayLeftBrack {
				iv.EolComment = s
				break
			}
			if flag == tokenArrayRightBrack {
				if it.kind == ArrayOfTables {
					ts := it.Table(-1)
					if ts != nil && ts[key] != nil {
						ts[key].EolComment = s
					}
					break
				}
				iv.EolComment = s
				break
			}
			/*
				[1, # comment
				2,
				]
			*/
			if flag == tokenComma {
				nit := iv.Index(-1)
				if nit != nil {
					nit.EolComment = s
					break
				}
			}
			err = InternalError

		case tokenString, tokenInteger, tokenFloat, tokenBoolean, tokenDatetime:
			if token == tokenString {
				s = string([]byte(s)[1 : len(s)-1])
			}
			if iv == nil {
				err = InternalError
				break
			}
			// plain value
			if iv.kind == InvalidKind {
				err = iv.SetAs(s, Kind(token))
				break
			}
			// plain Array or typeArray
			if iv.kind >= StringArray && iv.kind <= Array {
				nit := tm.newValue(0)
				nit.MultiComments, fc = fc, ""

				err = nit.SetAs(s, Kind(token))
				if err == nil {
					err = iv.Add(nit)
				}
				break
			}
			err = InValidFormat

		case tokenTableName:
			it = tm.newItem(TableName)
			iv = nil

			path = string(s[1 : len(s)-1])

			_, ok := tm[path]
			if ok {
				err = Redeclared
				break
			} else {
				tm[path] = it
			}

			it.MultiComments, fc = fc, ""

		case tokenArrayOfTables:
			path = string(s[2 : len(s)-2])
			iv = nil

			tmp, ok := tm[path]
			if !ok {
				it = tm.newItem(ArrayOfTables)
				tm[path] = it

			} else {
				it = tmp
				if it.kind != ArrayOfTables {
					err = Redeclared
					break
				}
			}

			err = it.AddTable(Table{}) // alow empty table

			if err != nil {
				break
			}

			if it.MultiComments != "" && fc != "" {
				it.MultiComments, fc = it.MultiComments+"\n"+fc, ""
			} else {
				it.MultiComments, fc = fc, ""
			}

		case tokenKey:
			if s == "" {
				err = NotSupported
				break
			}
			lastKey = s
			if path != "" {
				key = path + "." + s // table -> key
			} else {
				key = s // top key
			}

			if it != nil && it.kind == ArrayOfTables {
				t := it.Table(-1)
				if t == nil {
					err = InternalError
					break
				}

				iv = tm.newValue(0)
				iv.key = s

				t[s] = iv

			} else {
				nit := tm.newItem(0)
				nit.key = s

				tm[key] = nit
				iv = &nit.Value
			}

			iv.MultiComments, fc = fc, ""

		case tokenArrayLeftBrack: // [
			// TOML 未明确数组的维度限制, 其示例用了 2 级深度
			if iv != nil {
				if iv.kind == InvalidKind {
					iv.kind = Array
				} else {
					niv := tm.newValue(Array)
					err = iv.Add(niv)
					if err == nil {
						iv = niv
					}
				}
				break
			}

			iv = tm.newValue(Array)

			t := it.Table(-1)
			if t == nil {
				err = InternalError
				break
			}

			tiv := t[lastKey]
			if !tiv.IsValid() || tiv.kind != Array {
				err = InternalError
				break
			}

			tiv.Add(iv)

		case tokenArrayRightBrack: // ]

			if iv == nil || it == nil {
				err = InternalError
				break
			}
			var tiv *Value

			if it.kind == ArrayOfTables {

				t := it.Table(-1)
				if t == nil {
					err = InternalError
					break
				}
				tiv = t[lastKey]

			} else if !iv.IsValid() ||
				iv.kind < StringArray ||
				iv.kind > Array {

				err = InternalError
				break
			} else {
				// fetch table.key
				tit := tm[key]
				if !tit.IsValid() {
					err = InternalError
					break
				}
				tiv = &tit.Value
			}

			if tiv.Len() <= 0 {
				println("iv == nil || it == nil", lastKey)
				err = InternalError
				break
			}
			// 恢复 iv 的上级
			if tiv.idx == iv.idx {
				break
			}

			idx := iv.idx
			err = InternalError
			for tiv.Len() > 0 && tiv != nil {

				iv = tiv.Index(tiv.Len() - 1)

				if iv.IsValid() && iv.idx == idx {
					iv = tiv
					err = nil
					break
				}

				tiv = iv
			}

		case tokenComma:
			if iv == nil || iv.kind < StringArray || iv.kind > Array {
				err = InValidFormat
			}
		default:
			return NotSupported
		}
		flag = token
		return err
	}
	p.Handler(handler)
	r := recEmpty(p)
	for r != nil {
		r = r(p)
	}

	if err != nil {
		println(flag.String(), "===============", err.Error())
		s := p.Scanner.(*scanner)
		println(s.offset, s.size, string(s.buf))
	}
	// last comments
	if fc != "" {
		it = tm.newItem(0)
		it.MultiComments = fc
		tm[""] = it
	}
	return
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
