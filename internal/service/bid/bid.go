package bid

import (
	"errors"

	"github.com/google/uuid"

	"avi/internal/model"
	"avi/internal/repository/bid"
	"avi/internal/repository/organization"
	"avi/internal/repository/tender"
	"avi/internal/repository/user"
)

var ErrorUserNotFound = errors.New("user does not exist")
var ErrorUserIsNotOrgResponsible = errors.New("user is not organization responsible")
var ErrorTenderNotFound = errors.New("tender does not exist")
var ErrorOrgNotFound = errors.New("organization does not exist")
var ErrorBidNotFound = errors.New("bid does not exist")

type BidService struct {
	tenderRepo *tender.TenderRepo
	bidRepo    *bid.BidRepo
	userRepo   *user.UserRepo
	orgRepo    *organization.OrganizationRepo
}

func (service *BidService) CreateBid(
	name string,
	description string,
	tenderId uuid.UUID,
	authorType model.BidAuthorType,
	authorId uuid.UUID,
) (bid *model.Bid, err error) {
	_, err = service.tenderRepo.GetTenderById(tenderId)
	if err != nil {
		return nil, ErrorTenderNotFound
	}

	switch authorType {
	case model.OrgBidAuthorType:
		_, err = service.orgRepo.GetOrganizationById(authorId)
		if err != nil {
			return nil, ErrorOrgNotFound
		}
	case model.UserBidAuthorType:
		_, err = service.userRepo.GetUserById(authorId)
		if err != nil {
			return nil, ErrorUserNotFound
		}
	default:
		return nil, errors.New("not allowed author type")
	}
	bid, err = service.bidRepo.CreateBid(
		name, description, tenderId, authorType, authorId,
	)
	if err != nil {
		return nil, errors.New("can not create bid")
	}
	return
}

func (service *BidService) GetBidsByUsername(
	offset int,
	limit int,
	username string,
) (bids []*model.Bid, err error) {
	user, err := service.userRepo.GetUserByName(username)
	if err != nil {
		return nil, ErrorUserNotFound
	}
	bids, err = service.bidRepo.GetBidsByAuthorId(offset, limit, user.Id)
	if err != nil {
		return nil, errors.New("can not get bids")
	}
	return
}

func (service *BidService) GetBidsByTenderId(
	offset int,
	limit int,
	tenderId uuid.UUID,
) (bids []*model.Bid, err error) {
	_, err = service.tenderRepo.GetTenderById(tenderId)
	if err != nil {
		return nil, ErrorTenderNotFound
	}
	bids, err = service.bidRepo.GetBidsByTenderId(offset, limit, tenderId)
	if err != nil {
		return nil, errors.New("can not get bids")
	}
	return
}

func (service *BidService) GetBidById(id uuid.UUID) (bid *model.Bid, err error) {
	bid, err = service.bidRepo.GetBidById(id)
	if err != nil {
		return nil, ErrorBidNotFound
	}
	return
}

func (service *BidService) UpdateBidStatusById(
	id uuid.UUID, status model.BidStatus,
) (bid *model.Bid, err error) {
	if status != model.CreatedBidStatus &&
		status != model.PublishedBidStatus &&
		status != model.CanceledBidStatus {
		err = errors.New("not allowed status")
		return
	}

	_, err = service.bidRepo.GetBidById(id)
	if err != nil {
		return nil, ErrorBidNotFound
	}
	bid, err = service.bidRepo.UpdateBidStatusById(id, status)
	if err != nil {
		return nil, errors.New("can not update bid")
	}
	return
}

func (service *BidService) EditBidById(
	id uuid.UUID, name string, description string,
) (bid *model.Bid, err error) {
	bid, err = service.bidRepo.GetBidById(id)
	if err != nil {
		return nil, ErrorBidNotFound
	}

	if name == "" {
		name = bid.Name
	}

	if description == "" {
		description = bid.Description
	}

	bid, err = service.bidRepo.EditBidById(id, name, description)
	if err != nil {
		return nil, errors.New("can not edit bid")
	}
	return
}

