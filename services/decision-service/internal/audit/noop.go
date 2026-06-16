package audit

import (
	"context"

	"github.com/digital-twin/platform/services/decision-service/internal/decision"
)

type NoopPublisher struct{}

func (NoopPublisher) PublishRuleDecision(_ context.Context, _ PublishInput) error {
	return nil
}

func (NoopPublisher) Close() error { return nil }

type Publisher interface {
	PublishRuleDecision(ctx context.Context, in PublishInput) error
	Close() error
}

var _ Publisher = NoopPublisher{}

type pendingPublisher interface {
	PublishRuleDecision(ctx context.Context, in PublishInput) error
}

func RequiresAuditOutcome(d decision.RuleDecision) bool {
	return decision.RequiresAudit(d.Outcome)
}
