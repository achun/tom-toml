/**
硬编码实现 PEG.
*/
package toml

type stager interface {
	// 返回场景中全部角色
	Roles() []role
	// 通过 当前的 Token 和 当前的 stager 返回下一个 stager
	Next(Token, stager) stager
}

const (
	stageEnd constStage = iota
	stageInvalid
)

// 常量场景, 用于标记特殊状态
type constStage int

func (s constStage) Roles() []role {
	return nil
}

func (s constStage) Next(Token, stager) stager {
	return nil
}

type itsToken func(rune, int, bool) (Status, Token)

// 角色(rules?)
type role struct {
	/**
	Say 判断角色是否能识别传入的字符.
	返回值:
		Status 识别状态
		Token  此值被复用, 如果 Status 是 SMaybe, 会作为 flag 的值.
	参数:
		char  是待判定字符
		flag  是上一次 SMaybe 状态的 uint(Token), 默认为 0.
		race  表示是否有其他 role 竞争
	*/
	Say itsToken
	// 与角色绑定的预期 stager
	// nil 值表示保持 当前的 stager 不变
	Stager stager
}

// 基本场景, 固定的, 不变的
type stage struct {
	roles []role
}

func (s stage) Roles() []role {
	// ???可能不不需要 copy, 待证实
	//return s.Roles()
	roles := make([]role, len(s.roles))
	copy(roles, s.roles)
	return roles
}

func (s stage) Next(token Token, stage stager) stager {
	return s
}

// 升降场景, 升上去总要降下来, 用于嵌套
// 当嵌套层数nested 为 0 回退到 back 场景.
// 成员 stager 提供 roles
// 由Token open,close 开启和结束, 因 Toml 中只有数组这一种情况, open,close 被省略了
type liftStage struct {
	stager // stageArray
	back   stager
	level  int
	// open   Token // ommit: tokenArrayLeftBrack
	// close  Token // ommit: tokenArrayRightBrack
}

// 改变 roles, 截获场景
func (s liftStage) Roles() []role {
	roles := s.stager.Roles()
	if s.level != 0 {
		for i, t := range roles {
			_, token := t.Say(0, -1, true)
			if token == tokenArrayLeftBrack || token == tokenArrayRightBrack {
				roles[i].Stager = s
			}
		}
	}
	return roles
}

func (s liftStage) Next(token Token, current stager) stager {

	if s.back == nil && current != nil {
		s.back = current
	}

	if token == tokenArrayLeftBrack {
		s.level++
	} else if token == tokenArrayRightBrack {
		s.level--
	}

	if s.level == 0 {
		return s.back
	}
	return s
}

// 回退场景, stager 提供 roles,
// back 开始应该为 nil, 由 Next 进行设置, 下一次 Next 返回 back
// Token close 描述 toles 中要替换的 stager, TOML 中只有 tokenArrayRightBrack
type backStage struct {
	stager
	back stager
	// close Token // ommit: tokenArrayRightBrack
}

// 改变 roles, 截获场景
func (s backStage) Roles() []role {
	roles := s.stager.Roles()
	for i, t := range roles {
		_, token := t.Say(0, -1, true)
		if token == tokenArrayRightBrack {
			roles[i].Stager = s
		}
	}
	return roles
}

func (s backStage) Next(token Token, current stager) stager {

	if s.back == nil {
		s.back = current
		return s
	}
	return s.back.Next(token, nil)
}

// token 环
func circleToken(fns ...itsToken) itsToken {
	max := len(fns)
	if max == 0 {
		return func(char rune, flag int, race bool) (Status, Token) { return SNot, tokenError }
	}
	i := 0
	return func(char rune, flag int, race bool) (Status, Token) {
		if i == max {
			i = 0
		}

		s, t := fns[i](char, flag, race)

		if s == SYes || s == SYesKeep {
			i++
		}

		return s, t
	}
}

