package shared

import (
	"fmt"

	"github.com/keploy/keploy-review-agent/pkg/models"
)

// Global variable to store issues
var AllIssues []*models.Issue

// Function to get all issues
func GetAllIssues() []*models.Issue {
	return AllIssues
}

// Function to add an issue
func AddIssue(issue *models.Issue) error {
	fmt.Printf("Issue: %+v\n", issue)
	AllIssues = append(AllIssues, issue)
	return nil
}
