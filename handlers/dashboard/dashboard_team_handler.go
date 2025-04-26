package handlers

import (
	"strconv"
	"zatrano/models"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type TeamHandler struct {
	service services.ITeamService
}

func NewTeamHandler() *TeamHandler {
	return &TeamHandler{service: services.NewTeamService()}
}

func (h *TeamHandler) ListTeams(c *fiber.Ctx) error {
	flashData, flashErr := utils.GetFlashMessages(c)
	if flashErr != nil {
		utils.Log.Warn("Takım listesi: Flash mesajları alınamadı", zap.Error(flashErr))
	}

	var params utils.ListParams
	if err := c.QueryParser(&params); err != nil {
		utils.Log.Warn("Takım listesi: Query parametreleri parse edilemedi, varsayılanlar kullanılıyor.", zap.Error(err))
		params = utils.ListParams{
			Page:    utils.DefaultPage,
			PerPage: utils.DefaultPerPage,
			SortBy:  "id",
			OrderBy: utils.DefaultOrderBy,
		}
	}

	if params.Page <= 0 {
		params.Page = utils.DefaultPage
	}
	if params.PerPage <= 0 {
		params.PerPage = utils.DefaultPerPage
	} else if params.PerPage > utils.MaxPerPage {
		utils.Log.Warn("Sayfa başına istenen kayıt sayısı limiti aştı, varsayılana çekildi.",
			zap.Int("requested", params.PerPage), zap.Int("max", utils.MaxPerPage), zap.Int("default", utils.DefaultPerPage))
		params.PerPage = utils.DefaultPerPage
	}
	if params.SortBy == "" {
		params.SortBy = "id"
	}
	if params.OrderBy == "" {
		params.OrderBy = utils.DefaultOrderBy
	}

	paginatedResult, dbErr := h.service.GetAllTeamsPaginated(params)

	renderData := fiber.Map{
		"Title":     "Takımlar",
		"CsrfToken": c.Locals("csrf"),
		"Result":    paginatedResult,
		"Params":    params,
		"Success":   flashData.Success,
		"Error":     flashData.Error,
	}

	if dbErr != nil {
		dbErrMsg := "Takımlar getirilirken bir hata oluştu."
		utils.Log.Error("Takım listesi DB Hatası", zap.Error(dbErr))
		if existingErr, ok := renderData["Error"].(string); ok && existingErr != "" {
			renderData["Error"] = existingErr + " | " + dbErrMsg
		} else {
			renderData["Error"] = dbErrMsg
		}
		renderData["Result"] = &utils.PaginatedResult{
			Data: []models.Team{},
			Meta: utils.PaginationMeta{CurrentPage: params.Page, PerPage: params.PerPage, TotalItems: 0, TotalPages: 0},
		}
	}

	return c.Render("dashboard/teams/dashboard_teams_list", renderData, "layouts/dashboard_layout")
}

func (h *TeamHandler) ShowCreateTeam(c *fiber.Ctx) error {
	flashData, flashErr := utils.GetFlashMessages(c)
	if flashErr != nil {
		utils.Log.Warn("Takım oluşturma formu: Flash mesajları alınamadı", zap.Error(flashErr))
	}
	return c.Render("dashboard/teams/dashboard_teams_create", fiber.Map{
		"Title":     "Yeni Takım Ekle",
		"CsrfToken": c.Locals("csrf"),
		"Success":   flashData.Success,
		"Error":     flashData.Error,
	}, "layouts/dashboard_layout")
}

func (h *TeamHandler) CreateTeam(c *fiber.Ctx) error {
	type Request struct {
		Name   string `form:"name"`
		Status string `form:"status"`
	}
	var req Request

	renderError := func(errorMsg string, statusCode int, formData Request) error {
		return c.Status(statusCode).Render("dashboard/teams/dashboard_teams_create", fiber.Map{
			"Title":     "Yeni Takım Ekle",
			"CsrfToken": c.Locals("csrf"),
			"Error":     errorMsg,
			"FormData":  formData,
		}, "layouts/dashboard_layout")
	}

	if err := c.BodyParser(&req); err != nil {
		utils.SLog.Warnf("Takım oluşturma isteği ayrıştırılamadı: %v", err)
		return renderError("Geçersiz form verisi.", fiber.StatusBadRequest, req)
	}

	if req.Name == "" {
		return renderError("Takım adı boş olamaz.", fiber.StatusBadRequest, req)
	}

	status := req.Status == "true"
	team := models.Team{Name: req.Name, Status: status}

	if err := h.service.CreateTeam(&team); err != nil {
		utils.Log.Error("Takım oluşturulamadı (Servis Hatası)", zap.String("team_name", team.Name), zap.Error(err))
		return renderError("Takım oluşturulamadı: "+err.Error(), fiber.StatusInternalServerError, req)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Takım başarıyla oluşturuldu.")
	return c.Redirect("/dashboard/teams", fiber.StatusFound)
}

func (h *TeamHandler) ShowUpdateTeam(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		utils.Log.Warn("Takım güncelleme formu: Geçersiz ID parametresi", zap.String("param", c.Params("id")))
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz takım ID'si.")
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}
	teamID := uint(id)

	team, err := h.service.GetTeamByID(teamID)
	if err != nil {
		var errMsg string
		if err == services.ErrTeamNotFound {
			errMsg = "Düzenlenecek takım bulunamadı."
			utils.Log.Warn("Takım güncelleme formu: Takım bulunamadı", zap.Uint("team_id", teamID))
		} else {
			errMsg = "Takım bilgileri getirilirken bir hata oluştu."
			utils.Log.Error("Takım güncelleme formu: Takım alınamadı (Servis Hatası)", zap.Uint("team_id", teamID), zap.Error(err))
		}
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, errMsg)
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}

	flashData, flashErr := utils.GetFlashMessages(c)
	if flashErr != nil {
		utils.Log.Warn("Takım güncelleme formu: Flash mesajları alınamadı", zap.Uint("team_id", teamID), zap.Error(flashErr))
	}

	return c.Render("dashboard/teams/dashboard_teams_update", fiber.Map{
		"Title":     "Takım Düzenle",
		"Team":      team,
		"CsrfToken": c.Locals("csrf"),
		"Success":   flashData.Success,
		"Error":     flashData.Error,
	}, "layouts/dashboard_layout")
}

