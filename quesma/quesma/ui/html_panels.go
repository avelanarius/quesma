package ui

import (
	"bytes"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"mitmproxy/quesma/quesma/mux"
	"mitmproxy/quesma/stats"
	"mitmproxy/quesma/util"
	"strings"
	"time"
)

func generateQueries(debugKeyValueSlice []DebugKeyValue, withLinks bool) []byte {
	var buffer bytes.Buffer

	buffer.WriteString("\n" + `<div class="left" Id="left">` + "\n")
	buffer.WriteString(`<div class="title-bar">Query`)
	buffer.WriteString("\n</div>\n")
	buffer.WriteString(`<div class="debug-body">`)
	for _, v := range debugKeyValueSlice {
		if withLinks {
			buffer.WriteString(`<a href="/request-Id/` + v.Key + `">`)
		}
		buffer.WriteString("<p>RequestID:" + v.Key + "</p>\n")
		buffer.WriteString(`<pre Id="query` + v.Key + `">`)
		buffer.WriteString(util.JsonPrettify(string(v.Value.IncomingQueryBody), true))
		buffer.WriteString("\n</pre>")
		if withLinks {
			buffer.WriteString("\n</a>")
		}
	}
	buffer.WriteString("\n</div>")
	buffer.WriteString("\n</div>\n")

	buffer.WriteString(`<div class="right" Id="right">` + "\n")
	buffer.WriteString(`<div class="title-bar">Elasticsearch response` + "\n" + `</div>`)
	buffer.WriteString(`<div class="debug-body">`)
	for _, v := range debugKeyValueSlice {
		if withLinks {
			buffer.WriteString(`<a href="/request-Id/` + v.Key + `">`)
		}
		buffer.WriteString("<p>ResponseID:" + v.Key + "</p>\n")
		buffer.WriteString(`<pre Id="response` + v.Key + `">`)
		buffer.WriteString(util.JsonPrettify(string(v.Value.QueryResp), true))
		buffer.WriteString("\n</pre>")
		if withLinks {
			buffer.WriteString("\n</a>")
		}
	}
	buffer.WriteString("\n</div>")
	buffer.WriteString("\n</div>\n")

	buffer.WriteString(`<div class="bottom_left" Id="bottom_left">` + "\n")
	buffer.WriteString(`<div class="title-bar">Clickhouse translated query` + "\n" + `</div>`)
	buffer.WriteString(`<div class="debug-body">`)
	for _, v := range debugKeyValueSlice {
		if withLinks {
			buffer.WriteString(`<a href="/request-Id/` + v.Key + `">`)
		}
		buffer.WriteString("<p>RequestID:" + v.Key + "</p>\n")
		buffer.WriteString(`<pre Id="second_query` + v.Key + `">`)
		buffer.WriteString(sqlPrettyPrint(v.Value.QueryBodyTranslated))
		buffer.WriteString("\n</pre>")
		if withLinks {
			buffer.WriteString("\n</a>")
		}
	}
	buffer.WriteString("\n</div>")
	buffer.WriteString("\n</div>\n")

	buffer.WriteString(`<div class="bottom_right" Id="bottom_right">` + "\n")
	buffer.WriteString(`<div class="title-bar">Clickhouse response` + "\n" + `</div>`)
	buffer.WriteString(`<div class="debug-body">`)
	for _, v := range debugKeyValueSlice {
		if withLinks {
			buffer.WriteString(`<a href="/request-Id/` + v.Key + `">`)
		}
		buffer.WriteString("<p>ResponseID:" + v.Key + "</p>\n")
		buffer.WriteString(`<pre Id="second_response` + v.Key + `">`)
		buffer.WriteString(util.JsonPrettify(string(v.Value.QueryTranslatedResults), true))
		buffer.WriteString("\n\nThere are more results ...")
		buffer.WriteString("\n</pre>")
		if withLinks {
			buffer.WriteString("\n</a>")
		}
	}
	buffer.WriteString("\n</div>")
	buffer.WriteString("\n</div>\n")

	return buffer.Bytes()
}

