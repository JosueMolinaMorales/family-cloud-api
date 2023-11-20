package s3

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config/log"
	"github.com/JosueMolinaMorales/family-cloud-api/internal/middleware"
	"github.com/JosueMolinaMorales/family-cloud-api/pkg/error"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Routes returns the routes for the s3 package
func Routes(controller Controller) *chi.Mux {
	r := chi.NewRouter()

	h := &handler{
		controller: controller,
		logger:     log.NewLogger().With(context.Background(), "Version", "1.0.0"),
	}

	r.Use(middleware.AuthMiddlware)
	r.Get("/list", h.ListObjects)
	r.Get("/folder", h.ListFolder)
	r.Get("/folder/size", h.GetFolderSize)
	r.Post("/upload", h.UploadObject)
	r.Get("/download", h.GetObject)

	// Set middleware for error handling
	return r
}

type handler struct {
	controller Controller
	logger     log.Logger
}

func (h *handler) UploadObject(w http.ResponseWriter, r *http.Request) {
	// TODO: Look into DTOs
	// Get the body of the request
	var body struct {
		File string `json:"file"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		error.HandleError(w, r, error.NewRequestError(err, error.BadRequestError, "invalid request body", h.logger))
		return
	}

	// Get the presigned url
	url, err := h.controller.UploadObject(body.File)
	if err != nil {
		error.HandleError(w, r, err)
		return
	}

	// Return the url
	render.JSON(w, r, struct {
		URL string `json:"url"`
	}{
		URL: url,
	})
}

// listObjects lists all the objects in the bucket
// This method gets a list of all the objects in the bucket, then builds a file tree
// based on the keys of the objects. This method allows for collection of size of folders
func (h *handler) ListObjects(w http.ResponseWriter, r *http.Request) {
	tree, err := h.controller.ListObjects()
	if err != nil {
		error.HandleError(w, r, err)
		return
	}

	render.JSON(w, r, tree)
}

// listFolder lists all the items within a prefix in the bucket
// this method returns a file tree of the items within the prefix
// including files and folders. This method does not allow for collection
// of folder sizes
func (h *handler) ListFolder(w http.ResponseWriter, r *http.Request) {
	folder, err := h.controller.ListFolder(r.URL.Query().Get("prefix"))
	if err != nil {
		error.HandleError(w, r, err)
		return
	}
	render.JSON(w, r, folder)
}

// GetFolderSize returns the size of a folder
func (h *handler) GetFolderSize(w http.ResponseWriter, r *http.Request) {
	// Set a timeout for the request
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*500)
	defer cancel()

	ch := make(chan int64)
	go func() {
		size, err := h.controller.GetFolderSize(r.URL.Query().Get("prefix"))
		if err != nil {
			error.HandleError(w, r, err)
			return
		}
		ch <- size
	}()

	for {
		select {
		case <-ctx.Done():
			// Timeout
			error.HandleError(w, r, error.NewRequestError(ctx.Err(), error.BadRequestError, "Timeout", h.logger))
			return
		case size := <-ch:
			render.JSON(w, r, struct {
				Size int64 `json:"size"`
			}{Size: size})
			return
		}
	}
}

func (h *handler) GetObject(w http.ResponseWriter, r *http.Request) {
	// Get the presigned url
	url, err := h.controller.GetObject(r.URL.Query().Get("key"))
	if err != nil {
		error.HandleError(w, r, err)
		return
	}

	// Return the url
	render.JSON(w, r, struct {
		URL string `json:"url"`
	}{
		URL: url,
	})
}

func (h *handler) DeleteObject(w http.ResponseWriter, r *http.Request) {
}

func (h *handler) CreateFolder(w http.ResponseWriter, r *http.Request) {
}
