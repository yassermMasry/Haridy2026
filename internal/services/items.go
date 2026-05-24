package services

import (
	"errors"
	"math"
	"strings"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

type ItemService struct{ db *gorm.DB }

type ItemList struct {
	Items      []models.Item
	Page       int
	TotalPages int
	Query      string
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
	Rows   []PriceListRow
	Stats  PriceListStats
	Query  string
	Filter string
}

func NewItemService(db *gorm.DB) *ItemService { return &ItemService{db: db} }

func (s *ItemService) Categories() []models.ItemCategory {
	var categories []models.ItemCategory
	s.db.Order("name asc").Find(&categories)
	return categories
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
		"quantity":       item.Quantity,
		"minimum_stock":  item.MinimumStock,
		"category_id":    item.CategoryID,
	}).Error
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

func (s *ItemService) PriceList(query, filter string) PriceList {
	query = strings.TrimSpace(query)
	filter = strings.TrimSpace(filter)
	var items []models.Item
	q := s.db.Preload("Category").Order("code asc")
	if query != "" {
		like := "%" + query + "%"
		q = q.Where("name LIKE ? OR code LIKE ? OR barcode LIKE ?", like, like, like)
	}
	q.Find(&items)

	list := PriceList{Query: query, Filter: filter}
	for _, item := range items {
		row := priceListRow(item)
		switch row.Status {
		case PriceStatusNoPrice:
			list.Stats.NoPrice++
		case PriceStatusLoss:
			list.Stats.Loss++
		case PriceStatusWeak:
			list.Stats.Weak++
		case PriceStatusGood:
			list.Stats.Good++
		}
		if filter == "" || filter == "all" || string(row.Status) == filter {
			list.Rows = append(list.Rows, row)
		}
	}
	return list
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
