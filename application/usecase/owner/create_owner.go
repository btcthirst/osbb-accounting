// application/usecase/owner/create_owner.go
package owner

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// CreateOwnerUseCase створює нового власника.
type CreateOwnerUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewCreateOwnerUseCase створює новий use case.
func NewCreateOwnerUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *CreateOwnerUseCase {
	return &CreateOwnerUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateOwnerInput - вхідні дані.
type CreateOwnerInput struct {
	CurrentUserID     int64
	FirstName         string
	LastName          string
	MiddleName        *string
	Phone             *string
	Email             *string
	TaxNumber         *string
	PassportSeries    *string
	PassportNumber    *string
	RegisteredAddress *string
	ActualAddress     *string
	Notes             *string
}

// OwnerOutput - результат операцій з власником.
type OwnerOutput struct {
	ID                int64
	FirstName         string
	LastName          string
	MiddleName        *string
	Phone             *string
	Email             *string
	TaxNumber         *string
	PassportSeries    *string
	PassportNumber    *string
	RegisteredAddress *string
	ActualAddress     *string
	Notes             *string
	IsActive          bool
	FullName          string
	ShortName         string
}

// Execute виконує створення власника.
func (uc *CreateOwnerUseCase) Execute(ctx context.Context, input CreateOwnerInput) (*OwnerOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodePermissionDenied,
			"insufficient permissions to create owner",
			domainErrors.ErrPermissionDenied,
		)
	}

	// Перевірка унікальності ІПН (якщо вказано)
	if input.TaxNumber != nil && *input.TaxNumber != "" {
		exists, err := uc.ownerRepo.ExistsByTaxNumber(ctx, *input.TaxNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to check tax number: %w", err)
		}
		if exists {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"owner with this tax number already exists",
				domainErrors.ErrAlreadyExists,
			).WithDetails("field", "tax_number")
		}
	}

	// Створення entity
	owner, err := entity.NewOwner(
		input.FirstName,
		input.LastName,
		input.MiddleName,
		input.Phone,
		input.Email,
		input.TaxNumber,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid owner data",
			err,
		)
	}

	// Додаткові поля
	owner.PassportSeries = input.PassportSeries
	owner.PassportNumber = input.PassportNumber
	owner.RegisteredAddress = input.RegisteredAddress
	owner.ActualAddress = input.ActualAddress
	owner.Notes = input.Notes

	// Збереження
	if err := uc.ownerRepo.Create(ctx, owner); err != nil {
		return nil, fmt.Errorf("failed to create owner: %w", err)
	}

	return mapOwnerToOutput(owner), nil
}

// ============================================================================

