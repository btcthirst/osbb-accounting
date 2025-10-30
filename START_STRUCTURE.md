osbb-accounting/
├── main.go                          # Точка входу додатку
├── go.mod                           # Go модулі
├── go.sum
│
├── domain/                          # Entities (бізнес-об'єкти)
│   ├── user.go                      # Модель користувача
│   ├── apartment.go                 # Модель квартири (майбутнє)
│   └── payment.go                   # Модель платежу (майбутнє)
│
├── repository/                      # Data Access Layer (інтерфейси)
│   ├── user_repository.go           # Контракт для роботи з користувачами
│   └── sqlite/                      # SQLite реалізація
│       ├── user_repository_impl.go  # Імплементація UserRepository
│       └── connection.go            # Менеджер з'єднання з БД
│
├── service/                         # Business Logic Layer
│   ├── auth_service.go              # Сервіс аутентифікації
│   └── user_service.go              # Бізнес-логіка користувачів (майбутнє)
│
├── ui/                              # Presentation Layer (Fyne)
│   ├── login_screen.go              # Екран входу
│   ├── main_window.go               # Головне вікно
│   └── components/                  # UI компоненти
│
├── config/                          # Конфігурація
│   └── config.go                    # Налаштування додатку
│
└── migrations/                      # SQL міграції
    └── 001_init_schema.sql          # Початкова схема БД