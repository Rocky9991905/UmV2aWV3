package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/keploy/keploy-review-agent/internal/config"
	"github.com/keploy/keploy-review-agent/pkg/models"
)

// Engine implements the LLM-based code analysis
type Engine struct {
	cfg *config.Config
}
var Comment string
// NewEngine creates a new LLM analysis engine
func NewEngine(cfg *config.Config) *Engine {
	return &Engine{
		cfg: cfg,
	}
}

// Analyze runs LLM analysis on the provided files
func (e *Engine) Analyze(ctx context.Context, files []*models.File) ([]*models.Issue, error) {
	var issues []*models.Issue
	
	// Skip if LLM is disabled
	if !e.cfg.EnableLLM {
		return issues, nil
	}
	
	// Process each file
	for _, file := range files {
		// Skip files that are too large
		if len(file.Content) > int(e.cfg.MaxFileSizeBytes) {
			continue
		}
		
		// Skip non-code files
		if !isCodeFile(file.Path) {
			comment := "Add a code file to your PR."
			// fmt.Printf("Comment: in not a isCodefile non-uuu%s\n", comment)
			// fmt.Printf("Comment: in not a isCodefile %s\n", comment)
			if comment != "" {
				Comment = comment
				fmt.Printf("comment in isCodefile is non-empty")
			} else {
				Comment = ""
				fmt.Printf("comment in isCodefile in empty")
			}
			continue
		}
		
		// TODO: Implement LLM analysis
		// 1. Prepare prompt with file content
		// 2. Send to LLM API
		// 3. Parse response into issues
		
		// Example placeholder
		issues = append(issues, &models.Issue{
			Path:        file.Path,
			Line:        1,
			Severity:    models.SeverityInfo,
			Title:       "LLM analysis placeholder",
			Description: "This is a placeholder for LLM-based code analysis",
			Source:      "llm-engine",
		})
	}
	
	return issues, nil
}

// isCodeFile determines if a file is a code file based on extension
func isCodeFile(path string) bool {
	codeExtensions := []string{".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h"}
	for _, ext := range codeExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
