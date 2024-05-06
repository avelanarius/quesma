package ui

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"mitmproxy/quesma/buildinfo"
	"mitmproxy/quesma/quesma/ui/internal/builder"
	"mitmproxy/quesma/stats/errorstats"
	"net/url"
	"runtime"
	"strings"
	"time"
)

func (qmc *QuesmaManagementConsole) generateDashboard() []byte {
	buffer := newBufferWithHead()
	buffer.Write(generateTopNavigation("dashboard"))

	buffer.Html(`<main id="dashboard-main">` + "\n")

	// Unfortunately, we need tiny bit of javascript to pause the animation.
	buffer.Html(`<script type="text/javascript">`)
	buffer.Html(`var checkbox = document.getElementById("autorefresh");`)
	buffer.Html(`var dashboard = document.getElementById("dashboard-main");`)
	buffer.Html(`checkbox.addEventListener('change', function() {`)
	buffer.Html(`if (this.checked) {`)
	buffer.Html(`dashboard.classList.remove("paused");`)
	buffer.Html(`} else {`)
	buffer.Html(`dashboard.classList.add("paused");`)
	buffer.Html(`}`)
	buffer.Html(`});`)
	buffer.Html(`</script>` + "\n")

	buffer.Html(`<div id="svg-container">`)
	buffer.Html(`<svg width="100%" height="100%" viewBox="0 0 1000 1000" preserveAspectRatio="none">` + "\n")
	// One limitation is that, we don't update color of paths after initial draw.
	// They rarely change, so it's not a big deal for now.
	// Clickhouse -> Kibana
	if qmc.config.ReadsFromClickhouse() {
		status, _ := qmc.generateDashboardTrafficText(RequestStatisticKibana2Clickhouse)
		buffer.Html(fmt.Sprintf(`<path d="M 0 250 L 1000 250" fill="none" stroke="%s" />`, status))
	}
	// Elasticsearch -> Kibana
	if qmc.config.ReadsFromElasticsearch() {
		status, _ := qmc.generateDashboardTrafficText(RequestStatisticKibana2Elasticsearch)
		buffer.Html(fmt.Sprintf(`<path d="M 0 350 L 150 350 L 150 700 L 1000 700" fill="none" stroke="%s" />`, status))
	}

	// Ingest -> Clickhouse
	if qmc.config.WritesToClickhouse() {
		status, _ := qmc.generateDashboardTrafficText(RequestStatisticIngest2Clickhouse)
		buffer.Html(fmt.Sprintf(`<path d="M 1000 350 L 300 350 L 300 650 L 0 650" fill="none" stroke="%s" />`, status))
	}
	// Ingest -> Elasticsearch
	if qmc.config.WritesToElasticsearch() {
		status, _ := qmc.generateDashboardTrafficText(RequestStatisticIngest2Elasticsearch)
		buffer.Html(fmt.Sprintf(`<path d="M 1000 800 L 0 800" fill="none" stroke="%s" />`, status))
	}
	buffer.Html(`</svg>` + "\n")
	buffer.Write(qmc.generateDashboardTrafficPanel())
	buffer.Html(`</div>` + "\n")

	buffer.Html(`<div id="dashboard">` + "\n")
	buffer.Write(qmc.generateDashboardPanel())
	buffer.Html("</div>\n")
	buffer.Html("\n</main>\n\n")

	buffer.Html("\n</body>")
	buffer.Html("\n</html>")
	return buffer.Bytes()
}

func (qmc *QuesmaManagementConsole) generateDashboardTrafficText(typeName string) (string, string) {
	reqStats := qmc.requestsStore.GetRequestsStats(typeName)
	status := "green"
	if reqStats.ErrorRate > 0.20 {
		status = "red"
	}
	return status, fmt.Sprintf("%4.1f req/s, err:%5.1f%%, p99:%3dms",
		reqStats.RatePerMinute/60, reqStats.ErrorRate*100, reqStats.Duration99Percentile)
}

func (qmc *QuesmaManagementConsole) generateDashboardTrafficElement(typeName string, y int) string {
	status, text := qmc.generateDashboardTrafficText(typeName)
	return fmt.Sprintf(
		`<div style="left: 40%%; top: %d%%" id="traffic-%s" hx-swap-oob="true" class="traffic-element %s">%s</div>`,
		y, typeName, status, text)
}

func (qmc *QuesmaManagementConsole) generateDashboardTrafficPanel() []byte {
	var buffer builder.HtmlBuffer

	// Clickhouse -> Kibana
	if qmc.config.ReadsFromClickhouse() {
		buffer.Html(qmc.generateDashboardTrafficElement(RequestStatisticKibana2Clickhouse, 21))
	}

	// Elasticsearch -> Kibana
	if qmc.config.ReadsFromElasticsearch() {
		buffer.Html(qmc.generateDashboardTrafficElement(RequestStatisticKibana2Elasticsearch, 66))
	}

	// Ingest -> Clickhouse
	if qmc.config.WritesToClickhouse() {
		buffer.Html(qmc.generateDashboardTrafficElement(RequestStatisticIngest2Clickhouse, 31))
	}

	// Ingest -> Elasticsearch
	if qmc.config.WritesToElasticsearch() {
		buffer.Html(qmc.generateDashboardTrafficElement(RequestStatisticIngest2Elasticsearch, 76))
	}

	return buffer.Bytes()
}

