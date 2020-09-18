package hystrix

import (
	"errors"
	"time"

	"github.com/leon-yc/ggs/pkg/qlog"
)

//Reporter receive a circuit breaker Metrics and sink it to monitoring system
type Reporter func(cb *CircuitBreaker) error

//ErrDuplicated means you can not install reporter with same name
var ErrDuplicated = errors.New("duplicated reporter")
var reporterPlugins = make(map[string]Reporter)

//InstallReporter install reporter implementation
//it receives a circuit breaker and sink its Metrics to monitoring system
func InstallReporter(name string, reporter Reporter) error {
	_, ok := reporterPlugins[name]
	if ok {
		return ErrDuplicated
	}
	reporterPlugins[name] = reporter
	qlog.Trace("Install reporter plugin:" + name)
	return nil
}

//StartReporter starts reporting to reporters
func StartReporter() {
	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-tick:
			circuitBreakersMutex.RLock()
			for _, cb := range circuitBreakers {
				for k, report := range reporterPlugins {
					qlog.Trace("report circuit metrics to " + k)
					if err := report(cb); err != nil {
						qlog.Warn("can not report: " + err.Error())
					}
				}
			}
			circuitBreakersMutex.RUnlock()
		}
	}
}
