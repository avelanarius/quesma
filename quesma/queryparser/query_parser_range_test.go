package queryparser

import (
	"context"
	"github.com/stretchr/testify/assert"
	"mitmproxy/quesma/clickhouse"
	"mitmproxy/quesma/concurrent"
	"mitmproxy/quesma/quesma/config"
	"testing"
)

type parseRangeTest struct {
	name             string
	rangePartOfQuery QueryMap
	createTableQuery string
	expectedWhere    string
}

var parseRangeTests = []parseRangeTest{
	{
		"DateTime64",
		QueryMap{
			"timestamp": QueryMap{
				"format": "strict_date_optional_time",
				"gte":    "2024-02-02T13:47:16.029Z",
				"lte":    "2024-02-09T13:47:16.029Z",
			},
		},
		`CREATE TABLE ` + tableName + `
		( "message" String, "timestamp" DateTime64(3, 'UTC') )
		ENGINE = Memory`,
		`("timestamp">=parseDateTime64BestEffort('2024-02-02T13:47:16.029Z') AND "timestamp"<=parseDateTime64BestEffort('2024-02-09T13:47:16.029Z'))`,
	},
	{
		"parseDateTimeBestEffort",
		QueryMap{
			"timestamp": QueryMap{
				"format": "strict_date_optional_time",
				"gte":    "2024-02-02T13:47:16.029Z",
				"lte":    "2024-02-09T13:47:16.029Z",
			},
		},
		`CREATE TABLE ` + tableName + `
		( "message" String, "timestamp" DateTime )
		ENGINE = Memory`,
		`("timestamp">=parseDateTimeBestEffort('2024-02-02T13:47:16.029Z') AND "timestamp"<=parseDateTimeBestEffort('2024-02-09T13:47:16.029Z'))`,
	},
	{
		"numeric range",
		QueryMap{
			"time_taken": QueryMap{
				"gt": "100",
			},
		},
		`CREATE TABLE ` + tableName + `
		( "message" String, "timestamp" DateTime, "time_taken" UInt32 )
		ENGINE = Memory`,
		`"time_taken">100`,
	},
	{
		"DateTime64",
		QueryMap{
			"timestamp": QueryMap{
				"format": "strict_date_optional_time",
				"gte":    "2024-02-02T13:47:16",
				"lte":    "2024-02-09T13:47:16",
			},
		},
		`CREATE TABLE ` + tableName + `
		( "message" String, "timestamp" DateTime64(3, 'UTC') )
		ENGINE = Memory`,
		`("timestamp">=parseDateTime64BestEffort('2024-02-02T13:47:16') AND "timestamp"<=parseDateTime64BestEffort('2024-02-09T13:47:16'))`,
	},
}

func Test_parseRange(t *testing.T) {
	for _, test := range parseRangeTests {
		t.Run(test.name, func(t *testing.T) {
			table, err := clickhouse.NewTable(test.createTableQuery, clickhouse.NewNoTimestampOnlyStringAttrCHConfig())
			if err != nil {
				t.Fatal(err)
			}
			assert.NoError(t, err)
			lm := clickhouse.NewLogManager(concurrent.NewMapWith(tableName, table), config.QuesmaConfiguration{})
			cw := ClickhouseQueryTranslator{ClickhouseLM: lm, Table: table, Ctx: context.Background()}

			simpleQuery := cw.parseRange(test.rangePartOfQuery)
			assert.Equal(t, test.expectedWhere, simpleQuery.WhereClauseAsString())
		})
	}
}