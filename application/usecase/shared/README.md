# Shared Use Case Types

## DeleteOutput

Універсальний тип для результатів операцій видалення.

### Використання

```go
import "osbb-accounting/application/usecase/shared"

// У будь-якому use case:
func (uc *DeleteXxxUseCase) Execute(ctx context.Context, input DeleteXxxInput) (*shared.DeleteOutput, error) {
    // ... бізнес-логіка ...
    
    // Успішне видалення:
    return shared.NewDeleteOutput("Entity successfully deleted"), nil
    
    // Або з custom повідомленням:
    return &shared.DeleteOutput{
        Success: true,
        Message: fmt.Sprintf("Entity %d deleted", id),
    }, nil
}
```

### Приклад міграції

**До (з дублюванням):**
```go
// У кожному пакеті окремо:
type DeleteApartmentOutput struct {
    Success bool
    Message string
}

type DeleteChargeOutput struct {
    Success bool
    Message string
}

type DeletePaymentOutput struct {
    Success bool
    Message string
}
```

**Після (зі shared типом):**
```go
import "osbb-accounting/application/usecase/shared"

// В усіх пакетах використовується єдиний тип:
// apartment, charge, payment, etc.
return shared.NewDeleteOutput("Successfully deleted"), nil
```

### Переваги

1. **DRY принцип** - один тип замість багатьох дублікатів
2. **Консистентність** - однаковий формат відповіді всюди
3. **Простота підтримки** - зміни в одному місці
4. **Кращий TypeScript codegen** - один інтерфейс для фронтенду

### Поточні use cases з Delete операціями

Можуть використовувати shared.DeleteOutput:
- `apartment/delete_apartment.go`
- `charge/delete_charge.go`
- `contractorpayment/delete_contractorpayment.go`
- `expense_category/delete_expense_category.go`
- `owner/delete_owner.go`
- `ownership/delete_ownership_share.go`
- `payment/delete_payment.go`

### Розташування

```
application/
└── usecase/
    └── shared/
        └── output.go  # Спільні output типи
```
