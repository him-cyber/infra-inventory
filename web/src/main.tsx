import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import {
  Activity,
  BarChart3,
  Box,
  Building2,
  ChevronDown,
  ClipboardCheck,
  Cloud,
  Database,
  GitBranch,
  LockKeyhole,
  Monitor,
  Network,
  RefreshCcw,
  Router,
  Search,
  Server,
  ShieldAlert,
  ShieldCheck,
  TicketCheck,
  Wifi,
  PlayCircle,
  SlidersHorizontal,
  UserRound,
  Upload,
  PauseCircle,
  Wrench,
  CheckCircle2,
  PlusCircle,
} from 'lucide-react';
import './styles.css';

type Asset = {
  id: string;
  type: string;
  name: string;
  owner: string;
  environment: string;
  region: string;
  service: string;
  version: number;
  dependencies: string[];
  risk: string;
  updated_at: string;
};

type Event = {
  event_id: string;
  asset_id: string;
  event_type: string;
  timestamp: string;
  config_version: string;
  payload: Asset;
};

type TopologyNode = {
  id: string;
  label: string;
  kind: string;
  owner: string;
  risk: string;
  detail: string;
  tone: string;
};

type Topology = {
  story: {
    mission: string;
    vision: string;
    servicenow_use_case: string;
  };
  metrics: {
    assets: number;
    high_risk: number;
    recent_events: number;
    config_version: string;
    query_slo: string;
    indexing_lag_slo: string;
    dependency_guard: string;
    config_reload_cadence: string;
  };
  nodes: TopologyNode[];
  edges: { from: string; to: string; label: string }[];
  teams: { team: string; need: string; served_by: string }[];
};

type Analytics = {
  generated_at: string;
  summary: string;
  signals: { label: string; value: string; status: string }[];
  recommendations: string[];
  live_feed: { title: string; detail: string; timestamp: string }[];
};

type CVEAnalysis = {
  analysis_id: string;
  generated_at: string;
  model: string;
  summary: string;
  findings: { cve_id: string; asset_name: string; severity: string; score: number; confidence: number; remediation: string }[];
  incident_brief?: {
    provider: string;
    model: string;
    mode: string;
    executive_summary: string;
    probable_cause: string;
    blast_radius: string;
    recommended_actions: string[];
    servicenow_work_notes: string[];
  };
  kafka_topic: string;
  search_backend: string;
  servicenow_target: string;
};

type ServiceNowTicket = {
  number: string;
  created_at: string;
  table: string;
  short_description: string;
  assignment_group: string;
  priority: string;
  configuration_items: string[];
  work_notes: string[];
};

type ConfigIntelligence = {
  generated_at: string;
  model: string;
  summary: string;
  observed_assets: number;
  observed_owners: string[];
  observed_services: string[];
  proposals: { field: string; current: string; proposed: string; reason: string; confidence: number }[];
  servicenow_routing: Record<string, string>;
  preview_config: Record<string, unknown>;
};

type AuthSession = {
  mode: string;
  required: boolean;
  authenticated: boolean;
  login_url: string;
  logout_url: string;
  storage: string;
  transport: string;
  user?: { email?: string; name?: string; provider?: string; exp?: string };
};

type OrgImport = {
  import_id: string;
  provider: string;
  account_id: string;
  tenant_id: string;
  subscription_id: string;
  imported_assets: number;
  published_events: number;
  encrypted_at_rest: string;
  encrypted_in_transit: string;
  created_at: string;
};

type EvidenceItem = {
  id: string;
  name: string;
  url: string;
  size: string;
  status: string;
};

type ConfigAutomation = {
  automation_id: string;
  applied_at: string;
  status: string;
  policy: string;
  guardrails: string[];
  routes: Record<string, string>;
  replay_window: number;
  affected_types: string[];
  evidence: string[];
};

const apiBase = import.meta.env.VITE_API_URL ?? '';

const sampleInventoryPayload = JSON.stringify({
  assets: [
    {
      id: 'branch-vpn-gateway',
      type: 'network',
      name: 'Branch VPN Gateway',
      owner: 'it-ops',
      environment: 'prod',
      region: 'us-east-office',
      service: 'office-connectivity',
      version: 31,
      dependencies: ['az-fw-prod', 'queue-inventory-events'],
      risk: 'high'
    },
    {
      id: 'agent-assist-api',
      type: 'service',
      name: 'Agent Assist API',
      owner: 'product',
      environment: 'prod',
      region: 'us-east-office',
      service: 'servicenow-support',
      version: 44,
      dependencies: ['opensearch-assets', 'queue-inventory-events'],
      risk: 'medium'
    },
    {
      id: 'knowledge-worker',
      type: 'worker',
      name: 'Knowledge Generator Worker',
      owner: 'ml-platform',
      environment: 'prod',
      region: 'us-east-office',
      service: 'risk-knowledge',
      version: 18,
      dependencies: ['agent-assist-api', 'opensearch-assets'],
      risk: 'medium'
    }
  ]
}, null, 2);


const dataContracts = [
  {
    name: 'InventoryGraph',
    type: 'GraphQL-style read model',
    shape: 'asset(id) { owner service env risk dependencies }',
    note: 'CMDB-style relationship lookup',
  },
  {
    name: 'ControlEvidence',
    type: 'GraphQL-style read model',
    shape: 'controls(framework: SOC2) { asset status evidence }',
    note: 'SOC 2 / NIST / CIS evidence view',
  },
  {
    name: 'YamlInventoryImport',
    type: 'CLI ingestion contract',
    shape: 'assets[] -> POST /api/assets -> Kafka -> OpenSearch',
    note: 'bring your existing YAML or JSON inventory',
  },
  {
    name: 'StreamReplay',
    type: 'Kafka event schema',
    shape: 'asset.upserted -> indexer -> OpenSearch document',
    note: 'replayable data contract',
  },
];

const officeEndpoints = [
  { icon: <Wifi />, label: 'Branch Wi-Fi', detail: 'APs, VLANs, device health', tone: 'blue' },
  { icon: <Monitor />, label: 'Employee Laptops', detail: 'owned endpoints and risk', tone: 'green' },
  { icon: <Router />, label: 'Edge Router', detail: 'office gateway dependency', tone: 'violet' },
  { icon: <TicketCheck />, label: 'ServiceNow Desk', detail: 'CMDB, incident, change', tone: 'rose' },
];

