package services

import (
	"errors"
	"math"
	"strings"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ItemService struct{ db *gorm.DB }

type ItemList struct {
	Items      []models.Item
	Page       int
	TotalPages int
	Query      string
}

type SaleItemOption struct {
	Item      models.Item
	Quantity  float64
	Warehouse string
}

const WeakMarginPercent = 10

type PriceStatus string

const (
	PriceStatusNoPrice PriceStatus = "no_price"
	PriceStatusLoss    PriceStatus = "loss"
	PriceStatusWeak    PriceStatus = "weak"
	PriceStatusGood    PriceStatus = "good"
)

type PriceListRow struct {
	Item               models.Item
	AverageCost        *float64
	MarginPercent      float64
	WarehouseName      string
	Quantity           float64
	Status             PriceStatus
	StatusLabel        string
	StatusClass        string
	IsBelowCost        bool
	UpdateWarningLabel string
}

type PriceListStats struct {
	NoPrice int
	Loss    int
	Weak    int
	Good    int
}

type PriceList struct {
	Rows        []PriceListRow
	Stats       PriceListStats
	Query       string
	Filter      string
	WarehouseID uint
}

func NewItemService(db *gorm.DB) *ItemService { return &ItemService{db: db} }

func (s *ItemService) Categories(tenantID ...uint) []models.ItemCategory {
	var categories []models.ItemCategory
	query := s.db.Order("name asc")
	if len(tenantID) > 0 && tenantID[0] > 0 {
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantID[0])
	}
	query.Find(&categories)
	return categories
}

func (s *ItemService) Warehouses(tenantID uint) []models.Warehouse {
	var warehouses []models.Warehouse
	query := s.db.Order("name asc")
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantID)
	}
	query.Find(&warehouses)
	return warehouses
}

func (s *ItemService) SaleItems(warehouseID uint) []SaleItemOption {
	if warehouseID > 0 {
		var balances []models.ItemWarehouseBalance
		s.db.Preload("Item").Preload("Warehouse").
			Where("warehouse_id = ? AND quantity > 0", warehouseID).
			Find(&balances)
		options := make([]SaleItemOption, 0, len(balances))
		for _, balance := range balances {
			options = append(options, SaleItemOption{Item: balance.Item, Quantity: balance.Quantity, Warehouse: balance.Warehouse.Name})
		}
		return options
	}
	var items []models.Item
	s.db.Order("name asc").Find(&items)
	s.applyWarehouseQuantities(items)
	options := make([]SaleItemOption, 0, len(items))
	for _, item := range items {
		options = append(options, SaleItemOption{Item: item, Quantity: item.Quantity})
	}
	return options
}

func (s *ItemService) List(query string, page int) ItemList {
	if page < 1 {
		page = 1
	}
	limit := 12
	var items []models.Item
	var total int64
	q := s.db.Model(&models.Item{}).Preload("Category")
	if query != "" {
		like := "%" + query + "%"
		q = q.Where("name ILIKE ? OR code ILIKE ? OR barcode ILIKE ?", like, like, like)
	}
	q.Count(&total)
	q.Order("created_at desc").Limit(limit).Offset((page - 1) * limit).Find(&items)
	s.applyWarehouseQuantities(items)
	return ItemList{Items: items, Page: page, TotalPages: int(math.Ceil(float64(total) / float64(limit))), Query: query}
}

func (s *ItemService) Find(id uint) (*models.Item, error) {
	var item models.Item
	if err := s.db.Preload("Category").First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *ItemService) Save(item *models.Item) error {
	item.Name = strings.TrimSpace(item.Name)
	item.Code = strings.TrimSpace(item.Code)
	item.Barcode = strings.TrimSpace(item.Barcode)
	if item.Name == "" {
		return errors.New("اسم الصنف مطلوب")
	}
	if item.Code == "" {
		return errors.New("كود الصنف مطلوب")
	}
	if item.SalePrice < 0 || item.PurchasePrice < 0 || item.MinimumStock < 0 {
		return errors.New("الأسعار والحد الأدنى لا تكون بالسالب")
	}
	var count int64
	codeQuery := s.db.Model(&models.Item{}).Where("code = ?", item.Code)
	if item.ID > 0 {
		codeQuery = codeQuery.Where("id <> ?", item.ID)
	}
	if err := codeQuery.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("كود الصنف مستخدم من قبل")
	}
	if item.Barcode != "" {
		barcodeQuery := s.db.Model(&models.Item{}).Where("barcode = ?", item.Barcode)
		if item.ID > 0 {
			barcodeQuery = barcodeQuery.Where("id <> ?", item.ID)
		}
		if err := barcodeQuery.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("الباركود مستخدم من قبل")
		}
	}
	if item.ID == 0 {
		return s.db.Create(item).Error
	}
	return s.db.Model(&models.Item{}).Where("id = ?", item.ID).Updates(map[string]any{
		"name":           item.Name,
		"code":           item.Code,
		"barcode":        item.Barcode,
		"purchase_price": item.PurchasePrice,
		"sale_price":     item.SalePrice,
		"minimum_stock":  item.MinimumStock,
		"category_id":    item.CategoryID,
	}).Error
}

