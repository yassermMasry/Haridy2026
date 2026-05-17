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
	Name      string `gorm:"size:140;uniqueIndex;not null"`
	Code      string `gorm:"size:40;uniqueIndex;not null"`
	Address   string `gorm:"size:255"`
	IsActive  bool   `gorm:"not null;default:true;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Warehouse struct {
	ID        uint `gorm:"primaryKey"`
	BranchID  uint `gorm:"not null;index"`
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
	Name      string `gorm:"size:120;uniqueIndex;not null"`
	Items     []Item
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Item struct {
	ID                uint    `gorm:"primaryKey"`
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
	ID          uint `gorm:"primaryKey"`
	ItemID      uint `gorm:"not null;index"`
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
	Number      string `gorm:"size:40;uniqueIndex;not null"`
	UserID      uint   `gorm:"not null;index"`
	User        User
	CustomerID  *uint `gorm:"index"`
	Customer    *Customer
	BranchID    *uint   `gorm:"index"`
	WarehouseID *uint   `gorm:"index"`
	PaymentType string  `gorm:"size:20;not null;default:'cash';index"`
	Subtotal    float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Discount    float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Tax         float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Total       float64 `gorm:"type:numeric(14,2);not null;default:0"`
	PaidCash    float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Items       []SalesInvoiceItem
	CreatedAt   time.Time `gorm:"index"`
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
	Number      string `gorm:"size:40;uniqueIndex;not null"`
	SupplierID  uint   `gorm:"not null;index"`
	Supplier    Supplier
	BranchID    *uint `gorm:"index"`
	WarehouseID *uint `gorm:"index"`
	UserID      uint  `gorm:"not null;index"`
	User        User
	PaymentType string  `gorm:"size:20;not null;default:'cash';index"`
	Subtotal    float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Discount    float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Tax         float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Total       float64 `gorm:"type:numeric(14,2);not null;default:0"`
	PaidCash    float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Items       []PurchaseInvoiceItem
	CreatedAt   time.Time `gorm:"index"`
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
	Code      string      `gorm:"size:30;uniqueIndex;not null"`
	Name      string      `gorm:"size:120;not null;index"`
	Type      AccountType `gorm:"type:varchar(20);not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type JournalEntry struct {
	ID          uint   `gorm:"primaryKey"`
	Number      string `gorm:"size:40;uniqueIndex;not null"`
	Reference   string `gorm:"size:120;index"`
	Description string `gorm:"size:255"`
	Lines       []JournalEntryLine
	CreatedAt   time.Time `gorm:"index"`
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
	BranchID    *uint   `gorm:"index"`
	WarehouseID *uint   `gorm:"index"`
	Total       float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Reason      string  `gorm:"size:255"`
	UserID      uint    `gorm:"not null;index"`
	Items       []SalesReturnItem
	CreatedAt   time.Time `gorm:"index"`
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
	BranchID    *uint   `gorm:"index"`
	WarehouseID *uint   `gorm:"index"`
	Total       float64 `gorm:"type:numeric(14,2);not null;default:0"`
	Reason      string  `gorm:"size:255"`
	UserID      uint    `gorm:"not null;index"`
	Items       []PurchaseReturnItem
	CreatedAt   time.Time `gorm:"index"`
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
