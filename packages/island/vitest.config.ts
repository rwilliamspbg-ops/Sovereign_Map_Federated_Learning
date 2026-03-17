import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'node',
    include: ['**/*.test.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      include: ['index.ts'],
      thresholds: {
        lines: 25,
        functions: 25,
        statements: 25,
        branches: 20
      }
    }
  }
});
