package queryparser

import (
	"github.com/stretchr/testify/assert"
	"mitmproxy/quesma/clickhouse"
	"mitmproxy/quesma/util"
	"strconv"
	"testing"
)

// Simple unit tests, testing only "aggs" part of the request json query

const tableName = "logs-generic-default"
const tableNameQuoted = `"` + tableName + `"`

var aggregationTests = []struct {
	aggregationJson string
	translatedSqls  []string
}{
	{ // [0]
		`
		{
			"aggs": {
				"maxAgg": {
					"max": {
						"field": "AvgTicketPrice"
					}
				},
				"minAgg": {
					"min": {
						"field": "AvgTicketPrice"
					}
				}
			}
		}`,
		[]string{
			`SELECT max("AvgTicketPrice") FROM ` + tableNameQuoted + ` `,
			`SELECT min("AvgTicketPrice") FROM ` + tableNameQuoted + ` `,
		},
	},
	{ // [1]
		`
		{
			"aggs": {
				"0": {
					"aggs": {
						"1-bucket": {
							"filter": {
								"bool": {
									"filter": [
										{
											"bool": {
												"minimum_should_match": 1,
												"should": [
													{
														"match": {
															"FlightDelay": true
														}
													}
												]
											}
										}
									],
									"must": [],
									"must_not": [],
									"should": []
								}
							}
						},
						"3-bucket": {
							"filter": {
								"bool": {
									"filter": [
										{
											"bool": {
												"minimum_should_match": 1,
												"should": [
													{
														"match": {
															"Cancelled": true
														}
													}
												]
											}
										}
									],
									"must": [],
									"must_not": [],
									"should": []
								}
							}
						}
					},
					"terms": {
						"field": "OriginCityName",
						"order": {
							"_key": "asc"
						},
						"size": 1000
					}
				}
			}
		}`,
		[]string{
			`SELECT "OriginCityName", count() FROM ` + tableNameQuoted + `  GROUP BY ("OriginCityName")`,
			`SELECT "OriginCityName", count() FROM ` + tableNameQuoted + ` WHERE "Cancelled" == true  GROUP BY ("OriginCityName")`,
			`SELECT "OriginCityName", count() FROM ` + tableNameQuoted + ` WHERE "FlightDelay" == true  GROUP BY ("OriginCityName")`,
		},
	},
	{ // [2]
		`
		{
			"aggs": {
				"0": {
					"aggs": {
						"1": {
							"date_histogram": {
								"extended_bounds": {
									"max": 1707486436029,
									"min": 1706881636029
								},
								"field": "timestamp",
								"fixed_interval": "3h",
								"time_zone": "Europe/Warsaw"
							}
						}
					},
					"terms": {
						"field": "FlightDelayType",
						"order": {
							"_count": "desc"
						},
						"shard_size": 25,
						"size": 10
					}
				}
			}
		}`,
		[]string{
			`SELECT "FlightDelayType", count() FROM ` + tableNameQuoted + `  GROUP BY ("FlightDelayType")`,
			`SELECT "FlightDelayType", "timestamp", count() FROM ` + tableNameQuoted + `  GROUP BY ("FlightDelayType", "timestamp")`,
		},
	},
	{ // [3]
		`
		{
			"aggs": {
				"0-bucket": {
					"filter": {
						"bool": {
							"filter": [
								{
									"bool": {
										"minimum_should_match": 1,
										"should": [
											{
												"match": {
													"FlightDelay": true
												}
											}
										]
									}
								}
							],
							"must": [],
							"must_not": [],
							"should": []
						}
					}
				}
			}
		}`,
		[]string{`SELECT count() FROM ` + tableNameQuoted + ` WHERE "FlightDelay" == true `},
	},
	{ // [4]
		`
		{
			"aggs": {
				"time_offset_split": {
					"aggs": {},
					"filters": {
						"filters": {
							"0": {
								"range": {
									"timestamp": {
										"format": "strict_date_optional_time",
										"gte": "2024-02-02T13:47:16.029Z",
										"lte": "2024-02-09T13:47:16.029Z"
									}
								}
							},
							"604800000": {
								"range": {
									"timestamp": {
										"format": "strict_date_optional_time",
										"gte": "2024-01-26T13:47:16.029Z",
										"lte": "2024-02-02T13:47:16.029Z"
									}
								}
							}
						}
					}
				}
			}
		}`,
		[]string{
			`SELECT count() FROM ` + tableNameQuoted + ` WHERE "timestamp">=parseDateTime64BestEffort('2024-02-02T13:47:16.029Z') AND "timestamp"<=parseDateTime64BestEffort('2024-02-09T13:47:16.029Z') `,
			`SELECT count() FROM ` + tableNameQuoted + ` WHERE "timestamp">=parseDateTime64BestEffort('2024-01-26T13:47:16.029Z') AND "timestamp"<=parseDateTime64BestEffort('2024-02-02T13:47:16.029Z') `,
		},
	},
	{ // [5]
		`
		{
			"aggs": {
				"0": {
					"histogram": {
						"field": "FlightDelayMin",
						"interval": 1,
						"min_doc_count": 1
					}
				}
			}
		}`,
		[]string{`SELECT "FlightDelayMin", count() FROM ` + tableNameQuoted + `  GROUP BY ("FlightDelayMin")`},
	},
	{ // [6]
		`
		{
			"aggs": {
				"origins": {
					"aggs": {
						"distinations": {
							"aggs": {
								"destLocation": {
									"top_hits": {
										"_source": {
											"includes": [
												"DestLocation"
											]
										},
										"size": 1
									}
								}
							},
							"terms": {
								"field": "DestAirportID",
								"size": 10000
							}
						},
						"originLocation": {
							"top_hits": {
								"_source": {
									"includes": [
										"OriginLocation",
										"Origin"
									]
								},
								"size": 1
							}
						}
					},
					"terms": {
						"field": "OriginAirportID",
						"size": 10000
					}
				}
			}
		}`,
		[]string{
			`SELECT "OriginAirportID", "DestAirportID", "DestLocation" FROM "(SELECT DestLocation, ROW_NUMBER() OVER (PARTITION BY DestLocation) AS row_number FROM ` + tableName + `)"  GROUP BY ("OriginAirportID", "DestAirportID")`,
			`SELECT "OriginAirportID", "DestAirportID", count() FROM ` + tableNameQuoted + `  GROUP BY ("OriginAirportID", "DestAirportID")`,
			`SELECT "OriginAirportID", "OriginLocation", "Origin" FROM "(SELECT OriginLocation, Origin, ROW_NUMBER() OVER (PARTITION BY OriginLocation, Origin) AS row_number FROM ` + tableName + `)"  GROUP BY ("OriginAirportID")`,
			`SELECT "OriginAirportID", count() FROM ` + tableNameQuoted + `  GROUP BY ("OriginAirportID")`,
		},
	},
	{ // [7]
		`
		{
			"aggs": {
				"0": {
					"aggs": {
						"1": {
							"date_histogram": {
								"calendar_interval": "1d",
								"extended_bounds": {
									"max": 1707818397034,
									"min": 1707213597034
								},
								"field": "order_date",
								"time_zone": "Europe/Warsaw"
							}
						}
					},
					"terms": {
						"field": "category.keyword",
						"order": {
							"_count": "desc"
						},
						"shard_size": 25,
						"size": 10
					}
				}
			}
		}`,
		[]string{
			`SELECT "category.keyword", "order_date", count() FROM ` + tableNameQuoted + `  GROUP BY ("category.keyword", "order_date")`,
			`SELECT "category.keyword", count() FROM ` + tableNameQuoted + `  GROUP BY ("category.keyword")`,
		},
	},
	{ // [8]
		`
		{
			"aggs": {
				"0": {
					"sum": {
						"field": "taxful_total_price"
					}
				}
			}
		}`,
		[]string{`SELECT sum("taxful_total_price") FROM ` + tableNameQuoted + " "},
	},
	{ // [9]
		`
		{
			"aggs": {
				"0": {
					"percentiles": {
						"field": "taxful_total_price",
						"percents": [
							50
						]
					}
				}
			}
		}`,
		[]string{`SELECT quantile("taxful_total_price") FROM ` + tableNameQuoted + " "},
	},
	{ // [10]
		`
		{
			"aggs": {
				"0": {
					"avg": {
						"field": "total_quantity"
					}
				}
			}
		}`,
		[]string{`SELECT avg("total_quantity") FROM ` + tableNameQuoted + " "},
	},
	{ // [11]
		`
		{
			"aggs": {
				"1": {
					"aggs": {
						"2": {
							"aggs": {
								"4": {
									"top_metrics": {
										"metrics": {
											"field": "order_date"
										},
										"size": 10,
										"sort": {
											"order_date": "asc"
										}
									}
								},
								"5": {
									"top_metrics": {
										"metrics": {
											"field": "taxful_total_price"
										},
										"size": 10,
										"sort": {
											"order_date": "asc"
										}
									}
								}
							},
							"date_histogram": {
								"field": "order_date",
								"fixed_interval": "12h",
								"min_doc_count": 1,
								"time_zone": "Europe/Warsaw"
							}
						}
					},
					"filters": {
						"filters": {
							"c8c30be0-b88f-11e8-a451-f37365e9f268": {
								"bool": {
									"filter": [],
									"must": [{
										"query_string": {
											"analyze_wildcard": true,
											"query": "taxful_total_price:>250",
											"time_zone": "Europe/Warsaw"
										}
									}],
									"must_not": [],
									"should": []
								}
							}
						}
					}
				}
			}
		}`,
		[]string{
			`SELECT count() FROM "logs-generic-default" WHERE taxful_total_price>250 `,
			`SELECT "order_date" FROM "(SELECT order_date, ROW_NUMBER() OVER (PARTITION BY order_date) AS row_number FROM ` + tableName + `)" WHERE taxful_total_price>250 `,
			`SELECT "taxful_total_price" FROM "(SELECT taxful_total_price, ROW_NUMBER() OVER (PARTITION BY taxful_total_price) AS row_number FROM ` + tableName + `)" WHERE taxful_total_price>250 `,
		},
	},
}

func TestAggregationParser(t *testing.T) {
	testTable, err := clickhouse.NewTable(`CREATE TABLE `+tableName+`
		( "message" String, "timestamp" DateTime64(3, 'UTC') )
		ENGINE = Memory`,
		clickhouse.NewNoTimestampOnlyStringAttrCHConfig(),
	)
	if err != nil {
		t.Fatal(err)
	}
	lm := clickhouse.NewLogManager(chUrl, clickhouse.TableMap{tableName: testTable}, make(clickhouse.TableMap))
	cw := ClickhouseQueryTranslator{ClickhouseLM: lm, TableName: tableName}

	for testIdx, test := range aggregationTests {
		t.Run(strconv.Itoa(testIdx), func(t *testing.T) {
			aggregations, err := cw.ParseAggregationJson(test.aggregationJson)
			assert.NoError(t, err)
			assert.Equal(t, len(test.translatedSqls), len(aggregations))
			for _, aggregation := range aggregations {
				util.AssertContainsSqlEqual(t, test.translatedSqls, aggregation.String())
			}
		})
	}
}