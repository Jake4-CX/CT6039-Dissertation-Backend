package initializers

import (
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitializeDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("db/load_testing.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate all models
	err = DB.AutoMigrate(
		&structs.LoadTestModel{},
		&structs.LoadTestPlanModel{},
		&structs.LoadTestTestsModel{},
		&structs.LoadTestMetricsModel{},
		&structs.LoadTestHistoryModel{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	log.Info("Database connected and auto-migrated")

	resetRunningTests()

}

func resetRunningTests() {
	result := DB.Model(&structs.LoadTestTestsModel{}).Where("state = ?", structs.Running).Update("state", structs.Cancelled)

	if result.RowsAffected > 0 {
		log.Infof("Reset %v running load tests to cancelled", result.RowsAffected)
	} else if result.Error != nil {
		log.Errorf("Failed to reset running load tests: %v", result.Error)
	}
}
