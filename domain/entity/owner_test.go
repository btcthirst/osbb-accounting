// domain/entity/owner_test.go
package entity

import (
	"strings"
	"testing"
)

func TestNewOwner_Success(t *testing.T) {
	firstName := "Іван"
	lastName := "Петренко"

	owner, err := NewOwner(firstName, lastName, nil, nil, nil, nil)

	if err != nil {
		t.Fatalf("Очікувалось створення власника, отримано помилку: %v", err)
	}
	if owner.FirstName != firstName {
		t.Errorf("Очікувалось %s, отримано %s", firstName, owner.FirstName)
	}
	if !owner.IsActive {
		t.Error("Новий власник має бути активним")
	}
}

func TestNewOwner_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		wantErr   error
	}{
		{
			name:      "Empty First Name",
			firstName: "",
			lastName:  "Doe",
			wantErr:   ErrOwnerFirstNameRequired,
		},
		{
			name:      "Empty Last Name",
			firstName: "John",
			lastName:  "",
			wantErr:   ErrOwnerLastNameRequired,
		},
		{
			name:      "Name Too Long",
			firstName: strings.Repeat("a", 101),
			lastName:  "Doe",
			wantErr:   ErrOwnerNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOwner(tt.firstName, tt.lastName, nil, nil, nil, nil)
			if err != tt.wantErr {
				t.Errorf("Очікувалась помилка %v, отримано %v", tt.wantErr, err)
			}
		})
	}
}
