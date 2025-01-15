// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0
package main

import (
	"os"
	"quesma/ab_testing"
	"quesma/backend_connectors"
	"quesma/clickhouse"
	"quesma/elasticsearch"
	"quesma/frontend_connectors"
	"quesma/ingest"
	"quesma/logger"
	"quesma/processors"
	"quesma/quesma"
	"quesma/quesma/config"
	"quesma/quesma/ui"
	"quesma/schema"
	"quesma/table_resolver"
	"quesma/telemetry"
	quesma_api "quesma_v2/core"
)

const banner = `
               ________                                       
               \_____  \  __ __   ____   ______ _____ _____   
                /  / \  \|  |  \_/ __ \ /  ___//     \\__  \  
               /   \_/.  \  |  /\  ___/ \___ \|  Y Y  \/ __ \_
               \_____\ \_/____/  \___  >____  >__|_|  (____  /
                      \__>           \/     \/      \/     \/ 
`

const EnableConcurrencyProfiling = false

// Example of how to use the v2 module api in main function
//func main() {
//	q1 := buildQueryOnlyQuesma()
//	q1.Start()
//	stop := make(chan os.Signal, 1)
//	<-stop
//	q1.Stop(context.Background())
//}

func main() {
	var frontendConn = frontend_connectors.NewTCPConnector(":5432")
	var tcpProcessor quesma_api.Processor = processors.NewPostgresToDuckDBProcessor()
	var tcpPostgressHandler = frontend_connectors.TcpPostgresConnectionHandler{}
	frontendConn.AddConnectionHandler(&tcpPostgressHandler)
	var postgressPipeline quesma_api.PipelineBuilder = quesma_api.NewPipeline()
	postgressPipeline.AddProcessor(tcpProcessor)
	postgressPipeline.AddFrontendConnector(frontendConn)
	var quesmaBuilder quesma_api.QuesmaBuilder = quesma_api.NewQuesma(quesma_api.EmptyDependencies())
	postgressPipeline.AddBackendConnector(&backend_connectors.MySqlBackendConnector{})
	quesmaBuilder.AddPipeline(postgressPipeline)
	qb, _ := quesmaBuilder.Build()
	qb.Start()

	stop := make(chan os.Signal, 1)
	<-stop
}

func constructQuesma(cfg *config.QuesmaConfiguration, sl clickhouse.TableDiscovery, lm *clickhouse.LogManager, ip *ingest.IngestProcessor, im elasticsearch.IndexManagement, schemaRegistry schema.Registry, phoneHomeAgent telemetry.PhoneHomeAgent, quesmaManagementConsole *ui.QuesmaManagementConsole, logChan <-chan logger.LogWithLevel, abResultsrepository ab_testing.Sender, indexRegistry table_resolver.TableResolver) *quesma.Quesma {
	if cfg.TransparentProxy {
		return quesma.NewQuesmaTcpProxy(cfg, quesmaManagementConsole, logChan, false)
	} else {
		const quesma_v2 = false
		return quesma.NewHttpProxy(phoneHomeAgent, lm, ip, sl, im, schemaRegistry, cfg, quesmaManagementConsole, abResultsrepository, indexRegistry, quesma_v2)
	}
}
