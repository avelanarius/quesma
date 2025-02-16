#!/bin/env python3
import sqlparse
import sys

def run(input, output):
    with open(input, "r") as f:
        content = f.read()

    output = open(output, "w")

    for query in sqlparse.split(content):
        output.write(query + "\n<end_of_query/>\n")

    output.close()

if __name__ == '__main__':
    run(sys.argv[1], sys.argv[2])