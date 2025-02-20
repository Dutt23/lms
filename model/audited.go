package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type TimeWrapper time.Time

type Audited struct {
	Id uint64 `json:"id" gorm:"type:bigint;primaryKey;autoIncrement"`
}

func (t TimeWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(timestamppb.New(time.Time(t)))
}

func (t TimeWrapper) Value() (driver.Value, error) {
	return time.Time(t), nil
}
