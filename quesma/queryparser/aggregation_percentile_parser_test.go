package queryparser

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_parsePercentilesAggregationWithDefaultPercents(t *testing.T) {
	payload := QueryMap{
		"field": "custom_name",
	}
	fieldName, userSpecifiedPercents := parsePercentilesAggregation(payload)
	assert.Equal(t, "custom_name", fieldName)
	assert.Equal(t, DefaultPercentiles, userSpecifiedPercents)
}

func Test_parsePercentilesAggregationWithUserSpecifiedPercents(t *testing.T) {

	payload := QueryMap{
		"field":    "custom_name",
		"percents": []interface{}{0.001, 0.01, 0.05, 11.123123123123124, 63.4, 66.999999999999, float64(95), float64(99), 99.9, 99.9999, 99.99999999},
	}
	expectedOutputMap := map[string]float64{
		"0.001":              0.00001,
		"0.01":               0.0001,
		"0.05":               0.0005,
		"11.123123123123124": 0.1112312312,
		"63.4":               0.634,
		"66.999999999999":    0.6699999999,
		"95":                 0.95,
		"99":                 0.99,
		"99.9":               0.999,
		"99.9999":            0.999999,
		"99.99999999":        0.9999999999,
	}
	expectedOutputMapKeys := make([]string, 0, len(expectedOutputMap))
	for k := range expectedOutputMap {
		expectedOutputMapKeys = append(expectedOutputMapKeys, k)
	}

	fieldName, parsedMap := parsePercentilesAggregation(payload)
	assert.Equal(t, "custom_name", fieldName)

	parsedMapKeys := make([]string, 0, len(parsedMap))
	for k := range parsedMap {
		parsedMapKeys = append(parsedMapKeys, k)
	}
	assert.ElementsMatch(t, expectedOutputMapKeys, parsedMapKeys)

	assert.Equal(t, 0.00001, parsedMap["0.001"])
	assert.Equal(t, 0.0001, parsedMap["0.01"])
	assert.Equal(t, 0.0005, parsedMap["0.05"])
	assert.True(t, isBetween(parsedMap["11.123123123123124"], 0.111231231, 0.111231232))
	assert.Equal(t, 0.634, parsedMap["63.4"])
	assert.True(t, isBetween(parsedMap["66.999999999999"], 0.66999999, 0.67))
	assert.Equal(t, 0.95, parsedMap["95"])
	assert.Equal(t, 0.99, parsedMap["99"])
	assert.True(t, isBetween(parsedMap["99.9"], 0.999, 0.9991))
	assert.Equal(t, 0.999999, parsedMap["99.9999"])
	assert.Equal(t, maxPrecision, parsedMap["99.99999999"])

}

// For some numbers we might hit precision issues, so we need to check if the value is between the limits
func isBetween(value, lowerLimit, upperLimit float64) bool {
	if value < lowerLimit {
		fmt.Printf("value %f is lower than lower limit %f", value, lowerLimit)
		return false
	} else if value > upperLimit {
		fmt.Printf("value %f is higher than upper limit %f", value, upperLimit)
		return false
	}
	return true
}