// package static

// import (
// 	"bytes"
// 	"context"

// 	"encoding/json"

// 	"fmt"

// 	"io/ioutil"

// 	"log"

// 	"os"

// 	"os/exec"

// 	"path/filepath"

// 	"strings"

// 	"github.com/keploy/keploy-review-agent/internal/config"
// 	"github.com/keploy/keploy-review-agent/pkg/models"
// )

// type Linter struct {
// 	cfg *config.Config
// }
// var Comment string
// func NewLinter(cfg *config.Config) *Linter {
// 	return &Linter{
// 		cfg: cfg,
// 	}
// }

// func (l *Linter) Analyze(ctx context.Context, files []*models.File) ([]*models.Issue, error) {
// // func (l *Linter) Analyze(ctx context.Context, files []*models.File) (string , error) {

// 	var issues []*models.Issue
// 	tempDir, err := ioutil.TempDir("", "keploy-review-")
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create temp directory: %w", err)
// 	}
// 	defer os.RemoveAll(tempDir)
// 	fmt.Printf("Dinf \n")
// 	goFiles := []string{}
// 	for _, file := range files {
// 		if !strings.HasSuffix(file.Path, ".go") {
// 			comment := "Add a Go code file to your PR"
// 			fmt.Printf("Also  %s\n", comment)
// 			Comment = comment
// 			continue
// 		}
// 		log.Printf("mein static analyser mein hoon linter.go mein Analyze mein")
// 		filePath := filepath.Join(tempDir, filepath.Base(file.Path))
// 		if err := ioutil.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
// 			return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
// 		}
// 		goFiles = append(goFiles, filePath)
// 		fmt.Printf("goFiles: %s\n", goFiles)
// 	}

// 	if len(goFiles) == 0 {
// 		log.Printf("No Go files to analyze")
// 		return issues, nil
// 	}
// 	linterIssues, err := l.runGolangCILint(ctx, tempDir, goFiles)
// 	fmt.Printf("linterIssues: are ghhhhh %v\n", linterIssues)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to run golangci-lint: %w", err)
// 	}

// 	return issues,nil
// 	// return "ghoda hoon mein", nil
// }

// func (l *Linter) runGolangCILint(ctx context.Context, dir string, files []string) ([]*models.Issue, error) {
// 	var issues []*models.Issue

// 	if _, err := exec.LookPath("golangci-lint"); err != nil {
// 		return nil, fmt.Errorf("golangci-lint not found: %w", err)
// 	}

// 	configPath := filepath.Join(dir, ".golangci.yml")
// 	configContent := `
// linters:
//   enable:
//     - errcheck
//     - gosimple
//     - govet
//     - ineffassign
//     - staticcheck
//     - unused
//     - misspell
// issues:
//   max-issues-per-linter: 0
//   max-same-issues: 0
// `
// 	if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
// 		return nil, fmt.Errorf("failed to write golangci-lint config: %w", err)
// 	}

// 	args := []string{"run", "--config", configPath, "--out-format", "json"}
// 	args = append(args, files...)

// 	cmd := exec.CommandContext(ctx, "golangci-lint", args...)
// 	cmd.Dir = dir

// 	var stdout, stderr bytes.Buffer
// 	cmd.Stdout = &stdout
// 	cmd.Stderr = &stderr

// 	err := cmd.Run()
// 	if err != nil && stdout.Len() == 0 {
// 		return nil, fmt.Errorf("golangci-lint failed: %s", stderr.String())
// 	}

// 	if stdout.Len() > 0 {
// 		var result struct {
// 			Issues []struct {
// 				FromLinter  string `json:"from_linter"`
// 				Text        string `json:"text"`
// 				Pos         struct {
// 					Filename string `json:"filename"`
// 					Line     int    `json:"line"`
// 					Column   int    `json:"column"`
// 				} `json:"pos"`
// 			} `json:"Issues"`
// 		}

// 		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
// 			return nil, fmt.Errorf("failed to parse golangci-lint output: %w", err)
// 		}

