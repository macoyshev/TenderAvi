package model

import "github.com/google/uuid"

type OrganizationResponsible struct {
	Id             uuid.UUID
	OrganizationId uuid.UUID
	UserId         uuid.UUID
}
