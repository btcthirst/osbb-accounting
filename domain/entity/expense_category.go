// domain/entity/expense_category.go
package entity

import (
	"errors"
	"time"
)

// ExpenseCategory представляє категорію витрат (ієрархічна структура).
type ExpenseCategory struct {
	ID           int64
	Name         string
	Code         *string // Унікальний код категорії (напр. "UTIL-ELEC")
	ParentID     *int64  // ID батьківської категорії (NULL для кореневих)
	CategoryType CategoryType
	Description  *string
	IsActive     bool
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Додаткові поля для зручності роботи з деревом (не зберігаються в БД)
	Level    int                // Рівень вкладеності (0 - корінь)
	Path     string             // Повний шлях (напр. "Комунальні послуги / Електроенергія")
	Children []*ExpenseCategory // Дочірні категорії
}

// CategoryType представляє тип категорії витрат.
type CategoryType string

const (
	CategoryTypeUtility CategoryType = "utility" // Комунальні послуги
	CategoryTypeRepair  CategoryType = "repair"  // Ремонт та утримання
	CategoryTypeSalary  CategoryType = "salary"  // Зарплата
	CategoryTypeService CategoryType = "service" // Послуги
	CategoryTypeOther   CategoryType = "other"   // Інше
)

// Validation errors
var (
	ErrCategoryNameRequired      = errors.New("category name is required")
	ErrCategoryNameTooShort      = errors.New("category name must be at least 2 characters")
	ErrCategoryNameTooLong       = errors.New("category name must not exceed 100 characters")
	ErrCategoryCodeTooLong       = errors.New("category code must not exceed 20 characters")
	ErrCategoryTypeInvalid       = errors.New("invalid category type")
	ErrCategoryCircularReference = errors.New("circular reference detected in category hierarchy")
	ErrCategoryMaxDepthExceeded  = errors.New("maximum category depth exceeded (max 5 levels)")
	ErrCategoryHasChildren       = errors.New("cannot delete category with children")
	ErrCategoryParentNotFound    = errors.New("parent category not found")
	ErrCategoryParentInactive    = errors.New("parent category is inactive")
	ErrCategoryCannotBeOwnParent = errors.New("category cannot be its own parent")
)

// Константи для валідації
const (
	MaxCategoryDepth      = 5 // Максимальна глибина вкладеності
	MaxCategoryNameLength = 100
	MaxCategoryCodeLength = 20
)

// NewExpenseCategory створює нову категорію витрат з валідацією.
func NewExpenseCategory(
	name string,
	categoryType CategoryType,
	parentID *int64,
	code *string,
	description *string,
) (*ExpenseCategory, error) {
	category := &ExpenseCategory{
		Name:         name,
		CategoryType: categoryType,
		ParentID:     parentID,
		Code:         code,
		Description:  description,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Level:        0,
		Children:     make([]*ExpenseCategory, 0),
	}

	if err := category.Validate(); err != nil {
		return nil, err
	}

	return category, nil
}

// Validate перевіряє коректність всіх полів категорії.
func (c *ExpenseCategory) Validate() error {
	// Name validation
	if c.Name == "" {
		return ErrCategoryNameRequired
	}
	if len(c.Name) < 2 {
		return ErrCategoryNameTooShort
	}
	if len(c.Name) > MaxCategoryNameLength {
		return ErrCategoryNameTooLong
	}

	// Code validation (optional)
	if c.Code != nil && len(*c.Code) > MaxCategoryCodeLength {
		return ErrCategoryCodeTooLong
	}

	// CategoryType validation
	if !c.CategoryType.IsValid() {
		return ErrCategoryTypeInvalid
	}

	// Self-parent check
	if c.ParentID != nil && c.ID != 0 && *c.ParentID == c.ID {
		return ErrCategoryCannotBeOwnParent
	}

	return nil
}

