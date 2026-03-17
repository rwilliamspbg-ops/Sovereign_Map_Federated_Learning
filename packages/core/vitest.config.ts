import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'node',
    include: ['src/**/*.test.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      include: ['src/**/*.ts'],
      exclude: ['src/**/*.d.ts', 'src/ambient.d.ts', 'src/wasm/**/*.ts'],
      thresholds: {
        lines: 35,
        functions: 30,
        statements: 35,
        branches: 30
      }
    }
  }
});
