package toml

import (
	"errors"
	"fmt"
	"strings"
)

type Status int

const (
	SNot Status = iota
	SMaybe
	SYes
	SInvalid
)

const (
	BOM = 0xFEFF // UTF-8 encoded byte order mark
)

type Token uint

// don't change order
const (
	tokenEOF Token = iota
	tokenString
	tokenInteger
	tokenFloat
	tokenBoolean
	tokenDatetime
	tokenWhitespace
	tokenComment
	tokenTableName
	tokenArrayOfTables
	tokenNewLine
	tokenKey
	tokenEqual
	tokenArrayLeftBrack
	tokenArrayRightBrack
	tokenComma
	tokenError
	tokenRuneError
	lameValue // for lame
	lameArray
)

func (t Token) String() string {
	return tokensName[t]
}

var tokensName = [...]string{
	"EOF",
	"String",
	"Integer",
	"Float",
	"Boolean",
	"Datetime",
	"Whitespace",
	"Comment",
	"TableName",
	"ArrayOfTables",
	"NewLine",
	"Key",
	"Equal",
	"ArrayLeftBrack",
	"ArrayRightBrack",
	"Comma",
	"Error",
	"EncodingError",
	"Value",
	"Array",
}

type TokenHandler func(Token, string) error
type receptor func(parser) receptor
type tokenFn func(r rune, flag Token, maybe bool) (Status, Token)

type parser interface {
	Scanner
	Token(token Token) error
	NotMatch(token Token)
	Invalid(token Token)
	Unexpected(token Token)
	ArrayDimensions(add int) int
	Handler(TokenHandler)
}

type parse struct {
	Scanner
	arrayDimensions int
	err             error
	handler         TokenHandler
}

func (p *parse) run() {
	r := recEmpty(p)
	for r != nil {
		r = r(p)
	}
}
func (p *parse) NotMatch(token Token) {
	p.err = errors.New("incomplete " + tokensName[token])
	p.Token(tokenError)
}
func (p *parse) Invalid(token Token) {
	p.err = errors.New("invalid " + tokensName[token])
	p.Token(tokenError)
}
func (p *parse) Unexpected(token Token) {
	p.err = errors.New("unexpected token " + tokensName[token])
	p.Token(tokenError)
}
func (p *parse) Token(token Token) (err error) {
	var str string
	if token == tokenError {
		err = p.err
		str = err.Error()
	} else {
		if token != tokenEOF {
			str = strings.TrimSpace(p.Get())
		}
	}
	if p.handler == nil {
		fmt.Println(tokensName[token], str)
	} else {
		if token == tokenError {
			p.handler(token, str)
		} else {
			err = p.handler(token, str)
		}
	}
	return
}

// Handler to set TokenHandler
func (p *parse) Handler(h TokenHandler) {
	p.handler = h
}
func (p *parse) ArrayDimensions(add int) int {
	if add < 0 {
		add = -1
	}
	if add > 0 {
		add = 1
	}
	p.arrayDimensions += add
	return p.arrayDimensions
}

var tokensEmpty, tokensEqual, tokensTable, tokensArrayOfTables,
	tokensValues,
	tokensArray,
	tokensStringArray,
	tokensBooleanArray,
	tokensIntegerArray,
	tokensFloatArray,
	tokensDatetimeArray []tokenFn

var stateEmpty, stateEqual, stateTable, stateArrayOfTables,
	stateValues,
	stateArray,
	stateStringArray,
	stateBooleanArray,
	stateIntegerArray,
	stateFloatArray,
	stateDatetimeArray []receptor

