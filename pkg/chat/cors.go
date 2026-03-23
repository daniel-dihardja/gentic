package chat

import "net/http"

// CORS wraps next with CORS headers for the given allowed origins.
// Pass an empty slice or []string{"*"} to allow all origins.
func CORS(allowedOrigins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := resolveOrigin(origin, allowedOrigins)
		if allowed != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowed)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func resolveOrigin(origin string, allowed []string) string {
	if len(allowed) == 0 || (len(allowed) == 1 && allowed[0] == "*") {
		return "*"
	}
	for _, o := range allowed {
		if o == "*" || o == origin {
			return origin
		}
	}
	return ""
}
