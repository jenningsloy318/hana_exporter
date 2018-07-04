// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"

)

const (
	// Scrape query.
	licenseStatusQuery = `select  hardware_key,system_id,product_limit,days_between(TO_SECONDDATE(CURRENT_TIMESTAMP),TO_SECONDDATE(expiration_date)) as expire_days from sys.m_license`
	// Subsystem.
	licenseStatus = "sys_m_license"
)



// Metric descriptors.
var (
	licenseStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, licenseStatus, "expire_days"),
		"License expire days from sys.m_service_statistics.",
		[]string{"hana_instance","hardware_key","system_id","product_limit"}, nil,)
		
)

// ScrapeserviceStatistics collects from `sys.m_service_statistics`.
type ScrapeLicenseStatus struct{}

// Name of the Scraper. Should be unique.
func (ScrapeLicenseStatus) Name() string {
	return licenseStatus
}

// Help describes the role of the Scraper.
func (ScrapeLicenseStatus) Help() string {
	return "Collect  info from  sys.m_service_statistics"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeLicenseStatus) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	licenseStatusRows, err := db.Query(licenseStatusQuery)
	if err != nil {
		return err
	}
	defer licenseStatusRows.Close()
	
	var hardware_key string
	var system_id string 
	var product_limit string 
  var expire_days float64 


	for licenseStatusRows.Next() {
		if err := licenseStatusRows.Scan(&hardware_key, &system_id, &product_limit, &expire_days); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(licenseStatusDesc, prometheus.GaugeValue, expire_days,Hana_instance,hardware_key,system_id,product_limit)

			}
			return nil

	}


