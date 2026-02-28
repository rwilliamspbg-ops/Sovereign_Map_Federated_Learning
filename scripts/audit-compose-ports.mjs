#!/usr/bin/env node

import fs from "fs";
import path from "path";
import { execSync } from "child_process";

function walkForComposeFiles(dirPath, out = []) {
  const entries = fs.readdirSync(dirPath, { withFileTypes: true });
  for (const entry of entries) {
    if (entry.name === ".git" || entry.name === "node_modules") {
      continue;
    }
    const fullPath = path.join(dirPath, entry.name);
    if (entry.isDirectory()) {
      walkForComposeFiles(fullPath, out);
      continue;
    }
    if (/^docker-compose.*\.ya?ml$/i.test(entry.name)) {
      out.push(fullPath);
    }
  }
  return out;
}

function parseArgs(argv) {
  const args = {
    outDir: null,
    targetFile: null,
    failOnConflicts: false,
  };

  for (let index = 2; index < argv.length; index += 1) {
    const token = argv[index];
    if (token === "--out-dir") {
      args.outDir = argv[index + 1] || null;
      index += 1;
    } else if (token === "--target-file") {
      args.targetFile = argv[index + 1] || null;
      index += 1;
    } else if (token === "--fail-on-conflicts") {
      args.failOnConflicts = true;
    }
  }

  return args;
}

function run(command) {
  return execSync(command, {
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  });
}

function normalizePortEntry(entry) {
  if (typeof entry === "string") {
    const parts = entry.split(":");
    if (parts.length === 2) {
      const [published, target] = parts;
      return { hostIp: "0.0.0.0", published, target, protocol: "tcp" };
    }
    if (parts.length === 3) {
      const [hostIp, published, target] = parts;
      return { hostIp, published, target, protocol: "tcp" };
    }
    return null;
  }

  if (typeof entry === "object" && entry) {
    if (entry.published == null || entry.target == null) {
      return null;
    }
    return {
      hostIp: entry.host_ip || "0.0.0.0",
      published: String(entry.published),
      target: String(entry.target),
      protocol: entry.protocol || "tcp",
    };
  }

  return null;
}

function getComposeConfig(composeFilePath) {
  const safePath = composeFilePath.replace(/"/g, '\\"');
  const raw = run(`docker compose -f "${safePath}" config --format json`);
  return JSON.parse(raw);
}

function getRunningHostPorts() {
  const result = [];
  const seen = new Set();
  const lines = run("docker ps --format '{{.Names}}|{{.Ports}}'")
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);

  for (const line of lines) {
    const [name, portsRaw] = line.split("|");
    if (!portsRaw) {
      continue;
    }

    const regex = /(?:0\.0\.0\.0|:::|\[::\]):(\d+)->(\d+)\/(tcp|udp)/g;
    let match = regex.exec(portsRaw);
    while (match) {
      const fingerprint = `${name}|${match[1]}|${match[2]}|${match[3]}`;
      if (seen.has(fingerprint)) {
        match = regex.exec(portsRaw);
        continue;
      }
      seen.add(fingerprint);

      result.push({
        container: name,
        published: match[1],
        target: match[2],
        protocol: match[3],
      });
      match = regex.exec(portsRaw);
    }
  }

  return result;
}

