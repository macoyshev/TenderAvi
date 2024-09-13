package tender

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"avi/internal/api/apierror"
	"avi/internal/model"
	tenderService "avi/internal/service/tender"
)

type TenderRequest struct {
	Name            string                  `json:"name"            validate:"required,max=100"`
	Description     string                  `json:"description"     validate:"required,max=500"`
	ServiceType     model.TenderServiceType `json:"serviceType"     validate:"required,oneof=Construction Delivery Manufacture"`
	OrganizationId  uuid.UUID               `json:"organizationId"  validate:"required,max=100"`
	CreatorUsername string                  `json:"creatorUsername" validate:"required"`
}

type EditTenderRequest struct {
	Name        string                  `json:"name"        validate:"max=100"`
	Description string                  `json:"description" validate:"max=500"`
	ServiceType model.TenderServiceType `json:"serviceType" validate:"oneof=Construction Delivery Manufacture ''"`
}

func CreateTenderHandler(w http.ResponseWriter, r *http.Request) {
	tenderReq := TenderRequest{}
	json.NewDecoder(r.Body).Decode(&tenderReq)
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(tenderReq)
	if err != nil {
		slog.Error(err.Error())
		err = errors.New("incorrect request body")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := tenderService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("tenders service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	tender, err := service.CreateTender(
		tenderReq.Name,
		tenderReq.Description,
		model.TenderServiceType(tenderReq.ServiceType),
		tenderReq.OrganizationId,
		tenderReq.CreatorUsername,
	)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorUserNorFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, tenderService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(tender)
	w.Write(res)
}

func GetTendersHandler(w http.ResponseWriter, r *http.Request) {
	var offset int
	if offsetQ := r.URL.Query().Get("offset"); len(offsetQ) != 0 {
		offset, _ = strconv.Atoi(offsetQ)
	}

	var limit int
	if limitQ := r.URL.Query().Get("limit"); len(limitQ) != 0 {
		limit, _ = strconv.Atoi(limitQ)
	}

	var serviceType model.TenderServiceType
	if serviceTypeQ := r.URL.Query().Get("serviceType"); len(serviceTypeQ) != 0 {
		serviceType = model.TenderServiceType(serviceTypeQ)
	}

	service, err := tenderService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("tenders service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	tenders, err := service.GetTenders(serviceType, offset, limit, "")
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorUserNorFound) {
			httpStatus = http.StatusUnauthorized
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}
	if tenders == nil {
		tenders = []*model.Tender{}
	}
	res, _ := json.Marshal(tenders)
	w.Write(res)
}

func GetMyTendersHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username is not provided")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	var offset int
	if offsetQ := r.URL.Query().Get("offset"); len(offsetQ) != 0 {
		offset, _ = strconv.Atoi(offsetQ)
	}

	var limit int
	if limitQ := r.URL.Query().Get("limit"); len(limitQ) != 0 {
		limit, _ = strconv.Atoi(limitQ)
	}

	service, err := tenderService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("tenders service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	tenders, err := service.GetMyTenders(offset, limit, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorUserNorFound) {
			httpStatus = http.StatusUnauthorized
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}
	if tenders == nil {
		tenders = []*model.Tender{}
	}
	res, _ := json.Marshal(tenders)
	w.Write(res)
}

func GetTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenderId, err := uuid.Parse(chi.URLParam(r, "tenderId"))
	if err != nil {
		err = errors.New("incorrect tender uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := tenderService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("tenders service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	tender, err := service.GetTenderById(tenderId)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorTenderNorFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	username := r.URL.Query().Get("username")
	err = service.CheckReadRightByUsername(tenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorTenderNorFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, tenderService.ErrorUserNorFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, tenderService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(tender.Status)
	w.Write(res)
}

func UpdateTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenderId, err := uuid.Parse(chi.URLParam(r, "tenderId"))
	if err != nil {
		err = errors.New("incorrect tender uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		err := errors.New("status param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return

	}

	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := tenderService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("tenders service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = service.CheckWriteRightByUsername(tenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorTenderNorFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, tenderService.ErrorUserNorFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, tenderService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	tender, err := service.UpdateTenderStatus(tenderId, model.TenderStatus(status))
	if err != nil {
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	res, _ := json.Marshal(tender)
	w.Write(res)
}

func EditTenderHandler(w http.ResponseWriter, r *http.Request) {
	tenderId, err := uuid.Parse(chi.URLParam(r, "tenderId"))
	if err != nil {
		err = errors.New("incorrect tender uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	editTenderReq := EditTenderRequest{}
	json.NewDecoder(r.Body).Decode(&editTenderReq)
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(editTenderReq)
	if err != nil {
		slog.Error(err.Error())
		err = errors.New("incorrect request body")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	if editTenderReq.Name == "" &&
		editTenderReq.Description == "" &&
		editTenderReq.ServiceType == "" {
		err = errors.New("incorrect request body")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := tenderService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("tenders service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = service.CheckWriteRightByUsername(tenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorTenderNorFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, tenderService.ErrorUserNorFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, tenderService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	tender, err := service.UpdateTender(
		tenderId,
		editTenderReq.Name,
		editTenderReq.Description,
		editTenderReq.ServiceType,
	)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorTenderNorFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(tender)
	w.Write(res)
}

func RollbackTenderHandler(w http.ResponseWriter, r *http.Request) {
	tenderId, err := uuid.Parse(chi.URLParam(r, "tenderId"))
	if err != nil {
		err = errors.New("incorrect tender uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		err = errors.New("incorrect version")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := tenderService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("tenders service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = service.CheckWriteRightByUsername(tenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorTenderNorFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, tenderService.ErrorUserNorFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, tenderService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	tender, err := service.RollBackTender(tenderId, int32(version))

	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, tenderService.ErrorTenderNorFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(tender)
	w.Write(res)
}
