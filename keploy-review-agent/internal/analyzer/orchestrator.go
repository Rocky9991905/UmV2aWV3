package analyzer

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/keploy/keploy-review-agent/internal/analyzer/custom"
	"github.com/keploy/keploy-review-agent/internal/analyzer/dependency"
	"github.com/keploy/keploy-review-agent/internal/analyzer/llm"
	"github.com/keploy/keploy-review-agent/internal/analyzer/static"
	"github.com/keploy/keploy-review-agent/internal/config"
	"github.com/keploy/keploy-review-agent/internal/formatter"
	"github.com/keploy/keploy-review-agent/internal/reporter"
	"github.com/keploy/keploy-review-agent/internal/shared"
	"github.com/keploy/keploy-review-agent/pkg/github"
	"github.com/keploy/keploy-review-agent/pkg/models"
)

var pullnumber int
var AllIssues []*models.Issue

func PullRequestNumber(currentpullnumber int) int {
	pullnumber = currentpullnumber
	return pullnumber
}

// Job represents a code analysis job
type Job struct {
	Provider  string
	RepoOwner string
	RepoName  string
	PRNumber  int
}

// Orchestrator coordinates the different analysis engines
type Orchestrator struct {
	cfg            *config.Config
	staticAnalyzer *static.Linter
	depAnalyzer    *dependency.Scanner
	customAnalyzer *custom.Rules
	aiAnalyzer     *llm.GoogleAIClient
	githubClient   *github.Client
}

func NewOrchestrator(cfg *config.Config) *Orchestrator {
	aiConfig := &llm.AIConfig{
		MaxTokens:   cfg.AIMaxTokens,
		Temperature: cfg.AITemperature,
		MinSeverity: models.SeverityInfo,
	}

	return &Orchestrator{
		cfg:            cfg,
		staticAnalyzer: static.NewLinter(cfg),
		depAnalyzer:    dependency.NewScanner(cfg),
		customAnalyzer: custom.NewRules(cfg),
		githubClient:   github.NewClient(cfg.GitHubToken),
		aiAnalyzer:     llm.NewGoogleAIClient(cfg.GoogleAIKey, aiConfig),
	}
}

func (o *Orchestrator) AnalyzeCode(job *Job) ([]*models.Issue, error) {
	log.Printf("Starting analysis for %s/%s PR #%d", job.RepoOwner, job.RepoName, job.PRNumber)

	// Reset global issues array for this analysis
	AllIssues = []*models.Issue{}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(o.cfg.MaxProcessingTime)*time.Second,
	)
	defer cancel()

	// Fetch changed files
	files, err := o.fetchChangedFiles(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch changed files: %w", err)
	}
	log.Printf("Fetched %d changed files", len(files))

	// Create results channel and wait group
	resultsCh := make(chan *models.Issue)
	var wg sync.WaitGroup

	// Run analyzers in parallel
	// 1. Static Analysis
	if o.cfg.EnableStaticAnalysis {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.runAnalyzer("Static", func() ([]*models.Issue, error) {
				return o.staticAnalyzer.Analyze(ctx, files)
			}, resultsCh)
		}()
	}

	// 2. Dependency Analysis
	if o.cfg.EnableDependencyCheck {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.runAnalyzer("Dependency", func() ([]*models.Issue, error) {
				return o.depAnalyzer.Analyze(ctx, files)
			}, resultsCh)
		}()
	}

	// 3. AI Analysis
	if o.cfg.EnableAI {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.runAnalyzer("AI", func() ([]*models.Issue, error) {
				return o.aiAnalyzer.AnalyzeCode(ctx, files)
			}, resultsCh)
		}()
	}

	// 4. Custom Rules Analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.runAnalyzer("Custom", func() ([]*models.Issue, error) {
			return o.customAnalyzer.Analyze(ctx, files)
		}, resultsCh)
	}()

	// Close channel when all analyzers complete
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results
	for issue := range resultsCh {
		AllIssues = append(AllIssues, issue)
	}

	// Format and prepare comments
	comments := o.prepareComments(AllIssues)

	// Send review comments
	if err := o.sendReviewComment(ctx, job, comments); err != nil {
		log.Printf("Warning: Failed to send review comments: %v", err)
	}

	// Store issues in shared storage
	for _, issue := range AllIssues {
		if err := shared.AddIssue(issue); err != nil {
			log.Printf("Warning: Failed to add issue to shared storage: %v", err)
		}
	}

	log.Printf("Analysis completed for %s/%s PR #%d with %d issues",
		job.RepoOwner, job.RepoName, job.PRNumber, len(AllIssues))
	report := reporter.GenerateMarkdownReport(AllIssues)

	// Save to file
	if err := o.saveReport(report); err != nil {
		log.Printf("Failed to save report: %v", err)
	}
	fmt.Printf("GOLAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAASSSSSSSSSAAAAAAAAAAAVVVVVVVEEEEEE")
	return AllIssues, nil
}
func (o *Orchestrator) saveReport(report string) error {
	filename := "code-analysis-report.md"
	if o.cfg.ReportPath != "" {
		filename = o.cfg.ReportPath
	}
	return os.WriteFile(filename, []byte(report), 0644)
}

