#!/usr/bin/env node

import { Command } from 'commander';
import chalk from 'chalk';
import ora from 'ora';
import inquirer from 'inquirer';
import { SovereignNode, NodeState } from '@sovereignmap/core';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';

const program = new Command();
const CONFIG_DIR = path.join(os.homedir(), '.sovereignmap');
const CONFIG_FILE = path.join(CONFIG_DIR, 'config.json');

program
  .name('sovereignmap')
  .description('Sovereign Map Node Operator CLI')
  .version('0.1.0-alpha.1');

program
  .command('init')
  .description('Initialize a new Genesis Node')
  .option('-r, --region <region>', 'Geographic region')
  .option('--lat <latitude>', 'Latitude')
  .option('--lng <longitude>', 'Longitude')
  .option('--tpm', 'Enable TPM attestation')
  .option('--aggregator', 'Configure as aggregator node')
  .action(async (options) => {
    console.log(chalk.blue.bold('🗺️  Sovereign Map Node Initialization\n'));

    const answers = await inquirer.prompt([
      {
        type: 'input',
        name: 'region',
        message: 'Region name (e.g., bavaria-001):',
        default: options.region || 'genesis-node',
        validate: (input) => input.length >= 2 || 'Region must be at least 2 characters'
      },
      {
        type: 'input',
        name: 'lat',
        message: 'Latitude:',
        default: options.lat || '48.1351',
        validate: (v) => {
          const num = parseFloat(v);
          return !isNaN(num) && num >= -90 && num <= 90 || 'Invalid latitude';
        }
      },
      {
        type: 'input',
        name: 'lng',
        message: 'Longitude:',
        default: options.lng || '11.5820',
        validate: (v) => {
          const num = parseFloat(v);
          return !isNaN(num) && num >= -180 && num <= 180 || 'Invalid longitude';
        }
      },
      {
        type: 'confirm',
        name: 'privacy',
        message: 'Enable SGP-001 privacy (epsilon=1.0)?',
        default: true
      },
      {
        type: 'list',
        name: 'nodeType',
        message: 'Node type:',
        choices: [
          { name: 'Standard Node', value: 'standard' },
          { name: 'Edge Aggregator', value: 'edge' },
          { name: 'Regional Aggregator', value: 'regional' }
        ],
        default: options.aggregator ? 'edge' : 'standard'
      },
      {
        type: 'confirm',
        name: 'tpm',
        message: 'Enable TPM hardware attestation?',
        default: options.tpm || false
      }
    ]);

    const spinner = ora('Initializing node...').start();

    try {
      // Ensure config directory exists
      await fs.mkdir(CONFIG_DIR, { recursive: true });

      const config = {
        nodeId: `${answers.region}-${Date.now()}`,
        region: answers.region,
        coordinates: {
          lat: parseFloat(answers.lat),
          lng: parseFloat(answers.lng)
        },
        privacyBudget: answers.privacy ? {
          epsilon: 1.0,
          delta: 1e-5,
          mechanism: 'gaussian'
        } : undefined,
        hardware: {
          tpmAvailable: answers.tpm,
          npuTops: 0
        },
        consensus: {
          enabled: true,
          aggregatorTier: answers.nodeType === 'standard' ? 'none' : answers.nodeType,
          maxAggregationChildren: answers.nodeType === 'regional' ? 50 : 20
        },
        islandMode: {
          enabled: true,
          storagePath: path.join(CONFIG_DIR, 'island-storage'),
          maxOfflineHours: 24
        }
      };

      // Save configuration
      await fs.writeFile(CONFIG_FILE, JSON.stringify(config, null, 2));

      // Test initialization (don't keep it running)
      const node = new SovereignNode(config);
      
      // Quick validation
      const status = node.getStatus();
      
      spinner.succeed(`Node ${node.id} initialized successfully!`);

      console.log(chalk.green('\n✅ Configuration saved:'));
      console.log(`   ID: ${chalk.cyan(node.id)}`);
      console.log(`   Region: ${chalk.cyan(config.region)}`);
      console.log(`   Coordinates: ${chalk.cyan(`${config.coordinates.lat}, ${config.coordinates.lng}`)}`);
      console.log(`   Privacy: ${chalk.cyan(answers.privacy ? 'SGP-001 enabled' : 'Disabled')}`);
      console.log(`   Type: ${chalk.cyan(answers.nodeType)}`);
      console.log(`   Config: ${chalk.cyan(CONFIG_FILE)}`);

      console.log(chalk.yellow('\n📋 Next steps:'));
      console.log(`   1. Start node: ${chalk.bold('sovereignmap start')}`);
      console.log(`   2. Check status: ${chalk.bold('sovereignmap status')}`);
      console.log(`   3. View logs: ${chalk.bold('tail -f ~/.sovereignmap/logs/node.log')}`);

    } catch (error) {
      spinner.fail(`Initialization failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
      process.exit(1);
    }
  });

program
  .command('start')
  .description('Start the Genesis Node')
  .option('-d, --daemon', 'Run as daemon (background)')
  .option('--debug', 'Enable debug logging')
  .action(async (options) => {
    console.log(chalk.blue.bold('🚀 Starting Sovereign Map Node...\n'));

    // Load configuration
    let config;
    try {
      const configData = await fs.readFile(CONFIG_FILE, 'utf-8');
      config = JSON.parse(configData);
    } catch (error) {
      console.error(chalk.red('❌ No configuration found. Run `sovereignmap init` first.'));
      process.exit(1);
    }

    if (options.debug) {
      config.logging = { level: 'debug', pretty: true };
    }

    const node = new SovereignNode(config);

    // Setup logging
    const logFile = path.join(CONFIG_DIR, 'logs', 'node.log');
    await fs.mkdir(path.dirname(logFile), { recursive: true });

    // Handle events
    node.on('stateChange', ({ from, to }) => {
      console.log(chalk.gray(`State: ${from} → ${chalk.bold(to)}`));
    });

    node.on('connectivityLost', () => {
      console.log(chalk.yellow('⚠️  Network connectivity lost - entering Island Mode'));
    });

    node.on('connectivityRestored', (stats) => {
      console.log(chalk.green(`✅ Connectivity restored - synced ${stats.updatesSynced} updates`));
    });

    node.on('rewardEarned', ({ amount, reason }) => {
      console.log(chalk.green(`💰 Reward earned: ${amount} (${reason})`));
    });

    node.on('byzantineFaultDetected', ({ nodeId, faultType }) => {
      console.log(chalk.red(`🛡️  Byzantine fault detected: ${nodeId} (${faultType})`));
    });

    try {
      await node.initialize();

      console.log(chalk.green.bold('\n✅ Node is operational!'));
      console.log(chalk.gray(`   Node ID: ${node.id}`));
      console.log(chalk.gray(`   State: ${node.state}`));
      console.log(chalk.gray(`   Press Ctrl+C to stop\n`));

      // Keep running
      process.stdin.resume();

      // Graceful shutdown
      process.on('SIGINT', async () => {
        console.log(chalk.yellow('\n\n🛑 Shutting down...'));
        await node.shutdown();
        console.log(chalk.green('Goodbye!'));
        process.exit(0);
      });

    } catch (error) {
      console.error(chalk.red(`\n❌ Failed to start: ${error instanceof Error ? error.message : 'Unknown error'}`));
      process.exit(1);
    }
  });

program
  .command('status')
  .description('Check node status')
  .option('-j, --json', 'Output as JSON')
  .action(async (options) => {
    try {
      // This would connect to running node via IPC or API
      // For now, show configuration status
      
      const configData = await fs.readFile(CONFIG_FILE, 'utf-8').catch(() => null);
      
      if (!configData) {
        console.log(chalk.yellow('⚠️  No node configured. Run `sovereignmap init` first.'));
        return;
      }

      const config = JSON.parse(configData);

      if (options.json) {
        console.log(JSON.stringify(config, null, 2));
        return;
      }

      console.log(chalk.blue.bold('📊 Node Configuration\n'));
      console.log(`Node ID: ${chalk.cyan(config.nodeId)}`);
      console.log(`Region: ${chalk.cyan(config.region)}`);
      console.log(`Coordinates: ${chalk.cyan(`${config.coordinates.lat}, ${config.coordinates.lng}`)}`);
      console.log(`Privacy: ${chalk.cyan(config.privacyBudget ? `SGP-001 (ε=${config.privacyBudget.epsilon})` : 'Disabled')}`);
      console.log(`TPM: ${chalk.cyan(config.hardware.tpmAvailable ? 'Enabled' : 'Disabled')}`);
      console.log(`Aggregator: ${chalk.cyan(config.consensus.aggregatorTier)}`);
      console.log(`Island Mode: ${chalk.cyan(config.islandMode.enabled ? 'Enabled' : 'Disabled')}`);

      // Check if node is running (simplified)
      console.log(chalk.gray('\nNote: Use `sovereignmap start` to run the node and see live status.'));

    } catch (error) {
      console.error(chalk.red('Error reading status:', error));
    }
  });

program
  .command('test')
  .description('Run BFT test simulation')
  .option('-n, --nodes <count>', 'Number of nodes', '200')
  .option('-b, --byzantine <ratio>', 'Byzantine ratio', '0.555')
  .action(async (options) => {
    console.log(chalk.blue.bold(`🧪 Running BFT Test Simulation\n`));
    console.log(`Nodes: ${options.nodes}`);
    console.log(`Byzantine ratio: ${options.byzantine} (${Math.floor(parseInt(options.nodes) * parseFloat(options.byzantine))} nodes)`);
    console.log(chalk.gray('\nThis would run the full 200-node test suite.'));
    console.log(chalk.gray('For actual testing, use: npm run test-200-bft'));
  });

program.parse();
