package router

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"

	"github.com/dialogs/dialog-go-lib/service/info"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// AdminRouter router for administration functions
type AdminRouter struct {
	appinfo *info.Info
	mux     *http.ServeMux
}

// NewAdminRouter create router for administration functions
func NewAdminRouter(appinfo *info.Info) *AdminRouter {

	a := &AdminRouter{
		appinfo: appinfo,
		mux:     http.NewServeMux(),
	}

	a.HandleFunc("/health", a.health)
	a.HandleFunc("/info", a.info)
	a.Handle("/metrics", promhttp.Handler())
	a.HandleFunc("/debug/pprof/", pprof.Index)
	a.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	a.HandleFunc("/debug/pprof/profile", pprof.Profile)
	a.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	a.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return a
}

// Info return application info
func (a *AdminRouter) Info() *info.Info {
	return a.appinfo
}

// HandleFunc registers the handler function for the given pattern
func (a *AdminRouter) HandleFunc(path string, handler http.HandlerFunc) {
	a.mux.HandleFunc(path, handler)
}

// Handle registers the handler for the given pattern
func (a *AdminRouter) Handle(path string, handler http.Handler) {
	a.mux.Handle(path, handler)
}

// ServeHTTP dispatches the request (http.Handler implementation)
func (a *AdminRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.mux.ServeHTTP(w, req)
}

// Health handler function for livenness and readiness probes
func (a *AdminRouter) health(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Info returns service information
func (a *AdminRouter) info(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := json.NewEncoder(w).Encode(a.appinfo)
	if err != nil {
		_, _ = w.Write([]byte("unknown"))
	}
}
