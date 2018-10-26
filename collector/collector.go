package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"
	"go.opencensus.io/stats/view"
	"log"
	"sync"
)

type HANACollector interface {
	// return collector name
	CollectorName() string
	// return the views created by this collector
	NewViews() []*view.View
	// real scrape
	UpdateMeasurements(db *sql.DB)
}

var hanaCollectors = []HANACollector{LicenseCollector{},DisksCollector{}}
var hanaViews []*view.View
func NewHanaViews()  []*view.View {

for _, collector := range hanaCollectors {
	hanaViews=append(hanaViews,collector.NewViews()...)
	
}
return hanaViews

}

func NewHanaMeasurements(dsn string) {
	db, err := sql.Open("hdb",dsn)
	
	if err != nil {
		log.Fatalf("Failed to open the HANA database: %v", err)
	}

	wg := &sync.WaitGroup{}
	defer wg.Wait()

for _, collector := range hanaCollectors {
	wg.Add(1)
 go func(hanaCollector HANACollector ){
	collector.UpdateMeasurements(db)
 }(collector)
}

}