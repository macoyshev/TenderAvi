package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type History struct {
	Id         uuid.UUID
	Version    int
	UpdatedAt  time.Time
	ObjectId   uuid.UUID
	ObjectData json.RawMessage
}
