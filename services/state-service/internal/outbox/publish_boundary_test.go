package outbox

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Correctness-track exemplar: domain twin.state.updated publish must stay in outbox
// (consumer/dlq.go is allowed for poison routing only).
func TestKafkaWriterOnlyInAllowedPackages(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	allowedSubstrings := []string{
		filepath.Join("internal", "outbox"),
		filepath.Join("internal", "consumer", "dlq.go"),
	}
	writerRe := regexp.MustCompile(`kafka\.Writer`)

	var violations []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !writerRe.Match(data) {
			return nil
		}
		norm := filepath.ToSlash(path)
		allowed := false
		for _, sub := range allowedSubstrings {
			if strings.Contains(norm, filepath.ToSlash(sub)) {
				allowed = true
				break
			}
		}
		if !allowed {
			violations = append(violations, norm)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("kafka.Writer outside allowed publish paths: %v", violations)
	}
}
