package apartment_test

import (
	"context"
	"errors"
	"testing"

	"osbb-accounting/application/usecase/apartment"
	"osbb-accounting/application/usecase/mocks"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateApartmentUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		aptRepo := new(mocks.ApartmentRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := apartment.NewCreateApartmentUseCase(aptRepo, permRepo)

		input := apartment.CreateApartmentInput{
			CurrentUserID:   1,
			ApartmentNumber: "101",
			Floor:           1,
			AreaTotal:       50.5,
		}

		permRepo.On("HasPermissionForResource", ctx, int64(1), "apartments", entity.ActionCreate).Return(true, nil)
		aptRepo.On("ExistsByNumber", ctx, "101").Return(false, nil)
		aptRepo.On("Create", ctx, mock.AnythingOfType("*entity.Apartment")).Return(nil)

		output, err := useCase.Execute(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "101", output.ApartmentNumber)

		aptRepo.AssertExpectations(t)
		permRepo.AssertExpectations(t)
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		aptRepo := new(mocks.ApartmentRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := apartment.NewCreateApartmentUseCase(aptRepo, permRepo)

		permRepo.On("HasPermissionForResource", ctx, int64(1), "apartments", entity.ActionCreate).Return(false, nil)

		input := apartment.CreateApartmentInput{CurrentUserID: 1}
		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrPermissionDenied))
	})

	t.Run("DuplicateNumber", func(t *testing.T) {
		aptRepo := new(mocks.ApartmentRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := apartment.NewCreateApartmentUseCase(aptRepo, permRepo)

		permRepo.On("HasPermissionForResource", ctx, int64(1), "apartments", entity.ActionCreate).Return(true, nil)
		aptRepo.On("ExistsByNumber", ctx, "101").Return(true, nil)

		input := apartment.CreateApartmentInput{CurrentUserID: 1, ApartmentNumber: "101"}
		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrAlreadyExists))
	})

	t.Run("RepoError", func(t *testing.T) {
		aptRepo := new(mocks.ApartmentRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := apartment.NewCreateApartmentUseCase(aptRepo, permRepo)

		permRepo.On("HasPermissionForResource", ctx, int64(1), "apartments", entity.ActionCreate).Return(true, nil)
		aptRepo.On("ExistsByNumber", ctx, "101").Return(false, errors.New("db error"))

		input := apartment.CreateApartmentInput{CurrentUserID: 1, ApartmentNumber: "101"}
		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "db error")
	})
}
