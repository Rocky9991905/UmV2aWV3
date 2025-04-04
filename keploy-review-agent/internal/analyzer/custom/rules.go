package custom

import (
	"context"
	
	"github.com/keploy/keploy-review-agent/internal/config"
	"github.com/keploy/keploy-review-agent/pkg/models"
)

type Rules struct {
	cfg *config.Config
}

func NewRules(cfg *config.Config) *Rules {
	return &Rules{
		cfg: cfg,
	}
}

func (r *Rules) Analyze(ctx context.Context, files []*models.File) ([]*models.Issue, error) {
	var issues []*models.Issue


	
	return issues, nil
}
