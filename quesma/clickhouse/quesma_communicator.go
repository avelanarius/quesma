package clickhouse

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Implementation of API for Quesma

type FieldInfo int

const (
	NotExists FieldInfo = iota
	ExistsAndIsBaseType
	ExistsAndIsArray
)

type QueryResultCol struct {
	ColName string
	Value   interface{}
}

type QueryResultRow struct {
	Cols []QueryResultCol
}

type HistogramResult struct {
	start time.Time
	end   time.Time
	count int
}

func (c QueryResultCol) String() string {
	switch c.Value.(type) {
	case string:
		return fmt.Sprintf(`"%s": "%v"`, c.ColName, c.Value)
	default:
		return fmt.Sprintf(`"%s": %v`, c.ColName, c.Value)
	}
}

func (r QueryResultRow) String() string {
	str := strings.Builder{}
	str.WriteString(indent(1) + "{\n")
	for _, col := range r.Cols {
		str.WriteString(indent(2) + col.String() + ",\n")
	}
	str.WriteString("\n" + indent(1) + "}\n")
	return str.String()
}

func (hr HistogramResult) String() string {
	return fmt.Sprintf("%v - %v, count: %v", hr.start, hr.end, hr.count)
}

// (int, error) just for the 1st version. Should be changed to something more: rows, etc.
func (lm *LogManager) ProcessSelectQuery(query string) (int, error) {
	if lm.db == nil {
		connection, err := sql.Open("clickhouse", url)
		if err != nil {
			return -1, fmt.Errorf("open >> %v", err)
		}
		lm.db = connection
	}

	query = strings.Replace(query, "SELECT *", "SELECT count(*)", 1)
	rows, err := lm.db.Query(query)
	if err != nil {
		return -1, fmt.Errorf("query >> %v", err)
	}
	var cnt int
	if !rows.Next() {
		return -1, fmt.Errorf("no rows")
	}
	err = rows.Scan(&cnt)
	if err != nil {
		return -1, fmt.Errorf("scan >> %v", err)
	}
	return cnt, nil
}

func (lm *LogManager) GetAttributesList(tableName string) []Attribute {
	table := lm.findSchema(tableName)
	if table == nil {
		return make([]Attribute, 0)
	}
	return table.Config.attributes
}

// TODO Won't work with tuples, e.g. trying to access via tupleName.tupleField will return NotExists,
// instead of some other response. Fix this when needed (we seem to not need tuples right now)
func (lm *LogManager) GetFieldInfo(tableName string, fieldName string) FieldInfo {
	table := lm.findSchema(tableName)
	if table == nil {
		return NotExists
	}
	col, ok := table.Cols[fieldName]
	if !ok {
		return NotExists
	}
	if col.isArray() {
		return ExistsAndIsArray
	}
	return ExistsAndIsBaseType
}

// TODO again, fix tuples.
// t tuple(a String, b String) should return [t.a, t.b], now returns [t]
func (lm *LogManager) GetFieldsList(tableName string) []string {
	table := lm.findSchema(tableName)
	if table == nil {
		return make([]string, 0)
	}
	fieldNames := make([]string, 0, len(table.Cols))
	for colName := range table.Cols {
		fieldNames = append(fieldNames, colName)
	}
	return fieldNames
}

func (lm *LogManager) GetNMostRecentRows(tableName, timestampFieldName string, N int) ([]QueryResultRow, error) {
	table := lm.findSchema(tableName)
	if table == nil {
		table = lm.findSchema(tableName[1 : len(tableName)-1]) // try remove " " TODO improve this when we get out of the prototype phase
		if table == nil {
			return nil, fmt.Errorf("Table " + tableName + " not found")
		}
	}

	if lm.db == nil {
		connection, err := sql.Open("clickhouse", url)
		if err != nil {
			return nil, fmt.Errorf("open >> %v", err)
		}
		lm.db = connection
	}

	queryStr := strings.Builder{}
	queryStr.WriteString("SELECT ")
	row := make([]interface{}, 0, len(table.Cols))
	colNames := make([]string, 0, len(table.Cols))
	for colName, col := range table.Cols {
		colNames = append(colNames, fmt.Sprintf("\"%s\"", colName))
		if col.Type.isBool() {
			queryStr.WriteString("toInt8(" + fmt.Sprintf("\"%s\"", colName) + "),")
		} else {
			queryStr.WriteString(fmt.Sprintf("\"%s\"", colName) + ",")
		}
		row = append(row, col.Type.newZeroValue())
	}

	queryStr.WriteString(" FROM " + tableName + " ORDER BY " + fmt.Sprintf("\"%s\"", timestampFieldName) + " DESC LIMIT " + strconv.Itoa(N))
	fmt.Println("query string: ", queryStr.String())
	rowsDB, err := lm.db.Query(queryStr.String())
	if err != nil {
		return nil, fmt.Errorf("query >> %v", err)
	}

	rowDB := make([]interface{}, len(table.Cols))
	for i := 0; i < len(table.Cols); i++ {
		rowDB[i] = &row[i]
	}

	rows := make([]QueryResultRow, 0, N)
	for rowsDB.Next() {
		err = rowsDB.Scan(rowDB...)
		if err != nil {
			return nil, fmt.Errorf("scan >> %v", err)
		}
		resultRow := QueryResultRow{Cols: make([]QueryResultCol, 0, len(table.Cols))}
		for i, v := range row {
			resultRow.Cols = append(resultRow.Cols, QueryResultCol{ColName: colNames[i], Value: v})
		}
		rows = append(rows, resultRow)
	}

	return rows, nil
}

