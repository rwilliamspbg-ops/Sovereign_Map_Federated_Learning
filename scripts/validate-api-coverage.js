#!/usr/bin/env node

/*
Validates that Flask route decorators are represented in corresponding OpenAPI specs.
This is a lightweight guardrail to prevent doc drift.
*/

const fs = require("fs");
const path = require("path");

const root = process.cwd();

const targets = [
  {
    name: "control-plane",
    codeFile: "sovereignmap_production_backend_v2.py",
    openapiFile: "docs/api/openapi.yaml",
  },
  {
    name: "tokenomics-exporter",
    codeFile: "tokenomics_metrics_exporter.py",
    openapiFile: "docs/api/openapi.tokenomics.yaml",
  },
  {
    name: "tpm-exporter",
    codeFile: "tpm_metrics_exporter.py",
    openapiFile: "docs/api/openapi.tpm.yaml",
  },
  {
    name: "training-service",
    codeFile: "packages/training/api.py",
    openapiFile: "docs/api/openapi.training.yaml",
  },
];

function normalizeFlaskPath(routePath) {
  return routePath
    .replace(/<int:([a-zA-Z_][a-zA-Z0-9_]*)>/g, "{$1}")
    .replace(/<([a-zA-Z_][a-zA-Z0-9_]*)>/g, "{$1}");
}

function extractRouteOps(pythonSource) {
  const out = new Set();
  const routeRegex = /@app\.route\(\s*["']([^"']+)["']\s*,\s*methods\s*=\s*\[([^\]]+)\]\s*\)/g;

  let match;
  while ((match = routeRegex.exec(pythonSource)) !== null) {
    const route = normalizeFlaskPath(match[1].trim());
    const methodsRaw = match[2];
    const methodRegex = /["']([A-Z]+)["']/g;

    let methodMatch;
    while ((methodMatch = methodRegex.exec(methodsRaw)) !== null) {
      const method = methodMatch[1].toLowerCase();
      out.add(`${method} ${route}`);
    }
  }

  return out;
}

function extractOpenApiOps(openapiSource) {
  const out = new Set();
  const lines = openapiSource.split(/\r?\n/);

  let inPaths = false;
  let currentPath = null;

  for (const line of lines) {
    if (!inPaths) {
      if (/^paths:\s*$/.test(line)) {
        inPaths = true;
      }
      continue;
    }

    if (/^[^\s]/.test(line) && !/^paths:\s*$/.test(line)) {
      break;
    }

    const pathMatch = line.match(/^\s{2}(\/[^:]+):\s*$/);
    if (pathMatch) {
      currentPath = pathMatch[1];
      continue;
    }

    const opMatch = line.match(/^\s{4}(get|post|put|patch|delete|options|head|trace):\s*$/i);
    if (opMatch && currentPath) {
      out.add(`${opMatch[1].toLowerCase()} ${currentPath}`);
    }
  }

  return out;
}

let hasFailures = false;

for (const target of targets) {
  const codePath = path.join(root, target.codeFile);
  const specPath = path.join(root, target.openapiFile);

  if (!fs.existsSync(codePath)) {
    console.error(`[api:validate] missing code file: ${target.codeFile}`);
    hasFailures = true;
    continue;
  }

  if (!fs.existsSync(specPath)) {
    console.error(`[api:validate] missing spec file: ${target.openapiFile}`);
    hasFailures = true;
    continue;
  }

  const code = fs.readFileSync(codePath, "utf8");
  const spec = fs.readFileSync(specPath, "utf8");

  const codeOps = extractRouteOps(code);
  const specOps = extractOpenApiOps(spec);

  const missing = [...codeOps].filter((op) => !specOps.has(op));

  if (missing.length > 0) {
    hasFailures = true;
    console.error(`\n[api:validate] ${target.name}: missing operations in ${target.openapiFile}`);
    for (const op of missing) {
      console.error(`  - ${op}`);
    }
  } else {
    console.log(`[api:validate] ${target.name}: OK (${codeOps.size} operations mapped)`);
  }
}

if (hasFailures) {
  process.exit(1);
}

console.log("[api:validate] all API route coverage checks passed");
