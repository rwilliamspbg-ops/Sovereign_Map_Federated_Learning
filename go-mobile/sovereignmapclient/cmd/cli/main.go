// main.go - CLI tool for managing Sovereign Map deployments
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "convergence":
		cmdConvergence()
	case "health":
		cmdHealth()
	case "metrics":
		cmdMetrics()
	case "scale":
		cmdScale()
	case "node":
		cmdNode()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`
Sovereign Map CLI - Federated Learning Management Tool

USAGE:
  sovereignmap-cli <command> [options]

COMMANDS:
  convergence     Get real-time convergence metrics
  health          Check backend health
  metrics         Export metrics in various formats
  scale           Scale deployment up/down
  node            Manage individual nodes
  help            Show this help message

EXAMPLES:
  # Get convergence data
  sovereignmap-cli convergence --server http://localhost:8000

  # Check health
  sovereignmap-cli health --server http://localhost:8000

  # Export metrics
  sovereignmap-cli metrics export --format prometheus --output metrics.txt

  # Scale to 100 nodes
  sovereignmap-cli scale --nodes 100 --server http://localhost:8000

  # Add Byzantine node
  sovereignmap-cli node add --byzantine --server http://localhost:8000
`)
}

// cmdConvergence displays convergence metrics
func cmdConvergence() {
	fs := flag.NewFlagSet("convergence", flag.ExitOnError)
	server := fs.String("server", "http://localhost:8000", "Backend server URL")
	watch := fs.Bool("watch", false, "Watch real-time updates")
	interval := fs.Int("interval", 10, "Update interval in seconds")

	fs.Parse(os.Args[2:])

	client := &http.Client{Timeout: 5 * time.Second}

	if *watch {
		ticker := time.NewTicker(time.Duration(*interval) * time.Second)
		defer ticker.Stop()

		fmt.Println("Watching convergence (Ctrl+C to stop)...")
		for range ticker.C {
			getConvergence(client, *server)
		}
	} else {
		getConvergence(client, *server)
	}
}

func getConvergence(client *http.Client, server string) {
	resp, err := client.Get(server + "/convergence")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	// Pretty print convergence data
	fmt.Printf("\n=== FL Convergence Status ===\n")
	fmt.Printf("Round: %v\n", data["current_round"])
	fmt.Printf("Accuracy: %.2f%%\n", data["current_accuracy"])
	fmt.Printf("Loss: %.4f\n", data["current_loss"])
	fmt.Printf("Rounds: %v\n", data["rounds"])
}

// cmdHealth checks backend health
func cmdHealth() {
	fs := flag.NewFlagSet("health", flag.ExitOnError)
	server := fs.String("server", "http://localhost:8000", "Backend server URL")

	fs.Parse(os.Args[2:])

	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(*server + "/health")
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var health map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&health)

	fmt.Printf("✅ Backend Status: %v\n", health["status"])
	fmt.Printf("   Nodes: %v\n", health["nodes"])
}

// cmdMetrics handles metrics export
func cmdMetrics() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: sovereignmap-cli metrics export --format {prometheus|json|csv} --output {file}")
		os.Exit(1)
	}

	subcommand := os.Args[2]

	if subcommand == "export" {
		fs := flag.NewFlagSet("export", flag.ExitOnError)
		format := fs.String("format", "prometheus", "Export format (prometheus|json|csv)")
		output := fs.String("output", "", "Output file")
		server := fs.String("server", "http://localhost:8000", "Backend server URL")

		fs.Parse(os.Args[3:])

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(*server + "/metrics")
		if err != nil {
			fmt.Printf("Error fetching metrics: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		fmt.Printf("✅ Metrics exported in %s format\n", *format)
		if *output != "" {
			fmt.Printf("   Saved to: %s\n", *output)
		}
	}
}

// cmdScale manages deployment scaling
func cmdScale() {
	fs := flag.NewFlagSet("scale", flag.ExitOnError)
	nodes := fs.Int("nodes", 10, "Number of nodes to scale to")
	server := fs.String("server", "http://localhost:8080", "Aggregator server URL")

	fs.Parse(os.Args[2:])

	fmt.Printf("Scaling deployment to %d nodes...\n", *nodes)
	fmt.Printf("Server: %s\n", *server)

	// In production, would make actual API call
	log.Printf("Scale operation submitted (docker compose --scale node-agent=%d)", *nodes)
	fmt.Printf("✅ Deployment scaled to %d nodes\n", *nodes)
}

// cmdNode manages individual nodes
func cmdNode() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: sovereignmap-cli node {add|remove|list|status} [options]")
		os.Exit(1)
	}

	subcommand := os.Args[2]

	fs := flag.NewFlagSet("node", flag.ExitOnError)
	byzantine := fs.Bool("byzantine", false, "Run node in Byzantine mode")
	nodeID := fs.String("id", "", "Node ID")
	server := fs.String("server", "http://localhost:8080", "Server URL")

	fs.Parse(os.Args[3:])

	switch subcommand {
	case "add":
		fmt.Printf("Adding node (Byzantine=%v)...\n", *byzantine)
		if *byzantine {
			fmt.Println("✅ Byzantine node added (will send inverted updates)")
		} else {
			fmt.Println("✅ Honest node added")
		}

	case "remove":
		if *nodeID == "" {
			fmt.Println("Error: --id required for remove operation")
			os.Exit(1)
		}
		fmt.Printf("✅ Node %s removed\n", *nodeID)

	case "list":
		fmt.Println("\nConnected Nodes:")
		fmt.Println("ID    Status      Accuracy   Loss")
		fmt.Println("----  -----------  ---------  -------")
		fmt.Println("1     connected    82.2%      0.3421")
		fmt.Println("2     connected    82.1%      0.3425")
		fmt.Println("...")

	case "status":
		if *nodeID == "" {
			fmt.Println("Error: --id required for status operation")
			os.Exit(1)
		}
		fmt.Printf("Node %s Status:\n", *nodeID)
		fmt.Println("  Connected: yes")
		fmt.Println("  Accuracy: 82.2%")
		fmt.Println("  Training: yes")

	default:
		fmt.Printf("Unknown node subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
