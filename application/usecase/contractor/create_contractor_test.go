package contractor_test

import (
	"context"
	"testing"

	"osbb-accounting/application/usecase/contractor"
	"osbb-accounting/application/usecase/mocks"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateContractorUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo := new(mocks.ContractorRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := contractor.NewCreateContractorUseCase(repo, permRepo)

		edrpou := "12345678"
		input := contractor.CreateContractorInput{
			CurrentUserID:  1,
			Name:           "Test Contractor",
			EDRPOU:         &edrpou,
			ContractorType: entity.ContractorTypeSupplier,
		}

		permRepo.On("HasPermissionForResource", ctx, int64(1), "contractors", entity.ActionCreate).Return(true, nil)
		repo.On("ExistsByEDRPOU", ctx, "12345678").Return(false, nil)
		repo.On("Create", ctx, mock.AnythingOfType("*entity.Contractor")).Return(nil)

		output, err := useCase.Execute(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Test Contractor", output.Name)
		assert.Equal(t, "12345678", *output.EDRPOU)

		repo.AssertExpectations(t)
		permRepo.AssertExpectations(t)
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		repo := new(mocks.ContractorRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := contractor.NewCreateContractorUseCase(repo, permRepo)

		permRepo.On("HasPermissionForResource", ctx, int64(1), "contractors", entity.ActionCreate).Return(false, nil)

		input := contractor.CreateContractorInput{CurrentUserID: 1}
		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrPermissionDenied))
	})

	t.Run("DuplicateEDRPOU", func(t *testing.T) {
		repo := new(mocks.ContractorRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		useCase := contractor.NewCreateContractorUseCase(repo, permRepo)

		edrpou := "12345678"
		input := contractor.CreateContractorInput{
			CurrentUserID: 1,
			EDRPOU:        &edrpou,
		}

		permRepo.On("HasPermissionForResource", ctx, int64(1), "contractors", entity.ActionCreate).Return(true, nil)
		repo.On("ExistsByEDRPOU", ctx, "12345678").Return(true, nil)

		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrAlreadyExists))
	})
}
