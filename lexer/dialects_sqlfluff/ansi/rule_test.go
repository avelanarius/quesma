// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package ansi

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"lexer/core"
	"os"
	"testing"
)

func TestSqlfluffAnsiTestcases(t *testing.T) {
	testcases := loadParsedTestcases("../test_files/parsed-sqlfluff-ansi-testcases.txt")
	for _, testcase := range testcases {
		t.Run(testcase.query, func(t *testing.T) {
			tokens := core.Lex(testcase.query, SqlfluffAnsiRules)
			assert.Equal(t, len(testcase.expectedTokens), len(tokens))

			commonLength := min(len(testcase.expectedTokens), len(tokens))

			for i := 0; i < commonLength; i++ {
				assert.Equalf(t, testcase.expectedTokens[i].tokenType, tokens[i].Type.Name, "Token type at position %d", i)
				assert.Equalf(t, testcase.expectedTokens[i].tokenValue, tokens[i].RawValue, "Token value at position %d", i)
			}
		})
	}
}

// FIXME: code duplication
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
		tokens = tokens[:len(tokens)-2]

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
