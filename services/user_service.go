package services

import (
	"zatrano/models"
	"zatrano/repositories"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserServiceError string

func (e UserServiceError) Error() string {
	return string(e)
}

const (
	ErrUserServiceUserNotFound UserServiceError = "kullanıcı bulunamadı"
	ErrPasswordHashingFailed   UserServiceError = "şifre oluşturulurken bir hata oluştu"
	ErrPasswordUpdateFailed    UserServiceError = "şifre güncellenirken bir hata oluştu"
	ErrUserCreationFailed      UserServiceError = "kullanıcı veritabanına kaydedilemedi"
	ErrUserUpdateFailed        UserServiceError = "kullanıcı veritabanında güncellenemedi"
	ErrUserDeletionFailed      UserServiceError = "kullanıcı silinirken bir veritabanı hatası oluştu"
	ErrPasswordRequired        UserServiceError = "şifre alanı boş olamaz"
)

type IUserService interface {
	GetAllUsersPaginated(params utils.ListParams) (*utils.PaginatedResult, error)
	GetUserByID(id uint) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(id uint, userData *models.User) error
	DeleteUser(id uint) error
	GetUserCount() (int64, error)
}

type UserService struct {
	repo repositories.IUserRepository
}

func NewUserService() IUserService {
	return &UserService{repo: repositories.NewUserRepository()}
}

func (s *UserService) GetAllUsersPaginated(params utils.ListParams) (*utils.PaginatedResult, error) {
	if params.Page <= 0 {
		params.Page = utils.DefaultPage
	}
	if params.PerPage <= 0 {
		params.PerPage = utils.DefaultPerPage
	} else if params.PerPage > utils.MaxPerPage {
		utils.Log.Warn("Sayfa başına istenen kayıt sayısı limiti aştı, varsayılana çekildi.",
			zap.Int("requested", params.PerPage),
			zap.Int("max", utils.MaxPerPage),
			zap.Int("default", utils.DefaultPerPage),
		)
		params.PerPage = utils.DefaultPerPage
	}
	if params.SortBy == "" {
		params.SortBy = utils.DefaultSortBy
	}
	if params.OrderBy == "" {
		params.OrderBy = utils.DefaultOrderBy
	}

	users, totalCount, err := s.repo.FindAndPaginate(params)
	if err != nil {
		return nil, err
	}

	totalPages := utils.CalculateTotalPages(totalCount, params.PerPage)

	result := &utils.PaginatedResult{
		Data: users,
		Meta: utils.PaginationMeta{
			CurrentPage: params.Page,
			PerPage:     params.PerPage,
			TotalItems:  totalCount,
			TotalPages:  totalPages,
		},
	}

	return result, nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Log.Warn("Kullanıcı bulunamadı (ID ile arama)", zap.Uint("user_id", id))
			return nil, ErrUserServiceUserNotFound
		}
		utils.Log.Error("Kullanıcı alınırken hata oluştu (ID ile arama)", zap.Uint("user_id", id), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (s *UserService) CreateUser(user *models.User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}

	if err := user.SetPassword(user.Password); err != nil {
		utils.Log.Error("Kullanıcı oluşturma: Şifre ayarlanamadı/hashlenemedi", zap.String("account", user.Account), zap.Error(err))
		return ErrPasswordHashingFailed
	}

	utils.Log.Info("Kullanıcı oluşturuluyor...",
		zap.String("account", user.Account),
		zap.Any("type", user.Type),
		zap.Any("team_id", user.TeamID),
	)

	err := s.repo.Create(user)
	if err != nil {
		utils.Log.Error("Kullanıcı oluşturulurken veritabanı hatası",
			zap.String("account", user.Account),
			zap.Error(err),
		)
		modelErr, ok := err.(models.ModelError)
		if ok {
			return modelErr
		}
		return ErrUserCreationFailed
	}

	utils.SLog.Infof("Kullanıcı başarıyla oluşturuldu: %s (ID: %d)", user.Account, user.ID)
	return nil
}

func (s *UserService) UpdateUser(id uint, userData *models.User) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Log.Warn("Kullanıcı güncellenemedi: Kullanıcı bulunamadı (ön kontrol)", zap.Uint("user_id", id))
			return ErrUserServiceUserNotFound
		}
		utils.Log.Error("Kullanıcı güncellenemedi: Kullanıcı aranırken hata (ön kontrol)", zap.Uint("user_id", id), zap.Error(err))
		return err
	}

	updateData := map[string]interface{}{
		"name":    userData.Name,
		"account": userData.Account,
		"status":  userData.Status,
		"type":    userData.Type,
		"team_id": userData.TeamID,
	}

	passwordUpdated := false
	if userData.Password != "" {
		tempUserForHash := models.User{}
		if err := tempUserForHash.SetPassword(userData.Password); err != nil {
			utils.Log.Error("Kullanıcı güncelleme: Şifre ayarlanamadı/hashlenemedi", zap.Uint("user_id", id), zap.Error(err))
			return ErrPasswordUpdateFailed
		}
		updateData["password"] = tempUserForHash.Password
		passwordUpdated = true
	}

	utils.Log.Info("Kullanıcı güncelleniyor (map ile)...",
		zap.Uint("user_id", id),
		zap.Bool("password_updated", passwordUpdated),
		zap.Uintp("team_id", userData.TeamID),
		zap.String("type", string(userData.Type)),
	)

	err = s.repo.Update(id, updateData)
	if err != nil {
		utils.Log.Error("Kullanıcı güncellenirken veritabanı hatası (Update)",
			zap.Uint("user_id", id),
			zap.Error(err),
		)
		modelErr, ok := err.(models.ModelError)
		if ok {
			return modelErr
		}
		if err == gorm.ErrRecordNotFound {
			return ErrUserServiceUserNotFound
		}
		return ErrUserUpdateFailed
	}

	utils.SLog.Infof("Kullanıcı başarıyla güncellendi (map ile): ID %d, Hesap: %s", id, userData.Account)
	return nil
}

func (s *UserService) DeleteUser(id uint) error {
	err := s.repo.Delete(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Log.Warn("Kullanıcı silinemedi: Kullanıcı bulunamadı", zap.Uint("user_id", id))
			return ErrUserServiceUserNotFound
		}
		utils.Log.Error("Kullanıcı silinirken hata oluştu (Delete)", zap.Uint("user_id", id), zap.Error(err))
		return ErrUserDeletionFailed
	}
	utils.SLog.Infof("Kullanıcı başarıyla silindi: ID %d", id)
	return nil
}

func (s *UserService) GetUserCount() (int64, error) {
	count, err := s.repo.Count()
	if err != nil {
		utils.Log.Error("Kullanıcı sayısı alınırken hata oluştu", zap.Error(err))
		return 0, err
	}
	return count, nil
}

var _ IUserService = (*UserService)(nil)
