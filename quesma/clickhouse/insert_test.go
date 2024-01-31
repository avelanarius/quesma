package clickhouse

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// So far this file tests:
// 1) creating table through insert in 3 cases:
//   a) schema doesn't exist
//   b) schema exists in our memory (e.g. is predefined), but isn't created in ClickHouse, so CREATE TABLE needs to be sent
//   c) schema exists both in our memory and in ClickHouse
// 2) inserting into table (building insert query string with/without attrs)
// 3) that predefined schema trumps (is more important) schema from insert's JSON

const tableName = "test_table"

var insertTests = []struct {
	name                  string
	insertJson            string
	createTableLines      []string // those and only those lines should be in create table query
	createTableLinesAttrs []string
}{
	{
		"insert fields agree with schema",
		`{"@timestamp":"2024-01-27T16:11:19.94Z","host.name":"hermes","message":"User password reset failed","service.name":"frontend","severity":"debug","source":"rhel"}`,
		[]string{
			`CREATE TABLE IF NOT EXISTS "test_table"`,
			`(`,
			`	`,
			`	"@timestamp" DateTime64`,
			`	"host.name" String`,
			`	"message" String`,
			`	"service.name" String`,
			`	"severity" String`,
			`	"source" String`,
			`	INDEX severity_idx severity TYPE set(25) GRANULARITY 4`,
			`	`,
			``,
			`)`,
			`ENGINE = MergeTree`,
			`ORDER BY ("@timestamp")`,
		},
		[]string{
			`"attributes_float64_value" Array(Float64),`,
			`"attributes_float64_key" Array(String),`,
			`"attributes_string_value" Array(Float64),`,
			`"attributes_string_key" Array(String),`,
			`"attributes_int64_value" Array(Float64),`,
			`"attributes_int64_key" Array(String),`,
			`"attributes_bool_value" Array(Float64),`,
			`"attributes_bool_key" Array(String),`,
			``,
		},
	},
	{
		"insert fields disagree with schema",
		`{"@timestamp":"2024-01-27T16:11:19.94Z","host.name":"hermes","message":"User password reset failed","random1":["debug"],"random2":"random-string","severity":"frontend"}`,
		[]string{
			`CREATE TABLE IF NOT EXISTS "test_table"`,
			`(`,
			`	`,
			`	"@timestamp" DateTime64`,
			`	"host.name" String`,
			`	"message" String`,
			`	"random1" Array(String)`,
			`	"random2" string`,
			`	"severity" String`,
			`	INDEX severity_idx severity TYPE set(25) GRANULARITY 4`,
			`	`,
			``,
			`)`,
			`ENGINE = MergeTree`,
			`ORDER BY ("@timestamp")`,
		},
		[]string{
			`"attributes_float64_value" Array(Float64),`,
			`"attributes_float64_key" Array(String),`,
			`"attributes_string_value" Array(Float64),`,
			`"attributes_string_key" Array(String),`,
			`"attributes_int64_value" Array(Float64),`,
			`"attributes_int64_key" Array(String),`,
			`"attributes_bool_value" Array(Float64),`,
			`"attributes_bool_key" Array(String),`,
			``,
		},
	},
}

var configs = []*ChTableConfig{
	NewCHTableConfigNoAttrs(),
	NewDefaultCHConfig(),
}

var expectedInserts = []string{
	`INSERT INTO "` + tableName + `" FORMAT JSONEachRow ` + insertTests[0].insertJson,
	`INSERT INTO "` + tableName + `" FORMAT JSONEachRow {"attributes_string_key":\["service.name","severity","source"\],"attributes_string_value":\["frontend","debug","rhel"\],"@timestamp":"2024-01-27T16:11:19.94Z","host.name":"hermes","message":"User password reset failed"}`,
	`INSERT INTO "` + tableName + `" FORMAT JSONEachRow ` + strings.Replace(strings.Replace(insertTests[1].insertJson, "[", `\[`, 1), "]", `\]`, 1),
	`INSERT INTO "` + tableName + `" FORMAT JSONEachRow {"attributes_string_key":\["random1","random2","severity"\],"attributes_string_value":\["\[debug\]","random-string","frontend"\],"@timestamp":"2024-01-27T16:11:19.94Z","host.name":"hermes","message":"User password reset failed"}`,
}

type logManagerHelper struct {
	lm                  *LogManager
	tableAlreadyCreated bool
}

func logManagersNonEmpty(config *ChTableConfig) []logManagerHelper {
	lms := make([]logManagerHelper, 0, 4)
	for _, created := range []bool{true, false} {
		for _, predefinedNotRuntime := range []bool{true, false} {
			empty := make(TableMap)
			full := TableMap{
				tableName: &Table{
					Name:   tableName,
					Config: config,
					Cols: map[string]*Column{
						"@timestamp":       dateTime("@timestamp"),
						"host.name":        genericString("host.name"),
						"message":          lowCardinalityString("message"),
						"non-insert-field": genericString("non-insert-field"),
					},
					Created: created,
				},
			}
			if predefinedNotRuntime {
				lms = append(lms, logManagerHelper{NewLogManager(full, empty), created})
			} else {
				lms = append(lms, logManagerHelper{NewLogManager(empty, full), created})
			}
		}
	}
	return lms
}

