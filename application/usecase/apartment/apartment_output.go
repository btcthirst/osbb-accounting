// application/usecase/apartment/apartment_output.go
package apartment

import "osbb-accounting/domain/entity"

// ApartmentOutput - результат операцій з квартирою.
type ApartmentOutput struct {
	ID              int64
	ApartmentNumber string
	Floor           int
	Entrance        *int
	AreaTotal       float64
	AreaLiving      *float64
	RoomsCount      *int
	CadastralNumber *string
	Notes           *string
	IsActive        bool
	DisplayName     string
	FullInfo        string
	AreaInfo        string
}

// ApartmentStatisticsOutput - статистика квартир.
type ApartmentStatisticsOutput struct {
	TotalApartments int64
	TotalArea       float64
	AverageArea     float64
	MinArea         float64
	MaxArea         float64
	FloorCount      int
	EntranceCount   int
	WithOwners      int64
	WithoutOwners   int64
}

// mapApartmentToOutput - helper function для конвертації entity в output.
func mapApartmentToOutput(apt *entity.Apartment) *ApartmentOutput {
	return &ApartmentOutput{
		ID:              apt.ID,
		ApartmentNumber: apt.ApartmentNumber,
		Floor:           apt.Floor,
		Entrance:        apt.Entrance,
		AreaTotal:       apt.AreaTotal,
		AreaLiving:      apt.AreaLiving,
		RoomsCount:      apt.RoomsCount,
		CadastralNumber: apt.CadastralNumber,
		Notes:           apt.Notes,
		IsActive:        apt.IsActive,
		DisplayName:     apt.GetDisplayName(),
		FullInfo:        apt.GetFullInfo(),
		AreaInfo:        apt.GetAreaInfo(),
	}
}
