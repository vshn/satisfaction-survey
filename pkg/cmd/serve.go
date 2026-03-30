package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/stdr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/vshn/satisfaction-survey/pkg/api"
	"github.com/vshn/satisfaction-survey/pkg/metrics"
)

type prometheusConfig struct {
	URL     string
	Headers map[string]string
}

var (
	serverCommandName = "serve"
	serverConfig      = api.ApiServerConfig{}
	promConfig        = prometheusConfig{Headers: map[string]string{}}
	servePath         string
	serveCmd          = &cobra.Command{
		Use:   serverCommandName,
		Short: "Serve API endpoints",
		Long:  "Serve API endpoints",
		Run: func(cmd *cobra.Command, args []string) {
			l := stdr.New(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))
			serverConfig.Logger = &l

			var reg = prometheus.NewRegistry()

			var counter = metrics.NewCounter(reg)

			var server = api.NewApiServer(serverConfig, counter, reg, servePath)
			log.Println("Starting API server ...")

			go func() {
				err := server.Start()
				if err != nil {
					log.Fatal(err)
				}
				log.Println("Stopped serving new connections.")
			}()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownRelease()

			if err := server.Stop(shutdownCtx); err != nil {
				log.Fatalf("HTTP shutdown error: %v", err)
			}
			log.Println("Graceful shutdown complete.")

		},
	}
)

type headerInjector struct {
	headers map[string]string
}

func (h headerInjector) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	for key, value := range h.headers {
		r2.Header.Set(key, value)
	}
	return http.DefaultTransport.RoundTrip(r2)
}

func init() {
	serveCmd.Flags().StringVar(&serverConfig.AuthUser, "auth-user", "admin", "Username for authenticating with the API")
	serveCmd.Flags().StringVar(&serverConfig.AuthPass, "auth-pass", "", "Password for authenticating with the API")
	serveCmd.Flags().IntVar(&serverConfig.Port, "port", 8080, "Port at which to serve API")
	serveCmd.Flags().StringVar(&serverConfig.Host, "host", "0.0.0.0", "Host address to bind")
	serveCmd.Flags().StringVar(&servePath, "path", "./static", "Path from which to serve static files")

	rootCmd.AddCommand(serveCmd)
}
