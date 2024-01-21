package quesma

import (
	"errors"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
)

const UI_TCP_PORT = "9999"

type QueryDebugPrimarySource struct {
	id        string
	queryResp []byte
}

type QueryDebugSecondarySource struct {
	id string

	incomingQueryBody []byte

	queryBodyTranslated    []byte
	queryRawResults        []byte
	queryTranslatedResults []byte
}

type QueryDebugInfo struct {
	QueryDebugPrimarySource
	QueryDebugSecondarySource
}

type QueryDebugger struct {
	queryDebugPrimarySource   chan *QueryDebugPrimarySource
	queryDebugSecondarySource chan *QueryDebugSecondarySource
	ui                        *http.Server
	mutex                     sync.Mutex
	debugInfoMessages         map[string]QueryDebugInfo
}

func NewQueryDebugger() *QueryDebugger {
	return &QueryDebugger{
		queryDebugPrimarySource:   make(chan *QueryDebugPrimarySource, 5),
		queryDebugSecondarySource: make(chan *QueryDebugSecondarySource, 5),
		debugInfoMessages:         make(map[string]QueryDebugInfo),
	}
}

func (qd *QueryDebugger) PushPrimaryInfo(qdebugInfo *QueryDebugPrimarySource) {
	qd.queryDebugPrimarySource <- qdebugInfo
}

func (qd *QueryDebugger) PushSecondaryInfo(qdebugInfo *QueryDebugSecondarySource) {
	qd.queryDebugSecondarySource <- qdebugInfo
}

func copyMap(originalMap map[string]QueryDebugInfo) map[string]QueryDebugInfo {
	copiedMap := make(map[string]QueryDebugInfo)

	for key, value := range originalMap {
		copiedMap[key] = value
	}

	return copiedMap
}

func (qd *QueryDebugger) newHTTPServer() *http.Server {
	return &http.Server{
		Addr: ":" + UI_TCP_PORT,
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			buf := qd.generateReport()
			writer.Write(buf)
		})}
}

func (qd *QueryDebugger) listenAndServe() {
	if err := qd.ui.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Error starting server:", err)
	}
}

type DebugKeyValue struct {
	Key   string
	Value QueryDebugInfo
}

