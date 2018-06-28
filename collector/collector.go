package collector

import (
	"regexp"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Exporter namespace.
	namespace = "hana"
	// Math constant for picoseconds to seconds.
	picoSeconds = 1e12

)

var logRE = regexp.MustCompile(`.+\.(\d+)$`)

func newDesc(subsystem, name, help string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help, nil, nil,
	)
}