func (lm *LogManager) GetHistogram(tableName, timestampFieldName string, duration time.Duration) ([]HistogramResult, error) {
	if lm.db == nil {
		connection, err := sql.Open("clickhouse", url)
		if err != nil {
			return nil, err
		}
		lm.db = connection
	}

	histogramOneBar := durationToHistogramInterval(duration) // 1 bar duration
	gbyStmt := "toInt64(toUnixTimestamp64Milli(" + timestampFieldName + ")/" + strconv.FormatInt(histogramOneBar.Milliseconds(), 10) + ")"
	whrStmt := timestampFieldName + ">=timestamp_sub(SECOND," + strconv.FormatInt(int64(duration.Seconds()), 10) + ", now64())"
	query := "SELECT " + gbyStmt + ", count() FROM " + tableName + " WHERE " + whrStmt + " GROUP BY " + gbyStmt
	rows, err := lm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query >> %v", err)
	}
	histogram := make([]HistogramResult, 0)
	for rows.Next() {
		var start int64
		var count int
		err = rows.Scan(&start, &count)
		if err != nil {
			return nil, fmt.Errorf("scan >> %v", err)
		}
		startMs := start * histogramOneBar.Milliseconds()
		endMs := startMs + histogramOneBar.Milliseconds()
		histogram = append(histogram, HistogramResult{
			start: time.Unix(startMs/1_000, (startMs%1_000)*1_000_000),
			end:   time.Unix(endMs/1_000, (endMs%1_000)*1_000_000),
			count: count,
		})
	}
	sort.Slice(histogram, func(i, j int) bool {
		return histogram[i].start.Before(histogram[j].start)
	})
	return histogram, nil
}

/*
		How Kibana shows histogram (how long one bar is):
	    query duration -> one histogram's bar ...
	    10s  -> 200ms
		14s  -> 280ms
		20s  -> 400ms
		24s  -> 480ms
		25s  -> 1s
		[25s, 4m]   -> 1s
		[5m, 6m]    -> 5s
		[7m, 12m]   -> 10s
		[13m, 37m]  -> 30s
		[38m, 140m] -> 1m
		[150m, 7h]  -> 5m
		[8h, 16h]   -> 10m
		[17h, 37h]  -> 30m
		[38h, 99h]  -> 1h
		[100h, 12d] -> 3h
		[13d, 49d]  -> 12h
		[50d, 340d] -> 1d
		[350d, 34m] -> 7d
		[35m, 15y]  -> 1m
*/

func durationToHistogramInterval(d time.Duration) time.Duration {
	switch {
	case d < 25*time.Second:
		ms := d.Milliseconds() / 50
		ms += 20 - (ms % 20)
		return time.Millisecond * time.Duration(ms)
	case d <= 4*time.Minute:
		return time.Second
	case d < 7*time.Minute:
		return 5 * time.Second
	case d < 13*time.Minute:
		return 10 * time.Second
	case d < 38*time.Minute:
		return 30 * time.Second
	case d <= 140*time.Minute:
		return time.Minute
	case d <= 7*time.Hour:
		return 5 * time.Minute
	case d <= 16*time.Hour:
		return 10 * time.Minute
	case d <= 37*time.Hour:
		return 30 * time.Minute
	case d <= 99*time.Hour:
		return time.Hour
	case d <= 12*24*time.Hour:
		return 3 * time.Hour
	case d <= 49*24*time.Hour:
		return 12 * time.Hour
	case d <= 340*24*time.Hour:
		return 24 * time.Hour
	case d <= 35*30*24*time.Hour:
		return 7 * 24 * time.Hour
	default:
		return 30 * 24 * time.Hour
	}
}

// TODO make it faster? E.g. not search in all rows?
// TODO add support for autocomplete for attributes, if we'll find it needed
func (lm *LogManager) GetAutocompleteSuggestions(tableName, fieldName, prefix string, limit int) ([]QueryResultRow, error) {
	table := lm.findSchema(tableName)
	if table == nil {
		table = lm.findSchema(tableName[1 : len(tableName)-1]) // try remove " " TODO improve this when we get out of the prototype phase
		if table == nil {
			return nil, fmt.Errorf("Table " + tableName + " not found")
		}
	}

	if lm.db == nil {
		connection, err := sql.Open("clickhouse", url)
		if err != nil {
			return nil, err
		}
		lm.db = connection
	}

	// TODO add support for autocomplete for attributes, if we'll find it needed
	col, ok := table.Cols[fieldName]
	if !ok {
		return nil, fmt.Errorf("Column " + fieldName + " not found")
	}

	query := "SELECT DISTINCT " + fieldName + " FROM " + tableName
	if prefix != "" {
		if !col.Type.isString() {
			query += " WHERE toString(" + fieldName + ")"
		} else {
			query += " WHERE " + fieldName
		}
		query += " LIKE '" + prefix + "%'"
	}
	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}
	rowsDB, err := lm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query >> %v", err)
	}

	value := col.Type.newZeroValue()
	rows := make([]QueryResultRow, 0)
	for rowsDB.Next() {
		err = rowsDB.Scan(&value)
		if err != nil {
			return nil, fmt.Errorf("scan >> %v", err)
		}
		rows = append(rows, QueryResultRow{Cols: []QueryResultCol{{ColName: fieldName, Value: value}}})
	}
	return rows, nil
}