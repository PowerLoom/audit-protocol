package health

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/settings"
)

func HealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func HealthCheck(config *settings.Healthcheck) {
	http.Handle(config.Endpoint, HealthCheckHandler())

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
		if err != nil {
			log.WithError(err).Fatal("failed to start health check http server")
		}
	}()

	log.WithField("port", config.Port).Info("started health check http server")
}
