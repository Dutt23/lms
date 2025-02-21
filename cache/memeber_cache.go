package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dutt23/lms/model"
	"github.com/dutt23/lms/pkg/connectors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/rand"
)

type MemberCache interface {
	StoreMemberMetaInCache(c context.Context, member *model.Member) error
	IsEmailUnique(c context.Context, email string) bool
	GetMember(c context.Context, memberId uint64) *model.Member
	DoesMemberExist(c context.Context, memberId uint64) bool
	DeleteMember(c context.Context, memberId uint64) error
	GetMemberAnalytics(c context.Context, memberIds []uint64) (*MemberAnalytics, error)
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

func (cache *memberCache) GetMemberAnalytics(c context.Context, memberIds []uint64) (*MemberAnalytics, error) {
	db := cache.conn.DB(c)
	pipe := db.Pipeline()

	for _, memberId := range memberIds {
		key := fmt.Sprintf("SET_INTERNAL_ANALYTICS_MEMBER_%d", memberId)
		pipe.ZRangeWithScores(c, key, 0, 10)
	}

	res, err := pipe.Exec(c)

	if err != nil {
		fmt.Println("Error occured here ", err)
		return nil, err
	}

	analytics := make(map[string]*MemberAnalytic)
	for _, result := range res {
		zrangeRes := result.(*redis.ZSliceCmd)
		memberAnalytics := &MemberAnalytic{}
		memberFreq := make([]*MemberFreq, len(zrangeRes.Val()))
		args := result.Args()
		name := args[1].(string)
		// scores := args[4].(interface{})
		// fmt.Println(result.
		for idx, r := range zrangeRes.Val() {
			memberFreq[idx] = &MemberFreq{
				Week:  r.Member.(string),
				Count: uint64(r.Score),
			}
		}
		memberAnalytics.MemberFrequency = memberFreq
		id := strings.Replace(name, "SET_INTERNAL_ANALYTICS_MEMBER_", "", -1)
		analytics[id] = memberAnalytics
	}

	resp := &MemberAnalytics{}
	resp.Analytics = analytics
	return resp, nil
}
