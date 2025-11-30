package text

// Common Actions
const (
	ActionEdit       = "✏️ Редагувати"
	ActionDelete     = "🗑️ Видалити"
	ActionDetails    = "ℹ️ Деталі"
	ActionSave       = "Зберегти"
	ActionCancel     = "Скасувати"
	ActionRefresh    = "Оновити"
	ActionAdd        = "Додати"
	ActionSearch     = "Пошук"
	ActionImportXlsx = "📥 Імпортувати XLSX"
	ActionLogout     = "🚪 Вийти з системи"
	ActionCreateOSBB = "🏠 Створити ОСББ"
	ActionEditOSBB   = "✏️ Редагувати ОСББ"
	ActionExportXlsx = "📊 Експорт в XLSX"
	ActionApprove    = "✅ Затвердити"
	ActionUnapprove  = "↩️ Скасувати затвердження"
	ActionAddPayment = "Додати платіж"
	ActionClose      = "Закрити"
)

// Common Messages
const (
	MsgConfirmDeleteTitle         = "Підтвердження видалення"
	MsgConfirmDeleteBody          = "Ви впевнені, що хочете видалити %s?"
	MsgSuccessDeleted             = "%s успішно видалено"
	MsgErrorLoading               = "Помилка завантаження: %v"
	MsgErrorDeleting              = "Помилка видалення: %v"
	MsgAccessDenied               = "Доступ заборонено"
	MsgErrorEditApproved          = "Не можна редагувати затверджену витрату"
	MsgErrorDeleteApproved        = "Не можна видалити затверджену витрату"
	MsgSuccessExpenseDeleted      = "Витрату успішно видалено"
	MsgSuccessExpenseApproved     = "Витрату затверджено"
	MsgSuccessExpenseUnapproved   = "Затвердження витрати скасовано"
	MsgCashFlowStats              = "Показано: %d операцій | Надходження: %.2f грн | Витрати: %.2f грн | Сальдо: %.2f грн"
	MsgErrorServiceNotInitialized = "%s service not initialized"
	MsgErrorExport                = "Помилка експорту: %w"
	MsgErrorSaveFile              = "Помилка збереження файлу: %w"
	MsgSuccessFileSaved           = "Файл збережено: %s"
	MsgSuccessPaymentDeleted      = "Платіж успішно видалено"
	MsgContractorPaymentStats     = "Показано: %d платежів | Загальна сума: %.2f грн"
	MsgSuccessOwnerDeleted        = "Власник '%s' успішно видалено"
	MsgSuccessShareDeleted        = "Частку власності успішно видалено"
	MsgSuccessContractorDeleted   = "Контрагента успішно видалено"
	MsgErrorOwnerLoading          = "Помилка завантаження власників: %v"
	MsgErrorShareLoading          = "Помилка завантаження часток: %v"
	MsgSuccessChargeDeleted       = "Нарахування успішно видалено"
	MsgSuccessPaymentApproved     = "Платіж підтверджено"
	MsgSuccessPaymentUnapproved   = "Підтвердження платежу скасовано"
	MsgErrorChargeLoading         = "Помилка завантаження нарахувань: %v"
	MsgErrorPaymentLoading        = "Помилка завантаження платежів: %v"
	MsgErrorCategoryLoading       = "Помилка завантаження категорій: %v"
	MsgSuccessCategoryDeleted     = "Категорію успішно видалено"
	MsgImportCompleted            = "Імпорт завершено!\n\nФайл: %s\nАркушів: %d\nЗаписів: %d\nУспішно: %d\nПомилок: %d\nЧас: %.2f сек"
	MsgConfirmDeleteBatch         = "Ви впевнені, що хочете видалити цей імпорт?\n\nБатч #%d\nФайл: %s\nЗаписів: %d\n\nЦе дія незворотна!"
	MsgSuccessBatchDeleted        = "Батч #%d успішно видалено"
	MsgErrorBatchLoading          = "Помилка завантаження імпортів: %v"
	MsgErrorOSBBLoading           = "Помилка завантаження даних ОСББ: %v"
	MsgOSBBNotConfigured          = "Організація ОСББ ще не налаштована"
	MsgOSBBAdminRequired          = "Створити ОСББ (потрібні права адміністратора)"
	MsgSuccessTitle               = "Успішно"
)

