package service

import (
	"context"

	"github.com/dutt23/lms/cache"
	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
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
