package mocks

import (
	"context"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/stretchr/testify/mock"
)

// MockQuerier is a mock implementation of the db.Querier interface
type MockQuerier struct {
	mock.Mock
}

// Vendor data methods
func (m *MockQuerier) CreateVendorData(ctx context.Context, arg db.CreateVendorDataParams) (db.VendorDatum, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.VendorDatum), args.Error(1)
}

func (m *MockQuerier) DeleteVendorData(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) GetVendorData(ctx context.Context, name string) (db.GetVendorDataRow, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(db.GetVendorDataRow), args.Error(1)
}

func (m *MockQuerier) UpdateVendorData(ctx context.Context, arg db.UpdateVendorDataParams) (db.VendorDatum, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.VendorDatum), args.Error(1)
}

// Instance methods
func (m *MockQuerier) CreateInstance(ctx context.Context, arg db.CreateInstanceParams) (db.Instance, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Instance), args.Error(1)
}

func (m *MockQuerier) GetInstance(ctx context.Context, arg db.GetInstanceParams) (db.Instance, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Instance), args.Error(1)
}

func (m *MockQuerier) GetInstanceByID(ctx context.Context, id int64) (db.Instance, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Instance), args.Error(1)
}

func (m *MockQuerier) GetInstanceByIP(ctx context.Context, ipAddress *string) (db.Instance, error) {
	args := m.Called(ctx, ipAddress)
	return args.Get(0).(db.Instance), args.Error(1)
}

func (m *MockQuerier) ListInstances(ctx context.Context) ([]db.Instance, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Instance), args.Error(1)
}

func (m *MockQuerier) ListInstancesByProject(ctx context.Context, project string) ([]db.Instance, error) {
	args := m.Called(ctx, project)
	return args.Get(0).([]db.Instance), args.Error(1)
}

func (m *MockQuerier) UpdateInstance(ctx context.Context, arg db.UpdateInstanceParams) (db.Instance, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Instance), args.Error(1)
}

func (m *MockQuerier) UpdateInstanceIP(ctx context.Context, arg db.UpdateInstanceIPParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) DeleteInstance(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) HardDeleteInstance(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Instance state methods
func (m *MockQuerier) CreateOrUpdateInstanceState(ctx context.Context, arg db.CreateOrUpdateInstanceStateParams) (db.InstanceState, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.InstanceState), args.Error(1)
}

func (m *MockQuerier) GetInstanceState(ctx context.Context, instanceID int64) (db.InstanceState, error) {
	args := m.Called(ctx, instanceID)
	return args.Get(0).(db.InstanceState), args.Error(1)
}

func (m *MockQuerier) DeleteInstanceState(ctx context.Context, instanceID int64) error {
	args := m.Called(ctx, instanceID)
	return args.Error(0)
}

// Instance logs methods
func (m *MockQuerier) CreateInstanceLog(ctx context.Context, arg db.CreateInstanceLogParams) (db.InstanceLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.InstanceLog), args.Error(1)
}

func (m *MockQuerier) GetInstanceLogs(ctx context.Context, arg db.GetInstanceLogsParams) ([]db.InstanceLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.InstanceLog), args.Error(1)
}

func (m *MockQuerier) GetInstanceLogsByType(ctx context.Context, arg db.GetInstanceLogsByTypeParams) ([]db.InstanceLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.InstanceLog), args.Error(1)
}

func (m *MockQuerier) GetInstanceLogsByLevel(ctx context.Context, arg db.GetInstanceLogsByLevelParams) ([]db.InstanceLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.InstanceLog), args.Error(1)
}

func (m *MockQuerier) DeleteInstanceLogs(ctx context.Context, instanceID int64) error {
	args := m.Called(ctx, instanceID)
	return args.Error(0)
}

func (m *MockQuerier) DeleteOldInstanceLogs(ctx context.Context, arg db.DeleteOldInstanceLogsParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

// Profile methods
func (m *MockQuerier) CreateProfile(ctx context.Context, arg db.CreateProfileParams) (db.Profile, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Profile), args.Error(1)
}

func (m *MockQuerier) GetProfile(ctx context.Context, arg db.GetProfileParams) (db.Profile, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Profile), args.Error(1)
}

func (m *MockQuerier) ListProfiles(ctx context.Context) ([]db.Profile, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Profile), args.Error(1)
}

func (m *MockQuerier) ListProfilesByProject(ctx context.Context, project string) ([]db.Profile, error) {
	args := m.Called(ctx, project)
	return args.Get(0).([]db.Profile), args.Error(1)
}

func (m *MockQuerier) UpdateProfile(ctx context.Context, id int64) (db.Profile, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Profile), args.Error(1)
}

func (m *MockQuerier) DeleteProfile(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
