package structs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormModel struct {
	ID        uint           `gorm:"primarykey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
} //@name models.gormModel

type LoadTestState string
type LoadTestType string

const (
	Running   LoadTestState = "RUNNING"
	Cancelled LoadTestState = "CANCELLED"
	Completed LoadTestState = "COMPLETED"
)

const (
	Load   LoadTestType = "LOAD"
	Stress LoadTestType = "STRESS"
	Spike  LoadTestType = "SPIKE"
	Soak   LoadTestType = "SOAK"
)

type LoadTestModel struct {
	GormModel
	UUID      string               `json:"uuid"`
	Name      string               `json:"name"`
	TestPlan  LoadTestPlanModel    `gorm:"foreignKey:LoadTestModelId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"testPlan"`
	LoadTests []LoadTestTestsModel `gorm:"foreignKey:LoadTestModelId" json:"loadTests"` // 1 to many relationship - Of executed tests
}

type LoadTestPlanModel struct {
	GormModel
	LoadTestModelId uint   `json:"loadTestModelId"` // Foreign key
	ReactFlowPlan   string `gorm:"type:text" json:"reactFlowPlan"`
	TestPlan        string `gorm:"type:text" json:"testPlan"`
}

type LoadTestTestsModel struct {
	GormModel
	LoadTestModelId uint                 `json:"loadTestModelId"`          // Foreign key
	State           LoadTestState        `json:"state"`                    // Load test state (RUNNING, CANCELLED, COMPLETED)
	Duration        int                  `gorm:"type:int" json:"duration"` // In milliseconds
	VirtualUsers    int                  `gorm:"type:int" json:"virtualUsers"`
	LoadTestType    LoadTestType         `json:"loadTestType"`
	TestMetrics     LoadTestMetricsModel `json:"testMetrics" gorm:"foreignKey:LoadTestTestsModelID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // 1 to 1 relationship
}

type LoadTestMetricsModel struct {
	GormModel
	LoadTestTestsModelID uint  `json:"loadTestTestsModelId"` // Foreign key
	TotalRequests        int   `json:"totalRequests"`
	SuccessfulRequests   int   `json:"successfulRequests"`
	FailedRequests       int   `json:"failedRequests"`
	TotalResponseTime    int64 `json:"totalResponseTime"`
	AverageResponseTime  int64 `json:"averageResponseTime"`
}

func (loadTestModel *LoadTestModel) BeforeCreate(tx *gorm.DB) (err error) {
	loadTestModel.UUID = uuid.NewString()
	return
}
