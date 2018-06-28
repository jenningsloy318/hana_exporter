// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	hostServiceMemoryQuery = `select service_name,total_memory_used_size from "_SYS_STATISTICS"."HOST_SERVICE_MEMORY" where snapshot_id in (select distinct max(snapshot_id) as snapshot_id from "_SYS_STATISTICS"."HOST_SERVICE_MEMORY") 
	`
	// Subsystem.
	hostServiceMemory = "host_service_memory"
)



// Metric descriptors.
var (
	hostServiceMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceMemory, "total_memory_used_size"),
		"Total  memory used size by HANA service Byte .",
		[]string{"service_name"}, nil,
	)
)

// ScrapeHostServiceMemory collects from `SHOW GLOBAL STATUS`.
type ScrapeHostServiceMemory struct{}

// Name of the Scraper. Should be unique.
func (ScrapeHostServiceMemory) Name() string {
	return hostServiceMemory
}

// Help describes the role of the Scraper.
func (ScrapeHostServiceMemory) Help() string {
	return "Collect from  _SYS_STATISTICS.HOST_SERVICE_MEMORY"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeHostServiceMemory) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	hostServiceMemoryRows, err := db.Query(hostServiceMemoryQuery)
	if err != nil {
		return err
	}
	defer hostServiceMemoryRows.Close()

	var key string
	var val float64

	for hostServiceMemoryRows.Next() {
		if err := hostServiceMemoryRows.Scan(&key, &val); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(hostServiceMemoryDesc, prometheus.GaugeValue, val, key,)

			}
			return nil

	}


