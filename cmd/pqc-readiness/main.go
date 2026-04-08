// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/crypto/pqc"
)

func main() {
	status := pqc.Status()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(status); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode status: %v\n", err)
		os.Exit(1)
	}
}
