package entity_test

import (
	"testing"

	"osbb-accounting/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApartment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		apt, err := entity.NewApartment("101", 1, 50.5, nil, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, apt)
		assert.Equal(t, "101", apt.ApartmentNumber)
		assert.Equal(t, 50.5, apt.AreaTotal)
		assert.True(t, apt.IsActive)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name      string
			number    string
			floor     int
			areaTotal float64
			wantErr   error
		}{
			{
				name:      "Empty Number",
				number:    "",
				floor:     1,
				areaTotal: 50.0,
				wantErr:   entity.ErrApartmentNumberRequired,
			},
			{
				name:      "Negative Floor",
				number:    "101",
				floor:     -1,
				areaTotal: 50.0,
				wantErr:   entity.ErrApartmentFloorInvalid,
			},
			{
				name:      "Zero Area",
				number:    "101",
				floor:     1,
				areaTotal: 0,
				wantErr:   entity.ErrApartmentAreaTotalInvalid,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := entity.NewApartment(tt.number, tt.floor, tt.areaTotal, nil, nil, nil)
				assert.ErrorIs(t, err, tt.wantErr)
			})
		}
	})
}

func TestApartment_Update(t *testing.T) {
	apt, _ := entity.NewApartment("101", 1, 50.0, nil, nil, nil)

	t.Run("Success", func(t *testing.T) {
		newArea := 55.0
		err := apt.Update(2, newArea, nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 2, apt.Floor)
		assert.Equal(t, newArea, apt.AreaTotal)
	})

	t.Run("Validation Error", func(t *testing.T) {
		err := apt.Update(-1, 50.0, nil, nil, nil, nil, nil)
		assert.ErrorIs(t, err, entity.ErrApartmentFloorInvalid)
	})
}

func TestApartment_GetInfo(t *testing.T) {
	entrance := 1
	livingArea := 30.0
	apt, _ := entity.NewApartment("101", 2, 50.0, &entrance, &livingArea, nil)

	t.Run("GetDisplayName", func(t *testing.T) {
		assert.Equal(t, "Кв. 101", apt.GetDisplayName())
	})

	t.Run("GetFullInfo", func(t *testing.T) {
		info := apt.GetFullInfo()
		assert.Contains(t, info, "Кв. 101")
		assert.Contains(t, info, "під'їзд 1")
		assert.Contains(t, info, "2 пов.")
	})

	t.Run("GetAreaInfo", func(t *testing.T) {
		info := apt.GetAreaInfo()
		assert.Contains(t, info, "50.0 м²")
		assert.Contains(t, info, "житл. 30.0 м²")
	})
}
