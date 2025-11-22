package mocks

import (
	"context"
	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"

	"github.com/stretchr/testify/mock"
)

// UserRepositoryMock mocks repository.UserRepository
type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepositoryMock) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*entity.User, error) {
	args := m.Called(ctx, usernameOrEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *UserRepositoryMock) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepositoryMock) UpdatePasswordHash(ctx context.Context, userID int64, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}

func (m *UserRepositoryMock) UpdateLastLogin(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *UserRepositoryMock) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *UserRepositoryMock) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *UserRepositoryMock) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *UserRepositoryMock) List(ctx context.Context, filter repository.UserFilter) ([]*entity.User, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *UserRepositoryMock) Count(ctx context.Context, filter repository.UserFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// SessionRepositoryMock mocks repository.SessionRepository
type SessionRepositoryMock struct {
	mock.Mock
}

func (m *SessionRepositoryMock) Create(ctx context.Context, session *entity.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *SessionRepositoryMock) GetByToken(ctx context.Context, token string) (*entity.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Session), args.Error(1)
}

func (m *SessionRepositoryMock) GetByID(ctx context.Context, id int64) (*entity.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Session), args.Error(1)
}

func (m *SessionRepositoryMock) GetActiveByUserID(ctx context.Context, userID int64) ([]*entity.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Session), args.Error(1)
}

func (m *SessionRepositoryMock) Update(ctx context.Context, session *entity.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *SessionRepositoryMock) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *SessionRepositoryMock) DeleteByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *SessionRepositoryMock) DeleteAllByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *SessionRepositoryMock) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *SessionRepositoryMock) CleanupOldSessions(ctx context.Context, olderThan int) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

// RoleRepositoryMock mocks repository.RoleRepository
type RoleRepositoryMock struct {
	mock.Mock
}

func (m *RoleRepositoryMock) Create(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *RoleRepositoryMock) GetByID(ctx context.Context, id int64) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *RoleRepositoryMock) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *RoleRepositoryMock) GetAll(ctx context.Context) ([]*entity.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Role), args.Error(1)
}

func (m *RoleRepositoryMock) GetByUserID(ctx context.Context, userID int64) ([]*entity.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Role), args.Error(1)
}

func (m *RoleRepositoryMock) Update(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *RoleRepositoryMock) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *RoleRepositoryMock) AssignToUser(ctx context.Context, userRole *entity.UserRole) error {
	args := m.Called(ctx, userRole)
	return args.Error(0)
}

func (m *RoleRepositoryMock) RevokeFromUser(ctx context.Context, userID, roleID int64) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *RoleRepositoryMock) HasRole(ctx context.Context, userID, roleID int64) (bool, error) {
	args := m.Called(ctx, userID, roleID)
	return args.Bool(0), args.Error(1)
}

// PermissionRepositoryMock mocks repository.PermissionRepository
type PermissionRepositoryMock struct {
	mock.Mock
}

func (m *PermissionRepositoryMock) Create(ctx context.Context, permission *entity.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *PermissionRepositoryMock) GetByID(ctx context.Context, id int64) (*entity.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}

func (m *PermissionRepositoryMock) GetByCode(ctx context.Context, code string) (*entity.Permission, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}

func (m *PermissionRepositoryMock) GetAll(ctx context.Context) ([]*entity.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

func (m *PermissionRepositoryMock) GetByRoleID(ctx context.Context, roleID int64) ([]*entity.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

func (m *PermissionRepositoryMock) GetByUserID(ctx context.Context, userID int64) ([]*entity.Permission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

func (m *PermissionRepositoryMock) Update(ctx context.Context, permission *entity.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *PermissionRepositoryMock) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *PermissionRepositoryMock) AssignToRole(ctx context.Context, rolePermission *entity.RolePermission) error {
	args := m.Called(ctx, rolePermission)
	return args.Error(0)
}

func (m *PermissionRepositoryMock) RevokeFromRole(ctx context.Context, roleID, permissionID int64) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *PermissionRepositoryMock) HasPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	args := m.Called(ctx, userID, permissionCode)
	return args.Bool(0), args.Error(1)
}

func (m *PermissionRepositoryMock) HasPermissionForResource(ctx context.Context, userID int64, resource string, action entity.Action) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

// ApartmentRepositoryMock mocks repository.ApartmentRepository
type ApartmentRepositoryMock struct {
	mock.Mock
}

func (m *ApartmentRepositoryMock) Create(ctx context.Context, apartment *entity.Apartment) error {
	args := m.Called(ctx, apartment)
	return args.Error(0)
}

func (m *ApartmentRepositoryMock) GetByID(ctx context.Context, id int64) (*entity.Apartment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Apartment), args.Error(1)
}

func (m *ApartmentRepositoryMock) GetByNumber(ctx context.Context, number string) (*entity.Apartment, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Apartment), args.Error(1)
}

func (m *ApartmentRepositoryMock) Update(ctx context.Context, apartment *entity.Apartment) error {
	args := m.Called(ctx, apartment)
	return args.Error(0)
}

func (m *ApartmentRepositoryMock) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ApartmentRepositoryMock) Restore(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ApartmentRepositoryMock) List(ctx context.Context, filter repository.ApartmentFilter) ([]*entity.Apartment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Apartment), args.Error(1)
}

