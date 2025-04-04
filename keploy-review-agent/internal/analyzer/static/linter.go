





























































































































package static

import (

	"bytes"
	"context"
	"encoding/json"


	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

		if err := ioutil.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}

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

	if len(tsFiles) > 0 {
		linterOutput, err := l.RunESLint(ctx, tempDir, tsFiles)
		if err != nil {
			log.Fatalf("Error running ESLint: %v", err)
		}

		issues = append(issues, processLinterOutput(linterOutput)...)
	}









	fmt.Printf("Total Issues Found: %d\n", len(issues))
	for i, issue := range issues {
		fmt.Printf("Issue %d: %+v\n", i+1, issue)
	}

	return issues, nil
}

func processLinterOutput(output string) []*models.Issue {
	var issues []*models.Issue

	output = strings.TrimSpace(output)
	if output == "" {
		fmt.Println("Linter output is empty")
		return nil
	}

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

	for _, result := range lintResults {
		for _, msg := range result.Messages {
			if msg.Message == "File ignored because no matching configuration was supplied." {

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

	if _, err := exec.LookPath("/snap/bin/golangci-lint"); err != nil {
		fmt.Println("golangci-lint not found:", err)
		return "", nil // Don't stop execution
	}

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

	cmdArgs := []string{
		"/snap/bin/golangci-lint", "run",
		"--config", configPath,
		"--output.json.path", "stdout",
	}
	cmdArgs = append(cmdArgs, files...)

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("golangci-lint encountered an error:", err)
	}

	return string(output), nil
}

func (l *Linter) RunESLint(ctx context.Context, dir string, files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files provided for ESLint")
	}

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

	packagePath := filepath.Join(dir, "package.json")
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		packageContent := "{\"type\": \"module\"}"
		if err := os.WriteFile(packagePath, []byte(packageContent), 0644); err != nil {
			return "", fmt.Errorf("failed to create package.json: %w", err)
		}
	}

	cmdCheck := exec.CommandContext(ctx, "npm", "list", "@eslint/js")
	cmdCheck.Dir = dir
	if err := cmdCheck.Run(); err != nil {

		cmdInstall := exec.CommandContext(ctx, "npm", "install", "--save-dev", "@eslint/js")
		cmdInstall.Dir = dir
		if err := cmdInstall.Run(); err != nil {
			return "", fmt.Errorf("failed to install ESLint dependencies: %w", err)
		}
	}

	args := append([]string{"--format", "json", "--config", configPath}, files...)
	cmd := exec.CommandContext(ctx, "npx", append([]string{"eslint"}, args...)...)
	cmd.Dir = dir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		fmt.Printf("ESLint encountered issues but continuing: %s\n", err)
	}

	var lintResults []map[string]interface{}
	if jsonErr := json.Unmarshal(out.Bytes(), &lintResults); jsonErr != nil {
		return "", fmt.Errorf("invalid ESLint JSON output: %w\nOutput: %s", jsonErr, out.String())
	}

	fmt.Printf("ESLint output: %s\n", out.String())
	return out.String(), nil
}