func init() {
	tokensEmpty = []tokenFn{
		itsWhitespace,    // recEmpty
		itsNewLine,       // recEmpty
		itsComment,       // recEmpty
		itsTableName,     // recTable
		itsArrayOfTables, // recArrayOfTables
		itsKey,           // recEqual
	}
	stateEmpty = []receptor{
		recEmpty,         // itsWhitespace
		recEmpty,         // itsNewLine
		recEmpty,         // itsComment
		recEmpty,         // itsTableName
		recArrayOfTables, // itsArrayOfTables
		recEqual,         // itsKey
	}

	tokensEqual = []tokenFn{
		itsWhitespace, // recEqual
		itsEqual,      // recValues
	}
	stateEqual = []receptor{
		recEqual,  // itsWhitespace
		recValues, // itsEqual
		recNotMatch(tokenEqual),
	}

	tokensTable = []tokenFn{
		itsWhitespace,    // recEmpty
		itsComment,       // recEmpty
		itsNewLine,       // recEmpty
		itsKey,           // recEqual
		itsTableName,     // recTable
		itsArrayOfTables, // recArrayOfTables
	}
	stateTable = []receptor{
		recEmpty,         // itsWhitespace
		recEmpty,         // itsComment
		recEmpty,         // itsNewLine
		recEqual,         // itsKey
		recTable,         // itsTableName
		recArrayOfTables, // itsArrayOfTables
	}

	tokensArrayOfTables = []tokenFn{
		itsWhitespace, // recArrayOfTables
		itsComment,    // recArrayOfTables
		itsNewLine,    // recEmpty
		itsKey,        // recEqual
		itsTableName,  // recEmptyTable
	}
	stateArrayOfTables = []receptor{
		recArrayOfTables, // itsWhitespace
		recArrayOfTables, // itsComment
		recEmpty,         // itsNewLine
		recEqual,         // itsKey
		recEmptyTable,    // itsTableName
	}

	tokensValues = []tokenFn{
		itsWhitespace,     // recValues
		itsArrayLeftBrack, // recArrayLeftBrack
		itsString,         // recEmpty
		itsBoolean,        // recEmpty
		itsInteger,        // recEmpty
		itsFloat,          // recEmpty
		itsDatetime,       // recEmpty
	}
	stateValues = []receptor{
		recValues,         // itsWhitespace
		recArrayLeftBrack, // itsArrayLeftBrack
		recEmpty,          // itsString
		recEmpty,          // itsBoolean
		recEmpty,          // itsInteger
		recEmpty,          // itsFloat
		recEmpty,          // itsDatetime
		recNotMatch(lameValue),
	}

	tokensArray = []tokenFn{
		itsWhitespace,      // recArray
		itsNewLine,         // recArray
		itsComment,         // recArray
		itsComma,           // recArray
		itsString,          // recStringArray
		itsBoolean,         // recBooleanArray
		itsInteger,         // recIntegerArray
		itsFloat,           // recFloatArray
		itsDatetime,        // recDatetimeArray
		itsArrayLeftBrack,  // recArrayLeftBrack
		itsArrayRightBrack, // recArrayRightBrack
	}
	stateArray = []receptor{
		recArray,           // itsWhitespace
		recArray,           // itsNewLine
		recArray,           // itsComment
		recArray,           // itsComma
		recStringArray,     // itsString
		recBooleanArray,    // itsBoolean
		recIntegerArray,    // itsInteger
		recFloatArray,      // itsFloat
		recDatetimeArray,   // itsDatetime
		recArrayLeftBrack,  // itsArrayLeftBrack
		recArrayRightBrack, // itsArrayRightBrack
		recNotMatch(lameArray),
	}

	tokensStringArray = []tokenFn{
		itsWhitespace,      // recStringArray
		itsNewLine,         // recStringArray
		itsComment,         // recStringArray
		itsComma,           // recStringArray
		itsArrayRightBrack, // recArrayRightBrack
		itsString,          // recStringArray
	}
	stateStringArray = []receptor{
		recStringArray,     // itsWhitespace
		recStringArray,     // itsNewLine
		recStringArray,     // itsComment
		recStringArray,     // itsComma
		recArrayRightBrack, // itsArrayRightBrack
		recStringArray,     // itsString
		recNotMatch(lameArray),
	}

	tokensBooleanArray = []tokenFn{
		itsWhitespace,      // recBooleanArray
		itsNewLine,         // recBooleanArray
		itsComment,         // recBooleanArray
		itsComma,           // recBooleanArray
		itsArrayRightBrack, // recArrayRightBrack
		itsBoolean,         // recBooleanArray
	}
	stateBooleanArray = []receptor{
		recBooleanArray,    // itsWhitespace
		recBooleanArray,    // itsNewLine
		recBooleanArray,    // itsComment
		recBooleanArray,    // itsComma
		recArrayRightBrack, // itsArrayRightBrack
		recBooleanArray,    // itsBoolean
		recNotMatch(lameArray),
	}

	tokensIntegerArray = []tokenFn{
		itsWhitespace,      // recIntegerArray
		itsNewLine,         // recIntegerArray
		itsComment,         // recIntegerArray
		itsComma,           // recIntegerArray
		itsArrayRightBrack, // recArrayRightBrack
		itsInteger,         // recIntegerArray
	}
	stateIntegerArray = []receptor{
		recIntegerArray,    // itsWhitespace
		recIntegerArray,    // itsNewLine
		recIntegerArray,    // itsComment
		recIntegerArray,    // itsComma
		recArrayRightBrack, // itsArrayRightBrack
		recIntegerArray,    // itsInteger
		recNotMatch(lameArray),
	}

	tokensFloatArray = []tokenFn{
		itsWhitespace,      // recFloatArray
		itsNewLine,         // recFloatArray
		itsComment,         // recFloatArray
		itsComma,           // recFloatArray
		itsArrayRightBrack, // recArrayRightBrack
		itsFloat,           // recFloatArray
	}
	stateFloatArray = []receptor{
		recFloatArray,      // itsWhitespace
		recFloatArray,      // itsNewLine
		recFloatArray,      // itsComment
		recFloatArray,      // itsComma
		recArrayRightBrack, // itsArrayRightBrack
		recFloatArray,      // itsFloat
	}

	tokensDatetimeArray = []tokenFn{
		itsWhitespace,      // recDatetimeArray
		itsNewLine,         // recDatetimeArray
		itsComment,         // recDatetimeArray
		itsComma,           // recDatetimeArray
		itsArrayRightBrack, // recArrayRightBrack
		itsDatetime,        // recDatetimeArray
	}
	stateDatetimeArray = []receptor{
		recDatetimeArray,   // itsWhitespace
		recDatetimeArray,   // itsNewLine
		recDatetimeArray,   // itsComment
		recDatetimeArray,   // itsComma
		recArrayRightBrack, // itsArrayRightBrack
		recDatetimeArray,   // itsDatetime
	}

}
func rec(fns []tokenFn, rs []receptor, p parser) receptor {
	var (
		st    Status
		flag  Token
		maybe int
		r     rune
	)
	flags := make([]Token, len(fns))
	skips := make([]bool, len(fns))
	for {
		if p.Eof() {
			break
		}
		r = p.Next()

		if r == RuneError {
			p.Invalid(tokenRuneError)
			return nil
		}
		for i, skip := range skips {
			if skip {
				continue
			}
			switch st, flag = fns[i](r, flags[i], maybe != 0); st {
			case SYes:
				if p.Token(flag) != nil {
					return nil
				}
				if r == EOF {
					p.Token(tokenEOF)
				}
				return rs[i]

			case SNot:
				if flags[i] != 0 {
					maybe--
				}
				skips[i] = true

			case SInvalid:
				maybe--
				p.Invalid(flag)
				return nil

			case SMaybe:
				if flags[i] == 0 {
					maybe++
				}
				flags[i] = flag
			}
		}
		if 0 == maybe {
			break
		}
		if r == EOF {
			if len(rs) > len(fns) {
				return rs[len(fns)]
			}
			p.Token(tokenEOF)
			return nil
		}
	}
	if len(rs) > len(fns) {
		return rs[len(fns)]
	}
	return nil
}
func recUnexpected(token Token) receptor {
	return func(p parser) receptor {
		p.Unexpected(token)
		return nil
	}
}
func recNotMatch(token Token) receptor {
	return func(p parser) receptor {
		p.NotMatch(token)
		return nil
	}
}

