// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	hostServiceMemoryQuery = `select host,port,service_name,total_memory_used_size from "_SYS_STATISTICS"."HOST_SERVICE_MEMORY" where snapshot_id in (select distinct max(snapshot_id) as snapshot_id from "_SYS_STATISTICS"."HOST_SERVICE_MEMORY") 	`
	// Subsystem.
	hostServiceMemory = "host_service_memory"
)



// Metric descriptors.
var (
	hostServiceMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceMemory, "total_memory_used_size"),
		"Service memory usage.",
		[]string{"service_name","hana_instance","host","port"}, nil,
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

	var service_name string 
	var total_memory_used_size float64
	var host string
	var port string 
	for hostServiceMemoryRows.Next() {
		if err := hostServiceMemoryRows.Scan(&host, &port, &service_name, &total_memory_used_size); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(hostServiceMemoryDesc, prometheus.GaugeValue, total_memory_used_size, service_name,Hana_instance,host, port)

			}
			return nil

	}


