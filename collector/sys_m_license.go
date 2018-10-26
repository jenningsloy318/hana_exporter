// Scrape `sys_m_license`.

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

func (licenseCollector LicenseCollector)CollectorName() string {
	return "LicenseCollector"
}

func (licenseCollector LicenseCollector)NewViews() []*view.View {
	return LicenseCollectorViews
}

func (licenseCollector LicenseCollector)UpdateMeasurements(db *sql.DB) {

	licenseRow := db.QueryRow(licenseStatusQuery)

	var hardware_key string
	var system_id string
	var product_limit string
	var expire_days int64

	err := licenseRow.Scan(&hardware_key, &system_id, &product_limit, &expire_days)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No value returend")

	case err != nil:
		log.Fatal(err)

	}

	licensectx, err := tag.New(context.Background(),
		tag.Insert(hardwareKeyTag, hardware_key),
		tag.Insert(systemIdTag, system_id),
		tag.Insert(productLimitTag, product_limit),
	)

	if err != nil {
		log.Fatalf("Failed to insert tag: %v", err)

	}

	stats.Record(licensectx, expire_days_measure.M(expire_days))

}
