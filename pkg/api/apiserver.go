package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vshn/satisfaction-survey/pkg/api/client"
	"github.com/vshn/satisfaction-survey/pkg/api/metrics"
	"github.com/vshn/satisfaction-survey/pkg/api/satisfaction"
)

type ApiServerConfig struct {
	AuthUser string
	AuthPass string
	AuthDisable bool
	Port     int
	Host     string

	Logger *logr.Logger
}

type ApiServer struct {
	config  ApiServerConfig
	mux     *http.ServeMux
	counter satisfaction.ResultCounter
	server  *http.Server
}

func NewApiServer(config ApiServerConfig, counter satisfaction.ResultCounter, reg prometheus.Gatherer, servedir string) ApiServer {
	if config.Logger == nil {
		l := logr.Discard()
		config.Logger = &l
	}
	var mux = http.NewServeMux()
	satisfaction.Setup(mux, counter)
	metrics.Setup(mux, reg)
	client.Setup(mux, servedir)
	return ApiServer{
		config: config,
		mux:    mux,
	}
}

func (s *ApiServer) Start() error {
	var hostport = fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	var server = http.Server{
		Addr:    hostport,
		Handler: s.logInject(s.basicAuth(s.mux)),
	}
	s.server = &server
	s.config.Logger.Info("Listening on", "addr", hostport)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *ApiServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *ApiServer) logInject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.config.Logger.WithValues(
			"method", r.Method, "url", r.URL.String(), "remote", r.RemoteAddr,
			"request_id", r.Header.Get("X-Request-ID"), "internal_request_id", uuid.NewString(),
			"user_agent", r.UserAgent(),
		)
		ctx := logr.NewContext(r.Context(), logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *ApiServer) basicAuth(next http.Handler) http.Handler {
	if s.config.AuthDisable {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(s.config.AuthUser))
			expectedPasswordHash := sha256.Sum256([]byte(s.config.AuthPass))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logr.FromContextOrDiscard(r.Context()).Info("Unauthorized request", "username", username)
	})
}
