package model

import (
	"time"

	"github.com/google/uuid"
)

type BidStatus string
type BidAuthorType string

const (
	CreatedBidStatus   BidStatus = "Created"
	PublishedBidStatus BidStatus = "Published"
	CanceledBidStatus  BidStatus = "Canceled"
)

const (
	UserBidAuthorType BidAuthorType = "User"
	OrgBidAuthorType  BidAuthorType = "Organization"
)

type Bid struct {
	Id          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      BidStatus     `json:"status"`
	TenderId    uuid.UUID     `json:"tenderId"`
	AuthorType  BidAuthorType `json:"authorType"`
	AuthorId    uuid.UUID     `json:"authorId"`
	Version     int32         `json:"version"`
	CreatedAt   time.Time     `json:"createdAt"`
}

type Review struct {
	Id          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}
