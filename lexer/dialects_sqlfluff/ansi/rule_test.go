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

// FIXME: code duplication

func TestSqlfluffAnsiTestcases(t *testing.T) {
	testfiles := []string{
		"../test_files/parsed-sqlfluff-ansi-testcases.txt",
		"../test_files/parsed-sqlparse-testcases.txt",
	}

	for _, testfile := range testfiles {
		t.Run(testfile, func(t *testing.T) {
			testcases := loadParsedTestcases(testfile)
			for _, testcase := range testcases {
				t.Run(testcase.query, func(t *testing.T) {
					tokens := core.Lex(testcase.query, SqlfluffAnsiRules)
					assert.Equal(t, len(testcase.expectedTokens), len(tokens))

					commonLength := min(len(testcase.expectedTokens), len(tokens))

					for i := 0; i < commonLength; i++ {
						assert.Equalf(t, testcase.expectedTokens[i].tokenType, tokens[i].Type.Name, "Token type at position %d", i)
						assert.Equalf(t, testcase.expectedTokens[i].tokenValue, tokens[i].RawValue, "Token value at position %d", i)
					}

					if t.Failed() {
						for i := 0; i < commonLength; i++ {
							if testcase.expectedTokens[i].tokenType != tokens[i].Type.Name || testcase.expectedTokens[i].tokenValue != tokens[i].RawValue {
								t.Logf("Mismatch token at position %d: %s(%s). Got: %s(%s)", i, testcase.expectedTokens[i].tokenType, testcase.expectedTokens[i].tokenValue, tokens[i].Type.Name, tokens[i].RawValue)
							} else {
								t.Logf("Expected token at position %d: %s(%s). Got: %s(%s)", i, testcase.expectedTokens[i].tokenType, testcase.expectedTokens[i].tokenValue, tokens[i].Type.Name, tokens[i].RawValue)
							}
						}
					}
				})
			}
		})
	}
}

func FuzzLex(f *testing.F) {
	testcases := loadParsedTestcases("../test_files/parsed-sqlfluff-ansi-testcases.txt")
	for _, testcase := range testcases {
		f.Add(testcase.query)
	}

	f.Fuzz(func(t *testing.T, input string) {
		tokens := core.Lex(input, SqlfluffAnsiRules)

		// Basic checks:

		for _, token := range tokens {
			// Position should never be negative
			if token.Position < 0 {
				t.Errorf("Token position is negative: %d", token.Position)
			}

			// Token raw value should not be empty
			if len(token.RawValue) == 0 {
				t.Error("Token has empty raw value")
			}

			// Position should be within input string bounds
			if token.Position > len(input) {
				t.Errorf("Token position %d exceeds input length %d", token.Position, len(input))
			}
		}
	})
}

func BenchmarkLex(b *testing.B) {
	testCases := map[string]string{
		"empty":         "",
		"small_query":   "SELECT * FROM tabela",
		"medium_query":  "select * from foo where bar = 1 order by id desc",
		"subquery":      "select * from (select a, b + c as d from table) sub",
		"complex_query": "select 'abc' as foo, json_build_object('a', a,'b', b, 'c', c, 'd', d, 'e', e) as col2col3 from my_table",
		"long_query":    "SELECT t1.column1, t2.column2, t3.column3, SUM(t4.amount) FROM table1 t1 INNER JOIN table2 t2 ON t1.id = t2.id LEFT JOIN table3 t3 ON t2.id = t3.id INNER JOIN table4 t4 ON t3.id = t4.id WHERE t1.date >= '2023-01-01' AND t2.status = 'active' GROUP BY t1.column1, t2.column2, t3.column3 HAVING SUM(t4.amount) > 1000 ORDER BY t1.column1 DESC, t2.column2 ASC LIMIT 100",
		"invalid_query": "SELECT * FORM tabel WERE x = y",
		"garbage":       "!@#$%^&* )( asdf123 ;;; ~~~",
	}

	for name, tc := range testCases {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				core.Lex(tc, SqlfluffAnsiRules)
			}
		})
	}
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
