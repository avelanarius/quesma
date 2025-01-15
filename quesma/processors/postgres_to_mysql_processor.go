// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package processors

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
	_ "github.com/marcboeker/go-duckdb"
	"log"
	"net/http"
	"quesma/qpl_experiment"
	quesma_api "quesma_v2/core"
	"strings"
)

type PostgresToDuckDBProcessor struct {
	BaseProcessor
}

func NewPostgresToDuckDBProcessor() *PostgresToDuckDBProcessor {
	return &PostgresToDuckDBProcessor{
		BaseProcessor: NewBaseProcessor(),
	}
}

func (p *PostgresToDuckDBProcessor) InstanceName() string {
	return "PostgresToDuckDBProcessor"
}

func (p *PostgresToDuckDBProcessor) GetId() string {
	return "postgrestomysql_processor"
}

func (p *PostgresToDuckDBProcessor) respond(query string) ([][][]byte, []string, error) {
	var err error
	db, err := sql.Open("duckdb", "?access_mode=READ_WRITE")
	if err != nil {
		return nil, nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	parsedSql, err := qpl_experiment.ParseQPL(query)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing query: %w", err)
	}
	fmt.Println(parsedSql)

	var sqltvf *qpl_experiment.SQLTVF
	var iserr bool

	if len(parsedSql) >= 1 {
		sqltvf, iserr = parsedSql[0].(*qpl_experiment.SQLTVF)
		if !iserr {
			return nil, nil, fmt.Errorf("error parsing query: %w", err)
		}
		fmt.Println(sqltvf.Query)
		if strings.Contains(sqltvf.Query, "select * from input") {
			sqltvf = &qpl_experiment.SQLTVF{
				Query: query,
			}
		}
	} else {
		sqltvf = &qpl_experiment.SQLTVF{
			Query: query,
		}
	}

	stmt, err := db.PrepareContext(context.Background(), sqltvf.Query)
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing query: %w", err)
	}

	rows, err := stmt.QueryContext(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("error executing query: %w", err)
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting column names: %w", err)
	}

	// Prepare holders for row data
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	// Get all rows
	var allRows [][][]byte
	for rows.Next() {
		err = rows.Scan(valuePtrs...)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Convert values to []byte
		byteValues := make([][]byte, len(columns))
		for i, val := range values {
			switch v := val.(type) {
			case []byte:
				byteValues[i] = v
			case string:
				byteValues[i] = []byte(v)
			case nil:
				byteValues[i] = []byte("NULL")
			default:
				byteValues[i] = []byte(fmt.Sprintf("%v", v))
			}
		}
		allRows = append(allRows, byteValues)
	}

	if len(parsedSql) > 1 {
		if aiprompt, iserr := parsedSql[1].(*qpl_experiment.ChatGPTTVF); iserr {
			// Prepare table data for GPT
			var tableData string
			tableData = "Columns: " + strings.Join(columns, ", ") + "\n\nSample Data:\n"

			// Add sample rows (limit to first 5 rows to avoid token limits)
			rowLimit := 15
			if len(allRows) < rowLimit {
				rowLimit = len(allRows)
			}

			for i := 0; i < rowLimit; i++ {
				row := make([]string, len(columns))
				for j, val := range allRows[i] {
					row[j] = string(val)
				}
				tableData += strings.Join(row, ", ") + "\n"
			}

			// Create request body
			requestBody := map[string]interface{}{
				"model": "openai/gpt-4o",
				"messages": []map[string]string{
					{
						"role": "user",
						"content": fmt.Sprintf(
							"Given this table structure (from a partial query) and sample data:\n%s\n\nSuggest 5 useful SQL queries that could provide valuable insights from this data (some idea for it: %s). All suggested queries should start with the existing query: %s. The query uses an experimental Pipe SQL syntax (operator |>) - use it in your suggestions too. An example of Pipe SQL syntax: FROM orders\n|> WHERE order_date >= '2024-01-01'\n|> AGGREGATE SUM(order_amount) AS total_spent GROUP BY customer_id\n|> WHERE total_spent > 1000\n|> SELECT customer_id, total_spent. Note that the SELECT is at the end after GROUP BY - pipe sql has different order - a natual order, where each operator |> is effictively run sequentially. The syntax of AGGREGATE is: |> AGGREGATE ... GROUP BY ... (group is required) For each query, provide:\n1. The SQL query\n2. A description of what the query analyzes and why it would be useful",
							tableData,
							aiprompt.Prompt,
							query,
						),
					},
				},
				"response_format": map[string]interface{}{
					"type": "json_schema",
					"json_schema": map[string]interface{}{
						"name":   "sql_query_suggestions",
						"strict": true,
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"suggestions": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"query": map[string]interface{}{
												"type":        "string",
												"description": "The SQL query",
											},
											"description": map[string]interface{}{
												"type":        "string",
												"description": "Description of what the query analyzes and its usefulness",
											},
										},
										"required":             []string{"query", "description"},
										"additionalProperties": false,
									},
								},
							},
							"required":             []string{"suggestions"},
							"additionalProperties": false,
						},
					},
				},
			}

			jsonBody, err := json.Marshal(requestBody)
			if err != nil {
				return nil, nil, fmt.Errorf("error marshaling request body: %w", err)
			}

			// Create HTTP request
			req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
			if err != nil {
				return nil, nil, fmt.Errorf("error creating request: %w", err)
			}

			req.Header.Set("Authorization", "Bearer ")
			req.Header.Set("Content-Type", "application/json")

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return nil, nil, fmt.Errorf("error sending request: %w", err)
			}
			defer resp.Body.Close()

			// Read response
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return nil, nil, fmt.Errorf("error decoding response: %w", err)
			}

			// Parse suggestions JSON string into struct
			var suggestionsObj struct {
				Suggestions []struct {
					Description string `json:"description"`
					Query       string `json:"query"`
				} `json:"suggestions"`
			}

			suggestionsStr := result["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
			if err := json.Unmarshal([]byte(suggestionsStr), &suggestionsObj); err != nil {
				return nil, nil, fmt.Errorf("error parsing suggestions JSON: %w", err)
			}

			// Set columns
			columns = []string{"description", "query"}

			// Convert suggestions to rows
			allRows = make([][][]byte, 0)
			for _, suggestion := range suggestionsObj.Suggestions {
				row := [][]byte{
					[]byte(suggestion.Description),
					[]byte(strings.ReplaceAll(suggestion.Query, "\n", " ")),
				}
				allRows = append(allRows, row)
			}
		}
	}

	return allRows, columns, nil
}

