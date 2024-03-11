package main

import (
	"context"
	"mitmproxy/quesma/clickhouse"
	"mitmproxy/quesma/logger"
	"mitmproxy/quesma/quesma"
	"mitmproxy/quesma/quesma/config"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const banner = `
               ________                                       
               \_____  \  __ __   ____   ______ _____ _____   
                /  / \  \|  |  \_/ __ \ /  ___//     \\__  \  
               /   \_/.  \  |  /\  ___/ \___ \|  Y Y  \/ __ \_
               \_____\ \_/____/  \___  >____  >__|_|  (____  /
                      \__>           \/     \/      \/     \/ 
`

func main() {
	println(banner)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	doneCh := make(chan struct{})
	var cfg = config.Load()
	var connectionPool = clickhouse.InitDBConnectionPool(cfg)

	qmcLogChannel := logger.InitLogger(cfg, sig, doneCh)
	defer logger.StdLogFile.Close()
	defer logger.ErrLogFile.Close()

	lm := clickhouse.NewEmptyLogManager(cfg, connectionPool)

	logger.Info().Msgf("loaded config: %s", cfg.String())

	instance := constructQuesma(cfg, lm, qmcLogChannel)
	instance.Start()

	<-doneCh

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	instance.Close(ctx)

}

func constructQuesma(cfg config.QuesmaConfiguration, lm *clickhouse.LogManager, logChan <-chan string) *quesma.Quesma {

	switch cfg.Mode {
	case config.Proxy:
		return quesma.NewQuesmaTcpProxy(cfg, logChan, false)
	case config.ProxyInspect:
		return quesma.NewQuesmaTcpProxy(cfg, logChan, true)
	case config.DualWriteQueryElastic, config.DualWriteQueryClickhouse, config.DualWriteQueryClickhouseVerify, config.DualWriteQueryClickhouseFallback:
		return quesma.NewHttpProxy(lm, cfg, logChan)
	}
	logger.Panic().Msgf("unknown operation mode: %s", cfg.Mode.String())
	panic("unreachable")
}
