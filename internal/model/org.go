package model

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationType string

const (
	OrgTypeIE  OrganizationType = "IE"
	OrgTypeLLC OrganizationType = "LLC"
	OrgTypeJSC OrganizationType = "JSC"
)

type Organization struct {
	Id               uuid.UUID
	Name             string
	Description      string
	OrganizationType OrganizationType
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
