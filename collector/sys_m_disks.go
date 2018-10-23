// Scrape `sys_m_disks`.
package collector

import (
	"context"
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"log"
)

const (
	// Scrape query.
	disksQuery = `SELECT HOST,PATH,USAGE_TYPE,TOTAL_SIZE,USED_SIZE FROM SYS.M_DISKS`
	// Subsystem.
	disks = "sys_m_disks"
)

type DisksCollector struct{}

var (
	hostTag, _         = tag.NewKey("host")
	pathTag, _         = tag.NewKey("path")
	usageTypeTag, _    = tag.NewKey("UsageType")
	total_size_measure = stats.Int64("total_size", "Volume Size.", "MB")
	used_size_measure  = stats.Int64("used_size", "Volume Used Space.", "MB")

	totalSizeView = &view.View{
		Name:        "sys_m_disks/total_size",
		Description: "Volume Size.",
		TagKeys:     []tag.Key{hostTag, pathTag, usageTypeTag},
		Measure:     total_size_measure,
		Aggregation: view.LastValue(),
	}
	usedSizeView = &view.View{
		Name:        "sys_m_disks/used_size",
		Description: "Volume Size.",
		TagKeys:     []tag.Key{hostTag, pathTag, usageTypeTag},
		Measure:     used_size_measure,
		Aggregation: view.LastValue(),
	}

	DisksCollectorViews = []*view.View{
		totalSizeView,
		usedSizeView,
	}
)

func (DisksCollector) CollectorName() string {
	return "DisksCollector"
}

func (DisksCollector) Views() []*view.View {

	return DisksCollectorViews
}

func (DisksCollector) Scrape(ctx context.Context, db *sql.DB) {

	ctx, span := trace.StartSpan(ctx, "sql_query_sys_m_disks")
	span.Annotate([]trace.Attribute{trace.StringAttribute("step", "excute_sql_query_sys_m_disks")}, "excuting sql_query_sys_m_disks sql to get the disk info and exported as metrics")

	//get the sql data
	defer span.End()

	// if muliple row returned
	disksRow := db.QueryRowContext(ctx, disksQuery)

	var host string
	var path string
	var usage_type string
	var total_size int64
	var used_size int64

	sqlRowsScanCtx, sqlRowsScanSpan := trace.StartSpan(ctx, "sql_row_scan")
	sqlRowsScanSpan.Annotate([]trace.Attribute{trace.StringAttribute("step", "get the value from sql returns and update it to variables")}, "get the value from sql returns and update it to variables")

	err := disksRow.Scan(&host, &path, &usage_type, &total_size, &used_size)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No value returend")
		sqlRowsScanSpan.Annotate([]trace.Attribute{}, "no rows returned")

	case err != nil:
		log.Fatal(err)
		sqlRowsScanSpan.Annotate([]trace.Attribute{}, err.Error())

	}

	defer sqlRowsScanSpan.End()

	measureSetCtx, measureSetCtxSpan := trace.StartSpan(sqlRowsScanCtx, "measure_value_set")
	measureSetCtxSpan.Annotate([]trace.Attribute{trace.StringAttribute("step", "update_the_measurement")}, "use the variables to update the measurements")
	diskctx, err := tag.New(measureSetCtx,
		tag.Insert(hostTag, host),
		tag.Insert(pathTag, path),
		tag.Insert(usageTypeTag, usage_type),
	)

	if err != nil {
		log.Fatalf("Failed to insert tag: %v", err)

	}

	stats.Record(diskctx, total_size_measure.M(total_size))
	stats.Record(diskctx, used_size_measure.M(used_size))

	measureSetCtxSpan.End()
}
