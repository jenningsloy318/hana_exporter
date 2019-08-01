// Scrape `sys_m_service_replication`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	serviceReplicationQuery = ` SELECT HOST,PORT,VOLUME_ID,SECONDARY_HOST,SECONDARY_PORT,SECONDARY_ACTIVE_STATUS,SECONDARY_FULLY_RECOVERABLE,REPLICATION_MODE,REPLICATION_STATUS   from SYS.M_SERVICE_REPLICATION;`
	// Subsystem.
	serviceReplication = "sys_m_service_replication"
)

// Metric descriptors.
var (
	serviceReplicationLabels                    = []string{"host", "port", "volume_id", "secondary_host", "secondary_port"}
	serviceReplicationSecondaryActiveStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceReplication, "secondary_active_status"),
		"Secondary Active Status.",
		serviceReplicationLabels, nil)
	serviceReplicationSecondaryFullRecoverableDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceReplication, "secondary_fully_recoverable"),
		"Indicates if secondary is fully recoverable.",
		serviceReplicationLabels, nil)
	serviceReplicationReplicationStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceReplication, "replication_status"),
		"Replication Status.",
		serviceReplicationLabels, nil)
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
	var port string
	var volume_id string
	var secondary_host string
	var secondary_port string
	var secondary_active_status sql.RawBytes
	var secondary_fully_recoverable sql.RawBytes
	var replication_mode string
	var replication_status sql.RawBytes

	for serviceReplicationRows.Next() {
		if err := serviceReplicationRows.Scan(&host, &port, &volume_id, &secondary_host, &secondary_port, &secondary_active_status, &secondary_fully_recoverable, &replication_mode, &replication_status); err != nil {
			return err

		}
		serviceReplicationLabelValues := []string{host, port, volume_id, secondary_host, secondary_port}
		if secondary_active_statusVal, ok := parseStatus(secondary_active_status); ok {
			ch <- prometheus.MustNewConstMetric(serviceReplicationSecondaryActiveStatusDesc, prometheus.GaugeValue, secondary_active_statusVal, serviceReplicationLabelValues...)
		}
		if secondary_fully_recoverableVal, ok := parseStatus(secondary_fully_recoverable); ok {
			ch <- prometheus.MustNewConstMetric(serviceReplicationSecondaryFullRecoverableDesc, prometheus.GaugeValue, secondary_fully_recoverableVal, serviceReplicationLabelValues...)
		}
		if replication_statusVal, ok := parseStatus(replication_status); ok {
			ch <- prometheus.MustNewConstMetric(serviceReplicationReplicationStatusDesc, prometheus.GaugeValue, replication_statusVal, serviceReplicationLabelValues...)
		}
	}
	return nil

}
