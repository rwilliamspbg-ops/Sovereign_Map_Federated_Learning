// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package hva

import (
	"testing"
)

func TestMinimumHonestNodes(t *testing.T) {
	cases := []struct{ n, want int }{
		{0, 0}, {1, 1}, {9, 5}, {18, 10}, {200, 112}, {-5, 0},
	}
	for _, tc := range cases {
		got := MinimumHonestNodes(tc.n)
		if got != tc.want {
			t.Errorf("MinimumHonestNodes(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestMaximumByzantineNodes(t *testing.T) {
	cases := []struct{ n, want int }{
		{0, 0}, {1, 0}, {9, 4}, {18, 8}, {200, 88}, {-3, 0},
	}
	for _, tc := range cases {
		got := MaximumByzantineNodes(tc.n)
		if got != tc.want {
			t.Errorf("MaximumByzantineNodes(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestByzantinePlusHonestEqualsTotal(t *testing.T) {
	for _, n := range []int{1, 9, 10, 17, 18, 100, 200, 1000} {
		h := MinimumHonestNodes(n)
		b := MaximumByzantineNodes(n)
		if h+b != n {
			t.Errorf("n=%d: honest(%d)+byzantine(%d) != n", n, h, b)
		}
	}
}

func TestHonestNodesMajority(t *testing.T) {
	for _, n := range []int{1, 3, 9, 100, 1000} {
		h := MinimumHonestNodes(n)
		if 2*h <= n {
			t.Errorf("n=%d: honest=%d is not a majority", n, h)
		}
	}
}

func TestBuildPlanBasic(t *testing.T) {
	plan, err := BuildPlan(100, 2)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.TotalNodes != 100 {
		t.Errorf("TotalNodes = %d, want 100", plan.TotalNodes)
	}
	if plan.Dimensions != 2 {
		t.Errorf("Dimensions = %d, want 2", plan.Dimensions)
	}
	if len(plan.Levels) == 0 {
		t.Fatal("plan has no levels")
	}
	root := plan.Levels[len(plan.Levels)-1]
	if root.NodeCount != 1 {
		t.Errorf("root NodeCount = %d, want 1", root.NodeCount)
	}
	if plan.Levels[0].NodeCount != 100 {
		t.Errorf("edge NodeCount = %d, want 100", plan.Levels[0].NodeCount)
	}
}

func TestBuildPlanSingleNode(t *testing.T) {
	plan, err := BuildPlan(1, 1)
	if err != nil {
		t.Fatalf("BuildPlan(1,1): %v", err)
	}
	if len(plan.Levels) != 1 {
		t.Errorf("levels = %d, want 1", len(plan.Levels))
	}
}

func TestBuildPlanInvalidInputs(t *testing.T) {
	if _, err := BuildPlan(0, 1); err == nil {
		t.Error("expected error for totalNodes=0")
	}
	if _, err := BuildPlan(10, 0); err == nil {
		t.Error("expected error for dimensions=0")
	}
	if _, err := BuildPlan(-1, 2); err == nil {
		t.Error("expected error for negative totalNodes")
	}
}

func TestPlanValidateGoodPlan(t *testing.T) {
	plan, err := BuildPlan(50, 3)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if err := plan.Validate(); err != nil {
		t.Errorf("Validate on valid plan: %v", err)
	}
}

func TestPlanValidateRejectsBadPlan(t *testing.T) {
	bad := Plan{TotalNodes: 0, Dimensions: 0, Levels: nil}
	if err := bad.Validate(); err == nil {
		t.Error("expected Validate to reject zero-value Plan")
	}
}

func TestTierNameEdge(t *testing.T) {
	if got := tierName(0, 100); got != "edge" {
		t.Errorf("tierName(0,100) = %q, want edge", got)
	}
}

func TestTierNameGlobal(t *testing.T) {
	if got := tierName(5, 1); got != "global" {
		t.Errorf("tierName(5,1) = %q, want global", got)
	}
}
