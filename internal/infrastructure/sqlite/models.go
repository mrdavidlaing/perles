package sqlite

import (
	"time"

	"github.com/zjrosen/perles/internal/sessions/domain"
)

// SessionModel represents the database row for the sessions table.
// Fields map directly to SQL columns with Unix timestamps for time values.
type SessionModel struct {
	ID              int64
	GUID            string
	Project         string
	Name            *string // nullable
	State           string
	TemplateID      *string // nullable
	EpicID          *string // nullable
	WorkDir         *string // nullable
	WorktreePath    *string // nullable
	WorktreeBranch  *string // nullable
	OwnerCreatedPID *int64  // nullable
	OwnerCurrentPID *int64  // nullable
	CreatedAt       int64   // Unix timestamp
	StartedAt       *int64  // Unix timestamp, nullable
	PausedAt        *int64  // Unix timestamp, nullable
	UpdatedAt       int64   // Unix timestamp
	ArchivedAt      *int64  // Unix timestamp, nullable
	DeletedAt       *int64  // Unix timestamp, nullable
}

// toSessionModel converts a domain Session entity to a database SessionModel.
func toSessionModel(s *domain.Session) *SessionModel {
	m := &SessionModel{
		ID:        s.ID(),
		GUID:      s.GUID(),
		Project:   s.Project(),
		State:     string(s.State()),
		CreatedAt: s.CreatedAt().Unix(),
		UpdatedAt: s.UpdatedAt().Unix(),
	}
	if s.Name() != "" {
		name := s.Name()
		m.Name = &name
	}
	if s.TemplateID() != "" {
		templateID := s.TemplateID()
		m.TemplateID = &templateID
	}
	if s.EpicID() != "" {
		epicID := s.EpicID()
		m.EpicID = &epicID
	}
	if s.WorkDir() != "" {
		workDir := s.WorkDir()
		m.WorkDir = &workDir
	}
	if s.WorktreePath() != "" {
		worktreePath := s.WorktreePath()
		m.WorktreePath = &worktreePath
	}
	if s.WorktreeBranch() != "" {
		worktreeBranch := s.WorktreeBranch()
		m.WorktreeBranch = &worktreeBranch
	}
	if s.OwnerCreatedPID() != nil {
		pid := int64(*s.OwnerCreatedPID())
		m.OwnerCreatedPID = &pid
	}
	if s.OwnerCurrentPID() != nil {
		pid := int64(*s.OwnerCurrentPID())
		m.OwnerCurrentPID = &pid
	}
	if s.StartedAt() != nil {
		startedAt := s.StartedAt().Unix()
		m.StartedAt = &startedAt
	}
	if s.PausedAt() != nil {
		pausedAt := s.PausedAt().Unix()
		m.PausedAt = &pausedAt
	}
	if s.ArchivedAt() != nil {
		archivedAt := s.ArchivedAt().Unix()
		m.ArchivedAt = &archivedAt
	}
	if s.DeletedAt() != nil {
		deletedAt := s.DeletedAt().Unix()
		m.DeletedAt = &deletedAt
	}
	return m
}

// toDomain converts a database SessionModel to a domain Session entity.
func (m *SessionModel) toDomain() *domain.Session {
	var name, templateID, epicID, workDir, worktreePath, worktreeBranch string
	if m.Name != nil {
		name = *m.Name
	}
	if m.TemplateID != nil {
		templateID = *m.TemplateID
	}
	if m.EpicID != nil {
		epicID = *m.EpicID
	}
	if m.WorkDir != nil {
		workDir = *m.WorkDir
	}
	if m.WorktreePath != nil {
		worktreePath = *m.WorktreePath
	}
	if m.WorktreeBranch != nil {
		worktreeBranch = *m.WorktreeBranch
	}
	var ownerCreatedPID *int
	if m.OwnerCreatedPID != nil {
		pid := int(*m.OwnerCreatedPID)
		ownerCreatedPID = &pid
	}
	var ownerCurrentPID *int
	if m.OwnerCurrentPID != nil {
		pid := int(*m.OwnerCurrentPID)
		ownerCurrentPID = &pid
	}
	var startedAt *time.Time
	if m.StartedAt != nil {
		t := time.Unix(*m.StartedAt, 0)
		startedAt = &t
	}
	var pausedAt *time.Time
	if m.PausedAt != nil {
		t := time.Unix(*m.PausedAt, 0)
		pausedAt = &t
	}
	var archivedAt *time.Time
	if m.ArchivedAt != nil {
		t := time.Unix(*m.ArchivedAt, 0)
		archivedAt = &t
	}
	var deletedAt *time.Time
	if m.DeletedAt != nil {
		t := time.Unix(*m.DeletedAt, 0)
		deletedAt = &t
	}
	return domain.ReconstituteSession(
		m.ID,
		m.GUID,
		m.Project,
		name,
		domain.SessionState(m.State),
		templateID,
		epicID,
		workDir,
		worktreePath,
		worktreeBranch,
		ownerCreatedPID,
		ownerCurrentPID,
		time.Unix(m.CreatedAt, 0),
		startedAt,
		pausedAt,
		time.Unix(m.UpdatedAt, 0),
		archivedAt,
		deletedAt,
	)
}
