package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxOrganizationFileSize = 4 << 20

type OrganizationStore struct {
	mu           sync.Mutex
	path         string
	organization Organization
}

func OpenOrganizationStore(path string) (*OrganizationStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("organization path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create organization directory: %w", err)
	}
	store := &OrganizationStore{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		store.organization = Organization{Version: 1}
		if err := store.saveLocked(); err != nil {
			return nil, err
		}
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read organization: %w", err)
	}
	if len(data) > maxOrganizationFileSize {
		return nil, errors.New("organization file exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&store.organization); err != nil {
		return nil, fmt.Errorf("decode organization: %w", err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, errors.New("organization file contains trailing data")
	}
	if store.organization.Version != 1 {
		return nil, fmt.Errorf("unsupported organization version %d", store.organization.Version)
	}
	return store, nil
}

func DefaultOrganizationPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user configuration directory: %w", err)
	}
	return filepath.Join(configDir, "AI-Village", "organization.json"), nil
}

func (s *OrganizationStore) Snapshot() Organization {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, _ := json.Marshal(s.organization)
	var snapshot Organization
	_ = json.Unmarshal(data, &snapshot)
	return snapshot
}

func (s *OrganizationStore) RegisterWorkspace(id, root string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("workspace ID is required")
	}
	if !filepath.IsAbs(root) {
		return errors.New("workspace path must be absolute")
	}
	root, err := filepath.EvalSymlinks(filepath.Clean(root))
	if err != nil {
		return fmt.Errorf("resolve workspace path: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("inspect workspace path: %w", err)
	}
	if !info.IsDir() {
		return errors.New("workspace path must be a directory")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.organization.Workspaces {
		if s.organization.Workspaces[i].ID == id {
			return errors.New("workspace ID is already registered")
		}
	}
	s.organization.Workspaces = append(s.organization.Workspaces, ProjectWorkspace{
		ID: id, Path: root, Approved: true,
	})
	if err := s.appendEventLocked(ActivityEvent{
		Action: "workspace.register", Summary: "Owner registered and approved a project workspace",
		Outcome: "succeeded", Timestamp: time.Now().UTC(),
		Metadata: map[string]string{"workspace_id": id},
	}); err != nil {
		s.organization.Workspaces = s.organization.Workspaces[:len(s.organization.Workspaces)-1]
		return err
	}
	return nil
}

func (s *OrganizationStore) RequestApproval(request ApprovalRequest) (string, error) {
	if request.AgentID == "" || request.Action == "" || request.Summary == "" ||
		request.Reason == "" || len(request.AffectedResources) == 0 ||
		len(request.Risks) == 0 || request.RollbackPlan == "" {
		return "", errors.New("approval requires an agent, action, summary, reason, affected resources, risks, and rollback plan")
	}
	id, err := newRecordID()
	if err != nil {
		return "", err
	}
	request.ID = id
	request.Status = ApprovalPending
	request.RequestedAt = time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.organization.Approvals = append(s.organization.Approvals, request)
	if err := s.appendEventLocked(ActivityEvent{
		AgentID: request.AgentID, TaskID: request.TaskID, Tool: "approval_queue",
		Action: string(request.Action), Summary: request.Summary, Outcome: "awaiting_approval",
		Timestamp: request.RequestedAt, Metadata: map[string]string{"approval_id": id},
	}); err != nil {
		s.organization.Approvals = s.organization.Approvals[:len(s.organization.Approvals)-1]
		return "", err
	}
	return id, nil
}

func (s *OrganizationStore) DecideApproval(id string, approved bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.organization.Approvals {
		request := &s.organization.Approvals[i]
		if request.ID != id {
			continue
		}
		if request.Status != ApprovalPending {
			return errors.New("approval request has already been decided")
		}
		now := time.Now().UTC()
		if approved {
			request.Status = ApprovalApproved
		} else {
			request.Status = ApprovalRejected
		}
		request.DecidedAt = &now
		outcome := string(request.Status)
		if err := s.appendEventLocked(ActivityEvent{
			AgentID: request.AgentID, TaskID: request.TaskID, Tool: "approval_queue",
			Action: string(request.Action), Summary: "Owner decided an approval request",
			Outcome: outcome, Timestamp: now, Metadata: map[string]string{"approval_id": id},
		}); err != nil {
			request.Status = ApprovalPending
			request.DecidedAt = nil
			return err
		}
		return nil
	}
	return errors.New("approval request not found")
}

func (s *OrganizationStore) RecordActivity(event ActivityEvent) error {
	if event.Action == "" || event.Outcome == "" {
		return errors.New("activity action and outcome are required")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.appendEventLocked(event)
}

func (s *OrganizationStore) appendEventLocked(event ActivityEvent) error {
	if event.ID == "" {
		id, err := newRecordID()
		if err != nil {
			return err
		}
		event.ID = id
	}
	s.organization.Activity = append(s.organization.Activity, event)
	if err := s.saveLocked(); err != nil {
		s.organization.Activity = s.organization.Activity[:len(s.organization.Activity)-1]
		return err
	}
	return nil
}

func (s *OrganizationStore) saveLocked() error {
	data, err := json.MarshalIndent(s.organization, "", "  ")
	if err != nil {
		return fmt.Errorf("encode organization: %w", err)
	}
	if len(data) > maxOrganizationFileSize {
		return errors.New("organization state exceeds size limit")
	}
	temp, err := os.CreateTemp(filepath.Dir(s.path), ".organization-*.tmp")
	if err != nil {
		return fmt.Errorf("create organization temp file: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0600); err != nil {
		temp.Close()
		return fmt.Errorf("secure organization temp file: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("write organization: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("flush organization: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close organization: %w", err)
	}
	if err := os.Rename(tempName, s.path); err != nil {
		return fmt.Errorf("replace organization: %w", err)
	}
	return nil
}

func newRecordID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate record ID: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}
