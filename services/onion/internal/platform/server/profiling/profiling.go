package profiling

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

func PortFromEnv(defaultPort string) string {
	if port := os.Getenv("PPROF_PORT"); port != "" {
		return port
	}
	if os.Getenv("PPROF_ENABLED") == "true" {
		return defaultPort
	}
	return ""
}

func Start(ctx context.Context, port string) error {
	if port == "" {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{
		Addr:              net.JoinHostPort("127.0.0.1", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Ctx(ctx).Info("pprof server starting", logger.String("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Ctx(ctx).Warn("pprof server stopped unexpectedly", logger.Error(err))
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	return nil
}
