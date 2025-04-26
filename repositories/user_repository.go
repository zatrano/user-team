package repositories

import (
	"strings"
	"zatrano/configs"
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IUserRepository interface {
	FindAndPaginate(params utils.ListParams) ([]models.User, int64, error)
	FindByID(id uint) (*models.User, error)
	Create(user *models.User) error
	Update(id uint, data map[string]interface{}) error
	Delete(id uint) error
	Count() (int64, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository() IUserRepository {
	return &UserRepository{db: configs.GetDB()}
}

func (r *UserRepository) FindAndPaginate(params utils.ListParams) ([]models.User, int64, error) {
	var users []models.User
	var totalCount int64

	query := r.db.Model(&models.User{})

	if params.Name != "" {
		sqlQueryFragment, queryParams := utils.SQLFilter("name", params.Name)
		query = query.Where(sqlQueryFragment, queryParams...)
	}

	err := query.Count(&totalCount).Error
	if err != nil {
		utils.Log.Error("Kullanıcı sayısı alınırken hata (FindAndPaginate)", zap.Error(err))
		return nil, 0, err
	}

	if totalCount == 0 {
		return users, 0, nil
	}

	sortBy := params.SortBy
	orderBy := strings.ToLower(params.OrderBy)
	if orderBy != "asc" && orderBy != "desc" {
		orderBy = utils.DefaultOrderBy
	}
	allowedSortColumns := map[string]bool{"id": true, "name": true, "account": true, "team_id": true, "created_at": true, "status": true, "type": true}
	if _, ok := allowedSortColumns[sortBy]; !ok {
		sortBy = utils.DefaultSortBy
	}
	orderClause := sortBy + " " + orderBy
	query = query.Order(orderClause)

	query = query.Preload(clause.Associations)

	offset := params.CalculateOffset()
	query = query.Limit(params.PerPage).Offset(offset)

	err = query.Find(&users).Error
	if err != nil {
		utils.Log.Error("Kullanıcı verisi çekilirken hata (FindAndPaginate)", zap.Error(err))
		return nil, totalCount, err
	}

	return users, totalCount, nil
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload(clause.Associations).First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(id uint, data map[string]interface{}) error {
	result := r.db.Model(&models.User{}).Where("id = ?", id).Updates(data)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 && len(data) > 0 {
		utils.Log.Warn("UserRepository.Update: ID ile eşleşen kayıt bulunamadı veya güncelleme 0 satırı etkiledi", zap.Uint("user_id", id))
	}
	return nil
}

func (r *UserRepository) Delete(id uint) error {
	result := r.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		utils.Log.Warn("UserRepository.Delete: Silinecek kullanıcı bulunamadı", zap.Uint("user_id", id))
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

var _ IUserRepository = (*UserRepository)(nil)
