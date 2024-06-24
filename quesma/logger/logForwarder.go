package logger

import (
	"os"
	log_sender "quesma/logger/sender"
	"time"
)

type LogForwarder struct {
	logSender log_sender.LogSender
	logCh     chan []byte
	ticker    *time.Ticker
	sigCh     chan os.Signal
	doneCh    chan struct{}
}

func (l *LogForwarder) Run() {
	go func() {
		for {
			select {
			case p := <-l.logCh:
				result := l.logSender.EatLogMessage(p)
				if result.Err != nil {
					logger.Error().Msg(result.Err.Error())
				}
			case <-l.sigCh:
				err := l.logSender.FlushLogs()
				if err != nil {
					logger.Error().Msg(err.Error())
				}
				l.doneCh <- struct{}{}
			}
		}
	}()
}

func (l *LogForwarder) TriggerFlush() {
	go func() {
		for range l.ticker.C {
			l.logCh <- []byte{}
		}
	}()
}

func (l *LogForwarder) Write(p []byte) (n int, err error) {
	cloned := make([]byte, len(p))
	copy(cloned, p)
	l.logCh <- cloned
	return len(cloned), nil
}