func (service *BidService) SubmitDecisionById(
	id uuid.UUID, decision string,
) (bid *model.Bid, err error) {
	bid, err = service.bidRepo.GetBidById(id)
	if err != nil {
		return nil, ErrorBidNotFound
	}

	if decision != "Approved" && decision != "Rejected" {
		err = errors.New("now allowed decision")
		return
	}

	var orgId uuid.UUID
	if bid.AuthorType == model.OrgBidAuthorType {
		orgId = bid.AuthorId
	} else {
		orgs, err := service.orgRepo.GetOrganizationsByUserId(bid.AuthorId)
		if err != nil || orgs != nil && len(orgs) == 0 {
			return nil, ErrorUserIsNotOrgResponsible
		}
		orgId = orgs[0].Id
	}
	usersId, err := service.orgRepo.GetResponsibleUsersId(orgId)
	if err != nil {
		return nil, ErrorUserIsNotOrgResponsible
	}

	rejects, approves, err := service.bidRepo.GetDisicions(id)
	if err != nil {
		return nil, ErrorBidNotFound
	}

	if rejects > 0 {
		return nil, errors.New("bid is already rejected")
	}

	quorum := min(3, len(usersId))
	if approves > quorum {
		return nil, errors.New("bid is already approved")
	}

	if decision == "Approved" {
		approves += 1
		closeTender := approves >= quorum
		err = service.bidRepo.UpdateApproves(id, approves, bid.TenderId, closeTender)
		if err != nil {
			return nil, errors.New("can not approve bid")
		}
	} else {
		rejects += 1
		err = service.bidRepo.UpdateRejects(id, rejects)
		if err != nil {
			return nil, errors.New("can not reject bid")
		}
	}

	return
}

func (service *BidService) CreateReviewById(
	id uuid.UUID, description string,
) (bid *model.Bid, err error) {
	_, err = service.bidRepo.GetBidById(id)
	if err != nil {
		return nil, ErrorBidNotFound
	}
	bid, err = service.bidRepo.CreateReviewById(id, description)
	if err != nil {
		return nil, errors.New("can not create review")
	}
	return
}

func (service *BidService) RollbackById(
	id uuid.UUID, version int32,
) (bid *model.Bid, err error) {
	_, err = service.bidRepo.GetBidById(id)
	if err != nil {
		return nil, ErrorBidNotFound
	}
	bid, err = service.bidRepo.RollbackById(id, version)
	if err != nil {
		return nil, errors.New("can not rollback bid")
	}
	return
}

func (service *BidService) GetReviews(
	offset int,
	limit int,
	authorUsername string,
	tender_id uuid.UUID,
) (reviews []*model.Review, err error) {
	_, err = service.tenderRepo.GetTenderById(tender_id)
	if err != nil {
		return nil, ErrorTenderNotFound
	}
	author, err := service.userRepo.GetUserByName(authorUsername)
	if err != nil {
		return nil, ErrorUserNotFound
	}
	reviews, err = service.bidRepo.GetReviews(offset, limit, author.Id, tender_id)
	if err != nil {
		return nil, errors.New("can not get reviews")
	}
	return
}

func (service *BidService) CheckRWRightsByUsername(
	tenderId uuid.UUID, username string,
) error {
	tender, err := service.tenderRepo.GetTenderById(tenderId)
	if err != nil {
		err = ErrorTenderNotFound
		return err
	}

	if username == "" {
		return ErrorUserIsNotOrgResponsible
	}

	user, err := service.userRepo.GetUserByName(username)
	if err != nil {
		err = ErrorUserNotFound
		return err
	}

	usersId, err := service.orgRepo.GetResponsibleUsersId(
		tender.OrganizationId,
	)

	if err != nil {
		err = ErrorUserIsNotOrgResponsible
		return err
	}

	for _, userId := range usersId {
		if userId == user.Id {
			return nil
		}
	}
	return ErrorUserIsNotOrgResponsible
}

func NewService() (service *BidService, err error) {
	tenderRerository, err := tender.NewRepo()
	if err != nil {
		return
	}
	bidRepository, err := bid.NewRepo()
	if err != nil {
		return
	}
	userRepository, err := user.NewRepo()
	if err != nil {
		return
	}
	organizationRepository, err := organization.NewRepo()
	if err != nil {
		return
	}

	service = &BidService{
		tenderRepo: tenderRerository,
		bidRepo:    bidRepository,
		userRepo:   userRepository,
		orgRepo:    organizationRepository,
	}
	return
}
