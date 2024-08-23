// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0
package clickhouse

import (
	"github.com/stretchr/testify/assert"
	"quesma/concurrent"
	"quesma/quesma/config"
	"quesma/quesma/types"
	"testing"
)

func TestAlterTable(t *testing.T) {
	chConfig := &ChTableConfig{
		hasTimestamp:         true,
		timestampDefaultsNow: true,
		engine:               "MergeTree",
		orderBy:              "(timestamp)",
		partitionBy:          "",
		primaryKey:           "",
		ttl:                  "",
		attributes: []Attribute{
			NewDefaultStringAttribute(),
		},
		castUnsupportedAttrValueTypesToString: true,
		preferCastingToOthers:                 true,
	}
	rowsToInsert := []string{
		`{"Test1":1}`,
		`{"Test1":1,"Test2":2}`,
	}
	expectedInsert := []string{
		"{\"attributes_string_key\":[],\"attributes_string_type\":[],\"attributes_string_value\":[],\"Test1\":1}",
		"{\"attributes_string_key\":[],\"attributes_string_type\":[],\"attributes_string_value\":[],\"Test1\":1,\"Test2\":2}",
	}
	alters := []string{
		"ALTER TABLE \"\" ADD COLUMN IF NOT EXISTS \"Test1\" Nullable(Int64)",
		"ALTER TABLE \"\" ADD COLUMN IF NOT EXISTS \"Test2\" Nullable(Int64)",
	}
	columns := []string{"Test1", "Test2"}
	table := &Table{
		Cols: map[string]*Column{},
	}
	fieldsMap := concurrent.NewMapWith("tableName", table)

	lm := NewLogManager(fieldsMap, config.QuesmaConfiguration{})
	for i := range rowsToInsert {
		insert, alter, err := lm.BuildIngestSQLStatements("tableName", types.MustJSON(rowsToInsert[i]), nil, chConfig, true)
		assert.Equal(t, expectedInsert[i], insert)
		assert.Equal(t, alters[i], alter[0])
		// Table will grow with each iteration
		assert.Equal(t, i+1, len(table.Cols))
		for _, col := range columns[:i+1] {
			_, ok := table.Cols[col]
			assert.True(t, ok)
		}
		for k, col := range table.Cols {
			assert.Equal(t, k, col.Name)
			assert.Equal(t, "Nullable", col.Modifiers)
		}

		assert.NoError(t, err)
	}
}