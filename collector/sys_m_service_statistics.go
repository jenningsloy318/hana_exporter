// Scrape `sys_m_service_statistics`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	serviceStatisticsQuery = ` select SERVICE_NAME,HOST,PORT,ACTIVE_STATUS, seconds_between(TO_SECONDDATE(START_TIME),TO_SECONDDATE(SYS_TIMESTAMP)) as DURATION, PROCESS_CPU_TIME,TOTAL_CPU_TIME,TOTAL_CPU,PROCESS_PHYSICAL_MEMORY,PHYSICAL_MEMORY,REQUESTS_PER_SEC,RESPONSE_TIME,FINISHED_NON_INTERNAL_REQUEST_COUNT,ACTIVE_REQUEST_COUNT,PENDING_REQUEST_COUNT,ACTIVE_THREAD_COUNT,THREAD_COUNT from SYS.M_SERVICE_STATISTICS	`
	// Subsystem.
	serviceStatistics = "sys_m_service_statistics"
)

// Metric descriptors.
var (
	serviceStatisticsLabels =append(BaseLabelNames,"service_name", "service_status", "host", "port")

	serviceStatisticsActiveStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "status"),
		"Service Active Status from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "status_duration_seconds"),
		"Current service status duration (seconds) from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsProcessCPUTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "process_cpu_time"),
		"CPU usage of current process since start from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsTotalCPUTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "total_cpu_time"),
		"CPU usage of all processes	since start from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsTotalCPUDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "total_cpu"),
		"CPU usage of all processes from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)

	serviceStatisticsProcessPhysicalMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "process_physical_memory"),
		"Process physical memory usage from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsPhysicalMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "physical_memory"),
		"Process physical memory usage from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsRequestsPerSecDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "requests_per_sec"),
		"Requests per second. Average over last 1000 requests from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsResponseTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "response_time"),
		"Request response time. Average over last 1000 requests from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsFinishedNonInternalRequestCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "finished_non_internal_request_count"),
		"Finished requests from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsActiveRequestCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "active_request_count"),
		"Number of active requests from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsPendingRequestCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "pending_request_count"),
		"Number of pending requests from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsActiveThreadCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "active_thread_count"),
		"active_thread_count from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
	serviceStatisticsThreadCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceStatistics, "thread_count"),
		"active_thread_count from sys.m_service_statistics.",
		serviceStatisticsLabels, nil)
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
	var process_cpu_time float64
	var total_cpu_time float64
	var total_cpu float64
	var process_physical_memory float64
	var physical_memory float64
	var requests_per_sec float64
	var response_time float64
	var finished_non_internal_request_count float64
	var active_request_count float64
	var pending_request_count float64
	var ctive_thread_count float64
	var thread_count float64
	for serviceStatisticsRows.Next() {
		if err := serviceStatisticsRows.Scan(&service_name, &host, &port, &active_status, &duration, &process_cpu_time, &total_cpu_time, &total_cpu, &process_physical_memory, &physical_memory, &requests_per_sec, &response_time, &finished_non_internal_request_count, &active_request_count, &pending_request_count, &ctive_thread_count, &thread_count); err != nil {
			return err
		}

		serviceStatisticsLabelValues :=append(BaseLabelValues,service_name,string(active_status), host, port)

		if active_statusVal, ok := parseStatus(active_status); ok {
			ch <- prometheus.MustNewConstMetric(serviceStatisticsActiveStatusDesc, prometheus.GaugeValue, active_statusVal, serviceStatisticsLabelValues...)
		}

		ch <- prometheus.MustNewConstMetric(serviceStatisticsDurationDesc, prometheus.GaugeValue, duration, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsProcessCPUTimeDesc, prometheus.GaugeValue, process_cpu_time, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsTotalCPUTimeDesc, prometheus.GaugeValue, total_cpu_time, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsTotalCPUDesc, prometheus.GaugeValue, total_cpu, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsProcessPhysicalMemoryDesc, prometheus.GaugeValue, process_physical_memory, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsPhysicalMemoryDesc, prometheus.GaugeValue, physical_memory, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsRequestsPerSecDesc, prometheus.GaugeValue, requests_per_sec, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsResponseTimeDesc, prometheus.GaugeValue, response_time, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsFinishedNonInternalRequestCountDesc, prometheus.CounterValue, finished_non_internal_request_count, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsActiveRequestCountDesc, prometheus.CounterValue, active_request_count, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsPendingRequestCountDesc, prometheus.CounterValue, pending_request_count, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsActiveThreadCountDesc, prometheus.GaugeValue, ctive_thread_count, serviceStatisticsLabelValues...)
		ch <- prometheus.MustNewConstMetric(serviceStatisticsThreadCountDesc, prometheus.GaugeValue, thread_count, serviceStatisticsLabelValues...)

	}
	return nil

}
 