// IsValid перевіряє чи валідний тип категорії.
func (ct CategoryType) IsValid() bool {
	switch ct {
	case CategoryTypeUtility, CategoryTypeRepair, CategoryTypeSalary,
		CategoryTypeService, CategoryTypeOther:
		return true
	default:
		return false
	}
}

// Update оновлює дані категорії.
func (c *ExpenseCategory) Update(
	name string,
	categoryType CategoryType,
	description *string,
) error {
	c.Name = name
	c.CategoryType = categoryType
	c.Description = description
	c.UpdatedAt = time.Now()

	return c.Validate()
}

// UpdateCode оновлює код категорії.
func (c *ExpenseCategory) UpdateCode(code *string) error {
	c.Code = code
	c.UpdatedAt = time.Now()
	return c.Validate()
}

// SetParent встановлює батьківську категорію.
func (c *ExpenseCategory) SetParent(parentID *int64) error {
	// Перевірка на self-parent
	if parentID != nil && c.ID != 0 && *parentID == c.ID {
		return ErrCategoryCannotBeOwnParent
	}

	c.ParentID = parentID
	c.UpdatedAt = time.Now()
	return nil
}

// Deactivate деактивує категорію.
func (c *ExpenseCategory) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

// Activate активує категорію.
func (c *ExpenseCategory) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

// SoftDelete виконує м'яке видалення категорії.
func (c *ExpenseCategory) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.IsActive = false
	c.UpdatedAt = now
}

// IsDeleted перевіряє чи видалена категорія.
func (c *ExpenseCategory) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsRoot перевіряє чи є категорія кореневою.
func (c *ExpenseCategory) IsRoot() bool {
	return c.ParentID == nil
}

// HasChildren перевіряє чи має категорія дочірні елементи.
func (c *ExpenseCategory) HasChildren() bool {
	return len(c.Children) > 0
}

// AddChild додає дочірню категорію.
func (c *ExpenseCategory) AddChild(child *ExpenseCategory) error {
	// Перевірка глибини
	if child.Level >= MaxCategoryDepth {
		return ErrCategoryMaxDepthExceeded
	}

	child.ParentID = &c.ID
	child.Level = c.Level + 1
	c.Children = append(c.Children, child)

	return nil
}

// RemoveChild видаляє дочірню категорію.
func (c *ExpenseCategory) RemoveChild(childID int64) {
	for i, child := range c.Children {
		if child.ID == childID {
			c.Children = append(c.Children[:i], c.Children[i+1:]...)
			break
		}
	}
}

// GetFullPath повертає повний шлях категорії.
func (c *ExpenseCategory) GetFullPath() string {
	if c.Path != "" {
		return c.Path
	}
	return c.Name
}

// GetDisplayName повертає відображуване ім'я з кодом.
func (c *ExpenseCategory) GetDisplayName() string {
	if c.Code != nil && *c.Code != "" {
		return c.Name + " [" + *c.Code + "]"
	}
	return c.Name
}

// GetTypeName повертає локалізовану назву типу категорії.
func (ct CategoryType) GetDisplayName() string {
	switch ct {
	case CategoryTypeUtility:
		return "Комунальні послуги"
	case CategoryTypeRepair:
		return "Ремонт та утримання"
	case CategoryTypeSalary:
		return "Зарплата"
	case CategoryTypeService:
		return "Послуги"
	case CategoryTypeOther:
		return "Інше"
	default:
		return "Невідомо"
	}
}

// CalculateLevel розраховує рівень вкладеності на основі батьківської категорії.
func (c *ExpenseCategory) CalculateLevel(parent *ExpenseCategory) {
	if parent == nil {
		c.Level = 0
	} else {
		c.Level = parent.Level + 1
	}
}

// BuildPath будує повний шлях категорії.
func (c *ExpenseCategory) BuildPath(ancestors []string) {
	if len(ancestors) == 0 {
		c.Path = c.Name
	} else {
		c.Path = ""
		for _, ancestor := range ancestors {
			c.Path += ancestor + " / "
		}
		c.Path += c.Name
	}
}
