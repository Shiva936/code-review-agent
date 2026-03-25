package router

import (
	"net/http"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/router/middleware"
)

func Init(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()
	rl := middleware.NewMemoryRateLimiter()

	wrap := func(path string, requireAuth bool, h func(*config.Config, http.ResponseWriter, *http.Request)) http.Handler {
		base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ok := middleware.ValidatePath(r, w); !ok {
				return
			}
			h(cfg, w, r)
		})

		// Auth first (inner), rate limit outer so even auth failures are limited if configured.
		handler := http.Handler(base)
		if requireAuth {
			handler = middleware.BasicAuth(cfg, handler)
		}
		handler = middleware.RateLimit(cfg, path, rl, handler)
		return handler
	}

	mux.Handle("/health", wrap("/health", false, healthHandler))
	mux.Handle("/runs", wrap("/runs", false, runsHandler))

	// Protected
	mux.Handle("/run", wrap("/run", true, runHandler))
	mux.Handle("/run-groups", wrap("/run-groups", true, runGroupsHandler))

	// Catch-all for JSON 404/method handling.
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = middleware.ValidatePath(r, w)
	}))

	return mux
}
