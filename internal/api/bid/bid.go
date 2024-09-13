package bid

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
	bidService "avi/internal/service/bid"
)

type CreateBidRequest struct {
	Name        string              `json:"name"        validate:"required,max=100"`
	Description string              `json:"description" validate:"required,max=500"`
	TenderId    uuid.UUID           `json:"tenderId"    validate:"required,max=100"`
	AuthorType  model.BidAuthorType `json:"authorType"  validate:"required,oneof=User Organization"`
	AuthorId    uuid.UUID           `json:"authorID"    validate:"required,max=100"`
}

type EditBidRequest struct {
	Name        string `json:"name"        validate:"max=100"`
	Description string `json:"description" validate:"max=500"`
}

func CreateBidHandler(w http.ResponseWriter, r *http.Request) {
	bidReq := CreateBidRequest{}
	json.NewDecoder(r.Body).Decode(&bidReq)
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(bidReq)
	if err != nil {
		slog.Error(err.Error())
		err = errors.New("incorrect request body")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	bid, err := service.CreateBid(
		bidReq.Name,
		bidReq.Description,
		bidReq.TenderId,
		bidReq.AuthorType,
		bidReq.AuthorId,
	)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorUserNotFound) ||
			errors.Is(err, bidService.ErrorOrgNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, bidService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bid)
	w.Write(res)
}

func GetMyBidsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
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

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	bids, err := service.GetBidsByUsername(offset, limit, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorUserNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bids)
	w.Write(res)
}

func GetBidsHandler(w http.ResponseWriter, r *http.Request) {
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

	var offset int
	if offsetQ := r.URL.Query().Get("offset"); len(offsetQ) != 0 {
		offset, _ = strconv.Atoi(offsetQ)
	}

	var limit int
	if limitQ := r.URL.Query().Get("limit"); len(limitQ) != 0 {
		limit, _ = strconv.Atoi(limitQ)
	}

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = service.CheckRWRightsByUsername(tenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, bidService.ErrorUserNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, bidService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	bids, err := service.GetBidsByTenderId(offset, limit, tenderId)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bids)
	w.Write(res)
}

func GetBidStatusHandler(w http.ResponseWriter, r *http.Request) {
	bidId, err := uuid.Parse(chi.URLParam(r, "bidId"))
	if err != nil {
		err = errors.New("incorrect bid uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	bid, err := service.GetBidById(bidId)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	err = service.CheckRWRightsByUsername(bid.TenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, bidService.ErrorUserNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, bidService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bid.Status)
	w.Write(res)
}

func UpdateBidStatusHandler(w http.ResponseWriter, r *http.Request) {
	bidId, err := uuid.Parse(chi.URLParam(r, "bidId"))
	if err != nil {
		err = errors.New("incorrect bid uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		err := errors.New("status param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	bid, err := service.GetBidById(bidId)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	err = service.CheckRWRightsByUsername(bid.TenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, bidService.ErrorUserNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, bidService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	bid, err = service.UpdateBidStatusById(
		bidId, model.BidStatus(status),
	)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bid)
	w.Write(res)
}

func EditBidHandler(w http.ResponseWriter, r *http.Request) {
	bidId, err := uuid.Parse(chi.URLParam(r, "bidId"))
	if err != nil {
		err = errors.New("incorrect bid uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	bidReq := EditBidRequest{}
	json.NewDecoder(r.Body).Decode(&bidReq)
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(bidReq)
	if err != nil {
		slog.Error(err.Error())
		err = errors.New("incorrect request body")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	if bidReq.Name == "" && bidReq.Description == "" {
		err = errors.New("incorrect request body")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	bid, err := service.GetBidById(bidId)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	err = service.CheckRWRightsByUsername(bid.TenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, bidService.ErrorUserNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, bidService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	bid, err = service.EditBidById(
		bidId, bidReq.Name, bidReq.Description,
	)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bid)
	w.Write(res)
}

func SumbitDecisionHandler(w http.ResponseWriter, r *http.Request) {
	bidId, err := uuid.Parse(chi.URLParam(r, "bidId"))
	if err != nil {
		err = errors.New("incorrect bid uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	decision := r.URL.Query().Get("decision")
	if decision == "" {
		err := errors.New("decision param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}
	if decision != "Approved" && decision != "Rejected" {
		err := errors.New("not allowed decision")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	bid, err := service.GetBidById(bidId)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	err = service.CheckRWRightsByUsername(bid.TenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, bidService.ErrorUserNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, bidService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	bid, err = service.SubmitDecisionById(bidId, decision)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bid)
	w.Write(res)
}

func FeedbackHandler(w http.ResponseWriter, r *http.Request) {
	bidId, err := uuid.Parse(chi.URLParam(r, "bidId"))
	if err != nil {
		err = errors.New("incorrect bid uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	feedback := r.URL.Query().Get("bidFeedback")
	if feedback == "" {
		err := errors.New("bidFeedback param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	bid, err := service.CreateReviewById(bidId, feedback)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bid)
	w.Write(res)
}

func RollbackHandler(w http.ResponseWriter, r *http.Request) {
	bidId, err := uuid.Parse(chi.URLParam(r, "bidId"))
	if err != nil {
		err = errors.New("incorrect bid uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		err := errors.New("username param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		err = errors.New("incorrect version")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	bid, err := service.GetBidById(bidId)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	err = service.CheckRWRightsByUsername(bid.TenderId, username)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, bidService.ErrorUserNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, bidService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	bid, err = service.RollbackById(bidId, int32(version))
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorBidNotFound) {
			httpStatus = http.StatusNotFound
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bid)
	w.Write(res)
}

func GetReviewsHandler(w http.ResponseWriter, r *http.Request) {
	tenderId, err := uuid.Parse(chi.URLParam(r, "tenderId"))
	if err != nil {
		err = errors.New("incorrect bid uuid")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	reqUsername := r.URL.Query().Get("requesterUsername")
	if reqUsername == "" {
		err := errors.New("requesterUsername param is required")
		apierror.HandleError(w, r, err, http.StatusBadRequest)
		return
	}

	authorUsername := r.URL.Query().Get("authorUsername")
	if authorUsername == "" {
		err := errors.New("authorUsername param is required")
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

	service, err := bidService.NewService()
	if err != nil {
		slog.Error(err.Error())
		err := errors.New("bids service creation failed")
		apierror.HandleError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = service.CheckRWRightsByUsername(tenderId, reqUsername)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, bidService.ErrorUserNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, bidService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	bid, err := service.GetReviews(offset, limit, authorUsername, tenderId)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if errors.Is(err, bidService.ErrorTenderNotFound) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, bidService.ErrorUserNotFound) {
			httpStatus = http.StatusUnauthorized
		}
		if errors.Is(err, bidService.ErrorUserIsNotOrgResponsible) {
			httpStatus = http.StatusForbidden
		}
		apierror.HandleError(w, r, err, httpStatus)
		return
	}

	res, _ := json.Marshal(bid)
	w.Write(res)
}
