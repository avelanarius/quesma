// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	"fmt"
	lexer_core "lexer/core"
	"lexer/dialect_sqlparse"
	"parser/core"
	"parser/transforms"
)

func main() {
	tokens := lexer_core.Lex(`
SELECT * FROM (SELECT * FROM (SELECT * FROM tabela) |> WHERE b = 9) 
|> WHERE a = 3 |> WHERE c = 9 |> JOIN a = 9
`, dialect_sqlparse.SqlparseRules)
	node := core.TokensToNode(tokens)
	node, err := transforms.GroupParenthesis(node)
	if err != nil {
		panic(err)
	}
	node = transforms.TransformPipeSyntax(node)
	fmt.Println(PrettyPrint(node))
}
