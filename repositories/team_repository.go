package repositories

import (
	"strings"
	"zatrano/configs"
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ITeamRepository interface {
	FindAll() ([]models.Team, error)
	FindAndPaginate(params utils.ListParams) ([]models.Team, int64, error)
	FindByID(id uint) (*models.Team, error)
	Create(team *models.Team) error
	Update(id uint, data map[string]interface{}) error
	Delete(id uint) error
	Count() (int64, error)
}

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository() ITeamRepository {
	return &TeamRepository{db: configs.GetDB()}
}

func (r *TeamRepository) FindAll() ([]models.Team, error) {
	var teams []models.Team
	err := r.db.Find(&teams).Error
	return teams, err
}

func (r *TeamRepository) FindAndPaginate(params utils.ListParams) ([]models.Team, int64, error) {
	var teams []models.Team
	var totalCount int64

	query := r.db.Model(&models.Team{})

	if params.Name != "" {
		sqlQueryFragment, queryParams := utils.SQLFilter("name", params.Name)
		query = query.Where(sqlQueryFragment, queryParams...)
	}

	err := query.Count(&totalCount).Error
	if err != nil {
		utils.Log.Error("Takım sayısı alınırken hata (FindAndPaginate)", zap.Error(err))
		return nil, 0, err
	}

	if totalCount == 0 {
		return teams, 0, nil
	}

	sortBy := params.SortBy
	orderBy := strings.ToLower(params.OrderBy)
	if orderBy != "asc" && orderBy != "desc" {
		orderBy = utils.DefaultOrderBy
	}

	allowedSortColumns := map[string]bool{"id": true, "name": true, "status": true, "created_at": true}
	if _, ok := allowedSortColumns[sortBy]; !ok {
		sortBy = "id"
	}
	orderClause := sortBy + " " + orderBy
	query = query.Order(orderClause)

	offset := params.CalculateOffset()
	query = query.Limit(params.PerPage).Offset(offset)

	err = query.Find(&teams).Error
	if err != nil {
		utils.Log.Error("Takım verisi çekilirken hata (FindAndPaginate)", zap.Error(err))
		return nil, totalCount, err
	}

	return teams, totalCount, nil
}

func (r *TeamRepository) FindByID(id uint) (*models.Team, error) {
	var team models.Team
	err := r.db.First(&team, id).Error
	return &team, err
}
func (r *TeamRepository) Create(team *models.Team) error {
	return r.db.Create(team).Error
}
func (r *TeamRepository) Update(id uint, data map[string]interface{}) error {
	result := r.db.Model(&models.Team{}).Where("id = ?", id).Updates(data)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 && len(data) > 0 {
		utils.Log.Warn("TeamRepository.Update: ID ile eşleşen kayıt bulunamadı veya güncelleme 0 satırı etkiledi", zap.Uint("team_id", id))
	}
	return nil
}
func (r *TeamRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Team{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *TeamRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Team{}).Count(&count).Error
	return count, err
}

var _ ITeamRepository = (*TeamRepository)(nil)
