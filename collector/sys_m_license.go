// Scrape `sys_m_license`.

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
	licenseStatusQuery = `select  hardware_key,system_id,product_limit,days_between(TO_SECONDDATE(CURRENT_TIMESTAMP),TO_SECONDDATE(expiration_date)) as expire_days from sys.m_license`
	// Subsystem.
	licenseStatus = "sys_m_license"
)

var (
	hardwareKeyTag, _   = tag.NewKey("hardwareKey")
	systemIdTag, _      = tag.NewKey("systemId")
	productLimitTag, _  = tag.NewKey("productLimit")
	expire_days_measure = stats.Int64("expire_days", "Expired Days.", "Days")

	expireDaysView = &view.View{
		Name:        "sys_m_license/expire_days",
		Description: "License Expired Days.",
		TagKeys:     []tag.Key{hardwareKeyTag, systemIdTag, productLimitTag},
		Measure:     expire_days_measure,
		Aggregation: view.LastValue(),
	}
	LicenseCollectorViews = []*view.View{
		expireDaysView,
	}
)

// ScrapeserviceStatistics collects from `sys.m_service_statistics`.
type LicenseCollector struct{}

func (LicenseCollector) CollectorName() string {
	return "LicenseCollector"
}

func (LicenseCollector) Views() []*view.View {

	return LicenseCollectorViews
}

func (LicenseCollector) Scrape(ctx context.Context, db *sql.DB) {

	ctx, span := trace.StartSpan(ctx, "sql_query_sys_m_license")
	span.Annotate([]trace.Attribute{trace.StringAttribute("step", "excute_sql_query_sys_m_license")}, "excuting sql_query_sys_m_license sql to get the license info and exported as metrics")

	//get the sql data
	defer span.End()

	// if muliple row returned
	licenseRow := db.QueryRowContext(ctx, licenseStatusQuery)

	var hardware_key string
	var system_id string
	var product_limit string
	var expire_days int64

	sqlRowsScanCtx, sqlRowsScanSpan := trace.StartSpan(ctx, "sql_row_scan")
	sqlRowsScanSpan.Annotate([]trace.Attribute{trace.StringAttribute("step", "get the value from sql returns and update it to variables")}, "get the value from sql returns and update it to variables")
	err := licenseRow.Scan(&hardware_key, &system_id, &product_limit, &expire_days)
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
	licensectx, err := tag.New(measureSetCtx,
		tag.Insert(hardwareKeyTag, hardware_key),
		tag.Insert(systemIdTag, system_id),
		tag.Insert(productLimitTag, product_limit),
	)

	if err != nil {
		log.Fatalf("Failed to insert tag: %v", err)
		measureSetCtxSpan.Annotate([]trace.Attribute{}, err.Error())

	}

	stats.Record(licensectx, expire_days_measure.M(expire_days))

	measureSetCtxSpan.End()
}