// Navigation Menu Titles
const (
	NavDashboard          = "📊 Головна"
	NavOSBB               = "🏠 Про ОСББ"
	NavOwners             = "👥 Власники"
	NavApartments         = "🏢 Квартири"
	NavOwnerships         = "📝 Частки власності"
	NavCharges            = "💰 Нарахування"
	NavPayments           = "💳 Платежі"
	NavContractorPayments = "🏢 Платежі контрагентів"
	NavCashFlow           = "💰 Рух коштів"
	NavExpenses           = "💸 Витрати"
	NavExpenseCategories  = "📂 Категорії витрат"
	NavContractors        = "👷 Підрядники"
	NavImportExport       = "📥 Імпорт/Експорт"
	NavSettings           = "⚙️ Налаштування"
	NavLogout             = "Вийти"
)

// Screen Titles & Headers
const (
	TitleApartments        = "Квартири"
	TitleOwners            = "Власники"
	TitleCharges           = "Нарахування"
	TitlePayments          = "Платежі"
	TitleExpenses          = "Витрати"
	TitleContractors       = "Підрядники"
	TitleExpenseCategories = "Категорії витрат"
)

// Filters and Search
const (
	FilterAll         = "Всі"
	FilterActive      = "Активні"
	FilterInactive    = "Неактивні"
	SearchPlaceholder = "🔍 Пошук..."
)

// Months
const (
	MonthJanuary   = "Січень"
	MonthFebruary  = "Лютий"
	MonthMarch     = "Березень"
	MonthApril     = "Квітень"
	MonthMay       = "Травень"
	MonthJune      = "Червень"
	MonthJuly      = "Липень"
	MonthAugust    = "Серпень"
	MonthSeptember = "Вересень"
	MonthOctober   = "Жовтень"
	MonthNovember  = "Листопад"
	MonthDecember  = "Грудень"
)

var (
	Months = []string{
		MonthJanuary,
		MonthFebruary,
		MonthMarch,
		MonthApril,
		MonthMay,
		MonthJune,
		MonthJuly,
		MonthAugust,
		MonthSeptember,
		MonthOctober,
		MonthNovember,
		MonthDecember,
	}
)

// Additional Labels & Messages
const (
	LabelRoles            = "Ролі: %v"
	LabelTemplate         = "Template"
	LabelApartmentNumber  = "квартиру №%s"
	TitleApartmentDetails = "Інформація про квартиру"
	TitleApartmentActions = "Дії з квартирою № %s"
	MsgApartmentDetails   = "Квартира: №%s\n" +
		"Поверх: %d\n" +
		"Під'їзд: %s\n\n" +
		"Площа: %.2f м²\n" +
		"Жила площа: %s\n" +
		"Кількість кімнат: %s\n" +
		"Кадастровий номер: %s\n" +
		"Статус: %s\n"
)

// Apartment Filters
const (
	FilterApartmentAll       = "Всі квартири"
	FilterApartmentEntrance1 = "Перший підїзд"
	FilterApartmentEntrance2 = "Другий підїзд"
	FilterApartmentEntrance3 = "Третій підїзд"
	FilterApartmentEntrance4 = "Четвертий підїзд"
	FilterApartmentEntrance5 = "П'ятий підїзд"
	FilterApartmentEntrance6 = "Шостий підїзд"
)

// Expense Filters
const (
	FilterExpenseAll         = "Всі витрати"
	FilterExpenseApproved    = "Затверджені витрати"
	FilterExpenseNotApproved = "Не затверджені витрати"
)

// Contractor Payment Filters
const (
	FilterContractorPaymentAll      = "Всі платежі"
	FilterContractorPaymentCurrent  = "Поточний місяць"
	FilterContractorPaymentPrevious = "Минулий місяць"
)

// Owner Filters
const (
	FilterOwnerAll          = "Всі власники"
	FilterOwnerWithTax      = "Тільки з ІПН"
	FilterOwnerNoTax        = "Без ІПН"
	FilterOwnerWithContacts = "Тільки з контактами"
	FilterOwnerNoContacts   = "Без контактів"
	FilterOwnerActive       = "Тільки активні"
	FilterOwnerInactive     = "Неактивні"
)

// Ownership Share Filters
const (
	FilterShareAll      = "Всі частки"
	FilterShareActive   = "Тільки активні зараз"
	FilterShareFull     = "Повна власність"
	FilterShareShared   = "Часткова власність"
	FilterShareRent     = "Оренда"
	FilterShareFinished = "Завершені"
)

