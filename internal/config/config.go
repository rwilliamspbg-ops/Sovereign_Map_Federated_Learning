// Copyright 2026 Sovereign-Mohawk Core Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	NodeID          string
	AggregatorURL   string
	DatabaseURI     string
	BatchSize       int
	Timeout         time.Duration
	WASMBinaryPath  string
	TPMEnabled      bool
	LogLevel        string
}

// Load reads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		NodeID:          getEnv("NODE_ID", "node-1"),
		AggregatorURL:   getEnv("AGGREGATOR_URL", "http://aggregator:8080"),
		DatabaseURI:     getEnv("DATABASE_URI", "mongodb://mongo:27017/mydb"),
		BatchSize:       getEnvInt("BATCH_SIZE", 32),
		Timeout:         getEnvDuration("TIMEOUT", 30*time.Second),
		WASMBinaryPath:  getEnv("WASM_BINARY_PATH", "/app/wasm/verify.wasm"),
		TPMEnabled:      getEnvBool("TPM_ENABLED", false),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