func recArrayLeftBrack(p parser) receptor {
	deep := p.ArrayDimensions(1)
	if deep <= 2 {
		return rec(tokensArray, stateArray, p)
	}
	p.Unexpected(tokenArrayLeftBrack)
	return nil
}
func recArrayRightBrack(p parser) receptor {
	deep := p.ArrayDimensions(-1)
	if deep < 0 {
		p.Unexpected(tokenArrayRightBrack)
		return nil
	}
	if deep == 0 {
		return rec(tokensEmpty, stateEmpty, p)
	}
	return rec(tokensArray, stateArray, p)
}

func recEmpty(p parser) receptor {
	return rec(tokensEmpty, stateEmpty, p)
}

func recEqual(p parser) receptor {
	return rec(tokensEqual, stateEqual, p)
}

func recTable(p parser) receptor {
	return rec(tokensTable, stateTable, p)
}

func recArrayOfTables(p parser) receptor {
	return rec(tokensArrayOfTables, stateArrayOfTables, p)
}

// empty table is special
func recEmptyTable(p parser) receptor {
	return nil
}

func recValues(p parser) receptor {
	return rec(tokensValues, stateValues, p)
}

func recArray(p parser) receptor { // LIFO Stack
	return rec(tokensArray, stateArray, p)
}
func recStringArray(p parser) receptor {
	return rec(tokensStringArray, stateStringArray, p)
}
func recBooleanArray(p parser) receptor {
	return rec(tokensBooleanArray, stateBooleanArray, p)
}
func recIntegerArray(p parser) receptor {
	return rec(tokensIntegerArray, stateIntegerArray, p)
}
func recFloatArray(p parser) receptor {
	return rec(tokensFloatArray, stateFloatArray, p)
}
func recDatetimeArray(p parser) receptor {
	return rec(tokensDatetimeArray, stateDatetimeArray, p)
}