func (s *ItemService) SaveWithCategory(item *models.Item, newCategoryName string) error {
	if strings.TrimSpace(newCategoryName) == "" && item.CategoryID == nil {
		return errors.New("التصنيف مطلوب")
	}
	return s.Save(item)
}

func (s *ItemService) CreateWithOpeningBalance(item *models.Item, warehouseID uint) error {
	if item.Quantity > 0 && warehouseID == 0 {
		return errors.New("يجب اختيار المخزن عند إدخال كمية افتتاحية")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := (&ItemService{db: tx}).SaveWithCategory(item, ""); err != nil {
			return err
		}
		if item.Quantity <= 0 {
			return nil
		}
		var balance models.ItemWarehouseBalance
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("item_id = ? AND warehouse_id = ?", item.ID, warehouseID).
			FirstOrCreate(&balance, models.ItemWarehouseBalance{ItemID: item.ID, WarehouseID: warehouseID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&balance).Update("quantity", gorm.Expr("quantity + ?", item.Quantity)).Error; err != nil {
			return err
		}
		movement := models.StockMovement{
			TenantID:    item.TenantID,
			ItemID:      item.ID,
			WarehouseID: &warehouseID,
			Type:        models.StockOpen,
			Quantity:    item.Quantity,
			Reference:   "رصيد افتتاحي",
			Notes:       "رصيد افتتاحي",
			CreatedAt:   time.Now(),
		}
		return tx.Create(&movement).Error
	})
}

func (s *ItemService) Delete(id uint) error {
	if id == 0 {
		return errors.New("invalid item")
	}
	used, err := s.itemHasUsage(id)
	if err != nil {
		return err
	}
	if used {
		return errors.New("لا يمكن حذف الصنف لأنه مستخدم في عمليات سابقة")
	}
	return s.db.Delete(&models.Item{}, id).Error
}

