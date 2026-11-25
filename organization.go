package main

import "time"

// Organization is the persisted domain model for the AI Village workforce.
// It deliberately contains no executable capabilities; tools must enforce
// permissions through EvaluatePolicy before performing any side effect.
type Organization struct {
	Version     int               `json:"version"`
	Departments []Department      `json:"departments"`
	Agents      []AgentProfile    `json:"agents"`
	Projects    []Project         `json:"projects"`
	Tasks       []WorkTask        `json:"tasks"`
	Approvals   []ApprovalRequest `json:"approvals"`
	Activity    []ActivityEvent   `json:"activity"`
}

type Department struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ManagerAgentID string    `json:"manager_agent_id,omitempty"`
	BuildingID     string    `json:"building_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type AgentProfile struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Role           string   `json:"role"`
	DepartmentID   string   `json:"department_id,omitempty"`
	ManagerAgentID string   `json:"manager_agent_id,omitempty"`
	Instructions   string   `json:"instructions,omitempty"`
	Capabilities   []string `json:"capabilities,omitempty"`
	Active         bool     `json:"active"`
}

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	WorkspaceID string    `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type WorkTaskStatus string

const (
	TaskProposed   WorkTaskStatus = "proposed"
	TaskQueued     WorkTaskStatus = "queued"
	TaskRunning    WorkTaskStatus = "running"
	TaskBlocked    WorkTaskStatus = "blocked"
	TaskNeedsReview WorkTaskStatus = "needs_review"
	TaskCompleted  WorkTaskStatus = "completed"
	TaskFailed     WorkTaskStatus = "failed"
	TaskCancelled  WorkTaskStatus = "cancelled"
)

type WorkTask struct {
	ID              string         `json:"id"`
	ParentTaskID    string         `json:"parent_task_id,omitempty"`
	ProjectID       string         `json:"project_id,omitempty"`
	WorkspaceID     string         `json:"workspace_id,omitempty"`
	DepartmentID    string         `json:"department_id,omitempty"`
	AssignedAgentID string         `json:"assigned_agent_id,omitempty"`
	Title           string         `json:"title"`
	Description     string         `json:"description,omitempty"`
	Status          WorkTaskStatus `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Result          string         `json:"result,omitempty"`
}

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

type ApprovalRequest struct {
	ID                string         `json:"id"`
	TaskID            string         `json:"task_id,omitempty"`
	AgentID           string         `json:"agent_id"`
	Action            PolicyAction   `json:"action"`
	Summary           string         `json:"summary"`
	Reason            string         `json:"reason"`
	AffectedResources []string       `json:"affected_resources,omitempty"`
	Risks             []string       `json:"risks,omitempty"`
	RollbackPlan      string         `json:"rollback_plan,omitempty"`
	Status            ApprovalStatus `json:"status"`
	RequestedAt       time.Time      `json:"requested_at"`
	DecidedAt         *time.Time     `json:"decided_at,omitempty"`
}

type ActivityEvent struct {
	ID        string            `json:"id"`
	AgentID   string            `json:"agent_id,omitempty"`
	TaskID    string            `json:"task_id,omitempty"`
	Tool      string            `json:"tool,omitempty"`
	Action    string            `json:"action"`
	Summary   string            `json:"summary"`
	Outcome   string            `json:"outcome"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// PolicyAction identifies a requested capability, not a shell command.
type PolicyAction string

const (
	ActionReadWorkspace       PolicyAction = "workspace.read"
	ActionWriteWorkspace      PolicyAction = "workspace.write"
	ActionCreateProjectFile   PolicyAction = "workspace.create_file"
	ActionCreateProjectFolder PolicyAction = "workspace.create_folder"
	ActionRunTests            PolicyAction = "workspace.run_tests"
	ActionRunBuild            PolicyAction = "workspace.run_build"
	ActionInstallDependency   PolicyAction = "system.install_dependency"
	ActionPublishRelease      PolicyAction = "external.publish_release"
	ActionDeploy              PolicyAction = "external.deploy"
	ActionExternalMessage     PolicyAction = "external.send_message"
	ActionAccessOutside       PolicyAction = "workspace.access_outside"
	ActionDestructiveChange   PolicyAction = "workspace.destructive_change"
	ActionChangePolicy        PolicyAction = "security.change_policy"
	ActionPaidService         PolicyAction = "billing.add_paid_service"
	ActionModifyApplication   PolicyAction = "application.modify"
)

type PolicyDecision string

const (
	PolicyAllow          PolicyDecision = "allow"
	PolicyNeedsApproval  PolicyDecision = "needs_approval"
	PolicyDeny           PolicyDecision = "deny"
)

type PolicyContext struct {
	WorkspaceID          string
	TargetWorkspaceID    string
	WorkspaceApproved    bool
	CommandAllowlisted   bool
	ExplicitlyApproved   bool
}

type PolicyResult struct {
	Decision PolicyDecision
	Reason   string
}

// EvaluatePolicy is intentionally fail-closed. Routine project work is
// automatic only inside an explicitly approved workspace and only when the
// command is on the application's allowlist. Consequential actions wait for
// an explicit owner decision; unknown actions are denied.
func EvaluatePolicy(action PolicyAction, ctx PolicyContext) PolicyResult {
	if ctx.ExplicitlyApproved {
		switch action {
		case ActionReadWorkspace, ActionWriteWorkspace, ActionCreateProjectFile,
			ActionCreateProjectFolder, ActionRunTests, ActionRunBuild,
			ActionInstallDependency, ActionPublishRelease, ActionDeploy,
			ActionExternalMessage, ActionAccessOutside, ActionDestructiveChange,
			ActionChangePolicy, ActionPaidService, ActionModifyApplication:
			return PolicyResult{Decision: PolicyAllow, Reason: "Explicit owner approval recorded"}
		}
	}

	switch action {
	case ActionReadWorkspace, ActionWriteWorkspace, ActionCreateProjectFile, ActionCreateProjectFolder:
		if !workspaceIsScoped(ctx) {
			return PolicyResult{Decision: PolicyNeedsApproval, Reason: "The target is not inside an approved workspace"}
		}
		return PolicyResult{Decision: PolicyAllow, Reason: "Action is scoped to an approved workspace"}

	case ActionRunTests, ActionRunBuild:
		if !workspaceIsScoped(ctx) {
			return PolicyResult{Decision: PolicyNeedsApproval, Reason: "The target is not inside an approved workspace"}
		}
		if !ctx.CommandAllowlisted {
			return PolicyResult{Decision: PolicyNeedsApproval, Reason: "The requested command is not on the approved command allowlist"}
		}
		return PolicyResult{Decision: PolicyAllow, Reason: "Allowlisted build or test in an approved workspace"}

	case ActionInstallDependency, ActionPublishRelease, ActionDeploy,
		ActionExternalMessage, ActionAccessOutside, ActionDestructiveChange,
		ActionChangePolicy, ActionPaidService, ActionModifyApplication:
		return PolicyResult{Decision: PolicyNeedsApproval, Reason: "This action requires explicit owner approval"}

	default:
		return PolicyResult{Decision: PolicyDeny, Reason: "Unknown or unsupported action"}
	}
}

func workspaceIsScoped(ctx PolicyContext) bool {
	return ctx.WorkspaceApproved &&
		ctx.WorkspaceID != "" &&
		ctx.TargetWorkspaceID != "" &&
		ctx.WorkspaceID == ctx.TargetWorkspaceID
}
