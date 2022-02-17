package apiserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nais/console/pkg/models"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// In case err is not nil, write a suitable error code and error message, and returns true.
// Otherwise, does nothing and returns false.
func abort(ctx *gin.Context, err error) bool {
	// FIXME: incorrectly formatting the UUID will result in a 500 error.
	// This should be handled as a bad request instead.
	// SQLSTATE 22P02

	// var perr *pgconn.PgError
	// errors.As(err, &perr)
	// fmt.Println(perr.Code) // 23505

	switch err {
	case gorm.ErrRecordNotFound:
		ctx.AbortWithError(http.StatusNotFound, err)
	default:
		ctx.AbortWithError(http.StatusInternalServerError, err)
	case nil:
		return false
	}
	return true
}

func (h *Handler) GetTeam(ctx *gin.Context) {
	team := &models.Team{}
	tx := h.db.First(team, "id = ?", ctx.Param("id"))
	if abort(ctx, tx.Error) {
		return
	}
	ctx.JSON(http.StatusOK, team)
}

func (h *Handler) GetTeams(ctx *gin.Context) {
	teams := make([]*models.Team, 0)
	tx := h.db.Find(&teams)
	if abort(ctx, tx.Error) {
		return
	}
	ctx.JSON(http.StatusOK, teams)
}

func (h *Handler) PostTeam(ctx *gin.Context) {
	team := &models.Team{}

	err := ctx.BindJSON(team)
	if abort(ctx, err) {
		return
	}

	tx := h.db.Create(team)
	if abort(ctx, tx.Error) {
		return
	}

	ctx.Redirect(http.StatusCreated, "/api/v1/teams/"+team.ID.String())
}

func (h *Handler) PutTeam(ctx *gin.Context) {
	team := &models.Team{}

	// load from db
	tx := h.db.First(team, "id = ?", ctx.Param("id"))
	if abort(ctx, tx.Error) {
		return
	}

	// overwrite data
	err := ctx.BindJSON(team)
	if abort(ctx, err) {
		return
	}

	// persist to database
	tx = h.db.Save(team)
	if abort(ctx, tx.Error) {
		return
	}

	// read back object
	tx = h.db.First(team, "id = ?", ctx.Param("id"))
	if abort(ctx, tx.Error) {
		return
	}

	ctx.JSON(http.StatusOK, team)
}

func (h *Handler) DeleteTeam(ctx *gin.Context) {
	team := &models.Team{}

	// load from db
	tx := h.db.First(team, "id = ?", ctx.Param("id"))
	if abort(ctx, tx.Error) {
		return
	}

	// if found, try deleting
	tx = h.db.Delete(team, "id = ?", ctx.Param("id"))
	abort(ctx, tx.Error)
}
