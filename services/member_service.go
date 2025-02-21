package service

import (
	"context"
	"fmt"

	"github.com/dutt23/lms/cache"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	"gorm.io/gorm/clause"
)

type memberService struct {
	db    connectors.SqliteConnector
	cache cache.MemberCache
}

func NewMemberService(db connectors.SqliteConnector, cache cache.MemberCache) MemberService {
	return &memberService{db, cache}
}

func (service *memberService) GetMember(ctx context.Context, memberId uint64) (*model.Member, error) {
	res := service.cache.GetMember(ctx, memberId)
	if res != nil {
		return res, nil
	}

	db := service.db.DB(ctx)
	var member *model.Member
	if err := db.Last(&member, memberId).Error; err != nil {
		return nil, err
	}
	return member, nil
}

func (service *memberService) GetMembers(ctx context.Context, lastId uint64, pageSize int) ([]*model.Member, error) {
	db := service.db.DB(ctx)
	var members []*model.Member
	qry := db.Model(model.Member{}).Where("id > ?", lastId).Limit(pageSize)

	tx := qry.Order(clause.OrderByColumn{
		Column: clause.Column{Name: "join_date"},
		Desc:   true,
	}).Find(&members)

	if tx.Error != nil {
		fmt.Println("not able to find any loans", tx.Error)
		return nil, tx.Error
	}
	return members, nil
}

func (service *memberService) GetMemberByEmail(ctx context.Context, email string) (*model.Member, error) {
	db := service.db.DB(ctx)
	var member *model.Member
	if err := db.Where("email = ?", email).Find(&member).Error; err != nil {
		return nil, err
	}
	return member, nil
}
