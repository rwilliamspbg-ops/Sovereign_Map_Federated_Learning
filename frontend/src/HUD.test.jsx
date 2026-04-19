import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import HUD from './HUD';

const okResponse = (payload) => ({
  ok: true,
  status: 200,
  json: async () => payload,
});

const baseProps = {
  apiBase: '/backend',
  hudData: { simulation_counters: {} },
  health: { telemetry: { api_latency_ms: 24, api_error_rate: 0.08, ingress_mbps: 12 } },
  metricsSummary: { federated_learning: { current_round: 3, current_loss: 0.8 } },
  interactionSummary: null,
  interactionHistory: [],
  trustStatus: { fl_verification: { average_confidence_bps: 8425 }, trust_mode: 'governed-verification' },
  policyHistory: [],
  founders: [],
  voiceQuery: '',
  voiceResponse: '',
  policyDraft: { reject_on_verification_failure: false, min_confidence_bps: 7000 },
  policyToken: '',
  policyRole: 'admin',
  policyMessage: '',
  enclaveActionMessage: '',
  trainingStatus: { status: 'idle', active: false, round: 0, target_rounds: 10, current_metrics: { loss: 0.8 } },
  opsHealth: {
    status: 'ok',
    telemetry: { api_latency_ms: 24, api_error_rate: 0.08, ingress_mbps: 12 },
    privacy_security: { llm_policy_valid: 6, llm_policy_rejected: 1 },
    mapping: { coverage_pct: 97 },
    digital_twin: { lag_ms: 120, risk_score: 0.18 },
    sensors: {},
    alerts: [],
    dependencies: { prometheus: { reachable: true } },
    ports: { api_8000: true, flower_8080: true, prometheus_9090: true },
  },
  opsTrends: { api_latency_ms: [], api_error_rate_pct: [], ingress_mbps: [] },
  opsEvents: [],
  opsStreamStatus: 'connected',
  connectionDiagnostics: { apiBase: '/backend', eventsEndpoint: '/backend/ops/events', policyEndpoint: '/backend/ops/policy', policyStatus: 'ok' },
  webMetrics: {},
  loading: false,
  error: null,
  onTriggerFLRound: vi.fn(),
  onStartTraining: vi.fn(),
  onStopTraining: vi.fn(),
  onCreateEnclave: vi.fn(),
  onTriggerSimulation: vi.fn(),
  onSubmitVoiceQuery: vi.fn(),
  onPolicyChange: vi.fn(),
  onPolicyTokenChange: vi.fn(),
  onPolicyRoleChange: vi.fn(),
  onSubmitVerificationPolicy: vi.fn(),
  setVoiceQuery: vi.fn(),
};

