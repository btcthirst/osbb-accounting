// application/usecase/osbb/osbb_output.go
package osbb

// GetOSBBOutput - результат операцій з ОСББ.
type GetOSBBOutput struct {
	ID            int64
	Name          string
	EDRPOU        string
	LegalAddress  string
	ActualAddress *string
	Phone         *string
	Email         *string
	Website       *string
	ChairmanName  string
}
