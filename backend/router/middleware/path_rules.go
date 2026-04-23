package middleware

import (
	"encoding/json"
	"net/http"
)

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

var PathRules = map[string]struct {
	Methods []string
}{
	"/health":                     {Methods: []string{http.MethodGet}},
	"/run":                        {Methods: []string{http.MethodPost}},
	"/runs":                       {Methods: []string{http.MethodGet}},
	"/run-groups":                 {Methods: []string{http.MethodGet}},
	"/run-group-prompt-artifacts": {Methods: []string{http.MethodGet}},
}

func ValidatePath(r *http.Request, w http.ResponseWriter) bool {
	setCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return false
	}

	rules, exists := PathRules[r.URL.Path]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return false
	}
	methodAllowed := false
	for _, m := range rules.Methods {
		if r.Method == m {
			methodAllowed = true
			break
		}
	}
	if !methodAllowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return false
	}

	return true
}
