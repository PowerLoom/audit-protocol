package health

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func HealthCheck() {
	http.Handle("/health", HealthCheckHandler())

	go func() {
		err := http.ListenAndServe(":9000", nil)
		if err != nil {
			log.WithError(err).Fatal("failed to start health check http server")
		}
	}()
}
