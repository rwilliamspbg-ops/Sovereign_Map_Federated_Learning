import { describe, expect, it, vi } from 'vitest';

const pinoSpy = vi.hoisted(() =>
  Object.assign(vi.fn((opts?: Record<string, unknown>) => ({ opts, child: vi.fn() })), {
    stdTimeFunctions: { isoTime: vi.fn(() => 'iso') }
  })
);

vi.mock('pino', () => ({
  default: pinoSpy,
  stdTimeFunctions: pinoSpy.stdTimeFunctions
}));

describe('createLogger', () => {
  it('creates logger with standard options', async () => {
    const { createLogger } = await import('./logger.js');
    const logger = createLogger({ level: 'debug', nodeId: 'n1' });

    expect(logger).toBeTruthy();
    expect(pinoSpy).toHaveBeenCalled();
    const options = pinoSpy.mock.calls.at(-1)?.[0] as Record<string, unknown>;
    expect(options.level).toBe('debug');
    expect((options.base as Record<string, unknown>).node).toBe('n1');
  });

  it('creates pretty logger transport when pretty=true', async () => {
    const { createLogger } = await import('./logger.js');
    createLogger({ pretty: true, nodeId: 'n2' });

    const options = pinoSpy.mock.calls.at(-1)?.[0] as Record<string, unknown>;
    expect((options.transport as Record<string, unknown>).target).toBe('pino-pretty');
  });
});
