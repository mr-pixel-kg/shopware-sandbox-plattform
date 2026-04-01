package services

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	auditcontracts "github.com/manuel/shopware-testenv-platform/api/internal/auditlog"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type auditLogStoreStub struct {
	createFn          func(entry *models.AuditLog) error
	listFn            func(options repositories.AuditLogListOptions) ([]models.AuditLog, int64, error)
	listDistinctUsers func(options repositories.AuditLogFacetOptions) ([]models.User, error)
	lastCreate        *models.AuditLog
	lastList          repositories.AuditLogListOptions
	lastFacetList     repositories.AuditLogFacetOptions
}

func (s *auditLogStoreStub) Create(entry *models.AuditLog) error {
	s.lastCreate = entry
	if s.createFn != nil {
		return s.createFn(entry)
	}
	return nil
}

func (s *auditLogStoreStub) List(options repositories.AuditLogListOptions) ([]models.AuditLog, int64, error) {
	s.lastList = options
	if s.listFn != nil {
		return s.listFn(options)
	}
	return nil, 0, nil
}

func (s *auditLogStoreStub) ListDistinctUsers(options repositories.AuditLogFacetOptions) ([]models.User, error) {
	s.lastFacetList = options
	if s.listDistinctUsers != nil {
		return s.listDistinctUsers(options)
	}
	return nil, nil
}

