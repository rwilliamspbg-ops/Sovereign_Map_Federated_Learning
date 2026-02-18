/**
 * Structured logging for Sovereign Map SDK
 */

import pino, { Logger as PinoLogger, LoggerOptions } from 'pino';

export type Logger = PinoLogger;

export interface LogConfig {
  level?: 'trace' | 'debug' | 'info' | 'warn' | 'error';
  pretty?: boolean;
  nodeId?: string;
}

export function createLogger(config: LogConfig = {}): Logger {
  const options: LoggerOptions = {
    level: config.level || 'info',
    base: {
      sdk: '@sovereignmap/core',
      version: '0.1.0-alpha.1',
      node: config.nodeId
    },
    timestamp: pino.stdTimeFunctions.isoTime,
    formatters: {
      level: (label) => ({ level: label }),
      bindings: (bindings) => ({
        pid: bindings.pid,
        node: config.nodeId
      })
    }
  };

  if (config.pretty) {
    return pino({
      ...options,
      transport: {
        target: 'pino-pretty',
        options: {
          colorize: true,
          translateTime: 'HH:MM:ss Z',
          ignore: 'pid,hostname'
        }
      }
    });
  }

  return pino(options);
}
