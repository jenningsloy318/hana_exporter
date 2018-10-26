package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"
	"github.com/opencensus-integrations/ocsql"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	//	"go.opencensus.io/exporter/prometheus"
	"context"
	"log"
	"time"
)

type HANACollector interface {
	// return the views created by this collector
	NewViews() []*view.View
	// real scrape
	UpdateMeasurementData(ctx context.Context, db *sql.DB)
}

var hanaCollectors = []HANACollector{
	LicenseCollector{},
	DisksCollector{},
}
var hanaViews []*view.View

func NewHanaViews() []*view.View {
	for _, collector := range hanaCollectors {
		hanaViews = append(hanaViews, collector.NewViews()...)
	}
	return hanaViews

}

func CollectMeasurement(ctx context.Context, dsn string) {

	sqlDriverRegCtx, sqlDriverRegSpan := trace.StartSpan(ctx, "sql:register_driver")
	sqlDriverRegSpan.Annotate([]trace.Attribute{trace.StringAttribute("setp", "sql:register_driver")}, "Register the hana db driver into sql")

	driverName, err := ocsql.Register("hdb", ocsql.WithAllTraceOptions())

	if err != nil {
		log.Fatalf("Failed to register the ocsql driver: %v", err)

	}

	sqlDriverRegSpan.End()

	sqlOpenCtx, sqlOpenSpan := trace.StartSpan(sqlDriverRegCtx, "sql:open_db_connnection")
	sqlOpenSpan.AddAttributes(trace.StringAttribute("step", "sql:open_db_connnection"))
	sqlOpenSpan.Annotate([]trace.Attribute{trace.StringAttribute("setp", "sql:open_db_connnection")}, "open sql connection to database")

	db, err := sql.Open(driverName, dsn)

	if err != nil {
		log.Fatalf("Failed to open the HANA database: %v", err)
		sqlOpenSpan.Annotate([]trace.Attribute{}, err.Error())
	}
	sqlOpenSpan.End()

	sqlQueryctx, sqlQuerySpan := trace.StartSpan(sqlOpenCtx, "sql:exution_entrance")
	sqlQuerySpan.AddAttributes(trace.StringAttribute("step", "sql:exution_entrance"))
	sqlQuerySpan.Annotate([]trace.Attribute{trace.StringAttribute("step", "sql:exution_entrance")}, "prepare to execute the sqls, here is the entry point")
	defer sqlQuerySpan.End()

	for _, hanaCollector := range hanaCollectors {
			hanaCollector.UpdateMeasurementData(sqlQueryctx, db)
	}

	defer func() {
		db.Close()
		// Wait to 1 seconds so that the traces can be exported
		waitTime := 10 * time.Millisecond
		log.Printf("Waiting for %s Millisecond  to ensure all traces are exported before exiting", waitTime)
		<-time.After(waitTime)
	}()

}
