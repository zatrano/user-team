package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func ManagerHomeHandler(c *fiber.Ctx) error {
	return c.Render("manager/home/manager_home", fiber.Map{
		"Title": "Manager Ana Sayfa",
	}, "layouts/manager_layout")
}
