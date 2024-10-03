// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0
package quesma

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"quesma/ab_testing"
	"quesma/clickhouse"
	"quesma/common_table"
	"quesma/logger"
	"quesma/quesma/config"
	"quesma/quesma/types"
	"quesma/quesma/ui"
	"quesma/schema"
	"quesma/telemetry"
	"testing"
)

func TestSearchCommonTable(t *testing.T) {

	tests := []struct {
		Name         string
		QueryJson    string
		WantedSql    []string // array because of non-determinism
		Rows         []*sqlmock.Rows
		IndexPattern string
	}{

		{ // [0]
			Name:         "query non virtual table",
			IndexPattern: "logs-3",
			QueryJson: `
        {
          "query": {
            "match_all": {}
          },
          "track_total_hits": false,
          "runtime_mappings": {
        }
}`,
			WantedSql: []string{
				`SELECT "@timestamp", "message" FROM "logs-3" LIMIT 10`,
			},
		},

		{
			Name:         "query virtual table",
			IndexPattern: "logs-1",
			QueryJson: `
        {
          "query": {
            "match_all": {}
          },
          "track_total_hits": false,
          "runtime_mappings": {
        }
}`,
			WantedSql: []string{
				`SELECT "@timestamp", "message" FROM quesma_common_table WHERE "__quesma_index_name"='logs-1' LIMIT 10`,
			},
		},

		{
			Name:         "query virtual tables",
			IndexPattern: "logs-1,logs-2",
			QueryJson: `
        {
          "query": {
            "match_all": {}
          },
          "track_total_hits": false,
          "runtime_mappings": {
        }
}`,
			WantedSql: []string{
				`SELECT "@timestamp", "message", "__quesma_index_name" FROM quesma_common_table WHERE ("__quesma_index_name"='logs-1' OR "__quesma_index_name"='logs-2') LIMIT 10`,
			},
		},

		{
			Name:         "query all logs - we query only virtual tables",
			IndexPattern: "logs-*",
			QueryJson: `
        {
          "query": {
            "match_all": {}
          },
          "track_total_hits": false,
          "runtime_mappings": {
        }
}`,
			WantedSql: []string{
				`SELECT "@timestamp", "message", "__quesma_index_name" FROM quesma_common_table WHERE ("__quesma_index_name"='logs-1' OR "__quesma_index_name"='logs-2') LIMIT 10`,
			},
		},

		{
			Name:         "query all - we query only virtual tables",
			IndexPattern: "*",
			QueryJson: `
        {
          "query": {
            "match_all": {}
          },
          "track_total_hits": false,
          "runtime_mappings": {
        }
}`,
			WantedSql: []string{
				`SELECT "@timestamp", "message", "__quesma_index_name" FROM quesma_common_table WHERE ("__quesma_index_name"='logs-1' OR "__quesma_index_name"='logs-2') LIMIT 10`,
			},
		},

		{
			Name:         "aggregation query",
			IndexPattern: "*",
			QueryJson: `
        {
          "query": {
            "match_all": {}
          },

          "aggs": {
				"2": {
					"date_range": {
						"field": "timestamp",
						"ranges": [
							{
								"to": "now"
							},
							{
								"from": "now-3w/d",
								"to": "now"
							},
							{
								"from": "2024-04-14"
							}
						],
						"time_zone": "Europe/Warsaw"
					}
				}
			},
		  	

          "track_total_hits": false,
          "runtime_mappings": {
        }
}`,
			WantedSql: []string{
				`SELECT countIf("@timestamp"<toInt64(toUnixTimestamp(now()))) AS "range_0__aggr__2__count", countIf(("@timestamp">=toInt64(toUnixTimestamp(toStartOfDay(subDate(now(), INTERVAL 3 week)))) AND "@timestamp"<toInt64(toUnixTimestamp(now())))) AS "range_1__aggr__2__count", countIf("@timestamp">=toInt64(toUnixTimestamp('2024-04-14'))) AS "range_2__aggr__2__count" FROM quesma_common_table WHERE ("__quesma_index_name"='logs-1' OR "__quesma_index_name"='logs-2') -- optimizations: pancake(half)`,
				`SELECT "@timestamp", "message", "__quesma_index_name" FROM quesma_common_table WHERE ("__quesma_index_name"='logs-1' OR "__quesma_index_name"='logs-2') LIMIT 10`,
			},
			// we need to return some rows, otherwise pancakes will fail
			Rows: []*sqlmock.Rows{
				sqlmock.NewRows([]string{"range_0__aggr__2__count", "range_1__aggr__2__count", "range_2__aggr__2__count"}).AddRow(1, 2, 3),
				sqlmock.NewRows([]string{"@timestamp", "message", "__quesma_index_name"}).AddRow("2024-04-14", "message", "logs-1"),
			},
		},
	}

	quesmaConfig := &config.QuesmaConfiguration{
		IndexConfig: map[string]config.IndexConfiguration{
			"logs-1": {
				UseCommonTable: true,
				QueryTarget:    []string{config.ClickhouseTarget},
			},
			"logs-2": {
				UseCommonTable: true,
				QueryTarget:    []string{config.ClickhouseTarget},
			},
			"logs-3": {
				UseCommonTable: false,
				QueryTarget:    []string{config.ClickhouseTarget},
			},
		},
	}

	schemaRegistry := schema.StaticRegistry{
		Tables: make(map[schema.TableName]schema.Schema),
	}
	tableMap := clickhouse.NewTableMap()

	schemaRegistry.Tables["logs-1"] = schema.Schema{
		Fields: map[schema.FieldName]schema.Field{
			"@timestamp": {PropertyName: "@timestamp", InternalPropertyName: "@timestamp", Type: schema.QuesmaTypeDate},
			"message":    {PropertyName: "message", InternalPropertyName: "message", Type: schema.QuesmaTypeKeyword},
		},
	}

	tableMap.Store("logs-1", &clickhouse.Table{
		Name: "logs-1",
		Cols: map[string]*clickhouse.Column{
			"@timestamp": {Name: "@timestamp"},
			"message":    {Name: "message"},
		},
		VirtualTable: true,
	})

	schemaRegistry.Tables["logs-2"] = schema.Schema{
		Fields: map[schema.FieldName]schema.Field{
			"@timestamp": {PropertyName: "@timestamp", InternalPropertyName: "@timestamp", Type: schema.QuesmaTypeDate},
			"message":    {PropertyName: "message", InternalPropertyName: "message", Type: schema.QuesmaTypeKeyword},
		},
	}

	tableMap.Store("logs-2", &clickhouse.Table{
		Name: "logs-2",
		Cols: map[string]*clickhouse.Column{
			"@timestamp": {Name: "@timestamp"},
			"message":    {Name: "message"},
		},
		VirtualTable: true,
	})

	schemaRegistry.Tables["logs-3"] = schema.Schema{
		Fields: map[schema.FieldName]schema.Field{
			"@timestamp": {PropertyName: "@timestamp", InternalPropertyName: "@timestamp", Type: schema.QuesmaTypeDate},
			"message":    {PropertyName: "message", InternalPropertyName: "message", Type: schema.QuesmaTypeKeyword},
		},
	}

	tableMap.Store("logs-3", &clickhouse.Table{
		Name: "logs-3",
		Cols: map[string]*clickhouse.Column{
			"@timestamp": {Name: "@timestamp"},
			"message":    {Name: "message"},
		},
		VirtualTable: false,
	})

	schemaRegistry.Tables[common_table.TableName] = schema.Schema{
		Fields: map[schema.FieldName]schema.Field{
			"@timestamp":                 {PropertyName: "@timestamp", InternalPropertyName: "@timestamp", Type: schema.QuesmaTypeDate},
			"message":                    {PropertyName: "message", InternalPropertyName: "message", Type: schema.QuesmaTypeKeyword},
			common_table.IndexNameColumn: {PropertyName: common_table.IndexNameColumn, InternalPropertyName: common_table.IndexNameColumn, Type: schema.QuesmaTypeKeyword},
		},
	}

	tableMap.Store(common_table.TableName, &clickhouse.Table{
		Name: common_table.TableName,
		Cols: map[string]*clickhouse.Column{
			"@timestamp":                 {Name: "@timestamp"},
			"message":                    {Name: "message"},
			common_table.IndexNameColumn: {Name: common_table.IndexNameColumn},
		},
	})

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s(%d)", tt.Name, i), func(t *testing.T) {

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}

			defer db.Close()

			indexManagement := NewFixedIndexManagement()
			lm := clickhouse.NewLogManagerWithConnection(db, tableMap)
			managementConsole := ui.NewQuesmaManagementConsole(quesmaConfig, nil, indexManagement, make(<-chan logger.LogWithLevel, 50000), telemetry.NewPhoneHomeEmptyAgent(), nil)

			for i, query := range tt.WantedSql {

				rows := sqlmock.NewRows([]string{"@timestamp", "message"})
				if tt.Rows != nil {
					rows = tt.Rows[i]
				}

				mock.ExpectQuery(query).WillReturnRows(rows)
			}
			queryRunner := NewQueryRunner(lm, quesmaConfig, indexManagement, managementConsole, schemaRegistry, ab_testing.NewEmptySender())
			queryRunner.maxParallelQueries = 0

			_, err = queryRunner.handleSearch(ctx, tt.IndexPattern, types.MustJSON(tt.QueryJson))

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal("there were unfulfilled expections:", err)
			}
		})
	}
}