// tokens
func itsWhitespace(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if !maybe && isWhitespace(r) {
			return SMaybe, 1
		}
	case 1:
		if isWhitespace(r) {
			return SMaybe, 1
		}
		return SYes, tokenWhitespace
	}
	return SNot, tokenWhitespace
}
func itsComment(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if !maybe && r == '#' {
			return SMaybe, 1
		}
	case 1:
		if isNewLine(r) || isEOF(r) {
			return SYes, tokenComment
		}
		return SMaybe, 1
	}
	return SNot, tokenComment
}
func itsString(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if !maybe && r == '"' {
			return SMaybe, 2
		}
		return SNot, tokenString
	case 1:
		if !isNewLine(r) {
			return SMaybe, 2
		}
	case 2:
		if r == '\\' {
			return SMaybe, 1
		}
		if r == '"' {
			return SMaybe, 3
		}
		if !isNewLine(r) {
			return SMaybe, 2
		}
	case 3:
		if isSuffixOfValue(r) {
			return SYes, tokenString
		}
	}
	return SInvalid, tokenString
}
func itsInteger(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if r == '-' {
			return SMaybe, 1
		}
		if is09(r) {
			return SMaybe, 2
		}
	case 1:
		if is09(r) {
			return SMaybe, 2
		}
	case 2:
		if is09(r) {
			return SMaybe, 2
		}
		if isSuffixOfValue(r) {
			return SYes, tokenInteger
		}
	}
	return SNot, tokenInteger
}
func itsFloat(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if r == '-' {
			return SMaybe, 1
		}
		if is09(r) {
			return SMaybe, 2
		}
	case 1:
		if is09(r) {
			return SMaybe, 2
		}
	case 2:
		if is09(r) {
			return SMaybe, 2
		}
		if r == '.' {
			return SMaybe, 3
		}
	case 3:
		if is09(r) {
			return SMaybe, 4
		}
		return SInvalid, tokenFloat
	case 4:
		if is09(r) {
			return SMaybe, 4
		}
		if isSuffixOfValue(r) {
			return SYes, tokenFloat
		}
	}
	return SNot, tokenFloat
}
func itsBoolean(r rune, flag Token, maybe bool) (Status, Token) {
	const layout = "truefalse"
	switch flag {
	case 0:
		if maybe {
			return SNot, tokenBoolean
		}
		if r == 't' {
			return SMaybe, 1
		}
		if r == 'f' {
			return SMaybe, 5
		}
	case 1, 2, 3, 5, 6, 7, 8:
		if rune(layout[flag]) == r {
			return SMaybe, flag + 1
		}
	case 4, 9:
		if isSuffixOfValue(r) {
			return SYes, tokenBoolean
		}
	}
	return SNot, tokenBoolean
}
func itsDatetime(r rune, flag Token, maybe bool) (Status, Token) {
	const layout = "0000-00-00T00:00:00Z"
	if flag < 20 {
		if layout[flag] == '0' && is09(r) || r == rune(layout[flag]) {
			return SMaybe, flag + 1
		}
		if flag <= 4 {
			return SNot, tokenDatetime
		}
	}
	if flag == 20 && isSuffixOfValue(r) {
		return SYes, tokenDatetime
	}
	return SInvalid, tokenDatetime
}

