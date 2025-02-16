#!/bin/bash -e

cd /

git clone https://github.com/sqlfluff/sqlfluff.git
cd sqlfluff
git reset --hard 6666db9ed97f45161fb318f901392d9a214808d2

rm -f /mount/test_devel/sqlfluff/extract_testcases/extracted-sqlfluff-*-testcases.txt

for dialect_dir in test/fixtures/dialects/*/; do
    if [ -d "$dialect_dir" ]; then
        dialect=$(basename "$dialect_dir")
        output_file="/mount/test_devel/sqlfluff/extract_testcases/extracted-sqlfluff-${dialect}-testcases.txt"

        for sql_file in "${dialect_dir}"*.sql; do
            if [ -f "$sql_file" ]; then
                cat "$sql_file" >> "$output_file"
                printf "\n<end_of_query/>\n" >> "$output_file"
            fi
        done
        python3 /mount/test_devel/sqlfluff/extract_testcases/split_testcases.py "$output_file" "$output_file"
        echo "Extracted testcases for dialect ${dialect}, $(wc -l < "$output_file") lines"
    fi
done

cat /mount/test_devel/sqlfluff/extract_testcases/extracted-sqlfluff-*-testcases.txt > /mount/test_devel/sqlfluff/extract_testcases/extracted-sqlfluff-all-testcases.txt