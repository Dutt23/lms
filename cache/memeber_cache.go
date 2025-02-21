package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	"golang.org/x/exp/rand"
)

type MemberCache interface {
	StoreMemberMetaInCache(c context.Context, member *model.Member) error
	IsEmailUnique(c context.Context, email string) bool
	GetMember(c context.Context, memberId uint64) *model.Member
	DoesMemberExist(c context.Context, memberId uint64) bool
	DeleteMember(c context.Context, memberId uint64) error
}

type memberCache struct {
	conn connectors.CacheConnector
}

func NewMemberCache(client connectors.CacheConnector) MemberCache {
	return &memberCache{conn: client}
}

func (cache *memberCache) StoreMemberMetaInCache(c context.Context, member *model.Member) error {
	bookKey := CacheKey(c, "SET_MEMBER", fmt.Sprintf("%d", member.Id))

	db := cache.conn.DB(c)
	pipe := db.Pipeline()
	bookExpiryTime := 1 * time.Hour
	jitter := time.Duration(rand.Int63n(int64(bookExpiryTime)))
	data, err := json.Marshal(member)
	if err != nil {
		fmt.Errorf("Unable to cache the record as value is not marshalable %s", err, bookKey)
	}
	pipe.Set(c, bookKey, data, bookExpiryTime+jitter/2)

	pipe.BFAdd(c, MEMBER_EMAIL_FILTER, member.Id)

	_, err = pipe.Exec(c)
	return err
}

func (cache *memberCache) IsEmailUnique(c context.Context, email string) bool {
	db := cache.conn.DB(c)
	res, err := db.BFExists(c, MEMBER_EMAIL_FILTER, email).Result()

	if err != nil {
		fmt.Errorf("unable to determine result %w", err)
		// This will go to the database for confirmation
		return true
	}
	return !res
}

func (cache *memberCache) DoesMemberExist(c context.Context, memberId uint64) bool {
	db := cache.conn.DB(c)
	memberKey := CacheKey(c, "SET_MEMBER", fmt.Sprintf("%d", memberId))
	res, err := db.Exists(c, memberKey).Result()

	if err != nil {
		fmt.Errorf("unable to determine result %w", err)
		// This will go to the database for confirmation
		return true
	}
	return res > 0
}

func (cache *memberCache) GetMember(c context.Context, memberId uint64) *model.Member {
	db := cache.conn.DB(c)
	memberKey := CacheKey(c, "SET_MEMBER", fmt.Sprintf("%d", memberId))
	res, err := db.Get(c, memberKey).Bytes()

	if err != nil {
		fmt.Println(fmt.Errorf("unable to get result from cache %w", err))
		// This will go to the database for confirmation
		return nil
	}

	book := &model.Member{}
	err = json.Unmarshal(res, book)

	if err != nil {
		fmt.Println(fmt.Errorf("unable to get result from cache %w", err))
		// This will go to the database for confirmation
		return nil
	}
	return book
}

func (cache *memberCache) DeleteMember(c context.Context, memberId uint64) error {
	db := cache.conn.DB(c)
	bookKey := CacheKey(c, "SET_MEMBER", fmt.Sprintf("%d", memberId))
	return db.Del(c, bookKey).Err()
}
