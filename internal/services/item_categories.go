package services

import (
	"errors"
	"strings"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

const CategoryInUseMessage = "لا يمكن حذف التصنيف لأنه مرتبط بأصناف"

type ItemCategoryService struct{ db *gorm.DB }

func NewItemCategoryService(db *gorm.DB) *ItemCategoryService {
	return &ItemCategoryService{db: db}
}

func (s *ItemCategoryService) List(tenantID uint) []models.ItemCategory {
	var categories []models.ItemCategory
	query := s.db.Order("name asc")
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantID)
	}
	query.Find(&categories)
	return categories
}

func (s *ItemCategoryService) Find(id uint) (*models.ItemCategory, error) {
	var category models.ItemCategory
	if err := s.db.First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *ItemCategoryService) Create(tenantID uint, name string) (*models.ItemCategory, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("اسم التصنيف مطلوب")
	}
	if err := s.ensureUnique(tenantID, name, 0); err != nil {
		return nil, err
	}
	category := models.ItemCategory{Name: name}
	if tenantID > 0 {
		category.TenantID = &tenantID
	}
	if err := s.db.Create(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *ItemCategoryService) Update(tenantID, id uint, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("اسم التصنيف مطلوب")
	}
	if err := s.ensureUnique(tenantID, name, id); err != nil {
		return err
	}
	return s.db.Model(&models.ItemCategory{}).Where("id = ?", id).Update("name", name).Error
}

func (s *ItemCategoryService) Delete(id uint) error {
	var count int64
	if err := s.db.Model(&models.Item{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New(CategoryInUseMessage)
	}
	return s.db.Delete(&models.ItemCategory{}, id).Error
}

func (s *ItemCategoryService) ensureUnique(tenantID uint, name string, exceptID uint) error {
	query := s.db.Unscoped().Model(&models.ItemCategory{}).Where("LOWER(name) = LOWER(?)", name)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	if exceptID > 0 {
		query = query.Where("id <> ?", exceptID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("اسم التصنيف مستخدم من قبل")
	}
	return nil
}
