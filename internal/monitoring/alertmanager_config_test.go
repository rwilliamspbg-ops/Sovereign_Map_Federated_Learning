package monitoring

import (
	"os"
	"path/filepath"
	"testing"

	yaml "go.yaml.in/yaml/v2"
)

type alertmanagerConfig struct {
	Route struct {
		Receiver string   `yaml:"receiver"`
		GroupBy  []string `yaml:"group_by"`
		Routes   []struct {
			Receiver string   `yaml:"receiver"`
			Matchers []string `yaml:"matchers"`
		} `yaml:"routes"`
	} `yaml:"route"`
	Receivers []struct {
		Name string `yaml:"name"`
	} `yaml:"receivers"`
	InhibitRules []struct {
		SourceMatchers []string `yaml:"source_matchers"`
		TargetMatchers []string `yaml:"target_matchers"`
		Equal          []string `yaml:"equal"`
	} `yaml:"inhibit_rules"`
}

func TestAlertmanagerRoutingPolicy(t *testing.T) {
	cfg := readAlertmanagerConfig(t)

	if cfg.Route.Receiver != "default" {
		t.Fatalf("default receiver = %q, want default", cfg.Route.Receiver)
	}
	if !contains(cfg.Route.GroupBy, "service") || !contains(cfg.Route.GroupBy, "severity") {
		t.Fatalf("group_by must include service and severity, got %v", cfg.Route.GroupBy)
	}

	routeMap := make(map[string][]string)
	for _, route := range cfg.Route.Routes {
		routeMap[route.Receiver] = route.Matchers
	}

	mustRoute(t, routeMap, "consensus-critical", []string{"service=\"consensus\"", "severity=\"critical\""})
	mustRoute(t, routeMap, "consensus-warning", []string{"service=\"consensus\"", "severity=\"warning\""})
	mustRoute(t, routeMap, "fl-critical", []string{"service=\"federated-learning\"", "severity=\"critical\""})
	mustRoute(t, routeMap, "tokenomics-warning", []string{"service=\"tokenomics\"", "severity=\"warning\""})

	if hasMatch(routeMap["consensus-critical"], "severity=\"warning\"") {
		t.Fatalf("negative-path failed: consensus-critical route must not match warning severity")
	}

	receiverSet := make(map[string]struct{})
	for _, rec := range cfg.Receivers {
		receiverSet[rec.Name] = struct{}{}
	}
	for receiver := range routeMap {
		if _, ok := receiverSet[receiver]; !ok {
			t.Fatalf("route receiver %q is missing from receivers list", receiver)
		}
	}
}

func TestAlertmanagerInhibitionPolicy(t *testing.T) {
	cfg := readAlertmanagerConfig(t)

	foundConsensusInhibit := false
	foundCriticalSuppressWarning := false

	for _, rule := range cfg.InhibitRules {
		if hasMatch(rule.SourceMatchers, "alertname=\"ConsensusStatusEndpointDown\"") &&
			hasMatch(rule.SourceMatchers, "severity=\"critical\"") &&
			hasMatch(rule.TargetMatchers, "service=\"consensus\"") &&
			hasMatch(rule.TargetMatchers, "severity=\"warning\"") &&
			contains(rule.Equal, "service") {
			foundConsensusInhibit = true
		}

		if hasMatch(rule.SourceMatchers, "severity=\"critical\"") &&
			hasMatch(rule.TargetMatchers, "severity=\"warning\"") &&
			contains(rule.Equal, "service") {
			foundCriticalSuppressWarning = true
		}
	}

	if !foundConsensusInhibit {
		t.Fatal("missing inhibition rule: ConsensusStatusEndpointDown critical must inhibit consensus warning alerts")
	}
	if !foundCriticalSuppressWarning {
		t.Fatal("missing inhibition rule: critical alerts must suppress warning alerts within same service")
	}
}

func readAlertmanagerConfig(t *testing.T) alertmanagerConfig {
	t.Helper()

	path := filepath.Join("..", "..", "alertmanager.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read alertmanager.yml: %v", err)
	}

	var cfg alertmanagerConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse alertmanager.yml: %v", err)
	}
	return cfg
}

func mustRoute(t *testing.T, routeMap map[string][]string, receiver string, expectedMatchers []string) {
	t.Helper()

	matchers, ok := routeMap[receiver]
	if !ok {
		t.Fatalf("missing route for receiver %q", receiver)
	}
	for _, expected := range expectedMatchers {
		if !hasMatch(matchers, expected) {
			t.Fatalf("receiver %q missing matcher %q; got %v", receiver, expected, matchers)
		}
	}
}

func hasMatch(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
