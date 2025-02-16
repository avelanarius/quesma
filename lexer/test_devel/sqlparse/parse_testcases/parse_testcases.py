#!/bin/env python3
import sqlparse

def main():
    with open("/mount/test_devel/sqlparse/extract_testcases/extracted-sqlparse-testcases.txt", "rb") as f:
        content = f.read()
    queries = content.split(b"\n<end_of_query/>\n")[:-1]

    output = open("/mount/dialect_sqlparse/test_files/parsed-sqlparse-testcases.txt", "w")

    for query in queries:
        query_str = query.decode('utf-8')
        output.write(query_str)
        output.write("\n<end_of_query/>\n")

        for (token_type, token_content) in sqlparse.lexer.tokenize(query_str):
            output.write(repr(token_type))
            output.write("\n")
            output.write(token_content)
            output.write("\n<end_of_token/>\n")

        output.write("\n<end_of_tokens/>\n")

    print("Processed", len(queries), "queries")

    output.close()

if __name__ == '__main__':
    main()