func TestAuditServiceLogCreatesAuditEntry(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	clientToken := uuid.New()
	resourceID := uuid.New()
	ipAddress := "203.0.113.25"
	userAgent := "Shopware-Test/1.0"
	store := &auditLogStoreStub{}

	service := NewAuditService(store)
	resourceType := auditcontracts.ResourceTypeSandbox
	err := service.Log(AuditLogInput{
		Actor: AuditActor{
			UserID:      &userID,
			IPAddress:   &ipAddress,
			UserAgent:   &userAgent,
			ClientToken: &clientToken,
		},
		Action:       auditcontracts.ActionSandboxCreated,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Details: map[string]any{
			"imageId": resourceID.String(),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, store.lastCreate)
	assert.Equal(t, &userID, store.lastCreate.UserID)
	assert.Equal(t, string(auditcontracts.ActionSandboxCreated), store.lastCreate.Action)
	assert.Equal(t, &ipAddress, store.lastCreate.IPAddress)
	assert.Equal(t, &userAgent, store.lastCreate.UserAgent)
	assert.Equal(t, &clientToken, store.lastCreate.ClientToken)
	assert.Equal(t, "sandbox", *store.lastCreate.ResourceType)
	assert.Equal(t, &resourceID, store.lastCreate.ResourceID)
	assert.JSONEq(t, `{"imageId":"`+resourceID.String()+`"}`, string(store.lastCreate.Details))
	assert.WithinDuration(t, time.Now().UTC(), store.lastCreate.Timestamp, 2*time.Second)
}

func TestAuditServiceLogDefaultsDetailsToEmptyObject(t *testing.T) {
	t.Parallel()

	store := &auditLogStoreStub{}
	service := NewAuditService(store)

	err := service.Log(AuditLogInput{
		Action: auditcontracts.ActionAuthLoggedOut,
	})

	require.NoError(t, err)
	require.NotNil(t, store.lastCreate)
	assert.JSONEq(t, `{}`, string(store.lastCreate.Details))
	assert.Nil(t, store.lastCreate.UserID)
	assert.Nil(t, store.lastCreate.IPAddress)
	assert.Nil(t, store.lastCreate.UserAgent)
	assert.Nil(t, store.lastCreate.ClientToken)
	assert.Nil(t, store.lastCreate.ResourceType)
	assert.Nil(t, store.lastCreate.ResourceID)
}

func TestAuditServiceLogReturnsMarshalError(t *testing.T) {
	t.Parallel()

	service := NewAuditService(&auditLogStoreStub{})

	err := service.Log(AuditLogInput{
		Action: auditcontracts.ActionAuthLoggedIn,
		Details: map[string]any{
			"invalid": make(chan int),
		},
	})

	require.Error(t, err)
}

func TestAuditServiceListNormalizesFiltersAndPagination(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	resourceID := uuid.New()
	clientToken := uuid.New()
	from := time.Now().UTC().Add(-2 * time.Hour)
	to := time.Now().UTC()
	expectedLogs := []models.AuditLog{{ID: uuid.New()}, {ID: uuid.New()}}

	store := &auditLogStoreStub{
		listFn: func(options repositories.AuditLogListOptions) ([]models.AuditLog, int64, error) {
			return expectedLogs, 17, nil
		},
	}
	service := NewAuditService(store)

	action := "  sandbox.deleted  "
	resourceType := "  sandbox  "
	result, err := service.List(AuditLogListInput{
		Limit:        999,
		Offset:       -5,
		UserID:       &userID,
		Action:       &action,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		ClientToken:  &clientToken,
		From:         &from,
		To:           &to,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expectedLogs, result.Logs)
	assert.Equal(t, int64(17), result.Total)
	assert.Equal(t, 500, result.Limit)
	assert.Equal(t, 0, result.Offset)
	assert.Equal(t, &userID, store.lastList.UserID)
	require.NotNil(t, store.lastList.Action)
	assert.Equal(t, "sandbox.deleted", *store.lastList.Action)
	require.NotNil(t, store.lastList.ResourceType)
	assert.Equal(t, "sandbox", *store.lastList.ResourceType)
	assert.Equal(t, &resourceID, store.lastList.ResourceID)
	assert.Equal(t, &clientToken, store.lastList.ClientToken)
	assert.Equal(t, &from, store.lastList.From)
	assert.Equal(t, &to, store.lastList.To)
}

func TestAuditServiceListDropsBlankFiltersAndAppliesDefaults(t *testing.T) {
	t.Parallel()

	store := &auditLogStoreStub{
		listFn: func(options repositories.AuditLogListOptions) ([]models.AuditLog, int64, error) {
			return nil, 0, nil
		},
	}
	service := NewAuditService(store)

	action := "   "
	resourceType := "\t"
	result, err := service.List(AuditLogListInput{
		Action:       &action,
		ResourceType: &resourceType,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 50, result.Limit)
	assert.Equal(t, 0, result.Offset)
	assert.Nil(t, store.lastList.Action)
	assert.Nil(t, store.lastList.ResourceType)
}

func TestAuditServiceListReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("repo failed")
	store := &auditLogStoreStub{
		listFn: func(options repositories.AuditLogListOptions) ([]models.AuditLog, int64, error) {
			return nil, 0, expectedErr
		},
	}
	service := NewAuditService(store)

	result, err := service.List(AuditLogListInput{})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestAuditServiceListFacetsReturnsDistinctUsersAndKnownActions(t *testing.T) {
	t.Parallel()

	from := time.Now().UTC().Add(-24 * time.Hour)
	action := "  sandbox.deleted  "
	expectedUsers := []models.User{
		{ID: uuid.New(), Email: "a@example.com"},
		{ID: uuid.New(), Email: "b@example.com"},
	}
	store := &auditLogStoreStub{
		listDistinctUsers: func(options repositories.AuditLogFacetOptions) ([]models.User, error) {
			return expectedUsers, nil
		},
	}
	service := NewAuditService(store)

	result, err := service.ListFacets(AuditLogFacetInput{
		Action: &action,
		From:   &from,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expectedUsers, result.Users)
	require.NotNil(t, store.lastFacetList.Action)
	assert.Equal(t, "sandbox.deleted", *store.lastFacetList.Action)
	assert.Equal(t, &from, store.lastFacetList.From)
	assert.Contains(t, result.Actions, string(auditcontracts.ActionSandboxDeleted))
	assert.Contains(t, result.Actions, string(auditcontracts.ActionImageCreated))
}
