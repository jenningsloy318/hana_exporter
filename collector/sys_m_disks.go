// Scrape `sys_m_disks`.
package collector

import (
	"database/sql"
	_ "github.com/SAP/go-hdb/driver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"log"
	"context"
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

func (disksCollector DisksCollector)CollectorName() string {
	return "DisksCollector"
}

func (disksCollector DisksCollector)NewViews() []*view.View {
	return DisksCollectorViews
}

func (disksCollector DisksCollector)UpdateMeasurements(db *sql.DB) {

	disksRow := db.QueryRow(disksQuery)

	var host string
	var path string
	var usage_type string
	var total_size int64
	var used_size int64
 
	err := disksRow.Scan(&host, &path, &usage_type, &total_size, &used_size)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No value returend")
	case err != nil:
		log.Fatal(err)
	}

	diskctx, err := tag.New(context.Background(),
		tag.Insert(hostTag, host),
		tag.Insert(pathTag, path),
		tag.Insert(usageTypeTag, usage_type),
	)

	if err != nil {
		log.Fatalf("Failed to insert tag: %v", err)

	}

	stats.Record(diskctx, total_size_measure.M(total_size))
	stats.Record(diskctx, used_size_measure.M(used_size))

}
