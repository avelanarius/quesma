// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package core

type Token struct {
	RawValue string

	Type *TokenType

	Position int
}

var EmptyToken = Token{}

func MakeToken(rawValue string, tokenType *TokenType) Token {
	return Token{
		RawValue: rawValue,
		Type:     tokenType,
	}
}

type TokenType struct {
	Name        string
	Description string
}

var ErrorTokenType = &TokenType{
	Name:        "Error",
	Description: "Error token", // See RawValue for the actual error message
}