func (qd *QueryDebugger) generateReport() []byte {
	qd.mutex.Lock()
	localQueryDebugInfo := copyMap(qd.debugInfoMessages)
	qd.mutex.Unlock()

	var debugKeyValueSlice []DebugKeyValue
	for key, value := range localQueryDebugInfo {
		debugKeyValueSlice = append(debugKeyValueSlice, DebugKeyValue{key, value})
	}

	sort.Slice(debugKeyValueSlice, func(i, j int) bool {
		a, _ := strconv.Atoi(debugKeyValueSlice[i].Key)
		b, _ := strconv.Atoi(debugKeyValueSlice[j].Key)
		return a < b
	})

	var buf []byte
	head, err := os.ReadFile("/var/quesma/quesma/ui/head.html")
	buf = append(buf, head...)
	if err != nil {
		buf = append(buf, []byte(err.Error())...)
	}
	buf = append(buf, []byte("\n<div class=\"topnav\">")...)
	buf = append(buf, []byte("\n<h3>Quesma Debugging Interface</h3>")...)
	buf = append(buf, []byte("\n</div>")...)
	buf = append(buf, []byte("\n<div class=\"left\" id=\"left\">")...)
	buf = append(buf, []byte("\n<div class=\"title-bar\">Query")...)
	buf = append(buf, []byte("\n</div>")...)
	for _, v := range debugKeyValueSlice {
		buf = append(buf, []byte("<p>RequestID:"+v.Key+"</p>")...)
		buf = append(buf, []byte("\n<pre id=\"query"+v.Key+"\">")...)
		buf = append(buf, []byte(v.Value.incomingQueryBody)...)
		buf = append(buf, []byte("\n</pre>")...)
	}
	buf = append(buf, []byte("\n</div>")...)
	buf = append(buf, []byte("\n<div class=\"right\" id=\"right\">")...)
	buf = append(buf, []byte("\n<div class=\"title-bar\">Elasticsearch response")...)
	buf = append(buf, []byte("\n</div>")...)
	for _, v := range debugKeyValueSlice {
		buf = append(buf, []byte("<p>ResponseID:"+v.Key+"</p>")...)
		buf = append(buf, []byte("\n<pre id=\"response"+v.Key+"\">")...)
		buf = append(buf, []byte(v.Value.queryResp)...)
		buf = append(buf, []byte("\n</pre>")...)
	}
	buf = append(buf, []byte("\n</div>")...)
	buf = append(buf, []byte("\n<div class=\"bottom_left\" id=\"bottom_left\">")...)
	buf = append(buf, []byte("\n<div class=\"title-bar\">Clickhouse translated query")...)
	buf = append(buf, []byte("\n</div>")...)
	for _, v := range debugKeyValueSlice {
		buf = append(buf, []byte("<p>RequestID:"+v.Key+"</p>")...)
		buf = append(buf, []byte("\n<pre id=\"second_query"+v.Key+"\">")...)
		buf = append(buf, []byte(v.Value.queryBodyTranslated)...)
		buf = append(buf, []byte("\n</pre>")...)
	}
	buf = append(buf, []byte("\n</div>")...)
	buf = append(buf, []byte("\n<div class=\"bottom_right\" id=\"bottom_right\">")...)
	buf = append(buf, []byte("\n<div class=\"title-bar\">Clickhouse response")...)
	buf = append(buf, []byte("\n</div>")...)
	for _, v := range debugKeyValueSlice {
		buf = append(buf, []byte("<p>ResponseID:"+v.Key+"</p>")...)
		buf = append(buf, []byte("\n<pre id=\"second_response"+v.Key+"\">")...)
		buf = append(buf, []byte(v.Value.queryTranslatedResults)...)
		buf = append(buf, []byte("\n</pre>")...)
	}
	buf = append(buf, []byte("\n</div>")...)
	buf = append(buf, []byte("\n<div class=\"menu\">")...)
	buf = append(buf, []byte("\n<h2>Menu</h2>")...)
	buf = append(buf, []byte("&nbsp;<button id=\"find_query_by_id_button\" type=\"button\" class=\"btn\" onclick=\"findquerybyid_clicked(find_query_by_id_input.value)\">Find query by id</button><br>")...)
	buf = append(buf, []byte("&nbsp;<input type=\"text\" id=\"find_query_by_id_input\" class=\"input\" name=\"find_query_by_id_input\" value=\"\" required size=\"40\"><br><br>")...)

	buf = append(buf, []byte("&nbsp;<button id=\"find_query_by_str_button\" type=\"button\" class=\"btn\" onclick=\"findquerybystr_clicked(find_query_by_str_input.value)\">Find query by string</button><br>")...)
	buf = append(buf, []byte("&nbsp;<input type=\"text\" id=\"find_query_by_str_input\" class=\"input\" name=\"find_query_by_str_input\" value=\"\" required size=\"40\"><br><br>")...)
	buf = append(buf, []byte("&nbsp;<button id=\"unselect_button\" type=\"button\" class=\"btn\" onclick=\"unselect_clicked()\">Unselect</button><br>")...)

	buf = append(buf, []byte("\n</div>")...)
	buf = append(buf, []byte("\n</body>")...)
	buf = append(buf, []byte("\n</html>")...)
	return buf
}

func (qd *QueryDebugger) Run() {
	go func() {
		qd.ui = qd.newHTTPServer()
		qd.listenAndServe()
	}()
	for {
		select {
		case msg := <-qd.queryDebugPrimarySource:
			log.Println("Received debug info from primary source:", msg.id)
			qd.mutex.Lock()
			if value, ok := qd.debugInfoMessages[msg.id]; !ok {
				qd.debugInfoMessages[msg.id] = QueryDebugInfo{
					QueryDebugPrimarySource: QueryDebugPrimarySource{msg.id, msg.queryResp},
				}
			} else {
				value.QueryDebugPrimarySource = QueryDebugPrimarySource{msg.id, msg.queryResp}
				qd.debugInfoMessages[msg.id] = value
			}
			qd.mutex.Unlock()
		case msg := <-qd.queryDebugSecondarySource:
			log.Println("Received debug info from secondary source:", msg.id)
			qd.mutex.Lock()
			if value, ok := qd.debugInfoMessages[msg.id]; !ok {
				qd.debugInfoMessages[msg.id] = QueryDebugInfo{
					QueryDebugSecondarySource: QueryDebugSecondarySource{
						msg.id,
						msg.incomingQueryBody,
						msg.queryBodyTranslated,
						msg.queryRawResults,
						msg.queryTranslatedResults,
					},
				}
			} else {
				value.QueryDebugSecondarySource = QueryDebugSecondarySource{
					msg.id,
					msg.incomingQueryBody,
					msg.queryBodyTranslated,
					msg.queryRawResults,
					msg.queryTranslatedResults,
				}
				qd.debugInfoMessages[msg.id] = value
			}
			qd.mutex.Unlock()

		}
	}
}
