// Scrape `sys_m_shared_memory`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	sharedMemoryQuery = `SELECT HOST,PORT, CATEGORY,ALLOCATED_SIZE,USED_SIZE,FREE_SIZE FROM SYS.M_SHARED_MEMORY;`
	// Subsystem.
	sharedMemory = "sys_m_shared_memory"
)

// Metric descriptors.
var (
	sharedMemoryAllocatedSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sharedMemory, "allocated_size"),
		"Allocated shared memory size on the module.",
		[]string{"host", "port", "category"}, nil)
	sharedMemoryUsedSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sharedMemory, "used_size"),
		"Used shared memory size on the module.",
		[]string{"host", "port", "category"}, nil)
	sharedMemoryFreeSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sharedMemory, "free_size"),
		"Used shared memory size on the module.",
		[]string{"host", "port", "category"}, nil)
)

// Scrapedisks collects from `SYS.M_SHARED_MEMORY;`.
type ScrapeSharedMemory struct{}

// Name of the Scraper. Should be unique.
func (ScrapeSharedMemory) Name() string {
	return sharedMemory
}

// Help describes the role of the Scraper.
func (ScrapeSharedMemory) Help() string {
	return "Collect  info from  SYS.M_SHARED_MEMORY;"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeSharedMemory) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	sharedMemoryRows, err := db.Query(sharedMemoryQuery)
	if err != nil {
		return err
	}
	defer sharedMemoryRows.Close()

	var host string
	var port string
	var category string
	var allocated_size float64
	var used_size float64
	var free_size float64

	for sharedMemoryRows.Next() {
		if err := sharedMemoryRows.Scan(&host, &port, &category, &allocated_size, &used_size, &free_size); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(sharedMemoryAllocatedSizeDesc, prometheus.GaugeValue, allocated_size, host, port, category)
		ch <- prometheus.MustNewConstMetric(sharedMemoryUsedSizeDesc, prometheus.GaugeValue, used_size, host, port, category)
		ch <- prometheus.MustNewConstMetric(sharedMemoryFreeSizeDesc, prometheus.GaugeValue, free_size, host, port, category)

	}
	return nil

}