// Helper method to run analyzers and handle errors
func (o *Orchestrator) runAnalyzer(name string, analyzeFunc func() ([]*models.Issue, error), resultsCh chan<- *models.Issue) {
	issues, err := analyzeFunc()
	if err != nil {
		log.Printf("%s analysis failed: %v", name, err)
		return
	}

	log.Printf("%s analysis found %d issues", name, len(issues))
	for _, issue := range issues {
		resultsCh <- issue
	}
}

// Format issues into review comments
func (o *Orchestrator) prepareComments(issues []*models.Issue) []*models.ReviewComment {
	var comments []*models.ReviewComment

	for _, issue := range issues {
		comment := formatter.FormatLinterIssue(issue)
		// comment.CommitID = commitID
		comments = append(comments, comment)
	}

	return comments
}

// Fetch changed files from GitHub
func (o *Orchestrator) fetchChangedFiles(ctx context.Context, job *Job) ([]*models.File, error) {
	if job.Provider == "github" {
		return o.githubClient.GetChangedFiles(ctx, job.RepoOwner, job.RepoName, job.PRNumber)
	}
	return nil, fmt.Errorf("unsupported provider: %s", job.Provider)
}

// Send review comments to GitHub
func (o *Orchestrator) sendReviewComment(ctx context.Context, job *Job, comments []*models.ReviewComment) error {
	if job.Provider != "github" {
		return fmt.Errorf("unsupported provider: %s", job.Provider)
	}

	// If we have comments, send them as a review
	if len(comments) > 0 {
		if err := o.githubClient.CreateReview(ctx, job.RepoOwner, job.RepoName, job.PRNumber, comments); err != nil {
			return fmt.Errorf("failed to create review: %w", err)
		}
	}

	// Also process the pull request review (for backward compatibility)
	return o.githubClient.ProcessPullRequestReview(ctx, job.RepoOwner, job.RepoName, job.PRNumber)
}

// Add this to the existing postResults method in orchestrator.go
// func (o *Orchestrator) postResults(ctx context.Context, job *Job, issues []*models.Issue) error {
// 	// if len(issues) == 0 {
// 	// 	// No issues found, nothing to post
// 	// 	return nil
// 	// }
// 	log.Printf("issues are from postresults function are: %v", issues)
// 	// Convert issues to review comments
// 	var comments []*models.ReviewComment
// 	for _, issue := range issues {
// 		comment := &models.ReviewComment{
// 			Path: issue.Path,
// 			Line: issue.Line,
// 			Body: fmt.Sprintf("**%s**: %s\n\n%s",
// 				issue.Severity,
// 				issue.Title,
// 				issue.Description),
// 		}
// 		comments = append(comments, comment)
// 	}
// 	log.Printf("comments are from postresults function are: %v", comments)
// 	if job.Provider == "github" {
// 		return o.githubClient.CreateReview(ctx, job.RepoOwner, job.RepoName, job.PRNumber, comments)
// 	}

// 	return fmt.Errorf("unsupported provider: %s", job.Provider)
// }