function App() {
  const [page, setPage] = useState(() => window.location.hash.replace('#', '') || 'overview');
  const [assets, setAssets] = useState<Asset[]>([]);
  const [events, setEvents] = useState<Event[]>([]);
  const [topology, setTopology] = useState<Topology | null>(null);
  const [analytics, setAnalytics] = useState<Analytics | null>(null);
  const [analysis, setAnalysis] = useState<CVEAnalysis | null>(null);
  const [configIntel, setConfigIntel] = useState<ConfigIntelligence | null>(null);
  const [ticket, setTicket] = useState<ServiceNowTicket | null>(null);
  const [ticketDraft, setTicketDraft] = useState<ServiceNowTicket | null>(null);
  const [savedTickets, setSavedTickets] = useState<ServiceNowTicket[]>([]);
  const [ticketModalOpen, setTicketModalOpen] = useState(false);
  const [authSession, setAuthSession] = useState<AuthSession | null>(null);
  const [azureModalOpen, setAzureModalOpen] = useState(false);
  const [imports, setImports] = useState<OrgImport[]>([]);
  const [actionStatus, setActionStatus] = useState('ready');
  const [attachment, setAttachment] = useState(sampleInventoryPayload);
  const [attachResult, setAttachResult] = useState('ready for inventory');
  const [evidence, setEvidence] = useState<EvidenceItem[]>([]);
  const [cooldownAssets, setCooldownAssets] = useState<string[]>([]);
  const [automations, setAutomations] = useState<ConfigAutomation[]>([]);
  const [selectedAssetId, setSelectedAssetId] = useState('');
  const [shiftMode, setShiftMode] = useState<'infra' | 'api'>('infra');
  const [query, setQuery] = useState('');
  const [config, setConfig] = useState<{ version: string; updated_at: string; replay_window: number } | null>(null);
  const [status, setStatus] = useState('checking');

  async function refresh() {
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    const [searchResp, eventResp, configResp, topologyResp, analyticsResp, healthResp, authResp, importsResp, automationsResp, ticketsResp] = await Promise.all([
      fetch(`${apiBase}/api/assets/search?${params.toString()}`),
      fetch(`${apiBase}/api/stream/recent`),
      fetch(`${apiBase}/api/config/version`),
      fetch(`${apiBase}/api/topology`),
      fetch(`${apiBase}/api/analytics`),
      fetch(`${apiBase}/healthz`),
      fetch(`${apiBase}/api/auth/session`),
      fetch(`${apiBase}/api/org/imports`),
      fetch(`${apiBase}/api/config/automations`),
      fetch(`${apiBase}/api/servicenow/tickets`)
    ]);
    const searchData = await searchResp.json();
    const eventData = await eventResp.json();
    setAssets(searchData.assets ?? []);
    setEvents(eventData.events ?? []);
    setConfig(await configResp.json());
    setTopology(await topologyResp.json());
    setAnalytics(await analyticsResp.json());
    setAuthSession(await authResp.json());
    setImports((await importsResp.json()).imports ?? []);
    setAutomations((await automationsResp.json()).automations ?? []);
    setSavedTickets((await ticketsResp.json()).tickets ?? []);
    setStatus(healthResp.ok ? 'healthy' : 'degraded');
  }

  async function generateAnalytics() {
    setActionStatus('generating analytics');
    const response = await fetch(`${apiBase}/api/analytics`);
    setAnalytics(await response.json());
    setActionStatus('analytics refreshed');
  }

  async function startAnalysis() {
    setActionStatus('running CVE analysis');
    const response = await fetch(`${apiBase}/api/security/analyze`, { method: 'POST' });
    setAnalysis(await response.json());
    setActionStatus('CVE analysis ready');
  }

  async function reviewTicket() {
    setActionStatus('building ticket review');
    const response = await fetch(`${apiBase}/api/servicenow/tickets/draft`, { method: 'POST' });
    setTicketDraft(await response.json());
    setTicketModalOpen(true);
    setActionStatus('ticket ready for review');
  }

  async function submitTicket() {
    setActionStatus('submitting ticket');
    const response = await fetch(`${apiBase}/api/servicenow/tickets`, { method: 'POST' });
    const saved = await response.json();
    setTicket(saved);
    setSavedTickets((current) => [saved, ...current].slice(0, 20));
    setTicketModalOpen(false);
    setActionStatus('ticket saved');
  }

  async function buildConfig() {
    setActionStatus('learning better config from inventory');
    const response = await fetch(`${apiBase}/api/config/recommendations`, { method: 'POST' });
    setConfigIntel(await response.json());
    setActionStatus('config recommendations ready');
  }

  async function applyConfigAutomation() {
    setActionStatus('applying config guardrail');
    const response = await fetch(`${apiBase}/api/config/automation`, { method: 'POST' });
    const automation = await response.json();
    setAutomations((current) => [automation, ...current].slice(0, 10));
    setActionStatus('config guardrail applied');
  }

  async function importAzureOrg() {
    if (authSession?.mode === 'azure' && !authSession.authenticated) {
      setActionStatus('Azure sign-in required');
      changePage('cloud');
      return;
    }
    if (authSession?.mode !== 'azure') {
      changePage('cloud');
      setActionStatus('cloud gateway ready');
      return;
    }
    await importAzureSample('connected');
  }

  function startCloudGateway() {
    window.location.href = `${apiBase}${authSession?.login_url ?? '/auth/login'}`;
  }

  async function importAzureSample(source: 'local' | 'connected' = 'local') {
    setAzureModalOpen(false);
    setActionStatus(source === 'connected' ? 'importing connected Azure inventory' : 'importing local Azure inventory sample');
    const response = await fetch(`${apiBase}/api/org/import`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-Tenant': 'demo' },
      body: JSON.stringify({
        provider: 'azure',
        account_id: 'azure-prod-connectivity',
        tenant_id: 'demo-tenant',
        subscription_id: 'sub-prod-connectivity',
        region: 'eastus',
        assets: [
          { id: 'az-fw-prod', type: 'network', name: 'Azure Firewall Prod', owner: 'it-ops', environment: 'prod', region: 'eastus', service: 'network-security', version: 17, dependencies: ['network-edge-router'], risk: 'high' },
          { id: 'az-aks-support', type: 'kubernetes', name: 'AKS Support Cluster', owner: 'platform', environment: 'prod', region: 'eastus', service: 'support-platform', version: 29, dependencies: ['queue-inventory-events', 'opensearch-assets'], risk: 'medium' },
          { id: 'az-aks-cve-worker', type: 'container', name: 'CVE Worker Container', owner: 'ml-platform', environment: 'prod', region: 'eastus', service: 'risk-knowledge', version: 11, dependencies: ['az-aks-support', 'opensearch-assets'], risk: 'high' },
          { id: 'az-kv-config', type: 'keyvault', name: 'Key Vault Config Store', owner: 'security', environment: 'prod', region: 'eastus', service: 'secure-config', version: 13, dependencies: ['az-aks-support'], risk: 'medium' }
        ]
      })
    });
    const result = await response.json();
    setImports((current) => [result, ...current].slice(0, 10));
    setActionStatus('Azure inventory imported');
    await refresh();
  }

  async function attachInventory() {
    try {
      setActionStatus('attaching inventory');
      const parsed = JSON.parse(attachment) as { assets?: Asset[] } | Asset[];
      const items = Array.isArray(parsed) ? parsed : parsed.assets;
      if (!items?.length) {
        throw new Error('expected { "assets": [...] } or an asset array');
      }
      for (const asset of items) {
        const response = await fetch(`${apiBase}/api/assets`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'X-Tenant': 'demo' },
          body: JSON.stringify(asset),
        });
        if (!response.ok) {
          throw new Error(`asset ${asset.id || asset.name} failed with HTTP ${response.status}`);
        }
      }
      setAttachResult(`${items.length} assets attached and streamed`);
      setActionStatus('inventory attached');
      await refresh();
    } catch (error) {
      const message = error instanceof Error ? error.message : 'invalid inventory payload';
      setAttachResult(message);
      setActionStatus('attach failed');
    }
  }

  async function generateKnowledge() {
    setActionStatus('generating knowledge');
    const [analyticsResp, analysisResp, configResp] = await Promise.all([
      fetch(`${apiBase}/api/analytics`),
      fetch(`${apiBase}/api/security/analyze`, { method: 'POST' }),
      fetch(`${apiBase}/api/config/recommendations`, { method: 'POST' }),
    ]);
    setAnalytics(await analyticsResp.json());
    setAnalysis(await analysisResp.json());
    setConfigIntel(await configResp.json());
    setActionStatus('knowledge ready');
    await refresh();
  }

  function attachEvidenceFiles(files: FileList | File[]) {
    const images = Array.from(files).filter((file) => file.type.startsWith('image/'));
    if (!images.length) {
      setActionStatus('drop image evidence only');
      return;
    }
    setEvidence((current) => [
      ...images.map((file) => ({
        id: `${file.name}-${file.lastModified}-${file.size}`,
        name: file.name,
        url: URL.createObjectURL(file),
        size: `${Math.max(1, Math.round(file.size / 1024))} KB`,
        status: 'attached as ticket evidence',
      })),
      ...current,
    ].slice(0, 8));
    setActionStatus(`${images.length} evidence image${images.length === 1 ? '' : 's'} attached`);
  }

  function putTopFindingOnCooldown() {
    const target = topFinding?.asset_name ?? assets.find((asset) => asset.risk === 'high')?.name;
    if (!target) {
      setActionStatus('run CVE scoring first');
      return;
    }
    setCooldownAssets((current) => current.includes(target) ? current : [target, ...current].slice(0, 5));
    setActionStatus(`${target} queued for cooldown`);
  }

  function hardeningNote() {
    if (topFinding) {
      return topFinding.remediation;
    }
    if (cooldownAssets.length) {
      return `Hold traffic for ${cooldownAssets[0]} until owner review, patch window, and config replay complete.`;
    }
    return 'Run CVE scoring, attach evidence, then queue weak infrastructure for cooldown.';
  }

  function changePage(next: string) {
    setPage(next);
    window.location.hash = next;
  }

  useEffect(() => {
    refresh().catch(() => setStatus('degraded'));
    const id = window.setInterval(() => refresh().catch(() => setStatus('degraded')), 8000);
    return () => window.clearInterval(id);
  }, []);

  useEffect(() => {
    const onHash = () => setPage(window.location.hash.replace('#', '') || 'overview');
    window.addEventListener('hashchange', onHash);
    return () => window.removeEventListener('hashchange', onHash);
  }, []);

  const risks = useMemo(() => ({
    high: assets.filter((asset) => asset.risk === 'high').length,
    medium: assets.filter((asset) => asset.risk === 'medium').length,
    low: assets.filter((asset) => asset.risk === 'low').length
  }), [assets]);

  useEffect(() => {
    if (!assets.length) {
      setSelectedAssetId('');
      return;
    }
    if (!selectedAssetId || !assets.some((asset) => asset.id === selectedAssetId)) {
      setSelectedAssetId((assets.find((asset) => asset.risk === 'high') ?? assets[0]).id);
    }
  }, [assets, selectedAssetId]);

  const story = topology?.story ?? {
    mission: 'Keep local office support teams from guessing which endpoint, service, or dependency broke.',
    vision: 'Turn branch network changes into a searchable ServiceNow-ready CMDB stream within minutes.',
    servicenow_use_case: 'Support agents can search office assets, trace dependencies, replay changes, and attach evidence.'
  };
  const endpointCards = useMemo(() => {
    const ids = ['network-branch-wifi', 'endpoint-laptop-fleet', 'network-edge-router', 'servicenow-support-view'];
    const nodes = ids.map((id) => topology?.nodes.find((node) => node.id === id)).filter(Boolean) as TopologyNode[];
    if (nodes.length === ids.length) {
      return nodes.map((node) => ({
        icon: iconForKind(node.kind),
        label: node.label,
        detail: node.detail,
        tone: node.tone || 'blue'
      }));
    }
    return officeEndpoints;
  }, [topology]);
  const meshNodes = useMemo(() => {
    const ids = ['office-collector', 'go-api', 'kafka-stream', 'indexer', 'opensearch', 'servicenow-view'];
    return ids.map((id) => topology?.nodes.find((node) => node.id === id)).filter(Boolean) as TopologyNode[];
  }, [topology]);
  const meshLabels = ['REST upsert', 'Kafka event', 'consumer group', 'document write', 'search read'];
  const latestEvent = events[events.length - 1];
  const topFinding = analysis?.findings?.[0];
  const selectedAsset = assets.find((asset) => asset.id === selectedAssetId) ?? assets[0];
  const selectedFinding = selectedAsset ? analysis?.findings.find((finding) => finding.asset_name === selectedAsset.name) : undefined;
  const latestAutomation = automations[0];
  const gatewayUser = authSession?.user?.email || authSession?.user?.name || 'Local operator';
  const gatewayReady = Boolean(authSession?.authenticated || authSession?.mode === 'dev');

  return (
    <main>
      <header className="topbar">
        <div>
          <p className="eyebrow">Core Infrastructure Inventory</p>
          <h1>Infra Knowledge Console</h1>
          <p className="subtitle">Attach infrastructure data, generate searchable operational knowledge, score CVE risk, and produce ServiceNow-ready evidence.</p>
        </div>
        <button className="iconButton" onClick={() => refresh()} title="Refresh dashboard">
          <RefreshCcw size={18} />
        </button>
      </header>

      <section className="ribbon">
        <div>
          <span>Mission</span>
          <strong>{story.mission}</strong>
        </div>
        <div>
          <span>Vision</span>
          <strong>{story.vision}</strong>
        </div>
        <div>
          <span>ServiceNow use</span>
          <strong>{story.servicenow_use_case}</strong>
        </div>
      </section>

      <section className="commandBar">
        <div>
          <strong>{assets.length} assets</strong>
          <span>{risks.high} high risk · {events.length} stream events · {status} · {attachResult}</span>
        </div>
        <div className="commandActions">
          <button onClick={() => attachInventory()}><PlusCircle size={17} />Add sample</button>
          <button onClick={() => changePage('cloud')}><Cloud size={17} />Cloud gateway</button>
          <button onClick={() => generateKnowledge()}><BarChart3 size={17} />Generate</button>
          <button onClick={() => applyConfigAutomation()}><CheckCircle2 size={17} />Apply guardrail</button>
        </div>
      </section>

      <section className="remediationWorkbench">
        <div
          className="panel dropPanel"
          onDragOver={(event) => event.preventDefault()}
          onDrop={(event) => {
            event.preventDefault();
            attachEvidenceFiles(event.dataTransfer.files);
          }}
        >
          <div className="panelHeader">
            <h2>Evidence Drop</h2>
            <span>{evidence.length} images</span>
          </div>
          <label className="dropZone">
            <Upload size={22} />
            <strong>Drop screenshots or diagrams</strong>
            <span>Use them as ticket evidence for a CVE, failed endpoint, or weak service path.</span>
            <input
              type="file"
              accept="image/*"
              multiple
              onChange={(event) => event.target.files && attachEvidenceFiles(event.target.files)}
            />
          </label>
          <div className="evidenceList">
            {evidence.map((item) => (
              <div className="evidenceItem" key={item.id}>
                <img src={item.url} alt={item.name} />
                <div>
                  <strong>{item.name}</strong>
                  <span>{item.size} · {item.status}</span>
                </div>
              </div>
            ))}
            {!evidence.length && <div className="empty small">No evidence attached yet.</div>}
          </div>
        </div>

        <div className="panel remediationPanel">
          <div className="panelHeader">
            <h2>Remediation Queue</h2>
            <span>{cooldownAssets.length} on cooldown</span>
          </div>
          <div className="hardeningCard">
            <Wrench size={20} />
            <div>
              <strong>{topFinding ? `${topFinding.cve_id} on ${topFinding.asset_name}` : 'Hardening plan'}</strong>
              <span>{hardeningNote()}</span>
            </div>
          </div>
          <div className="workbenchActions">
            <button onClick={() => putTopFindingOnCooldown()}><PauseCircle size={17} />Put on cooldown</button>
            <button onClick={() => reviewTicket()}><TicketCheck size={17} />Review ticket</button>
          </div>
          <div className="cooldownList">
            {cooldownAssets.map((asset) => <code key={asset}>{asset}</code>)}
            {!cooldownAssets.length && <span>Cooldown isolates the riskiest CI while the ticket carries CVE findings, evidence images, and config notes.</span>}
          </div>
          <div className="savedTicketList">
            <strong>Saved tickets</strong>
            {savedTickets.slice(0, 3).map((item) => (
              <button key={item.number} onClick={() => setTicket(item)}>
                <span>{item.number}</span>
                <em>{item.configuration_items.join(', ') || 'no CI'}</em>
              </button>
            ))}
            {!savedTickets.length && <span>No submitted tickets yet.</span>}
          </div>
        </div>
      </section>

      <nav className="pageNav">
        {[
          ['overview', 'Overview'],
          ['security', 'Security'],
          ['config', 'Config'],
          ['cloud', 'Cloud gateway'],
          ['org', 'Org access'],
          ['data', 'Data']
        ].map(([id, label]) => (
          <button className={page === id ? 'active' : ''} key={id} onClick={() => changePage(id)}>{label}</button>
        ))}
      </nav>

      {page === 'cloud' && (
        <section className="cloudGateway">
          <div className="gatewayHero">
            <div>
              <p className="eyebrow">Cloud Integration Gateway</p>
              <h2>{gatewayReady ? `Ready as ${gatewayUser}` : 'Sign in to connect cloud inventory'}</h2>
              <span>Use this gateway before importing Azure, Kubernetes, API, or YAML inventory into the incident-intelligence graph.</span>
            </div>
            <div className="gatewayActions">
              <button onClick={() => startCloudGateway()}><UserRound size={17} />{authSession?.authenticated ? 'Refresh session' : 'Start login'}</button>
              <button onClick={() => importAzureSample('local')}><Cloud size={17} />Use local Azure sample</button>
            </div>
          </div>

          <div className="gatewayGrid">
            <div className="gatewayCard active">
              <Cloud size={22} />
              <strong>Azure</strong>
              <span>{authSession?.mode === 'azure' ? 'Entra ID + OIDC gateway' : 'Local gateway session; Azure SSO configurable'}</span>
              <code>{authSession?.storage ?? 'encrypted HttpOnly SameSite cookie'}</code>
              <button onClick={() => importAzureOrg()}>Import Azure inventory</button>
            </div>
            <div className="gatewayCard">
              <Network size={22} />
              <strong>Kubernetes</strong>
              <span>Attach AKS/K8s manifests or exported cluster inventory as YAML.</span>
              <code>make import FILE=examples/local-office.yaml</code>
            </div>
            <div className="gatewayCard">
              <Box size={22} />
              <strong>API inventory</strong>
              <span>Post services, endpoints, queues, and dependencies directly into the Go API.</span>
              <code>POST /api/assets</code>
            </div>
          </div>

          <div className="shiftPanel">
            <div>
              <strong>Shift detection mode</strong>
              <span>{shiftMode === 'infra' ? 'Watching asset, dependency, cloud, and CVE drift.' : 'Watching API/service ownership, version, and dependency drift.'}</span>
            </div>
            <div className="segmented">
              <button className={shiftMode === 'infra' ? 'active' : ''} onClick={() => setShiftMode('infra')}>Infra</button>
              <button className={shiftMode === 'api' ? 'active' : ''} onClick={() => setShiftMode('api')}>API</button>
            </div>
          </div>
        </section>
      )}

      <section className="stats">
        <Metric icon={<Server />} label="Assets" value={assets.length.toString()}>
          <MetricDetail label="ServiceNow CI types" value={[...new Set(assets.map((asset) => asset.type))].join(', ') || 'waiting for seed'} />
          <MetricDetail label="Owners" value={[...new Set(assets.map((asset) => asset.owner))].slice(0, 4).join(', ') || 'none'} />
        </Metric>
        <Metric icon={<Activity />} label="Stream" value={status}>
          <MetricDetail label="Kafka topic" value="inventory.events" />
          <MetricDetail label="Latest event" value={latestEvent ? latestEvent.payload.name : 'waiting for events'} />
        </Metric>
        <Metric icon={<GitBranch />} label="Config" value={config?.version ?? 'unknown'}>
          <MetricDetail label="Config source" value="hot-reloaded JSON / Azure Blob path" />
          <MetricDetail label="Replay window" value={`${config?.replay_window ?? '-'} events`} />
          <MetricDetail label="Proposal" value={configIntel ? `${configIntel.proposals.length} changes` : 'not generated'} />
        </Metric>
        <Metric icon={<ShieldAlert />} label="Risk" value={`${risks.high} high`}>
          <MetricDetail label="CVE model" value={analysis?.model ?? 'weighted-logistic-cve-risk-v1'} />
          <MetricDetail label="Findings" value={analysis ? `${analysis.findings.length} scored` : 'run live CVE analysis'} />
          <MetricDetail label="Top risk" value={topFinding ? `${topFinding.asset_name} ${topFinding.score.toFixed(1)}` : 'waiting'} />
        </Metric>
        <Metric icon={<Database />} label="DB" value="OpenSearch">
          <MetricDetail label="Index" value="inventory-assets" />
          <MetricDetail label="Compatibility" value="Elasticsearch-style search API" />
        </Metric>
        <Metric icon={<ShieldCheck />} label="Identity" value={authSession?.mode ?? 'dev'}>
          <MetricDetail label="Session" value={authSession?.authenticated ? 'signed in' : 'local anonymous'} />
          <MetricDetail label="Browser storage" value={authSession?.storage ?? 'encrypted cookie'} />
        </Metric>
      </section>

      <section className={`inventoryConsole ${!['overview', 'data'].includes(page) ? 'hidden' : ''}`}>
        <div className="panel assetListPanel">
          <div className="panelHeader">
            <h2>Inventory</h2>
            <span>{assets.length} indexed</span>
          </div>
          <div className="searchMini">
            <Search size={17} />
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={(event) => event.key === 'Enter' && refresh()}
              placeholder="Search assets, owners, services"
            />
            <button onClick={() => refresh()}>Search</button>
          </div>
          <div className="assetRows">
            {assets.map((asset) => {
              const finding = analysis?.findings.find((item) => item.asset_name === asset.name);
              return (
                <button
                  className={`assetRow ${selectedAsset?.id === asset.id ? 'selected' : ''}`}
                  key={asset.id}
                  onClick={() => setSelectedAssetId(asset.id)}
                >
                  <span className={`nodeDot ${asset.risk}`} />
                  <span>
                    <strong>{asset.name}</strong>
                    <em>{asset.type} · {asset.service}</em>
                  </span>
                  <code>{finding ? finding.score.toFixed(1) : asset.risk}</code>
                </button>
              );
            })}
            {assets.length === 0 && <div className="empty">Attach JSON or import Azure sample to start.</div>}
          </div>
        </div>

        <div className="panel topologyPanel">
          <div className="panelHeader">
            <h2>Live Infra Map</h2>
            <span>{assets.reduce((count, asset) => count + asset.dependencies.length, 0)} links</span>
          </div>
          <TopologyMap
            assets={assets}
            selectedId={selectedAsset?.id ?? ''}
            findings={analysis?.findings ?? []}
            onSelect={setSelectedAssetId}
          />
        </div>

        <aside className="panel inspectorPanel">
          <div className="panelHeader">
            <h2>Asset Detail</h2>
            <span>{selectedAsset?.risk ?? 'none'}</span>
          </div>
          {selectedAsset ? (
            <div className="assetInspector">
              <div>
                <strong>{selectedAsset.name}</strong>
                <span>{selectedAsset.type} · {selectedAsset.owner} · {selectedAsset.region}</span>
              </div>
              <div className="inspectorStats">
                <span><b>{selectedAsset.dependencies.length}</b> deps</span>
                <span><b>{selectedFinding ? selectedFinding.score.toFixed(1) : '-'}</b> CVE</span>
                <span><b>{selectedAsset.version}</b> version</span>
              </div>
              <div className="dependencyChips">
                {selectedAsset.dependencies.map((dep) => <button key={dep} onClick={() => setSelectedAssetId(dep)}>{dep}</button>)}
                {!selectedAsset.dependencies.length && <span>No dependencies</span>}
              </div>
              <div className="findingNote">
                <ShieldAlert size={18} />
                <span>{selectedFinding?.remediation ?? 'Run CVE scoring to attach risk and remediation to this asset.'}</span>
              </div>
              <div className="workbenchActions tightActions">
                <button onClick={() => startAnalysis()}><ShieldAlert size={17} />Score</button>
                <button onClick={() => putTopFindingOnCooldown()}><PauseCircle size={17} />Cooldown</button>
                <button onClick={() => reviewTicket()}><TicketCheck size={17} />Ticket</button>
              </div>
            </div>
          ) : <div className="empty">No selected asset.</div>}
        </aside>
      </section>

      <section className={`automationConsole ${page !== 'config' && page !== 'overview' ? 'hidden' : ''}`}>
        <div className="panel automationPanel">
          <div className="panelHeader">
            <h2>Configuration Automation</h2>
            <span>{latestAutomation?.status ?? 'ready'}</span>
          </div>
          <div className="automationBody">
            <div className="automationHero">
              <SlidersHorizontal size={20} />
              <div>
                <strong>{latestAutomation?.policy ?? 'inventory-risk-guardrail'}</strong>
                <span>{latestAutomation ? `${latestAutomation.guardrails.length} guardrails · replay window ${latestAutomation.replay_window}` : 'Apply routing, evidence, and stateful-asset guardrails from current inventory.'}</span>
              </div>
            </div>
            <div className="guardrailGrid">
              {(latestAutomation?.guardrails ?? ['route high CVEs by owner', 'require stateful dependency evidence', 'align search fields to support lookups']).map((item) => <code key={item}>{item}</code>)}
            </div>
            <div className="workbenchActions tightActions">
              <button onClick={() => buildConfig()}><BarChart3 size={17} />Preview</button>
              <button onClick={() => applyConfigAutomation()}><CheckCircle2 size={17} />Apply guardrail</button>
            </div>
          </div>
        </div>

        <div className="panel azurePanel">
          <div className="panelHeader">
            <h2>Azure Integration</h2>
            <span>{imports[0]?.provider ?? 'dev'}</span>
          </div>
          <div className="azureBody">
            <Cloud size={22} />
            <strong>{imports[0]?.account_id ?? 'No cloud import yet'}</strong>
            <span>{imports[0] ? `${imports[0].imported_assets} assets · ${imports[0].encrypted_in_transit}` : 'Import an Azure subscription sample or wire Entra ID in cloud mode.'}</span>
            <button onClick={() => importAzureOrg()}><Cloud size={17} />Connect / import</button>
          </div>
        </div>
      </section>

      <section className={`pipeline ${page !== 'overview' ? 'hidden' : ''}`}>
        <Step tone="yellow" icon={<Building2 />} label="YAML / Cloud Import" detail="existing infra files or Azure account data" />
        <Step tone="blue" icon={<Box />} label="Go API" detail="validates and publishes changes" />
        <Step tone="green" icon={<Database />} label="OpenSearch" detail="CMDB-style searchable state" />
        <Step tone="sky" icon={<Cloud />} label="Azure Path" detail="Event Hubs + Blob config" />
      </section>

      <section className={`actionGrid ${!['overview', 'security', 'config'].includes(page) ? 'hidden' : ''}`}>
        <div className="panel actionPanel">
          <div className="panelHeader">
            <h2>Generate Knowledge</h2>
            <span>{actionStatus}</span>
          </div>
          <div className="actionButtons">
            <button onClick={() => generateAnalytics()}>
              <BarChart3 size={17} />
              Refresh signals
            </button>
            <button onClick={() => startAnalysis()}>
              <ShieldAlert size={17} />
              Score CVEs
            </button>
            <button onClick={() => buildConfig()}>
              <SlidersHorizontal size={17} />
              Suggest config
            </button>
            <button onClick={() => reviewTicket()}>
              <PlayCircle size={17} />
              Review ticket
            </button>
          </div>
          <div className="signalGrid">
            {(analytics?.signals ?? []).map((signal) => (
              <div className={`signal ${signal.status}`} key={signal.label}>
                <span>{signal.label}</span>
                <strong>{signal.value}</strong>
                <em>{signal.status}</em>
              </div>
            ))}
          </div>
        </div>

        <div className="panel actionPanel">
          <div className="panelHeader">
            <h2>{ticket?.number ?? analysis?.analysis_id ?? 'Live Feed'}</h2>
            <ClipboardCheck size={18} />
          </div>
          {ticket ? (
            <div className="brief">
              <strong>{ticket.short_description}</strong>
              <span>{ticket.table} · priority {ticket.priority} · {ticket.assignment_group}</span>
              <div>
                {ticket.configuration_items.map((asset) => <code key={asset}>{asset}</code>)}
              </div>
            </div>
          ) : analysis ? (
            <div className="brief">
              <strong>{analysis.summary}</strong>
              <span>{analysis.model} · {analysis.search_backend} · Kafka {analysis.kafka_topic}</span>
              <div>
                {analysis.findings.slice(0, 4).map((finding) => <code key={`${finding.cve_id}-${finding.asset_name}`}>{finding.cve_id} · {finding.asset_name}</code>)}
              </div>
            </div>
          ) : (
            <div className="feed compactFeed">
              {(analytics?.live_feed ?? []).map((item) => (
                <div className="event" key={`${item.title}-${item.timestamp}`}>
                  <strong>{item.title}</strong>
                  <span>{item.detail} · {new Date(item.timestamp).toLocaleTimeString()}</span>
                </div>
              ))}
              {!analytics?.live_feed?.length && <div className="empty small">No live feed yet.</div>}
            </div>
          )}
        </div>
      </section>

      <section className={`intelligenceGrid ${!['overview', 'security', 'config'].includes(page) ? 'hidden' : ''}`}>
        <details className="panel insightPanel" open>
          <summary>
            <span>Real-time CVE Risk Scoring</span>
            <ChevronDown size={18} />
          </summary>
          <div className="insightBody">
            <div className="modelLine">
              <strong>{analysis?.incident_brief ? `${analysis.incident_brief.provider} · ${analysis.incident_brief.model}` : analysis?.model ?? 'weighted-logistic-cve-risk-v1'}</strong>
              <span>{analysis?.incident_brief?.executive_summary ?? analysis?.summary ?? 'Run analysis to score every indexed infrastructure asset from OpenSearch.'}</span>
            </div>
            {analysis?.incident_brief && (
              <div className="aiBrief">
                <div>
                  <strong>Probable cause</strong>
                  <span>{analysis.incident_brief.probable_cause}</span>
                </div>
                <div>
                  <strong>Blast radius</strong>
                  <span>{analysis.incident_brief.blast_radius}</span>
                </div>
                <div>
                  <strong>Actions</strong>
                  {analysis.incident_brief.recommended_actions.map((item) => <code key={item}>{item}</code>)}
                </div>
              </div>
            )}
            <div className="findingList">
              {(analysis?.findings ?? []).slice(0, 6).map((finding) => (
                <div className="finding" key={`${finding.cve_id}-${finding.asset_name}`}>
                  <div>
                    <strong>{finding.asset_name}</strong>
                    <span>{finding.cve_id} · {finding.severity} · confidence {(finding.confidence * 100).toFixed(0)}%</span>
                  </div>
                  <div className="scoreBar" aria-label={`Risk score ${finding.score.toFixed(1)}`}>
                    <span style={{ width: `${Math.min(100, finding.score * 10)}%` }} />
                  </div>
                  <code>{finding.remediation}</code>
                </div>
              ))}
              {!analysis && <div className="empty">Click Score CVEs to score live assets from the OpenSearch read model.</div>}
            </div>
          </div>
        </details>

        <details className="panel insightPanel" open>
          <summary>
            <span>Configuration Recommendations</span>
            <ChevronDown size={18} />
          </summary>
          <div className="insightBody">
            <div className="modelLine">
              <strong>{configIntel?.model ?? 'config-recommendation-rules-v1'}</strong>
              <span>{configIntel?.summary ?? 'Suggest config from owners, services, CVE pressure, and Kafka replay behavior.'}</span>
            </div>
            <div className="proposalList">
              {(configIntel?.proposals ?? []).map((proposal) => (
                <div className="proposal" key={proposal.field}>
                  <div>
                    <strong>{proposal.field}</strong>
                    <em>{(proposal.confidence * 100).toFixed(0)}% confidence</em>
                  </div>
                  <span>{proposal.reason}</span>
                  <code>{proposal.current} {'->'} {proposal.proposed}</code>
                </div>
              ))}
              {!configIntel && <div className="empty">Click Suggest config to generate a ServiceNow-aware preview config from indexed infrastructure.</div>}
            </div>
          </div>
        </details>
      </section>

      {page === 'org' && (
        <section className="intelligenceGrid">
          <details className="panel insightPanel" open>
            <summary>
              <span>Azure SSO and Secure Session</span>
              <ChevronDown size={18} />
            </summary>
            <div className="insightBody">
              <div className="modelLine">
                <strong>{authSession?.authenticated ? authSession.user?.email || authSession.user?.name || 'Signed in' : 'Azure Entra ID ready'}</strong>
                <span>{authSession?.transport}</span>
              </div>
              <div className="proposalList">
                <div className="proposal">
                  <div><strong>OIDC flow</strong><em>{authSession?.mode ?? 'dev'}</em></div>
                  <span>Azure mode uses Microsoft identity platform discovery, authorization-code flow with PKCE, and verified ID tokens.</span>
                  <code>AUTH_MODE=azure · AZURE_TENANT_ID · AZURE_CLIENT_ID · AZURE_CLIENT_SECRET</code>
                </div>
                <div className="proposal">
                  <div><strong>Browser data</strong><em>encrypted</em></div>
                  <span>Session claims are stored in an AES-GCM encrypted HttpOnly SameSite cookie. The UI does not store raw user data in localStorage.</span>
                  <code>{authSession?.storage ?? 'encrypted HttpOnly SameSite cookie'}</code>
                </div>
              </div>
              <a className="linkButton" href={`${apiBase}${authSession?.login_url ?? '/auth/login'}`}>
                <UserRound size={17} />
                Start sign-in
              </a>
            </div>
          </details>

          <details className="panel insightPanel" open>
            <summary>
              <span>Organization Inventory Import</span>
              <ChevronDown size={18} />
            </summary>
            <div className="insightBody">
              <div className="modelLine">
                <strong>Bring cloud account data into the knowledge graph</strong>
                <span>Import account inventory, publish each asset as Kafka events, then index the searchable state in OpenSearch for CVE scoring and configuration recommendations.</span>
              </div>
              <button onClick={() => importAzureOrg()}>
                <Cloud size={17} />
                Connect / import Azure
              </button>
              <div className="proposalList">
                {imports.map((item) => (
                  <div className="proposal" key={item.import_id}>
                    <div><strong>{item.provider} · {item.account_id}</strong><em>{item.imported_assets} assets</em></div>
                    <span>{item.encrypted_in_transit}</span>
                    <code>{item.encrypted_at_rest}</code>
                  </div>
                ))}
                {!imports.length && <div className="empty">No organization import has run in this session.</div>}
              </div>
            </div>
          </details>
        </section>
      )}

      <details className={`disclosure ${!['overview', 'security'].includes(page) ? 'hidden' : ''}`}>
        <summary>
          <span>Ticket Evidence</span>
          <ChevronDown size={18} />
        </summary>
        <section className="assistGrid">
          <div className="assistCard">
            <strong>Native ServiceNow shape</strong>
            <span>Incident/change record with configuration items, priority, assignment group, work notes, and evidence links.</span>
            <code>{ticket?.number ?? 'POST /api/servicenow/tickets'}</code>
          </div>
          <div className="assistCard">
            <strong>Assistant context</strong>
            <span>CVE findings, impacted CIs, Kafka replay window, OpenSearch query evidence, and topology dependencies.</span>
            <code>{analysis?.analysis_id ?? 'POST /api/security/analyze'}</code>
          </div>
          <div className="assistCard">
            <strong>Operator action</strong>
            <span>Run CVE analysis live, create the ServiceNow ticket payload, then attach the generated evidence to a review.</span>
            <code>{ticket ? ticket.configuration_items.join(', ') : 'no ticket created yet'}</code>
          </div>
        </section>
      </details>

      <section className={`meshPanel ${page !== 'data' ? 'hidden' : ''}`}>
        <div className="panelHeader">
          <h2>Local Office Topology</h2>
          <span>ServiceNow support path</span>
        </div>
        <div className="endpointGrid">
          {endpointCards.map((endpoint) => (
            <EndpointCard key={endpoint.label} {...endpoint} />
          ))}
        </div>
        <div className="meshMap">
          {(meshNodes.length ? meshNodes : [
            { id: 'office-collector', label: 'Office Collector', kind: 'network', owner: 'it-ops', risk: 'low', detail: 'branch network signals', tone: 'blue' },
            { id: 'go-api', label: 'Go API', kind: 'service', owner: 'core-infra', risk: 'medium', detail: 'rate limit + DFS guard', tone: 'violet' },
            { id: 'kafka-stream', label: 'Kafka Stream', kind: 'queue', owner: 'data', risk: 'low', detail: 'replayable events', tone: 'green' },
            { id: 'indexer', label: 'Indexer', kind: 'worker', owner: 'search', risk: 'medium', detail: 'OpenSearch writer', tone: 'violet' },
            { id: 'opensearch', label: 'OpenSearch', kind: 'database', owner: 'search', risk: 'medium', detail: 'search index', tone: 'green' },
            { id: 'servicenow-view', label: 'ServiceNow View', kind: 'frontend', owner: 'product', risk: 'low', detail: 'CMDB/change evidence', tone: 'rose' },
          ]).map((node, index, nodes) => (
            <React.Fragment key={node.id}>
              <MeshNode tone={node.tone} title={node.label} subtitle={node.detail} />
              {index < nodes.length - 1 && <span className="meshLink">{meshLabels[index]}</span>}
            </React.Fragment>
          ))}
        </div>
      </section>

      <details className={`disclosure ${!['overview', 'data'].includes(page) ? 'hidden' : ''}`}>
        <summary>
          <span>Developer Setup and Data Contracts</span>
          <ChevronDown size={18} />
        </summary>
        <section className="explainGrid">
          <div className="panel compactPanel">
            <div className="panelHeader">
              <h2>Onboard Your Data</h2>
              <span>local first</span>
            </div>
            <div className="apiList">
              <CodeLine method="1" path="make demo" note="starts Go API, indexer, web, Kafka-compatible Redpanda, and OpenSearch" />
              <CodeLine method="2" path="make import FILE=examples/local-office.yaml" note="loads existing YAML/JSON inventory into the API" />
              <CodeLine method="3" path="POST /api/security/analyze" note="scores CVE risk from the indexed infrastructure graph" />
              <CodeLine method="4" path="POST /api/config/automation" note="applies a guardrail preview from observed services, owners, and risk" />
            </div>
          </div>

          <div className="panel compactPanel">
            <div className="panelHeader">
              <h2>APIs</h2>
              <span>live endpoints</span>
            </div>
            <div className="apiList">
              <CodeLine method="POST" path="/api/assets" note="publish inventory change" />
              <CodeLine method="GET" path="/api/assets/search" note="OpenSearch query" />
              <CodeLine method="GET" path="/api/topology" note="mesh read model" />
              <CodeLine method="GET" path="/api/analytics" note="ops snapshot" />
              <CodeLine method="POST" path="/api/security/analyze" note="ML-style CVE scoring" />
              <CodeLine method="POST" path="/api/config/recommendations" note="configuration proposal" />
              <CodeLine method="POST" path="/api/config/automation" note="apply guardrail preview" />
              <CodeLine method="POST" path="/api/servicenow/tickets" note="ServiceNow ticket" />
              <CodeLine method="GET" path="/api/auth/session" note="encrypted session status" />
              <CodeLine method="POST" path="/api/org/import" note="cloud account inventory import" />
              <CodeLine method="GET" path="/api/stream/recent" note="event replay buffer" />
              <CodeLine method="GET" path="/api/config/version" note="hot reload proof" />
            </div>
          </div>

          <div className="panel compactPanel">
            <div className="panelHeader">
              <h2>GraphQL-style / Data</h2>
              <span>contracts</span>
            </div>
            <div className="contractList">
              {dataContracts.map((contract) => (
                <div className="contract" key={contract.name}>
                  <strong>{contract.name}</strong>
                  <span>{contract.type}</span>
                  <code>{contract.shape}</code>
                  <em>{contract.note}</em>
                </div>
              ))}
            </div>
          </div>
        </section>
      </details>

      <details className={`disclosure ${page !== 'config' ? 'hidden' : ''}`} open>
        <summary>
          <span>Controls and ServiceNow Fit</span>
          <LockKeyhole size={18} />
        </summary>
        <div className="panel compactPanel">
            <div className="controlGrid">
              <Control label="ServiceNow CMDB" value="inventory-ready assets" />
              <Control label="Incident support" value="endpoint dependency search" />
              <Control label="NIST CSF" value="identify, protect, detect" />
              <Control label="SOC 2 / CIS" value="change stream + config checks" />
            </div>
          {configIntel && (
            <div className="routingGrid">
              {Object.entries(configIntel.servicenow_routing).map(([key, value]) => (
                <div className="routing" key={key}>
                  <strong>{key}</strong>
                  <span>{value}</span>
                </div>
              ))}
            </div>
          )}
          <div className="teamGrid">
            {(topology?.teams ?? []).map((team) => (
              <div className="team" key={team.team}>
                <strong>{team.team}</strong>
                <span>{team.need}</span>
                <code>{team.served_by}</code>
              </div>
            ))}
          </div>
        </div>
      </details>

      <section className={`toolbar ${page !== 'data' ? 'hidden' : ''}`}>
        <Search size={18} />
        <input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          onKeyDown={(event) => event.key === 'Enter' && refresh()}
          placeholder="Search ServiceNow asset, office endpoint, owner, region, or risk"
        />
        <button onClick={() => refresh()}>Search</button>
      </section>

      <section className={`grid ${page !== 'data' ? 'hidden' : ''}`}>
        <div className="panel inventoryPanel">
          <div className="panelHeader">
            <h2>Inventory</h2>
            <span>{assets.length} indexed</span>
          </div>
          <div className="table">
            <div className="row head">
              <span>Name</span><span>Type</span><span>Owner</span><span>Env</span><span>Risk</span>
            </div>
            {assets.map((asset) => (
              <div className="row" key={asset.id}>
                <span className="strong">{asset.name}</span>
                <span>{asset.type}</span>
                <span>{asset.owner}</span>
                <span>{asset.environment}</span>
                <span className={`risk ${asset.risk}`}>{asset.risk}</span>
              </div>
            ))}
            {assets.length === 0 && (
              <div className="empty">Attach JSON or import the sample to start.</div>
            )}
          </div>
        </div>

        <aside className="panel">
          <div className="panelHeader">
            <h2>Stream</h2>
            <Database size={18} />
          </div>
          <div className="feed">
            {events.slice(-8).reverse().map((event) => (
              <div className="event" key={event.event_id}>
                <strong>{event.payload.name}</strong>
                <span>{event.event_type} · {new Date(event.timestamp).toLocaleTimeString()}</span>
              </div>
            ))}
            {events.length === 0 && <div className="empty small">No events yet.</div>}
          </div>
          <div className="configLine">
            <span>Replay window</span>
            <strong>{config?.replay_window ?? '-'}</strong>
          </div>
        </aside>
      </section>

      {ticketModalOpen && ticketDraft && (
        <div className="modalBackdrop" role="presentation" onClick={() => setTicketModalOpen(false)}>
          <section className="modalPanel" role="dialog" aria-modal="true" aria-label="Review ServiceNow ticket" onClick={(event) => event.stopPropagation()}>
            <div className="modalHeader">
              <div>
                <span>Ticket Review</span>
                <h2>{ticketDraft.number}</h2>
              </div>
              <button className="iconButton" onClick={() => setTicketModalOpen(false)} title="Close ticket review">X</button>
            </div>
            <div className="ticketReview">
              <div className="reviewHero">
                <TicketCheck size={22} />
                <div>
                  <strong>{ticketDraft.short_description}</strong>
                  <span>{ticketDraft.table} · priority {ticketDraft.priority} · {ticketDraft.assignment_group}</span>
                </div>
              </div>
              <div className="reviewGrid">
                <div>
                  <strong>Configuration items</strong>
                  <div className="dependencyChips">
                    {ticketDraft.configuration_items.map((item) => <button key={item}>{item}</button>)}
                    {!ticketDraft.configuration_items.length && <span>No impacted CI found.</span>}
                  </div>
                </div>
                <div>
                  <strong>Evidence</strong>
                  <span>{evidence.length} image attachments · {events.length} replay events · {analysis?.findings.length ?? 0} CVE findings</span>
                </div>
              </div>
              <div className="workNotes">
                {ticketDraft.work_notes.map((note) => <code key={note}>{note}</code>)}
              </div>
              <div className="modalActions">
                <button onClick={() => setTicketModalOpen(false)}>Keep draft</button>
                <button onClick={() => submitTicket()}><CheckCircle2 size={17} />Submit ticket</button>
              </div>
            </div>
          </section>
        </div>
      )}

      {azureModalOpen && (
        <div className="modalBackdrop" role="presentation" onClick={() => setAzureModalOpen(false)}>
          <section className="modalPanel azureModal" role="dialog" aria-modal="true" aria-label="Connect Azure inventory" onClick={(event) => event.stopPropagation()}>
            <div className="modalHeader">
              <div>
                <span>Azure Inventory</span>
                <h2>Connect Entra ID or use a local sample</h2>
              </div>
              <button className="iconButton" onClick={() => setAzureModalOpen(false)} title="Close Azure onboarding">X</button>
            </div>
            <div className="ticketReview">
              <div className="reviewHero">
                <Cloud size={22} />
                <div>
                  <strong>Local mode has no Azure token.</strong>
                  <span>Use the sample to fingerprint Azure Firewall, AKS, container worker, and Key Vault assets, or enable Azure SSO to route imports through Entra ID.</span>
                </div>
              </div>
              <div className="workNotes">
                <code>AUTH_MODE=azure</code>
                <code>AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET</code>
                <code>AUTH_SESSION_KEY=$(openssl rand -base64 32)</code>
              </div>
              <div className="modalActions">
                <a className="linkButton" href={`${apiBase}${authSession?.login_url ?? '/auth/login'}`}>Start SSO</a>
                <button onClick={() => importAzureSample('local')}><Cloud size={17} />Use local Azure sample</button>
              </div>
            </div>
          </section>
        </div>
      )}
    </main>
  );
}

