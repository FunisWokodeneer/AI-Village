package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOrganizationStorePersistsWorkspacesApprovalsAndActivity(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(t.TempDir(), "organization.json")
	store, err := OpenOrganizationStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RegisterWorkspace("project-a", root); err != nil {
		t.Fatal(err)
	}
	approvalID, err := store.RequestApproval(ApprovalRequest{
		AgentID: "agent-a", Action: ActionDeploy, Summary: "Deploy project",
		Reason: "Publish the reviewed result", AffectedResources: []string{"project-a"},
		Risks: []string{"Public users will see the release"}, RollbackPlan: "Disable the release",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.DecideApproval(approvalID, true); err != nil {
		t.Fatal(err)
	}

	reloaded, err := OpenOrganizationStore(path)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.Snapshot()
	if len(got.Workspaces) != 1 || got.Workspaces[0].ID != "project-a" {
		t.Fatalf("persisted workspaces = %#v", got.Workspaces)
	}
	if len(got.Approvals) != 1 || got.Approvals[0].Status != ApprovalApproved {
		t.Fatalf("persisted approvals = %#v", got.Approvals)
	}
	if len(got.Activity) != 3 {
		t.Fatalf("activity event count = %d, want 3", len(got.Activity))
	}
	if got.Activity[0].Outcome != "succeeded" || got.Activity[1].Outcome != "awaiting_approval" ||
		got.Activity[2].Outcome != "approved" {
		t.Fatalf("unexpected activity outcomes: %#v", got.Activity)
	}
}

func TestOrganizationStoreRejectsIncompleteApprovalAndRepeatDecision(t *testing.T) {
	store, err := OpenOrganizationStore(filepath.Join(t.TempDir(), "organization.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.RequestApproval(ApprovalRequest{Action: ActionDeploy}); err == nil {
		t.Fatal("incomplete approval request was accepted")
	}
	id, err := store.RequestApproval(ApprovalRequest{
		AgentID: "agent-a", Action: ActionDeploy, Summary: "Deploy",
		Reason: "Release approved build", AffectedResources: []string{"project-a"},
		Risks: []string{"Public exposure"}, RollbackPlan: "Revert deployment",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.DecideApproval(id, false); err != nil {
		t.Fatal(err)
	}
	if err := store.DecideApproval(id, true); err == nil {
		t.Fatal("already-decided request was changed")
	}
	if got := store.Snapshot().Approvals[0].Status; got != ApprovalRejected {
		t.Fatalf("approval status = %q, want %q", got, ApprovalRejected)
	}
}

func TestOrganizationStoreRejectsUnsupportedAndMalformedState(t *testing.T) {
	for name, contents := range map[string]string{
		"version":   `{"version":2}`,
		"malformed": `{"version":`,
		"trailing":  `{"version":1} {"version":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "organization.json")
			if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := OpenOrganizationStore(path); err == nil {
				t.Fatal("invalid organization state was accepted")
			}
		})
	}
}
