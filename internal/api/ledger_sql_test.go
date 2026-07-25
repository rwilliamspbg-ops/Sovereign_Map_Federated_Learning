// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package api

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// A thread-safe counter of executed database queries.
type queryTracker struct {
	mu           sync.Mutex
	queryCount   map[string]int
	totalQueries int
}

var tracker = &queryTracker{
	queryCount: make(map[string]int),
}

func (t *queryTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.queryCount = make(map[string]int)
	t.totalQueries = 0
}

func (t *queryTracker) Log(q string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	// Normalize spacing a bit
	normalized := strings.Join(strings.Fields(q), " ")
	t.queryCount[normalized]++
	t.totalQueries++
}

func (t *queryTracker) GetCount(q string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	normalized := strings.Join(strings.Fields(q), " ")
	return t.queryCount[normalized]
}

type benchMockDriver struct{}

func (d *benchMockDriver) Open(name string) (driver.Conn, error) {
	return &benchMockConn{}, nil
}

type benchMockConn struct{}

func (c *benchMockConn) Prepare(query string) (driver.Stmt, error) {
	return &benchMockStmt{query: query}, nil
}

func (c *benchMockConn) Close() error {
	return nil
}

func (c *benchMockConn) Begin() (driver.Tx, error) {
	return &benchMockTx{}, nil
}

type benchMockStmt struct {
	query string
}

func (s *benchMockStmt) Close() error {
	return nil
}

func (s *benchMockStmt) NumInput() int {
	return -1
}

func (s *benchMockStmt) Exec(args []driver.Value) (driver.Result, error) {
	tracker.Log(s.query)
	return &benchMockResult{}, nil
}

func (s *benchMockStmt) Query(args []driver.Value) (driver.Rows, error) {
	tracker.Log(s.query)
	return &benchMockRows{query: s.query}, nil
}

type benchMockResult struct{}

func (r *benchMockResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (r *benchMockResult) RowsAffected() (int64, error) {
	return 1, nil
}

type benchMockRows struct {
	query string
	step  int
}

func (r *benchMockRows) Columns() []string {
	if strings.Contains(r.query, "count(*)") {
		return []string{"count"}
	}
	if strings.Contains(r.query, "INSERT INTO mohawk_ledger_entries") {
		return []string{"id", "created_at"}
	}
	if strings.Contains(r.query, "mohawk_ledger_entries") {
		return []string{"id", "entry_id", "stream_id", "seq_no", "created_at", "event_type", "proof_hash", "prev_hash", "entry_hash", "idempotency_key", "role", "accepted", "latency_ms", "error_text"}
	}
	if strings.Contains(r.query, "mohawk_ledger_stream_heads") {
		return []string{"seq_no", "entry_hash"}
	}
	return []string{"id"}
}

func (r *benchMockRows) Close() error {
	return nil
}

func (r *benchMockRows) Next(dest []driver.Value) error {
	if r.step > 0 {
		return io.EOF
	}
	r.step++

	if strings.Contains(r.query, "count(*)") {
		dest[0] = int64(100) // Dummy database total of 100 entries
		return nil
	}

	if strings.Contains(r.query, "INSERT INTO mohawk_ledger_entries") {
		dest[0] = int64(1)
		dest[1] = time.Now()
		return nil
	}

	if strings.Contains(r.query, "mohawk_ledger_entries") {
		// Mock loadEntryByIdempotency returning empty rows
		return io.EOF
	}

	if strings.Contains(r.query, "mohawk_ledger_stream_heads") {
		dest[0] = int64(50)
		dest[1] = "dummy_entry_hash"
		return nil
	}

	dest[0] = int64(1)
	return nil
}

type benchMockTx struct{}

func (t *benchMockTx) Commit() error {
	return nil
}

func (t *benchMockTx) Rollback() error {
	return nil
}

func init() {
	sql.Register("bench-mock-driver", &benchMockDriver{})
}

func TestSQLProofLedgerCountCaching(t *testing.T) {
	tracker.Reset()

	l, err := NewSQLProofLedger("bench-mock-driver", "mock-dsn", 1000, 10)
	if err != nil {
		t.Fatalf("failed to initialize SQLProofLedger: %v", err)
	}

	// Verify count query occurred once as warmup in NewSQLProofLedger
	warmupQueries := tracker.GetCount("SELECT count(*) FROM mohawk_ledger_entries")
	if warmupQueries != 1 {
		t.Errorf("expected exactly 1 warmup count query, got %d", warmupQueries)
	}

	// Verify that l.countCached is set and count matches our mock's 100
	if !l.countCached {
		t.Errorf("expected count to be cached")
	}
	if l.cachedCount != 100 {
		t.Errorf("expected cached count to be 100, got %d", l.cachedCount)
	}

	// Call Len() 10 times and verify no database query happens
	for i := 0; i < 10; i++ {
		length := l.Len()
		if length != 100 {
			t.Errorf("iteration %d: expected length 100, got %d", i, length)
		}
	}

	countAfterLen := tracker.GetCount("SELECT count(*) FROM mohawk_ledger_entries")
	if countAfterLen != 1 {
		t.Errorf("expected no additional count queries after calling Len(), but got total %d", countAfterLen)
	}

	// Record a new entry
	_, replay := l.RecordWithOptions("verify", []byte("proof"), "user", true, 42, nil, LedgerRecordOptions{
		StreamID: "test-stream",
	})
	if replay {
		t.Errorf("did not expect replay")
	}

	// Cached count should have incremented to 101
	if l.cachedCount != 101 {
		t.Errorf("expected cached count to increment to 101, got %d", l.cachedCount)
	}

	// Call Len() again and verify it returns 101 without database count queries
	length := l.Len()
	if length != 101 {
		t.Errorf("expected length 101, got %d", length)
	}

	finalCountQueries := tracker.GetCount("SELECT count(*) FROM mohawk_ledger_entries")
	if finalCountQueries != 1 {
		t.Errorf("expected total count queries to remain 1, got %d", finalCountQueries)
	}
}

func TestSQLProofLedgerCapacityLimiting(t *testing.T) {
	tracker.Reset()

	l, err := NewSQLProofLedger("bench-mock-driver", "mock-dsn", 50, 10)
	if err != nil {
		t.Fatalf("failed to initialize SQLProofLedger: %v", err)
	}

	// Total database count is mock-returned as 100.
	// But capacity is 50, so Len() should limit the returned length to 50.
	length := l.Len()
	if length != 50 {
		t.Errorf("expected length to be limited to capacity of 50, got %d (cached: %d)", length, l.cachedCount)
	}
}

func BenchmarkSQLProofLedgerLenUncached(b *testing.B) {
	l, err := NewSQLProofLedger("bench-mock-driver", "mock-dsn", 1000, 10)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.mu.Lock()
		l.countCached = false
		l.mu.Unlock()
		_ = l.Len()
	}
}

func BenchmarkSQLProofLedgerLenCached(b *testing.B) {
	l, err := NewSQLProofLedger("bench-mock-driver", "mock-dsn", 1000, 10)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Len()
	}
}