function TopologyMap({
  assets,
  selectedId,
  findings,
  onSelect,
}: {
  assets: Asset[];
  selectedId: string;
  findings: CVEAnalysis['findings'];
  onSelect: (id: string) => void;
}) {
  const visible = assets.slice(0, 18);
  const positions = new Map<string, { x: number; y: number }>();
  const cols = Math.min(6, Math.max(1, Math.ceil(Math.sqrt(Math.max(visible.length, 1)))));
  visible.forEach((asset, index) => {
    const col = index % cols;
    const row = Math.floor(index / cols);
    positions.set(asset.id, {
      x: 76 + col * 132 + (row % 2) * 26,
      y: 72 + row * 112,
    });
  });
  const links = visible.flatMap((asset) => (
    asset.dependencies
      .filter((dep) => positions.has(dep))
      .map((dep) => ({ from: asset.id, to: dep }))
  ));
  const height = Math.max(260, 120 + Math.ceil(visible.length / Math.max(cols, 1)) * 112);
  const scoreByAsset = new Map(findings.map((finding) => [finding.asset_name, finding.score]));

  return (
    <div className="topologyCanvas">
      <svg viewBox={`0 0 840 ${height}`} role="img" aria-label="Live infrastructure dependency map">
        {links.map((link) => {
          const from = positions.get(link.from)!;
          const to = positions.get(link.to)!;
          return <line key={`${link.from}-${link.to}`} x1={from.x} y1={from.y} x2={to.x} y2={to.y} />;
        })}
        {visible.map((asset) => {
          const point = positions.get(asset.id)!;
          const score = scoreByAsset.get(asset.name);
          return (
            <g
              className={`topologyNode ${asset.risk} ${selectedId === asset.id ? 'selected' : ''}`}
              key={asset.id}
              onClick={() => onSelect(asset.id)}
              tabIndex={0}
              role="button"
            >
              <circle cx={point.x} cy={point.y} r={26} />
              <text x={point.x} y={point.y + 4} textAnchor="middle">{shortKind(asset.type)}</text>
              <text className="nodeLabel" x={point.x} y={point.y + 45} textAnchor="middle">{asset.name}</text>
              <text className="nodeScore" x={point.x} y={point.y + 62} textAnchor="middle">{score ? `CVE ${score.toFixed(1)}` : asset.owner}</text>
            </g>
          );
        })}
      </svg>
      {!assets.length && <div className="empty">Import or attach inventory to plot the dependency map.</div>}
    </div>
  );
}

