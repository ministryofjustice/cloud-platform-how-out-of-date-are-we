package bin

import (
	"encoding/json"
	"fmt"
	"time"
)

// OrphanedResourcesReporter represents a reporter for orphaned AWS resources
type OrphanedResourcesReporter struct{}

// Run generates the orphaned AWS resources report
func (r *OrphanedResourcesReporter) Run() map[string]interface{} {
	// Placeholder for the actual logic to generate the report
	return map[string]interface{}{
		"resource1": "details1",
		"resource2": "details2",
	}
}

// Report represents the final report structure
type Report struct {
	OrphanedAWSResources map[string]interface{} `json:"orphaned_aws_resources"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

func main() {
	reporter := &OrphanedResourcesReporter{}
	orphanedResources := reporter.Run()

	report := Report{
		OrphanedAWSResources: orphanedResources,
		UpdatedAt:            time.Now(),
	}

	reportJSON, err := json.Marshal(report)
	if err != nil {
		fmt.Println("Error marshalling report to JSON:", err)
		return
	}

	fmt.Println(string(reportJSON))
}