// 开启新舞台, 返回第一个场景
func openStage() stager {
	stageEmpty := &stage{}
	stageEqual := &stage{}
	stageValues := &stage{}
	stageArray := &stage{}
	stageStringArray := &stage{}
	stageBooleanArray := &stage{}
	stageIntegerArray := &stage{}
	stageFloatArray := &stage{}
	stageDatetimeArray := &stage{}

	stageEmpty.roles = []role{
		{itsEOF, stageEnd},
		{itsWhitespace, nil},
		{itsNewLine, nil},
		{itsComment, nil},
		{itsTableName, nil},
		{itsArrayOfTables, nil},
		{itsKey, stageEqual},
	}

	stageEqual.roles = []role{
		{itsWhitespace, nil},
		{itsEqual, stageValues},
	}

	stageValues.roles = []role{
		{itsWhitespace, nil},
		{itsArrayLeftBrack, liftStage{stageArray, stageEmpty, 0}},
		{itsString, stageEmpty},
		{itsBoolean, stageEmpty},
		{itsInteger, stageEmpty},
		{itsFloat, stageEmpty},
		{itsDatetime, stageEmpty},
	}

	stageArray.roles = []role{
		{itsWhitespace, nil},
		{itsComment, nil},
		{itsNewLine, nil},
		{circleToken(itsArrayLeftBrack, itsComma), nil}, // stageArray 被 liftStage 代理
		{itsArrayRightBrack, nil},                       //
		{itsString,
			backStage{stageStringArray, nil}},
		{itsBoolean,
			backStage{stageBooleanArray, nil}},
		{itsInteger,
			backStage{stageIntegerArray, nil}},
		{itsFloat,
			backStage{stageFloatArray, nil}},
		{itsDatetime,
			backStage{stageDatetimeArray, nil}},
	}

	stageStringArray.roles = []role{
		{itsWhitespace, nil},
		{itsNewLine, nil},
		{itsComment, nil},
		{itsArrayRightBrack, stageInvalid},
		{circleToken(itsComma, itsString), nil},
	}

	stageBooleanArray.roles = []role{
		{itsWhitespace, nil},
		{itsNewLine, nil},
		{itsComment, nil},
		{itsArrayRightBrack, stageInvalid},
		{circleToken(itsComma, itsBoolean), nil},
	}

	stageIntegerArray.roles = []role{
		{itsWhitespace, nil},
		{itsNewLine, nil},
		{itsComment, nil},
		{itsArrayRightBrack, stageInvalid},
		{circleToken(itsComma, itsInteger), nil},
	}

	stageFloatArray.roles = []role{
		{itsWhitespace, nil},
		{itsNewLine, nil},
		{itsComment, nil},
		{itsArrayRightBrack, stageInvalid},
		{circleToken(itsComma, itsFloat), nil},
	}

	stageDatetimeArray.roles = []role{
		{itsWhitespace, nil},
		{itsNewLine, nil},
		{itsComment, nil},
		{itsArrayRightBrack, stageInvalid},
		{circleToken(itsComma, itsDatetime), nil},
	}

	return stageEmpty
}

func stagePlay(p parser, sta stager) {
Loop:
	for sta != nil {
		if sta == stageEnd {
			break
		}

		if sta == stageInvalid {
			p.Invalid(tokenError)
			break
		}

		roles := sta.Roles()
		skip := make([]bool, len(roles))
		flag := make([]int, len(roles))

		if len(roles) == 0 {
			p.Invalid(tokenError)
			break
		}

		var (
			st    Status
			token Token // flag 是 uint 和 Token 复用
			maybe int
			r     rune
		)

		for {

			r = p.Next()

			if r == RuneError {
				p.Invalid(tokenRuneError)
				return
			}

			//println("\nlength:", len(roles))
			for i, role := range roles {
				if skip[i] {
					continue
				}
				st, token = role.Say(r, flag[i], maybe != 0)
				//println(i, string(r), maybe, flag[i], st.String(), token.String(), uint(token))
				switch st {
				case SMaybe:
					if flag[i] == 0 {
						maybe++
					}
					flag[i] = int(token)

				case SYes, SYesKeep:
					if st == SYesKeep {
						p.Keep()
					}
					if p.Token(token) != nil {
						return
					}

					// 替代 stagePlay 递归
					if role.Stager != nil {
						sta = role.Stager.Next(token, sta)
					}
					continue Loop

				case SNot:
					if flag[i] != 0 {
						maybe--
					}
					skip[i] = true

				case SInvalid:
					p.Invalid(token)
					return
				}
			}

			if 0 == maybe {
				tokens := make([]Token, len(roles))
				for i, role := range roles {
					_, t := role.Say(0, -1, true)
					tokens[i] = t
				}

				p.NotMatch(tokens...)

				return
			}
		}
	}

}