func (qmc *QuesmaManagementConsole) generateRouterStatistics() []byte {
	var buffer bytes.Buffer

	matchStatistics := mux.MatchStatistics()

	buffer.WriteString("\n<h2>Matched URLs</h2>\n<ul>")
	for _, url := range matchStatistics.Matched {
		buffer.WriteString(fmt.Sprintf("<li>%s</li>\n", url))
	}

	buffer.WriteString("</ul>\n")
	buffer.WriteString("\n<h2>Not matched URLs</h2>\n<ul>")
	for _, url := range matchStatistics.Nonmatched {
		buffer.WriteString(fmt.Sprintf("<li>%s</li>\n", url))
	}
	buffer.WriteString("</ul>\n")

	return buffer.Bytes()
}

func (qmc *QuesmaManagementConsole) generateStatistics() []byte {
	var buffer bytes.Buffer
	const maxTopValues = 5

	statistics := stats.GlobalStatistics

	for _, index := range statistics.SortedIndexNames() {
		buffer.WriteString("\n" + fmt.Sprintf(`<h2>Stats for "%s" <small>from %d requests</small></h2>`, index.IndexName, index.Requests) + "\n")

		buffer.WriteString("<table>\n")

		buffer.WriteString("<thead>\n")
		buffer.WriteString("<tr>\n")
		buffer.WriteString(`<th class="key">Key</th>` + "\n")
		buffer.WriteString(`<th class="key-count">Count</th>` + "\n")
		buffer.WriteString(`<th class="value">Value</th>` + "\n")
		buffer.WriteString(`<th class="value-count">Count</th>` + "\n")
		buffer.WriteString(`<th class="value-count">Percentage</th>` + "\n")
		buffer.WriteString(`<th class="types">Potential type</th>` + "\n")
		buffer.WriteString("</tr>\n")
		buffer.WriteString("</thead>\n")
		buffer.WriteString("<tbody>\n")

		for _, keyStats := range index.SortedKeyStatistics() {
			topValuesCount := maxTopValues
			if len(keyStats.Values) < maxTopValues {
				topValuesCount = len(keyStats.Values)
			}

			buffer.WriteString("<tr>\n")
			buffer.WriteString(fmt.Sprintf(`<td class="key" rowspan="%d">%s</td>`+"\n", topValuesCount, keyStats.KeyName))
			buffer.WriteString(fmt.Sprintf(`<td class="key-count" rowspan="%d">%d</td>`+"\n", topValuesCount, keyStats.Occurrences))

			for i, value := range keyStats.TopNValues(topValuesCount) {
				if i > 0 {
					buffer.WriteString("</tr>\n<tr>\n")
				}

				buffer.WriteString(fmt.Sprintf(`<td class="value">%s</td>`, value.ValueName))
				buffer.WriteString(fmt.Sprintf(`<td class="value-count">%d</td>`, value.Occurrences))
				buffer.WriteString(fmt.Sprintf(`<td class="value-count">%.1f%%</td>`, 100*float32(value.Occurrences)/float32(keyStats.Occurrences)))
				buffer.WriteString(fmt.Sprintf(`<td class="types">%s</td>`, strings.Join(value.Types, ", ")))
			}
			buffer.WriteString("</tr>\n")
		}

		buffer.WriteString("</tbody>\n")

		buffer.WriteString("</table>\n")
	}

	return buffer.Bytes()
}

func secondsToTerseString(second uint64) string {
	return (time.Duration(second) * time.Second).String()
}