// 		for _, issue := range result.Issues {
// 			issues = append(issues, &models.Issue{
// 				Path:        issue.Pos.Filename,
// 				Line:        issue.Pos.Line,
// 				Column:      issue.Pos.Column,
// 				Severity:    models.SeverityWarning,
// 				Title:       fmt.Sprintf("[%s] Code issue", issue.FromLinter),
// 				Description: issue.Text,
// 				Source:      "golangci-lint",
// 			})
// 		}
// 	}
// 	log.Printf("issues are from runGolangCILint function are: %v", issues)
// 	return issues, nil
// }

package static

import (
	// "bytes"
	"bytes"
	"context"
	"encoding/json"

	// "io"

	// "encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	// "github.com/golangci/golangci-lint/pkg/lint/linter"
	"github.com/keploy/keploy-review-agent/internal/config"
	"github.com/keploy/keploy-review-agent/pkg/models"
)

type Linter struct {
	cfg *config.Config
}

var Comment string

func NewLinter(cfg *config.Config) *Linter {
	return &Linter{
		cfg: cfg,
	}
}

func (l *Linter) Analyze(ctx context.Context, files []*models.File) ([]*models.Issue, error) {
	var issues []*models.Issue

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "keploy-review-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	var goFiles, tsFiles []string
	hasGo, hasTS := false, false

	for _, file := range files {
		filePath := filepath.Join(tempDir, filepath.Base(file.Path))

		switch {
		case strings.HasSuffix(file.Path, ".go"):
			hasGo = true
			goFiles = append(goFiles, filePath)
		case strings.HasSuffix(file.Path, ".ts"):
			hasTS = true
			tsFiles = append(tsFiles, filePath)
		default:
			continue // Ignore non-Go/TS files
		}

		// Write file to temp directory
		if err := ioutil.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}

	// Handle missing file cases
	if !hasGo && !hasTS {
		log.Println("No Go or TypeScript files detected in PR")
		Comment = "Add either a Go or TypeScript file to your PR"
		return issues, nil
	}
	if !hasGo {
		log.Println("No Go files detected in PR")
		Comment = "Add a Go code file to your PR"
	}
	if !hasTS {
		log.Println("No TypeScript files detected in PR")
		Comment = "Add a TypeScript code file to your PR"
	}

	// Run ESLint for TypeScript files if available
	if len(tsFiles) > 0 {
		linterOutput, err := l.RunESLint(ctx, tempDir, tsFiles)
		if err != nil {
			log.Fatalf("Error running ESLint: %v", err)
		}

		// Process linter output (TS files)
		issues = append(issues, processLinterOutput(linterOutput)...)
	}

	// Uncomment this if you want to process Go files as well
	// if len(goFiles) > 0 {
	// 	linterOutput, err := l.runGolangCILint(ctx, tempDir, goFiles)
	// 	if err != nil {
	// 		log.Fatalf("Error running Go linter: %v", err)
	// 	}
	// 	issues = append(issues, processLinterOutput(linterOutput)...)
	// }

	fmt.Printf("Total Issues Found: %d\n", len(issues))
	for i, issue := range issues {
		fmt.Printf("Issue %d: %+v\n", i+1, issue)
	}

	return issues, nil
}

