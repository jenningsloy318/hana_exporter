// Scrape `sys_m_cs_tables`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	csTablesQuery = `SELECT TOP 5  HOST,PORT,SCHEMA_NAME,TABLE_NAME,PART_ID,MEMORY_SIZE_IN_TOTAL,RECORD_COUNT,READ_COUNT,WRITE_COUNT,MERGE_COUNT  FROM SYS.M_CS_TABLES   WHERE SCHEMA_NAME NOT LIKE 'SYS%'  AND SCHEMA_NAME NOT LIKE '_SYS%' AND SCHEMA_NAME NOT LIKE 'HANA%' AND  SCHEMA_NAME NOT LIKE 'UI%'   ORDER BY MEMORY_SIZE_IN_TOTAL DESC ;`
	// Subsystem.
	csTables = "sys_m_cs_tables"
)

// Metric descriptors.
var (
	csTablesLabels                = []string{"host", "port", "schema_name", "table_name", "part_id"}
	csTablesMemorySizeInTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "memory_size_in_total"),
		"total shared memory size of this table,Byte.",
		csTablesLabels, nil)
	csTablesRecordCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "record_count"),
		"Record count of this table or partition.",
		csTablesLabels, nil)
	csTablesReadCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "read_count"),
		"Number of read accesses on the table or partition.",
		csTablesLabels, nil)
	csTablesWriteCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "write_count"),
		"Number of write accesses on the table or partition.",
		csTablesLabels, nil)
	csTablesMergeCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "merge_count"),
		"Number of delta merges	done on the table or partition.",
		csTablesLabels, nil)
)

// Scrapedisks collects from `SYS.M_CS_TABLES;`.
type ScrapeCsTables struct{}

// Name of the Scraper. Should be unique.
func (ScrapeCsTables) Name() string {
	return csTables
}

// Help describes the role of the Scraper.
func (ScrapeCsTables) Help() string {
	return "Collect  info from  SYS.M_CS_TABLES;"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeCsTables) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	csTablesRows, err := db.Query(csTablesQuery)
	if err != nil {
		return err
	}
	defer csTablesRows.Close()

	var host string
	var port string
	var schema_name string
	var table_name string
	var part_id string
	var memory_size_in_total float64
	var record_count float64
	var read_count float64
	var write_count float64
	var merge_count float64

	for csTablesRows.Next() {
		if err := csTablesRows.Scan(&host, &port, &schema_name, &table_name, &part_id, &memory_size_in_total, &record_count, &read_count, &write_count, &merge_count); err != nil {
			return err
		}
		csTablesLabelValues := []string{host, port, schema_name, table_name, part_id}
		ch <- prometheus.MustNewConstMetric(csTablesMemorySizeInTotalDesc, prometheus.GaugeValue, memory_size_in_total, csTablesLabelValues...)
		ch <- prometheus.MustNewConstMetric(csTablesRecordCountDesc, prometheus.GaugeValue, record_count, csTablesLabelValues...)
		ch <- prometheus.MustNewConstMetric(csTablesReadCountDesc, prometheus.GaugeValue, read_count, csTablesLabelValues...)
		ch <- prometheus.MustNewConstMetric(csTablesWriteCountDesc, prometheus.GaugeValue, write_count, csTablesLabelValues...)
		ch <- prometheus.MustNewConstMetric(csTablesMergeCountDesc, prometheus.GaugeValue, merge_count, csTablesLabelValues...)

	}
	return nil

}
