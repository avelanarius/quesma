#!/bin/bash -e

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH

sqlparse/extract_testcases/extract_testcases.sh
sqlfluff/extract_testcases/extract_testcases.sh

sqlparse/parse_testcases/parse_testcases.sh
sqlfluff/parse_testcases/parse_testcases.sh