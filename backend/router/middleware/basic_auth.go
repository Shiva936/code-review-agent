package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Shiva936/code-review-agent/backend/config"
)

func BasicAuth(cfg *config.Config, next http.Handler) http.Handler {
	// If credentials are not set, do not enforce auth.
	if cfg == nil || (cfg.Auth.Username == "" && cfg.Auth.Password == "") {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := parseBasicAuth(r.Header.Get("Authorization"))
		if !ok || user != cfg.Auth.Username || pass != cfg.Auth.Password {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseBasicAuth(header string) (string, string, bool) {
	if header == "" {
		return "", "", false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Basic") {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}
	creds := string(decoded)
	i := strings.IndexByte(creds, ':')
	if i < 0 {
		return "", "", false
	}
	return creds[:i], creds[i+1:], true
}
