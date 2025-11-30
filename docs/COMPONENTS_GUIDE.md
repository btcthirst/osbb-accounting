# UI Components Helper Guide

## Доступні Helper Функції

Файл: `presentation/fyne/screens/components.go`

### Кнопки

```go
// Кнопка "Додати" з високим пріоритетом
createBtn := newCreateButton("Додати квартиру", func() {
    // action
})

// Кнопка "Оновити"
refreshBtn := newRefreshButton(func() {
    // action
})
```

### Поля вводу

```go
// Поле пошуку зі стандартним placeholder "🔍 Пошук..."
searchEntry := newSearchEntry("", func(query string) {
    // onChange handler
})

// Поле пошуку з кастомним placeholder
searchEntry := newSearchEntry("Пошук за номером...", func(query string) {
    // onChange handler
})

// Поле пошуку без onChange handler
searchEntry := newSearchEntry("", nil)
// Потім можна додати handler окремо:
searchEntry.OnChanged = func(query string) {
    // handler
}
```

### Випадаючі списки

```go
// Filter select з опціями
filterSelect := newFilterSelect(
    []string{"Всі", "Активні", "Неактивні"},
    "Всі", // default option
    func(value string) {
        // onChange handler
    },
)

// Filter select без default option
filterSelect := newFilterSelect(
    []string{"Опція 1", "Опція 2"},
    "", // no default
    func(value string) {
        // handler
    },
)
```

## Приклад використання в екрані

```go
func (s *ApartmentsScreen) buildUI() {
    // Створення пошукового поля
    s.searchEntry = newSearchEntry("🔍 Пошук за номером...", func(query string) {
        s.applyFilters()
    })

    // Створення фільтра
    s.filterSelect = newFilterSelect(
        []string{"Всі квартири", "Тільки активні", "Неактивні"},
        "Всі квартири",
        func(value string) {
            s.applyFilters()
        },
    )

    // Кнопки
    s.createButton = newCreateButton("Додати квартиру", func() {
        s.showCreateDialog()
    })
    
    s.refreshButton = newRefreshButton(func() {
        s.loadApartments()
    })
}
```

## Переваги

✅ **Консистентний вигляд** - всі компоненти виглядають однаково
✅ **Менше коду** - не потрібно повторювати налаштування
✅ **Легка підтримка** - зміни в одному місці
✅ **Стандартна поведінка** - всі компоненти працюють однаково
