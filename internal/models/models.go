package models

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleCashier Role = "cashier"
)

type User struct {
	ID              uint   `gorm:"primaryKey"`
	TenantID        *uint  `gorm:"index"`
	Username        string `gorm:"size:80;uniqueIndex;not null"`
	PasswordHash    string `gorm:"size:255;not null"`
	Role            Role   `gorm:"type:varchar(20);not null;index"`
	CurrentBranchID *uint  `gorm:"index"`
	CurrentBranch   *Branch
	Roles           []RBACRole `gorm:"many2many:user_roles;"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type Branch struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  *uint  `gorm:"index"`
	Name      string `gorm:"size:140;uniqueIndex;not null"`
	Code      string `gorm:"size:40;uniqueIndex;not null"`
	Address   string `gorm:"size:255"`
	IsActive  bool   `gorm:"not null;default:true;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Warehouse struct {
	ID        uint  `gorm:"primaryKey"`
	TenantID  *uint `gorm:"index"`
	BranchID  uint  `gorm:"not null;index"`
	Branch    Branch
	Name      string `gorm:"size:140;not null;index"`
	Code      string `gorm:"size:40;uniqueIndex;not null"`
	IsActive  bool   `gorm:"not null;default:true;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ItemCategory struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  *uint  `gorm:"index"`
	Name      string `gorm:"size:120;uniqueIndex;not null"`
	Items     []Item `gorm:"foreignKey:CategoryID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Item struct {
	ID                uint    `gorm:"primaryKey"`
	TenantID          *uint   `gorm:"index"`
	Name              string  `gorm:"size:160;not null;index"`
	Code              string  `gorm:"size:80;uniqueIndex;not null"`
	Barcode           string  `gorm:"size:120;index"`
	PurchasePrice     float64 `gorm:"type:numeric(14,2);not null;default:0"`
	SalePrice         float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Quantity          float64 `gorm:"type:numeric(14,3);not null;default:0;index"`
	MinimumStock      float64 `gorm:"type:numeric(14,3);not null;default:0"`
	CategoryID        *uint   `gorm:"index"`
	Category          *ItemCategory
	WarehouseBalances []ItemWarehouseBalance
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

type ItemWarehouseBalance struct {
	ID          uint `gorm:"primaryKey"`
	ItemID      uint `gorm:"not null;uniqueIndex:idx_item_warehouse"`
	Item        Item
	WarehouseID uint `gorm:"not null;uniqueIndex:idx_item_warehouse"`
	Warehouse   Warehouse
	Quantity    float64 `gorm:"type:numeric(14,3);not null;default:0;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WarehouseTransfer struct {
	ID              uint   `gorm:"primaryKey"`
	Number          string `gorm:"size:40;uniqueIndex;not null"`
	FromWarehouseID uint   `gorm:"not null;index"`
	FromWarehouse   Warehouse
	ToWarehouseID   uint `gorm:"not null;index"`
	ToWarehouse     Warehouse
	ItemID          uint `gorm:"not null;index"`
	Item            Item
	Quantity        float64 `gorm:"type:numeric(14,3);not null"`
	UserID          uint    `gorm:"not null;index"`
	User            User
	CreatedAt       time.Time `gorm:"index"`
}

type StockMovementType string

const (
	StockIn   StockMovementType = "IN"
	StockOut  StockMovementType = "OUT"
	StockSale StockMovementType = "SALE"
	StockBuy  StockMovementType = "PURCHASE"
)

type StockMovement struct {
	ID          uint  `gorm:"primaryKey"`
	TenantID    *uint `gorm:"index"`
	ItemID      uint  `gorm:"not null;index"`
	Item        Item
	BranchID    *uint             `gorm:"index"`
	WarehouseID *uint             `gorm:"index"`
	Type        StockMovementType `gorm:"type:varchar(20);not null;index"`
	Quantity    float64           `gorm:"type:numeric(14,3);not null"`
	Reference   string            `gorm:"size:120;index"`
	Notes       string            `gorm:"size:255"`
	PerformedBy *uint             `gorm:"index"`
	CreatedAt   time.Time         `gorm:"index"`
}

type SalesInvoice struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    *uint  `gorm:"index"`
	Number      string `gorm:"size:40;uniqueIndex;not null"`
	UserID      uint   `gorm:"not null;index"`
	User        User
	CustomerID  *uint `gorm:"index"`
	Customer    *Customer
	BranchID    *uint              `gorm:"index"`
	WarehouseID *uint              `gorm:"index"`
	PaymentType string             `gorm:"size:20;not null;default:'cash';index"`
	Subtotal    float64            `gorm:"type:numeric(14,2);not null;default:0"`
	Discount    float64            `gorm:"type:numeric(14,2);not null;default:0"`
	Tax         float64            `gorm:"type:numeric(14,2);not null;default:0"`
	Total       float64            `gorm:"type:numeric(14,2);not null;default:0"`
	PaidCash    float64            `gorm:"type:numeric(14,2);not null;default:0"`
	Items       []SalesInvoiceItem `gorm:"foreignKey:InvoiceID"`
	CreatedAt   time.Time          `gorm:"index"`
	UpdatedAt   time.Time
}

type SalesInvoiceItem struct {
	ID        uint `gorm:"primaryKey"`
	InvoiceID uint `gorm:"not null;index"`
	Invoice   SalesInvoice
	ItemID    uint `gorm:"not null;index"`
	Item      Item
	Quantity  float64 `gorm:"type:numeric(14,3);not null"`
	UnitPrice float64 `gorm:"type:numeric(14,2);not null"`
	Total     float64 `gorm:"type:numeric(14,2);not null"`
}

type Treasury struct {
	ID           uint    `gorm:"primaryKey"`
	TenantID     *uint   `gorm:"index"`
	BranchID     *uint   `gorm:"index"`
	Name         string  `gorm:"size:120;uniqueIndex;not null"`
	Balance      float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Transactions []TreasuryTransaction
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TreasuryTransactionType string

const (
	TreasurySale        TreasuryTransactionType = "SALE"
	TreasuryReceive     TreasuryTransactionType = "RECEIVE"
	TreasuryExpense     TreasuryTransactionType = "EXPENSE"
	TreasuryPurchase    TreasuryTransactionType = "PURCHASE"
	TreasurySupplierPay TreasuryTransactionType = "SUPPLIER_PAYMENT"
)

type TreasuryTransaction struct {
	ID          uint `gorm:"primaryKey"`
	TreasuryID  uint `gorm:"not null;index"`
	Treasury    Treasury
	Type        TreasuryTransactionType `gorm:"type:varchar(20);not null;index"`
	Amount      float64                 `gorm:"type:numeric(14,2);not null"`
	Reference   string                  `gorm:"size:120;index"`
	Description string                  `gorm:"size:255"`
	UserID      *uint                   `gorm:"index"`
	User        *User
	CreatedAt   time.Time `gorm:"index"`
}

type Customer struct {
	ID          uint    `gorm:"primaryKey"`
	TenantID    *uint   `gorm:"index"`
	Name        string  `gorm:"size:160;not null;index"`
	Phone       string  `gorm:"size:40;index"`
	Address     string  `gorm:"size:255"`
	Notes       string  `gorm:"size:500"`
	CreditLimit float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Balance     float64 `gorm:"type:numeric(14,2);not null;default:0;index"`
	BranchID    *uint   `gorm:"index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type CustomerTransactionType string

const (
	CustomerSale    CustomerTransactionType = "SALE"
	CustomerPayment CustomerTransactionType = "PAYMENT"
)

type CustomerTransaction struct {
	ID          uint `gorm:"primaryKey"`
	CustomerID  uint `gorm:"not null;index"`
	Customer    Customer
	Type        CustomerTransactionType `gorm:"type:varchar(20);not null;index"`
	Debit       float64                 `gorm:"type:numeric(14,2);not null;default:0"`
	Credit      float64                 `gorm:"type:numeric(14,2);not null;default:0"`
	Reference   string                  `gorm:"size:120;index"`
	Description string                  `gorm:"size:255"`
	UserID      *uint                   `gorm:"index"`
	CreatedAt   time.Time               `gorm:"index"`
}

type Supplier struct {
	ID        uint    `gorm:"primaryKey"`
	TenantID  *uint   `gorm:"index"`
	Name      string  `gorm:"size:160;not null;index"`
	Phone     string  `gorm:"size:40;index"`
	Address   string  `gorm:"size:255"`
	Notes     string  `gorm:"size:500"`
	Balance   float64 `gorm:"type:numeric(14,2);not null;default:0;index"`
	BranchID  *uint   `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type SupplierTransactionType string

const (
	SupplierPurchase SupplierTransactionType = "PURCHASE"
	SupplierPayment  SupplierTransactionType = "PAYMENT"
)

type SupplierTransaction struct {
	ID          uint `gorm:"primaryKey"`
	SupplierID  uint `gorm:"not null;index"`
	Supplier    Supplier
	Type        SupplierTransactionType `gorm:"type:varchar(20);not null;index"`
	Debit       float64                 `gorm:"type:numeric(14,2);not null;default:0"`
	Credit      float64                 `gorm:"type:numeric(14,2);not null;default:0"`
	Reference   string                  `gorm:"size:120;index"`
	Description string                  `gorm:"size:255"`
	UserID      *uint                   `gorm:"index"`
	CreatedAt   time.Time               `gorm:"index"`
}

type PurchaseInvoice struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    *uint  `gorm:"index"`
	Number      string `gorm:"size:40;uniqueIndex;not null"`
	SupplierID  uint   `gorm:"not null;index"`
	Supplier    Supplier
	BranchID    *uint `gorm:"index"`
	WarehouseID *uint `gorm:"index"`
	UserID      uint  `gorm:"not null;index"`
	User        User
	PaymentType string                `gorm:"size:20;not null;default:'cash';index"`
	Subtotal    float64               `gorm:"type:numeric(14,2);not null;default:0"`
	Discount    float64               `gorm:"type:numeric(14,2);not null;default:0"`
	Tax         float64               `gorm:"type:numeric(14,2);not null;default:0"`
	Total       float64               `gorm:"type:numeric(14,2);not null;default:0"`
	PaidCash    float64               `gorm:"type:numeric(14,2);not null;default:0"`
	Items       []PurchaseInvoiceItem `gorm:"foreignKey:InvoiceID"`
	CreatedAt   time.Time             `gorm:"index"`
	UpdatedAt   time.Time
}

type PurchaseInvoiceItem struct {
	ID        uint `gorm:"primaryKey"`
	InvoiceID uint `gorm:"not null;index"`
	Invoice   PurchaseInvoice
	ItemID    uint `gorm:"not null;index"`
	Item      Item
	Quantity  float64 `gorm:"type:numeric(14,3);not null"`
	UnitCost  float64 `gorm:"type:numeric(14,2);not null"`
	Total     float64 `gorm:"type:numeric(14,2);not null"`
}

type AccountType string

const (
	AccountAsset     AccountType = "ASSET"
	AccountRevenue   AccountType = "REVENUE"
	AccountExpense   AccountType = "EXPENSE"
	AccountLiability AccountType = "LIABILITY"
)

type ChartOfAccount struct {
	ID        uint        `gorm:"primaryKey"`
	TenantID  *uint       `gorm:"index"`
	Code      string      `gorm:"size:30;uniqueIndex;not null"`
	Name      string      `gorm:"size:120;not null;index"`
	Type      AccountType `gorm:"type:varchar(20);not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type JournalEntry struct {
	ID          uint               `gorm:"primaryKey"`
	TenantID    *uint              `gorm:"index"`
	Number      string             `gorm:"size:40;uniqueIndex;not null"`
	Reference   string             `gorm:"size:120;index"`
	Description string             `gorm:"size:255"`
	Lines       []JournalEntryLine `gorm:"foreignKey:EntryID"`
	CreatedAt   time.Time          `gorm:"index"`
}

type JournalEntryLine struct {
	ID        uint `gorm:"primaryKey"`
	EntryID   uint `gorm:"not null;index"`
	Entry     JournalEntry
	AccountID uint `gorm:"not null;index"`
	Account   ChartOfAccount
	Debit     float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Credit    float64 `gorm:"type:numeric(14,2);not null;default:0"`
}

type AuditLog struct {
	ID        uint      `gorm:"primaryKey"`
	TenantID  *uint     `gorm:"index"`
	UserID    *uint     `gorm:"index"`
	Action    string    `gorm:"size:80;not null;index"`
	Entity    string    `gorm:"size:80;not null;index"`
	EntityID  uint      `gorm:"index"`
	Details   string    `gorm:"size:1000"`
	CreatedAt time.Time `gorm:"index"`
}

type RBACRole struct {
	ID          uint         `gorm:"primaryKey"`
	Name        string       `gorm:"size:80;uniqueIndex;not null"`
	Description string       `gorm:"size:255"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
	Users       []User       `gorm:"many2many:user_roles;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Permission struct {
	ID          uint   `gorm:"primaryKey"`
	Code        string `gorm:"size:80;uniqueIndex;not null"`
	Description string `gorm:"size:255"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RolePermission struct {
	RBACRoleID   uint `gorm:"primaryKey;column:rbac_role_id"`
	PermissionID uint `gorm:"primaryKey"`
}

type UserRole struct {
	UserID     uint `gorm:"primaryKey"`
	RBACRoleID uint `gorm:"primaryKey;column:rbac_role_id"`
}

type SalesReturn struct {
	ID          uint   `gorm:"primaryKey"`
	Number      string `gorm:"size:40;uniqueIndex;not null"`
	InvoiceID   uint   `gorm:"not null;index"`
	Invoice     SalesInvoice
	BranchID    *uint             `gorm:"index"`
	WarehouseID *uint             `gorm:"index"`
	Total       float64           `gorm:"type:numeric(14,2);not null;default:0"`
	Reason      string            `gorm:"size:255"`
	UserID      uint              `gorm:"not null;index"`
	Items       []SalesReturnItem `gorm:"foreignKey:ReturnID"`
	CreatedAt   time.Time         `gorm:"index"`
}

type SalesReturnItem struct {
	ID        uint `gorm:"primaryKey"`
	ReturnID  uint `gorm:"not null;index"`
	ItemID    uint `gorm:"not null;index"`
	Item      Item
	Quantity  float64 `gorm:"type:numeric(14,3);not null"`
	UnitPrice float64 `gorm:"type:numeric(14,2);not null"`
	Total     float64 `gorm:"type:numeric(14,2);not null"`
}

type PurchaseReturn struct {
	ID          uint   `gorm:"primaryKey"`
	Number      string `gorm:"size:40;uniqueIndex;not null"`
	InvoiceID   uint   `gorm:"not null;index"`
	Invoice     PurchaseInvoice
	BranchID    *uint                `gorm:"index"`
	WarehouseID *uint                `gorm:"index"`
	Total       float64              `gorm:"type:numeric(14,2);not null;default:0"`
	Reason      string               `gorm:"size:255"`
	UserID      uint                 `gorm:"not null;index"`
	Items       []PurchaseReturnItem `gorm:"foreignKey:ReturnID"`
	CreatedAt   time.Time            `gorm:"index"`
}

type PurchaseReturnItem struct {
	ID       uint `gorm:"primaryKey"`
	ReturnID uint `gorm:"not null;index"`
	ItemID   uint `gorm:"not null;index"`
	Item     Item
	Quantity float64 `gorm:"type:numeric(14,3);not null"`
	UnitCost float64 `gorm:"type:numeric(14,2);not null"`
	Total    float64 `gorm:"type:numeric(14,2);not null"`
}

type ReceiptVoucher struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    *uint  `gorm:"index"`
	Number      string `gorm:"size:40;uniqueIndex;not null"`
	BranchID    *uint  `gorm:"index"`
	CustomerID  *uint  `gorm:"index"`
	Customer    *Customer
	Amount      float64   `gorm:"type:numeric(14,2);not null"`
	Description string    `gorm:"size:255"`
	UserID      uint      `gorm:"not null;index"`
	CreatedAt   time.Time `gorm:"index"`
}

type PaymentVoucher struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    *uint  `gorm:"index"`
	Number      string `gorm:"size:40;uniqueIndex;not null"`
	BranchID    *uint  `gorm:"index"`
	SupplierID  *uint  `gorm:"index"`
	Supplier    *Supplier
	Amount      float64   `gorm:"type:numeric(14,2);not null"`
	Description string    `gorm:"size:255"`
	UserID      uint      `gorm:"not null;index"`
	CreatedAt   time.Time `gorm:"index"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey"`
	BranchID  *uint     `gorm:"index"`
	Type      string    `gorm:"size:40;not null;index"`
	Title     string    `gorm:"size:160;not null"`
	Message   string    `gorm:"size:500"`
	IsRead    bool      `gorm:"not null;default:false;index"`
	UserID    *uint     `gorm:"index"`
	CreatedAt time.Time `gorm:"index"`
}

type LoginAttempt struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"size:80;index"`
	IP        string    `gorm:"size:64;index"`
	Success   bool      `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"index"`
}

type SecurityEvent struct {
	ID        uint      `gorm:"primaryKey"`
	TenantID  *uint     `gorm:"index"`
	UserID    *uint     `gorm:"index"`
	Type      string    `gorm:"size:80;not null;index"`
	Severity  string    `gorm:"size:20;not null;default:'info';index"`
	IP        string    `gorm:"size:64;index"`
	UserAgent string    `gorm:"size:255"`
	Details   string    `gorm:"size:1000"`
	CreatedAt time.Time `gorm:"index"`
}

type ReconciliationRun struct {
	ID          uint       `gorm:"primaryKey"`
	TenantID    *uint      `gorm:"index"`
	Scope       string     `gorm:"size:80;not null;index"`
	Status      string     `gorm:"size:30;not null;index"`
	Findings    string     `gorm:"type:text"`
	StartedAt   time.Time  `gorm:"index"`
	CompletedAt *time.Time `gorm:"index"`
}

type BackupVerification struct {
	ID          uint       `gorm:"primaryKey"`
	BackupName  string     `gorm:"size:255;not null;index"`
	StorageURI  string     `gorm:"size:500"`
	Checksum    string     `gorm:"size:128"`
	Encrypted   bool       `gorm:"not null;default:false"`
	Status      string     `gorm:"size:30;not null;index"`
	Details     string     `gorm:"size:1000"`
	CompletedAt *time.Time `gorm:"index"`
	CreatedAt   time.Time  `gorm:"index"`
}

type ReportJob struct {
	ID          uint       `gorm:"primaryKey"`
	TenantID    *uint      `gorm:"index"`
	Type        string     `gorm:"size:80;not null;index"`
	Status      string     `gorm:"size:30;not null;default:'queued';index"`
	Parameters  string     `gorm:"type:text"`
	ResultPath  string     `gorm:"size:500"`
	Error       string     `gorm:"size:1000"`
	RequestedBy *uint      `gorm:"index"`
	StartedAt   *time.Time `gorm:"index"`
	CompletedAt *time.Time `gorm:"index"`
	CreatedAt   time.Time  `gorm:"index"`
	UpdatedAt   time.Time
}

type Tenant struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:160;not null;index"`
	Slug      string `gorm:"size:80;uniqueIndex;not null"`
	Domain    string `gorm:"size:180;index"`
	Subdomain string `gorm:"size:80;uniqueIndex"`
	Status    string `gorm:"size:30;not null;default:'trial';index"`
	Settings  CompanySetting
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Plan struct {
	ID           uint    `gorm:"primaryKey"`
	Code         string  `gorm:"size:50;uniqueIndex;not null"`
	Name         string  `gorm:"size:120;not null"`
	PriceMonthly float64 `gorm:"type:numeric(14,2);not null;default:0"`
	MaxUsers     int     `gorm:"not null;default:5"`
	MaxBranches  int     `gorm:"not null;default:1"`
	Features     string  `gorm:"type:text"`
	IsActive     bool    `gorm:"not null;default:true;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Subscription struct {
	ID        uint `gorm:"primaryKey"`
	TenantID  uint `gorm:"not null;index"`
	Tenant    Tenant
	PlanID    uint `gorm:"not null;index"`
	Plan      Plan
	Status    string `gorm:"size:30;not null;index"`
	StartsAt  time.Time
	EndsAt    *time.Time `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TenantUser struct {
	ID        uint `gorm:"primaryKey"`
	TenantID  uint `gorm:"not null;uniqueIndex:idx_tenant_user"`
	UserID    uint `gorm:"not null;uniqueIndex:idx_tenant_user"`
	IsOwner   bool `gorm:"not null;default:false"`
	CreatedAt time.Time
}

type CompanySetting struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null;uniqueIndex"`
	LegalName string `gorm:"size:180"`
	TaxNumber string `gorm:"size:80"`
	LogoURL   string `gorm:"size:255"`
	Currency  string `gorm:"size:10;not null;default:'EGP'"`
	TimeZone  string `gorm:"size:80;not null;default:'Africa/Cairo'"`
	Address   string `gorm:"size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FiscalYear struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null;index"`
	Name      string `gorm:"size:80;not null"`
	StartDate time.Time
	EndDate   time.Time
	IsClosed  bool `gorm:"not null;default:false;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FinancialPeriod struct {
	ID           uint   `gorm:"primaryKey"`
	TenantID     uint   `gorm:"not null;index"`
	FiscalYearID uint   `gorm:"not null;index"`
	Name         string `gorm:"size:80;not null"`
	StartDate    time.Time
	EndDate      time.Time
	IsClosed     bool `gorm:"not null;default:false;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type OpeningBalance struct {
	ID           uint    `gorm:"primaryKey"`
	TenantID     uint    `gorm:"not null;index"`
	AccountID    uint    `gorm:"not null;index"`
	Debit        float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Credit       float64 `gorm:"type:numeric(14,2);not null;default:0"`
	FiscalYearID uint    `gorm:"not null;index"`
	CreatedAt    time.Time
}

type ClosingEntry struct {
	ID             uint `gorm:"primaryKey"`
	TenantID       uint `gorm:"not null;index"`
	FiscalYearID   uint `gorm:"not null;index"`
	JournalEntryID uint `gorm:"not null;index"`
	CreatedAt      time.Time
}

type EInvoice struct {
	ID        uint    `gorm:"primaryKey"`
	TenantID  uint    `gorm:"not null;index"`
	InvoiceID uint    `gorm:"not null;index"`
	UUID      string  `gorm:"size:80;uniqueIndex;not null"`
	QRPayload string  `gorm:"type:text"`
	XMLBody   string  `gorm:"type:text"`
	VATAmount float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Status    string  `gorm:"size:30;not null;default:'draft';index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ApprovalWorkflow struct {
	ID        uint           `gorm:"primaryKey"`
	TenantID  uint           `gorm:"not null;index"`
	Name      string         `gorm:"size:120;not null"`
	Module    string         `gorm:"size:50;not null;index"`
	IsActive  bool           `gorm:"not null;default:true;index"`
	Steps     []ApprovalStep `gorm:"foreignKey:WorkflowID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ApprovalStep struct {
	ID         uint   `gorm:"primaryKey"`
	WorkflowID uint   `gorm:"not null;index"`
	StepOrder  int    `gorm:"not null"`
	RoleID     *uint  `gorm:"index"`
	UserID     *uint  `gorm:"index"`
	Name       string `gorm:"size:120;not null"`
	CreatedAt  time.Time
}

type ApprovalLog struct {
	ID         uint      `gorm:"primaryKey"`
	TenantID   uint      `gorm:"not null;index"`
	WorkflowID uint      `gorm:"not null;index"`
	EntityType string    `gorm:"size:80;not null;index"`
	EntityID   uint      `gorm:"not null;index"`
	StepID     *uint     `gorm:"index"`
	Status     string    `gorm:"size:30;not null;index"`
	Comment    string    `gorm:"size:500"`
	UserID     uint      `gorm:"not null;index"`
	CreatedAt  time.Time `gorm:"index"`
}

type ItemSerial struct {
	ID           uint   `gorm:"primaryKey"`
	TenantID     uint   `gorm:"not null;index"`
	ItemID       uint   `gorm:"not null;index"`
	SerialNumber string `gorm:"size:120;uniqueIndex;not null"`
	Status       string `gorm:"size:30;not null;default:'available';index"`
	CreatedAt    time.Time
}

type ItemBatch struct {
	ID         uint       `gorm:"primaryKey"`
	TenantID   uint       `gorm:"not null;index"`
	ItemID     uint       `gorm:"not null;index"`
	BatchNo    string     `gorm:"size:120;not null;index"`
	ExpiryDate *time.Time `gorm:"index"`
	Quantity   float64    `gorm:"type:numeric(14,3);not null;default:0"`
	UnitCost   float64    `gorm:"type:numeric(14,2);not null;default:0"`
	CreatedAt  time.Time
}

type StockValuationLayer struct {
	ID                uint      `gorm:"primaryKey"`
	TenantID          uint      `gorm:"not null;index"`
	ItemID            uint      `gorm:"not null;index"`
	Method            string    `gorm:"size:20;not null;index"`
	Quantity          float64   `gorm:"type:numeric(14,3);not null"`
	UnitCost          float64   `gorm:"type:numeric(14,2);not null"`
	RemainingQuantity float64   `gorm:"type:numeric(14,3);not null"`
	CreatedAt         time.Time `gorm:"index"`
}

type CRMActivity struct {
	ID         uint       `gorm:"primaryKey"`
	TenantID   uint       `gorm:"not null;index"`
	CustomerID uint       `gorm:"not null;index"`
	Type       string     `gorm:"size:40;not null;index"`
	Subject    string     `gorm:"size:160;not null"`
	Notes      string     `gorm:"size:1000"`
	DueAt      *time.Time `gorm:"index"`
	DoneAt     *time.Time `gorm:"index"`
	UserID     *uint      `gorm:"index"`
	CreatedAt  time.Time
}

type SalesPipeline struct {
	ID        uint                 `gorm:"primaryKey"`
	TenantID  uint                 `gorm:"not null;index"`
	Name      string               `gorm:"size:120;not null"`
	Stages    []SalesPipelineStage `gorm:"foreignKey:PipelineID"`
	CreatedAt time.Time
}

type SalesPipelineStage struct {
	ID         uint   `gorm:"primaryKey"`
	PipelineID uint   `gorm:"not null;index"`
	Name       string `gorm:"size:120;not null"`
	SortOrder  int    `gorm:"not null;default:0"`
}

type Deal struct {
	ID              uint       `gorm:"primaryKey"`
	TenantID        uint       `gorm:"not null;index"`
	CustomerID      uint       `gorm:"not null;index"`
	StageID         uint       `gorm:"not null;index"`
	Title           string     `gorm:"size:160;not null"`
	Value           float64    `gorm:"type:numeric(14,2);not null;default:0"`
	Status          string     `gorm:"size:30;not null;default:'open';index"`
	ExpectedCloseAt *time.Time `gorm:"index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Employee struct {
	ID         uint    `gorm:"primaryKey"`
	TenantID   uint    `gorm:"not null;index"`
	BranchID   *uint   `gorm:"index"`
	Name       string  `gorm:"size:160;not null;index"`
	Email      string  `gorm:"size:160;index"`
	Phone      string  `gorm:"size:40;index"`
	JobTitle   string  `gorm:"size:120"`
	BaseSalary float64 `gorm:"type:numeric(14,2);not null;default:0"`
	IsActive   bool    `gorm:"not null;default:true;index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type AttendanceRecord struct {
	ID         uint      `gorm:"primaryKey"`
	TenantID   uint      `gorm:"not null;index"`
	EmployeeID uint      `gorm:"not null;index"`
	WorkDate   time.Time `gorm:"index"`
	CheckIn    *time.Time
	CheckOut   *time.Time
	Status     string `gorm:"size:30;not null;default:'present';index"`
	CreatedAt  time.Time
}

type PayrollRun struct {
	ID         uint    `gorm:"primaryKey"`
	TenantID   uint    `gorm:"not null;index"`
	PeriodName string  `gorm:"size:80;not null"`
	GrossTotal float64 `gorm:"type:numeric(14,2);not null;default:0"`
	NetTotal   float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Status     string  `gorm:"size:30;not null;default:'draft';index"`
	CreatedAt  time.Time
}

type MobileDevice struct {
	ID         uint   `gorm:"primaryKey"`
	TenantID   uint   `gorm:"not null;index"`
	UserID     uint   `gorm:"not null;index"`
	Platform   string `gorm:"size:30;not null"`
	PushToken  string `gorm:"size:255;index"`
	LastSeenAt *time.Time
	CreatedAt  time.Time
}

type RefreshToken struct {
	ID        uint       `gorm:"primaryKey"`
	TenantID  *uint      `gorm:"index"`
	UserID    uint       `gorm:"not null;index"`
	TokenHash string     `gorm:"size:120;uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"index"`
	RevokedAt *time.Time `gorm:"index"`
	CreatedAt time.Time
}

type LicenseKey struct {
	ID            uint       `gorm:"primaryKey"`
	TenantID      *uint      `gorm:"index"`
	Key           string     `gorm:"size:120;uniqueIndex"`
	KeyHash       string     `gorm:"size:120;uniqueIndex"`
	PlanCode      string     `gorm:"size:50;not null;default:'yearly';index"`
	MaxOperations int64      `gorm:"not null;default:250"`
	Status        string     `gorm:"size:30;not null;default:'active';index"`
	ExpiresAt     *time.Time `gorm:"index"`
	UsedAt        *time.Time `gorm:"index"`
	CreatedAt     time.Time
}

type SystemUpdate struct {
	ID        uint   `gorm:"primaryKey"`
	Version   string `gorm:"size:40;uniqueIndex;not null"`
	Notes     string `gorm:"type:text"`
	Status    string `gorm:"size:30;not null;default:'available';index"`
	CreatedAt time.Time
}
