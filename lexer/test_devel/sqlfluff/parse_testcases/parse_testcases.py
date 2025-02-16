#!/bin/env python3

from sqlfluff.core.parser import Lexer

def run(dialect, input, output):
    with open(input, "rb") as f:
        content = f.read()
    queries = content.split(b"\n<end_of_query/>\n")[:-1]

    output = open(output, "w")
    lexer = Lexer(dialect=dialect)

    for query in queries:
        query_str = query.decode('utf-8')
        output.write(query_str)
        output.write("\n<end_of_query/>\n")

        segments, errors = lexer.lex(query_str)
        segments = list(segments)

        assert not errors

        for segment in segments:
            output.write(type(segment).__name__)
            output.write("\n")
            output.write(segment._raw)
            output.write("\n<end_of_token/>\n")

        output.write("\n<end_of_tokens/>\n")

    print(input, ": processed", len(queries), "queries")

    output.close()

if __name__ == '__main__':
    run("ansi", "/mount/test_devel/sqlfluff/extract_testcases/extracted-sqlfluff-ansi-testcases.txt", "/mount/dialects_sqlfluff/test_files/parsed-sqlfluff-ansi-testcases.txt")