// [ASCII] http://en.wikipedia.org/wiki/ASCII#ASCII_printable_characters
func itsTableName(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if !maybe && r == '[' {
			return SMaybe, 1
		}
	case 1:
		if r == '[' {
			return SNot, tokenTableName
		}
		if isNewLine(r) || isWhitespace(r) || r == ']' {
			return SInvalid, tokenTableName
		}
		return SMaybe, 2
	case 2:
		if isNewLine(r) || isWhitespace(r) {
			return SInvalid, tokenTableName
		}
		if r != ']' {
			return SMaybe, 2
		}
		return SMaybe, 3
	case 3:
		return SYes, tokenTableName
	}
	return SNot, tokenTableName
}

func itsArrayOfTables(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0, 1:
		if r == '[' {
			return SMaybe, flag + 1
		}
	case 2:
		if isNewLine(r) || isWhitespace(r) || r == ']' {
			return SInvalid, tokenArrayOfTables
		}
		return SMaybe, 3
	case 3:
		if isNewLine(r) || isWhitespace(r) {
			return SInvalid, tokenArrayOfTables
		}
		if r != ']' {
			return SMaybe, 3
		}
		return SMaybe, 4
	case 4:
		if r == ']' {
			return SMaybe, 5
		}
		return SInvalid, tokenArrayOfTables
	case 5:
		return SYes, tokenArrayOfTables
	}
	return SNot, tokenArrayOfTables
}
func itsNewLine(r rune, flag Token, maybe bool) (Status, Token) {
	if isNewLine(r) {
		return SYes, tokenNewLine
	}
	return SNot, tokenNewLine
}
func itsSpace(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if isWhitespace(r) || isNewLine(r) {
			return SMaybe, 1
		}
	case 1:
		if isWhitespace(r) || isNewLine(r) {
			return SMaybe, 1
		}
		return SYes, tokenNewLine
	}
	return SNot, tokenNewLine
}

func itsKey(r rune, flag Token, maybe bool) (Status, Token) {
	if maybe && flag == 0 {
		return SNot, tokenKey
	}
	switch flag {
	case 0:
		return SMaybe, 1
	case 1:
		if isNewLine(r) || isEOF(r) {
			return SInvalid, tokenKey
		}
		if r == '=' || isWhitespace(r) {
			return SYes, tokenKey
		}
		return SMaybe, 1
	}
	return SNot, tokenKey
}
func itsEqual(r rune, flag Token, maybe bool) (Status, Token) {
	if maybe && flag == 0 {
		return SNot, tokenEqual
	}
	if flag == 1 {
		return SYes, tokenEqual
	}
	if r == '=' {
		return SMaybe, 1
	}
	return SNot, tokenEqual
}

func itsComma(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if !maybe && r == ',' {
			return SMaybe, 1
		}
	case 1:
		return SYes, tokenComma
	}
	return SNot, tokenComma
}

func itsArrayLeftBrack(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if !maybe && r == '[' {
			return SMaybe, 1
		}
	case 1:
		return SYes, tokenArrayLeftBrack
	}
	return SNot, tokenArrayLeftBrack
}
func itsArrayRightBrack(r rune, flag Token, maybe bool) (Status, Token) {
	switch flag {
	case 0:
		if !maybe && r == ']' {
			return SMaybe, 1
		}
	case 1:
		return SYes, tokenArrayRightBrack
	}
	return SNot, tokenArrayRightBrack
}

func isEOF(r rune) bool {
	return r == EOF
}
func is09(r rune) bool {

	return r >= '0' && r <= '9'
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' // || r == '\v' || r == '\f' // 0x85, 0xA0 ?
}

// [NewLine](http://en.wikipedia.org/wiki/Newline)
//	LF    = 0x0A   // Line feed, \n
//	CR    = 0x0D   // Carriage return, \r
//	LFCR  = 0x0A0D // \n\r
//	CRLF  = 0x0D0A // \r\n
//	RS    = 0x1E   // QNX pre-POSIX implementation.
func isNewLine(r rune) bool {
	return r == '\n' || r == '\r' || r == 0x1E
}
func isSuffixOfValue(r rune) bool {
	return isWhitespace(r) || isNewLine(r) || isEOF(r) || r == '#' || r == ',' || r == ']'
}
