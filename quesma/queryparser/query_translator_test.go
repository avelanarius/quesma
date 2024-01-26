package queryparser

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mitmproxy/quesma/model"
	"testing"
)

type Row struct {
}

const searchResponseExpectedString = `
{
  "took": 0,
  "timed_out": false,
  "_shards": {
    "total": 0,
    "successful": 0,
    "failed": 0,
    "failures": null,
    "skipped": 0
  },
  "hits": {
    "total": {
      "value": 0,
      "relation": ""
    },
    "max_score": 0,
    "hits": [
      {
        "_index": "",
        "_id": "",
        "_score": 0,
        "_source": {
          "@timestamp": "2024-01-01"
        },
        "_type": "",
        "sort": null
      }
    ]
  },
  "errors": false,
  "aggregations": null
}
`

func (row Row) String() string {
	return "{ \"@timestamp\" : " + "\"2024-01-01\"}"
}

func TestSearchResponse(t *testing.T) {
	row := []Row{Row{}}

	searchRespBuf, err := MakeResponse(row)
	require.NoError(t, err)
	var searchResponseResult model.SearchResp
	err = json.Unmarshal([]byte(searchRespBuf), &searchResponseResult)
	require.NoError(t, err)
	var searchResponseExpected model.SearchResp
	err = json.Unmarshal([]byte(searchResponseExpectedString), &searchResponseExpected)
	require.NoError(t, err)

	assert.Equal(t, searchResponseExpected, searchResponseResult)
	require.NoError(t, err)
}
