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

type ViewCollector interface {
	// return collector name
	CollectorName() string
	// return the views created by this collector
	Views() []*view.View
	// real scrape
	Scrape(ctx context.Context, db *sql.DB)
}

func Collect(ctx context.Context, dsn string, viewCollectors map[ViewCollector]bool) {

	sqlDriverRegCtx, sqlDriverRegSpan := trace.StartSpan(ctx, "sql_register_driver")
	sqlDriverRegSpan.Annotate([]trace.Attribute{trace.StringAttribute("setp", "register_sql_driver")}, "Register the hana db driver into sql")

	driverName, err := ocsql.Register("hdb", ocsql.WithAllTraceOptions())

	if err != nil {
		log.Fatalf("Failed to register the ocsql driver: %v", err)

	}

	sqlDriverRegSpan.End()

	sqlOpenCtx, sqlOpenSpan := trace.StartSpan(sqlDriverRegCtx, "sql_open_db_conn")
	sqlOpenSpan.AddAttributes(trace.StringAttribute("step", "sql_opendb_conn"))
	sqlOpenSpan.Annotate([]trace.Attribute{trace.StringAttribute("setp", "sql_opendb_conn")}, "open sql connection to database")

	db, err := sql.Open(driverName, dsn)

	if err != nil {
		log.Fatalf("Failed to open the HANA database: %v", err)
		sqlOpenSpan.Annotate([]trace.Attribute{}, err.Error())
	}
	sqlOpenSpan.End()

	defer func() {
		db.Close()
		// Wait to 1 seconds so that the traces can be exported
		waitTime := 1 * time.Second
		log.Printf("Waiting for %s seconds to ensure all traces are exported before exiting", waitTime)
		<-time.After(waitTime)
	}()

	sqlQueryctx, sqlQuerySpan := trace.StartSpan(sqlOpenCtx, "sql_queries")
	sqlQuerySpan.AddAttributes(trace.StringAttribute("step", "sql_entry"))
	sqlQuerySpan.Annotate([]trace.Attribute{trace.StringAttribute("step", "prepare_sql_execution")}, "prepare to execute the sqls, here is the entry point")
	defer sqlQuerySpan.End()

	//enabledCollectors := viewCollectors
	//
	for viewCollector := range viewCollectors {
		viewCollector.Scrape(sqlQueryctx, db)

	}
}