func (qmc *QuesmaManagementConsole) generateDashboardPanel() []byte {
	var buffer bytes.Buffer

	buffer.WriteString(`<div id="dashboard-kibana" class="component">`)
	buffer.WriteString(`<h3>Kibana</h3>`)
	buffer.WriteString(`<div class="status">OK</div>`)
	buffer.WriteString(`</div>`)

	buffer.WriteString(`<div id="dashboard-ingest" class="component">`)
	buffer.WriteString(`<h3>Ingest</h3>`)
	buffer.WriteString(`<div class="status">OK</div>`)
	buffer.WriteString(`</div>`)

	buffer.WriteString(`<div id="dashboard-elasticsearch" class="component">`)
	buffer.WriteString(`<h3>Elastic</h3><h3>search</h3>`)
	buffer.WriteString(`<div class="status">OK</div>`)
	buffer.WriteString(`</div>`)

	buffer.WriteString(`<div id="dashboard-clickhouse" class="component">`)
	buffer.WriteString(`<h3>ClickHouse</h3>`)
	buffer.WriteString(`<div class="status">OK</div>`)
	buffer.WriteString(`</div>`)

	buffer.WriteString(`<div id="dashboard-traffic" class="component">`)

	buffer.WriteString(`<div id="dashboard-quesma" class="component">`)
	buffer.WriteString(`<h3>Quesma</h3>`)

	cpuStr := ""
	c0, err0 := cpu.Percent(0, false)

	if err0 == nil {
		cpuStr = fmt.Sprintf("CPU: %.1f%%", c0[0])
	} else {
		cpuStr = fmt.Sprintf("CPU: N/A (error: %s)", err0.Error())
	}

	buffer.WriteString(fmt.Sprintf(`<div class="status">%s</div>`, cpuStr))

	memStr := ""
	v, errV := mem.VirtualMemory()
	if errV == nil {
		total := float64(v.Total) / 1024.0 / 1024.0 / 1024.0
		used := float64(v.Used) / 1024.0 / 1024.0 / 1024.0
		memStr = fmt.Sprintf("Memory - used: %1.f GB, total: %.1f GB", used, total)
	} else {
		memStr = fmt.Sprintf("Memory: N/A (error: %s)", errV.Error())
	}
	buffer.WriteString(fmt.Sprintf(`<div class="status">%s</div>`, memStr))

	// TODO: Currently we check host uptime, not application uptime
	uptimeStr := ""
	h, errH := host.Info()
	if errH == nil {
		uptimeStr = fmt.Sprintf("Uptime: %s", secondsToTerseString(h.Uptime))
	} else {
		uptimeStr = fmt.Sprintf("Uptime: N/A (error: %s)", errH.Error())
	}
	buffer.WriteString(fmt.Sprintf(`<div class="status">%s</div>`, uptimeStr))
	buffer.WriteString(fmt.Sprintf(`<div class="status">Mode: %s</div>`, qmc.config.Mode.String()))
	buffer.WriteString(`</div>`)

	buffer.WriteString(`<div id="dashboard-errors" class="component">`)
	buffer.WriteString(`<h3>Top errors:</h3>`)
	buffer.WriteString(`<div class="status">7: Unknown error</div>`)
	buffer.WriteString(`<div class="status">2: Parsing error</div>`)
	buffer.WriteString(`<div class="status">1: Request out of bound</div>`)
	buffer.WriteString(`</div>`)
	buffer.WriteString(`</div>`)

	return buffer.Bytes()
}

func (qmc *QuesmaManagementConsole) generateDashboardTrafficPanel() []byte {
	var buffer bytes.Buffer
	var rps, err, latency int32

	buffer.WriteString(`<svg width="100%" height="100%" viewBox="0 0 1000 1000">`)

	// Clickhouse -> Kibana
	if qmc.config.ReadsFromClickhouse() {
		rps = rand.Int31n(100) + 5
		err = rand.Int31n(5)
		latency = rand.Int31n(20) + 10
		buffer.WriteString(fmt.Sprintf(`<text x="500" y="240" class="green">%d rps / %d%% err / %dms</text>`,
			rps, err, latency))
	}

	// Elasticsearch -> Kibana
	if qmc.config.ReadsFromElasticsearch() {
		rps = rand.Int31n(100) + 5
		err = 0
		latency = rand.Int31n(20) + 10
		buffer.WriteString(fmt.Sprintf(`<text x="500" y="690" class="green">%d rps / %d%% err / %dms</text>`,
			rps, err, latency))
	}

	// Ingest -> Clickhouse
	if qmc.config.WritesToClickhouse() {
		buffer.WriteString(`<text x="500" y="340" class="green">29 rps / 1% err / 3ms</text>`)
	}

	// Ingest -> Elasticsearch
	if qmc.config.WritesToElasticsearch() {
		buffer.WriteString(`<text x="500" y="790" class="green">29 rps / 0% err / 1ms</text>`)
	}
	buffer.WriteString(`</svg>`)

	return buffer.Bytes()
}