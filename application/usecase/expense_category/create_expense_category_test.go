package expense_category_test

import (
	"context"
	"testing"

	"osbb-accounting/application/usecase/expense_category"
	"osbb-accounting/application/usecase/mocks"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateExpenseCategoryUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo := new(mocks.ExpenseCategoryRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := expense_category.NewCreateExpenseCategoryUseCase(repo, permRepo)

		code := "TEST-001"
		input := expense_category.CreateExpenseCategoryInput{
			CurrentUserID: 1,
			Name:          "Test Category",
			Code:          &code,
			CategoryType:  entity.CategoryTypeOther,
		}

		permRepo.On("HasPermissionForResource", ctx, int64(1), "expenses", entity.ActionCreate).Return(true, nil)
		repo.On("ExistsByCode", ctx, "TEST-001").Return(false, nil)
		repo.On("Create", ctx, mock.AnythingOfType("*entity.ExpenseCategory")).Return(nil)
		repo.On("HasChildren", ctx, mock.AnythingOfType("int64")).Return(false, nil)

		output, err := useCase.Execute(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Test Category", output.Name)
		assert.Equal(t, "TEST-001", *output.Code)

		repo.AssertExpectations(t)
		permRepo.AssertExpectations(t)
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		repo := new(mocks.ExpenseCategoryRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := expense_category.NewCreateExpenseCategoryUseCase(repo, permRepo)

		permRepo.On("HasPermissionForResource", ctx, int64(1), "expenses", entity.ActionCreate).Return(false, nil)

		input := expense_category.CreateExpenseCategoryInput{CurrentUserID: 1}
		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrPermissionDenied))
	})

	t.Run("DuplicateCode", func(t *testing.T) {
		repo := new(mocks.ExpenseCategoryRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := expense_category.NewCreateExpenseCategoryUseCase(repo, permRepo)

		code := "TEST-001"
		input := expense_category.CreateExpenseCategoryInput{
			CurrentUserID: 1,
			Code:          &code,
		}

		permRepo.On("HasPermissionForResource", ctx, int64(1), "expenses", entity.ActionCreate).Return(true, nil)
		repo.On("ExistsByCode", ctx, "TEST-001").Return(true, nil)

		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrAlreadyExists))
	})

	t.Run("ParentNotFound", func(t *testing.T) {
		repo := new(mocks.ExpenseCategoryRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := expense_category.NewCreateExpenseCategoryUseCase(repo, permRepo)

		parentID := int64(999)
		input := expense_category.CreateExpenseCategoryInput{
			CurrentUserID: 1,
			ParentID:      &parentID,
		}

		permRepo.On("HasPermissionForResource", ctx, int64(1), "expenses", entity.ActionCreate).Return(true, nil)
		repo.On("GetByID", ctx, int64(999)).Return(nil, domainErrors.ErrNotFound)

		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "parent category not found")
	})
}
