package entity

import "time"

// ImportedCashFlowRecord represents a record from the Cash Flow import file (311 account).
type ImportedCashFlowRecord struct {
	ID             int64
	ImportBatchID  int64
	Date           time.Time
	ContractorName string
	OperationType  string // "debit" or "credit"
	Amount         float64
	CategoryCode   string // e.g., "313", "63", "641"
	Description    string
	IsMigrated     bool
	CreatedAt      time.Time
}

const (
	CashFlowOperationDebit  = "debit"
	CashFlowOperationCredit = "credit"
)
