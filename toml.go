package toml

import (
	"errors"
	"io/ioutil"
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
		if it == nil || !it.IsValid() {
			continue
		}
		pos := strings.LastIndex(key, ".")

		if pos == -1 && it.kind < Table {
			keys = append(keys, key)
			keyidx = append(keyidx, it.idx)
			it.key = key
		} else {
			if it.kind < Table {
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
// such as:
//	p.Fetch("")        // returns all valid elements in p
//	p.Fetch("prefix")
// 从 Toml 中提取出 prefix 开头的所有 Table 元素, 返回值也是一个 Toml.
// 注意返回值是原 Toml 的子集, 没有进行克隆.
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

// TablesName returns all name of Table/ArrayOfTables
// 返回 Toml 包含的所有 Table/ArrayOfTables的名称.
func (p Toml) TablesName() []string {
	var re []string
	for key, it := range p {
		if it != nil && it.IsValid() && (it.kind == Table || it.kind == ArrayOfTables) {
			re = append(re, key)
		}
	}
	return re
}

func (p Toml) newItem(k Kind) *Item {
	return NewItem(k)
}

var (
	InValidFormat = errors.New("invalid TOML format")
	Redeclared    = errors.New("duplicate definitionin TOML")
)

// 从 TOML 格式 source 解析出 Toml 对象.
func Parse(source []byte) (tm Toml, err error) {
	var (
		it, iv        *Item
		path, key, fc string // fc is Multi-line comments cache
		handler       TokenHandler
		flag          Token
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
					ts := it.Tables(-1)
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
				nit := iv.Value.Index(-1)
				if nit != nil {
					nit.EolComment = s
					break
				}
			}
			println(flag.String())
			err = InternalError

		case tokenString, tokenInteger, tokenFloat, tokenBoolean, tokenDatetime:
			if token == tokenString {
				s = string([]byte(s)[1 : len(s)-1])
			}
			// plain value
			if iv.kind == InvalidKind {
				err = iv.SetAs(s, Kind(token))
				break
			}
			// plain Array or typeArray
			if iv.kind >= StringArray && iv.kind <= Array {
				nit := tm.newItem(0)
				nit.MultiComments, fc = fc, ""

				err = nit.SetAs(s, Kind(token))
				if err == nil {
					err = iv.Add(nit)
				}
				break
			}
			err = InValidFormat

		case tokenTable:
			it = tm.newItem(Table)
			iv = nil

			path = string(s[1 : len(s)-1])

			tmp, ok := tm[path]
			if ok {
				it = tmp
				if it.kind != Table {
					err = Redeclared
					break
				}
			} else {
				tm[path] = it
			}
			if it.MultiComments != "" && fc != "" {
				it.MultiComments, fc = it.MultiComments+"\n"+fc, ""
			} else {
				it.MultiComments, fc = fc, ""
			}

		case tokenArrayOfTables:
			iv = nil

			path = string(s[2 : len(s)-2])

			tmp, ok := tm[path]
			if !ok {

				it = tm.newItem(ArrayOfTables)
				err = it.AddTables(Tables{})
				if err != nil {
					break
				}
				tm[path] = it

			} else {

				it = tmp
				if it.kind != ArrayOfTables {
					err = Redeclared
					break
				}
				it.AddTables(Tables{})

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
			iv = tm.newItem(0)
			iv.key = s
			if path != "" {
				key = path + "." + s
			} else {
				key = s
			}

			if it != nil && it.kind == ArrayOfTables {

				ts := it.Tables(-1)
				if ts == nil {
					err = InternalError
					break
				}
				ts[s] = &iv.Value

			} else {
				tm[key] = iv
			}
			iv.MultiComments, fc = fc, ""

		case tokenArrayLeftBrack:
			if it.kind == ArrayOfTables {
				err = InValidFormat
				break
			}
			// [[..],[<-..]]
			if iv.kind == Array && tm[key] == iv {
				iv = tm.newItem(Array)
				break
			}
			if iv.kind == InvalidKind { // [
				iv.kind = Array
			} else if iv.kind == Array { // [[
				if iv.v != nil {
					err = InternalError
					break
				}
				iv = tm.newItem(Array) // new iv
			}

		case tokenArrayRightBrack:
			if it != nil && it.kind == ArrayOfTables {

				ts := it.Tables(-1)
				if ts == nil {
					err = InternalError
					break
				}
				nit := ts[key]
				if nit == &iv.Value {
					break
				}
				err = nit.Add(&iv.Value)
				if err == nil {
					iv = tm.newItem(Array)
				}

			} else {
				// [[..->]]
				nit := tm[key]
				if nit != iv {
					err = nit.Add(iv)
					if err == nil {
						iv = nit // nit == iv
					}
					break
				}
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
