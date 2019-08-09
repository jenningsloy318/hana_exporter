// Scrape `sys_m_rs_tables`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	rsTablesQuery = `select TOP 5 (ALLOCATED_FIXED_PART_SIZE+ALLOCATED_VARIABLE_PART_SIZE) as TOTAL_ALLOCATED_SIZE  ,
	(USED_FIXED_PART_SIZE+USED_VARIABLE_PART_SIZE)as TOTAL_USED_SIZE ,	 SCHEMA_NAME,TABLE_NAME  from  SYS.M_RS_TABLES
	 WHERE SCHEMA_NAME NOT LIKE 'SYS%' AND SCHEMA_NAME NOT LIKE '_SYS%' AND SCHEMA_NAME NOT LIKE 'HANA%' AND  SCHEMA_NAME NOT LIKE 'UI%' and(ALLOCATED_FIXED_PART_SIZE+ALLOCATED_VARIABLE_PART_SIZE) !=0	 ORDER BY TOTAL_ALLOCATED_SIZE  DESC , TOTAL_USED_SIZE DESC;`
	// Subsystem.
	rsTables = "sys_m_rs_tables"
)

// Metric descriptors.
var (
	rsTablesLabels                = []string{ "schema_name", "table_name"}
	rsTablesTotalAllocatedSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, rsTables, "total_allocated_size"),
		"Total allocated memory size on this table, Byte.",
		rsTablesLabels, nil)
	rsTablesTotalUsedSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, rsTables, "total_used_size"),
		"Total Used memory of this table, Byte.",
		rsTablesLabels, nil)
)

// Scrapedisks collects from `SYS.M_RS_TABLES;`.
type ScrapeRsTables struct{}

// Name of the Scraper. Should be unique.
func (ScrapeRsTables) Name() string {
	return rsTables
}

// Help describes the role of the Scraper.
func (ScrapeRsTables) Help() string {
	return "Collect  info from  SYS.M_RS_TABLES;"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeRsTables) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	rsTablesRows, err := db.Query(rsTablesQuery)
	if err != nil {
		return err
	}
	defer rsTablesRows.Close()

	var total_allocated_size float64
	var total_used_size float64
	var schema_name string
	var table_name string

	for rsTablesRows.Next() {
		if err := rsTablesRows.Scan(&total_allocated_size, &total_used_size, &schema_name, &table_name); err != nil {
			return err
		}
		rsTablesLabelValues := []string{ schema_name, table_name}
		ch <- prometheus.MustNewConstMetric(rsTablesTotalAllocatedSizeDesc, prometheus.GaugeValue, total_allocated_size, rsTablesLabelValues...)
		ch <- prometheus.MustNewConstMetric(rsTablesTotalUsedSizeDesc, prometheus.GaugeValue, total_used_size, rsTablesLabelValues...)

	}
	return nil

}
