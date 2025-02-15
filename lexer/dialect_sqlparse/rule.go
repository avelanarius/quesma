// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package dialect_sqlparse

import (
	"lexer/core"
	"strings"
)

// Based on https://github.com/andialbrecht/sqlparse/blob/38c065b86ac43f76ffd319747e57096ed78bfa63/sqlparse/keywords.py

var SQL_REGEX = []core.Rule{
	core.NewRegexRule(`(--|# )\+.*?(\r\n|\r|\n|$)`, &SingleHintTokenType),
	core.NewRegexRule(`/\*\+[\s\S]*?\*/`, &MultilineHintTokenType),

	core.NewRegexRule(`(--|# ).*?(\r\n|\r|\n|$)`, &SingleCommentTokenType),
	core.NewRegexRule(`/\*[\s\S]*?\*/`, &MultilineCommentTokenType),

	core.NewRegexRule(`(\r\n|\r|\n)`, &NewlineTokenType),
	core.NewRegexRule(`\s+?`, &WhitespaceTokenType),

	core.NewRegexRule(`:=`, &AssignmentTokenType),
	core.NewRegexRule(`::`, &PunctuationTokenType),

	core.NewRegexRule(`\*`, &WildcardTokenType),

	core.NewRegexRule("`(``|[^`])*`", &NameTokenType),
	core.NewRegexRule(`´(´´|[^´])*´`, &NameTokenType),
	// TODO: doesn't compile in golang
	// core.NewRegexRule(`((?<![\w\"\$])\$(?:[_A-ZÀ-Ü]\w*)?\$)[\s\S]*?\1`, &LiteralTokenType),

	core.NewRegexRule(`\?`, &PlaceholderNameTokenType),
	core.NewRegexRule(`%(\(\w+\))?s`, &PlaceholderNameTokenType),

	// TODO: doesn't compile in golang
	//core.NewRegexRule(`(?<!\w)[$:?]\w+`, &PlaceholderNameTokenType),

	core.NewRegexRule(`\\\w+`, &CommandTokenType),

	// FIXME(andi): VALUES shouldn't be listed here
	// see https://github.com/andialbrecht/sqlparse/pull/64
	// AS and IN are special, it may be followed by a parenthesis, but
	// are never functions, see issue183 and issue507
	core.NewRegexRule(`(CASE|IN|VALUES|USING|FROM|AS)\b`, &KeywordTokenType),

	core.NewRegexRule(`(@|##|#)[A-ZÀ-Ü]\w+`, &NameTokenType),

	// see issue #39
	// Spaces around period `schema . name` are valid identifier
	// TODO: Spaces before period not implemented
	// TODO: doesn't compile in golang
	// core.NewRegexRule(`[A-ZÀ-Ü]\w*(?=\s*\.)`, &NameTokenType), // 'Name'.
	// FIXME(atronah): never match,
	// because `re.match` doesn't work with look-behind regexp feature
	// TODO: doesn't compile in golang
	//core.NewRegexRule(`(?<=\.)[A-ZÀ-Ü]\w*`, &NameTokenType), // .'Name'
	// TODO: doesn't compile in golang
	//core.NewRegexRule(`[A-ZÀ-Ü]\w*(?=\()`, &NameTokenType), // side effect: change kw to func
	core.NewRegexRule(`-?0x[\dA-F]+`, &HexadecimalNumberTokenType),
	core.NewRegexRule(`-?\d+(\.\d+)?E-?\d+`, &FloatNumberTokenType),
	// TODO: doesn't compile in golang
	//core.NewRegexRule(`(?![_A-ZÀ-Ü])-?(\d+(\.\d*)|\.\d+)(?![_A-ZÀ-Ü])`, &FloatNumberTokenType),
	// TODO: doesn't compile in golang
	//core.NewRegexRule(`(?![_A-ZÀ-Ü])-?\d+(?![_A-ZÀ-Ü])`, &IntegerNumberTokenType),
	core.NewRegexRule(`'(''|\\'|[^'])*'`, &SingleStringTokenType),
	// not a real string literal in ANSI SQL:
	core.NewRegexRule(`"(""|\\"|[^"])*"`, &SymbolStringTokenType),
	core.NewRegexRule(`(""|".*?[^\\]")`, &SymbolStringTokenType),
	// sqlite names can be escaped with [square brackets]. left bracket
	// cannot be preceded by word character or a right bracket --
	// otherwise it's probably an array index
	// TODO: doesn't compile in golang
	//core.NewRegexRule(`(?<![\w\])])(\[[^\]\[]+\])`, &NameTokenType),
	core.NewRegexRule(`((LEFT\s+|RIGHT\s+|FULL\s+)?(INNER\s+|OUTER\s+|STRAIGHT\s+)?`+
		`|(CROSS\s+|NATURAL\s+)?)?JOIN\b`, &KeywordTokenType),
	core.NewRegexRule(`END(\s+IF|\s+LOOP|\s+WHILE)?\b`, &KeywordTokenType),
	core.NewRegexRule(`NOT\s+NULL\b`, &KeywordTokenType),
	core.NewRegexRule(`(ASC|DESC)(\s+NULLS\s+(FIRST|LAST))?\b`, &OrderKeywordTokenType),
	core.NewRegexRule(`(ASC|DESC)\b`, &OrderKeywordTokenType),
	core.NewRegexRule(`NULLS\s+(FIRST|LAST)\b`, &OrderKeywordTokenType),
	core.NewRegexRule(`UNION\s+ALL\b`, &KeywordTokenType),
	core.NewRegexRule(`CREATE(\s+OR\s+REPLACE)?\b`, &DDLKeywordTokenType),
	core.NewRegexRule(`DOUBLE\s+PRECISION\b`, &BuiltinNameTokenType),
	core.NewRegexRule(`GROUP\s+BY\b`, &KeywordTokenType),
	core.NewRegexRule(`ORDER\s+BY\b`, &KeywordTokenType),
	core.NewRegexRule(`PRIMARY\s+KEY\b`, &KeywordTokenType),
	core.NewRegexRule(`HANDLER\s+FOR\b`, &KeywordTokenType),
	core.NewRegexRule(`GO(\s\d+)\b`, &KeywordTokenType),
	core.NewRegexRule(`(LATERAL\s+VIEW\s+)`+
		`(EXPLODE|INLINE|PARSE_URL_TUPLE|POSEXPLODE|STACK)\b`, &KeywordTokenType),
	core.NewRegexRule(`(AT|WITH')\s+TIME\s+ZONE\s+'[^']+'`, &TZCastKeywordTokenType),
	core.NewRegexRule(`(NOT\s+)?(LIKE|ILIKE|RLIKE)\b`, &ComparisonOperatorTokenType),
	core.NewRegexRule(`(NOT\s+)?(REGEXP)\b`, &ComparisonOperatorTokenType),
	// Check for keywords, also returns tokens.Name if regex matches
	// but the match isn't a keyword.
	NewProcessAsKeywordRule(`\w[$#\w]*`, &NameTokenType, ALL_KEYWORDS),
	core.NewRegexRule(`[;:()\[\],\.]`, &PunctuationTokenType),
	// JSON operators
	core.NewRegexRule(`(\->>?|#>>?|@>|<@|\?\|?|\?&|\-|#\-)`, &OperatorTokenType),
	core.NewRegexRule(`[<>=~!]+`, &ComparisonOperatorTokenType),
	core.NewRegexRule(`[+/@#%^&|^-]+`, &OperatorTokenType),
}

var SqlparseRules = core.NewRuleList(SQL_REGEX...)

// PROCESS_AS_KEYWORD in sqlparse
type ProcessAsKeywordRule struct {
	regexRule *core.RegexRule
	keywords  map[string]*core.TokenType
}

func NewProcessAsKeywordRule(regex string, defaultTokenType *core.TokenType, keywords map[string]*core.TokenType) *ProcessAsKeywordRule {
	// TODO: add safeguard to ensure that keywords are uppercase
	return &ProcessAsKeywordRule{regexRule: core.NewRegexRule(regex, defaultTokenType), keywords: keywords}
}

func (p *ProcessAsKeywordRule) Match(input string) (core.Token, bool) {
	token, matched := p.regexRule.Match(input)
	if matched {
		if keywordTokenType, found := p.keywords[strings.ToUpper(token.RawValue)]; found {
			token.Type = keywordTokenType
		}
	}
	return token, matched
}

func (p *ProcessAsKeywordRule) Name() string {
	return "ProcessAsKeywordRule"
}