func (m *ApartmentRepositoryMock) Count(ctx context.Context, filter repository.ApartmentFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *ApartmentRepositoryMock) ExistsByNumber(ctx context.Context, number string) (bool, error) {
	args := m.Called(ctx, number)
	return args.Bool(0), args.Error(1)
}

func (m *ApartmentRepositoryMock) GetByFloor(ctx context.Context, floor int) ([]*entity.Apartment, error) {
	args := m.Called(ctx, floor)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Apartment), args.Error(1)
}

func (m *ApartmentRepositoryMock) GetByEntrance(ctx context.Context, entrance int) ([]*entity.Apartment, error) {
	args := m.Called(ctx, entrance)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Apartment), args.Error(1)
}

func (m *ApartmentRepositoryMock) GetStatistics(ctx context.Context) (*repository.ApartmentStatistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ApartmentStatistics), args.Error(1)
}

// PasswordHasherMock mocks service.PasswordHasher
type PasswordHasherMock struct {
	mock.Mock
}

func (m *PasswordHasherMock) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *PasswordHasherMock) Verify(password, hash string) error {
	args := m.Called(password, hash)
	return args.Error(0)
}

// ContractorRepositoryMock mocks repository.ContractorRepository
type ContractorRepositoryMock struct {
	mock.Mock
}

func (m *ContractorRepositoryMock) Create(ctx context.Context, contractor *entity.Contractor) error {
	args := m.Called(ctx, contractor)
	return args.Error(0)
}

func (m *ContractorRepositoryMock) GetByID(ctx context.Context, id int64) (*entity.Contractor, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Contractor), args.Error(1)
}

func (m *ContractorRepositoryMock) GetByEDRPOU(ctx context.Context, edrpou string) (*entity.Contractor, error) {
	args := m.Called(ctx, edrpou)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Contractor), args.Error(1)
}

func (m *ContractorRepositoryMock) Update(ctx context.Context, contractor *entity.Contractor) error {
	args := m.Called(ctx, contractor)
	return args.Error(0)
}

func (m *ContractorRepositoryMock) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ContractorRepositoryMock) Restore(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ContractorRepositoryMock) List(ctx context.Context, filter repository.ContractorFilter) ([]*entity.Contractor, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Contractor), args.Error(1)
}

func (m *ContractorRepositoryMock) Count(ctx context.Context, filter repository.ContractorFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *ContractorRepositoryMock) ExistsByEDRPOU(ctx context.Context, edrpou string) (bool, error) {
	args := m.Called(ctx, edrpou)
	return args.Bool(0), args.Error(1)
}

func (m *ContractorRepositoryMock) Search(ctx context.Context, query string, limit int) ([]*entity.Contractor, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Contractor), args.Error(1)
}

func (m *ContractorRepositoryMock) GetByType(ctx context.Context, contractorType entity.ContractorType) ([]*entity.Contractor, error) {
	args := m.Called(ctx, contractorType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Contractor), args.Error(1)
}

// ExpenseCategoryRepositoryMock mocks repository.ExpenseCategoryRepository
type ExpenseCategoryRepositoryMock struct {
	mock.Mock
}

func (m *ExpenseCategoryRepositoryMock) Create(ctx context.Context, category *entity.ExpenseCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *ExpenseCategoryRepositoryMock) GetByID(ctx context.Context, id int64) (*entity.ExpenseCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ExpenseCategory), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) GetByCode(ctx context.Context, code string) (*entity.ExpenseCategory, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ExpenseCategory), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) Update(ctx context.Context, category *entity.ExpenseCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *ExpenseCategoryRepositoryMock) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ExpenseCategoryRepositoryMock) Restore(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ExpenseCategoryRepositoryMock) List(ctx context.Context, filter repository.ExpenseCategoryFilter) ([]*entity.ExpenseCategory, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ExpenseCategory), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) Count(ctx context.Context, filter repository.ExpenseCategoryFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) GetRootCategories(ctx context.Context) ([]*entity.ExpenseCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ExpenseCategory), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) GetChildren(ctx context.Context, parentID int64) ([]*entity.ExpenseCategory, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ExpenseCategory), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) GetTree(ctx context.Context) ([]*entity.ExpenseCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ExpenseCategory), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) GetAncestors(ctx context.Context, categoryID int64) ([]*entity.ExpenseCategory, error) {
	args := m.Called(ctx, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ExpenseCategory), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) GetDescendants(ctx context.Context, categoryID int64) ([]*entity.ExpenseCategory, error) {
	args := m.Called(ctx, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ExpenseCategory), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) HasChildren(ctx context.Context, categoryID int64) (bool, error) {
	args := m.Called(ctx, categoryID)
	return args.Bool(0), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) ExistsByCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) GetDepth(ctx context.Context, categoryID int64) (int, error) {
	args := m.Called(ctx, categoryID)
	return args.Int(0), args.Error(1)
}

func (m *ExpenseCategoryRepositoryMock) Move(ctx context.Context, categoryID int64, newParentID *int64) error {
	args := m.Called(ctx, categoryID, newParentID)
	return args.Error(0)
}

func (m *ExpenseCategoryRepositoryMock) GetByType(ctx context.Context, categoryType entity.CategoryType) ([]*entity.ExpenseCategory, error) {
	args := m.Called(ctx, categoryType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ExpenseCategory), args.Error(1)
}
