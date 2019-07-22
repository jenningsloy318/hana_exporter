// Scrape `sys_m_host_resource_utilization`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	hostResourceUtilizationQuery = ` select HOST,USED_PHYSICAL_MEMORY,FREE_PHYSICAL_MEMORY from SYS.M_HOST_RESOURCE_UTILIZATION
	`
	// Subsystem.
	hostResourceUtilization = "sys_m_host_resource_utilization"
)

// Metric descriptors.
var (
	hostResourceUtilizationLabels = []string{"host"}
	hostResourceUtilizationUsedPhysicalMemorydesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostResourceUtilization, "used_physical_memory_bytes"),
		"Used physical memory on the host (bytes) from sys.m_host_resource_utilization.",
		hostResourceUtilizationLabels, nil)
	hostResourceUtilizationFreePhysicalMemorydesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostResourceUtilization, "free_physical_memory_bytes"),
		"Free physical memory on the host(bytes) from sys.m_host_resource_utilization.",
		hostResourceUtilizationLabels, nil)
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
	var host string
	for hostResourceUtilizationRows.Next() {
		if err := hostResourceUtilizationRows.Scan(&host, &used_physical_memory, &free_physical_memory); err != nil {
			return err
		}
		hostResourceUtilizationLabelValues :=[]string{host}
		ch <- prometheus.MustNewConstMetric(hostResourceUtilizationUsedPhysicalMemorydesc, prometheus.GaugeValue, used_physical_memory, hostResourceUtilizationLabelValues...)
		ch <- prometheus.MustNewConstMetric(hostResourceUtilizationFreePhysicalMemorydesc, prometheus.GaugeValue, free_physical_memory, hostResourceUtilizationLabelValues...)

	}
	return nil

}
