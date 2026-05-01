package api

import (
	"fmt"
	"net/http"
	"runtime"
)

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP go_memstats_alloc_bytes Bytes allocated and still in use.\n# TYPE go_memstats_alloc_bytes gauge\ngo_memstats_alloc_bytes %d\n", m.Alloc)
	fmt.Fprintf(w, "go_memstats_sys_bytes %d\n", m.Sys)
	fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())
	fmt.Fprintf(w, "go_gc_duration_seconds %f\n", float64(m.PauseTotalNs)/1e9)
}
