import { useEffect, useMemo, useRef, useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  Area,
  AreaChart,
  Legend
} from 'recharts';
import './BrowserFLDemo.css';

const MAX_POINTS = 120;
const DEFAULT_ROUNDS = 50;
const API_BASE = import.meta.env.VITE_HUD_API_BASE || (import.meta.env.DEV ? '/backend' : 'http://localhost:8000');
const TRAINING_API_BASE = API_BASE;

function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}

function randomFloat(min, max) {
  return min + Math.random() * (max - min);
}

function envInt(name, fallback) {
  const raw = import.meta.env?.[name];
  const parsed = Number(raw);
  if (!Number.isFinite(parsed) || parsed < 0) return fallback;
  return Math.floor(parsed);
}

function envBool(name, fallback) {
  const raw = import.meta.env?.[name];
  if (raw === undefined || raw === null || raw === '') return fallback;
  const value = String(raw).trim().toLowerCase();
  if (['1', 'true', 'yes', 'on'].includes(value)) return true;
  if (['0', 'false', 'no', 'off'].includes(value)) return false;
  return fallback;
}

function envFloat(name, fallback) {
  const raw = import.meta.env?.[name];
  const parsed = Number(raw);
  if (!Number.isFinite(parsed)) return fallback;
  return parsed;
}

function ema(previous, next, alpha = 0.35) {
  if (!Number.isFinite(next)) return previous;
  if (!Number.isFinite(previous)) return next;
  return (alpha * next) + ((1 - alpha) * previous);
}

const IMPACT_DEFAULT_COOLDOWN_SECONDS = envInt('VITE_IMPACT_COOLDOWN_DEFAULT_SECONDS', 0);
const IMPACT_RISKY_COOLDOWN_SECONDS = envInt('VITE_IMPACT_COOLDOWN_RISKY_SECONDS', 3);
const IMPACT_REQUIRE_TYPED_RISKY = envBool('VITE_IMPACT_REQUIRE_TYPED_RISKY', true);
const UI_ACC_ALPHA = envFloat('VITE_UI_ACC_SMOOTHING_ALPHA', 0.35);
const UI_LOSS_ALPHA = envFloat('VITE_UI_LOSS_SMOOTHING_ALPHA', 0.3);
const UI_COMP_ALPHA = envFloat('VITE_UI_COMP_SMOOTHING_ALPHA', 0.35);
const UI_LAT_ALPHA = envFloat('VITE_UI_LATENCY_SMOOTHING_ALPHA', 0.25);
const UI_BW_ALPHA = envFloat('VITE_UI_BANDWIDTH_SMOOTHING_ALPHA', 0.25);
const CHART_THROTTLE_MS = Math.max(50, envInt('VITE_CHART_THROTTLE_MS', 350));

const NETWORK_HELPER_TEXT = {
  en: 'Share trusted compute, request an invite, and help the federated network grow safely.',
  es: 'Comparte computo confiable, solicita invitacion y ayuda a crecer la red federada.',
  fr: 'Partagez un calcul fiable, demandez une invitation et aidez le reseau federatif a grandir.'
};

const IMPACT_GUARDS = {
  approve_invite_request: {
    cooldownSeconds: envInt('VITE_IMPACT_COOLDOWN_APPROVE_INVITE_SECONDS', IMPACT_DEFAULT_COOLDOWN_SECONDS),
  },
  reject_invite_request: {
    cooldownSeconds: envInt('VITE_IMPACT_COOLDOWN_REJECT_INVITE_SECONDS', IMPACT_DEFAULT_COOLDOWN_SECONDS),
  },
  submit_dispute: {
    cooldownSeconds: envInt('VITE_IMPACT_COOLDOWN_SUBMIT_DISPUTE_SECONDS', IMPACT_DEFAULT_COOLDOWN_SECONDS),
  },
  revoke_invite: {
    cooldownSeconds: envInt('VITE_IMPACT_COOLDOWN_REVOKE_INVITE_SECONDS', IMPACT_RISKY_COOLDOWN_SECONDS),
    requiredLabel: 'Type invite ID to confirm revoke',
    requiredText: ({ inviteId }) => (IMPACT_REQUIRE_TYPED_RISKY ? inviteId : null),
  },
  revoke_registration: {
    cooldownSeconds: envInt('VITE_IMPACT_COOLDOWN_REVOKE_REGISTRATION_SECONDS', IMPACT_RISKY_COOLDOWN_SECONDS),
    requiredLabel: 'Type node id to confirm revoke',
    requiredText: ({ nodeId }) => (IMPACT_REQUIRE_TYPED_RISKY ? `node-${nodeId}` : null),
  },
  release_escrow: {
    cooldownSeconds: envInt('VITE_IMPACT_COOLDOWN_RELEASE_ESCROW_SECONDS', IMPACT_RISKY_COOLDOWN_SECONDS),
    requiredLabel: 'Type contract id to confirm release',
    requiredText: ({ contractId }) => (IMPACT_REQUIRE_TYPED_RISKY ? contractId : null),
  },
};

function detectWalletLabel(provider) {
  if (!provider) return 'Injected Wallet';
  if (provider.isMetaMask) return 'MetaMask';
  if (provider.isCoinbaseWallet) return 'Coinbase Wallet';
  if (provider.isRabby) return 'Rabby';
  if (provider.isBraveWallet) return 'Brave Wallet';
  return 'Injected Wallet';
}

