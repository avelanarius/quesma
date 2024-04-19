package lucene

import (
	"math"
	"mitmproxy/quesma/logger"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// Mainly based on this doc: https://lucene.apache.org/core/2_9_4/queryparsersyntax.html
// Alternatively: https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-query-string-query.html

// We don't support:
// - Fuzzy search (e.g. roam~0.8, ~0.8 is simply removed)
// - Wildcards ? and * - they are treated as regular characters
//   (I think I'll add at least some basic support for them quite soon, it's needed for sample dashboards)
// - escaped " inside quoted fieldnames, so e.g.
//     * "a\"b" - not supported
//     * abc"def - supported
// - +, -, &&, ||, ! operators. But AND, OR, NOT are fully supported and they seem equivalent.

// Date ranges are only in format YYYY-MM-DD, as in docs there are no other examples. That can be changed if needed.

// Used in parsing one Lucene query. During parsing lastExpression keeps parsed part of the query,
// and tokens keep the rest (unparsed yet) part of the query.
// After parsing, the result expression is in lastExpression.
// If you have multiple queries to parse, create a new luceneParser for each query.
type luceneParser struct {
	tokens            []token
	defaultFieldNames []string
	lastExpression    expression
}

func newLuceneParser(defaultFieldNames []string) luceneParser {
	return luceneParser{defaultFieldNames: defaultFieldNames, lastExpression: nil, tokens: make([]token, 0)}
}

const fuzzyOperator = '~'
const boostingOperator = '^'
const escapeCharacter = '\\'

const delimiterCharacter = ':'

const leftParenthesis = '('
const rightParenthesis = ')'
const inclusiveRangeOpeningCharacter = '['
const inclusiveRangeClosingCharacter = ']'
const exclusiveRangeOpeningCharacter = '{'
const exclusiveRangeClosingCharacter = '}'
const rangeSeparator = " TO "
const infiniteRange = "*"

var specialOperators = map[string]token{
	"AND ":                   andToken{},
	"OR ":                    orToken{},
	"NOT ":                   notToken{},
	string(leftParenthesis):  leftParenthesisToken{},
	string(rightParenthesis): rightParenthesisToken{},
}

func TranslateToSQL(query string, fields []string) string {
	parser := newLuceneParser(fields)
	return parser.translateToSQL(query)
}

func (p *luceneParser) translateToSQL(query string) string {
	query = p.removeFuzzySearchOperator(query)
	query = p.removeBoostingOperator(query)
	p.tokenizeQuery(query)
	for len(p.tokens) > 0 {
		p.lastExpression = p.buildExpression(true)
	}
	if p.lastExpression == nil {
		return "true"
	}
	return p.lastExpression.toSQL()
}

func (p *luceneParser) tokenizeQuery(query string) {
	query = strings.TrimSpace(query)
	for len(query) > 0 {
		nextTokens, remainingQuery := p.nextToken(query)
		p.tokens = append(p.tokens, nextTokens...)
		query = strings.TrimSpace(remainingQuery)
	}
}

func (p *luceneParser) nextToken(query string) (tokens []token, remainingQuery string) {
	// parsing special operators
	for operator, operatorToken := range specialOperators {
		if strings.HasPrefix(query, operator) {
			return []token{operatorToken}, query[len(operator):]
		}
	}

	// parsing term(:value)
	term, remainingQuery := p.parseTerm(query, false)

	// case 1. there's no ":value"
	remainingQuery = strings.TrimSpace(remainingQuery)
	if len(remainingQuery) == 0 || remainingQuery[0] != delimiterCharacter {
		return []token{term}, remainingQuery
	}

	// case 2. query[len(term)] == ':" => there's ":value"
	if termCasted, termIsFieldName := term.(termToken); termIsFieldName {
		// this branch should always be used, but being cautious and wrapping in if
		// to not panic in case of invalid query
		return []token{newFieldNameToken(termCasted.term), newSeparatorToken()}, remainingQuery[1:]
	}
	return []token{term, newSeparatorToken()}, remainingQuery[1:]
}

// query - non-empty string
// closingBoundTerm is true <=> we're parsing the second bound of the range.
// Then we finish when we encounter ']' or '}'. Otherwise we don't.
func (p *luceneParser) parseTerm(query string, closingBoundTerm bool) (token token, remainingQuery string) {
	switch query[0] {
	case '"':
		for i, r := range query[1:] {
			if r == '"' {
				return newTermToken(query[:i+2]), query[i+2:]
			}
		}
		logger.Error().Msgf("unterminated quoted term, query: %s", query)
		return newInvalidToken(), ""
	case '>', '<', inclusiveRangeOpeningCharacter, exclusiveRangeOpeningCharacter:
		return p.parseRange(query)
	default:
		for i, r := range query {
			if r == ' ' || r == delimiterCharacter || r == rightParenthesis || (closingBoundTerm && (r == exclusiveRangeClosingCharacter || r == inclusiveRangeClosingCharacter)) {
				return newTermToken(query[:i]), query[i:]
			}
		}
		return newTermToken(query), ""
	}
}

func (p *luceneParser) parseRange(query string) (token token, remainingQuery string) {
	var number float64
	switch query[0] {
	case '>', '<':
		if len(query) == 1 {
			logger.Error().Msgf("parseRange: invalid range, missing value, query: %s", query)
			return newInvalidToken(), ""
		}
		acceptableCharactersAfterNumber := []rune{' ', rightParenthesis}
		if query[1] == '=' { // >=, <=
			number, remainingQuery = p.parseNumber(query[2:], true, acceptableCharactersAfterNumber)
			switch query[0] {
			case '>':
				return newRangeToken(newRangeValueGte(number)), remainingQuery
			case '<':
				return newRangeToken(newRangeValueLte(number)), remainingQuery
			}
		} else {
			number, remainingQuery = p.parseNumber(query[1:], true, acceptableCharactersAfterNumber)
			switch query[0] {
			case '>':
				return newRangeToken(newRangeValueGt(number)), remainingQuery
			case '<':
				return newRangeToken(newRangeValueLt(number)), remainingQuery
			}
		}
	case inclusiveRangeOpeningCharacter, exclusiveRangeOpeningCharacter:
		var lowerBound, upperBound any
		lowerBound, remainingQuery = p.parseOneBound(query[1:], false)
		if _, isInvalid := lowerBound.(invalidToken); isInvalid {
			return newInvalidToken(), ""
		}
		if len(remainingQuery) < len(rangeSeparator) || remainingQuery[:len(rangeSeparator)] != rangeSeparator {
			return newInvalidToken(), ""
		}
		upperBound, remainingQuery = p.parseOneBound(remainingQuery[len(rangeSeparator):], true)
		if _, isInvalid := upperBound.(invalidToken); isInvalid || len(remainingQuery) == 0 {
			return newInvalidToken(), ""
		}
		inclusiveOpening := query[0] == inclusiveRangeOpeningCharacter
		inclusiveClosing := remainingQuery[0] == inclusiveRangeClosingCharacter
		return newRangeToken(newRangeValue(lowerBound, inclusiveOpening, upperBound, inclusiveClosing)), remainingQuery[1:]
	}
	logger.Error().Msgf("parseRange: invalid range, query: %s", query)
	return newInvalidToken(), ""
}

// parseNumber returns (math.NaN, "") if parsing failed
// acceptableCharsAfterNumber - what character is acceptable as first character after the number,
// e.g. when acceptableCharsAfterNumber = {']', '}'}, then 200} or 200] parses to 200, but parsing 200( fails.
func (p *luceneParser) parseNumber(query string, reportErrors bool, acceptableCharsAfterNumber []rune) (number float64, remainingQuery string) {
	var i, dotCount = 0, 0
	for i = 0; i < len(query); i++ {
		r := rune(query[i])
		if r == '.' {
			dotCount++
			if dotCount > 1 {
				if reportErrors {
					logger.Error().Msgf("invalid number, multiple dots, query: %s", query)
				}
				return math.NaN(), ""
			}
			continue
		}
		if !unicode.IsDigit(r) {
			if !slices.Contains(acceptableCharsAfterNumber, r) {
				if reportErrors {
					logger.Error().Msgf("invalid number, query: %s", query)
				}
				return math.NaN(), ""
			}
			break
		}
	}
	var err error
	number, err = strconv.ParseFloat(query[:i], 64)
	if err != nil {
		if reportErrors {
			logger.Error().Msgf("invalid number, query: %s, error: %v", query, err)
		}
		return math.NaN(), ""
	}
	return number, query[i:]
}

// parseOneBound returns invalidToken{} if parsing failed
// closingBound == true <=> it's second bound, so ] or } are totally fine after the number
func (p *luceneParser) parseOneBound(query string, closingBound bool) (bound any, remainingQuery string) {
	// let's try parsing a number first, only if it fails, we'll parse it as a string
	var acceptableCharactersAfterNumber []rune
	if closingBound {
		acceptableCharactersAfterNumber = []rune{inclusiveRangeClosingCharacter, exclusiveRangeClosingCharacter}
	} else {
		acceptableCharactersAfterNumber = []rune{' '}
	}
	var number float64
	number, remainingQuery = p.parseNumber(query, false, acceptableCharactersAfterNumber)
	if !math.IsNaN(number) {
		return number, remainingQuery
	}

	var tok token
	tok, remainingQuery = p.parseTerm(query, closingBound)
	if term, isTerm := tok.(termToken); isTerm {
		if term.term == infiniteRange {
			bound = unbounded
		} else {
			bound = term.term
		}
		return bound, remainingQuery
	} else {
		logger.Error().Msgf("parseRange: invalid range, query: %s", query)
		return newInvalidToken(), ""
	}
}

func (p *luceneParser) removeFuzzySearchOperator(query string) string {
	return p.removeSpecialCharacter(query, fuzzyOperator)
}

func (p *luceneParser) removeBoostingOperator(query string) string {
	return p.removeSpecialCharacter(query, boostingOperator)
}

func (p *luceneParser) removeSpecialCharacter(query string, specialChar byte) string {
	var afterRemoval strings.Builder
	for i := 0; i < len(query); i++ {
		if query[i] == escapeCharacter && i+1 < len(query) && query[i+1] == specialChar {
			// it's escaped, we don't remove it
			i++
		} else if query[i] == specialChar {
			// remove the character together with the following number (may be float)
			for ; i+1 < len(query) && unicode.IsDigit(rune(query[i+1])); i++ {
			}
			if i+1 < len(query) && query[i+1] == '.' {
				i++
				for ; i+1 < len(query) && unicode.IsDigit(rune(query[i+1])); i++ {
				}
			}
		} else {
			afterRemoval.WriteByte(query[i])
		}
	}
	return afterRemoval.String()
}