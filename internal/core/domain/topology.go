package domain

type ProductStory struct {
	Mission           string `json:"mission"`
	Vision            string `json:"vision"`
	ServiceNowUseCase string `json:"servicenow_use_case"`
}

type TopologyNode struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Kind   string `json:"kind"`
	Owner  string `json:"owner"`
	Risk   string `json:"risk"`
	Detail string `json:"detail"`
	Tone   string `json:"tone"`
}

type TopologyEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

type TeamNeed struct {
	Team     string `json:"team"`
	Need     string `json:"need"`
	ServedBy string `json:"served_by"`
}

type TopologyMetrics struct {
	Assets              int    `json:"assets"`
	HighRisk            int    `json:"high_risk"`
	RecentEvents        int    `json:"recent_events"`
	ConfigVersion       string `json:"config_version"`
	QuerySLO            string `json:"query_slo"`
	IndexingLagSLO      string `json:"indexing_lag_slo"`
	ReplayWindow        int    `json:"replay_window"`
	DependencyGuard     string `json:"dependency_guard"`
	ConfigReloadCadence string `json:"config_reload_cadence"`
}

type SupportTopology struct {
	Story   ProductStory    `json:"story"`
	Metrics TopologyMetrics `json:"metrics"`
	Nodes   []TopologyNode  `json:"nodes"`
	Edges   []TopologyEdge  `json:"edges"`
	Teams   []TeamNeed      `json:"teams"`
}