export default function BrowserFLDemo({ enableBackendMetrics = false, onMetricsUpdate = null }) {
  const [trainingMode, setTrainingMode] = useState('simulation');

  const [participants, setParticipants] = useState(120);
  const [localEpochs, setLocalEpochs] = useState(2);
  const [targetRounds, setTargetRounds] = useState(DEFAULT_ROUNDS);
  const [epsilon, setEpsilon] = useState(1.2);
  const [compressionBits, setCompressionBits] = useState(8);
  const [learningRate, setLearningRate] = useState(0.02);

  const [runtimeBackend, setRuntimeBackend] = useState('detecting');
  const [running, setRunning] = useState(false);
  const [round, setRound] = useState(0);

  const [accuracy, setAccuracy] = useState(0.51);
  const [loss, setLoss] = useState(1.7);
  const [compressionRatio, setCompressionRatio] = useState(4.0);
  const [latencyMs, setLatencyMs] = useState(0);
  const [bandwidthKB, setBandwidthKB] = useState(0);

  const [history, setHistory] = useState([]);
  const [realHistory, setRealHistory] = useState([]);

  const [usingBackendData, setUsingBackendData] = useState(false);

  const [phase3dStatus, setPhase3dStatus] = useState('idle');
  const [phase3dRound, setPhase3dRound] = useState(0);
  const [phase3dTotal, setPhase3dTotal] = useState(10);
  const [offerTitle, setOfferTitle] = useState('Regional Image Pack');
  const [offerModality, setOfferModality] = useState('image');
  const [offerPrice, setOfferPrice] = useState(12.5);
  const [offerQuality, setOfferQuality] = useState(0.8);
  const [intentTask, setIntentTask] = useState('classification');
  const [intentBudget, setIntentBudget] = useState(80);
  const [intentMinQuality, setIntentMinQuality] = useState(0.7);
  const [selectedIntentId, setSelectedIntentId] = useState('');
  const [marketplaceOffers, setMarketplaceOffers] = useState([]);
  const [marketplaceIntents, setMarketplaceIntents] = useState([]);
  const [marketplaceContracts, setMarketplaceContracts] = useState([]);
  const [marketplaceStatus, setMarketplaceStatus] = useState('idle');
  const [marketplaceLastError, setMarketplaceLastError] = useState('');
  const [marketplaceDisputes, setMarketplaceDisputes] = useState([]);
  const [governanceActions, setGovernanceActions] = useState([]);
  const [disputeReason, setDisputeReason] = useState('Insufficient update quality evidence');
  const [previewBudget, setPreviewBudget] = useState(80);
  const [previewMinQuality, setPreviewMinQuality] = useState(0.7);
  const [policyPreview, setPolicyPreview] = useState(null);
  const [governanceProposals, setGovernanceProposals] = useState([]);
  const [proposalTitle, setProposalTitle] = useState('Tighten image quality floor');
  const [proposalType, setProposalType] = useState('policy_update');
  const [proposalDescription, setProposalDescription] = useState('Raise min_quality_score to 0.8 for image intents');
  const [selectedProposalId, setSelectedProposalId] = useState('');
  const [voteDecision, setVoteDecision] = useState('yes');
  const [voteWeight, setVoteWeight] = useState(1);
  const [voteReason, setVoteReason] = useState('Aligned with trust policy');
  const [networkSummary, setNetworkSummary] = useState(null);
  const [attestationFeed, setAttestationFeed] = useState([]);
  const [attestParticipantName, setAttestParticipantName] = useState('community-node');
  const [attestComputeType, setAttestComputeType] = useState('gpu');
  const [attestComputeCapacity, setAttestComputeCapacity] = useState('1x A100 40GB');
  const [attestStatus, setAttestStatus] = useState('verified');
  const [attestRegion, setAttestRegion] = useState('global');
  const [attestationSortBy, setAttestationSortBy] = useState('reputation');
  const [helperLanguage, setHelperLanguage] = useState('en');
  const [inviteRequestName, setInviteRequestName] = useState('new-community-node');
  const [inviteRequestEmail, setInviteRequestEmail] = useState('community@example.org');
  const [inviteRequestComputeType, setInviteRequestComputeType] = useState('gpu');
  const [inviteRequestRegion, setInviteRequestRegion] = useState('global');
  const [inviteRequestMotivation, setInviteRequestMotivation] = useState('Contribute verified compute to classification rounds');
  const [walletProviders, setWalletProviders] = useState([]);
  const [selectedWalletProviderId, setSelectedWalletProviderId] = useState('');
  const [walletAddress, setWalletAddress] = useState('');
  const [walletChainId, setWalletChainId] = useState('');
  const [walletStatus, setWalletStatus] = useState('wallet not connected');
  const [walletError, setWalletError] = useState('');
  const [adminToken, setAdminToken] = useState('');
  const [adminAuthMethods, setAdminAuthMethods] = useState(null);
  const [adminInviteRequests, setAdminInviteRequests] = useState([]);
  const [adminInvites, setAdminInvites] = useState([]);
  const [adminRegistrations, setAdminRegistrations] = useState([]);
  const [adminInviteRequestsTotal, setAdminInviteRequestsTotal] = useState(0);
  const [adminInvitesTotal, setAdminInvitesTotal] = useState(0);
  const [adminRegistrationsTotal, setAdminRegistrationsTotal] = useState(0);
  const [adminStatus, setAdminStatus] = useState('idle');
  const [adminError, setAdminError] = useState('');
  const [adminInviteParticipantName, setAdminInviteParticipantName] = useState('admin-manual-node');
  const [adminInviteMaxUses, setAdminInviteMaxUses] = useState(1);
  const [adminInviteExpiresHours, setAdminInviteExpiresHours] = useState(24);
  const [rejectReason, setRejectReason] = useState('Not enough validation signals');
  const [adminRequestQuery, setAdminRequestQuery] = useState('');
  const [adminInviteQuery, setAdminInviteQuery] = useState('');
  const [adminRegistrationQuery, setAdminRegistrationQuery] = useState('');
  const [adminRequestStatusFilter, setAdminRequestStatusFilter] = useState('pending');
  const [adminInviteStatusFilter, setAdminInviteStatusFilter] = useState('active');
  const [adminRegistrationStatusFilter, setAdminRegistrationStatusFilter] = useState('all');
  const [adminRequestSortBy, setAdminRequestSortBy] = useState('created_at');
  const [adminRequestSortDir, setAdminRequestSortDir] = useState('desc');
  const [adminInviteSortBy, setAdminInviteSortBy] = useState('created_at');
  const [adminInviteSortDir, setAdminInviteSortDir] = useState('desc');
  const [adminRegistrationSortBy, setAdminRegistrationSortBy] = useState('registered_at');
  const [adminRegistrationSortDir, setAdminRegistrationSortDir] = useState('desc');
  const [adminRequestPage, setAdminRequestPage] = useState(1);
  const [adminInvitePage, setAdminInvitePage] = useState(1);
  const [adminRegistrationPage, setAdminRegistrationPage] = useState(1);
  const [adminPageSize, setAdminPageSize] = useState(5);
  const [assistantMode, setAssistantMode] = useState('auto');
  const [impactPreview, setImpactPreview] = useState(null);
  const [impactConfirmInput, setImpactConfirmInput] = useState('');
  const [impactCooldownRemaining, setImpactCooldownRemaining] = useState(0);
  const smoothedRef = useRef({
    accuracy: 0.51,
    loss: 1.7,
    compressionRatio: 4.0,
    latencyMs: 0,
    bandwidthKB: 0
  });
  const chartThrottleRef = useRef({
    simulationLastEmit: 0,
    simulationPending: null,
    simulationTimer: null,
    realLastEmit: 0,
    realPending: null,
    realTimer: null
  });

  const commitChartPoint = (mode, point) => {
    const setSeries = mode === 'real' ? setRealHistory : setHistory;
    setSeries((previous) => {
      const merged = [...previous, point].slice(-MAX_POINTS);
      onMetricsUpdate?.(merged);
      return merged;
    });
  };

  const flushChartPoint = (mode) => {
    const pendingKey = mode === 'real' ? 'realPending' : 'simulationPending';
    const lastEmitKey = mode === 'real' ? 'realLastEmit' : 'simulationLastEmit';
    const timerKey = mode === 'real' ? 'realTimer' : 'simulationTimer';
    const pendingPoint = chartThrottleRef.current[pendingKey];

    chartThrottleRef.current[timerKey] = null;
    if (!pendingPoint) {
      return;
    }

    chartThrottleRef.current[pendingKey] = null;
    chartThrottleRef.current[lastEmitKey] = Date.now();
    commitChartPoint(mode, pendingPoint);
  };

  const pushChartPoint = (mode, point) => {
    const pendingKey = mode === 'real' ? 'realPending' : 'simulationPending';
    const lastEmitKey = mode === 'real' ? 'realLastEmit' : 'simulationLastEmit';
    const timerKey = mode === 'real' ? 'realTimer' : 'simulationTimer';
    const now = Date.now();
    const elapsed = now - chartThrottleRef.current[lastEmitKey];

    if (elapsed >= CHART_THROTTLE_MS && !chartThrottleRef.current[timerKey]) {
      chartThrottleRef.current[lastEmitKey] = now;
      commitChartPoint(mode, point);
      return;
    }

    chartThrottleRef.current[pendingKey] = point;
    if (!chartThrottleRef.current[timerKey]) {
      const delay = Math.max(0, CHART_THROTTLE_MS - elapsed);
      chartThrottleRef.current[timerKey] = setTimeout(() => {
        flushChartPoint(mode);
      }, delay);
    }
  };

  useEffect(() => {
    const detectWebGPU = async () => {
      try {
        if (typeof navigator !== 'undefined' && navigator.gpu) {
          const adapter = await navigator.gpu.requestAdapter();
          setRuntimeBackend(adapter ? 'webgpu' : 'cpu-fallback');
        } else {
          setRuntimeBackend('cpu-fallback');
        }
      } catch {
        setRuntimeBackend('cpu-fallback');
      }
    };

    detectWebGPU();
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const eth = window.ethereum;
    if (!eth) {
      setWalletProviders([]);
      return;
    }

    const providers = Array.isArray(eth.providers) ? eth.providers : [eth];
    const normalized = providers.map((provider, index) => ({
      id: `${detectWalletLabel(provider)}-${index}`,
      label: detectWalletLabel(provider),
      provider
    }));
    setWalletProviders(normalized);
    if (!selectedWalletProviderId && normalized[0]?.id) {
      setSelectedWalletProviderId(normalized[0].id);
    }
  }, [selectedWalletProviderId]);

  useEffect(() => {
    if (!enableBackendMetrics) return undefined;

    const fetchBackendMetrics = async () => {
      try {
        const response = await fetch(`${API_BASE}/metrics_summary`);
        setUsingBackendData(response.ok);
      } catch {
        setUsingBackendData(false);
      }
    };

    fetchBackendMetrics();
    const interval = setInterval(fetchBackendMetrics, 5000);
    return () => clearInterval(interval);
  }, [enableBackendMetrics]);

  useEffect(() => {
    const loadMarketplace = async () => {
      try {
        const [offersResp, intentsResp, contractsResp, disputesResp, actionsResp, proposalsResp, networkResp, attestationsResp] = await Promise.all([
          fetch(`${API_BASE}/marketplace/offers?status=active&limit=10`),
          fetch(`${API_BASE}/marketplace/round_intents?limit=10`),
          fetch(`${API_BASE}/marketplace/contracts`),
          fetch(`${API_BASE}/marketplace/disputes?limit=10`),
          fetch(`${API_BASE}/governance/actions?limit=10`),
          fetch(`${API_BASE}/governance/proposals?limit=10`),
          fetch(`${API_BASE}/network/expansion_summary`),
          fetch(`${API_BASE}/attestations/feed?sort_by=${attestationSortBy}&limit=8`)
        ]);

        if (offersResp.ok) {
          const payload = await offersResp.json();
          setMarketplaceOffers(Array.isArray(payload.offers) ? payload.offers : []);
        }
        if (intentsResp.ok) {
          const payload = await intentsResp.json();
          const intents = Array.isArray(payload.round_intents) ? payload.round_intents : [];
          setMarketplaceIntents(intents);
          if (!selectedIntentId) {
            const openIntent = intents.find((item) => item.status === 'open');
            if (openIntent?.round_intent_id) {
              setSelectedIntentId(openIntent.round_intent_id);
            }
          }
        }
        if (contractsResp.ok) {
          const payload = await contractsResp.json();
          setMarketplaceContracts(Array.isArray(payload.contracts) ? payload.contracts : []);
        }
        if (disputesResp.ok) {
          const payload = await disputesResp.json();
          setMarketplaceDisputes(Array.isArray(payload.disputes) ? payload.disputes : []);
        }
        if (actionsResp.ok) {
          const payload = await actionsResp.json();
          setGovernanceActions(Array.isArray(payload.actions) ? payload.actions : []);
        }
        if (proposalsResp.ok) {
          const payload = await proposalsResp.json();
          const proposals = Array.isArray(payload.proposals) ? payload.proposals : [];
          setGovernanceProposals(proposals);
          if (!selectedProposalId && proposals[0]?.proposal_id) {
            setSelectedProposalId(proposals[0].proposal_id);
          }
        }
        if (networkResp.ok) {
          const payload = await networkResp.json();
          setNetworkSummary(payload);
        }
        if (attestationsResp.ok) {
          const payload = await attestationsResp.json();
          setAttestationFeed(Array.isArray(payload.attestations) ? payload.attestations : []);
        }
      } catch {
        setMarketplaceStatus('marketplace unavailable');
      }
    };

    loadMarketplace();
    const interval = setInterval(loadMarketplace, 8000);
    return () => clearInterval(interval);
  }, [selectedIntentId, attestationSortBy]);

  useEffect(() => {
    if (trainingMode !== 'real' || phase3dStatus !== 'training') return undefined;

    const pollTrainingStatus = async () => {
      try {
        const response = await fetch(`${TRAINING_API_BASE}/training/status`);
        if (!response.ok) return;

        const status = await response.json();
        setPhase3dRound(status.current_round || 0);
        setPhase3dTotal((prev) => {
          const total = Number(status.total_rounds);
          return Number.isFinite(total) && total > 0 ? total : prev;
        });

        if (status.current_metrics) {
          const metric = status.current_metrics;
          const nextAccRaw = Number(metric.accuracy);
          const nextLossRaw = Number(metric.round_loss);
          const nextCompRaw = Number(metric.compression_ratio);

          const smoothedAccuracy = ema(smoothedRef.current.accuracy, nextAccRaw, UI_ACC_ALPHA);
          const smoothedLoss = ema(smoothedRef.current.loss, nextLossRaw, UI_LOSS_ALPHA);
          const smoothedCompression = ema(smoothedRef.current.compressionRatio, nextCompRaw, UI_COMP_ALPHA);

          smoothedRef.current.accuracy = Number.isFinite(smoothedAccuracy) ? smoothedAccuracy : smoothedRef.current.accuracy;
          smoothedRef.current.loss = Number.isFinite(smoothedLoss) ? smoothedLoss : smoothedRef.current.loss;
          smoothedRef.current.compressionRatio = Number.isFinite(smoothedCompression) ? smoothedCompression : smoothedRef.current.compressionRatio;

          setAccuracy(smoothedRef.current.accuracy);
          setLoss(smoothedRef.current.loss);
          setCompressionRatio(smoothedRef.current.compressionRatio);

          const point = {
            round: status.current_round || 0,
            accuracy: Number(smoothedRef.current.accuracy.toFixed(4)),
            loss: Number(smoothedRef.current.loss.toFixed(4)),
            compressionRatio: Number(smoothedRef.current.compressionRatio.toFixed(2)),
            latencyMs: Number.isFinite(smoothedRef.current.latencyMs)
              ? Math.round(smoothedRef.current.latencyMs)
              : latencyMs,
            bandwidthKB: Number.isFinite(smoothedRef.current.bandwidthKB)
              ? Number(smoothedRef.current.bandwidthKB.toFixed(1))
              : bandwidthKB
          };
          pushChartPoint('real', point);
        }

        if (status.status === 'completed') {
          setPhase3dStatus('completed');
        } else if (status.status === 'error') {
          setPhase3dStatus('idle');
        }
      } catch {
        setPhase3dStatus('idle');
      }
    };

    pollTrainingStatus();
    const interval = setInterval(pollTrainingStatus, 2000);
    return () => clearInterval(interval);
  }, [
    trainingMode,
    phase3dStatus,
    latencyMs,
    bandwidthKB,
    onMetricsUpdate
  ]);

  useEffect(() => {
    if (!running || trainingMode !== 'simulation') return undefined;

    const tick = () => {
      setRound((previousRound) => {
        const nextRound = previousRound + 1;

        const roundDecay = Math.exp(-nextRound / (targetRounds * 0.7));
        const participationGain = clamp(participants / 200, 0.4, 1.5);
        const epochGain = clamp(localEpochs / 2.5, 0.7, 1.6);
        const backendBoost = runtimeBackend === 'webgpu' ? 0.0045 : 0;

        const privacyPenalty = epsilon < 0.8 ? (0.8 - epsilon) * 0.006 : 0;
        const compressionPenalty = compressionBits < 8 ? (8 - compressionBits) * 0.0012 : 0;

        const gain =
          learningRate * 0.42 * roundDecay * participationGain * epochGain +
          backendBoost -
          privacyPenalty -
          compressionPenalty +
          randomFloat(-0.0015, 0.0015);

        const nextAccuracy = clamp(accuracy + gain, 0.45, 0.995);
        const nextLoss = clamp(loss * (1 - gain * 1.5 + randomFloat(-0.003, 0.002)), 0.05, 2.3);

        const quantRatio = 32 / compressionBits;
        const entropyBonus = compressionBits <= 8 ? 1.25 : 1.05;
        const nextCompressionRatio = clamp(quantRatio * entropyBonus, 1.2, 14);

        const bytesPerClient = 30000;
        const originalBytes = participants * bytesPerClient;
        const compressedBytes = originalBytes / nextCompressionRatio;
        const nextBandwidthKB = compressedBytes / 1024;

        const baseLatency = 130 + participants * 1.35 + localEpochs * 22;
        const compressionCost = compressionBits <= 4 ? 30 : compressionBits <= 8 ? 14 : 8;
        const backendMultiplier = runtimeBackend === 'webgpu' ? 0.63 : 1;
        const nextLatency = Math.round((baseLatency + compressionCost) * backendMultiplier);

        const smoothedLatency = ema(smoothedRef.current.latencyMs, nextLatency, UI_LAT_ALPHA);
        const smoothedBandwidth = ema(smoothedRef.current.bandwidthKB, nextBandwidthKB, UI_BW_ALPHA);

        smoothedRef.current.accuracy = nextAccuracy;
        smoothedRef.current.loss = nextLoss;
        smoothedRef.current.compressionRatio = nextCompressionRatio;
        smoothedRef.current.latencyMs = Number.isFinite(smoothedLatency) ? smoothedLatency : nextLatency;
        smoothedRef.current.bandwidthKB = Number.isFinite(smoothedBandwidth) ? smoothedBandwidth : nextBandwidthKB;

        setAccuracy(nextAccuracy);
        setLoss(nextLoss);
        setCompressionRatio(nextCompressionRatio);
        setLatencyMs(Math.round(smoothedRef.current.latencyMs));
        setBandwidthKB(smoothedRef.current.bandwidthKB);

        const point = {
          round: nextRound,
          accuracy: Number(nextAccuracy.toFixed(4)),
          loss: Number(nextLoss.toFixed(4)),
          latencyMs: Math.round(smoothedRef.current.latencyMs),
          bandwidthKB: Number(smoothedRef.current.bandwidthKB.toFixed(1)),
          compressionRatio: Number(nextCompressionRatio.toFixed(2))
        };
        pushChartPoint('simulation', point);

        if (nextRound >= targetRounds) {
          setRunning(false);
        }

        return nextRound;
      });
    };

    const timer = setInterval(tick, 700);
    return () => clearInterval(timer);
  }, [
    running,
    trainingMode,
    targetRounds,
    participants,
    localEpochs,
    epsilon,
    compressionBits,
    learningRate,
    runtimeBackend,
    accuracy,
    loss,
    onMetricsUpdate
  ]);

  const backendLabel = useMemo(() => {
    if (runtimeBackend === 'webgpu') return 'WebGPU Active';
    if (runtimeBackend === 'cpu-fallback') return 'CPU Fallback';
    return 'Detecting Runtime';
  }, [runtimeBackend]);

  const backendClass = runtimeBackend === 'webgpu' ? 'badge-webgpu' : 'badge-cpu';

  const activeRound = trainingMode === 'real' ? phase3dRound : round;
  const activeTarget = trainingMode === 'real' ? phase3dTotal : targetRounds;
  const activeHistory = trainingMode === 'real' ? realHistory : history;

  const progressPercent = clamp((activeRound / Math.max(activeTarget, 1)) * 100, 0, 100);
  const compressionSavedPercent = clamp((1 - 1 / compressionRatio) * 100, 0, 99.9);

  const resetScenario = () => {
    setRunning(false);
    setRound(0);
    setPhase3dRound(0);
    setAccuracy(0.51);
    setLoss(1.7);
    setCompressionRatio(4.0);
    setLatencyMs(0);
    setBandwidthKB(0);
    setHistory([]);
    setRealHistory([]);
    smoothedRef.current = {
      accuracy: 0.51,
      loss: 1.7,
      compressionRatio: 4.0,
      latencyMs: 0,
      bandwidthKB: 0
    };
    if (chartThrottleRef.current.simulationTimer) {
      clearTimeout(chartThrottleRef.current.simulationTimer);
    }
    if (chartThrottleRef.current.realTimer) {
      clearTimeout(chartThrottleRef.current.realTimer);
    }
    chartThrottleRef.current = {
      simulationLastEmit: 0,
      simulationPending: null,
      simulationTimer: null,
      realLastEmit: 0,
      realPending: null,
      realTimer: null
    };
    onMetricsUpdate?.([]);
  };

  useEffect(() => () => {
    if (chartThrottleRef.current.simulationTimer) {
      clearTimeout(chartThrottleRef.current.simulationTimer);
    }
    if (chartThrottleRef.current.realTimer) {
      clearTimeout(chartThrottleRef.current.realTimer);
    }
  }, []);

  const startPhase3dTraining = async () => {
    try {
      const config = {
        num_rounds: phase3dTotal,
        num_clients: participants,
        local_epochs: localEpochs,
        learning_rate: learningRate,
        epsilon,
        compression_bits: compressionBits,
        use_compression: true,
        use_privacy: true,
        dataset: 'cifar10',
        device: 'cuda',
        multi_gpu: true
      };

      const response = await fetch(`${TRAINING_API_BASE}/training/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        setPhase3dStatus('idle');
        return;
      }

      setPhase3dStatus('training');
      setRunning(false);
      setRound(0);
      setPhase3dRound(0);
      setRealHistory([]);
      onMetricsUpdate?.([]);
    } catch {
      setPhase3dStatus('idle');
    }
  };

  const cancelPhase3dTraining = async () => {
    try {
      await fetch(`${TRAINING_API_BASE}/training/stop`, { method: 'POST' });
      setPhase3dStatus('idle');
    } catch {
      setPhase3dStatus('idle');
    }
  };

  const refreshMarketplace = async () => {
    try {
      const [offersResp, intentsResp, contractsResp, disputesResp, actionsResp, proposalsResp, networkResp, attestationsResp] = await Promise.all([
        fetch(`${API_BASE}/marketplace/offers?status=active&limit=10`),
        fetch(`${API_BASE}/marketplace/round_intents?limit=10`),
        fetch(`${API_BASE}/marketplace/contracts`),
        fetch(`${API_BASE}/marketplace/disputes?limit=10`),
        fetch(`${API_BASE}/governance/actions?limit=10`),
        fetch(`${API_BASE}/governance/proposals?limit=10`),
        fetch(`${API_BASE}/network/expansion_summary`),
        fetch(`${API_BASE}/attestations/feed?sort_by=${attestationSortBy}&limit=8`)
      ]);
      if (offersResp.ok) {
        const payload = await offersResp.json();
        setMarketplaceOffers(Array.isArray(payload.offers) ? payload.offers : []);
      }
      if (intentsResp.ok) {
        const payload = await intentsResp.json();
        const intents = Array.isArray(payload.round_intents) ? payload.round_intents : [];
        setMarketplaceIntents(intents);
      }
      if (contractsResp.ok) {
        const payload = await contractsResp.json();
        setMarketplaceContracts(Array.isArray(payload.contracts) ? payload.contracts : []);
      }
      if (disputesResp.ok) {
        const payload = await disputesResp.json();
        setMarketplaceDisputes(Array.isArray(payload.disputes) ? payload.disputes : []);
      }
      if (actionsResp.ok) {
        const payload = await actionsResp.json();
        setGovernanceActions(Array.isArray(payload.actions) ? payload.actions : []);
      }
      if (proposalsResp.ok) {
        const payload = await proposalsResp.json();
        const proposals = Array.isArray(payload.proposals) ? payload.proposals : [];
        setGovernanceProposals(proposals);
        if (!selectedProposalId && proposals[0]?.proposal_id) {
          setSelectedProposalId(proposals[0].proposal_id);
        }
      }
      if (networkResp.ok) {
        const payload = await networkResp.json();
        setNetworkSummary(payload);
      }
      if (attestationsResp.ok) {
        const payload = await attestationsResp.json();
        setAttestationFeed(Array.isArray(payload.attestations) ? payload.attestations : []);
      }
    } catch {
      setMarketplaceStatus('marketplace refresh failed');
    }
  };

  const requestJoinInvite = async () => {
    setMarketplaceStatus('submitting invite request...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/join/request_invite`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          participant_name: inviteRequestName,
          contact_email: inviteRequestEmail,
          compute_type: inviteRequestComputeType,
          region: inviteRequestRegion,
          preferred_language: helperLanguage,
          motivation: inviteRequestMotivation
        })
      });

      if (response.ok) {
        setMarketplaceStatus('invite request submitted');
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('invite request failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('invite request failed');
      setMarketplaceLastError('network error');
    }
  };

  const adminHeaders = () => {
    const headers = { 'Content-Type': 'application/json' };
    if (adminToken) {
      headers['X-Join-Admin-Token'] = adminToken;
    }
    if (walletAddress) {
      headers['X-Admin-Wallet'] = walletAddress;
    }
    return headers;
  };

  const connectSelectedWallet = async () => {
    const selected = walletProviders.find((item) => item.id === selectedWalletProviderId);
    if (!selected?.provider) {
      setWalletError('no wallet provider selected');
      return;
    }
    setWalletStatus('connecting wallet...');
    setWalletError('');
    try {
      const accounts = await selected.provider.request({ method: 'eth_requestAccounts' });
      const chainId = await selected.provider.request({ method: 'eth_chainId' });
      const address = Array.isArray(accounts) && accounts[0] ? String(accounts[0]) : '';
      if (!address) {
        setWalletStatus('wallet connection failed');
        setWalletError('no account returned by provider');
        return;
      }
      setWalletAddress(address.toLowerCase());
      setWalletChainId(String(chainId || ''));
      setWalletStatus('wallet connected');
    } catch (error) {
      setWalletStatus('wallet connection failed');
      setWalletError(error?.message || 'wallet connection error');
    }
  };

  const refreshAdminData = async (overrides = {}) => {
    setAdminStatus('loading admin data...');
    setAdminError('');

    const nextRequestQuery = Object.prototype.hasOwnProperty.call(overrides, 'adminRequestQuery')
      ? overrides.adminRequestQuery
      : adminRequestQuery;
    const nextInviteQuery = Object.prototype.hasOwnProperty.call(overrides, 'adminInviteQuery')
      ? overrides.adminInviteQuery
      : adminInviteQuery;
    const nextRegistrationQuery = Object.prototype.hasOwnProperty.call(overrides, 'adminRegistrationQuery')
      ? overrides.adminRegistrationQuery
      : adminRegistrationQuery;
    const nextRequestStatus = Object.prototype.hasOwnProperty.call(overrides, 'adminRequestStatusFilter')
      ? overrides.adminRequestStatusFilter
      : adminRequestStatusFilter;
    const nextInviteStatus = Object.prototype.hasOwnProperty.call(overrides, 'adminInviteStatusFilter')
      ? overrides.adminInviteStatusFilter
      : adminInviteStatusFilter;
    const nextRegistrationStatus = Object.prototype.hasOwnProperty.call(overrides, 'adminRegistrationStatusFilter')
      ? overrides.adminRegistrationStatusFilter
      : adminRegistrationStatusFilter;
    const nextRequestSortBy = Object.prototype.hasOwnProperty.call(overrides, 'adminRequestSortBy')
      ? overrides.adminRequestSortBy
      : adminRequestSortBy;
    const nextRequestSortDir = Object.prototype.hasOwnProperty.call(overrides, 'adminRequestSortDir')
      ? overrides.adminRequestSortDir
      : adminRequestSortDir;
    const nextInviteSortBy = Object.prototype.hasOwnProperty.call(overrides, 'adminInviteSortBy')
      ? overrides.adminInviteSortBy
      : adminInviteSortBy;
    const nextInviteSortDir = Object.prototype.hasOwnProperty.call(overrides, 'adminInviteSortDir')
      ? overrides.adminInviteSortDir
      : adminInviteSortDir;
    const nextRegistrationSortBy = Object.prototype.hasOwnProperty.call(overrides, 'adminRegistrationSortBy')
      ? overrides.adminRegistrationSortBy
      : adminRegistrationSortBy;
    const nextRegistrationSortDir = Object.prototype.hasOwnProperty.call(overrides, 'adminRegistrationSortDir')
      ? overrides.adminRegistrationSortDir
      : adminRegistrationSortDir;
    const nextRequestPage = Object.prototype.hasOwnProperty.call(overrides, 'adminRequestPage')
      ? overrides.adminRequestPage
      : adminRequestPage;
    const nextInvitePage = Object.prototype.hasOwnProperty.call(overrides, 'adminInvitePage')
      ? overrides.adminInvitePage
      : adminInvitePage;
    const nextRegistrationPage = Object.prototype.hasOwnProperty.call(overrides, 'adminRegistrationPage')
      ? overrides.adminRegistrationPage
      : adminRegistrationPage;
    const nextPageSize = Object.prototype.hasOwnProperty.call(overrides, 'adminPageSize')
      ? overrides.adminPageSize
      : adminPageSize;

    const limit = Math.max(1, Number(nextPageSize) || 5);
    const requestOffset = Math.max(0, (Math.max(1, Number(nextRequestPage) || 1) - 1) * limit);
    const inviteOffset = Math.max(0, (Math.max(1, Number(nextInvitePage) || 1) - 1) * limit);
    const registrationOffset = Math.max(0, (Math.max(1, Number(nextRegistrationPage) || 1) - 1) * limit);

    const requestParams = new URLSearchParams({
      limit: String(limit),
      offset: String(requestOffset)
    });
    if (nextRequestStatus && nextRequestStatus !== 'all') {
      requestParams.set('status', nextRequestStatus);
    }
    requestParams.set('sort_by', nextRequestSortBy || 'created_at');
    requestParams.set('sort_dir', nextRequestSortDir || 'desc');
    if ((nextRequestQuery || '').trim()) {
      requestParams.set('q', String(nextRequestQuery).trim());
    }

    const inviteParams = new URLSearchParams({
      limit: String(limit),
      offset: String(inviteOffset),
      include_revoked: 'true',
      status: nextInviteStatus || 'all'
    });
    if ((nextInviteQuery || '').trim()) {
      inviteParams.set('q', String(nextInviteQuery).trim());
    }
    inviteParams.set('sort_by', nextInviteSortBy || 'created_at');
    inviteParams.set('sort_dir', nextInviteSortDir || 'desc');

    const registrationParams = new URLSearchParams({
      limit: String(limit),
      offset: String(registrationOffset),
      status: nextRegistrationStatus || 'all'
    });
    if ((nextRegistrationQuery || '').trim()) {
      registrationParams.set('q', String(nextRegistrationQuery).trim());
    }
    registrationParams.set('sort_by', nextRegistrationSortBy || 'registered_at');
    registrationParams.set('sort_dir', nextRegistrationSortDir || 'desc');

    try {
      const headers = adminHeaders();
      const [methodsResp, reqResp, invitesResp, registrationsResp] = await Promise.all([
        fetch(`${API_BASE}/admin/auth/methods`),
        fetch(`${API_BASE}/join/invite_requests?${requestParams.toString()}`, { headers }),
        fetch(`${API_BASE}/join/invites?${inviteParams.toString()}`, { headers }),
        fetch(`${API_BASE}/join/registrations?${registrationParams.toString()}`, { headers })
      ]);

      if (methodsResp.ok) {
        setAdminAuthMethods(await methodsResp.json());
      }

      if (!reqResp.ok || !invitesResp.ok || !registrationsResp.ok) {
        setAdminStatus('admin access unavailable');
        const firstError = await reqResp.json().catch(() => ({}));
        setAdminError(firstError?.message || firstError?.error || 'unauthorized or invalid admin credentials');
        return;
      }

      const reqPayload = await reqResp.json();
      const invitesPayload = await invitesResp.json();
      const registrationsPayload = await registrationsResp.json();
      setAdminInviteRequests(Array.isArray(reqPayload.requests) ? reqPayload.requests : []);
      setAdminInvites(Array.isArray(invitesPayload.invites) ? invitesPayload.invites : []);
      setAdminRegistrations(Array.isArray(registrationsPayload.registrations) ? registrationsPayload.registrations : []);
      setAdminInviteRequestsTotal(Number(reqPayload.total || 0));
      setAdminInvitesTotal(Number(invitesPayload.total || 0));
      setAdminRegistrationsTotal(Number(registrationsPayload.total || 0));
      setAdminStatus('admin data loaded');
    } catch {
      setAdminStatus('admin refresh failed');
      setAdminError('network error');
    }
  };

  const requestTotalPages = Math.max(1, Math.ceil(adminInviteRequestsTotal / Math.max(1, adminPageSize)));
  const inviteTotalPages = Math.max(1, Math.ceil(adminInvitesTotal / Math.max(1, adminPageSize)));
  const registrationTotalPages = Math.max(1, Math.ceil(adminRegistrationsTotal / Math.max(1, adminPageSize)));

  const logAssistantEvent = async (actionType, payload = {}) => {
    try {
      await fetch(`${API_BASE}/governance/actions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          action_type: actionType,
          actor: walletAddress || 'local-assistant',
          source: 'assistant',
          payload
        })
      });
    } catch {
      // Assistant telemetry is best-effort only.
    }
  };

  const showImpactPreview = (config) => {
    setImpactConfirmInput('');
    setImpactCooldownRemaining(Math.max(0, Number(config?.cooldownSeconds || 0)));
    setImpactPreview(config);
  };

  const withGuardPolicy = (actionKey, config, context = {}) => {
    const policy = IMPACT_GUARDS[actionKey] || {};
    const resolvedRequiredText = typeof policy.requiredText === 'function'
      ? policy.requiredText(context)
      : policy.requiredText;
    return {
      ...config,
      cooldownSeconds: Object.prototype.hasOwnProperty.call(config, 'cooldownSeconds')
        ? config.cooldownSeconds
        : Number(policy.cooldownSeconds || 0),
      requiredText: Object.prototype.hasOwnProperty.call(config, 'requiredText')
        ? config.requiredText
        : resolvedRequiredText,
      requiredLabel: Object.prototype.hasOwnProperty.call(config, 'requiredLabel')
        ? config.requiredLabel
        : policy.requiredLabel,
    };
  };

  const closeImpactPreview = () => {
    setImpactConfirmInput('');
    setImpactCooldownRemaining(0);
    setImpactPreview(null);
  };

  const confirmImpactPreview = async () => {
    if (impactPreview?.requiredText) {
      const expected = String(impactPreview.requiredText).trim();
      if (impactConfirmInput.trim() !== expected) {
        return;
      }
    }
    if (!impactPreview?.onConfirm) {
      setImpactPreview(null);
      return;
    }
    const runner = impactPreview.onConfirm;
    setImpactPreview(null);
    setImpactCooldownRemaining(0);
    await runner();
  };

  useEffect(() => {
    if (!impactPreview || impactCooldownRemaining <= 0) return undefined;
    const timer = setInterval(() => {
      setImpactCooldownRemaining((value) => {
        if (value <= 1) return 0;
        return value - 1;
      });
    }, 1000);
    return () => clearInterval(timer);
  }, [impactPreview, impactCooldownRemaining]);

  const approveInviteRequest = async (requestId, skipPreview = false) => {
    if (!skipPreview) {
      showImpactPreview(withGuardPolicy('approve_invite_request', {
        title: 'Approve Invite Request',
        endpoint: `/join/invite_requests/${requestId}/approve`,
        method: 'POST',
        payload: { max_uses: 1, expires_in_hours: 24 },
        impact: [
          'Creates a join invite for this request.',
          'Request status changes from pending to approved.'
        ],
        confirmLabel: 'Approve Request',
        onConfirm: () => approveInviteRequest(requestId, true)
      }));
      return;
    }
    setAdminStatus('approving invite request...');
    setAdminError('');
    try {
      const response = await fetch(`${API_BASE}/join/invite_requests/${requestId}/approve`, {
        method: 'POST',
        headers: adminHeaders(),
        body: JSON.stringify({ max_uses: 1, expires_in_hours: 24 })
      });
      if (!response.ok) {
        const err = await response.json().catch(() => ({}));
        setAdminStatus('approve failed');
        setAdminError(err?.message || err?.error || 'approve failed');
        return;
      }
      setAdminStatus('invite request approved');
      await logAssistantEvent('assistant_admin_approve_invite_request', { request_id: requestId });
      await refreshAdminData();
      await refreshMarketplace();
    } catch {
      setAdminStatus('approve failed');
      setAdminError('network error');
    }
  };

  const rejectInviteRequest = async (requestId, skipPreview = false) => {
    if (!skipPreview) {
      showImpactPreview(withGuardPolicy('reject_invite_request', {
        title: 'Reject Invite Request',
        endpoint: `/join/invite_requests/${requestId}/reject`,
        method: 'POST',
        payload: { reason: rejectReason },
        impact: [
          'Request status changes from pending to rejected.',
          'Rejection reason is recorded for audit history.'
        ],
        confirmLabel: 'Reject Request',
        onConfirm: () => rejectInviteRequest(requestId, true)
      }));
      return;
    }
    setAdminStatus('rejecting invite request...');
    setAdminError('');
    try {
      const response = await fetch(`${API_BASE}/join/invite_requests/${requestId}/reject`, {
        method: 'POST',
        headers: adminHeaders(),
        body: JSON.stringify({ reason: rejectReason })
      });
      if (!response.ok) {
        const err = await response.json().catch(() => ({}));
        setAdminStatus('reject failed');
        setAdminError(err?.message || err?.error || 'reject failed');
        return;
      }
      setAdminStatus('invite request rejected');
      await logAssistantEvent('assistant_admin_reject_invite_request', {
        request_id: requestId,
        reason: rejectReason
      });
      await refreshAdminData();
      await refreshMarketplace();
    } catch {
      setAdminStatus('reject failed');
      setAdminError('network error');
    }
  };

  const createAdminInvite = async () => {
    setAdminStatus('creating invite...');
    setAdminError('');
    try {
      const response = await fetch(`${API_BASE}/join/invite`, {
        method: 'POST',
        headers: adminHeaders(),
        body: JSON.stringify({
          participant_name: adminInviteParticipantName,
          max_uses: adminInviteMaxUses,
          expires_in_hours: adminInviteExpiresHours
        })
      });
      if (!response.ok) {
        const err = await response.json().catch(() => ({}));
        setAdminStatus('invite creation failed');
        setAdminError(err?.message || err?.error || 'invite creation failed');
        return;
      }
      setAdminStatus('invite created');
      await refreshAdminData();
      await refreshMarketplace();
    } catch {
      setAdminStatus('invite creation failed');
      setAdminError('network error');
    }
  };

  const revokeInvite = async (inviteId, skipPreview = false) => {
    if (!skipPreview) {
      showImpactPreview(withGuardPolicy('revoke_invite', {
        title: 'Revoke Invite',
        endpoint: `/join/invites/${inviteId}/revoke`,
        method: 'POST',
        payload: {},
        impact: [
          'Invite code can no longer be used by new participants.',
          'Existing registrations are not removed.'
        ],
        confirmLabel: 'Revoke Invite',
        onConfirm: () => revokeInvite(inviteId, true)
      }, { inviteId }));
      return;
    }
    setAdminStatus('revoking invite...');
    setAdminError('');
    try {
      const response = await fetch(`${API_BASE}/join/invites/${inviteId}/revoke`, {
        method: 'POST',
        headers: adminHeaders()
      });
      if (!response.ok) {
        const err = await response.json().catch(() => ({}));
        setAdminStatus('invite revoke failed');
        setAdminError(err?.message || err?.error || 'invite revoke failed');
        return;
      }
      setAdminStatus('invite revoked');
      await logAssistantEvent('assistant_admin_revoke_invite', { invite_id: inviteId });
      await refreshAdminData();
      await refreshMarketplace();
    } catch {
      setAdminStatus('invite revoke failed');
      setAdminError('network error');
    }
  };

  const revokeRegistration = async (nodeId, skipPreview = false) => {
    if (!skipPreview) {
      showImpactPreview(withGuardPolicy('revoke_registration', {
        title: 'Revoke Node Certificate',
        endpoint: `/join/revoke/${nodeId}`,
        method: 'POST',
        payload: {},
        impact: [
          `Node-${nodeId} may be blocked from future participation until re-onboarded.`,
          'Action is intended for compromise or policy violation scenarios.'
        ],
        confirmLabel: 'Revoke Node',
        onConfirm: () => revokeRegistration(nodeId, true)
      }, { nodeId }));
      return;
    }
    setAdminStatus('revoking node certificate...');
    setAdminError('');
    try {
      const response = await fetch(`${API_BASE}/join/revoke/${nodeId}`, {
        method: 'POST',
        headers: adminHeaders()
      });
      if (!response.ok) {
        const err = await response.json().catch(() => ({}));
        setAdminStatus('node revoke failed');
        setAdminError(err?.message || err?.error || 'node revoke failed');
        return;
      }
      setAdminStatus('node certificate revoked');
      await logAssistantEvent('assistant_admin_revoke_registration', { node_id: nodeId });
      await refreshAdminData();
      await refreshMarketplace();
    } catch {
      setAdminStatus('node revoke failed');
      setAdminError('network error');
    }
  };

  const shareComputeAttestation = async () => {
    setMarketplaceStatus('sharing compute attestation...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/attestations/share`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          participant_name: attestParticipantName,
          node_name: attestParticipantName,
          compute_type: attestComputeType,
          compute_capacity: attestComputeCapacity,
          attestation_status: attestStatus,
          region: attestRegion,
          notes: 'shared from browser studio'
        })
      });
      if (response.ok) {
        setMarketplaceStatus('attestation shared');
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('attestation share failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('attestation share failed');
      setMarketplaceLastError('network error');
    }
  };

  const previewMarketplacePolicy = async () => {
    const intentId = selectedIntentId || marketplaceIntents.find((item) => item.status === 'open')?.round_intent_id;
    if (!intentId) {
      setMarketplaceStatus('no open intent for preview');
      setMarketplaceLastError('');
      return;
    }

    setMarketplaceStatus('running policy preview...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/marketplace/policy/preview`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          round_intent_id: intentId,
          max_offers: 3,
          policy_overrides: {
            budget_total: previewBudget,
            min_quality_score: previewMinQuality
          }
        })
      });

      if (response.ok) {
        const payload = await response.json();
        setPolicyPreview(payload);
        setMarketplaceStatus('policy preview complete');
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('policy preview failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('policy preview failed');
      setMarketplaceLastError('network error');
    }
  };

  const createGovernanceProposal = async () => {
    setMarketplaceStatus('creating proposal...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/governance/proposals`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: proposalTitle,
          proposal_type: proposalType,
          description: proposalDescription,
          created_by: 'hud-operator'
        })
      });

      if (response.ok) {
        const payload = await response.json();
        setSelectedProposalId(payload.proposal_id || '');
        setMarketplaceStatus('proposal created');
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('proposal creation failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('proposal creation failed');
      setMarketplaceLastError('network error');
    }
  };

  const castGovernanceVote = async () => {
    if (!selectedProposalId) {
      setMarketplaceStatus('no proposal selected for vote');
      setMarketplaceLastError('');
      return;
    }

    setMarketplaceStatus('casting vote...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/governance/proposals/${selectedProposalId}/vote`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          voter: 'hud-operator',
          decision: voteDecision,
          weight: voteWeight,
          reason: voteReason
        })
      });

      if (response.ok) {
        setMarketplaceStatus('vote recorded');
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('vote failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('vote failed');
      setMarketplaceLastError('network error');
    }
  };

  const submitDispute = async (skipPreview = false) => {
    if (!latestContract?.contract_id) {
      setMarketplaceStatus('no contract available for dispute');
      setMarketplaceLastError('');
      return;
    }
    if (!skipPreview) {
      showImpactPreview(withGuardPolicy('submit_dispute', {
        title: 'Submit Contract Dispute',
        endpoint: '/marketplace/disputes',
        method: 'POST',
        payload: {
          contract_id: latestContract.contract_id,
          reporter: 'hud-operator',
          reason: disputeReason
        },
        impact: [
          'Creates an open dispute and governance review trail.',
          'Can delay contract settlement workflows.'
        ],
        confirmLabel: 'Submit Dispute',
        onConfirm: () => submitDispute(true)
      }));
      return;
    }
    setMarketplaceStatus('submitting dispute...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/marketplace/disputes`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          contract_id: latestContract.contract_id,
          reporter: 'hud-operator',
          reason: disputeReason
        })
      });
      if (response.ok) {
        setMarketplaceStatus('dispute submitted');
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('dispute submission failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      if (response.ok) {
        await logAssistantEvent('assistant_dispute_submitted', {
          contract_id: latestContract.contract_id,
          reason: disputeReason
        });
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('dispute submission failed');
      setMarketplaceLastError('network error');
    }
  };

  const recordGovernanceAction = async () => {
    setMarketplaceStatus('recording governance action...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/governance/actions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          action_type: 'marketplace_policy_review_requested',
          actor: 'hud-operator',
          source: 'hud',
          payload: {
            selected_intent_id: selectedIntentId || null,
            contract_id: latestContract?.contract_id || null
          }
        })
      });
      if (response.ok) {
        setMarketplaceStatus('governance action recorded');
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('governance action failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('governance action failed');
      setMarketplaceLastError('network error');
    }
  };

  const createOffer = async () => {
    setMarketplaceStatus('creating offer...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/marketplace/offers`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          seller_node_id: `hud-node-${participants}`,
          dataset_fingerprint: `sha256:hud-${Date.now()}`,
          title: offerTitle,
          modality: offerModality,
          quality_score: offerQuality,
          allowed_tasks: [intentTask],
          price_per_round: offerPrice,
          min_rounds: 1,
          attestation_status: 'verified'
        })
      });
      if (response.ok) {
        setMarketplaceStatus('offer created');
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('offer creation failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('offer creation failed');
      setMarketplaceLastError('network error');
    }
  };

  const createIntent = async () => {
    setMarketplaceStatus('creating intent...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/marketplace/round_intents`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model_owner_id: 'hud-operator',
          task_type: intentTask,
          required_modalities: [offerModality],
          min_quality_score: intentMinQuality,
          budget_total: intentBudget
        })
      });

      if (response.ok) {
        const payload = await response.json();
        setSelectedIntentId(payload.round_intent_id || '');
        setMarketplaceStatus('intent created');
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('intent creation failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('intent creation failed');
      setMarketplaceLastError('network error');
    }
  };

  const createMatch = async () => {
    const intentId = selectedIntentId || marketplaceIntents.find((item) => item.status === 'open')?.round_intent_id;
    if (!intentId) {
      setMarketplaceStatus('no open intent to match');
      setMarketplaceLastError('');
      return;
    }

    setMarketplaceStatus('matching offers...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/marketplace/match`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ round_intent_id: intentId, max_offers: 3 })
      });
      if (response.ok) {
        setMarketplaceStatus('match contract created');
      } else {
        const err = await response.json().catch(() => ({}));
        const reasons = err?.details?.rejection_reasons
          ? Object.entries(err.details.rejection_reasons)
              .map(([key, value]) => `${key}:${value}`)
              .join(', ')
          : '';
        setMarketplaceStatus('match failed');
        setMarketplaceLastError(
          reasons
            ? `${err?.message || err?.error || 'match failed'} (${reasons})`
            : (err?.message || err?.error || 'match failed')
        );
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('match failed');
      setMarketplaceLastError('network error');
    }
  };

  const releaseLatestContract = async (skipPreview = false) => {
    const pending = marketplaceContracts.find((item) => item.payout_status === 'pending');
    if (!pending?.contract_id) {
      setMarketplaceStatus('no pending contract found');
      setMarketplaceLastError('');
      return;
    }

    if (!skipPreview) {
      showImpactPreview(withGuardPolicy('release_escrow', {
        title: 'Release Contract Escrow',
        endpoint: '/marketplace/escrow/release',
        method: 'POST',
        payload: { contract_id: pending.contract_id },
        impact: [
          `Releases payout for ${pending.contract_id}.`,
          `Amount: ${Number(pending.agreed_price_per_round_total || 0).toFixed(2)}.`
        ],
        confirmLabel: 'Release Escrow',
        onConfirm: () => releaseLatestContract(true)
      }, { contractId: pending.contract_id }));
      return;
    }

    setMarketplaceStatus('releasing escrow...');
    setMarketplaceLastError('');
    try {
      const response = await fetch(`${API_BASE}/marketplace/escrow/release`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ contract_id: pending.contract_id })
      });
      if (response.ok) {
        setMarketplaceStatus('escrow released');
        await logAssistantEvent('assistant_escrow_released', {
          contract_id: pending.contract_id,
          agreed_price_per_round_total: pending.agreed_price_per_round_total
        });
      } else {
        const err = await response.json().catch(() => ({}));
        setMarketplaceStatus('escrow release failed');
        setMarketplaceLastError(err?.message || err?.error || 'unknown error');
      }
      await refreshMarketplace();
    } catch {
      setMarketplaceStatus('escrow release failed');
      setMarketplaceLastError('network error');
    }
  };

  const latestContract = marketplaceContracts.length > 0 ? marketplaceContracts[0] : null;
  const latestTimeline = Array.isArray(latestContract?.timeline) ? latestContract.timeline : [];
  const pendingContract = marketplaceContracts.find((item) => item.payout_status === 'pending');
  const openIntent = marketplaceIntents.find((item) => item.status === 'open');
  const openDisputes = marketplaceDisputes.filter(
    (item) => item.status === 'open' || item.status === 'under_review'
  ).length;

  const assistantGuidance = useMemo(() => {
    const hasAdminAuth = Boolean(String(adminToken || '').trim() || walletAddress);
    const inviteReady = Boolean(
      String(inviteRequestName || '').trim() &&
      String(inviteRequestEmail || '').includes('@') &&
      String(inviteRequestComputeType || '').trim()
    );
    const attestationReady = Boolean(
      String(attestParticipantName || '').trim() &&
      String(attestComputeType || '').trim() &&
      String(attestComputeCapacity || '').trim()
    );

    const checks = [
      {
        key: 'intent_ready',
        label: 'Intent ready for matching',
        ok: Boolean(openIntent),
        detail: openIntent ? openIntent.round_intent_id : 'create an open intent first',
        fieldId: 'intent-match-select'
      },
      {
        key: 'offer_available',
        label: 'At least one active offer',
        ok: marketplaceOffers.length > 0,
        detail: marketplaceOffers.length > 0 ? `${marketplaceOffers.length} offers` : 'create an offer',
        fieldId: 'offer-title-input'
      },
      {
        key: 'release_safe',
        label: 'Contract can be released safely',
        ok: Boolean(pendingContract) && openDisputes === 0,
        detail: pendingContract
          ? openDisputes > 0
            ? `${openDisputes} open disputes`
            : pendingContract.contract_id
          : 'no pending contract',
        fieldId: 'dispute-reason-input'
      },
      {
        key: 'invite_form_valid',
        label: 'Invite request form is valid',
        ok: inviteReady,
        detail: inviteReady ? 'ready to submit' : 'name/email/compute required',
        fieldId: 'invite-email-input'
      },
      {
        key: 'attestation_form_valid',
        label: 'Attestation form is valid',
        ok: attestationReady,
        detail: attestationReady ? 'ready to share' : 'participant/type/capacity required',
        fieldId: 'attest-participant-input'
      },
      {
        key: 'admin_auth_ready',
        label: 'Admin credentials available',
        ok: hasAdminAuth,
        detail: hasAdminAuth ? 'token or wallet provided' : 'connect wallet or set admin token',
        fieldId: 'admin-token-input'
      }
    ];

    const risks = [];
    if (offerQuality < intentMinQuality) {
      risks.push('Current offer quality is below intent minimum quality, so matching can fail.');
    }
    if (intentBudget > 0 && offerPrice > intentBudget) {
      risks.push('Offer price is above intent budget and may be rejected by policy.');
    }
    if (pendingContract && openDisputes > 0) {
      risks.push('Escrow release is risky while disputes are still open or under review.');
    }
    if (!String(inviteRequestEmail || '').includes('@')) {
      risks.push('Invite request email appears invalid and will be rejected by API validation.');
    }

    const suggestions = [];
    if (!openIntent) suggestions.push('Create an intent, then run Match with max 3 offers.');
    if (marketplaceOffers.length === 0) suggestions.push('Create at least one verified offer before matching.');
    if (pendingContract && openDisputes === 0) suggestions.push('A pending contract can be safely released now.');
    if (!hasAdminAuth) suggestions.push('Connect wallet or provide admin token before moderation actions.');
    if (suggestions.length === 0) suggestions.push('System is in a healthy state for marketplace and admin operations.');

    let modeSummary = 'Balanced guidance for contributor, operator, and admin actions.';
    if (assistantMode === 'contributor') {
      modeSummary = 'Contributor mode emphasizes attestation sharing and invite readiness.';
    } else if (assistantMode === 'operator') {
      modeSummary = 'Operator mode emphasizes matching quality, budgets, and release safety.';
    } else if (assistantMode === 'admin') {
      modeSummary = 'Admin mode emphasizes auth readiness and safe moderation workflows.';
    }

    const score = Math.round((checks.filter((item) => item.ok).length / checks.length) * 100);
    return { checks, risks, suggestions, score, modeSummary };
  }, [
    adminToken,
    assistantMode,
    attestComputeCapacity,
    attestComputeType,
    attestParticipantName,
    inviteRequestComputeType,
    inviteRequestEmail,
    inviteRequestName,
    intentBudget,
    intentMinQuality,
    marketplaceDisputes,
    marketplaceIntents,
    marketplaceOffers,
    marketplaceContracts,
    offerPrice,
    offerQuality,
    openIntent,
    openDisputes,
    pendingContract,
    walletAddress
  ]);

  const applyAssistantDefaults = () => {
    setIntentMinQuality((value) => Math.max(value, 0.75));
    setPreviewMinQuality((value) => Math.max(value, 0.75));
    setPreviewBudget((value) => Math.max(value, intentBudget || 60));
    setCompressionBits(8);
    setEpsilon((value) => clamp(value, 0.8, 2.0));
    setMarketplaceStatus('assistant applied safe defaults');
    setMarketplaceLastError('');
    logAssistantEvent('assistant_apply_safe_defaults', {
      intent_min_quality: intentMinQuality,
      preview_min_quality: previewMinQuality,
      preview_budget: previewBudget,
      compression_bits: compressionBits,
      epsilon
    });
  };

  const focusAssistantField = (fieldId, checkKey) => {
    if (!fieldId) return;
    if (typeof document === 'undefined') return;
    const node = document.getElementById(fieldId);
    if (!node) return;
    node.scrollIntoView({ behavior: 'smooth', block: 'center' });
    node.focus();
    logAssistantEvent('assistant_focus_blocked_field', { check_key: checkKey, field_id: fieldId });
  };

  return (
    <section className="demo-shell">
      <div className="demo-topbar">
        <div>
          <h2>Browser Federated Learning Studio</h2>
          <p>
            {trainingMode === 'simulation'
              ? 'Interactive simulation of local training, privacy noise, gradient compression, and network bandwidth tradeoffs.'
              : 'Real federated learning with CIFAR-10, gradient compression, and CUDA multi-GPU support.'}
            {usingBackendData ? ' (Real network metrics enabled)' : ''}
          </p>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <div className={`runtime-badge ${backendClass}`}>{backendLabel}</div>
          {enableBackendMetrics && (
            <div className={`runtime-badge ${usingBackendData ? 'badge-sync' : 'badge-offline'}`}>
              {usingBackendData ? 'Network Live' : 'Simulation Only'}
            </div>
          )}
          {trainingMode === 'real' && (
            <div className={`runtime-badge ${phase3dStatus === 'training' ? 'badge-sync' : phase3dStatus === 'completed' ? 'badge-webgpu' : 'badge-offline'}`}>
              {phase3dStatus === 'training' ? 'Training Live' : phase3dStatus === 'completed' ? 'Training Complete' : 'Phase 3D Ready'}
            </div>
          )}
        </div>
      </div>

      <div className="demo-layout">
        <aside className="control-card">
          <h3>Round Controls</h3>

          <label>
            Training Mode
            <select value={trainingMode} onChange={(event) => setTrainingMode(event.target.value)}>
              <option value="simulation">Simulation</option>
              <option value="real">Phase 3D (Real)</option>
            </select>
          </label>

          <label>
            Participants
            <input
              type="range"
              min="10"
              max="500"
              value={participants}
              onChange={(event) => setParticipants(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
            />
            <span>{participants}</span>
          </label>

          <label>
            Local Epochs
            <input
              type="range"
              min="1"
              max="10"
              value={localEpochs}
              onChange={(event) => setLocalEpochs(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
            />
            <span>{localEpochs}</span>
          </label>

          <label>
            Privacy Epsilon
            <input
              type="range"
              min="0.2"
              max="3"
              step="0.1"
              value={epsilon}
              onChange={(event) => setEpsilon(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
            />
            <span>{epsilon.toFixed(1)}</span>
          </label>

          <label>
            Compression Bits
            <input
              type="range"
              min="4"
              max="16"
              step="1"
              value={compressionBits}
              onChange={(event) => setCompressionBits(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
            />
            <span>{compressionBits}-bit</span>
          </label>

          <label>
            Learning Rate
            <input
              type="range"
              min="0.005"
              max="0.08"
              step="0.001"
              value={learningRate}
              onChange={(event) => setLearningRate(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
            />
            <span>{learningRate.toFixed(3)}</span>
          </label>

          <label>
            Target Rounds
            <input
              type="number"
              min="5"
              max="400"
              value={trainingMode === 'real' ? phase3dTotal : targetRounds}
              onChange={(event) => {
                const value = clamp(Number(event.target.value) || 1, 5, 400);
                if (trainingMode === 'real') {
                  setPhase3dTotal(value);
                } else {
                  setTargetRounds(value);
                }
              }}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
            />
          </label>

          <div className="control-actions">
            {trainingMode === 'simulation' ? (
              <button onClick={() => setRunning((value) => !value)}>{running ? 'Pause Training' : 'Start Training'}</button>
            ) : (
              <button onClick={phase3dStatus === 'training' ? cancelPhase3dTraining : startPhase3dTraining}>
                {phase3dStatus === 'training' ? 'Cancel Training' : 'Start Real Training'}
              </button>
            )}
            <button className="ghost" onClick={resetScenario}>
              Reset
            </button>
          </div>

          <div className="marketplace-panel">
            <h4>Local Data Marketplace</h4>

            <label>
              Offer Title
              <input
                id="offer-title-input"
                type="text"
                value={offerTitle}
                onChange={(event) => setOfferTitle(event.target.value)}
              />
            </label>

            <label>
              Offer Modality
              <select value={offerModality} onChange={(event) => setOfferModality(event.target.value)}>
                <option value="image">image</option>
                <option value="tabular">tabular</option>
                <option value="text">text</option>
              </select>
            </label>

            <label>
              Offer Price per Round
              <input
                type="number"
                min="0"
                step="0.1"
                value={offerPrice}
                onChange={(event) => setOfferPrice(Number(event.target.value) || 0)}
              />
            </label>

            <label>
              Offer Quality
              <input
                type="number"
                min="0"
                max="1"
                step="0.01"
                value={offerQuality}
                onChange={(event) => setOfferQuality(clamp(Number(event.target.value) || 0, 0, 1))}
              />
            </label>

            <label>
              Intent Task
              <input
                type="text"
                value={intentTask}
                onChange={(event) => setIntentTask(event.target.value || 'classification')}
              />
            </label>

            <label>
              Intent Budget
              <input
                type="number"
                min="0"
                step="1"
                value={intentBudget}
                onChange={(event) => setIntentBudget(Number(event.target.value) || 0)}
              />
            </label>

            <label>
              Intent Min Quality
              <input
                type="number"
                min="0"
                max="1"
                step="0.01"
                value={intentMinQuality}
                onChange={(event) => setIntentMinQuality(clamp(Number(event.target.value) || 0, 0, 1))}
              />
            </label>

            <label>
              Match Intent
              <select id="intent-match-select" value={selectedIntentId} onChange={(event) => setSelectedIntentId(event.target.value)}>
                <option value="">Auto-select open intent</option>
                {marketplaceIntents.map((intent) => (
                  <option key={intent.round_intent_id} value={intent.round_intent_id}>
                    {intent.round_intent_id} ({intent.status})
                  </option>
                ))}
              </select>
            </label>

            <div className="marketplace-actions">
              <button type="button" onClick={createOffer}>Create Offer</button>
              <button type="button" onClick={createIntent}>Create Intent</button>
              <button type="button" onClick={createMatch}>Match</button>
              <button type="button" onClick={previewMarketplacePolicy}>Policy Preview</button>
              <button type="button" onClick={releaseLatestContract}>Release Escrow</button>
            </div>

            <label>
              Preview Budget Override
              <input
                type="number"
                min="0"
                step="1"
                value={previewBudget}
                onChange={(event) => setPreviewBudget(Number(event.target.value) || 0)}
              />
            </label>

            <label>
              Preview Min Quality Override
              <input
                type="number"
                min="0"
                max="1"
                step="0.01"
                value={previewMinQuality}
                onChange={(event) => setPreviewMinQuality(clamp(Number(event.target.value) || 0, 0, 1))}
              />
            </label>

            <p className="marketplace-status">{marketplaceStatus}</p>
            {marketplaceLastError ? <p className="marketplace-error">{marketplaceLastError}</p> : null}
            <p className="marketplace-summary">
              Offers: {marketplaceOffers.length} | Intents: {marketplaceIntents.length} | Contracts: {marketplaceContracts.length} | Disputes: {marketplaceDisputes.length}
            </p>

            <div className="assistant-panel">
              <h5>AI Assistant (Local Guidance)</h5>
              <label>
                Guidance Mode
                <select value={assistantMode} onChange={(event) => setAssistantMode(event.target.value)}>
                  <option value="auto">auto</option>
                  <option value="contributor">contributor</option>
                  <option value="operator">operator</option>
                  <option value="admin">admin</option>
                </select>
              </label>
              <p className="assistant-summary">Readiness: {assistantGuidance.score}% | {assistantGuidance.modeSummary}</p>
              <ul className="assistant-checklist">
                {assistantGuidance.checks.map((item) => (
                  <li key={item.label} className={item.ok ? 'check-ok' : 'check-warn'}>
                    <span>{item.ok ? 'Ready' : 'Check'} {item.label}</span>
                    <small>{item.detail}</small>
                    {!item.ok && item.fieldId ? (
                      <button
                        type="button"
                        className="assistant-jump"
                        onClick={() => focusAssistantField(item.fieldId, item.key)}
                      >
                        Why blocked? Fix field
                      </button>
                    ) : null}
                  </li>
                ))}
              </ul>
              {assistantGuidance.risks.length > 0 ? (
                <div className="assistant-risks">
                  <p>Risk Flags</p>
                  <ul>
                    {assistantGuidance.risks.map((risk) => (
                      <li key={risk}>{risk}</li>
                    ))}
                  </ul>
                </div>
              ) : null}
              <div className="assistant-next">
                <p>Recommended Next Steps</p>
                <ul>
                  {assistantGuidance.suggestions.map((item) => (
                    <li key={item}>{item}</li>
                  ))}
                </ul>
              </div>
              <div className="marketplace-actions">
                <button type="button" onClick={applyAssistantDefaults}>Apply Safe Defaults</button>
              </div>
            </div>

            {latestContract ? (
              <div className="marketplace-timeline">
                <h5>Latest Contract Timeline</h5>
                <p className="timeline-contract-id">{latestContract.contract_id}</p>
                {latestContract.selected_offers?.[0]?.score_breakdown ? (
                  <div className="score-breakdown">
                    <p>Top Offer Score Breakdown</p>
                    <p>
                      Q: {latestContract.selected_offers[0].score_breakdown.quality_component.toFixed(3)} |
                      C: {latestContract.selected_offers[0].score_breakdown.cost_component.toFixed(3)} |
                      T: {latestContract.selected_offers[0].score_breakdown.trust_component.toFixed(3)} |
                      Total: {latestContract.selected_offers[0].score_breakdown.total.toFixed(3)}
                    </p>
                  </div>
                ) : null}
                {latestTimeline.length === 0 ? (
                  <p className="timeline-empty">No lifecycle events yet.</p>
                ) : (
                  <ul>
                    {latestTimeline
                      .slice()
                      .reverse()
                      .slice(0, 6)
                      .map((entry, index) => (
                        <li key={`${entry.ts}-${entry.event}-${index}`}>
                          <span className="timeline-event">{entry.event}</span>
                          <span className="timeline-ts">{new Date((entry.ts || 0) * 1000).toLocaleString()}</span>
                        </li>
                      ))}
                  </ul>
                )}

                <label>
                  Dispute Reason
                  <input
                    id="dispute-reason-input"
                    type="text"
                    value={disputeReason}
                    onChange={(event) => setDisputeReason(event.target.value)}
                  />
                </label>
                <div className="marketplace-actions">
                  <button type="button" onClick={submitDispute}>Submit Dispute</button>
                  <button type="button" onClick={recordGovernanceAction}>Log Governance Action</button>
                </div>

                <div className="governance-log">
                  <h5>Governance Activity</h5>
                  {governanceActions.length === 0 ? (
                    <p className="timeline-empty">No governance actions yet.</p>
                  ) : (
                    <ul>
                      {governanceActions.slice(0, 5).map((action) => (
                        <li key={action.action_id}>
                          <span className="timeline-event">{action.action_type}</span>
                          <span className="timeline-ts">{action.actor}</span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>

                <div className="proposal-panel">
                  <h5>Governance Proposals</h5>

                  <label>
                    Proposal Title
                    <input
                      type="text"
                      value={proposalTitle}
                      onChange={(event) => setProposalTitle(event.target.value)}
                    />
                  </label>

                  <label>
                    Proposal Type
                    <input
                      type="text"
                      value={proposalType}
                      onChange={(event) => setProposalType(event.target.value)}
                    />
                  </label>

                  <label>
                    Proposal Description
                    <input
                      type="text"
                      value={proposalDescription}
                      onChange={(event) => setProposalDescription(event.target.value)}
                    />
                  </label>

                  <div className="marketplace-actions">
                    <button type="button" onClick={createGovernanceProposal}>Create Proposal</button>
                  </div>

                  <label>
                    Vote Proposal
                    <select value={selectedProposalId} onChange={(event) => setSelectedProposalId(event.target.value)}>
                      <option value="">Select proposal</option>
                      {governanceProposals.map((proposal) => (
                        <option key={proposal.proposal_id} value={proposal.proposal_id}>
                          {proposal.proposal_id} ({proposal.status})
                        </option>
                      ))}
                    </select>
                  </label>

                  <label>
                    Vote Decision
                    <select value={voteDecision} onChange={(event) => setVoteDecision(event.target.value)}>
                      <option value="yes">yes</option>
                      <option value="no">no</option>
                      <option value="abstain">abstain</option>
                    </select>
                  </label>

                  <label>
                    Vote Weight
                    <input
                      type="number"
                      min="0"
                      step="0.1"
                      value={voteWeight}
                      onChange={(event) => setVoteWeight(Number(event.target.value) || 0)}
                    />
                  </label>

                  <label>
                    Vote Reason
                    <input
                      type="text"
                      value={voteReason}
                      onChange={(event) => setVoteReason(event.target.value)}
                    />
                  </label>

                  <div className="marketplace-actions">
                    <button type="button" onClick={castGovernanceVote}>Cast Vote</button>
                  </div>

                  {policyPreview?.preview ? (
                    <div className="policy-preview">
                      <p>Policy Preview Result</p>
                      <p>
                        Selected: {policyPreview.preview.selected_offers?.length || 0} |
                        Compatible: {policyPreview.preview.selection_diagnostics?.compatible_offers || 0} |
                        Budget Rejected: {policyPreview.preview.selection_diagnostics?.budget_rejected || 0}
                      </p>
                    </div>
                  ) : null}

                  {governanceProposals.length > 0 ? (
                    <ul className="proposal-list">
                      {governanceProposals.slice(0, 3).map((proposal) => (
                        <li key={proposal.proposal_id}>
                          <span className="timeline-event">{proposal.title}</span>
                          <span className="timeline-ts">
                            {proposal.status} | yes {proposal.tally?.yes ?? 0} / {proposal.tally?.total_weight ?? 0}
                          </span>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="timeline-empty">No proposals yet.</p>
                  )}
                </div>

                <div className="network-panel">
                  <h5>Network Expansion</h5>
                  <label>
                    Help Language
                    <select value={helperLanguage} onChange={(event) => setHelperLanguage(event.target.value)}>
                      <option value="en">English</option>
                      <option value="es">Espanol</option>
                      <option value="fr">Francais</option>
                    </select>
                  </label>
                  <p className="network-helper-text">{NETWORK_HELPER_TEXT[helperLanguage] || NETWORK_HELPER_TEXT.en}</p>

                  {networkSummary ? (
                    <p className="marketplace-summary">
                      Active Nodes: {networkSummary.active_nodes || 0} | Open Invites: {networkSummary.open_invites || 0} |
                      Pending Requests: {networkSummary.pending_invite_requests || 0} | Verified Ratio: {((networkSummary.verified_ratio || 0) * 100).toFixed(1)}%
                    </p>
                  ) : (
                    <p className="timeline-empty">Network expansion summary unavailable.</p>
                  )}

                  <label>
                    Participant Name
                    <input
                        id="attest-participant-input"
                      type="text"
                      value={attestParticipantName}
                      onChange={(event) => setAttestParticipantName(event.target.value)}
                    />
                  </label>

                  <label>
                    Compute Type
                    <select value={attestComputeType} onChange={(event) => setAttestComputeType(event.target.value)}>
                      <option value="gpu">gpu</option>
                      <option value="cpu">cpu</option>
                      <option value="npu">npu</option>
                      <option value="tpu">tpu</option>
                    </select>
                  </label>

                  <label>
                    Compute Capacity
                    <input
                      type="text"
                      value={attestComputeCapacity}
                      onChange={(event) => setAttestComputeCapacity(event.target.value)}
                    />
                  </label>

                  <label>
                    Attestation Status
                    <select value={attestStatus} onChange={(event) => setAttestStatus(event.target.value)}>
                      <option value="verified">verified</option>
                      <option value="pending">pending</option>
                      <option value="unverified">unverified</option>
                    </select>
                  </label>

                  <label>
                    Region
                    <input
                      type="text"
                      value={attestRegion}
                      onChange={(event) => setAttestRegion(event.target.value)}
                    />
                  </label>

                  <label>
                    Feed Sort
                    <select value={attestationSortBy} onChange={(event) => setAttestationSortBy(event.target.value)}>
                      <option value="reputation">reputation</option>
                      <option value="recent">recent</option>
                      <option value="capacity">capacity</option>
                    </select>
                  </label>

                  <div className="marketplace-actions">
                    <button type="button" onClick={shareComputeAttestation}>Share Compute Attestation</button>
                  </div>

                  {attestationFeed.length > 0 ? (
                    <ul className="proposal-list">
                      {attestationFeed.slice(0, 4).map((item) => (
                        <li key={item.attestation_id}>
                          <span className="timeline-event">{item.participant_name} ({item.compute_type})</span>
                          <span className="timeline-ts">{item.attestation_status} | {item.region} | rep {Number(item.reputation_score || 0).toFixed(3)}</span>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="timeline-empty">No shared compute attestations yet.</p>
                  )}

                  <div className="invite-request-panel">
                    <h5>Request Join Invite</h5>
                    <label>
                      Name
                      <input
                        type="text"
                        value={inviteRequestName}
                        onChange={(event) => setInviteRequestName(event.target.value)}
                      />
                    </label>
                    <label>
                      Contact Email
                      <input
                        id="invite-email-input"
                        type="text"
                        value={inviteRequestEmail}
                        onChange={(event) => setInviteRequestEmail(event.target.value)}
                      />
                    </label>
                    <label>
                      Compute Type
                      <select value={inviteRequestComputeType} onChange={(event) => setInviteRequestComputeType(event.target.value)}>
                        <option value="gpu">gpu</option>
                        <option value="cpu">cpu</option>
                        <option value="npu">npu</option>
                        <option value="tpu">tpu</option>
                      </select>
                    </label>
                    <label>
                      Region
                      <input
                        type="text"
                        value={inviteRequestRegion}
                        onChange={(event) => setInviteRequestRegion(event.target.value)}
                      />
                    </label>
                    <label>
                      Motivation
                      <input
                        type="text"
                        value={inviteRequestMotivation}
                        onChange={(event) => setInviteRequestMotivation(event.target.value)}
                      />
                    </label>
                    <div className="marketplace-actions">
                      <button type="button" onClick={requestJoinInvite}>Request Invite</button>
                    </div>
                  </div>

                  <div className="admin-panel">
                    <h5>Admin Operations</h5>

                    <label>
                      Select Web3 Wallet
                      <select value={selectedWalletProviderId} onChange={(event) => setSelectedWalletProviderId(event.target.value)}>
                        {walletProviders.length === 0 ? (
                          <option value="">No injected wallet found</option>
                        ) : (
                          walletProviders.map((wallet) => (
                            <option key={wallet.id} value={wallet.id}>{wallet.label}</option>
                          ))
                        )}
                      </select>
                    </label>
                    <div className="marketplace-actions">
                      <button type="button" onClick={connectSelectedWallet}>Connect Wallet</button>
                      <button type="button" onClick={refreshAdminData}>Refresh Admin Data</button>
                    </div>
                    <p className="marketplace-summary">{walletStatus}{walletAddress ? ` | ${walletAddress}` : ''}{walletChainId ? ` | chain ${walletChainId}` : ''}</p>
                    {walletError ? <p className="marketplace-error">{walletError}</p> : null}

                    <label>
                      Admin Token (optional if wallet allowlisted)
                      <input
                        id="admin-token-input"
                        type="text"
                        value={adminToken}
                        onChange={(event) => setAdminToken(event.target.value)}
                        placeholder="local-dev-admin-token"
                      />
                    </label>
                    {adminAuthMethods ? (
                      <p className="marketplace-summary">
                        Token auth: {adminAuthMethods.token_admin_enabled ? 'enabled' : 'disabled'} |
                        Wallet allowlist: {adminAuthMethods.wallet_allowlist_enabled ? `enabled (${adminAuthMethods.wallet_allowlist_count})` : 'disabled'}
                      </p>
                    ) : null}
                    <p className="marketplace-summary">{adminStatus}</p>
                    {adminError ? <p className="marketplace-error">{adminError}</p> : null}
                    <label>
                      Page Size
                      <select
                        value={adminPageSize}
                        onChange={(event) => {
                          const nextSize = Math.max(1, Number(event.target.value) || 5);
                          setAdminPageSize(nextSize);
                          setAdminRequestPage(1);
                          setAdminInvitePage(1);
                          setAdminRegistrationPage(1);
                          refreshAdminData({
                            adminPageSize: nextSize,
                            adminRequestPage: 1,
                            adminInvitePage: 1,
                            adminRegistrationPage: 1
                          });
                        }}
                      >
                        <option value="5">5</option>
                        <option value="10">10</option>
                        <option value="20">20</option>
                      </select>
                    </label>

                    <div className="admin-section">
                      <h6>Create Invite</h6>
                      <label>
                        Participant Name
                        <input
                          type="text"
                          value={adminInviteParticipantName}
                          onChange={(event) => setAdminInviteParticipantName(event.target.value)}
                        />
                      </label>
                      <label>
                        Max Uses
                        <input
                          type="number"
                          min="1"
                          step="1"
                          value={adminInviteMaxUses}
                          onChange={(event) => setAdminInviteMaxUses(Math.max(1, Number(event.target.value) || 1))}
                        />
                      </label>
                      <label>
                        Expires in Hours
                        <input
                          type="number"
                          min="1"
                          step="1"
                          value={adminInviteExpiresHours}
                          onChange={(event) => setAdminInviteExpiresHours(Math.max(1, Number(event.target.value) || 24))}
                        />
                      </label>
                      <div className="marketplace-actions">
                        <button type="button" onClick={createAdminInvite}>Create Invite</button>
                      </div>
                    </div>

                    <div className="admin-section">
                      <h6>Pending Invite Requests</h6>
                      <label>
                        Search Requests
                        <input
                          type="text"
                          value={adminRequestQuery}
                          onChange={(event) => {
                            const nextQuery = event.target.value;
                            setAdminRequestQuery(nextQuery);
                            setAdminRequestPage(1);
                            refreshAdminData({ adminRequestQuery: nextQuery, adminRequestPage: 1 });
                          }}
                          placeholder="name, email, compute, region"
                        />
                      </label>
                      <label>
                        Request Status
                        <select value={adminRequestStatusFilter} onChange={(event) => {
                          const nextStatus = event.target.value;
                          setAdminRequestStatusFilter(nextStatus);
                          setAdminRequestPage(1);
                          refreshAdminData({ adminRequestStatusFilter: nextStatus, adminRequestPage: 1 });
                        }}>
                          <option value="pending">pending</option>
                          <option value="approved">approved</option>
                          <option value="rejected">rejected</option>
                          <option value="all">all</option>
                        </select>
                      </label>
                      <label>
                        Sort Requests
                        <select value={`${adminRequestSortBy}:${adminRequestSortDir}`} onChange={(event) => {
                          const [nextSortBy, nextSortDir] = event.target.value.split(':');
                          setAdminRequestSortBy(nextSortBy);
                          setAdminRequestSortDir(nextSortDir);
                          setAdminRequestPage(1);
                          refreshAdminData({
                            adminRequestSortBy: nextSortBy,
                            adminRequestSortDir: nextSortDir,
                            adminRequestPage: 1
                          });
                        }}>
                          <option value="created_at:desc">Newest first</option>
                          <option value="created_at:asc">Oldest first</option>
                          <option value="participant_name:asc">Name A-Z</option>
                          <option value="participant_name:desc">Name Z-A</option>
                          <option value="compute_type:asc">Compute type A-Z</option>
                          <option value="status:asc">Status A-Z</option>
                        </select>
                      </label>
                      <label>
                        Reject Reason
                        <input
                          type="text"
                          value={rejectReason}
                          onChange={(event) => setRejectReason(event.target.value)}
                        />
                      </label>
                      {adminInviteRequests.length === 0 ? (
                        <p className="timeline-empty">No pending requests.</p>
                      ) : (
                        <ul className="proposal-list">
                          {adminInviteRequests.map((req) => (
                            <li key={req.request_id}>
                              <span className="timeline-event">{req.participant_name} ({req.compute_type})</span>
                              <span className="timeline-ts">{req.contact_email}</span>
                              <div className="admin-inline-actions">
                                <button type="button" onClick={() => approveInviteRequest(req.request_id)}>Approve</button>
                                <button type="button" className="danger" onClick={() => rejectInviteRequest(req.request_id)}>Reject</button>
                              </div>
                            </li>
                          ))}
                        </ul>
                      )}
                      <div className="admin-pagination">
                        <button
                          type="button"
                          onClick={() => {
                            const nextPage = Math.max(1, adminRequestPage - 1);
                            setAdminRequestPage(nextPage);
                            refreshAdminData({ adminRequestPage: nextPage });
                          }}
                        >
                          Prev
                        </button>
                        <span>{adminRequestPage}/{requestTotalPages}</span>
                        <button
                          type="button"
                          onClick={() => {
                            const nextPage = Math.min(requestTotalPages, adminRequestPage + 1);
                            setAdminRequestPage(nextPage);
                            refreshAdminData({ adminRequestPage: nextPage });
                          }}
                        >
                          Next
                        </button>
                      </div>
                    </div>

                    <div className="admin-section">
                      <h6>Invites</h6>
                      <label>
                        Search Invites
                        <input
                          type="text"
                          value={adminInviteQuery}
                          onChange={(event) => {
                            const nextQuery = event.target.value;
                            setAdminInviteQuery(nextQuery);
                            setAdminInvitePage(1);
                            refreshAdminData({ adminInviteQuery: nextQuery, adminInvitePage: 1 });
                          }}
                          placeholder="participant, invite id, source"
                        />
                      </label>
                      <label>
                        Invite Status
                        <select value={adminInviteStatusFilter} onChange={(event) => {
                          const nextStatus = event.target.value;
                          setAdminInviteStatusFilter(nextStatus);
                          setAdminInvitePage(1);
                          refreshAdminData({ adminInviteStatusFilter: nextStatus, adminInvitePage: 1 });
                        }}>
                          <option value="active">active</option>
                          <option value="revoked">revoked</option>
                          <option value="all">all</option>
                        </select>
                      </label>
                      <label>
                        Sort Invites
                        <select value={`${adminInviteSortBy}:${adminInviteSortDir}`} onChange={(event) => {
                          const [nextSortBy, nextSortDir] = event.target.value.split(':');
                          setAdminInviteSortBy(nextSortBy);
                          setAdminInviteSortDir(nextSortDir);
                          setAdminInvitePage(1);
                          refreshAdminData({
                            adminInviteSortBy: nextSortBy,
                            adminInviteSortDir: nextSortDir,
                            adminInvitePage: 1
                          });
                        }}>
                          <option value="created_at:desc">Newest first</option>
                          <option value="created_at:asc">Oldest first</option>
                          <option value="participant_name:asc">Participant A-Z</option>
                          <option value="participant_name:desc">Participant Z-A</option>
                          <option value="used:desc">Most used</option>
                          <option value="used:asc">Least used</option>
                          <option value="status:asc">Status A-Z</option>
                        </select>
                      </label>
                      {adminInvites.length === 0 ? (
                        <p className="timeline-empty">No invites found.</p>
                      ) : (
                        <ul className="proposal-list">
                          {adminInvites.map((invite) => (
                            <li key={invite.invite_id}>
                              <span className="timeline-event">{invite.participant_name}</span>
                              <span className="timeline-ts">used {invite.used}/{invite.max_uses} {invite.revoked ? '| revoked' : ''}</span>
                              {!invite.revoked ? (
                                <div className="admin-inline-actions">
                                  <button type="button" className="danger" onClick={() => revokeInvite(invite.invite_id)}>Revoke</button>
                                </div>
                              ) : null}
                            </li>
                          ))}
                        </ul>
                      )}
                      <div className="admin-pagination">
                        <button
                          type="button"
                          onClick={() => {
                            const nextPage = Math.max(1, adminInvitePage - 1);
                            setAdminInvitePage(nextPage);
                            refreshAdminData({ adminInvitePage: nextPage });
                          }}
                        >
                          Prev
                        </button>
                        <span>{adminInvitePage}/{inviteTotalPages}</span>
                        <button
                          type="button"
                          onClick={() => {
                            const nextPage = Math.min(inviteTotalPages, adminInvitePage + 1);
                            setAdminInvitePage(nextPage);
                            refreshAdminData({ adminInvitePage: nextPage });
                          }}
                        >
                          Next
                        </button>
                      </div>
                    </div>

                    <div className="admin-section">
                      <h6>Registrations</h6>
                      <label>
                        Search Registrations
                        <input
                          type="text"
                          value={adminRegistrationQuery}
                          onChange={(event) => {
                            const nextQuery = event.target.value;
                            setAdminRegistrationQuery(nextQuery);
                            setAdminRegistrationPage(1);
                            refreshAdminData({ adminRegistrationQuery: nextQuery, adminRegistrationPage: 1 });
                          }}
                          placeholder="participant, node name, node id"
                        />
                      </label>
                      <label>
                        Registration Status
                        <select value={adminRegistrationStatusFilter} onChange={(event) => {
                          const nextStatus = event.target.value;
                          setAdminRegistrationStatusFilter(nextStatus);
                          setAdminRegistrationPage(1);
                          refreshAdminData({ adminRegistrationStatusFilter: nextStatus, adminRegistrationPage: 1 });
                        }}>
                          <option value="all">all</option>
                          <option value="active">active</option>
                          <option value="revoked">revoked</option>
                        </select>
                      </label>
                      <label>
                        Sort Registrations
                        <select value={`${adminRegistrationSortBy}:${adminRegistrationSortDir}`} onChange={(event) => {
                          const [nextSortBy, nextSortDir] = event.target.value.split(':');
                          setAdminRegistrationSortBy(nextSortBy);
                          setAdminRegistrationSortDir(nextSortDir);
                          setAdminRegistrationPage(1);
                          refreshAdminData({
                            adminRegistrationSortBy: nextSortBy,
                            adminRegistrationSortDir: nextSortDir,
                            adminRegistrationPage: 1
                          });
                        }}>
                          <option value="registered_at:desc">Newest first</option>
                          <option value="registered_at:asc">Oldest first</option>
                          <option value="node_id:asc">Node ID ascending</option>
                          <option value="node_id:desc">Node ID descending</option>
                          <option value="participant_name:asc">Participant A-Z</option>
                          <option value="node_name:asc">Node name A-Z</option>
                          <option value="status:asc">Status A-Z</option>
                        </select>
                      </label>
                      {adminRegistrations.length === 0 ? (
                        <p className="timeline-empty">No registrations found.</p>
                      ) : (
                        <ul className="proposal-list">
                          {adminRegistrations.map((reg) => (
                            <li key={`${reg.node_id}-${reg.registered_at}`}>
                              <span className="timeline-event">node-{reg.node_id} {reg.node_name}</span>
                              <span className="timeline-ts">{reg.participant_name}</span>
                              <div className="admin-inline-actions">
                                <button type="button" className="danger" onClick={() => revokeRegistration(reg.node_id)}>Revoke Node</button>
                              </div>
                            </li>
                          ))}
                        </ul>
                      )}
                      <div className="admin-pagination">
                        <button
                          type="button"
                          onClick={() => {
                            const nextPage = Math.max(1, adminRegistrationPage - 1);
                            setAdminRegistrationPage(nextPage);
                            refreshAdminData({ adminRegistrationPage: nextPage });
                          }}
                        >
                          Prev
                        </button>
                        <span>{adminRegistrationPage}/{registrationTotalPages}</span>
                        <button
                          type="button"
                          onClick={() => {
                            const nextPage = Math.min(registrationTotalPages, adminRegistrationPage + 1);
                            setAdminRegistrationPage(nextPage);
                            refreshAdminData({ adminRegistrationPage: nextPage });
                          }}
                        >
                          Next
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ) : null}
          </div>
        </aside>

        <div className="metrics-column">
          <div className="kpi-grid">
            <article>
              <h4>Round</h4>
              <p>{activeRound}</p>
            </article>
            <article>
              <h4>Accuracy</h4>
              <p>{(accuracy * 100).toFixed(2)}%</p>
            </article>
            <article>
              <h4>Loss</h4>
              <p>{loss.toFixed(3)}</p>
            </article>
            <article>
              <h4>Compression</h4>
              <p>{compressionRatio.toFixed(2)}x</p>
            </article>
            <article>
              <h4>Bandwidth</h4>
              <p>{bandwidthKB.toFixed(1)} KB</p>
            </article>
            <article>
              <h4>Round Latency</h4>
              <p>{latencyMs} ms</p>
            </article>
          </div>

          <div className="progress-wrap">
            <div className="progress-labels">
              <span>Progress</span>
              <span>{progressPercent.toFixed(0)}%</span>
            </div>
            <div className="progress-track">
              <div className="progress-fill" style={{ width: `${progressPercent}%` }} />
            </div>
            <div className="progress-meta">
              <span>
                {activeRound}/{activeTarget} rounds
              </span>
              <span>{compressionSavedPercent.toFixed(1)}% payload saved</span>
            </div>
          </div>

          <div className="chart-grid">
            <article className="chart-card">
              <h3>Accuracy and Loss</h3>
              <ResponsiveContainer width="100%" height={260}>
                <LineChart data={activeHistory}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="round" />
                  <YAxis yAxisId="left" domain={[0, 1]} />
                  <YAxis yAxisId="right" orientation="right" domain={[0, 2.5]} />
                  <Tooltip />
                  <Legend />
                  <Line yAxisId="left" type="monotone" dataKey="accuracy" stroke="#1d4ed8" dot={false} />
                  <Line yAxisId="right" type="monotone" dataKey="loss" stroke="#f97316" dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </article>

            <article className="chart-card">
              <h3>Bandwidth and Compression</h3>
              <ResponsiveContainer width="100%" height={260}>
                <AreaChart data={activeHistory}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="round" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Area type="monotone" dataKey="bandwidthKB" stroke="#0ea5e9" fill="#bae6fd" />
                  <Line type="monotone" dataKey="compressionRatio" stroke="#16a34a" dot={false} />
                </AreaChart>
              </ResponsiveContainer>
            </article>
          </div>
        </div>
      </div>

      {impactPreview ? (
        <div className="impact-modal-backdrop" role="dialog" aria-modal="true" aria-label="Action impact summary">
          <div className="impact-modal">
            <h3>{impactPreview.title || 'Confirm Action'}</h3>
            <p className="impact-line">Method: {impactPreview.method || 'POST'}</p>
            <p className="impact-line">Endpoint: {impactPreview.endpoint || 'n/a'}</p>
            <div className="impact-payload">
              <p>Payload</p>
              <pre>{JSON.stringify(impactPreview.payload || {}, null, 2)}</pre>
            </div>
            <div className="impact-effects">
              <p>Expected Impact</p>
              <ul>
                {(impactPreview.impact || []).map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ul>
            </div>
            {impactPreview.requiredText ? (
              <label className="impact-confirm-label">
                {impactPreview.requiredLabel || 'Type confirmation text'}
                <input
                  type="text"
                  value={impactConfirmInput}
                  onChange={(event) => setImpactConfirmInput(event.target.value)}
                  placeholder={String(impactPreview.requiredText)}
                />
              </label>
            ) : null}
            {impactCooldownRemaining > 0 ? (
              <p className="impact-cooldown">Please review details. Confirm enabled in {impactCooldownRemaining}s.</p>
            ) : null}
            <div className="impact-actions">
              <button type="button" className="ghost" onClick={closeImpactPreview}>Cancel</button>
              <button
                type="button"
                onClick={confirmImpactPreview}
                disabled={
                  impactCooldownRemaining > 0 ||
                  Boolean(impactPreview.requiredText) &&
                  impactConfirmInput.trim() !== String(impactPreview.requiredText).trim()
                }
              >
                {impactPreview.confirmLabel || 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
}
