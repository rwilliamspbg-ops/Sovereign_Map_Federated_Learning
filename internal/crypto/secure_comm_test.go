// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"
)

func TestNewSecureChannel(t *testing.T) {
	sc, err := NewSecureChannel()
	if err != nil {
		t.Fatalf("NewSecureChannel: %v", err)
	}
	if sc.privateKey == nil || sc.publicKey == nil {
		t.Fatal("keys are nil after creation")
	}
}

func TestExportImportPublicKey(t *testing.T) {
	sc, _ := NewSecureChannel()
	pemBytes, err := sc.ExportPublicKey()
	if err != nil {
		t.Fatalf("ExportPublicKey: %v", err)
	}
	pub, err := ImportPublicKey(pemBytes)
	if err != nil {
		t.Fatalf("ImportPublicKey: %v", err)
	}
	if pub.X.Cmp(sc.publicKey.X) != 0 || pub.Y.Cmp(sc.publicKey.Y) != 0 {
		t.Fatal("imported public key does not match original")
	}
}

func TestSignAndVerify(t *testing.T) {
	sc, _ := NewSecureChannel()
	_ = sc.RegisterPeer("self", sc.publicKey)
	data := []byte("sovereign-map test payload 1234")
	sig, err := sc.SignData(data)
	if err != nil {
		t.Fatalf("SignData: %v", err)
	}
	if err := sc.VerifySignature("self", data, sig); err != nil {
		t.Fatalf("VerifySignature: %v", err)
	}
}

func TestSignRejectsTampered(t *testing.T) {
	sc, _ := NewSecureChannel()
	_ = sc.RegisterPeer("self", sc.publicKey)
	sig, _ := sc.SignData([]byte("authentic message"))
	if err := sc.VerifySignature("self", []byte("tampered!"), sig); err == nil {
		t.Fatal("expected failure on tampered data")
	}
}

func TestRegisterPeerRejectsNil(t *testing.T) {
	sc, _ := NewSecureChannel()
	if err := sc.RegisterPeer("ghost", (*ecdsa.PublicKey)(nil)); err == nil {
		t.Fatal("expected error for nil public key")
	}
}

func TestGetTLSConfig(t *testing.T) {
	sc, _ := NewSecureChannel()
	if sc.GetTLSConfig() == nil {
		t.Fatal("GetTLSConfig returned nil")
	}
}

// TestTLSConfigMinVersion validates TLS 1.3 enforcement and no stray cipher suite list.
func TestTLSConfigMinVersion(t *testing.T) {
	cfg := createTLSConfig()
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Fatalf("expected MinVersion TLS 1.3 (0x%04x), got 0x%04x", tls.VersionTLS13, cfg.MinVersion)
	}
	if len(cfg.CipherSuites) != 0 {
		t.Fatalf("CipherSuites must be empty for TLS 1.3-only config; got %v", cfg.CipherSuites)
	}
}

// TestRotateSessionKeyNoDeadlock verifies RotateSessionKey completes without deadlock.
func TestRotateSessionKeyNoDeadlock(t *testing.T) {
	alice, _ := NewSecureChannel()
	bob, _ := NewSecureChannel()
	_ = alice.RegisterPeer("bob", bob.publicKey)

	if _, err := alice.EncryptMessage("bob", []byte("ping")); err != nil {
		t.Fatalf("initial encrypt: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- alice.RotateSessionKey("bob") }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RotateSessionKey: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RotateSessionKey timed out -- possible deadlock")
	}
}

// TestHandshakeVerification performs an in-memory TLS 1.3 handshake and
// asserts the negotiated version is TLS 1.3.
func TestHandshakeVerification(t *testing.T) {
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate server key: %v", err)
	}
	cert, err := makeSelfSignedTLSCert(serverKey)
	if err != nil {
		t.Fatalf("self-signed cert: %v", err)
	}

	serverTLS := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
	}
	clientTLS := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true,
	}

	serverRaw, clientRaw := net.Pipe()
	serverConn := tls.Server(serverRaw, serverTLS)
	clientConn := tls.Client(clientRaw, clientTLS)
	defer func() { _ = serverConn.Close() }()
	defer func() { _ = clientConn.Close() }()

	errCh := make(chan error, 2)
	go func() { errCh <- serverConn.Handshake() }()
	go func() { errCh <- clientConn.Handshake() }()

	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("TLS handshake error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("TLS handshake timed out")
		}
	}

	state := clientConn.ConnectionState()
	if state.Version != tls.VersionTLS13 {
		t.Fatalf("expected TLS 1.3 (0x%04x), negotiated 0x%04x", tls.VersionTLS13, state.Version)
	}
}

// makeSelfSignedTLSCert builds a minimal self-signed ECDSA certificate for tests.
func makeSelfSignedTLSCert(key *ecdsa.PrivateKey) (tls.Certificate, error) {
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return tls.X509KeyPair(certPEM, keyPEM)
}
