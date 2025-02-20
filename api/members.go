package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
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
	filter *bloom.BloomFilter
	cache  service.MemberCacheService
}

func NewMembersApi(config *config.AppConfig, db connectors.SqliteConnector, bookFilter *bloom.BloomFilter, cache service.MemberCacheService) *membersApi {
	return &membersApi{
		config: config,
		db:     db,
		filter: bookFilter,
		cache:  cache,
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

func (api *membersApi) GetMember(ctx *gin.Context) {
	var req getMemberRequestBody

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	res := api.cache.GetMember(ctx, (req.ID))
	if res != nil {
		ctx.JSON(http.StatusOK, res)
		return
	}

	db := api.db.DB(ctx)
	var member *model.Member
	if err := db.Last(&member, req.ID).Error; err != nil {
		fmt.Errorf("error : %w", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New(fmt.Sprintf("unable to locate member with Id %d", req.ID))))
		return
	}

	c := context.Background()
	go api.storeMemberMeta(c, member)
	ctx.JSON(http.StatusOK, member)
}

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
	api.filter.Add([]byte(member.Email))
	api.storeMemberMeta(ctx, member)
}

func (api *membersApi) storeMemberMeta(ctx context.Context, member *model.Member) {
	if err := api.cache.StoreMemberMetaInCache(ctx, member); err != nil {
		fmt.Printf("Error occurred while adding book to cache %w", err)
	}
}
