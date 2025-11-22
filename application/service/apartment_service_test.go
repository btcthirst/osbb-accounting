package service_test

import (
	"context"
	"testing"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/apartment"
	"osbb-accounting/application/usecase/mocks"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestApartmentService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo := new(mocks.ApartmentRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		svc := service.NewApartmentService(repo, permRepo)

		entrance := 1
		areaLiving := 30.0
		input := apartment.CreateApartmentInput{
			CurrentUserID:   1,
			ApartmentNumber: "101",
			Floor:           1,
			Entrance:        &entrance,
			AreaTotal:       50.5,
			AreaLiving:      &areaLiving,
		}

		permRepo.On("HasPermissionForResource", ctx, int64(1), "apartments", entity.ActionCreate).Return(true, nil)
		repo.On("ExistsByNumber", ctx, "101").Return(false, nil)
		repo.On("Create", ctx, mock.AnythingOfType("*entity.Apartment")).Return(nil)

		output, err := svc.Create(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "101", output.ApartmentNumber)

		repo.AssertExpectations(t)
		permRepo.AssertExpectations(t)
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		repo := new(mocks.ApartmentRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		svc := service.NewApartmentService(repo, permRepo)

		permRepo.On("HasPermissionForResource", ctx, int64(1), "apartments", entity.ActionCreate).Return(false, nil)

		input := apartment.CreateApartmentInput{CurrentUserID: 1}
		output, err := svc.Create(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrPermissionDenied))
	})
}

func TestApartmentService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo := new(mocks.ApartmentRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		svc := service.NewApartmentService(repo, permRepo)

		input := apartment.ListApartmentsInput{
			CurrentUserID: 1,
			Limit:         10,
			Offset:        0,
		}

		permRepo.On("HasPermissionForResource", ctx, int64(1), "apartments", entity.ActionRead).Return(true, nil)
		repo.On("List", ctx, mock.AnythingOfType("repository.ApartmentFilter")).Return([]*entity.Apartment{
			{ID: 1, ApartmentNumber: "101"},
			{ID: 2, ApartmentNumber: "102"},
		}, nil)
		repo.On("Count", ctx, mock.AnythingOfType("repository.ApartmentFilter")).Return(int64(2), nil)

		output, err := svc.List(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Len(t, output.Apartments, 2)
		assert.Equal(t, int64(2), output.Total)
	})
}
