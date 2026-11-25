package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const maxWorkspaceFileSize = 1 << 20

type WorkspaceFiles struct {
	store       *OrganizationStore
	workspaceID string
	agentID     string
	taskID      string
}

func NewWorkspaceFiles(store *OrganizationStore, workspaceID, agentID, taskID string) (*WorkspaceFiles, error) {
	if store == nil {
		return nil, errors.New("organization store is required")
	}
	if strings.TrimSpace(workspaceID) == "" {
		return nil, errors.New("workspace ID is required")
	}
	return &WorkspaceFiles{store: store, workspaceID: workspaceID, agentID: agentID, taskID: taskID}, nil
}

func (f *WorkspaceFiles) ReadFile(name string) ([]byte, error) {
	workspace, err := f.authorize(ActionReadWorkspace, name)
	if err != nil {
		return nil, err
	}
	if err := f.record(string(ActionReadWorkspace), name, "started"); err != nil {
		return nil, err
	}
	path, err := resolveWorkspacePath(workspace.Path, name)
	if err != nil {
		return nil, f.finish(string(ActionReadWorkspace), name, "failed", err)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, f.finish(string(ActionReadWorkspace), name, "failed", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, f.finish(string(ActionReadWorkspace), name, "failed", errors.New("workspace path is not a regular file"))
	}
	if info.Size() > maxWorkspaceFileSize {
		return nil, f.finish(string(ActionReadWorkspace), name, "failed", errors.New("workspace file exceeds size limit"))
	}
	data, err := io.ReadAll(io.LimitReader(file, maxWorkspaceFileSize+1))
	if err != nil {
		return nil, f.finish(string(ActionReadWorkspace), name, "failed", err)
	}
	if len(data) > maxWorkspaceFileSize {
		return nil, f.finish(string(ActionReadWorkspace), name, "failed", errors.New("workspace file exceeds size limit"))
	}
	if err := f.finish(string(ActionReadWorkspace), name, "succeeded", nil); err != nil {
		return nil, err
	}
	return data, nil
}

func (f *WorkspaceFiles) WriteFile(name string, data []byte) error {
	workspace, err := f.authorize(ActionWriteWorkspace, name)
	if err != nil {
		return err
	}
	if err := f.record(string(ActionWriteWorkspace), name, "started"); err != nil {
		return err
	}
	if len(data) > maxWorkspaceFileSize {
		return f.finish(string(ActionWriteWorkspace), name, "failed", errors.New("workspace file exceeds size limit"))
	}
	path, err := resolveWorkspacePath(workspace.Path, name)
	if err != nil {
		return f.finish(string(ActionWriteWorkspace), name, "failed", err)
	}
	info, err := os.Lstat(path)
	if err == nil && !info.Mode().IsRegular() {
		return f.finish(string(ActionWriteWorkspace), name, "failed", errors.New("workspace target is not a regular file"))
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return f.finish(string(ActionWriteWorkspace), name, "failed", fmt.Errorf("inspect workspace target: %w", err))
	}
	parentInfo, err := os.Stat(filepath.Dir(path))
	if err != nil || !parentInfo.IsDir() {
		return f.finish(string(ActionWriteWorkspace), name, "failed", errors.New("workspace parent directory does not exist"))
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".ai-village-*.tmp")
	if err != nil {
		return f.finish(string(ActionWriteWorkspace), name, "failed", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0600); err != nil {
		temp.Close()
		return f.finish(string(ActionWriteWorkspace), name, "failed", err)
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return f.finish(string(ActionWriteWorkspace), name, "failed", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return f.finish(string(ActionWriteWorkspace), name, "failed", err)
	}
	if err := temp.Close(); err != nil {
		return f.finish(string(ActionWriteWorkspace), name, "failed", err)
	}
	if err := os.Rename(tempName, path); err != nil {
		return f.finish(string(ActionWriteWorkspace), name, "failed", err)
	}
	return f.finish(string(ActionWriteWorkspace), name, "succeeded", nil)
}

func (f *WorkspaceFiles) CreateDirectory(name string) error {
	workspace, err := f.authorize(ActionCreateProjectFolder, name)
	if err != nil {
		return err
	}
	if err := f.record(string(ActionCreateProjectFolder), name, "started"); err != nil {
		return err
	}
	path, err := resolveWorkspacePath(workspace.Path, name)
	if err != nil {
		return f.finish(string(ActionCreateProjectFolder), name, "failed", err)
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		return f.finish(string(ActionCreateProjectFolder), name, "failed", err)
	}
	return f.finish(string(ActionCreateProjectFolder), name, "succeeded", nil)
}

func (f *WorkspaceFiles) authorize(action PolicyAction, name string) (ProjectWorkspace, error) {
	var workspace ProjectWorkspace
	for _, registered := range f.store.Snapshot().Workspaces {
		if registered.ID == f.workspaceID {
			workspace = registered
			break
		}
	}
	if workspace.ID == "" {
		_ = f.record(string(action), name, "denied")
		return ProjectWorkspace{}, errors.New("workspace is not registered")
	}
	result := EvaluatePolicy(action, PolicyContext{
		WorkspaceID: workspace.ID, TargetWorkspaceID: f.workspaceID,
		WorkspaceApproved: workspace.Approved,
	})
	if result.Decision != PolicyAllow {
		_ = f.record(string(action), name, "denied")
		return ProjectWorkspace{}, fmt.Errorf("workspace policy denied action: %s", result.Reason)
	}
	if _, err := resolveWorkspacePath(workspace.Path, name); err != nil {
		_ = f.record(string(action), name, "denied")
		return ProjectWorkspace{}, err
	}
	return workspace, nil
}

func (f *WorkspaceFiles) record(action, name, outcome string) error {
	return f.store.RecordActivity(ActivityEvent{
		AgentID: f.agentID, TaskID: f.taskID, Tool: "workspace_files",
		Action: action, Summary: "Project workspace file operation",
		Outcome: outcome, Timestamp: time.Now().UTC(),
		Metadata: map[string]string{"workspace_id": f.workspaceID, "path": filepath.ToSlash(name)},
	})
}

func (f *WorkspaceFiles) finish(action, name, outcome string, cause error) error {
	if err := f.record(action, name, outcome); err != nil {
		return fmt.Errorf("record workspace activity: %w", err)
	}
	if cause != nil {
		return cause
	}
	return nil
}

func resolveWorkspacePath(root, name string) (string, error) {
	if name == "" || strings.ContainsRune(name, '\x00') || filepath.IsAbs(name) {
		return "", errors.New("workspace path must be a non-empty relative path")
	}
	registeredRoot := filepath.Clean(root)
	resolvedRoot, err := filepath.EvalSymlinks(registeredRoot)
	if err != nil {
		return "", errors.New("registered workspace is unavailable")
	}
	if !sameWorkspacePath(registeredRoot, filepath.Clean(resolvedRoot)) {
		return "", errors.New("registered workspace root changed")
	}
	root = registeredRoot
	cleanName := filepath.Clean(filepath.FromSlash(name))
	if cleanName == "." || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", errors.New("workspace path escapes its root")
	}
	target := filepath.Join(root, cleanName)
	if !pathWithin(root, target) {
		return "", errors.New("workspace path escapes its root")
	}

	current := root
	for _, component := range strings.Split(cleanName, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		resolved, resolveErr := filepath.EvalSymlinks(current)
		if resolveErr != nil {
			if errors.Is(resolveErr, os.ErrNotExist) {
				continue
			}
			return "", errors.New("workspace path contains an invalid link")
		}
		if !pathWithin(root, resolved) {
			return "", errors.New("workspace path escapes its root")
		}
	}
	return target, nil
}

func pathWithin(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	if err != nil || filepath.IsAbs(relative) || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	if runtime.GOOS == "windows" {
		return !strings.HasPrefix(strings.ToLower(relative), "..\\")
	}
	return true
}

func sameWorkspacePath(left, right string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
