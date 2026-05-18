package domain

import "time"

type CVEFinding struct {
	CVEID       string   `json:"cve_id"`
	AssetID     string   `json:"asset_id"`
	AssetName   string   `json:"asset_name"`
	Severity    string   `json:"severity"`
	Score       float64  `json:"score"`
	Confidence  float64  `json:"confidence"`
	Signals     []string `json:"signals"`
	Remediation string   `json:"remediation"`
}

type CVEAnalysisReport struct {
	AnalysisID       string           `json:"analysis_id"`
	GeneratedAt      time.Time        `json:"generated_at"`
	Model            string           `json:"model"`
	Summary          string           `json:"summary"`
	Findings         []CVEFinding     `json:"findings"`
	IncidentBrief    *AIIncidentBrief `json:"incident_brief,omitempty"`
	KafkaTopic       string           `json:"kafka_topic"`
	SearchBackend    string           `json:"search_backend"`
	ServiceNowTarget string           `json:"servicenow_target"`
}

type AIIncidentBrief struct {
	Provider            string   `json:"provider"`
	Model               string   `json:"model"`
	Mode                string   `json:"mode"`
	ExecutiveSummary    string   `json:"executive_summary"`
	ProbableCause       string   `json:"probable_cause"`
	BlastRadius         string   `json:"blast_radius"`
	RecommendedActions  []string `json:"recommended_actions"`
	ServiceNowWorkNotes []string `json:"servicenow_work_notes"`
}

type ServiceNowTicket struct {
	Number             string    `json:"number"`
	CreatedAt          time.Time `json:"created_at"`
	Table              string    `json:"table"`
	ShortDescription   string    `json:"short_description"`
	AssignmentGroup    string    `json:"assignment_group"`
	Impact             string    `json:"impact"`
	Urgency            string    `json:"urgency"`
	Priority           string    `json:"priority"`
	ConfigurationItems []string  `json:"configuration_items"`
	WorkNotes          []string  `json:"work_notes"`
}
