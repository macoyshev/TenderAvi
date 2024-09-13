package tender

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/google/uuid"

	"avi/internal/model"
	"avi/internal/repository/organization"
	"avi/internal/repository/tender"
	"avi/internal/repository/user"
)

var ErrorUserNorFound = errors.New("user does not exist")
var ErrorUserIsNotOrgResponsible = errors.New("user is not organization responsible")
var ErrorTenderNorFound = errors.New("tender does not exist")

type TenderService struct {
	tenderRepo *tender.TenderRepo
	orgRepo    *organization.OrganizationRepo
	userRepo   *user.UserRepo
}

func (service *TenderService) CreateTender(
	name string,
	description string,
	serviceType model.TenderServiceType,
	organizarionId uuid.UUID,
	createUsername string,
) (tender *model.Tender, err error) {
	user, err := service.userRepo.GetUserByName(createUsername)
	if err != nil {
		err = ErrorUserNorFound
		return
	}
	orgs, err := service.orgRepo.GetOrganizationsByUserId(user.Id)
	if err != nil || orgs == nil {
		err = ErrorUserIsNotOrgResponsible
		return
	}
	isOrgResponsible := false
	for _, org := range orgs {
		if organizarionId == org.Id {
			isOrgResponsible = true
			break
		}
	}
	if !isOrgResponsible {
		err = ErrorUserIsNotOrgResponsible
		return
	}

	tender, err = service.tenderRepo.CreateTender(
		name, description, serviceType, organizarionId, user.Id,
	)
	if err != nil {
		err = errors.New("tender creation failed, check fields")
		return
	}

	return
}

func (service *TenderService) GetMyTenders(
	offset int,
	limit int,
	username string,
) ([]*model.Tender, error) {
	user, err := service.userRepo.GetUserByName(username)
	if err != nil {
		err = ErrorUserNorFound
		return nil, err
	}

	tenders, err := service.tenderRepo.GetTendersByUserId(
		offset, limit, user.Id,
	)
	if err != nil {
		err = errors.New("can not get tenders")
		return nil, err
	}

	return tenders, err
}

func (service *TenderService) GetTenders(
	serviceType model.TenderServiceType,
	offset int,
	limit int,
	username string,
) ([]*model.Tender, error) {
	var offsetFlt, limitFlt, serviceTypeFlt string
	var orgsId []uuid.UUID

	if offset > 0 {
		offsetFlt = strconv.Itoa(offset)
	}
	if limit > 0 {
		limitFlt = strconv.Itoa(limit)
	}
	if username != "" {
		user, err := service.userRepo.GetUserByName(username)
		if err != nil {
			err = ErrorUserNorFound
			return nil, err
		}

		orgs, err := service.orgRepo.GetOrganizationsByUserId(user.Id)
		if err != nil {
			err = errors.New("organization does not exist")
			return nil, err
		}
		orgsId = make([]uuid.UUID, len(orgs))
		for _, org := range orgs {
			orgsId = append(orgsId, org.Id)
		}
	}
	if serviceType != "" &&
		serviceType != model.TenderServiceTypeConstruction &&
		serviceType != model.TenderServiceTypeDelivery &&
		serviceType != model.TenderServiceTypeManufacture {
		err := errors.New("not allowed serviceType")
		return nil, err
	}
	serviceTypeFlt = string(serviceType)

	tenders, err := service.tenderRepo.GetTenders(
		serviceTypeFlt,
		offsetFlt,
		limitFlt,
		orgsId,
	)
	if err != nil {
		slog.Info(err.Error())
		err = errors.New("can not get tenders")
		return nil, err
	}

	return tenders, err
}

func (service *TenderService) GetTenderById(id uuid.UUID) (*model.Tender, error) {
	tender, err := service.tenderRepo.GetTenderById(id)
	if err != nil {
		err = ErrorTenderNorFound
		return nil, err
	}
	return tender, err
}

func (service *TenderService) UpdateTenderStatus(
	id uuid.UUID, status model.TenderStatus,
) (*model.Tender, error) {
	if status != model.TenderStatusClosed &&
		status != model.TenderStatusCreated &&
		status != model.TenderStatusPublished {
		return nil, errors.New("not allowed status")
	}

	tender, err := service.tenderRepo.GetTenderById(id)
	if err != nil {
		return tender, ErrorTenderNorFound
	}
	tender.Status = status
	tenderUpd, err := service.tenderRepo.UpdateTender(tender)
	if err != nil {
		return nil, errors.New("can not update tender")
	}

	return tenderUpd, nil
}

func (service *TenderService) UpdateTender(
	id uuid.UUID,
	name string,
	description string,
	serviceType model.TenderServiceType,
) (*model.Tender, error) {
	tender, err := service.tenderRepo.GetTenderById(id)
	if err != nil {
		return tender, ErrorTenderNorFound
	}
	if name != "" {
		tender.Name = name
	}
	if description != "" {
		tender.Description = description
	}
	if serviceType != "" {
		if serviceType != model.TenderServiceTypeConstruction &&
			serviceType != model.TenderServiceTypeDelivery &&
			serviceType != model.TenderServiceTypeManufacture {
			err = errors.New("not allowed service type")
			return nil, err
		}
		tender.ServiceType = serviceType
	}
	tenderUpd, err := service.tenderRepo.UpdateTender(tender)
	if err != nil {
		return nil, errors.New("can not update tender")
	}
	return tenderUpd, err
}

func (service *TenderService) RollBackTender(id uuid.UUID, version int32) (*model.Tender, error) {
	tender, err := service.tenderRepo.GetTenderById(id)
	if err != nil {
		return nil, ErrorTenderNorFound
	}
	tender, err = service.tenderRepo.RollBackTender(tender.Id, version)
	if err != nil {
		return nil, errors.New("cat not rollback tender")
	}
	return tender, nil
}

func (service *TenderService) CheckReadRightByUsername(
	tenderId uuid.UUID, username string,
) error {
	tender, err := service.tenderRepo.GetTenderById(tenderId)
	if err != nil {
		err = ErrorTenderNorFound
		return err
	}
	if tender.Status == model.TenderStatusPublished {
		return nil
	}

	if username == "" {
		return ErrorUserIsNotOrgResponsible
	}

	user, err := service.userRepo.GetUserByName(username)
	if err != nil {
		err = ErrorUserNorFound
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

func (service *TenderService) CheckWriteRightByUsername(
	tenderId uuid.UUID, username string,
) error {
	tender, err := service.tenderRepo.GetTenderById(tenderId)
	if err != nil {
		err = ErrorTenderNorFound
		return err
	}

	if username == "" {
		return ErrorUserIsNotOrgResponsible
	}

	user, err := service.userRepo.GetUserByName(username)
	if err != nil {
		err = ErrorUserNorFound
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

func NewService() (service *TenderService, err error) {
	tenderRerository, err := tender.NewRepo()
	if err != nil {
		return
	}
	organizationRepository, err := organization.NewRepo()
	if err != nil {
		return
	}
	userRepository, err := user.NewRepo()

	service = &TenderService{
		tenderRepo: tenderRerository,
		orgRepo:    organizationRepository,
		userRepo:   userRepository,
	}
	return
}
