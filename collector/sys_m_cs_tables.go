// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	csTablesQuery = `SELECT HOST,PORT,SCHEMA_NAME,TABLE_NAME,PART_ID,MEMORY_SIZE_IN_TOTAL,RECORD_COUNT,READ_COUNT,WRITE_COUNT,MERGE_COUNT FROM SYS.M_CS_TABLES WHERE SCHEMA_NAME != '_SYS_STATISTICS'  ORDER BY MEMORY_SIZE_IN_TOTAL DESC LIMIT 10;`
	// Subsystem.
	csTables = "sys_m_cs_tables"
)

// Metric descriptors.
var (
	csTablesMemorySizeInTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "memory_size_in_total"),
		"Allocated shared memory size on the module.",
		[]string{"hana_instance", "host", "port", "schema_name", "table_name", "part_id"}, nil)
	csTablesRecordCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "record_count"),
		"Used shared memory size on the module.",
		[]string{"hana_instance", "host", "port", "schema_name", "table_name", "part_id"}, nil)
	csTablesReadCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "read_count"),
		"Used shared memory size on the module.",
		[]string{"hana_instance", "host", "port", "schema_name", "table_name", "part_id"}, nil)
	csTablesWriteCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "write_count"),
		"Used shared memory size on the module.",
		[]string{"hana_instance", "host", "port", "schema_name", "table_name", "part_id"}, nil)
	csTablesMergeCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csTables, "merge_count"),
		"Used shared memory size on the module.",
		[]string{"hana_instance", "host", "port", "schema_name", "table_name", "part_id"}, nil)
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
		ch <- prometheus.MustNewConstMetric(csTablesMemorySizeInTotalDesc, prometheus.GaugeValue, memory_size_in_total, Hana_instance, host, port, schema_name, table_name, part_id)
		ch <- prometheus.MustNewConstMetric(csTablesRecordCountDesc, prometheus.GaugeValue, record_count, Hana_instance, host, port, schema_name, table_name, part_id)
		ch <- prometheus.MustNewConstMetric(csTablesReadCountDesc, prometheus.GaugeValue, read_count, Hana_instance, host, port, schema_name, table_name, part_id)
		ch <- prometheus.MustNewConstMetric(csTablesWriteCountDesc, prometheus.GaugeValue, write_count, Hana_instance, host, port, schema_name, table_name, part_id)
		ch <- prometheus.MustNewConstMetric(csTablesMergeCountDesc, prometheus.GaugeValue, merge_count, Hana_instance, host, port, schema_name, table_name, part_id)

	}
	return nil

}
