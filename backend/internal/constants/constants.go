package constants

// Invoice Types
const (
	InvoiceTypeCash   = "CASH"
	InvoiceTypeCredit = "CREDIT"
	InvoiceTypeDebit  = "DEBIT"
)

// Payment Status
const (
	PaymentStatusPending = "PENDING"
	PaymentStatusPartial = "PARTIAL"
	PaymentStatusPaid    = "PAID"
)

// Payment Methods
const (
	PaymentMethodCash         = "CASH"
	PaymentMethodBankTransfer = "BANK_TRANSFER"
	PaymentMethodCheque       = "CHEQUE"
	PaymentMethodUPI          = "UPI"
	PaymentMethodCard         = "CARD"
)

// Default Values
const (
	DefaultUnit           = "pcs"
	DefaultDueDays        = 30
	MinPasswordLength     = 6
	DefaultPageLimit      = 10
	DefaultGSTRateZero    = 0
	DefaultGSTRateFive    = 5
	DefaultGSTRateTwelve  = 12
	DefaultGSTRateEighteen = 18
	DefaultGSTRateTwentyEight = 28
)

// JWT
const (
	JWTExpirationHours = 24
	BearerPrefix       = "Bearer "
	BearerPrefixLength = 7
)

// HTTP Headers
const (
	AuthorizationHeader = "Authorization"
)

// Valid invoice types slice
var ValidInvoiceTypes = []string{
	InvoiceTypeCash,
	InvoiceTypeCredit,
	InvoiceTypeDebit,
}

// Valid payment methods slice
var ValidPaymentMethods = []string{
	PaymentMethodCash,
	PaymentMethodBankTransfer,
	PaymentMethodCheque,
	PaymentMethodUPI,
	PaymentMethodCard,
}

// Valid payment statuses slice
var ValidPaymentStatuses = []string{
	PaymentStatusPending,
	PaymentStatusPartial,
	PaymentStatusPaid,
}
