// Scrape `sys_m_system_replication`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	systemReplicationQuery = `select SITE_NAME,SITE_ID,SECONDARY_SITE_NAME,SECONDARY_SITE_ID,REPLICATION_MODE,REPLICATION_STATUS,OPERATION_MODE,TIER from SYS.M_SYSTEM_REPLICATION;`
	// Subsystem.
	systemReplication = "sys_m_system_replication"
)

// Metric descriptors.
var (
	systemReplicationLabels                    = []string{"site_name", "site_id", "secondary_site_name", "secondary_site_id", "replication_mode","operation_mode","tier"}
	systemReplicationStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemReplication, "status"),
		"system replication Status, 1(ACTIVE).",
		systemReplicationLabels, nil)
)

// Scrapedisks collects from `M_SYSTEM_REPLICATION;`.
type ScrapeSystemReplication struct{}

// Name of the Scraper. Should be unique.
func (ScrapeSystemReplication) Name() string {
	return systemReplication
}

// Help describes the role of the Scraper.
func (ScrapeSystemReplication) Help() string {
	return "Collect  info from  SYS.M_SYSTEM_REPLICATION;"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeSystemReplication) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	systemReplicationRows, err := db.Query(systemReplicationQuery)
	if err != nil {
		return err
	}
	defer systemReplicationRows.Close()

	var site_name string
	var site_id string
	var secondary_site_name string
	var secondary_site_id string
	var replication_mode string
	var operation_mode string
	var tier string
	var replication_status  sql.RawBytes
	for systemReplicationRows.Next() {
		if err := systemReplicationRows.Scan(&site_name, &site_id, &secondary_site_name, &secondary_site_id, &replication_mode, &replication_status,&operation_mode, &tier); err != nil {
			return err

		}
		systemReplicationLabelValues := []string{site_name, site_id, secondary_site_name, secondary_site_id, replication_mode,operation_mode,tier}
		if replication_statusVal, ok := parseStatus(replication_status); ok {
			ch <- prometheus.MustNewConstMetric(systemReplicationStatusDesc, prometheus.GaugeValue, replication_statusVal, systemReplicationLabelValues...)
		}
	}
	return nil

}