describe('HUD AI interaction console', () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    fetchMock.mockImplementation((url) => {
      if (String(url).includes('/ai/interaction/decision')) {
        return Promise.resolve(okResponse({ status: 'recorded' }));
      }
      return Promise.reject(new Error(`Unexpected URL ${url}`));
    });
    global.fetch = fetchMock;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders review history and opens the interaction drawer', async () => {
    render(
      <HUD
        {...baseProps}
        interactionSummary={{
          quick_actions: [
            {
              id: 'start-training-10',
              label: 'Start 10-round training',
              command: 'start_training',
              kind: 'control_action',
              parameters: { rounds: 10 },
              requires_confirmation: true,
              undo_command: 'stop_training',
            },
          ],
          recommendations: [
            {
              action: 'hold_position_monitor',
              label: 'Hold position and observe',
              reason: 'system is stable and within guardrails',
              confidence: 0.91,
              risk: 0.08,
              expected_gain: 0.45,
            },
          ],
          intent_examples: ['start training for 10 rounds'],
          model_route: 'summary',
          context: { safety_state: 'safe' },
        }}
        interactionHistory={[
          {
            review_id: 'review-1',
            decision: 'approve',
            action_label: 'Start 10-round training',
            reason: 'operator approved training window',
          },
        ]}
      />
    );

    expect(await screen.findByText('AI Interaction Console')).toBeInTheDocument();
    const historyArticle = screen.getByRole('heading', { name: 'Decision History' }).closest('article');
    expect(within(historyArticle).getAllByText('approve').length).toBeGreaterThan(0);
    expect(within(historyArticle).getByText('Start 10-round training')).toBeInTheDocument();

    const quickActionsSection = screen.getByRole('heading', { name: 'Quick Actions' }).closest('article');
    fireEvent.click(within(quickActionsSection).getByRole('button', { name: 'Review' }));

    expect(await screen.findByText('Review Drawer')).toBeInTheDocument();
  });

  it('approves a reviewed action and records the decision', async () => {
    render(
      <HUD
        {...baseProps}
        interactionSummary={{
          quick_actions: [
            {
              id: 'start-training-10',
              label: 'Start 10-round training',
              command: 'start_training',
              kind: 'control_action',
              parameters: { rounds: 10 },
              requires_confirmation: true,
              undo_command: 'stop_training',
            },
          ],
          recommendations: [],
          intent_examples: ['start training for 10 rounds'],
          model_route: 'summary',
          context: { safety_state: 'safe' },
        }}
      />
    );

    const quickActionsSection = screen.getByRole('heading', { name: 'Quick Actions' }).closest('article');
    fireEvent.click(within(quickActionsSection).getByRole('button', { name: 'Review' }));
    fireEvent.change(screen.getByLabelText('Decision reason'), { target: { value: 'fits the current maintenance window' } });
    const reviewArticle = screen.getByRole('heading', { name: 'Review Drawer' }).closest('article');
    fireEvent.click(within(reviewArticle).getByRole('button', { name: 'Approve' }));

    await waitFor(() => {
      expect(baseProps.onStartTraining).toHaveBeenCalledWith(
        10,
        expect.objectContaining({
          decision: 'approve',
          source_action_id: 'start-training-10',
        })
      );
      expect(fetchMock).toHaveBeenCalledWith(
        '/backend/ai/interaction/decision',
        expect.objectContaining({ method: 'POST' })
      );
    });

    const decisionRequest = fetchMock.mock.calls.find(([url]) => String(url).includes('/ai/interaction/decision'));
    expect(decisionRequest).toBeTruthy();
    const [, decisionOptions] = decisionRequest;
    const decisionBody = JSON.parse(decisionOptions.body);
    expect(decisionBody.decision).toBe('approve');
    expect(decisionBody.action_label).toBe('Start 10-round training');
    expect(decisionBody.reason).toBe('fits the current maintenance window');
  });

  it('supports undo for reversible reviewed actions', async () => {
    render(
      <HUD
        {...baseProps}
        interactionSummary={{
          quick_actions: [
            {
              id: 'start-training-10',
              label: 'Start 10-round training',
              command: 'start_training',
              kind: 'control_action',
              parameters: { rounds: 10 },
              requires_confirmation: true,
              undo_command: 'stop_training',
            },
          ],
          recommendations: [],
          intent_examples: ['start training for 10 rounds'],
          model_route: 'summary',
          context: { safety_state: 'safe' },
        }}
      />
    );

    const quickActionsSection = screen.getByRole('heading', { name: 'Quick Actions' }).closest('article');
    fireEvent.click(within(quickActionsSection).getByRole('button', { name: 'Review' }));
    const reviewArticle = screen.getByRole('heading', { name: 'Review Drawer' }).closest('article');
    fireEvent.click(within(reviewArticle).getByRole('button', { name: 'Undo' }));

    await waitFor(() => {
      expect(baseProps.onStopTraining).toHaveBeenCalled();
    });

    const decisionRequest = fetchMock.mock.calls.find(([url]) => String(url).includes('/ai/interaction/decision'));
    expect(decisionRequest).toBeTruthy();
    const [, decisionOptions] = decisionRequest;
    const decisionBody = JSON.parse(decisionOptions.body);
    expect(decisionBody.decision).toBe('undo');
  });

  it('filters decision history and renders replay detail context', async () => {
    render(
      <HUD
        {...baseProps}
        interactionSummary={{
          quick_actions: [],
          recommendations: [],
          intent_examples: ['show planner insights'],
          model_route: 'planner',
          context: { safety_state: 'safe' },
        }}
        interactionHistory={[
          {
            review_id: 'review-a',
            ts: 1710000001,
            decision: 'approve',
            model_route: 'planner',
            action_label: 'Replan risky corridor',
            reason: 'maintenance window fit',
            prompt: 'replan corridor alpha',
          },
          {
            review_id: 'review-b',
            ts: 1710000002,
            decision: 'reject',
            model_route: 'summary',
            action_label: 'Reduce replanning rate',
            reason: 'latency stabilized',
            prompt: 'hold replanning cadence',
          },
        ]}
      />
    );

    const historyArticle = screen.getByRole('heading', { name: 'Decision History' }).closest('article');
    fireEvent.change(within(historyArticle).getByLabelText('Search decisions'), { target: { value: 'maintenance' } });
    expect(within(historyArticle).getByText('Replan risky corridor')).toBeInTheDocument();
    expect(within(historyArticle).queryByText('Reduce replanning rate')).not.toBeInTheDocument();

    fireEvent.change(within(historyArticle).getByLabelText('Decision filter'), { target: { value: 'reject' } });
    expect(within(historyArticle).getByText('No interaction decisions match the current filters.')).toBeInTheDocument();

    fireEvent.change(within(historyArticle).getByLabelText('Decision filter'), { target: { value: 'all' } });

    const replayArticle = screen.getByRole('heading', { name: 'Twin Replay Context' }).closest('article');
    fireEvent.click(within(replayArticle).getAllByRole('button')[0]);
    expect(within(replayArticle).getByText(/reason maintenance window fit/i)).toBeInTheDocument();
    expect(within(replayArticle).getByText(/prompt replan corridor alpha/i)).toBeInTheDocument();
  });

});
