package apiserver

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/models"
	"github.com/wI2L/fizz"
	"gorm.io/gorm"
)

type TeamsHandler struct {
	db *gorm.DB
}

func (h *TeamsHandler) RouterGroup(parent *fizz.RouterGroup) *fizz.RouterGroup {
	return parent.Group(
		"/teams",
		"Teams",
		"manage teams, how to work with teams",
	)
}

func (h *TeamsHandler) CrudSpec() crudspec {
	return crudspec{
		create:   h.PostTeam,
		read:     h.GetTeam,
		list:     h.GetTeams,
		update:   h.PutTeam,
		delete:   h.DeleteTeam,
		singular: "team",
		plural:   "teams",
	}
}

func (h *TeamsHandler) GetTeam(_ *gin.Context, req *GenericRequest) (*models.Team, error) {
	team := &models.Team{}
	tx := h.db.First(team, "id = ?", req.ID)
	return team, tx.Error
}

func (h *TeamsHandler) GetTeams(_ *gin.Context) ([]*models.Team, error) {
	teams := make([]*models.Team, 0)
	tx := h.db.Find(&teams)
	return teams, tx.Error
}

func (h *TeamsHandler) PostTeam(_ *gin.Context, team *models.Team) (*models.Team, error) {
	tx := h.db.Create(team)
	return team, tx.Error
}

func (h *TeamsHandler) PutTeam(_ *gin.Context, req *TeamRequest) (*models.Team, error) {
	u, _ := uuid.Parse(req.ID)
	team := &req.Team
	team.ID = &u
	tx := h.db.Updates(team)
	if tx.Error != nil {
		return nil, tx.Error
	}
	tx = h.db.First(team)
	return team, tx.Error
}

func (h *TeamsHandler) DeleteTeam(_ *gin.Context, req *GenericRequest) error {
	team := &models.Team{}
	tx := h.db.First(team, "id = ?", req.ID)
	if tx.Error != nil {
		return tx.Error
	}
	tx = h.db.Delete(team)
	return tx.Error
}
