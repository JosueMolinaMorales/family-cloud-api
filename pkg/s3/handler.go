package s3

import (
	"context"
	"net/http"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Routes returns the routes for the s3 package
func Routes(controller Controller) *chi.Mux {
	r := chi.NewRouter()

	h := &handler{
		controller: controller,
		logger:     config.NewLogger().With(context.Background(), "Version", "1.0.0"),
	}

	r.Get("/list", h.ListObjects)
	r.Get("/folder", h.GetFolderSize)
	r.Get("/folder/size", h.GetFolderSize)

	return r
}

type handler struct {
	controller Controller
	logger     config.Logger
}

// listObjects lists all the objects in the bucket
// This method gets a list of all the objects in the bucket, then builds a file tree
// based on the keys of the objects. This method allows for collection of size of folders
func (h *handler) ListObjects(w http.ResponseWriter, r *http.Request) {
	tree, err := h.controller.ListObjects()
	if err != nil {
		panic("failed to list objects, " + err.Error())
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
		panic("failed to list folder, " + err.Error())
	}

	render.JSON(w, r, folder)
}

// GetFolderSize returns the size of a folder
func (h *handler) GetFolderSize(w http.ResponseWriter, r *http.Request) {
	size, err := h.controller.GetFolderSize(r.URL.Query().Get("prefix"))
	if err != nil {
		panic("failed to get folder size, " + err.Error())
	}

	render.JSON(w, r, struct {
		Size int64 `json:"size"`
	}{
		Size: size,
	})
}

func (h *handler) GetObject(w http.ResponseWriter, r *http.Request) {
}

func (h *handler) UploadObject(w http.ResponseWriter, r *http.Request) {
}

func (h *handler) DeleteObject(w http.ResponseWriter, r *http.Request) {
}
