package quesma

import (
	"mitmproxy/quesma/clickhouse"
	"time"
)

// Feel free to suggest a better name for this file

func (cw *ClickhouseQueryTranslator) getAttributesList(tableName string) []clickhouse.Attribute {
	return cw.clickhouseLM.GetAttributesList(tableName)
}

func (cw *ClickhouseQueryTranslator) getFieldInfo(tableName string, fieldName string) clickhouse.FieldInfo {
	return cw.clickhouseLM.GetFieldInfo(tableName, fieldName)
}

// TODO flatten tuples, I think (or just don't support them for now, we don't want them at the moment in production schemas)
func (cw *ClickhouseQueryTranslator) getFieldsList(tableName string) []string {
	return []string{"message"}
}

func (cw *ClickhouseQueryTranslator) queryClickhouse(query string) (int, error) {
	return cw.clickhouseLM.ProcessSelectQuery(query)
}

func (cw *ClickhouseQueryTranslator) getNMostRecentRows(tableName, timestampFieldName string, limit int) ([]clickhouse.QueryResultRow, error) {
	return cw.clickhouseLM.GetNMostRecentRows(tableName, timestampFieldName, limit)
}

func (cw *ClickhouseQueryTranslator) getHistogram(tableName string) ([]clickhouse.HistogramResult, error) {
	return cw.clickhouseLM.GetHistogram(tableName, "timestamp", 15*time.Minute)
}

//lint:ignore U1000 Not used yet
func (cw *ClickhouseQueryTranslator) getAutocompleteSuggestions(tableName, fieldName string, prefix string, limit int) ([]clickhouse.QueryResultRow, error) {
	return cw.clickhouseLM.GetAutocompleteSuggestions(tableName, fieldName, prefix, limit)
}