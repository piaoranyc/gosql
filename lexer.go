package gosql

import (
	"fmt"
	"strings"
)

type Location struct {
	Line uint
	Col  uint
}

type keyword string

const (
	SelectKeyword keyword = "select"
	FromKeyword   keyword = "from"
	AsKeyword     keyword = "as"
	TableKeyword  keyword = "table"
	CreateKeyword keyword = "create"
	InsertKeyword keyword = "insert"
	IntoKeyword   keyword = "into"
	ValuesKeyword keyword = "values"
	IntKeyword    keyword = "int"
	TextKeyword   keyword = "text"
	WhereKeyword  keyword = "where"
)

type Symbol string

const (
	SemiColonSymbol  Symbol = ";"
	AsteriskSymbol   Symbol = "*"
	CommaSymbol      Symbol = ","
	LeftparenSymbol  Symbol = "("
	RightparenSymbol Symbol = ")"
)

type TokenKind uint

const (
	KeywordKind TokenKind = iota
	SymbolKind
	IdentifierKind
	StringKind
	NumericKind
)

type Token struct {
	Value string
	Kind  TokenKind
	Loc   Location
}

type cursor struct {
	pointer uint
	loc     Location
}

func (t *Token) equals(other *Token) bool {
	return t.Value == other.Value && t.Kind == other.Kind
}

type lexer func(string, cursor) (*Token, cursor, bool)

func lex(source string) ([]*Token, error) {
	tokens := []*Token{}
	cur := cursor{}

lex:

	for cur.pointer < uint(len(source)) {
		lexers := []lexer{lexKeyword, lexSymbol, lexNumeric, lexString, lexIdentifier}
		for _, l := range lexers {
			if token, newCursor, ok := l(source, cur); ok {
				cur = newCursor
				if token != nil {
					tokens = append(tokens, token)
				}
				continue lex
			}
		}
		hint := ""
		if len(tokens) > 0 {
			hint = " after " + tokens[len(tokens)-1].Value
		}
		return nil, fmt.Errorf("Unable to lex token%s, at %d:%d", hint, cur.loc.Line, cur.loc.Col)
	}
	return tokens, nil

}
func lexNumeric(source string, ic cursor) (*Token, cursor, bool) {
	cur := ic
	periodFound := false
	expMarkerFound := false

	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		cur.loc.Col++
		isDigit := c >= '0' && c <= '9'
		isPeriod := c == '.'
		isExpMarker := c == 'e'

		if cur.pointer == ic.pointer {
			if !isDigit && !isPeriod {
				return nil, ic, false
			}
			periodFound = isPeriod
			continue
		}
		if isPeriod {
			//period comes only one
			if periodFound {
				return nil, ic, false
			}
			continue
		}

		if isExpMarker {
			if expMarkerFound {
				return nil, ic, false
			}
			periodFound = true
			expMarkerFound = true

			if cur.pointer == uint(len(source))-1 {
				return nil, ic, false
			}
			cNext := source[cur.pointer+1]
			if cNext == '-' || cNext == '+' {
				cur.pointer++
				cur.loc.Col++
			}
			continue
		}
		if !isDigit {
			break
		}
	}
	if cur.pointer == ic.pointer {
		return nil, ic, false
	}
	return &Token{
		Value: source[ic.pointer:cur.pointer],
		Loc:   ic.loc,
		Kind:  NumericKind,
	}, cur, true
}

func lexCharacterDelimited(source string, ic cursor, delimiter byte) (*Token, cursor, bool) {
	cur := ic
	if len(source[cur.pointer:]) == 0 {
		return nil, ic, false
	}
	if source[cur.pointer] != delimiter {
		return nil, ic, false
	}
	cur.loc.Col++
	cur.pointer++
	var value []byte
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		if c == delimiter {
			if cur.pointer+1 >= uint(len(source)) || source[cur.pointer+1] != delimiter {
				return &Token{
					Value: string(value),
					Loc:   ic.loc,
					Kind:  StringKind,
				}, cur, true
			} else {
				value = append(value, delimiter)
				cur.pointer++
				cur.loc.Col++
			}

		}
		value = append(value, c)
		cur.loc.Col++
	}
	return nil, ic, false
}
func lexString(source string, ic cursor) (*Token, cursor, bool) {
	return lexCharacterDelimited(source, ic, '\'')
}

func longestMatch(source string, ic cursor, options []string) string {
	var value []byte
	var skipList []int
	var match string

	cur := ic
	for cur.pointer < uint(len(source)) {
		value = append(value, strings.ToLower(string(source[cur.pointer]))...)
		cur.pointer++
	match:
		for i, option := range options {
			for _, skip := range skipList {
				if i == skip {
					continue match
				}
			}
			if option == string(value) {
				skipList = append(skipList, i)
				if len(option) > len(match) {
					match = option
				}
				continue
			}
			sharePrefix := string(value) == option[:cur.pointer-ic.pointer]
			tooLong := len(value) > len(options)
			if tooLong || !sharePrefix {
				skipList = append(skipList, i)
			}
		}

		if len(skipList) == len(options) {
			break
		}
	}
	return match
}

func lexSymbol(source string, ic cursor) (*Token, cursor, bool) {
	cur := ic
	c := source[cur.pointer]
	cur.loc.Col++
	cur.pointer++
	switch c {
	case '\n':
		cur.loc.Line++
		cur.loc.Col = 0
		fallthrough
	case '\t':
		fallthrough
	case ' ':
		return nil, cur, true

	}
	symbols := []Symbol{
		SemiColonSymbol,
		AsteriskSymbol,
		CommaSymbol,
		LeftparenSymbol,
		RightparenSymbol,
	}
	var options []string
	for _, s := range symbols {
		options = append(options, string(s))
	}
	match := longestMatch(source, ic, options)
	if match == "" {
		return nil, ic, false
	}
	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.Col = ic.loc.Col + uint(len(match))

	return &Token{
		Value: match,
		Loc:   ic.loc,
		Kind:  SymbolKind,
	}, cur, true
}

func lexKeyword(source string, ic cursor) (*Token, cursor, bool) {
	cur := ic
	keywords := []keyword{
		SelectKeyword,
		InsertKeyword,
		ValuesKeyword,
		TableKeyword,
		CreateKeyword,
		WhereKeyword,
		FromKeyword,
		IntoKeyword,
		TextKeyword,
		IntKeyword,
	}
	var options []string
	for _, k := range keywords {
		options = append(options, string(k))
	}
	match := longestMatch(source, ic, options)
	if match == "" {
		return nil, ic, false
	}
	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.Col = ic.loc.Col + uint(len(match))

	return &Token{
		Value: match,
		Loc:   ic.loc,
		Kind:  KeywordKind,
	}, cur, true
}

func lexIdentifier(source string, ic cursor) (*Token, cursor, bool) {

	if token, newCursor, ok := lexCharacterDelimited(source, ic, '"'); ok {
		return token, newCursor, true
	}
	cur := ic
	c := source[cur.pointer]
	isAlpha := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	if !isAlpha {
		return nil, ic, false
	}
	cur.pointer++
	cur.loc.Col++
	value := []byte{c}
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c = source[cur.pointer]
		isAlpha := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
		isNumeric := c >= '0' && c <= '9'
		if isAlpha || isNumeric || c == '$' || c == '_' {
			value = append(value, c)
			cur.loc.Col++
			continue
		}
		break
	}
	if len(value) == 0 {
		return nil, ic, false
	}
	return &Token{
		Value: strings.ToLower(string(value)),
		Loc:   ic.loc,
		Kind:  IdentifierKind,
	}, cur, true
}
