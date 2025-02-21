package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	cache "github.com/dutt23/lms/cache"
	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	service "github.com/dutt23/lms/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type membersApi struct {
	config *config.AppConfig
	db     connectors.SqliteConnector
	cache  cache.MemberCache
  service service.MemberService
}

func NewMembersApi(config *config.AppConfig, db connectors.SqliteConnector, cache cache.MemberCache, service service.MemberService) *membersApi {
	return &membersApi{
		config,
		db,
		cache,
    service,
	}
}

type addMemberRequestBody struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required,gt=1"`
}

type getMemberRequestBody struct {
	ID uint64 `uri:"id" binding:"required,min=1"`
}

type getMembersRequestBody struct {
	LastId   int32 `json:"last_id"`
	PageSize int32 `json:"page_size"`
}

type getMembersResponse struct {
	Members []*model.Member `json:"members"`
}

type updateMembersRequestBody struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// AddMember godoc
// @Summary endpoint to create member
// @Description add a member
// @Tags member
// @Accept json
// @Produce json
// @Param member body addMemberRequestBody true "Member data"
// @Success 200 {object} model.Member
// @Router /v1/member [post]
func (api *membersApi) AddMember(ctx *gin.Context) {
	var req addMemberRequestBody

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !api.cache.IsEmailUnique(ctx, req.Email) {
		ctx.JSON(http.StatusBadRequest, errors.New("please give a unique email"))
		return
	}

	member := &model.Member{
		Email:    req.Email,
		Name:     req.Name,
		JoinDate: model.TimeWrapper(time.Now()),
	}

	err := api.db.DB(ctx).Create(member).Error
	if err != nil {
		fmt.Errorf("error occurred %w", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("Unable to add member to library")))
		return
	}

	go api.postProcessAddingMember(member)
	ctx.JSON(http.StatusOK, member)
}

// GetMember godoc
// @Summary endpoint to get member
// @Description get a member
// @Tags member
// @Produce json
// @param id path integer false "member id"
// @Success 200 {object} model.Member
// @Router /v1/member/:id [get]
func (api *membersApi) GetMember(ctx *gin.Context) {
	var req getMemberRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

  member, err := api.service.GetMember(ctx, req.ID)
  if err != nil {
    fmt.Errorf("error : %w", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New(fmt.Sprintf("unable to locate member with Id %d", req.ID))))
		return 
  }

	c := context.Background()
	go api.storeMemberMeta(c, member)
	ctx.JSON(http.StatusOK, member)
}

// DeleteMember godoc
// @Summary endpoint to delete member
// @Description delete a member
// @Tags member
// @param id path integer false "member id"
// @Success 200
// @Router /v1/member/:id [delete]
func (api *membersApi) DeleteMember(ctx *gin.Context) {
	var req deleteBookRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !api.cache.DoesMemberExist(ctx, uint64(req.ID)) {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New(fmt.Sprintf("unable to locate book with Id %d", req.ID))))
		return
	}

	if err := api.db.DB(ctx).Delete(&model.Book{}, req.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New(fmt.Sprintf("unable to delete book %d", req.ID))))
		return
	}

	go api.invalidateMemberCache(req.ID)
	ctx.Status(http.StatusOK)
}

// GetMembers godoc
// @Summary endpoint to get members
// @Description get list of members
// @Tags member
// @Produce json
// @Accept json
// @Param member body getMembersRequestBody true "Member data"
// @Success 200 {object} []model.Member
// @Router /v1/member [get]
func (api *membersApi) GetMembers(ctx *gin.Context) {
	var req getMembersRequestBody

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var members []*model.Member

	lastId := 0
	pageSize := 10

	if req.LastId > 0 {
		lastId = int(req.LastId)
	}

	if req.PageSize >= 100 || req.PageSize < 1 {
		pageSize = 10
	}

	qry := api.db.DB(ctx).Model(model.Member{}).Where("id > ?", lastId).Limit(pageSize)

	tx := qry.Order(clause.OrderByColumn{
		Column: clause.Column{Name: "join_date"},
		Desc:   true,
	}).Find(&members)

	if tx.Error != nil {
		fmt.Println("not able to find any members", tx.Error)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("Unable to find any members")))
		return
	}

	ctx.JSON(http.StatusOK, getMembersResponse{Members: members})
}

// UpdateMembers godoc
// @Summary endpoint to update member
// @Description update member data
// @Tags member
// @Produce json
// @Accept json
// @Param member body addMemberRequestBody true "Member data"
// @param id path integer false "member id"
// @Success 200 {object} model.Member
// @Router /v1/member/:id [put]
func (api *membersApi) UpdateMember(ctx *gin.Context) {
	var req updateMembersRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var body addMemberRequestBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !api.cache.DoesMemberExist(ctx, uint64(req.ID)) {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New(fmt.Sprintf("unable to locate member with Id %d", req.ID))))
		return
	}

	if !api.cache.IsEmailUnique(ctx, body.Email) {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New(fmt.Sprintf("please provide a unique email Id %s", body.Email))))
		return
	}

	member := &model.Member{
		Email:    body.Email,
		Name:     body.Name,
		JoinDate: model.TimeWrapper(time.Now()),
	}

	if err := api.db.DB(ctx).Save(member).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	go api.postProcessAddingMember(member)
	ctx.JSON(http.StatusOK, member)
}

func (api *membersApi) invalidateMemberCache(bookId uint64) {
	ctx := context.Background()
	if err := api.cache.DeleteMember(ctx, bookId); err != nil {
		fmt.Println("Unable to remove entry from cache ", err)
	}
}

func (api *membersApi) postProcessAddingMember(member *model.Member) {
	ctx := context.Background()
	api.storeMemberMeta(ctx, member)
}

func (api *membersApi) storeMemberMeta(ctx context.Context, member *model.Member) {
	if err := api.cache.StoreMemberMetaInCache(ctx, member); err != nil {
		fmt.Printf("Error occurred while adding book to cache %w", err)
	}
}