func secondsToTerseString(second uint64) string {
	return (time.Duration(second) * time.Second).String()
}

func statusToDiv(s healthCheckStatus) string {
	return fmt.Sprintf(`<div class="status %s" title="%s">%s</div>`, s.status, s.tooltip, s.message)
}

func (qmc *QuesmaManagementConsole) generateDashboardPanel() []byte {
	var buffer builder.HtmlBuffer

	dashboardName := "<h3>Kibana</h3>"
	storeName := "<h3>Elasticsearch</h3>"
	if qmc.config.Elasticsearch.Url != nil && strings.Contains(qmc.config.Elasticsearch.Url.String(), "opensearch") {
		dashboardName = "<h3>OpenSearch</h3><h3>Dashboards</h3>"
		storeName = "<h3>OpenSearch</h3>"
	}

	clickhouseName := "<h3>ClickHouse</h3>"
	if qmc.config.Hydrolix.Url != nil {
		clickhouseName = "<h3>Hydrolix</h3>"
	}

	buffer.Html(`<div id="dashboard-kibana" class="component">`)
	if qmc.config.Elasticsearch.AdminUrl != nil {
		buffer.Html(fmt.Sprintf(`<a href="%s">`, qmc.config.Elasticsearch.AdminUrl.String()))
	}
	buffer.Html(dashboardName)
	if qmc.config.Elasticsearch.AdminUrl != nil {
		buffer.Html(`</a>`)
	}
	buffer.Html(statusToDiv(qmc.checkKibana()))
	buffer.Html(`</div>`)

	buffer.Html(`<div id="dashboard-ingest" class="component">`)
	buffer.Html(`<h3>Ingest</h3>`)
	buffer.Html(statusToDiv(qmc.checkIngest()))
	buffer.Html(`</div>`)

	buffer.Html(`<div id="dashboard-elasticsearch" class="component">`)
	buffer.Html(storeName)
	buffer.Html(statusToDiv(qmc.checkElasticsearch()))
	buffer.Html(`</div>`)

	buffer.Html(`<div id="dashboard-clickhouse" class="component">`)
	if qmc.config.ClickHouse.AdminUrl != nil {
		buffer.Html(fmt.Sprintf(`<a href="%s">`, qmc.config.ClickHouse.AdminUrl.String()))
	}
	buffer.Html(clickhouseName)
	if qmc.config.ClickHouse.AdminUrl != nil {
		buffer.Html(`</a>`)
	}
	buffer.Html(statusToDiv(qmc.checkClickhouseHealth()))
	buffer.Html(`</div>`)

	buffer.Html(`<div id="dashboard-traffic" class="component">`)

	buffer.Html(`<div id="dashboard-quesma" class="component">`)
	buffer.Html(`<h3>Quesma</h3>`)

	cpuStr := ""
	c0, err0 := cpu.Percent(0, false)

	if err0 == nil {
		cpuStr = fmt.Sprintf("Host CPU: %.1f%%", c0[0])
	} else {
		cpuStr = fmt.Sprintf("Host CPU: N/A (error: %s)", err0.Error())
	}

	buffer.Html(fmt.Sprintf(`<div class="status">%s</div>`, cpuStr))

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memStr := fmt.Sprintf("Memory used: %1.f MB", float64(m.Alloc)/1024.0/1024.0)
	if v, errV := mem.VirtualMemory(); errV == nil {
		total := float64(v.Total) / 1024.0 / 1024.0 / 1024.0
		memStr += fmt.Sprintf(", avail: %.1f GB", total)
	}
	buffer.Html(fmt.Sprintf(`<div class="status">%s</div>`, memStr))

	duration := uint64(time.Since(qmc.startedAt).Seconds())

	buffer.Html(fmt.Sprintf(`<div class="status">Started: %s ago</div>`, secondsToTerseString(duration)))
	buffer.Html(fmt.Sprintf(`<div class="status">Mode: %s</div>`, qmc.config.Mode.String()))

	if h, errH := host.Info(); errH == nil {
		buffer.Html(fmt.Sprintf(`<div class="status">Host uptime: %s</div>`, secondsToTerseString(h.Uptime)))
	}

	buffer.Html("<div>Version: ")
	buffer.Text(buildinfo.Version)
	buffer.Html("</div>")

	buffer.Html(`</div>`)

	buffer.Html(`<div id="dashboard-errors" class="component">`)
	errors := errorstats.GlobalErrorStatistics.ReturnTopErrors(5)
	if len(errors) > 0 {
		buffer.Html(`<h3>Top errors:</h3>`)
		for _, e := range errors {
			buffer.Html(fmt.Sprintf(`<div class="status">%d: <a href="/error/%s">%s</a></div>`,
				e.Count, url.PathEscape(e.Reason), e.Reason))
		}
	} else {
		buffer.Html(`<h3>No errors</h3>`)
	}
	buffer.Html(`</div>`)
	buffer.Html(`</div>`)

	return buffer.Bytes()
}