func (p *PostgresToDuckDBProcessor) Handle(metadata map[string]interface{}, message ...any) (map[string]interface{}, any, error) {
	fmt.Println("PostgresToMySql processor ")
	for _, m := range message {
		msg := m.(pgproto3.FrontendMessage)
		switch msg.(type) {
		case *pgproto3.Query:
			query := msg.(*pgproto3.Query).String
			responses, columns, err := p.respond(query)
			if err != nil {
				log.Fatal(err)
			}

			var fields []pgproto3.FieldDescription
			for _, column := range columns {
				fields = append(fields, pgproto3.FieldDescription{
					Name:                 []byte(column),
					TableOID:             0,
					TableAttributeNumber: 0,
					DataTypeOID:          25,
					DataTypeSize:         -1,
					TypeModifier:         -1,
					Format:               0,
				})
			}

			buf := mustEncode((&pgproto3.RowDescription{Fields: fields}).Encode(nil))
			for _, response := range responses {
				buf = mustEncode((&pgproto3.DataRow{Values: response}).Encode(buf))
			}
			buf = mustEncode((&pgproto3.CommandComplete{CommandTag: []byte("")}).Encode(buf))
			buf = mustEncode((&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf))
			return metadata, buf, nil
		case *pgproto3.Terminate:
			return metadata, nil, nil

		default:
			fmt.Println("Received other than query")
			return metadata, nil, fmt.Errorf("received message other than Query from client: %#v", msg)
		}
	}
	return metadata, nil, nil
}

func mustEncode(buf []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return buf
}

func (p *PostgresToDuckDBProcessor) GetSupportedBackendConnectors() []quesma_api.BackendConnectorType {
	return []quesma_api.BackendConnectorType{quesma_api.MySQLBackend}
}
