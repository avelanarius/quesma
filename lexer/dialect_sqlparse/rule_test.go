// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package dialect_sqlparse

import (
	"lexer/core"
	"testing"

	"github.com/stretchr/testify/assert"
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
