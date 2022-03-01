package apiserver

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/models"
	"github.com/nais/console/pkg/requests"
	"gorm.io/gorm"
)

type UsersHandler struct {
	db *gorm.DB
}

func (h *UsersHandler) Read(_ *gin.Context, req *requests.GenericRequest) (*models.User, error) {
	user := &models.User{}
	tx := h.db.First(user, "id = ?", req.ID)
	return user, tx.Error
}

func (h *UsersHandler) List(_ *gin.Context) ([]*models.User, error) {
	users := make([]*models.User, 0)
	tx := h.db.Find(&users)
	return users, tx.Error
}

func (h *UsersHandler) Create(_ *gin.Context, req *requests.UserRequest) (*models.User, error) {
	user := &req.User
	tx := h.db.Create(user)
	return user, tx.Error
}

func (h *UsersHandler) Update(_ *gin.Context, req *requests.UserIDRequest) (*models.User, error) {
	u, _ := uuid.Parse(req.ID)
	user := &req.User
	user.ID = &u
	tx := h.db.Updates(user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	tx = h.db.First(user)
	return user, tx.Error
}

func (h *UsersHandler) Delete(_ *gin.Context, req *requests.GenericRequest) error {
	user := &models.User{}
	tx := h.db.First(user, "id = ?", req.ID)
	if tx.Error != nil {
		return tx.Error
	}
	tx = h.db.Delete(user)
	return tx.Error
}
