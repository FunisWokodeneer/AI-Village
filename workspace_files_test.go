package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceFilesReadWriteAndCreateDirectory(t *testing.T) {
	root := t.TempDir()
	store := newTestOrganizationStore(t)
	if err := store.RegisterWorkspace("project-a", root); err != nil {
		t.Fatal(err)
	}
	files, err := NewWorkspaceFiles(store, "project-a", "agent-a", "task-a")
	if err != nil {
		t.Fatal(err)
	}
	if err := files.CreateDirectory("src/pkg"); err != nil {
		t.Fatal(err)
	}
	if err := files.WriteFile("src/pkg/main.go", []byte("package pkg\n")); err != nil {
		t.Fatal(err)
	}
	got, err := files.ReadFile("src/pkg/main.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "package pkg\n" {
		t.Fatalf("file contents = %q", got)
	}
	events := store.Snapshot().Activity
	if len(events) != 7 {
		t.Fatalf("activity event count = %d, want 7", len(events))
	}
	for _, event := range events[1:] {
		if event.AgentID != "agent-a" || event.TaskID != "task-a" || event.Tool != "workspace_files" {
			t.Errorf("activity event lacks agent/task/tool attribution: %#v", event)
		}
	}
}

func TestWorkspaceFilesRejectUnregisteredAndEscapingPaths(t *testing.T) {
	root := t.TempDir()
	store := newTestOrganizationStore(t)
	if err := store.RegisterWorkspace("project-a", root); err != nil {
		t.Fatal(err)
	}
	files, err := NewWorkspaceFiles(store, "project-a", "agent-a", "task-a")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"../outside.txt", filepath.Join("..", "outside.txt"), ""} {
		if err := files.WriteFile(name, []byte("no")); err == nil {
			t.Errorf("WriteFile(%q) unexpectedly succeeded", name)
		}
	}
	if _, err := NewWorkspaceFiles(store, "not-registered", "agent-a", "task-a"); err != nil {
		t.Fatal(err)
	} else if _, err := filesForUnregistered(store); err == nil {
		t.Fatal("unregistered workspace operation unexpectedly succeeded")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "outside.txt")); !os.IsNotExist(err) {
		t.Fatalf("outside path was created, stat error = %v", err)
	}
}

func TestWorkspaceFilesRejectSymlinkEscape(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "external")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	store := newTestOrganizationStore(t)
	if err := store.RegisterWorkspace("project-a", root); err != nil {
		t.Fatal(err)
	}
	files, err := NewWorkspaceFiles(store, "project-a", "agent-a", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := files.ReadFile("external/secret.txt"); err == nil {
		t.Fatal("read through symlink outside workspace succeeded")
	}
	if err := files.WriteFile("external/new.txt", []byte("no")); err == nil {
		t.Fatal("write through symlink outside workspace succeeded")
	}
	if len(store.Snapshot().Activity) < 3 {
		t.Fatal("denied attempts were not recorded")
	}
}

func TestWorkspaceFilesRejectsReplacedWorkspaceRoot(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	store := newTestOrganizationStore(t)
	if err := store.RegisterWorkspace("project-a", root); err != nil {
		t.Fatal(err)
	}
	originalRoot := root + "-original"
	if err := os.Rename(root, originalRoot); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, root); err != nil {
		t.Skipf("directory symlinks are unavailable: %v", err)
	}
	files, err := NewWorkspaceFiles(store, "project-a", "agent-a", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := files.ReadFile("secret.txt"); err == nil {
		t.Fatal("read through a replacement workspace-root symlink succeeded")
	}
}

func TestWorkspaceFilesNeverRecordFileContents(t *testing.T) {
	root := t.TempDir()
	store := newTestOrganizationStore(t)
	if err := store.RegisterWorkspace("project-a", root); err != nil {
		t.Fatal(err)
	}
	files, err := NewWorkspaceFiles(store, "project-a", "agent-a", "")
	if err != nil {
		t.Fatal(err)
	}
	secret := "do-not-put-this-in-the-audit-log"
	if err := files.WriteFile("notes.txt", []byte(secret)); err != nil {
		t.Fatal(err)
	}
	for _, event := range store.Snapshot().Activity {
		if strings.Contains(event.Summary, secret) {
			t.Fatal("file contents leaked into activity history")
		}
		for _, value := range event.Metadata {
			if strings.Contains(value, secret) {
				t.Fatal("file contents leaked into activity history")
			}
		}
	}
}

func newTestOrganizationStore(t *testing.T) *OrganizationStore {
	t.Helper()
	store, err := OpenOrganizationStore(filepath.Join(t.TempDir(), "organization.json"))
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func filesForUnregistered(store *OrganizationStore) ([]byte, error) {
	files, err := NewWorkspaceFiles(store, "unknown", "agent-a", "")
	if err != nil {
		return nil, err
	}
	return files.ReadFile("any.txt")
}
