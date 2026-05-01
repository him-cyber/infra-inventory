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

const apiBase = import.meta.env.VITE_API_URL ?? '';

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
  const [authSession, setAuthSession] = useState<AuthSession | null>(null);
  const [imports, setImports] = useState<OrgImport[]>([]);
  const [actionStatus, setActionStatus] = useState('ready');
  const [query, setQuery] = useState('');
  const [config, setConfig] = useState<{ version: string; updated_at: string; replay_window: number } | null>(null);
  const [status, setStatus] = useState('checking');

  async function refresh() {
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    const [searchResp, eventResp, configResp, topologyResp, analyticsResp, healthResp, authResp, importsResp] = await Promise.all([
      fetch(`${apiBase}/api/assets/search?${params.toString()}`),
      fetch(`${apiBase}/api/stream/recent`),
      fetch(`${apiBase}/api/config/version`),
      fetch(`${apiBase}/api/topology`),
      fetch(`${apiBase}/api/analytics`),
      fetch(`${apiBase}/healthz`),
      fetch(`${apiBase}/api/auth/session`),
      fetch(`${apiBase}/api/org/imports`)
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

  async function createTicket() {
    setActionStatus('creating ServiceNow ticket');
    const response = await fetch(`${apiBase}/api/servicenow/tickets`, { method: 'POST' });
    setTicket(await response.json());
    setActionStatus('ServiceNow ticket ready');
  }

  async function buildConfig() {
    setActionStatus('learning better config from inventory');
    const response = await fetch(`${apiBase}/api/config/recommendations`, { method: 'POST' });
    setConfigIntel(await response.json());
    setActionStatus('config recommendations ready');
  }

  async function importAzureOrg() {
    setActionStatus('importing Azure subscription inventory');
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
          { id: 'az-aks-support', type: 'service', name: 'AKS Support Workloads', owner: 'platform', environment: 'prod', region: 'eastus', service: 'support-platform', version: 29, dependencies: ['queue-inventory-events', 'opensearch-assets'], risk: 'medium' },
          { id: 'az-kv-config', type: 'database', name: 'Key Vault Config Store', owner: 'security', environment: 'prod', region: 'eastus', service: 'secure-config', version: 13, dependencies: [], risk: 'medium' }
        ]
      })
    });
    const result = await response.json();
    setImports((current) => [result, ...current].slice(0, 10));
    setActionStatus('Azure inventory imported');
    await refresh();
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

  return (
    <main>
      <header className="topbar">
        <div>
          <p className="eyebrow">Core Infrastructure Inventory</p>
          <h1>ServiceNow Office Support Topology</h1>
          <p className="subtitle">A local-office inventory pipeline for ServiceNow-style CMDB, support workflow, and change evidence. Go services stream endpoint updates through Kafka and index searchable operational state in OpenSearch.</p>
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

      <section className="launchStrip">
        <div>
          <span>Launch</span>
          <code>make demo</code>
        </div>
        <div>
          <span>Import YAML</span>
          <code>make import FILE=examples/local-office.yaml</code>
        </div>
        <div>
          <span>Read</span>
          <code>OpenSearch + Kafka feed the CVE and config models</code>
        </div>
        <div>
          <span>Secure</span>
          <code>make security-login && make security-export</code>
        </div>
      </section>

      <nav className="pageNav">
        {[
          ['overview', 'Overview'],
          ['security', 'Security'],
          ['config', 'Config'],
          ['org', 'Org access'],
          ['data', 'Data']
        ].map(([id, label]) => (
          <button className={page === id ? 'active' : ''} key={id} onClick={() => changePage(id)}>{label}</button>
        ))}
      </nav>

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
          <MetricDetail label="AI proposal" value={configIntel ? `${configIntel.proposals.length} changes` : 'not generated'} />
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

      <section className={`pipeline ${page !== 'overview' ? 'hidden' : ''}`}>
        <Step tone="yellow" icon={<Building2 />} label="YAML / Cloud Import" detail="existing infra files or Azure account data" />
        <Step tone="blue" icon={<Box />} label="Go API" detail="validates and publishes changes" />
        <Step tone="green" icon={<Database />} label="OpenSearch" detail="CMDB-style searchable state" />
        <Step tone="sky" icon={<Cloud />} label="Azure Path" detail="Event Hubs + Blob config" />
      </section>

      <section className={`actionGrid ${!['overview', 'security', 'config'].includes(page) ? 'hidden' : ''}`}>
        <div className="panel actionPanel">
          <div className="panelHeader">
            <h2>Executive Actions</h2>
            <span>{actionStatus}</span>
          </div>
          <div className="actionButtons">
            <button onClick={() => generateAnalytics()}>
              <BarChart3 size={17} />
              Create analytics
            </button>
            <button onClick={() => startAnalysis()}>
              <ShieldAlert size={17} />
              Start CVE analysis
            </button>
            <button onClick={() => buildConfig()}>
              <SlidersHorizontal size={17} />
              Build better config
            </button>
            <button onClick={() => createTicket()}>
              <PlayCircle size={17} />
              Create ServiceNow ticket
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
              <strong>{analysis?.model ?? 'weighted-logistic-cve-risk-v1'}</strong>
              <span>{analysis?.summary ?? 'Run analysis to score every indexed infrastructure asset from OpenSearch.'}</span>
            </div>
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
              {!analysis && <div className="empty">Click Start CVE analysis to score live assets from the OpenSearch read model.</div>}
            </div>
          </div>
        </details>

        <details className="panel insightPanel" open>
          <summary>
            <span>AI-assisted Configuration Management</span>
            <ChevronDown size={18} />
          </summary>
          <div className="insightBody">
            <div className="modelLine">
              <strong>{configIntel?.model ?? 'config-recommendation-rules-v1'}</strong>
              <span>{configIntel?.summary ?? 'Build better config from owners, services, CVE pressure, and Kafka replay behavior.'}</span>
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
              {!configIntel && <div className="empty">Click Build better config to generate a ServiceNow-aware preview config from indexed infrastructure.</div>}
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
                Import Azure subscription sample
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

      <details className={`disclosure ${!['overview', 'security'].includes(page) ? 'hidden' : ''}`} open>
        <summary>
          <span>Now Assist Handoff</span>
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

      <section className={`meshPanel ${!['overview', 'data'].includes(page) ? 'hidden' : ''}`}>
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

      <details className={`disclosure ${!['overview', 'data'].includes(page) ? 'hidden' : ''}`} open>
        <summary>
          <span>Launch, Import, and Data Contracts</span>
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
              <CodeLine method="4" path="POST /api/config/recommendations" note="builds better config from observed services, owners, and risk" />
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
              <CodeLine method="POST" path="/api/config/recommendations" note="AI-assisted config proposal" />
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

      <section className={`toolbar ${!['overview', 'data'].includes(page) ? 'hidden' : ''}`}>
        <Search size={18} />
        <input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          onKeyDown={(event) => event.key === 'Enter' && refresh()}
          placeholder="Search ServiceNow asset, office endpoint, owner, region, or risk"
        />
        <button onClick={() => refresh()}>Search</button>
      </section>

      <section className={`grid ${!['overview', 'data'].includes(page) ? 'hidden' : ''}`}>
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
              <div className="empty">Run make seed, then search for Search or catalog.</div>
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
    </main>
  );
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
