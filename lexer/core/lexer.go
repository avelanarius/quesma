// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package core

import "fmt"

func Lex(input string, rule Rule) []Token {
	var tokens []Token
	position := 0

	for len(input) > 0 {
		token, matched := rule.Match(input)

		if matched {
			token.Position = position
			tokens = append(tokens, token)

			input = input[len(token.RawValue):]
			position += len(token.RawValue)
		} else {
			// FIXME: don't put the entire input in the error message, only first ~20 or so characters
			errorToken := MakeToken(fmt.Sprintf("no rule matched input at position %d: '%s'", position, input), ErrorTokenType)
			errorToken.Position = position

			tokens = append(tokens, errorToken)
			break
		}
	}

	return tokens
}