// Contractor Filters
const (
	FilterContractorAll      = "Всі контрагенти"
	FilterContractorActive   = "Активні"
	FilterContractorInactive = "Неактивні"
)

// Charge Filters
const (
	FilterChargeAll         = "Всі нарахування"
	FilterChargeCurrent     = "Поточний місяць"
	FilterChargePrevious    = "Минулий місяць"
	FilterChargeMaintenance = "Утримання"
	FilterChargeUtility     = "Комунальні"
	FilterChargeRepair      = "Ремонт"
	FilterChargePenalty     = "Пеня"
	FilterChargeOther       = "Інше"
)

// Payment Filters
const (
	FilterPaymentAll         = "Всі платежі"
	FilterPaymentCurrent     = "Поточний місяць"
	FilterPaymentPrevious    = "Минулий місяць"
	FilterPaymentNotApproved = "Непідтверджені платежі"
	FilterPaymentApproved    = "Підтверджені платежі"
)

// Expense Category Filters
const (
	FilterCategoryAll      = "Всі категорії"
	FilterCategoryActive   = "Активні"
	FilterCategoryInactive = "Неактивні"
)

// General Labels
const (
	LabelYes               = "Так"
	LabelNo                = "Ні"
	LabelExpenseID         = "витрату #%d"
	TitleExpenseActions    = "Витрата #%d"
	LabelPeriod            = "Період:"
	LabelContractorID      = "Контрагент #%d"
	TitlePaymentActions    = "Платіж #%d"
	LabelPaymentID         = "платіж #%d"
	LabelOwnerName         = "власника %s"
	TitleOwnerActions      = "Дії з власником: %s"
	TitleOwnerDetails      = "Інформація про власника"
	LabelShareID           = "частку власності #%d"
	TitleShareActions      = "Дії з часткою #%d"
	TitleShareDetails      = "Деталі частки власності"
	LabelContractorName    = "контрагента '%s'"
	TitleContractorActions = "Дії з контрагентом: %s"
	LabelActiveNow         = "✅ Активна зараз"
	LabelFinished          = "❌ Завершена"
	LabelApartmentShort    = "Кв. %s"
	LabelShareFormat       = "%s (%.1f%%)"
	LabelDatePresent       = "теперішній час"
	LabelChargeID          = "нарахування #%d"
	TitleChargeActions     = "Дії з нарахуванням #%d"
	TitleChargeDetails     = "Деталі нарахування"
	LabelPayerID           = "ID: %d"
	TitleCategoryActions   = "Дії з категорією: %s"
	LabelCategoryName      = "категорію '%s'"

	// Import
	TitleImportDetails = "Деталі імпорту по місяцях"
	TitleBatchActions  = "Дії з батчем #%d"
	LabelBatchInfo     = "Батч #%d: %s (Всього: %d записів)"
	LabelNoData        = "Немає даних"
	LabelNoRecords     = "Немає записів для відображення"

	// Settings
	TitleUserProfile  = "Профіль користувача"
	TitleAppInfo      = "Про програму"
	LabelName         = "Ім'я: %s"
	LabelLogin        = "Логін: %s"
	LabelEmail        = "Email: %s"
	LabelAppName      = "OSBB Accounting System"
	LabelAppVersion   = "Версія: 1.0.0"
	LabelAppDeveloper = "Розробник: Google Deepmind Team"

	// Dashboard
	TitleFinanceOverview = "Огляд фінансів"
	LabelTotalCharges    = "Нараховано (всього)"
	LabelTotalPayments   = "Сплачено (всього)"
	LabelOutstandingDebt = "Заборгованість"
	TitlePaymentMethods  = "Методи оплати"
	TitleChargeTypes     = "Типи нарахувань"
	LabelMethodCash      = "Готівка"
	LabelMethodCard      = "Картка"
	LabelMethodBank      = "Банк"
	LabelMethodOther     = "Інше"

	// OSBB
	TitleOSBBInfo       = "Інформація про ОСББ"
	LabelOSBBName       = "Назва"
	LabelOSBBEDRPOU     = "ЄДРПОУ"
	LabelOSBBAddress    = "Юридична адреса"
	LabelOSBBChairman   = "Голова правління"
	LabelOSBBActualAddr = "Фактична адреса"
	LabelOSBBPhone      = "Телефон"
	LabelOSBBEmail      = "Email"
	LabelOSBBWebsite    = "Веб-сайт"
)

// Status
const (
	ActiveStatus   = "✅ Активний"
	InactiveStatus = "❌ Неактивний"
)
