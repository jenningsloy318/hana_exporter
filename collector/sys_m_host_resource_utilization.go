// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	hostResourceUtilizationQuery = ` select USED_PHYSICAL_MEMORY,FREE_PHYSICAL_MEMORY from SYS.M_HOST_RESOURCE_UTILIZATION
	`
	// Subsystem.
	hostResourceUtilization = "sys_m_host_resource_utilization"
)



// Metric descriptors.
var (
	hostResourceUtilizationUsedPhysicalMemorydesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostResourceUtilization, "used_physical_memory"),
		"Used physical memory on the host.",
		[]string{"hana_instance"}, nil,)
	hostResourceUtilizationFreePhysicalMemorydesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, hostResourceUtilization, "free_physical_memory"),
			"Free physical memory on the host.",
			[]string{"hana_instance"}, nil,)		
)

// ScrapeHostResourceUtilization collects from `SYS.M_HOST_RESOURCE_UTILIZATION`.
type ScrapeHostResourceUtilization struct{}

// Name of the Scraper. Should be unique.
func (ScrapeHostResourceUtilization) Name() string {
	return hostResourceUtilization
}

// Help describes the role of the Scraper.
func (ScrapeHostResourceUtilization) Help() string {
	return "Collect  info from  SYS.M_HOST_RESOURCE_UTILIZATION"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeHostResourceUtilization) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	hostResourceUtilizationRows, err := db.Query(hostResourceUtilizationQuery)
	if err != nil {
		return err
	}
	defer hostResourceUtilizationRows.Close()

	var used_physical_memory float64 
	var free_physical_memory float64

	for hostResourceUtilizationRows.Next() {
		if err := hostResourceUtilizationRows.Scan(&used_physical_memory, &free_physical_memory); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(hostResourceUtilizationUsedPhysicalMemorydesc, prometheus.GaugeValue, used_physical_memory,Hana_instance)
		ch <- prometheus.MustNewConstMetric(hostResourceUtilizationFreePhysicalMemorydesc, prometheus.GaugeValue, free_physical_memory,Hana_instance)

			}
			return nil

	}


