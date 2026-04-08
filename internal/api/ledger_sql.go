// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultSQLDriver = "postgres"

// SQLProofLedger stores ledger entries in a CockroachDB/PostgreSQL-compatible SQL backend.
type SQLProofLedger struct {
	db              *sql.DB
	cap             int
	checkpointEvery uint64
}

// NewSQLProofLedgerFromEnv initializes a SQL ledger from environment variables.
func NewSQLProofLedgerFromEnv(capacity int) (*SQLProofLedger, error) {
	driver := strings.TrimSpace(os.Getenv("MOHAWK_LEDGER_SQL_DRIVER"))
	if driver == "" {
		driver = defaultSQLDriver
	}
	dsn := strings.TrimSpace(os.Getenv("MOHAWK_LEDGER_SQL_DSN"))
	checkpointEvery := uint64(defaultCheckpointEvery)
	if raw := strings.TrimSpace(os.Getenv("MOHAWK_LEDGER_CHECKPOINT_EVERY")); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse MOHAWK_LEDGER_CHECKPOINT_EVERY: %w", err)
		}
		checkpointEvery = parsed
	}
	return NewSQLProofLedger(driver, dsn, capacity, checkpointEvery)
}

// NewSQLProofLedger initializes a SQL-backed ledger and ensures schema exists.
func NewSQLProofLedger(driver string, dsn string, capacity int, checkpointEvery uint64) (*SQLProofLedger, error) {
	if strings.TrimSpace(driver) == "" {
		return nil, errors.New("sql driver is required")
	}
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("MOHAWK_LEDGER_SQL_DSN is required for SQL ledger backend")
	}
	if capacity <= 0 {
		capacity = defaultLedgerCap
	}
	if checkpointEvery == 0 {
		checkpointEvery = defaultCheckpointEvery
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open sql ledger: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sql ledger: %w", err)
	}

	ledger := &SQLProofLedger{
		db:              db,
		cap:             capacity,
		checkpointEvery: checkpointEvery,
	}
	if err := ledger.ensureSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return ledger, nil
}

