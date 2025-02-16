// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package dialect_sqlparse

import (
	"bytes"
	"lexer/core"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleSelect(t *testing.T) {
	input := "SELECT * FROM tabela"
	tokens := core.Lex(input, SqlparseRules)

	assert.Equal(t, 7, len(tokens))

	assert.Equal(t, "Token.Keyword.DML", tokens[0].Type.Name)
	assert.Equal(t, "SELECT", tokens[0].RawValue)

	assert.Equal(t, "Token.Text.Whitespace", tokens[1].Type.Name)
	assert.Equal(t, " ", tokens[1].RawValue)

	assert.Equal(t, "Token.Wildcard", tokens[2].Type.Name)
	assert.Equal(t, "*", tokens[2].RawValue)

	assert.Equal(t, "Token.Text.Whitespace", tokens[3].Type.Name)
	assert.Equal(t, " ", tokens[3].RawValue)

	assert.Equal(t, "Token.Keyword", tokens[4].Type.Name)
	assert.Equal(t, "FROM", tokens[4].RawValue)

	assert.Equal(t, "Token.Text.Whitespace", tokens[5].Type.Name)
	assert.Equal(t, " ", tokens[5].RawValue)

	assert.Equal(t, "Token.Name", tokens[6].Type.Name)
	assert.Equal(t, "tabela", tokens[6].RawValue)
}

func TestSqlparseTestcases(t *testing.T) {
	testcases := loadParsedTestcases("test_files/parsed-sqlparse-testcases.txt")
	for _, testcase := range testcases {
		t.Run(testcase.query, func(t *testing.T) {
			tokens := core.Lex(testcase.query, SqlparseRules)
			require.Equal(t, len(testcase.expectedTokens), len(tokens))

			for i, expectedToken := range testcase.expectedTokens {
				assert.Equalf(t, expectedToken.tokenType, tokens[i].Type.Name, "Token type at position %d", i)
				assert.Equalf(t, expectedToken.tokenValue, tokens[i].RawValue, "Token value at position %d", i)
			}

			if t.Failed() {
				for i, expectedToken := range testcase.expectedTokens {
					t.Logf("Expected token at position %d: %s(%s). Got: %s(%s)", i, expectedToken.tokenType, expectedToken.tokenValue, tokens[i].Type.Name, tokens[i].RawValue)
				}
			}
		})
	}
}

func FuzzLex(f *testing.F) {
	testcases := loadParsedTestcases("test_files/parsed-sqlparse-testcases.txt")
	for _, testcase := range testcases {
		f.Add(testcase.query)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_ = core.Lex(input, SqlparseRules)
	})
}

type parsedTestcase struct {
	query          string
	expectedTokens []expectedToken
}

type expectedToken struct {
	tokenType  string
	tokenValue string
}

func loadParsedTestcases(filename string) []parsedTestcase {
	contents, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	testcases := bytes.Split(contents, []byte("\n<end_of_tokens/>\n"))
	testcases = testcases[:len(testcases)-1]

	var parsedTestcases []parsedTestcase
	for _, testcase := range testcases {
		endOfQuerySplit := bytes.Split(testcase, []byte("\n<end_of_query/>\n"))

		query := string(endOfQuerySplit[0])

		tokens := bytes.Split(endOfQuerySplit[1], []byte("\n<end_of_token/>\n"))
		tokens = tokens[:len(tokens)-1]

		var expectedTokens []expectedToken
		for _, tokenDescription := range tokens {
			tokenDescriptionSplit := bytes.SplitN(tokenDescription, []byte("\n"), 2)
			tokenType := string(tokenDescriptionSplit[0])
			tokenValue := string(tokenDescriptionSplit[1])
			expectedTokens = append(expectedTokens, expectedToken{tokenType, tokenValue})
		}

		parsedTestcases = append(parsedTestcases, parsedTestcase{query, expectedTokens})
	}
	return parsedTestcases
}
