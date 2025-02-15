package main

import (
	"fmt"
	"lexer/core"
	"lexer/dialect_sqlparse"
)

func main() {
	for _, rule := range dialect_sqlparse.SQL_REGEX {
		fmt.Println(rule.(*core.RegexRule).DeleteMe)
	}
}
