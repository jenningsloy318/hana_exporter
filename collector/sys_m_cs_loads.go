// Scrape `sys_m_cs_unloads`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	csLoadsQuery = `SELECT COUNT(*) CS_LOAD_COUNT,SCHEMA_NAME  FROM "SYS"."M_CS_LOADS"   WHERE SCHEMA_NAME NOT  LIKE '_SYS%'  GROUP BY SCHEMA_NAME  ORDER BY COUNT(*) DESC;`
	// Subsystem.
	csLoads = "sys_m_cs_loads"
)

// Metric descriptors.
var (
	csLoadsLabels                    = []string{"schema"}
	csLoadsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, csLoads, "count"),
		"column unloads count.",
		csLoadsLabels, nil)
)

// Scrapedisks collects from `M_CS_LOADS;`.
type ScrapeCsLoads struct{}

// Name of the Scraper. Should be unique.
func (ScrapeCsLoads) Name() string {
	return csLoads
}

// Help describes the role of the Scraper.
func (ScrapeCsLoads) Help() string {
	return "Collect  info from  SYS.M_CS_LOADS;"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeCsLoads) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	csLoadsRows, err := db.Query(csLoadsQuery)
	if err != nil {
		return err
	}
	defer csLoadsRows.Close()

	var schema string
	var cs_load_count  float64
	for csLoadsRows.Next() {
		if err := csLoadsRows.Scan(&cs_load_count, &schema); err != nil {
			return err

		}
		csLoadsLabelValues := []string{schema}
			ch <- prometheus.MustNewConstMetric(csLoadsCountDesc, prometheus.GaugeValue, cs_load_count, csLoadsLabelValues...)
	}
	return nil

}
