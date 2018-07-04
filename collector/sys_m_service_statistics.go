// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"

)

const (
	// Scrape query.
	serviceStatisticsQuery = ` select SERVICE_NAME,HOST,PORT,ACTIVE_STATUS, seconds_between(TO_SECONDDATE(START_TIME),TO_SECONDDATE(SYS_TIMESTAMP)) as DURATION  from SYS.M_SERVICE_STATISTICS	`
	// Subsystem.
	serviceStatistics = "sys_m_service_statistics"
)



// Metric descriptors.
var (
	serviceStatisticsActiveStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "status"),
		"Service Active Status from sys.m_service_statistics.",
		[]string{"hana_instance","service_name","host","port"}, nil,)
	serviceStatisticsDurationDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, serviceStatistics, "status_duration_seconds"),
			"Current service status duration (seconds) from sys.m_service_statistics.",
			[]string{"hana_instance","service_name","service_status","host","port"}, nil,)		
)

// ScrapeserviceStatistics collects from `SYS.M_SERVICE_STATISTICS`.
type ScrapeServiceStatistics struct{}

// Name of the Scraper. Should be unique.
func (ScrapeServiceStatistics) Name() string {
	return serviceStatistics
}

// Help describes the role of the Scraper.
func (ScrapeServiceStatistics) Help() string {
	return "Collect  info from  SYS.M_SERVICE_STATISTICS"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeServiceStatistics) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	serviceStatisticsRows, err := db.Query(serviceStatisticsQuery)
	if err != nil {
		return err
	}
	defer serviceStatisticsRows.Close()

	var service_name string 
	var host string
	var port string
	var active_status sql.RawBytes
	var duration float64


	for serviceStatisticsRows.Next() {
		if err := serviceStatisticsRows.Scan(&service_name, &host, &port, &active_status, &duration); err != nil {
			return err
		}
		if active_statusVal, ok := parseStatus(active_status); ok { 
			ch <- prometheus.MustNewConstMetric(serviceStatisticsActiveStatusDesc, prometheus.GaugeValue, active_statusVal, Hana_instance,service_name,host,port)
		}
		ch <- prometheus.MustNewConstMetric(serviceStatisticsDurationDesc, prometheus.GaugeValue, duration,Hana_instance,service_name,string(active_status),host,port)

			}
			return nil

	}


