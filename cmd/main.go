package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"avi/internal/api/bid"
	"avi/internal/api/tender"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/api", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			res, _ := json.Marshal("ok")
			w.Write(res)
		})
		r.Route("/tenders", func(r chi.Router) {
			r.Get("/", tender.GetTendersHandler)
			r.Post("/new", tender.CreateTenderHandler)
			r.Get("/my", tender.GetMyTendersHandler)
			r.Patch("/{tenderId}/edit", tender.EditTenderHandler)
			r.Get("/{tenderId}/status", tender.GetTenderStatusHandler)
			r.Put("/{tenderId}/status", tender.UpdateTenderStatusHandler)
			r.Put("/{tenderId}/rollback/{version}", tender.RollbackTenderHandler)
		})
		r.Route("/bids", func(r chi.Router) {
			r.Post("/new", bid.CreateBidHandler)
			r.Get("/my", bid.GetMyBidsHandler)
			r.Get("/{tenderId}/list", bid.GetBidsHandler)
			r.Get("/{bidId}/status", bid.GetBidStatusHandler)
			r.Put("/{bidId}/status", bid.UpdateBidStatusHandler)
			r.Patch("/{bidId}/edit", bid.EditBidHandler)
			r.Put("/{bidId}/submit_decision", bid.SumbitDecisionHandler)
			r.Put("/{bidId}/feedback", bid.FeedbackHandler)
			r.Put("/{bidId}/rollback/{version}", bid.RollbackHandler)
			r.Get("/{tenderId}/reviews", bid.GetReviewsHandler)
		})
	})

	serverAdd, ok := os.LookupEnv("SERVER_ADDRESS")
	if !ok || serverAdd == "" {
		errMsg := "no SERVER_ADDRESS environment variable or empty"
		slog.Error(errMsg)
		return
	}
	err := http.ListenAndServe(serverAdd, r)
	if err != nil {
		slog.Error(err.Error())
	}
}
