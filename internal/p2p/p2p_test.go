package p2p

import (
	"context"
	"testing"
	"time"
)

func TestVerifierHappyPath(t *testing.T) {
	v := NewVerifier("node-main", 2, time.Second)

	if err := v.RegisterPeer(&PeerDetail{ID: "peer-a"}); err != nil {
		t.Fatalf("register peer-a: %v", err)
	}
	if err := v.RegisterPeer(&PeerDetail{ID: "peer-b"}); err != nil {
		t.Fatalf("register peer-b: %v", err)
	}

	requestID, err := v.RequestVerification(context.Background(), &ModelVerificationRequest{
		ModelWeights: []byte("weights"),
		Proof:        []byte("proof"),
		ProposerID:   "node-main",
		Round:        1,
		Timestamp:    time.Now(),
	})
	if err != nil {
		t.Fatalf("request verification: %v", err)
	}

	err = v.SubmitVerification(context.Background(), &ModelVerificationResponse{
		RequestID:  requestID,
		VerifierID: "peer-a",
		Valid:      true,
		Timestamp:  time.Now(),
	})
	if err != nil {
		t.Fatalf("submit verification peer-a: %v", err)
	}

	err = v.SubmitVerification(context.Background(), &ModelVerificationResponse{
		RequestID:  requestID,
		VerifierID: "peer-b",
		Valid:      true,
		Timestamp:  time.Now(),
	})
	if err != nil {
		t.Fatalf("submit verification peer-b: %v", err)
	}

	complete, confidence, err := v.CheckVerificationStatus(requestID)
	if err != nil {
		t.Fatalf("check verification status: %v", err)
	}
	if !complete {
		t.Fatal("expected verification to complete")
	}
	if confidence <= 0.66 {
		t.Fatalf("expected confidence > 0.66, got %f", confidence)
	}
}

func TestVerifierRejectsUnknownPeer(t *testing.T) {
	v := NewVerifier("node-main", 1, time.Second)

	requestID, err := v.RequestVerification(context.Background(), &ModelVerificationRequest{
		ProposerID: "node-main",
		Round:      1,
		Timestamp:  time.Now(),
	})
	if err != nil {
		t.Fatalf("request verification: %v", err)
	}

	err = v.SubmitVerification(context.Background(), &ModelVerificationResponse{
		RequestID:  requestID,
		VerifierID: "missing-peer",
		Valid:      true,
		Timestamp:  time.Now(),
	})
	if err == nil {
		t.Fatal("expected unknown verifier error")
	}
}

func TestVerificationProtocolRequestLifecycle(t *testing.T) {
	vp := NewVerificationProtocol("node-main", 2, time.Second)
	_ = vp.RegisterPeer("peer-1")
	_ = vp.RegisterPeer("peer-2")

	requestID, err := vp.RequestVerification(context.Background(), []byte("data"), []byte("sig"))
	if err != nil {
		t.Fatalf("request verification: %v", err)
	}

	err = vp.SubmitVerificationResponse(context.Background(), &VerificationResponse{
		RequestID:  requestID,
		Valid:      true,
		VerifierID: "peer-1",
		Confidence: 0.9,
		VerifiedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("submit response peer-1: %v", err)
	}

	err = vp.SubmitVerificationResponse(context.Background(), &VerificationResponse{
		RequestID:  requestID,
		Valid:      true,
		VerifierID: "peer-2",
		Confidence: 0.8,
		VerifiedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("submit response peer-2: %v", err)
	}

	complete, confidence, err := vp.CheckVerificationStatus(requestID)
	if err != nil {
		t.Fatalf("check verification status: %v", err)
	}
	if !complete {
		t.Fatal("expected verification protocol consensus to complete")
	}
	if confidence <= 0 {
		t.Fatalf("expected positive confidence, got %f", confidence)
	}
}
