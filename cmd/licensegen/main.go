package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"haridy2026/configs"
	"haridy2026/internal/database"
	"haridy2026/internal/services"
)

func main() {
	plan := flag.String("plan", "yearly", "license plan code")
	maxOperations := flag.Int64("max-operations", 100000, "maximum allowed operations")
	expires := flag.String("expires", "", "optional expiry date in YYYY-MM-DD")
	durationDays := flag.Int("duration-days", 0, "license duration from today in days")
	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Println("  go run ./cmd/licensegen --plan yearly --max-operations 100000 --duration-days 365")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  monthly:  --duration-days 30 --max-operations 5000")
		fmt.Println("  yearly:   --duration-days 365 --max-operations 100000")
		fmt.Println("  lifetime: --duration-days 36500 --max-operations 999999999")
		fmt.Println()
		flag.PrintDefaults()
	}
	flag.Parse()

	var expiresAt *time.Time
	if *durationDays > 0 {
		expiresOn := time.Now().AddDate(0, 0, *durationDays)
		expiresAt = &expiresOn
	} else if *expires != "" {
		parsed, err := time.Parse("2006-01-02", *expires)
		if err != nil {
			log.Fatalf("invalid --expires: %v", err)
		}
		expiresAt = &parsed
	}

	cfg := configs.Load()
	db, err := configs.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	code, license, err := services.NewActivationService(db).CreateLicense(*plan, *maxOperations, expiresAt)
	if err != nil {
		log.Fatalf("create license: %v", err)
	}

	fmt.Printf("Activation code: %s\n", code)
	fmt.Printf("License ID: %d\n", license.ID)
	fmt.Printf("Plan: %s\n", license.PlanCode)
	fmt.Printf("Max operations: %d\n", license.MaxOperations)
	if license.ExpiresAt != nil {
		fmt.Printf("Expires at: %s\n", license.ExpiresAt.Format("2006-01-02"))
	}
}
