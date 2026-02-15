// Package crypto implements secure communication layer for federated learning
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// SecureChannel manages encrypted peer-to-peer communication
type SecureChannel struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	peerKeys   map[string]*ecdsa.PublicKey
	sessionKeys map[string][]byte
	mu         sync.RWMutex
	tlsConfig  *tls.Config
}

// NewSecureChannel creates a new secure communication channel
func NewSecureChannel() (*SecureChannel, error) {
	// Generate ECDSA key pair using P-256 curve
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &SecureChannel{
		privateKey:  privateKey,
		publicKey:   &privateKey.PublicKey,
		peerKeys:    make(map[string]*ecdsa.PublicKey),
		sessionKeys: make(map[string][]byte),
		tlsConfig:   createTLSConfig(),
	}, nil
}

// RegisterPeer registers a peer's public key for secure communication
func (sc *SecureChannel) RegisterPeer(peerID string, publicKey *ecdsa.PublicKey) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if publicKey == nil {
		return errors.New("public key cannot be nil")
	}

	sc.peerKeys[peerID] = publicKey
	return nil
}

// EncryptMessage encrypts a message for a specific peer using AES-GCM
func (sc *SecureChannel) EncryptMessage(peerID string, plaintext []byte) ([]byte, error) {
	sc.mu.RLock()
	sessionKey, exists := sc.sessionKeys[peerID]
	sc.mu.RUnlock()

	if !exists {
		// Establish session key if it doesn't exist
		var err error
		sessionKey, err = sc.establishSessionKey(peerID)
		if err != nil {
			return nil, fmt.Errorf("failed to establish session key: %w", err)
		}
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptMessage decrypts a message from a peer
func (sc *SecureChannel) DecryptMessage(peerID string, ciphertext []byte) ([]byte, error) {
	sc.mu.RLock()
	sessionKey, exists := sc.sessionKeys[peerID]
	sc.mu.RUnlock()

	if !exists {
		return nil, errors.New("no session key for peer")
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// SignData signs data using ECDSA for authentication
func (sc *SecureChannel) SignData(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := ecdsa.SignASN1(rand.Reader, sc.privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

// VerifySignature verifies a signature from a peer
func (sc *SecureChannel) VerifySignature(peerID string, data, signature []byte) error {
	sc.mu.RLock()
	publicKey, exists := sc.peerKeys[peerID]
	sc.mu.RUnlock()

	if !exists {
		return errors.New("peer public key not found")
	}

	hash := sha256.Sum256(data)
	valid := ecdsa.VerifyASN1(publicKey, hash[:], signature)
	if !valid {
		return errors.New("invalid signature")
	}

	return nil
}

// establishSessionKey creates a shared session key using ECDH
func (sc *SecureChannel) establishSessionKey(peerID string) ([]byte, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	peerKey, exists := sc.peerKeys[peerID]
	if !exists {
		return nil, errors.New("peer public key not registered")
	}

	// ECDH key agreement
	x, _ := peerKey.Curve.ScalarMult(peerKey.X, peerKey.Y, sc.privateKey.D.Bytes())
	sharedSecret := x.Bytes()

	// Derive session key using SHA-256
	sessionKey := sha256.Sum256(sharedSecret)
	sc.sessionKeys[peerID] = sessionKey[:]

	return sessionKey[:], nil
}

// RotateSessionKey rotates the session key for a peer
func (sc *SecureChannel) RotateSessionKey(peerID string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	delete(sc.sessionKeys, peerID)
	_, err := sc.establishSessionKey(peerID)
	return err
}

// GetTLSConfig returns the TLS configuration for secure connections
func (sc *SecureChannel) GetTLSConfig() *tls.Config {
	return sc.tlsConfig
}

// ExportPublicKey exports the public key in PEM format
func (sc *SecureChannel) ExportPublicKey() ([]byte, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(sc.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}

// ImportPublicKey imports a public key from PEM format
func ImportPublicKey(pemData []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}

	return ecdsaPub, nil
}

// createTLSConfig creates a secure TLS configuration
func createTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		SessionTicketsDisabled: false,
		ClientSessionCache:     tls.NewLRUClientSessionCache(128),
	}
}

// SecureMessage wraps an encrypted message with metadata
type SecureMessage struct {
	SenderID   string
	RecipientID string
	Timestamp  time.Time
	Ciphertext []byte
	Signature  []byte
}

// SecureModelUpdate encrypts and signs a model update
func (sc *SecureChannel) SecureModelUpdate(peerID string, modelData []byte) (*SecureMessage, error) {
	ciphertext, err := sc.EncryptMessage(peerID, modelData)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	signature, err := sc.SignData(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	return &SecureMessage{
		SenderID:    "self",
		RecipientID: peerID,
		Timestamp:   time.Now(),
		Ciphertext:  ciphertext,
		Signature:   signature,
	}, nil
}

// VerifyAndDecryptMessage verifies and decrypts a secure message
func (sc *SecureChannel) VerifyAndDecryptMessage(msg *SecureMessage) ([]byte, error) {
	// Verify signature
	if err := sc.VerifySignature(msg.SenderID, msg.Ciphertext, msg.Signature); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// Decrypt message
	plaintext, err := sc.DecryptMessage(msg.SenderID, msg.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
