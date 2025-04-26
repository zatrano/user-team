package services

import (
	"zatrano/models"
	"zatrano/repositories"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TeamServiceError string

func (e TeamServiceError) Error() string {
	return string(e)
}

const (
	ErrTeamNotFound       TeamServiceError = "takım bulunamadı"
	ErrTeamCreationFailed TeamServiceError = "takım oluşturulamadı"
	ErrTeamUpdateFailed   TeamServiceError = "takım güncellenemedi"
	ErrTeamDeletionFailed TeamServiceError = "takım silinemedi"
)

type ITeamService interface {
	GetAllTeams() ([]models.Team, error)
	GetAllTeamsPaginated(params utils.ListParams) (*utils.PaginatedResult, error)
	GetTeamByID(id uint) (*models.Team, error)
	CreateTeam(team *models.Team) error
	UpdateTeam(id uint, teamData *models.Team) error
	DeleteTeam(id uint) error
	GetTeamCount() (int64, error)
}

type TeamService struct {
	repo repositories.ITeamRepository
}

func NewTeamService() ITeamService {
	return &TeamService{repo: repositories.NewTeamRepository()}
}

func (s *TeamService) GetAllTeams() ([]models.Team, error) {
	teams, err := s.repo.FindAll()
	if err != nil {
		utils.Log.Error("Tüm takımlar alınırken hata oluştu", zap.Error(err))
		return nil, err
	}
	return teams, nil
}

func (s *TeamService) GetAllTeamsPaginated(params utils.ListParams) (*utils.PaginatedResult, error) {
	if params.Page <= 0 {
		params.Page = utils.DefaultPage
	}
	if params.PerPage <= 0 {
		params.PerPage = utils.DefaultPerPage
	} else if params.PerPage > utils.MaxPerPage {
		params.PerPage = utils.DefaultPerPage
	}
	if params.SortBy == "" {
		params.SortBy = "id"
	}
	if params.OrderBy == "" {
		params.OrderBy = utils.DefaultOrderBy
	}

	teams, totalCount, err := s.repo.FindAndPaginate(params)
	if err != nil {
		return nil, err
	}

	totalPages := utils.CalculateTotalPages(totalCount, params.PerPage)

	result := &utils.PaginatedResult{
		Data: teams,
		Meta: utils.PaginationMeta{
			CurrentPage: params.Page, PerPage: params.PerPage,
			TotalItems: totalCount, TotalPages: totalPages,
		},
	}
	return result, nil
}

func (s *TeamService) GetTeamByID(id uint) (*models.Team, error) {
	team, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTeamNotFound
		}
		utils.Log.Error("Takım alınırken hata oluştu (ID ile arama)", zap.Uint("team_id", id), zap.Error(err))
		return nil, err
	}
	return team, nil
}
func (s *TeamService) CreateTeam(team *models.Team) error {
	err := s.repo.Create(team)
	if err != nil {
		utils.Log.Error("Takım oluşturulurken veritabanı hatası", zap.String("team_name", team.Name), zap.Error(err))
		return ErrTeamCreationFailed
	}
	utils.SLog.Infof("Takım başarıyla oluşturuldu: %s (ID: %d)", team.Name, team.ID)
	return nil
}
func (s *TeamService) UpdateTeam(id uint, teamData *models.Team) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTeamNotFound
		}
		utils.Log.Error("Takım güncellenemedi: Takım aranırken hata (ön kontrol)", zap.Uint("team_id", id), zap.Error(err))
		return err
	}
	updateData := map[string]interface{}{"name": teamData.Name, "status": teamData.Status}
	err = s.repo.Update(id, updateData)
	if err != nil {
		utils.Log.Error("Takım güncellenirken veritabanı hatası", zap.Uint("team_id", id), zap.String("new_name", teamData.Name), zap.Error(err))
		if err == gorm.ErrRecordNotFound {
			return ErrTeamNotFound
		}
		return ErrTeamUpdateFailed
	}
	utils.SLog.Infof("Takım başarıyla güncellendi: ID %d, Yeni Ad: %s", id, teamData.Name)
	return nil
}
func (s *TeamService) DeleteTeam(id uint) error {
	err := s.repo.Delete(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTeamNotFound
		}
		utils.Log.Error("Takım silinirken hata oluştu", zap.Uint("team_id", id), zap.Error(err))
		return ErrTeamDeletionFailed
	}
	utils.SLog.Infof("Takım başarıyla silindi: ID %d", id)
	return nil
}
func (s *TeamService) GetTeamCount() (int64, error) {
	count, err := s.repo.Count()
	if err != nil {
		utils.Log.Error("Takım sayısı alınırken hata oluştu", zap.Error(err))
		return 0, err
	}
	return count, nil
}

var _ ITeamService = (*TeamService)(nil)
