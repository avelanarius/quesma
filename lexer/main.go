package main

import (
	"fmt"
	"lexer/core"
	"lexer/dialect_sqlparse"
)

func main() {
	for _, rule := range dialect_sqlparse.SQL_REGEX {
		if rule, isRegex := rule.(*core.RegexRule); isRegex {
			fmt.Println(rule.DeleteMe)
		}
	}
}
