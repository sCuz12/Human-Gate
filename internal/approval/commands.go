package approval

import "encoding/json"

type SubmitApprovalRequestCommand struct {
	WorkspaceID     string
	IdempotencyKey  string
	Action          ProposedAction
	Context         RequestContext
	Source          RequestSource
	Continuation    Continuation
	AuthenticatedBy AuthenticatedActor
}

type ProposedAction struct {
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type RequestContext struct {
	Reason          string            `json:"reason"`
	Evidence        json.RawMessage   `json:"evidence"`
	Confidence      *float64          `json:"confidence"`
	Reversible      *bool             `json:"reversible"`
	AffectedSystems []string          `json:"affected_systems"`
	Deadline        string            `json:"deadline"`
	Metadata        map[string]string `json:"metadata"`
}

type RequestSource struct {
	Platform    string `json:"platform"`
	WorkflowID  string `json:"workflow_id"`
	ExecutionID string `json:"execution_id"`
}

type Continuation struct {
	Strategy string `json:"strategy"`
	URL      string `json:"url"`
}

type AuthenticatedActor struct {
	ActorType string
	ActorID   string
}
