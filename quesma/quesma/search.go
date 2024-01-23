package quesma

import (
	"fmt"
	"mitmproxy/quesma/clickhouse"
	"strconv"
)

func handleSearch(index string, body []byte, lm *clickhouse.LogManager,
	responseMatcher *ResponseMatcher,
	queryDebugger *QueryDebugger,
	requestId string) {
	// TODO: implement
	var translatedQueryBody []byte
	queryTranslator := &ClickhouseQueryTranslator{clickhouseLM: lm}

	// TODO index argument is not used yet
	_ = index

	// old TODO: query clickhouse
	// get response
	// and translate

	query := queryTranslator.Write(body)
	var responseBody []byte
	if query.canParse {
		cnt, err := queryTranslator.queryClickhouse(query.sql)
		if err != nil {
			responseBody = []byte("Error processing query: " + query.sql + ", err: " + err.Error())
		}
		rows, err := queryTranslator.getNMostRecentRows(tableName, "@timestamp", 2)
		if err == nil {
			fmt.Println(rows)
		} else {
			fmt.Println(err)
		}
		responseTranslator := &ClickhouseResultReader{clickhouseLM: lm}
		responseTranslator.Read(responseBody) // TODO implement this, not line below
		responseBody = []byte(strconv.Itoa(cnt))
		histogram, err := queryTranslator.getHistogram(tableName)
		fmt.Printf("Histogram: %+v, err: %+v\n", histogram, err)
	} else {
		responseBody = []byte("Invalid Query, err: " + query.sql)
	}

	var rawResults []byte
	responseMatcher.Push(&QResponse{requestId, responseBody})
	translatedQueryBody = []byte(query.sql)
	queryDebugger.PushSecondaryInfo(&QueryDebugSecondarySource{
		id:                     requestId,
		incomingQueryBody:      body,
		queryBodyTranslated:    translatedQueryBody,
		queryRawResults:        rawResults,
		queryTranslatedResults: responseBody,
	})
}