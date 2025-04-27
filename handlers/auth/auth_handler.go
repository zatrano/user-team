package handlers

import (
	"zatrano/models"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthHandler struct {
	service services.IAuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{service: services.NewAuthService()}
}

func (h *AuthHandler) ShowLogin(c *fiber.Ctx) error {
	flashData, err := utils.GetFlashMessages(c)
	if err != nil {
		utils.Log.Warn("Giriş sayfası: Flash mesajları alınamadı", zap.Error(err))
	}

	return c.Render("auth/auth_login", fiber.Map{
		"Title":     "Giriş",
		"CsrfToken": c.Locals("csrf"),
		"Success":   flashData.Success,
		"Error":     flashData.Error,
	}, "layouts/auth_layout")
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request struct {
		Account  string `form:"account"`
		Password string `form:"password"`
	}

	if err := c.BodyParser(&request); err != nil {
		utils.SLog.Warnf("Login isteği ayrıştırılamadı: %v", err)
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Lütfen hesap adı ve şifre alanlarını doldurun.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	if request.Account == "" || request.Password == "" {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Lütfen hesap adı ve şifre alanlarını doldurun.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	user, err := h.service.Authenticate(request.Account, request.Password)
	if err != nil {
		var errMsg string
		switch err {
		case services.ErrInvalidCredentials:
			errMsg = "Kullanıcı adı veya şifre hatalı."
		case services.ErrUserInactive:
			errMsg = "Hesabınız aktif değil. Lütfen yöneticinizle iletişime geçin."
		default:
			errMsg = "Giriş işlemi sırasında bir sorun oluştu. Lütfen tekrar deneyin."
			utils.Log.Error("Kimlik doğrulama servisinde beklenmeyen hata",
				zap.String("account", request.Account),
				zap.Error(err),
			)
		}
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, errMsg)
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	sess, sessionErr := utils.SessionStart(c)
	if sessionErr != nil {
		utils.Log.Error("Oturum başlatılamadı (Login)",
			zap.Uint("user_id", user.ID),
			zap.String("account", user.Account),
			zap.Error(sessionErr),
		)
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Oturum başlatılamadı. Lütfen tekrar deneyin.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	sess.Set("user_id", user.ID)
	sess.Set("user_type", string(user.Type))
	sess.Set("user_status", user.Status)
	sess.Set("user_name", user.Name)

	if saveErr := sess.Save(); saveErr != nil {
		utils.Log.Error("Oturum kaydedilemedi (Login)",
			zap.Uint("user_id", user.ID),
			zap.String("account", user.Account),
			zap.Error(saveErr),
		)
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Oturum bilgileri kaydedilemedi.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	var redirectURL string
	switch user.Type {
	case models.Panel:
		redirectURL = "/panel/home"
	case models.System:
		redirectURL = "/dashboard/home"
	default:
		utils.Log.Error("Geçersiz kullanıcı tipi (Login sonrası yönlendirme)",
			zap.Uint("user_id", user.ID),
			zap.String("account", user.Account),
			zap.String("type", string(user.Type)),
		)
		_ = sess.Destroy()
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Hesabınız için tanımlanmış bir rol bulunamadı.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Başarıyla giriş yapıldı.")
	return c.Redirect(redirectURL, fiber.StatusFound)
}

func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	flashData, flashErr := utils.GetFlashMessages(c)
	if flashErr != nil {
		utils.Log.Warn("Profil: Flash mesajları alınamadı", zap.Error(flashErr))
	}

	userID, ok := c.Locals("userID").(uint)
	if !ok {
		utils.SLog.Debug("Profil: UserID locals'ta bulunamadı, session kontrol ediliyor.")
		sess, sessionErr := utils.SessionStart(c)
		if sessionErr != nil {
			utils.Log.Error("Profil: Oturum başlatılamadı (locals'ta ID yok)", zap.Error(sessionErr))
			_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Oturum hatası, lütfen tekrar giriş yapın.")
			return c.Redirect("/auth/login", fiber.StatusSeeOther)
		}
		userIDValue := sess.Get("user_id")
		switch v := userIDValue.(type) {
		case uint:
			userID = v
			ok = true
		case int:
			userID = uint(v)
			ok = true
		case float64:
			userID = uint(v)
			ok = true
		default:
			ok = false
		}
		if !ok {
			utils.Log.Warn("Profil: Session'da geçersiz veya eksik user_id", zap.Any("value", userIDValue))
			_ = sess.Destroy()
			_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz oturum bilgisi, lütfen tekrar giriş yapın.")
			return c.Redirect("/auth/login", fiber.StatusSeeOther)
		}
		utils.SLog.Debugf("Profil: UserID session'dan alındı: %d", userID)
	}

	user, err := h.service.GetUserProfile(userID)
	if err != nil {
		var errMsg string
		if err == services.ErrUserNotFound {
			errMsg = "Profil bilgileri bulunamadı, lütfen tekrar giriş yapın."
			utils.Log.Warn("Profil: Kullanıcı bulunamadı", zap.Uint("user_id", userID))
			sess, _ := utils.SessionStart(c)
			if sess != nil {
				_ = sess.Destroy()
			}
		} else {
			errMsg = "Profil bilgileri alınırken bir hata oluştu."
			utils.Log.Error("Profil: Kullanıcı profili alınırken hata", zap.Uint("user_id", userID), zap.Error(err))
		}
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, errMsg)
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	return c.Render("auth/auth_profile", fiber.Map{
		"Title":     "Profilim",
		"User":      user,
		"CsrfToken": c.Locals("csrf"),
		"Success":   flashData.Success,
		"Error":     flashData.Error,
	}, "layouts/auth_layout")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sess, err := utils.SessionStart(c)
	if err != nil {
		utils.Log.Warn("Çıkış: Oturum başlatılamadı (muhtemelen zaten yok)", zap.Error(err))
		_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Çıkış yapıldı.")
		return c.Redirect("/auth/login", fiber.StatusFound)
	}

	flashMsg := "Başarıyla çıkış yapıldı."
	if destroyErr := sess.Destroy(); destroyErr != nil {
		utils.Log.Error("Çıkış: Oturum yok edilemedi", zap.Error(destroyErr))
		flashMsg = "Çıkış yapıldı (ancak oturum temizlenirken bir sorun oluştu)."
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, flashMsg)
	return c.Redirect("/auth/login", fiber.StatusFound)
}

func (h *AuthHandler) UpdatePassword(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		if !ok {
			utils.Log.Warn("Parola Güncelleme: Session'da geçersiz veya eksik user_id", zap.Any("value", c.Locals("userID")))
			sess, _ := utils.SessionStart(c)
			if sess != nil {
				_ = sess.Destroy()
			}
			_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz oturum bilgisi, lütfen tekrar giriş yapın.")
			return c.Redirect("/auth/login", fiber.StatusSeeOther)
		}
	}

	var request struct {
		CurrentPassword string `form:"current_password"`
		NewPassword     string `form:"new_password"`
		ConfirmPassword string `form:"confirm_password"`
	}
	if err := c.BodyParser(&request); err != nil {
		utils.SLog.Warnf("Parola güncelleme isteği ayrıştırılamadı: %v", err)
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Lütfen tüm şifre alanlarını doldurun.")
		return c.Redirect("/auth/profile", fiber.StatusSeeOther)
	}
	if request.CurrentPassword == "" || request.NewPassword == "" || request.ConfirmPassword == "" {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Lütfen tüm şifre alanlarını doldurun.")
		return c.Redirect("/auth/profile", fiber.StatusSeeOther)
	}
	if request.NewPassword != request.ConfirmPassword {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Yeni şifreler uyuşmuyor.")
		return c.Redirect("/auth/profile", fiber.StatusSeeOther)
	}

	err := h.service.UpdatePassword(userID, request.CurrentPassword, request.NewPassword)
	if err != nil {
		var errMsg string
		flashKey := utils.FlashErrorKey
		redirectTarget := "/auth/profile"
		logoutUser := false

		switch err {
		case services.ErrCurrentPasswordIncorrect:
			errMsg = "Mevcut şifreniz hatalı."
		case services.ErrPasswordTooShort:
			errMsg = err.Error()
		case services.ErrPasswordSameAsOld:
			errMsg = err.Error()
		case services.ErrUserNotFound:
			errMsg = "Kullanıcı bulunamadı, lütfen tekrar giriş yapın."
			logoutUser = true
			redirectTarget = "/auth/login"
			utils.Log.Warn("Parola Güncelleme: Kullanıcı bulunamadı (servis hatası)", zap.Uint("user_id", userID))
		default:
			errMsg = "Şifre güncellenirken bilinmeyen bir hata oluştu."
			utils.Log.Error("Parola güncelleme servisinde beklenmeyen hata", zap.Uint("user_id", userID), zap.Error(err))
		}

		if logoutUser {
			sess, _ := utils.SessionStart(c)
			if sess != nil {
				_ = sess.Destroy()
			}
		}

		_ = utils.SetFlashMessage(c, flashKey, errMsg)
		return c.Redirect(redirectTarget, fiber.StatusSeeOther)
	}

	flashMsg := "Şifre başarıyla güncellendi. Lütfen yeni şifrenizle tekrar giriş yapın."
	sess, sessionErr := utils.SessionStart(c)
	if sess != nil {
		if destroyErr := sess.Destroy(); destroyErr != nil {
			utils.Log.Error("Parola güncellendi ancak oturum yok edilemedi",
				zap.Uint("user_id", userID),
				zap.Error(destroyErr),
			)
			flashMsg = "Şifre başarıyla güncellendi (ancak mevcut oturum sonlandırılamadı). Lütfen tekrar giriş yapın."
		}
	} else if sessionErr != nil {
		utils.Log.Error("Parola güncellendi ancak oturum başlatılamadı/alınamadı", zap.Uint("user_id", userID), zap.Error(sessionErr))
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, flashMsg)
	return c.Redirect("/auth/login", fiber.StatusFound)
}