// GetOwnerUseCase отримує власника за ID.
type GetOwnerUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewGetOwnerUseCase створює новий use case.
func NewGetOwnerUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *GetOwnerUseCase {
	return &GetOwnerUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// GetOwnerInput - вхідні дані.
type GetOwnerInput struct {
	CurrentUserID int64
	OwnerID       int64
}

// Execute виконує отримання власника.
func (uc *GetOwnerUseCase) Execute(ctx context.Context, input GetOwnerInput) (*OwnerOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання власника
	owner, err := uc.ownerRepo.GetByID(ctx, input.OwnerID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"owner not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}

	return mapOwnerToOutput(owner), nil
}

// ============================================================================

// UpdateOwnerUseCase оновлює дані власника.
type UpdateOwnerUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewUpdateOwnerUseCase створює новий use case.
func NewUpdateOwnerUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateOwnerUseCase {
	return &UpdateOwnerUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// UpdateOwnerInput - вхідні дані.
type UpdateOwnerInput struct {
	CurrentUserID     int64
	OwnerID           int64
	FirstName         string
	LastName          string
	MiddleName        *string
	Phone             *string
	Email             *string
	TaxNumber         *string
	PassportSeries    *string
	PassportNumber    *string
	RegisteredAddress *string
	ActualAddress     *string
	Notes             *string
}

// Execute виконує оновлення власника.
func (uc *UpdateOwnerUseCase) Execute(ctx context.Context, input UpdateOwnerInput) (*OwnerOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання поточних даних
	owner, err := uc.ownerRepo.GetByID(ctx, input.OwnerID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"owner not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}

	// Перевірка унікальності ІПН (якщо змінився)
	if input.TaxNumber != nil && *input.TaxNumber != "" {
		if owner.TaxNumber == nil || *owner.TaxNumber != *input.TaxNumber {
			exists, err := uc.ownerRepo.ExistsByTaxNumber(ctx, *input.TaxNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to check tax number: %w", err)
			}
			if exists {
				return nil, domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"owner with this tax number already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "tax_number")
			}
		}
	}

	// Оновлення основних полів
	if err := owner.Update(
		input.FirstName,
		input.LastName,
		input.MiddleName,
		input.Phone,
		input.Email,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid owner data",
			err,
		)
	}

	// Оновлення ІПН
	if err := owner.UpdateTaxNumber(input.TaxNumber); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid tax number",
			err,
		)
	}

	// Оновлення паспорту та адрес
	owner.UpdatePassport(input.PassportSeries, input.PassportNumber)
	owner.UpdateAddresses(input.RegisteredAddress, input.ActualAddress)
	owner.Notes = input.Notes

	// Збереження
	if err := uc.ownerRepo.Update(ctx, owner); err != nil {
		return nil, fmt.Errorf("failed to update owner: %w", err)
	}

	return mapOwnerToOutput(owner), nil
}

// ============================================================================

// DeleteOwnerUseCase видаляє (деактивує) власника.
type DeleteOwnerUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewDeleteOwnerUseCase створює новий use case.
func NewDeleteOwnerUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteOwnerUseCase {
	return &DeleteOwnerUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// DeleteOwnerInput - вхідні дані.
type DeleteOwnerInput struct {
	CurrentUserID int64
	OwnerID       int64
}

// DeleteOwnerOutput - результат.
type DeleteOwnerOutput struct {
	Success bool
	Message string
}

// Execute виконує видалення власника.
func (uc *DeleteOwnerUseCase) Execute(ctx context.Context, input DeleteOwnerInput) (*DeleteOwnerOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Перевірка існування
	owner, err := uc.ownerRepo.GetByID(ctx, input.OwnerID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"owner not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}

	// Soft delete
	if err := uc.ownerRepo.SoftDelete(ctx, input.OwnerID); err != nil {
		return nil, fmt.Errorf("failed to delete owner: %w", err)
	}

	return &DeleteOwnerOutput{
		Success: true,
		Message: fmt.Sprintf("Owner '%s' successfully deleted", owner.FullName()),
	}, nil
}

// ============================================================================

// ListOwnersUseCase отримує список власників.
type ListOwnersUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewListOwnersUseCase створює новий use case.
func NewListOwnersUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *ListOwnersUseCase {
	return &ListOwnersUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// ListOwnersInput - вхідні дані.
type ListOwnersInput struct {
	CurrentUserID  int64
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string
	HasTaxNumber   *bool
	HasContact     *bool
	Limit          int
	Offset         int
	OrderBy        string
	OrderDesc      bool
}

// ListOwnersOutput - результат.
type ListOwnersOutput struct {
	Owners  []*OwnerOutput
	Total   int64
	Limit   int
	Offset  int
	HasMore bool
}

// Execute виконує отримання списку власників.
func (uc *ListOwnersUseCase) Execute(ctx context.Context, input ListOwnersInput) (*ListOwnersOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Підготовка фільтру
	filter := repository.OwnerFilter{
		IncludeDeleted: input.IncludeDeleted,
		IsActive:       input.IsActive,
		SearchQuery:    input.SearchQuery,
		HasTaxNumber:   input.HasTaxNumber,
		HasContact:     input.HasContact,
		Limit:          input.Limit,
		Offset:         input.Offset,
		OrderBy:        input.OrderBy,
		OrderDesc:      input.OrderDesc,
	}

	// За замовчуванням
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Отримання списку
	owners, err := uc.ownerRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list owners: %w", err)
	}

	// Підрахунок загальної кількості
	total, err := uc.ownerRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count owners: %w", err)
	}

	// Конвертація в output
	ownerOutputs := make([]*OwnerOutput, len(owners))
	for i, owner := range owners {
		ownerOutputs[i] = mapOwnerToOutput(owner)
	}

	return &ListOwnersOutput{
		Owners:  ownerOutputs,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: int64(filter.Offset+filter.Limit) < total,
	}, nil
}

// ============================================================================

// SearchOwnersUseCase шукає власників (для autocomplete).
type SearchOwnersUseCase struct {
	ownerRepo      repository.OwnerRepository
	permissionRepo repository.PermissionRepository
}

// NewSearchOwnersUseCase створює новий use case.
func NewSearchOwnersUseCase(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *SearchOwnersUseCase {
	return &SearchOwnersUseCase{
		ownerRepo:      ownerRepo,
		permissionRepo: permissionRepo,
	}
}

// SearchOwnersInput - вхідні дані.
type SearchOwnersInput struct {
	CurrentUserID int64
	Query         string
	Limit         int
}

// Execute виконує швидкий пошук власників.
func (uc *SearchOwnersUseCase) Execute(ctx context.Context, input SearchOwnersInput) ([]*OwnerOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "owners", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Мінімальна довжина запиту
	if len(input.Query) < 2 {
		return []*OwnerOutput{}, nil
	}

	// Ліміт за замовчуванням
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	// Пошук
	owners, err := uc.ownerRepo.Search(ctx, input.Query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search owners: %w", err)
	}

	// Конвертація
	result := make([]*OwnerOutput, len(owners))
	for i, owner := range owners {
		result[i] = mapOwnerToOutput(owner)
	}

	return result, nil
}

// ============================================================================

// Helper функція для конвертації entity в output
func mapOwnerToOutput(owner *entity.Owner) *OwnerOutput {
	return &OwnerOutput{
		ID:                owner.ID,
		FirstName:         owner.FirstName,
		LastName:          owner.LastName,
		MiddleName:        owner.MiddleName,
		Phone:             owner.Phone,
		Email:             owner.Email,
		TaxNumber:         owner.TaxNumber,
		PassportSeries:    owner.PassportSeries,
		PassportNumber:    owner.PassportNumber,
		RegisteredAddress: owner.RegisteredAddress,
		ActualAddress:     owner.ActualAddress,
		Notes:             owner.Notes,
		IsActive:          owner.IsActive,
		FullName:          owner.FullName(),
		ShortName:         owner.ShortName(),
	}
}