func processLinterOutput(output string) []*models.Issue {
	var issues []*models.Issue

	// Trim spaces and check if output is empty
	output = strings.TrimSpace(output)
	if output == "" {
		fmt.Println("Linter output is empty")
		return nil
	}

	// Parse JSON as an array
	var lintResults []struct {
		FilePath string `json:"filePath"`
		Messages []struct {
			RuleID   string `json:"ruleId"`
			Severity int    `json:"severity"`
			Message  string `json:"message"`
			Line     int    `json:"line"`
			Column   int    `json:"column"`
		} `json:"messages"`
	}

	if err := json.Unmarshal([]byte(output), &lintResults); err != nil {
		fmt.Printf("Error parsing linter output: %v\nOutput: %s\n", err, output)
		return nil
	}

	// Convert ESLint issues to models.Issue format
	for _, result := range lintResults {
		for _, msg := range result.Messages {
			if msg.Message == "File ignored because no matching configuration was supplied." {
				// Skip ignored files
				continue
			}

			issue := &models.Issue{
				Path:        result.FilePath,
				Line:        msg.Line,
				Column:      msg.Column,
				Severity:    models.Severity(severityToString(msg.Severity)),
				Title:       fmt.Sprintf("ESLint Issue: %s", msg.RuleID),
				Description: msg.Message,
				Suggestion:  "Consider fixing this issue based on the linter's feedback.",
				Source:      "ESLint",
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// Converts ESLint severity to string
func severityToString(severity int) string {
	switch severity {
	case 1:
		return "warning"
	case 2:
		return "error"
	default:
		return "info"
	}
}


func (l *Linter) runGolangCILint(ctx context.Context, dir string, files []string) (string, error) {
	// Ensure golangci-lint is installed
	if _, err := exec.LookPath("/snap/bin/golangci-lint"); err != nil {
		fmt.Println("golangci-lint not found:", err)
		return "", nil // Don't stop execution
	}

	// Write .golangci.yml configuration file
	configPath := filepath.Join(dir, ".golangci.yml")
	configContent := `
version: 2 
linters:
  enable:
    - govet
    - staticcheck
    - errcheck
    - ineffassign
    - unused
    - misspell
    - gocritic
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		fmt.Println("Failed to write golangci-lint config:", err)
		return "", nil // Don't stop execution
	}

	// Build the command dynamically
	cmdArgs := []string{
		"/snap/bin/golangci-lint", "run",
		"--config", configPath,
		"--output.json.path", "stdout",
	}
	cmdArgs = append(cmdArgs, files...)

	// Run the linter and capture output/errors
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.CombinedOutput()

	// Print errors but DON'T return them (to avoid stopping execution)
	if err != nil {
		fmt.Println("golangci-lint encountered an error:", err)
	}

	return string(output), nil
}

func (l *Linter) RunESLint(ctx context.Context, dir string, files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files provided for ESLint")
	}

	// Ensure ESLint config file exists in the working directory
	configPath := filepath.Join(dir, "eslint.config.js")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configContent :=
		`
		export default [
  {
    ignores: [],
    files: ["**/*.ts", "**/*.tsx", "**/*.js", "**/*.jsx"],
    languageOptions: {
      parserOptions: {
        ecmaVersion: "latest",
        sourceType: "module",
      },
    },
    rules: {
      "no-unused-vars": "warn",
      "no-console": "warn",
    },
  },
];
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			return "", fmt.Errorf("failed to create ESLint config file: %w", err)
		}
	}

	// Ensure package.json exists with module type
	packagePath := filepath.Join(dir, "package.json")
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		packageContent := "{\"type\": \"module\"}"
		if err := os.WriteFile(packagePath, []byte(packageContent), 0644); err != nil {
			return "", fmt.Errorf("failed to create package.json: %w", err)
		}
	}

	// Check if @eslint/js is installed
	cmdCheck := exec.CommandContext(ctx, "npm", "list", "@eslint/js")
	cmdCheck.Dir = dir
	if err := cmdCheck.Run(); err != nil {
		// Install @eslint/js if not found
		cmdInstall := exec.CommandContext(ctx, "npm", "install", "--save-dev", "@eslint/js")
		cmdInstall.Dir = dir
		if err := cmdInstall.Run(); err != nil {
			return "", fmt.Errorf("failed to install ESLint dependencies: %w", err)
		}
	}

	// Run ESLint
	args := append([]string{"--format", "json", "--config", configPath}, files...)
	cmd := exec.CommandContext(ctx, "npx", append([]string{"eslint"}, args...)...)
	cmd.Dir = dir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		fmt.Printf("ESLint encountered issues but continuing: %s\n", err)
	}

	// Ensure valid JSON output
	var lintResults []map[string]interface{}
	if jsonErr := json.Unmarshal(out.Bytes(), &lintResults); jsonErr != nil {
		return "", fmt.Errorf("invalid ESLint JSON output: %w\nOutput: %s", jsonErr, out.String())
	}

	fmt.Printf("ESLint output: %s\n", out.String())
	return out.String(), nil
}