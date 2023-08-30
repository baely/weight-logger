package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"

	"github.com/baely/weightloss-tracker/internal/database"
	"github.com/baely/weightloss-tracker/internal/integrations/apple"
	"github.com/baely/weightloss-tracker/internal/integrations/gcs"
	"github.com/baely/weightloss-tracker/internal/integrations/meta"
	"github.com/baely/weightloss-tracker/internal/integrations/ntfy"
	"github.com/baely/weightloss-tracker/internal/util"
)

type Server struct {
	s http.Server
}

func NewServer() (*Server, error) {
	s := Server{}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	addr := fmt.Sprintf(":%s", port)

	r := chi.NewRouter()

	r.Get("/", s.GetIndex)
	r.Post("/data", s.PostData)
	r.Get("/privacy-policy", s.GetPrivacyPolicy)
	r.Get("/trigger-post", s.TriggerPost)
	r.Get("/refresh-token", s.RefreshToken)
	r.Get("/new-token", s.NewLongToken)
	r.Get("/latest-image", s.LatestImage)

	s.s = http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &s, nil
}

func (s *Server) Run() {
	if err := s.s.ListenAndServe(); err != nil {
		panic(err)
	}
}

func (s *Server) GetIndex(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("hello, world"))
	s.LatestImage(w, r)
	return
}

func (s *Server) PostData(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading from request body", http.StatusBadRequest)
	}

	buf := bytes.NewBuffer(b)
	err = gcs.UploadFile(util.PrivateBucket, fmt.Sprintf("%s_data.json", time.Now().String()), buf)
	if err != nil {
		fmt.Println("error saving raw data to bucket:", err)
	}

	export := apple.Export{}
	err = json.Unmarshal(b, &export)
	if err != nil {
		http.Error(w, "error unmarshalling request body", http.StatusBadRequest)
		return
	}

	documents := database.ExportToDocuments(export.Data)
	go database.InsertOrUpdateDocuments(documents)
}

func (s *Server) GetPrivacyPolicy(w http.ResponseWriter, r *http.Request) {
	policy := meta.Policy()
	w.Write([]byte(policy))
}

func (s *Server) TriggerPost(w http.ResponseWriter, r *http.Request) {
	token, err := database.GetToken()
	if err != nil {
		fmt.Println("error getting long token:", err)
		return
	}

	pageId, busToken, err := meta.GetDetails(token.Token)
	if err != nil {
		fmt.Println("error getting details:", err)
		return
	}

	igId, err := meta.BusinessAccount(pageId, busToken)
	if err != nil {
		fmt.Println("error getting business account:", err)
		return
	}

	// Hack to avoid loading tz files
	date := time.Now().Add(10*time.Hour).AddDate(0, 0, -1).Format("2006-01-02")
	file := fmt.Sprintf("https://storage.googleapis.com/res.xbd.au/weightlog/%s.jpg", date)
	containerId, err := meta.CreateContainer(igId, file, date, busToken)
	if err != nil {
		fmt.Println("error creating container:", err)
		_ = ntfy.Notify(fmt.Sprintf("Error creating Instagram container:\n%s", err))
		return
	}

	err = meta.PublishContent(igId, containerId, busToken)
	if err != nil {
		fmt.Println("error publishing content:", err)
		_ = ntfy.Notify(fmt.Sprintf("Error publishing Instagram content:\n%s", err))
		return
	}
}

func (s *Server) RefreshToken(w http.ResponseWriter, r *http.Request) {
	oldToken, err := database.GetToken()
	if err != nil {
		fmt.Println("error retrieving old token:", oldToken)
		return
	}

	longToken, err := meta.GetLongToken(oldToken.Token)
	if err != nil {
		fmt.Println("error getting long token:", err)
		return
	}

	t := database.TokenDocument{Token: longToken}
	err = t.InsertOrUpdate()
	if err != nil {
		fmt.Println("error saving token to firestore:", err)
		return
	}

	w.Write([]byte("Successfully updated long token"))
}

func (s *Server) NewLongToken(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	code := query.Get("code")

	if code == "" {
		authUrl := meta.AuthUrl()
		page := fmt.Sprintf("<html><body><a href=\"%s\">%s</a></body></html>", authUrl, authUrl)
		w.Write([]byte(page))
		return
	}

	token, err := meta.GetToken(code)
	if err != nil {
		fmt.Println("error getting token:", err)
		return
	}

	longToken, err := meta.GetLongToken(token)
	if err != nil {
		fmt.Println("error getting long token:", err)
		return
	}

	t := database.TokenDocument{Token: longToken}
	err = t.InsertOrUpdate()
	if err != nil {
		fmt.Println("error saving token to firestore:", err)
		return
	}

	w.Write([]byte("Successfully set long token"))
}

func (s *Server) LatestImage(w http.ResponseWriter, r *http.Request) {
	date := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	fileName := fmt.Sprintf("weightlog/%s.jpg", date)
	file, err := gcs.ReadFile(util.ResourceBucket, fileName)
	if err != nil {
		fmt.Println("error getting gcs file:", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("error reading gcs file:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(b)
}
