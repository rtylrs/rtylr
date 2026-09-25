package probe

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

type Check func(context.Context) error

type Config struct {
	ServiceName string
	Port        string
	StopToken   string
	ReadyChecks []Check
	Stop        func()
}

func PortFromEnv(defaultPort string) string {
	if port := os.Getenv("PROBE_PORT"); port != "" {
		return port
	}
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return defaultPort
	}
	return ""
}

// Start launches a small probe server. Empty Port disables it, which keeps
// local multi-process dev from fighting over :8888.
func Start(ctx context.Context, cfg Config) error {
	if cfg.Port == "" {
		return nil
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = logger.ServiceName()
	}
	stopToken := strings.TrimSpace(cfg.StopToken)
	if stopToken == "" {
		stopToken = strings.TrimSpace(os.Getenv("PROBE_STOP_TOKEN"))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": cfg.ServiceName,
		})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		for _, check := range cfg.ReadyChecks {
			if check == nil {
				continue
			}
			checkCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			err := check(checkCtx)
			cancel()
			if err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{
					"status": "unready",
				})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		handleStop(w, r, stopToken, cfg.Stop)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("PROBE_METRICS_ENABLED") != "true" {
			http.NotFound(w, r)
			return
		}
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		writeJSON(w, http.StatusOK, map[string]any{
			"goroutines": runtime.NumGoroutine(),
			"alloc":      mem.Alloc,
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Ctx(ctx).Info("probe server starting",
			logger.String("port", cfg.Port),
			logger.String("service", cfg.ServiceName),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Ctx(ctx).Warn("probe server stopped unexpectedly", logger.Error(err))
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

func authorizedStopRequest(r *http.Request, expectedToken string) bool {
	if r == nil || expectedToken == "" {
		return false
	}
	providedToken := strings.TrimSpace(r.Header.Get("X-Probe-Token"))
	if providedToken == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)) == 1
}

func handleStop(w http.ResponseWriter, r *http.Request, expectedToken string, stop func()) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !authorizedStopRequest(r, expectedToken) {
		http.NotFound(w, r)
		return
	}
	if stop == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"status": "unsupported"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopping"})
	go stop()
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Global().Warn("probe: failed to encode JSON response", logger.Error(err))
	}
}
