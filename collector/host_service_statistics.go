// Scrape `SHOW GLOBAL STATUS`.

package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Scrape query.
	hostServiceStatisticsQuery = ` select SERVICE_NAME,ACTIVE_STATUS,ACTIVE_REQUEST_COUNT,PENDING_REQUEST_COUNT,ALL_FINISHED_REQUEST_COUNT,ALL_FINISHED_REQUEST_COUNT_DELTA,FINISHED_NON_INTERNAL_REQUEST_COUNT,FINISHED_NON_INTERNAL_REQUEST_COUNT_DELTA,REQUESTS_PER_SEC,RESPONSE_TIME,	PROCESS_ID,ACTIVE_THREAD_COUNT,THREAD_COUNT,PROCESS_CPU_TIME,TOTAL_CPU_TIME,PROCESS_PHYSICAL_MEMORY,PROCESS_MEMORY,PHYSICAL_MEMORY
	from "_SYS_STATISTICS"."HOST_SERVICE_STATISTICS" where snapshot_id in (select distinct max(snapshot_id) as snapshot_id from "_SYS_STATISTICS"."HOST_SERVICE_STATISTICS") AND SERVICE_NAME  != 'daemon';	`
	// Subsystem.
	hostServiceStatistics = "host_service_statistics"
)



// Metric descriptors.
var (
	hostServiceStatisticsActiveStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "active_status"),
		"Service Active Status from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)
	hostServiceStatisticsActiveRequestCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "active_request_count"),
		"Active Request Count from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)
	hostServiceStatisticsPendingRequestCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "pending_request_count"),
		"Pending Request Count from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)	
	hostServiceStatisticsAllFinishedRequestCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "all_finished_request_count"),
		"All Finished Request Count from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)	
	hostServiceStatisticsAllFinishedRequestCountDeltaDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "all_finished_request_count_delta"),
		"All Finished Request Count Delta from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)			
	finishedNonInternalRequestCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "finished_non_internal_request_count"),
		"Finished requests from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)						
	finishedNonInternalRequestCountDeltaDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "finished_non_internal_request_count_delta"),
		"Finished requests Delta from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)							
	requestsPerSecDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "requests_per_sec"),
		"Requests per second(Average over last 1000 requests) from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)						
	responseTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "response_time"),
		"Request response time(Average over last 1000 requests) from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)														
	processIdDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "process_id"),
		"process_id from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)						
	activeThreadCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "active_thread_count"),
		"Number of active threads from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)			
	threadCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "thread_count"),
		"Number of total threads from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)					
	processCpuTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "process_cpu_time"),
		"CPU usage of current process since start from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)							
	totalCpuTimeDesc = prometheus.NewDesc(
	  prometheus.BuildFQName(namespace, hostServiceStatistics, "total_cpu_time"),
	  "CPU usage of all processes since start from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)			
	processPhysicalMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "process_physical_memory"),
		"Process physical memory usage from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)								
	processMemoryDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, hostServiceStatistics, "process_memory"),
			"Process logical memory usage from Service statstics.",
			[]string{"service_name","hana_instance"}, nil,)					
	physicalMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, hostServiceStatistics, "physical_memory"),
		"Host physical memory size from Service statstics.",
		[]string{"service_name","hana_instance"}, nil,)						
)

// ScrapeHostServiceStatistics collects from `SHOW GLOBAL STATUS`.
type ScrapeHostServiceStatistics struct{}

// Name of the Scraper. Should be unique.
func (ScrapeHostServiceStatistics) Name() string {
	return hostServiceStatistics
}

// Help describes the role of the Scraper.
func (ScrapeHostServiceStatistics) Help() string {
	return "Collect service cpu and memory info from  _SYS_STATISTICS.HOST_SERVICE_MEMORY"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeHostServiceStatistics) Scrape(db *sql.DB, ch chan<- prometheus.Metric) error {
	hostServiceStatisticsRows, err := db.Query(hostServiceStatisticsQuery)
	if err != nil {
		return err
	}
	defer hostServiceStatisticsRows.Close()

	var service_name  string 
	var active_status  float64
	var active_request_count  float64
	var pending_request_count  float64
	var all_finished_request_count  float64
	var all_finished_request_count_delta  float64
	var finished_non_internal_request_count  float64
	var finished_non_internal_request_count_delta  float64
	var requests_per_sec  float64
	var response_time  float64
	var process_id  float64
	var active_thread_count  float64
	var thread_count  float64
	var process_cpu_time  float64
	var total_cpu_time  float64
	var process_physical_memory  float64
	var process_memory  float64
	var physical_memory  float64

	for hostServiceStatisticsRows.Next() {
		if err := hostServiceStatisticsRows.Scan(&service_name,&active_status,&active_request_count,&pending_request_count,&all_finished_request_count,&all_finished_request_count_delta,&finished_non_internal_request_count,&finished_non_internal_request_count_delta,&requests_per_sec,&response_time,&process_id,&active_thread_count,&thread_count,&process_cpu_time,&total_cpu_time,&process_physical_memory,&process_memory,&physical_memory); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(hostServiceStatisticsActiveStatusDesc, prometheus.GaugeValue, active_status, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(hostServiceStatisticsActiveRequestCountDesc, prometheus.GaugeValue, active_request_count, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(hostServiceStatisticsPendingRequestCountDesc, prometheus.GaugeValue, pending_request_count, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(hostServiceStatisticsAllFinishedRequestCountDesc, prometheus.GaugeValue, all_finished_request_count, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(hostServiceStatisticsAllFinishedRequestCountDeltaDesc, prometheus.GaugeValue, all_finished_request_count_delta, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(finishedNonInternalRequestCountDesc, prometheus.GaugeValue, finished_non_internal_request_count, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(finishedNonInternalRequestCountDeltaDesc, prometheus.GaugeValue, finished_non_internal_request_count_delta, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(requestsPerSecDesc, prometheus.GaugeValue, requests_per_sec, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(responseTimeDesc, prometheus.GaugeValue, response_time, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(processIdDesc, prometheus.GaugeValue, process_id, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(activeThreadCountDesc, prometheus.GaugeValue, active_thread_count, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(threadCountDesc, prometheus.GaugeValue, thread_count, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(processCpuTimeDesc, prometheus.GaugeValue, process_cpu_time, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(totalCpuTimeDesc, prometheus.GaugeValue, total_cpu_time, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(processPhysicalMemoryDesc, prometheus.GaugeValue, process_physical_memory, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(processMemoryDesc, prometheus.GaugeValue, process_memory, service_name,Hana_instance)
		ch <- prometheus.MustNewConstMetric(physicalMemoryDesc, prometheus.GaugeValue, physical_memory, service_name,Hana_instance)

			}
			return nil

	}


