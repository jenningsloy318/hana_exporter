package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	//	"go.opencensus.io/exporter/prometheus"
	"context"
	"log"
	"time"
)

type ViewCollector interface {
	// return collector name
	CollectorName() string
	// return the views created by this collector
	NewViews() []*view.View
	// real scrape
	UpdateMeasurements(db *sql.DB)
}


func NewHanaViews(dsn string)  []*view.View {
	driverName, err := sql.Register("hdb")
	if err != nil {
		log.Fatalf("Failed to register the ocsql driver: %v", err)

	}
	db, err := sql.Open(driverName, dsn)

	if err != nil {
		log.Fatalf("Failed to open the HANA database: %v", err)
	}
	var disksCollector = DisksCollector{}
	var licenseCollector = LicenseCollector{}
 var HanaCollectors = []ViewCollector{licenseCollector,licenseCollector}
 var HanaViews []*view.View
for collector = range HanaCollectors {
	HanaViews=append(HanaViews,collector.NewViews(db))
}

}