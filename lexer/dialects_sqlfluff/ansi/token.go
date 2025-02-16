// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package ansi

import "lexer/core"

var WhitespaceTokenType = core.TokenType{
	Name:        "WhitespaceSegment",
	Description: "Whitespace segment",
}

var CommentTokenType = core.TokenType{
	Name:        "CommentSegment",
	Description: "Comment segment",
}

var CodeTokenType = core.TokenType{
	Name:        "CodeSegment",
	Description: "Code segment",
}

var LiteralTokenType = core.TokenType{
	Name:        "LiteralSegment",
	Description: "Literal segment",
}

var ComparisonOperatorTokenType = core.TokenType{
	Name:        "ComparisonOperatorSegment",
	Description: "Comparison operator segment",
}

var NewlineTokenType = core.TokenType{
	Name:        "NewlineSegment",
	Description: "Newline segment",
}

var WordTokenType = core.TokenType{
	Name:        "WordSegment",
	Description: "Word segment",
}