func (h *TeamHandler) UpdateTeam(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		utils.Log.Warn("Takım güncelleme: Geçersiz ID parametresi", zap.String("param", c.Params("id")))
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz takım ID'si.")
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}
	teamID := uint(id)
	redirectPathOnSuccess := "/dashboard/teams"
	redirectPathOnError := "/dashboard/teams/update/" + strconv.Itoa(id)

	type Request struct {
		Name   string `form:"name"`
		Status string `form:"status"`
	}
	var req Request

	renderError := func(errorMsg string, statusCode int, formData Request) error {
		team, _ := h.service.GetTeamByID(teamID)
		return c.Status(statusCode).Render("dashboard/teams/dashboard_teams_update", fiber.Map{
			"Title":     "Takım Düzenle",
			"CsrfToken": c.Locals("csrf"),
			"Error":     errorMsg,
			"Team":      team,
			"FormData":  formData,
		}, "layouts/dashboard_layout")
	}

	if err := c.BodyParser(&req); err != nil {
		utils.Log.Warn("Takım güncelleme: Form verileri okunamadı", zap.Uint("team_id", teamID), zap.Error(err))
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Form verileri okunamadı.")
		return c.Redirect(redirectPathOnError, fiber.StatusSeeOther)
	}

	if req.Name == "" {
		return renderError("Takım adı boş olamaz.", fiber.StatusBadRequest, req)
	}

	newStatus := req.Status == "true"
	teamToUpdate := &models.Team{
		Name:   req.Name,
		Status: newStatus,
	}

	if err := h.service.UpdateTeam(teamID, teamToUpdate); err != nil {
		var errMsg string
		statusCode := fiber.StatusInternalServerError

		if err == services.ErrTeamNotFound {
			utils.Log.Warn("Takım güncelleme: Takım bulunamadı (Servis hatası)", zap.Uint("team_id", teamID))
			errMsg = "Güncellenecek takım bulunamadı."
			_ = utils.SetFlashMessage(c, utils.FlashErrorKey, errMsg)
			return c.Redirect(redirectPathOnSuccess, fiber.StatusSeeOther)
		} else {
			utils.Log.Error("Takım güncelleme: Servis hatası", zap.Uint("team_id", teamID), zap.Error(err))
			errMsg = "Takım güncellenemedi: " + err.Error()
		}
		return renderError(errMsg, statusCode, req)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Takım başarıyla güncellendi.")
	return c.Redirect(redirectPathOnSuccess, fiber.StatusFound)
}

func (h *TeamHandler) DeleteTeam(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		utils.Log.Warn("Takım silme: Geçersiz ID parametresi", zap.String("param", c.Params("id")))
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz takım ID'si.")
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}
	teamID := uint(id)

	if err := h.service.DeleteTeam(teamID); err != nil {
		var errMsg string
		if err == services.ErrTeamNotFound {
			errMsg = "Silinecek takım bulunamadı."
			utils.Log.Warn("Takım silme: Takım bulunamadı", zap.Uint("team_id", teamID))
		} else {
			errMsg = "Takım silinemedi: " + err.Error()
			utils.Log.Error("Takım silme: Servis hatası", zap.Uint("team_id", teamID), zap.Error(err))
		}
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, errMsg)
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Takım başarıyla silindi.")
	return c.Redirect("/dashboard/teams", fiber.StatusFound)
}