function main() {
  const args = parseArgs(process.argv);
  const cwd = process.cwd();
  const composeFiles = walkForComposeFiles(cwd)
    .map((p) => path.relative(cwd, p))
    .sort();

  if (composeFiles.length === 0) {
    console.error("No docker-compose*.yml files found.");
    process.exit(1);
  }

  const hostPortMap = new Map();
  const perFileEntries = [];

  for (const composeFile of composeFiles) {
    let config;
    try {
      config = getComposeConfig(path.join(cwd, composeFile));
    } catch (error) {
      perFileEntries.push({
        composeFile,
        error: String(error.message || error),
      });
      continue;
    }

    const services = config.services || {};
    for (const [serviceName, serviceConfig] of Object.entries(services)) {
      const ports = serviceConfig.ports || [];
      for (const rawPort of ports) {
        const normalized = normalizePortEntry(rawPort);
        if (!normalized) {
          continue;
        }

        const key = `${normalized.hostIp}:${normalized.published}/${normalized.protocol}`;
        const item = {
          composeFile,
          service: serviceName,
          hostIp: normalized.hostIp,
          published: normalized.published,
          target: normalized.target,
          protocol: normalized.protocol,
        };

        if (!hostPortMap.has(key)) {
          hostPortMap.set(key, []);
        }
        hostPortMap.get(key).push(item);
      }
    }
  }

  const staticConflicts = [];
  for (const [key, entries] of hostPortMap.entries()) {
    if (entries.length > 1) {
      staticConflicts.push({ key, entries });
    }
  }

  const runningPorts = getRunningHostPorts();
  const runningPortMap = new Map();
  for (const entry of runningPorts) {
    const key = `${entry.published}/${entry.protocol}`;
    if (!runningPortMap.has(key)) {
      runningPortMap.set(key, []);
    }
    runningPortMap.get(key).push(entry);
  }

  const runtimeConflicts = [];
  const targetFile = args.targetFile ? path.normalize(args.targetFile) : null;
  if (targetFile) {
    const targetPorts = [];
    const targetSeen = new Set();
    for (const [key, entries] of hostPortMap.entries()) {
      for (const entry of entries) {
        if (path.normalize(entry.composeFile) === targetFile || entry.composeFile.endsWith(targetFile)) {
          const fingerprint = `${entry.composeFile}|${entry.service}|${entry.published}|${entry.protocol}`;
          if (targetSeen.has(fingerprint)) {
            continue;
          }
          targetSeen.add(fingerprint);
          targetPorts.push({ key, ...entry });
        }
      }
    }

    for (const targetEntry of targetPorts) {
      const runtimeKey = `${targetEntry.published}/${targetEntry.protocol}`;
      const holders = runningPortMap.get(runtimeKey) || [];
      if (holders.length > 0) {
        runtimeConflicts.push({
          target: targetEntry,
          holders,
        });
      }
    }
  }

  const report = {
    generatedAt: new Date().toISOString(),
    composeFiles,
    staticConflictCount: staticConflicts.length,
    staticConflicts,
    runtimeConflictCount: runtimeConflicts.length,
    runtimeConflicts,
    parseIssues: perFileEntries.filter((x) => x.error),
  };

  if (args.outDir) {
    fs.mkdirSync(args.outDir, { recursive: true });
    fs.writeFileSync(
      path.join(args.outDir, "compose-port-audit.json"),
      `${JSON.stringify(report, null, 2)}\n`,
      "utf8",
    );

    const md = [];
    md.push("# Compose Port Audit");
    md.push("");
    md.push(`Generated: ${report.generatedAt}`);
    md.push(`Compose files scanned: ${report.composeFiles.length}`);
    md.push(`Static conflicts: ${report.staticConflictCount}`);
    md.push(`Runtime conflicts: ${report.runtimeConflictCount}`);
    md.push("");

    if (report.staticConflicts.length > 0) {
      md.push("## Static Conflicts (Across Compose Files)");
      md.push("");
      for (const conflict of report.staticConflicts) {
        md.push(`- ${conflict.key}`);
        for (const entry of conflict.entries) {
          md.push(
            `  - ${entry.composeFile} :: ${entry.service} publishes ${entry.published}->${entry.target}/${entry.protocol}`,
          );
        }
      }
      md.push("");
    }

    if (report.runtimeConflicts.length > 0) {
      md.push("## Runtime Conflicts (Against Running Containers)");
      md.push("");
      for (const conflict of report.runtimeConflicts) {
        const target = conflict.target;
        md.push(`- ${target.composeFile} :: ${target.service} wants ${target.published}/${target.protocol}`);
        for (const holder of conflict.holders) {
          md.push(`  - held by running container ${holder.container} (${holder.published}->${holder.target}/${holder.protocol})`);
        }
      }
      md.push("");
    }

    if (report.parseIssues.length > 0) {
      md.push("## Parse Issues");
      md.push("");
      for (const issue of report.parseIssues) {
        md.push(`- ${issue.composeFile}: ${issue.error}`);
      }
      md.push("");
    }

    fs.writeFileSync(path.join(args.outDir, "compose-port-audit.md"), `${md.join("\n")}\n`, "utf8");
  }

  console.log(`Scanned ${composeFiles.length} compose files`);
  console.log(`Static conflicts: ${staticConflicts.length}`);
  console.log(`Runtime conflicts: ${runtimeConflicts.length}`);

  if (args.failOnConflicts && (staticConflicts.length > 0 || runtimeConflicts.length > 0)) {
    process.exit(2);
  }
}

main();
