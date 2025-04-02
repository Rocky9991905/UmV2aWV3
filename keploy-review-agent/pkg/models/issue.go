package models

// Severity represents the severity level of an issue
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)
// type Dependency struct {
//     System  string
//     Name    string
//     Version string
// }
// Issue represents a code issue found during analysis
type Issue struct {
	Path        string   // File path
	Line        int      // Line number
	Column      int      // Column number (optional)
	Severity    Severity // Issue severity
	Title       string   // Short issue title
	Description string   // Detailed description
	Suggestion  string   // Suggested fix (optional)
	Source      string   // Source of the issue (e.g., "golangci-lint", "llm")
}

type AffectedVersion struct {
    Introduced string
    Fixed      string
}
// type Vulnerability struct {
//     Title            string
//     Severity         string
//     Dependency       Dependency
//     CVEs             []string
//     AffectedVersions []AffectedVersion
//     AdvisoryID       string
// }
// File represents a source code file
type File struct {
	Path    string // File path
	Content string // File content
}

// ReviewComment represents a comment to be posted on a code review
type ReviewComment struct {
	Path      string // File path
	Line      int    // Line number
	Body      string // Comment body
	CommitID  string // Commit ID
	Position  int    // Position in the diff
}
