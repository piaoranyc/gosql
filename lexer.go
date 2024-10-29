package gosql

type location struct {
	line uint
	col  uint
}

type keyword string

const (
	selectKeyword keyword = "select"
	fromKeyword   keyword = "from"
	asKeyword     keyword = "as"
	tableKeyword  keyword = "table"
	createKeyword keyword = "create"
	insertKeyword keyword = "insert"
	intoKeyword   keyword = "into"
	valuesKeyword keyword = "values"
	intKeyword    keyword = "int"
	textKeyword   keyword = "text"
)

type symbol string

const (
	semicolonSymbol  symbol = ";"
	asteriskSymbol   symbol = "*"
	commaSymbol      symbol = ","
	leftparenSymbol  symbol = "("
	rightparenSymbol symbol = ")"
)

type tokenKind uint

const (
	keywordKind tokenKind = iota
	symbolKind
	identifierKind
	stringKind
	numericKind
)

type token struct {
	value string
	kind  tokenKind
	loc   location
}

type cursor struct {
	pointer uint
	loc     location
}

func (t *token) equals(other *token) bool {
	return t.value == other.value && t.kind == other.kind
}

type lexer func(string, cursor) (*token, cursor, error)

//	func lex(source string) ([]*token, error) {
//		tokens := []*token{}
//		cur := cursor{}
//
// lex:
//
//		for cur.pointer < uint(len(source)) {
//			lexers := []lexer{lexKeword}
//		}
//	}
func lexNumeric(source string, ic cursor) (*token, cursor, bool) {
	cur := ic
	periodFound := false
	expMarkerFound := false

	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		cur.loc.col++
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
				cur.loc.col++
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
	return &token{
		value: source[ic.pointer:cur.pointer],
		loc:   ic.loc,
		kind:  numericKind,
	}, cur, true
}

func lexCharacterDelimited(source string, ic cursor, delimiter byte) (*token, cursor, bool) {
	cur := ic
	if len(source[cur.pointer:]) == 0 {
		return nil, ic, false
	}
	if source[cur.pointer] != delimiter {
		return nil, ic, false
	}
	cur.loc.col++
	cur.pointer++
	var value []byte
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		if c == delimiter {
			if cur.pointer+1 >= uint(len(source)) || source[cur.pointer+1] != delimiter {
				return &token{
					value: string(value),
					loc:   ic.loc,
					kind:  stringKind,
				}, cur, true
			} else {
				value = append(value, delimiter)
				cur.pointer++
				cur.loc.col++
			}

		}
		value = append(value, c)
		cur.loc.col++
	}
	return nil, ic, false
}
func lexString(source string, ic cursor) (*token, cursor, bool) {
	return lexCharacterDelimited(source, ic, '\'')
}

//func lexSymbol(source string, ic cursor) (*token, cursor, bool) {
//	cur := ic
//	c := source[cur.pointer]
//	cur.loc.col++
//	cur.pointer++
//	switch c {
//	case '\n':
//		cur.loc.line++
//		cur.loc.col = 0
//		fallthrough
//	case '\t':
//		fallthrough
//	case ' ':
//		return nil, cur, true
//
//	}
//	symbols := []symbol{
//		semicolonSymbol,
//		asteriskSymbol,
//		commaSymbol,
//		leftparenSymbol,
//		rightparenSymbol,
//	}
//	var options []string
//	for _, s := range symbols {
//		options = append(options, string(s))
//	}
//	longestMatch()
//
//}
