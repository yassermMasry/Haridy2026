package services

import (
	"errors"
	"math"

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
	if item.Name == "" || item.Code == "" {
		return errors.New("اسم وكود الصنف مطلوبان")
	}
	if item.SalePrice < 0 || item.PurchasePrice < 0 || item.Quantity < 0 {
		return errors.New("القيم المالية والكمية يجب ألا تكون سالبة")
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
	return s.db.Delete(&models.Item{}, id).Error
}

func (s *ItemService) Alerts() []models.Item {
	var items []models.Item
	s.db.Where("quantity <= minimum_stock").Order("quantity asc").Limit(10).Find(&items)
	return items
}
