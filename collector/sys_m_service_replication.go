// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	serviceReplicationQuery = `SELECT HOST,PATH,USAGE_TYPE,TOTAL_SIZE,USED_SIZE FROM M_SERVICE_REPLICATION`
	// Subsystem.
	serviceReplication = "sys_m_disks"
)

// Metric descriptors.
var (
	disksTotalSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, disks, "total_size"),
		"Volume Size.",
		[]string{"hana_instance", "host", "path", "usage_type"}, nil)
	disksUsedSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, disks, "used_size"),
		"Volume Used Space.",
		[]string{"hana_instance", "host", "path", "usage_type"}, nil)
)

// Scrapedisks collects from `M_SERVICE_REPLICATION;`.
type ScrapeServiceReplication struct{}

// Name of the Scraper. Should be unique.
func (ScrapeServiceReplication) Name() string {
	return serviceReplication
}

// Help describes the role of the Scraper.
func (ScrapeServiceReplication) Help() string {
	return "Collect  info from  M_SERVICE_REPLICATION;"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeServiceReplication) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	serviceReplicationRows, err := db.Query(serviceReplicationQuery)
	if err != nil {
		return err
	}
	defer serviceReplicationRows.Close()

	var host string
	var path string
	var usage_type string
	var total_size float64
	var used_size float64

	for serviceReplicationRows.Next() {
		if err := serviceReplicationRows.Scan(&host, &path, &usage_type, &total_size, &used_size); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(disksTotalSizeDesc, prometheus.GaugeValue, total_size, Hana_instance, host, path, usage_type)
		ch <- prometheus.MustNewConstMetric(disksUsedSizeDesc, prometheus.GaugeValue, used_size, Hana_instance, host, path, usage_type)

	}
	return nil

}
