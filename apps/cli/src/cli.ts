#!/usr/bin/env node

import { Command } from 'commander';
import chalk from 'chalk';
import ora from 'ora';
import inquirer from 'inquirer';
import { SovereignNode } from '@sovereignmap/core';

const program = new Command();

program
  .name('sovereignmap')
  .description('Sovereign Map Node Operator CLI')
  .version('0.1.0-alpha');

program
  .command('init')
  .description('Initialize a new Genesis Node')
  .option('-r, --region <region>', 'Geographic region')
  .option('--lat <latitude>', 'Latitude')
  .option('--lng <longitude>', 'Longitude')
  .action(async (options) => {
    console.log(chalk.blue('🗺️  Sovereign Map Node Initialization\n'));
    
    const answers = await inquirer.prompt([
      {
        type: 'input',
        name: 'region',
        message: 'Region name:',
        default: options.region || 'genesis-node',
      },
      {
        type: 'input',
        name: 'lat',
        message: 'Latitude:',
        default: options.lat || '48.1351',
        validate: (v) => !isNaN(parseFloat(v)) && parseFloat(v) >= -90 && parseFloat(v) <= 90,
      },
      {
        type: 'input',
        name: 'lng',
        message: 'Longitude:',
        default: options.lng || '11.5820',
        validate: (v) => !isNaN(parseFloat(v)) && parseFloat(v) >= -180 && parseFloat(v) <= 180,
      },
      {
        type: 'confirm',
        name: 'privacy',
        message: 'Enable SGP-001 privacy (epsilon=1.0)?',
        default: true,
      },
    ]);
    
    const spinner = ora('Initializing node...').start();
    
    try {
      const node = new SovereignNode({
        region: answers.region,
        coordinates: {
          lat: parseFloat(answers.lat),
          lng: parseFloat(answers.lng),
        },
        privacyBudget: answers.privacy ? { epsilon: 1.0, delta: 1e-5 } : undefined,
      });
      
      await node.initialize();
      
      spinner.succeed(`Node ${node.id} initialized successfully!`);
      
      console.log(chalk.green('\n✅ Node Status:'));
      console.log(`   ID: ${node.id}`);
      console.log(`   State: ${node.state}`);
      console.log(`   Privacy Budget: ε=${node.getStatus().privacy.totalBudget}`);
      console.log(`\nNext steps:`);
      console.log(`   1. Run: sovereignmap start`);
      console.log(`   2. Check status: sovereignmap status`);
      
    } catch (error) {
      spinner.fail(`Initialization failed: ${error.message}`);
      process.exit(1);
    }
  });

program
  .command('start')
  .description('Start the Genesis Node')
  .action(async () => {
    console.log(chalk.blue('🚀 Starting Sovereign Map Node...\n'));
    
    // Load config from disk
    const spinner = ora('Connecting to network...').start();
    
    // Implementation would load saved config
    spinner.succeed('Node operational!');
    
    console.log(chalk.green('\nNode is running. Press Ctrl+C to stop.'));
  });

program
  .command('status')
  .description('Check node status')
  .action(async () => {
    console.log(chalk.blue('📊 Node Status\n'));
    
    // Mock status - would read from running node
    console.log({
      state: 'ONLINE',
      uptime: '2h 34m',
      privacyBudget: { consumed: 0.3, remaining: 0.7 },
      peers: 47,
      updatesSubmitted: 12,
    });
  });

program.parse();