func (s *ItemService) itemHasUsage(id uint) (bool, error) {
	checks := []struct {
		model any
		where string
		args  []any
	}{
		{model: &models.SalesInvoiceItem{}, where: "item_id = ?", args: []any{id}},
		{model: &models.PurchaseInvoiceItem{}, where: "item_id = ?", args: []any{id}},
		{model: &models.StockMovement{}, where: "item_id = ?", args: []any{id}},
		{model: &models.SalesReturnItem{}, where: "item_id = ?", args: []any{id}},
		{model: &models.PurchaseReturnItem{}, where: "item_id = ?", args: []any{id}},
		{model: &models.ItemWarehouseBalance{}, where: "item_id = ? AND quantity > ?", args: []any{id, 0}},
	}
	for _, check := range checks {
		var count int64
		if err := s.db.Model(check.model).Where(check.where, check.args...).Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}

func (s *ItemService) Alerts() []models.Item {
	var items []models.Item
	s.db.Where("quantity <= minimum_stock").Order("quantity asc").Limit(10).Find(&items)
	return items
}

func (s *ItemService) applyWarehouseQuantities(items []models.Item) {
	if len(items) == 0 {
		return
	}
	ids := make([]uint, 0, len(items))
	index := map[uint]int{}
	for i := range items {
		ids = append(ids, items[i].ID)
		index[items[i].ID] = i
	}
	var rows []struct {
		ItemID uint
		Total  float64
	}
	if err := s.db.Model(&models.ItemWarehouseBalance{}).
		Select("item_id, COALESCE(SUM(quantity),0) AS total").
		Where("item_id IN ?", ids).
		Group("item_id").
		Scan(&rows).Error; err != nil {
		return
	}
	for _, row := range rows {
		if i, ok := index[row.ItemID]; ok {
			items[i].Quantity = row.Total
		}
	}
}

func (s *ItemService) PriceList(query, filter string, warehouseID ...uint) PriceList {
	query = strings.TrimSpace(query)
	filter = strings.TrimSpace(filter)
	selectedWarehouseID := uint(0)
	if len(warehouseID) > 0 {
		selectedWarehouseID = warehouseID[0]
	}
	list := PriceList{Query: query, Filter: filter, WarehouseID: selectedWarehouseID}

	if selectedWarehouseID > 0 {
		var balances []models.ItemWarehouseBalance
		s.db.Preload("Item.Category").Preload("Warehouse").
			Where("warehouse_id = ?", selectedWarehouseID).
			Find(&balances)
		for _, balance := range balances {
			item := balance.Item
			if !matchesItemQuery(item, query) {
				continue
			}
			item.Quantity = balance.Quantity
			row := priceListRow(item)
			row.WarehouseName = balance.Warehouse.Name
			row.Quantity = balance.Quantity
			list.addPriceListRow(row, filter)
		}
		return list
	}

	var items []models.Item
	q := s.db.Preload("Category").Order("code asc")
	if query != "" {
		like := "%" + query + "%"
		q = q.Where("name LIKE ? OR code LIKE ? OR barcode LIKE ?", like, like, like)
	}
	q.Find(&items)
	s.applyWarehouseQuantities(items)

	for _, item := range items {
		row := priceListRow(item)
		row.Quantity = item.Quantity
		list.addPriceListRow(row, filter)
	}
	return list
}

func (l *PriceList) addPriceListRow(row PriceListRow, filter string) {
	switch row.Status {
	case PriceStatusNoPrice:
		l.Stats.NoPrice++
	case PriceStatusLoss:
		l.Stats.Loss++
	case PriceStatusWeak:
		l.Stats.Weak++
	case PriceStatusGood:
		l.Stats.Good++
	}
	if filter == "" || filter == "all" || string(row.Status) == filter {
		l.Rows = append(l.Rows, row)
	}
}

func matchesItemQuery(item models.Item, query string) bool {
	if query == "" {
		return true
	}
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(item.Name), query) ||
		strings.Contains(strings.ToLower(item.Code), query) ||
		strings.Contains(strings.ToLower(item.Barcode), query)
}

func (s *ItemService) UpdateSalePrice(id uint, salePrice float64) (bool, error) {
	if id == 0 {
		return false, errors.New("invalid item")
	}
	if salePrice < 0 {
		return false, errors.New("sale price cannot be negative")
	}
	var item models.Item
	if err := s.db.First(&item, id).Error; err != nil {
		return false, err
	}
	if err := s.db.Model(&models.Item{}).Where("id = ?", id).Update("sale_price", salePrice).Error; err != nil {
		return false, err
	}
	return salePrice > 0 && salePrice < item.PurchasePrice, nil
}

func priceListRow(item models.Item) PriceListRow {
	cost := item.PurchasePrice
	margin := 0.0
	if item.SalePrice > 0 && cost > 0 {
		margin = ((item.SalePrice - cost) / item.SalePrice) * 100
	}
	row := PriceListRow{Item: item, MarginPercent: margin}
	switch {
	case item.SalePrice <= 0:
		row.Status = PriceStatusNoPrice
		row.StatusLabel = "لم يتم تسجيل سعر بيع"
		row.StatusClass = "bg-amber-100 text-amber-800"
	case item.SalePrice < cost:
		row.Status = PriceStatusLoss
		row.StatusLabel = "خسارة بيع"
		row.StatusClass = "bg-red-100 text-red-800"
		row.IsBelowCost = true
	case margin < WeakMarginPercent:
		row.Status = PriceStatusWeak
		row.StatusLabel = "هامش ضعيف"
		row.StatusClass = "bg-yellow-100 text-yellow-800"
	default:
		row.Status = PriceStatusGood
		row.StatusLabel = "جيد"
		row.StatusClass = "bg-green-100 text-green-800"
	}
	if row.IsBelowCost {
		row.UpdateWarningLabel = "أقل من التكلفة"
	}
	return row
}
