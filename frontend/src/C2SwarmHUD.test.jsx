import React from 'react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import C2SwarmHUD from './C2SwarmHUD';

const okResponse = (payload) => ({
  ok: true,
  status: 200,
  json: async () => payload
});

describe('C2SwarmHUD', () => {
  const fetchMock = vi.fn();
  let _originalFetch;

  beforeEach(() => {
    _originalFetch = global.fetch;
    fetchMock.mockImplementation((url) => {
      if (url.includes('/swarm/status')) {
        return Promise.resolve(okResponse({
          status: 'ok',
          nodes_active: 24,
          coverage_pct: 92.1,
          avg_latency_ms: 28.4,
          error_rate_pct: 0.17
        }));
      }
      if (url.includes('/swarm/map')) {
        return Promise.resolve(okResponse({
          node_count: 3,
          nodes: [
            { id: 'node-0001', status: 'ok', position: { x: 10, y: 20 } },
            { id: 'node-0002', status: 'ok', position: { x: 40, y: 50 } },
            { id: 'node-0003', status: 'warning', position: { x: 70, y: 80 } }
          ],
          paths: [{ id: 'path-1', from: 'node-0001', to: 'node-0002' }],
          risk_zones: [{ id: 'risk-a', x: 30, y: 30, radius: 5 }],
          coverage: { percent: 88.9 }
        }));
      }
      if (url.includes('/swarm/commands?')) {
        return Promise.resolve(okResponse({
          commands: [{ action_id: 'swarm-1', ts: 1730000000, command: 'hold_position', target_scope: 'global', target_ids: [], status: 'accepted' }]
        }));
      }
      if (url.includes('/swarm/audit/recent')) {
        return Promise.resolve(okResponse({
          audits: [
            {
              ts: 1730000000,
              action_id: 'swarm-1',
              command: 'hold_position',
              role: 'admin',
              signature: 'abc123def456'
            }
          ]
        }));
      }
      if (url.includes('/swarm/command')) {
        return Promise.resolve({
          ok: true,
          status: 202,
          json: async () => ({
            action_id: 'swarm-abc123',
            command: 'hold_position',
            normalized: { command: 'hold_position' }
          })
        });
      }
      return Promise.reject(new Error(`Unexpected URL ${url}`));
    });
    global.fetch = fetchMock;
  });

  afterEach(() => {
    global.fetch = _originalFetch;
    vi.restoreAllMocks();
  });

  it('renders status, map, and command log', async () => {
    render(<C2SwarmHUD apiBase="/backend" />);

    expect(await screen.findByText('Autonomous Swarm C2')).toBeInTheDocument();
    expect(await screen.findByText('24')).toBeInTheDocument();
    expect(await screen.findByText('3 rendered nodes')).toBeInTheDocument();
    expect(await screen.findByLabelText('command-log')).toBeInTheDocument();
    expect(await screen.findByLabelText('audit-log')).toBeInTheDocument();
  });

  it('renders shared AI interaction actions and recommendations', async () => {
    render(
      <C2SwarmHUD
        apiBase="/backend"
        interactionSummary={{
          quick_actions: [
            {
              id: 'ask-twin',
              label: 'Ask for twin summary',
              prompt: 'summarize the digital twin status and top risks',
              kind: 'assistant_query',
              model_route: 'summary',
              requires_confirmation: false
            }
          ],
          recommendations: [
            {
              action: 'hold_position_monitor',
              label: 'Hold position and observe',
              reason: 'system is stable and within guardrails',
              confidence: 0.91,
              risk: 0.08,
              expected_gain: 0.45
            }
          ]
        }}
      />
    );

    expect(await screen.findByRole('button', { name: 'Ask for twin summary' })).toBeInTheDocument();
    expect(screen.getByText(/Hold position and observe: system is stable and within guardrails/)).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Ask for twin summary' }));

    await waitFor(() => {
      expect(screen.getByText(/Assistant prompt loaded: summarize the digital twin status and top risks/)).toBeInTheDocument();
    });
  });

  it('submits command with auth header', async () => {
    render(<C2SwarmHUD apiBase="/backend" />);

    await screen.findByText('Autonomous Swarm C2');
    fireEvent.change(screen.getByPlaceholderText('Bearer token'), { target: { value: 'token-123' } });
    fireEvent.change(screen.getByLabelText('Nonce'), { target: { value: 'nonce-1' } });

    fireEvent.click(screen.getByRole('button', { name: /submit command/i }));

    await waitFor(() => {
      expect(screen.getByText(/Accepted hold_position/)).toBeInTheDocument();
    });

    const commandCall = fetchMock.mock.calls.find(([url]) => url.endsWith('/swarm/command'));
    expect(commandCall).toBeTruthy();
    const [, commandRequest] = commandCall;
    expect(commandRequest.method).toBe('POST');
    expect(commandRequest.headers.Authorization).toBe('Bearer token-123');
    expect(commandRequest.headers['X-API-Role']).toBe('admin');

    const body = JSON.parse(commandRequest.body);
    expect(body.command).toBe('hold_position');
    expect(body.client_nonce).toBe('nonce-1');
  });
});
