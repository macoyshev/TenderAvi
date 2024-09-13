package model

import (
	"time"

	"github.com/google/uuid"
)

type TenderStatus string
type TenderServiceType string

const (
	TenderStatusCreated   TenderStatus = "Created"
	TenderStatusPublished TenderStatus = "Published"
	TenderStatusClosed    TenderStatus = "Closed"
)

const (
	TenderServiceTypeConstruction TenderServiceType = "Construction"
	TenderServiceTypeDelivery     TenderServiceType = "Delivery"
	TenderServiceTypeManufacture  TenderServiceType = "Manufacture"
)

type Tender struct {
	Id             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	ServiceType    TenderServiceType `json:"serviceType"`
	Status         TenderStatus      `json:"status"`
	OrganizationId uuid.UUID         `json:"organizationId"`
	Version        int32             `json:"version"`
	CreatedAt      time.Time         `json:"createdAt"`
}