func logManagers(config *ChTableConfig) []logManagerHelper {
	return append([]logManagerHelper{{NewLogManagerEmpty(), false}}, logManagersNonEmpty(config)...)
}

func TestAutomaticTableCreationAtInsert(t *testing.T) {
	for index1, tt := range insertTests {
		for index2, config := range configs {
			for index3, lm := range logManagers(config) {
				t.Run("case insertTest["+strconv.Itoa(index1)+"], config["+strconv.Itoa(index2)+"], logManager["+strconv.Itoa(index3)+"]", func(t *testing.T) {
					query, err := buildCreateTableQueryNoOurFields(tableName, tt.insertJson, config)
					assert.NoError(t, err)
					table, err := NewTable(query, config)
					assert.NoError(t, err)
					query = addOurFieldsToCreateTableQuery(query, config, table)

					// check if CREATE TABLE string is OK
					queryByLine := strings.Split(query, "\n")
					if len(config.attributes) > 0 {
						assert.Equal(t, len(tt.createTableLines)+2*len(config.attributes)+1, len(queryByLine))
						for _, line := range tt.createTableLines {
							assert.True(t, slices.Contains(tt.createTableLines, line) || slices.Contains(tt.createTableLinesAttrs, line))
						}
					} else {
						assert.Equal(t, len(tt.createTableLines), len(queryByLine))
						for _, line := range tt.createTableLines {
							assert.Contains(t, tt.createTableLines, line)
						}
					}
					logManagerEmpty := len(lm.lm.newRuntimeTables) == 0 && len(lm.lm.predefinedTables) == 0

					// check if we properly create table in our tables table :) (:) suggested by Copilot) if needed
					tableInMemory := lm.lm.findSchema(tableName)
					needCreate := true
					if tableInMemory != nil && tableInMemory.Created {
						needCreate = false
					}
					noSuchTable := lm.lm.addSchemaIfDoesntExist(table)
					assert.Equal(t, needCreate, noSuchTable)

					// and Created is set to true
					tableInMemory = lm.lm.findSchema(tableName)
					assert.NotNil(t, tableInMemory)
					assert.True(t, tableInMemory.Created)

					// and we have a schema in memory in every case
					assert.Equal(t, 1, len(lm.lm.newRuntimeTables)+len(lm.lm.predefinedTables))

					// and that schema in memory is what it should be (predefined, if it was predefined, new if it was new)
					if logManagerEmpty {
						assert.Equal(t, 6+2*len(config.attributes), len(lm.lm.newRuntimeTables[tableName].Cols))
					} else if len(lm.lm.predefinedTables) > 0 {
						assert.Equal(t, 4, len(lm.lm.predefinedTables[tableName].Cols))
					} else {
						assert.Equal(t, 4, len(lm.lm.newRuntimeTables[tableName].Cols))
					}
				})
			}
		}
	}
}

func TestProcessInsertQuery(t *testing.T) {
	for index1, tt := range insertTests {
		for index2, config := range configs {
			for index3, lm := range logManagers(config) {
				t.Run("case insertTest["+strconv.Itoa(index1)+"], config["+strconv.Itoa(index2)+"], logManager["+strconv.Itoa(index3)+"]", func(t *testing.T) {
					db, mock, err := sqlmock.New()
					assert.NoError(t, err)
					lm.lm.db = db
					defer db.Close()

					// info: result values aren't important, this '.WillReturnResult[...]' just needs to be there
					if !lm.tableAlreadyCreated {
						// we check here if we try to create table from predefined schema, not from insert's JSON
						if len(lm.lm.predefinedTables) > 0 || len(lm.lm.newRuntimeTables) > 0 {
							mock.ExpectExec(`CREATE TABLE IF NOT EXISTS "` + tableName + `.*non-insert-field`).WillReturnResult(sqlmock.NewResult(0, 0))
						} else {
							mock.ExpectExec(`CREATE TABLE IF NOT EXISTS "` + tableName).WillReturnResult(sqlmock.NewResult(0, 0))
						}
					}
					if len(config.attributes) == 0 || (len(lm.lm.predefinedTables) == 0 && len(lm.lm.newRuntimeTables) == 0) {
						mock.ExpectExec(expectedInserts[2*index1]).WillReturnResult(sqlmock.NewResult(545, 54))
					} else {
						mock.ExpectExec(expectedInserts[2*index1+1]).WillReturnResult(sqlmock.NewResult(1, 1))
					}

					err = lm.lm.ProcessInsertQuery(tableName, tt.insertJson)
					assert.NoError(t, err)
					if err := mock.ExpectationsWereMet(); err != nil {
						t.Fatal("there were unfulfilled expections:", err)
					}
				})
			}
		}
	}
}