function shortKind(kind: string) {
  switch (kind) {
    case 'database':
      return 'DB';
    case 'network':
      return 'NET';
    case 'endpoint':
      return 'END';
    case 'queue':
      return 'Q';
    case 'worker':
      return 'WK';
    case 'frontend':
      return 'UI';
    case 'kubernetes':
      return 'K8S';
    case 'container':
      return 'CTR';
    case 'keyvault':
      return 'KV';
    default:
      return 'API';
  }
}

function Step({ icon, label, detail, tone }: { icon: React.ReactNode; label: string; detail: string; tone: string }) {
  return (
    <div className={`step ${tone}`}>
      <div className="metricIcon">{icon}</div>
      <strong>{label}</strong>
      <span>{detail}</span>
    </div>
  );
}

function MeshNode({ tone, title, subtitle }: { tone: string; title: string; subtitle: string }) {
  return (
    <div className={`meshNode ${tone}`}>
      <Network size={18} />
      <strong>{title}</strong>
      <span>{subtitle}</span>
    </div>
  );
}

function EndpointCard({ icon, label, detail, tone }: { icon: React.ReactNode; label: string; detail: string; tone: string }) {
  return (
    <div className={`endpoint ${tone}`}>
      <div className="metricIcon">{icon}</div>
      <strong>{label}</strong>
      <span>{detail}</span>
    </div>
  );
}

function iconForKind(kind: string) {
  switch (kind) {
    case 'endpoint':
      return <Monitor />;
    case 'network':
      return <Router />;
    case 'frontend':
      return <TicketCheck />;
    case 'database':
      return <Database />;
    case 'queue':
      return <Activity />;
    default:
      return <Server />;
  }
}

function CodeLine({ method, path, note }: { method: string; path: string; note: string }) {
  return (
    <div className="codeLine">
      <span>{method}</span>
      <code>{path}</code>
      <em>{note}</em>
    </div>
  );
}

function Control({ label, value }: { label: string; value: string }) {
  return (
    <div className="control">
      <strong>{label}</strong>
      <span>{value}</span>
    </div>
  );
}

function Metric({ icon, label, value, children }: { icon: React.ReactNode; label: string; value: string; children?: React.ReactNode }) {
  return (
    <details className="metric">
      <summary>
        <div className="metricIcon">{icon}</div>
        <span>{label}</span>
        <strong>{value}</strong>
        <ChevronDown size={16} />
      </summary>
      <div className="metricDetails">{children}</div>
    </details>
  );
}

function MetricDetail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

createRoot(document.getElementById('root')!).render(<App />);
