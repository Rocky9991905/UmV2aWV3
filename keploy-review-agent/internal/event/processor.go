package event

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/keploy/keploy-review-agent/internal/analyzer"
	"github.com/keploy/keploy-review-agent/internal/config"
)

type Processor struct {
	cfg        *config.Config
	orchestrator *analyzer.Orchestrator
}

func NewProcessor(cfg *config.Config) *Processor {
	return &Processor{
		cfg:        cfg,
		orchestrator: analyzer.NewOrchestrator(cfg),
	}
}

func (p *Processor) ProcessGitHubEvent(eventType string, payload []byte) error {
	// Parse GitHub event
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse GitHub event: %w", err)
	}
	// Extract PR info
	action, ok := event["action"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid action in GitHub event")
	}
	
	// Process only opened or synchronized PRs
	if action != "opened" && action != "synchronize" {
		return nil // Ignore other actions
	}
	
	// Extract PR details
	pr, ok := event["pull_request"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid pull_request in GitHub event")
	}
	
	// Extract repo info
	repo, ok := event["repository"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid repository in GitHub event")
	}
	
	// Create analysis job
	job := &analyzer.Job{
		Provider:    "github",
		RepoOwner:   repo["owner"].(map[string]interface{})["login"].(string),
		RepoName:    repo["name"].(string),
		PRNumber:    int(pr["number"].(float64)),
		HeadSHA:     pr["head"].(map[string]interface{})["sha"].(string),
		BaseSHA:     pr["base"].(map[string]interface{})["sha"].(string),
	}
	
	// Start analysis
	log.Printf("Starting analysis for %s/%s PR ", job.RepoOwner, job.RepoName)
	if issues, err := p.orchestrator.AnalyzeCode(job); err != nil {
		return fmt.Errorf("failed to analyze code: %w", err)
	} else {
		log.Printf("Analysis result: %v", issues)
	}
	
	return nil
}

func (p *Processor) ProcessGitLabEvent(eventType string, payload []byte) error {
	// TODO: Implement GitLab event processing
	return fmt.Errorf("GitLab event processing not implemented")
}