func (l *SQLProofLedger) ensureSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS mohawk_ledger_entries (
			id BIGSERIAL PRIMARY KEY,
			entry_id TEXT NOT NULL UNIQUE,
			stream_id TEXT NOT NULL,
			seq_no BIGINT NOT NULL,
			event_type TEXT NOT NULL,
			proof_hash TEXT NOT NULL,
			prev_hash TEXT NOT NULL,
			entry_hash TEXT NOT NULL,
			idempotency_key TEXT NOT NULL UNIQUE,
			role TEXT NOT NULL,
			accepted BOOL NOT NULL,
			latency_ms BIGINT NOT NULL,
			error_text TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (stream_id, seq_no)
		)`,
		`CREATE TABLE IF NOT EXISTS mohawk_ledger_stream_heads (
			stream_id TEXT PRIMARY KEY,
			seq_no BIGINT NOT NULL,
			entry_hash TEXT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS mohawk_ledger_checkpoints (
			stream_id TEXT NOT NULL,
			through_seq BIGINT NOT NULL,
			root_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			PRIMARY KEY (stream_id, through_seq)
		)`,
	}
	for _, stmt := range stmts {
		if _, err := l.db.Exec(stmt); err != nil {
			return fmt.Errorf("ensure ledger schema: %w", err)
		}
	}
	return nil
}

func (l *SQLProofLedger) Record(eventType string, proofBytes []byte, role string, accepted bool, latencyMs int64, verifyErr error) {
	_, _ = l.RecordWithOptions(eventType, proofBytes, role, accepted, latencyMs, verifyErr, LedgerRecordOptions{})
}

func (l *SQLProofLedger) RecordWithOptions(eventType string, proofBytes []byte, role string, accepted bool, latencyMs int64, verifyErr error, opts LedgerRecordOptions) (LedgerEntry, bool) {
	errStr := ""
	if verifyErr != nil {
		errStr = verifyErr.Error()
	}
	streamID := strings.TrimSpace(opts.StreamID)
	if streamID == "" {
		streamID = strings.TrimSpace(eventType)
	}

	proofHash := computeProofHash(proofBytes)
	idempotencyKey := strings.TrimSpace(opts.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = computeDeterministicID(streamID, proofHash, role, "")
	}

	tx, err := l.db.Begin()
	if err != nil {
		return LedgerEntry{}, false
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	existing, found, err := loadEntryByIdempotency(tx, idempotencyKey)
	if err == nil && found {
		existing.Replay = true
		_ = tx.Rollback()
		tx = nil
		return existing, true
	}

	seqNo, prevHash, err := loadStreamHeadForUpdate(tx, streamID)
	if err != nil {
		return LedgerEntry{}, false
	}
	nextSeq := seqNo + 1
	entryHash := computeEntryHash(streamID, nextSeq, proofHash, role, accepted, errStr, idempotencyKey, prevHash)
	entryID := computeDeterministicID(streamID, proofHash, role, idempotencyKey)

	entry := LedgerEntry{
		EntryID:        entryID,
		StreamID:       streamID,
		SeqNo:          nextSeq,
		EventType:      eventType,
		ProofHash:      proofHash,
		PrevHash:       prevHash,
		EntryHash:      entryHash,
		IdempotencyKey: idempotencyKey,
		Role:           role,
		Accepted:       accepted,
		LatencyMs:      latencyMs,
		Error:          errStr,
	}

	if err := insertEntry(tx, &entry); err != nil {
		return LedgerEntry{}, false
	}
	if err := upsertStreamHead(tx, streamID, nextSeq, entryHash); err != nil {
		return LedgerEntry{}, false
	}
	if l.checkpointEvery > 0 && nextSeq%l.checkpointEvery == 0 {
		if err := upsertCheckpoint(tx, streamID, nextSeq, entryHash); err != nil {
			return LedgerEntry{}, false
		}
	}

	if err := tx.Commit(); err != nil {
		return LedgerEntry{}, false
	}
	tx = nil
	return entry, false
}

func (l *SQLProofLedger) Entries() []LedgerEntry {
	query := `SELECT id, entry_id, stream_id, seq_no, created_at, event_type, proof_hash, prev_hash, entry_hash, idempotency_key, role, accepted, latency_ms, error_text
		FROM mohawk_ledger_entries ORDER BY id DESC`
	args := []interface{}{}
	if l.cap > 0 {
		query += ` LIMIT $1`
		args = append(args, l.cap)
	}
	rows, err := l.db.Query(query, args...)
	if err != nil {
		return []LedgerEntry{}
	}
	defer rows.Close()

	entries := make([]LedgerEntry, 0)
	for rows.Next() {
		entry, scanErr := scanLedgerEntry(rows)
		if scanErr != nil {
			return []LedgerEntry{}
		}
		entries = append(entries, entry)
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
}

func (l *SQLProofLedger) Len() int {
	var total int
	if l.cap > 0 {
		if err := l.db.QueryRow(
			`SELECT count(*) FROM (SELECT 1 FROM mohawk_ledger_entries LIMIT $1) AS limited_entries`,
			l.cap,
		).Scan(&total); err != nil {
			return 0
		}
		return total
	}
	if err := l.db.QueryRow(`SELECT count(*) FROM mohawk_ledger_entries`).Scan(&total); err != nil {
		return 0
	}
	return total
}

func (l *SQLProofLedger) Checkpoints() []LedgerCheckpoint {
	query := `SELECT stream_id, through_seq, root_hash, created_at FROM mohawk_ledger_checkpoints ORDER BY stream_id, through_seq DESC`
	args := []interface{}{}
	if l.cap > 0 {
		query += ` LIMIT $1`
		args = append(args, l.cap)
	}
	rows, err := l.db.Query(query, args...)
	if err != nil {
		return []LedgerCheckpoint{}
	}
	defer rows.Close()

	out := make([]LedgerCheckpoint, 0)
	for rows.Next() {
		var cp LedgerCheckpoint
		var ts time.Time
		if err := rows.Scan(&cp.StreamID, &cp.ThroughSeq, &cp.RootHash, &ts); err != nil {
			return []LedgerCheckpoint{}
		}
		cp.Timestamp = ts.UTC().Format(time.RFC3339Nano)
		out = append(out, cp)
	}
	return out
}

func (l *SQLProofLedger) Reconcile() LedgerReconcileReport {
	return reconcileEntries(l.Entries())
}

func (l *SQLProofLedger) StorageMode() string {
	return "cockroach-sql"
}

func (l *SQLProofLedger) Capacity() int {
	return l.cap
}

// Readiness validates SQL connectivity for deployment gating.
func (l *SQLProofLedger) Readiness() (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := l.db.PingContext(ctx); err != nil {
		return false, err.Error()
	}
	return true, ""
}

func scanLedgerEntry(scanner interface {
	Scan(dest ...interface{}) error
}) (LedgerEntry, error) {
	var e LedgerEntry
	var ts time.Time
	if err := scanner.Scan(&e.ID, &e.EntryID, &e.StreamID, &e.SeqNo, &ts, &e.EventType, &e.ProofHash, &e.PrevHash, &e.EntryHash, &e.IdempotencyKey, &e.Role, &e.Accepted, &e.LatencyMs, &e.Error); err != nil {
		return LedgerEntry{}, err
	}
	e.Timestamp = ts.UTC().Format(time.RFC3339Nano)
	return e, nil
}

func loadEntryByIdempotency(tx *sql.Tx, idempotencyKey string) (LedgerEntry, bool, error) {
	row := tx.QueryRow(`SELECT id, entry_id, stream_id, seq_no, created_at, event_type, proof_hash, prev_hash, entry_hash, idempotency_key, role, accepted, latency_ms, error_text
		FROM mohawk_ledger_entries WHERE idempotency_key = $1`, idempotencyKey)
	entry, err := scanLedgerEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LedgerEntry{}, false, nil
		}
		return LedgerEntry{}, false, err
	}
	return entry, true, nil
}

func loadStreamHeadForUpdate(tx *sql.Tx, streamID string) (uint64, string, error) {
	row := tx.QueryRow(`SELECT seq_no, entry_hash FROM mohawk_ledger_stream_heads WHERE stream_id = $1 FOR UPDATE`, streamID)
	var seqNo uint64
	var entryHash string
	if err := row.Scan(&seqNo, &entryHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", nil
		}
		return 0, "", err
	}
	return seqNo, entryHash, nil
}

func insertEntry(tx *sql.Tx, entry *LedgerEntry) error {
	row := tx.QueryRow(`INSERT INTO mohawk_ledger_entries (entry_id, stream_id, seq_no, event_type, proof_hash, prev_hash, entry_hash, idempotency_key, role, accepted, latency_ms, error_text)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at`,
		entry.EntryID,
		entry.StreamID,
		entry.SeqNo,
		entry.EventType,
		entry.ProofHash,
		entry.PrevHash,
		entry.EntryHash,
		entry.IdempotencyKey,
		entry.Role,
		entry.Accepted,
		entry.LatencyMs,
		entry.Error,
	)
	var ts time.Time
	if err := row.Scan(&entry.ID, &ts); err != nil {
		return err
	}
	entry.Timestamp = ts.UTC().Format(time.RFC3339Nano)
	return nil
}

func upsertStreamHead(tx *sql.Tx, streamID string, seqNo uint64, entryHash string) error {
	_, err := tx.Exec(`INSERT INTO mohawk_ledger_stream_heads (stream_id, seq_no, entry_hash)
		VALUES ($1,$2,$3)
		ON CONFLICT (stream_id)
		DO UPDATE SET seq_no = EXCLUDED.seq_no, entry_hash = EXCLUDED.entry_hash, updated_at = now()`, streamID, seqNo, entryHash)
	return err
}

func upsertCheckpoint(tx *sql.Tx, streamID string, throughSeq uint64, rootHash string) error {
	_, err := tx.Exec(`INSERT INTO mohawk_ledger_checkpoints (stream_id, through_seq, root_hash)
		VALUES ($1,$2,$3)
		ON CONFLICT (stream_id, through_seq)
		DO UPDATE SET root_hash = EXCLUDED.root_hash, created_at = now()`, streamID, throughSeq, rootHash)
	return err
}

func computeProofHash(proofBytes []byte) string {
	h := sha256.Sum256(proofBytes)
	return hex.EncodeToString(h[:])
}
