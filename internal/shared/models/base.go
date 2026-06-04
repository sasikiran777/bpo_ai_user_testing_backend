package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type BaseModel struct {
	bun.BaseModel

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt time.Time `bun:"created_at,notnull,default:now()"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:now()"`
}
