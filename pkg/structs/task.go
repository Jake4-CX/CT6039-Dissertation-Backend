package structs

type TaskAssignment struct {
	LoadTestTestsModel LoadTestTestsModel `json:"loadTest"` // Contains VirtualUsers, Duration, LoadTestType
	LoadTestPlanModel  LoadTestPlanModel  `json:"testPlan"` // Contains the test plan
	AssignedWorkerID   string             `json:"assignedWorkerId"`
}