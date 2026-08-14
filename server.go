package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func NewServer(app *App, logger Logger) http.Handler {
	if logger == nil {
		logger = standardLogger{}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")
		mac := r.URL.Query().Get("mac")
		remote := remoteIP(r.RemoteAddr)
		logger.Printf("%s %s action=%s mac=%s remote=%s", r.Method, r.URL.RequestURI(), action, mac, remote)

		if r.URL.Path == "/healthz" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("ok\n"))
			return
		}
		if r.URL.Path == "/debug/status" {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(app.DebugStatus()); err != nil {
				logger.Printf("encode debug status failed: %v", err)
			}
			return
		}
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		app.RecordRequest(r.Method, r.URL.RequestURI(), action, mac, remote)

		switch r.Method {
		case http.MethodGet:
			if action == "" {
				writeDebugHelp(w, app)
				return
			}
			if action != "poll" {
				http.Error(w, "unsupported GET action", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(app.Poll(mac, remote)); err != nil {
				logger.Printf("encode poll response failed: %v", err)
			}
		case http.MethodPost:
			if action == "" {
				http.Error(w, "missing action", http.StatusBadRequest)
				return
			}
			if err := app.HandleEvent(r.Context(), action, mac, remote); err != nil {
				logger.Printf("event action=%s failed: %v", action, err)
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func writeDebugHelp(w http.ResponseWriter, app *App) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	status := app.DebugStatus()
	lines := []string{
		"PuKK PoC server is running.",
		"",
		"Health: /healthz",
		"Debug:  /debug/status",
		"PuKK poll endpoint: /?action=poll&mac=<device-mac>",
		"",
		"Requests seen by this process:",
		"  total: " + intString(status.TotalRequests),
		"  poll:  " + intString(status.PollRequests),
		"  event: " + intString(status.EventRequests),
	}
	_, _ = w.Write([]byte(strings.Join(lines, "\n") + "\n"))
}

func intString(value int) string {
	return strconv.Itoa(value)
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}
