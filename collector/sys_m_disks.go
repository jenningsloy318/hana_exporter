// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	disksQuery = `SELECT HOST,PATH,USAGE_TYPE,TOTAL_SIZE,USED_SIZE FROM SYS.M_DISKS`
	// Subsystem.
	disks = "sys_m_disks"
)

// Metric descriptors.
var (
	disksTotalSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, disks, "totalSize"),
		"Volume Size.",
		[]string{"hana_instance", "host", "path", "usage_type"}, nil)
	disksUsedSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, disks, "usedSize"),
		"Volume Used Space.",
		[]string{"hana_instance", "host", "path", "usage_type"}, nil)
)

// Scrapedisks collects from `SYS.M_DISKS;`.
type ScrapeDisks struct{}

// Name of the Scraper. Should be unique.
func (ScrapeDisks) Name() string {
	return disks
}

// Help describes the role of the Scraper.
func (ScrapeDisks) Help() string {
	return "Collect  info from  SYS.M_DISKS;"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeDisks) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	disksRows, err := db.Query(disksQuery)
	if err != nil {
		return err
	}
	defer disksRows.Close()

	var host string
	var path string
	var usage_type string
	var total_size float64
	var used_size float64

	for disksRows.Next() {
		if err := disksRows.Scan(&host, &path, &usage_type, &total_size, &used_size); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(disksTotalSizeDesc, prometheus.GaugeValue, total_size, Hana_instance, host, path, usage_type)
		ch <- prometheus.MustNewConstMetric(disksUsedSizeDesc, prometheus.GaugeValue, used_size, Hana_instance, host, path, usage_type)

	}
	return nil

}
