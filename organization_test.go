package main

import "testing"

func TestEvaluatePolicyAllowsRoutineWorkInsideApprovedWorkspace(t *testing.T) {
	ctx := PolicyContext{
		WorkspaceID:        "project-a",
		TargetWorkspaceID:  "project-a",
		WorkspaceApproved:  true,
		CommandAllowlisted: true,
	}

	for _, action := range []PolicyAction{
		ActionReadWorkspace,
		ActionWriteWorkspace,
		ActionCreateProjectFile,
		ActionCreateProjectFolder,
		ActionRunTests,
		ActionRunBuild,
	} {
		t.Run(string(action), func(t *testing.T) {
			got := EvaluatePolicy(action, ctx)
			if got.Decision != PolicyAllow {
				t.Fatalf("decision = %q, want %q (reason: %s)", got.Decision, PolicyAllow, got.Reason)
			}
		})
	}
}

func TestEvaluatePolicyRequiresApprovalOutsideApprovedWorkspace(t *testing.T) {
	ctx := PolicyContext{
		WorkspaceID:       "project-a",
		TargetWorkspaceID: "project-b",
		WorkspaceApproved: true,
	}

	for _, action := range []PolicyAction{
		ActionReadWorkspace,
		ActionWriteWorkspace,
		ActionCreateProjectFile,
		ActionCreateProjectFolder,
		ActionRunTests,
		ActionRunBuild,
	} {
		t.Run(string(action), func(t *testing.T) {
			got := EvaluatePolicy(action, ctx)
			if got.Decision != PolicyNeedsApproval {
				t.Fatalf("decision = %q, want %q", got.Decision, PolicyNeedsApproval)
			}
		})
	}
}

func TestEvaluatePolicyRequiresAllowlistedCommands(t *testing.T) {
	ctx := PolicyContext{
		WorkspaceID:       "project-a",
		TargetWorkspaceID: "project-a",
		WorkspaceApproved: true,
	}

	for _, action := range []PolicyAction{ActionRunTests, ActionRunBuild} {
		got := EvaluatePolicy(action, ctx)
		if got.Decision != PolicyNeedsApproval {
			t.Errorf("%s decision = %q, want %q", action, got.Decision, PolicyNeedsApproval)
		}
	}
}

func TestEvaluatePolicyRequiresApprovalForConsequentialActions(t *testing.T) {
	ctx := PolicyContext{}

	for _, action := range []PolicyAction{
		ActionInstallDependency,
		ActionPublishRelease,
		ActionDeploy,
		ActionExternalMessage,
		ActionAccessOutside,
		ActionDestructiveChange,
		ActionChangePolicy,
		ActionPaidService,
		ActionModifyApplication,
	} {
		t.Run(string(action), func(t *testing.T) {
			got := EvaluatePolicy(action, ctx)
			if got.Decision != PolicyNeedsApproval {
				t.Fatalf("decision = %q, want %q", got.Decision, PolicyNeedsApproval)
			}
		})
	}
}

func TestEvaluatePolicyHonorsExplicitOwnerApproval(t *testing.T) {
	got := EvaluatePolicy(ActionPublishRelease, PolicyContext{ExplicitlyApproved: true})
	if got.Decision != PolicyAllow {
		t.Fatalf("decision = %q, want %q", got.Decision, PolicyAllow)
	}
}

func TestEvaluatePolicyDeniesUnknownActions(t *testing.T) {
	got := EvaluatePolicy(PolicyAction("unknown.action"), PolicyContext{ExplicitlyApproved: true})
	if got.Decision != PolicyDeny {
		t.Fatalf("decision = %q, want %q", got.Decision, PolicyDeny)
	}
}

func TestEvaluatePolicyRequiresApprovalForUnscopedOrConsequentialActions(t *testing.T) {
	for _, action := range []PolicyAction{
		ActionReadWorkspace, ActionWriteWorkspace, ActionCreateProjectFile,
		ActionCreateProjectFolder, ActionRunTests, ActionRunBuild,
		ActionInstallDependency, ActionPublishRelease, ActionDeploy,
		ActionExternalMessage, ActionAccessOutside, ActionDestructiveChange,
		ActionChangePolicy, ActionPaidService, ActionModifyApplication,
	} {
		t.Run(string(action), func(t *testing.T) {
			if got := EvaluatePolicy(action, PolicyContext{}); got.Decision != PolicyNeedsApproval {
				t.Fatalf("decision = %q, want %q", got.Decision, PolicyNeedsApproval)
			}
		})
	}
}
