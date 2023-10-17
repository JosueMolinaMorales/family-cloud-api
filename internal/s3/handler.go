package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
)

func Routes(r *chi.Mux, logger *config.Logger) {
	r.Route("/s3", func(r chi.Router) {
		r.Get("/list", listObjects)
	})
}

// FileItem is an interface for a file or folder
type FileItem interface {
	GetName() string
	GetSize() int64
	GetItems() []FileItem
	GetLastModified() time.Time
	IsDirectory() bool
}

type Folder struct {
	Name         string     `json:"name"`
	Size         int64      `json:"size"`
	Items        []FileItem `json:"items"`
	LastModified time.Time  `json:"lastModified"`
	IsDir        bool       `json:"isDir"`
}

func (f *Folder) GetName() string {
	return f.Name
}

func (f *Folder) GetSize() int64 {
	// Calculate the size of the folder
	var size int64
	for _, item := range f.Items {
		size += item.GetSize()
	}
	return size
}

func (f *Folder) GetItems() []FileItem {
	return f.Items
}

func (f *Folder) GetLastModified() time.Time {
	// Get the earliest last modified date
	var lastModified time.Time
	for _, item := range f.Items {
		if item.GetLastModified().Before(lastModified) {
			lastModified = item.GetLastModified()
		}
	}
	return lastModified
}

func (f *Folder) IsDirectory() bool {
	return true
}

type File struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	IsDir        bool      `json:"isDir"`
}

func (f *File) GetName() string {
	return f.Name
}

func (f *File) GetSize() int64 {
	return f.Size
}

func (f *File) GetItems() []FileItem {
	return nil
}

func (f *File) GetLastModified() time.Time {
	return f.LastModified
}

func (f *File) IsDirectory() bool {
	return false
}

// listObjects lists all the objects in the bucket
func listObjects(w http.ResponseWriter, r *http.Request) {
	cfg, err := aws_config.LoadDefaultConfig(context.TODO(), aws_config.WithSharedConfigProfile("personal"), aws_config.WithRegion("us-east-1"))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)
	bucket := "morales-storage-drive"
	res, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &bucket,
	})
	if err != nil {
		panic("failed to list objects, " + err.Error())
	}

	// Create root folder
	folder := &Folder{
		Name:         "/",
		Size:         0,
		Items:        make([]FileItem, 0),
		LastModified: time.Now(),
		IsDir:        true,
	}
	fmt.Println(len(res.Contents))
	for _, item := range res.Contents {
		if item.Key == nil {
			continue
		}
		buildFileTree(folder, *item.Key, item.Size)
		// add the size of the file
		folder.Size += item.Size
	}
	// Print the file tree
	json, err := json.Marshal(folder)
	if err != nil {
		panic("failed to marshal json, " + err.Error())
	}

	// Return the file tree
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func buildFileTree(root *Folder, path string, size int64) {
	pathParts := strings.Split(path, "/")
	fileName := pathParts[len(pathParts)-1]

	for i, part := range pathParts {
		// Skip the last part since it's the file name
		if i == len(pathParts)-1 {
			continue
		}
		if !containsFolder(root.Items, part) {
			root.Items = append(root.Items, &Folder{
				Name:         part,
				Size:         0,
				Items:        make([]FileItem, 0),
				LastModified: time.Now(),
				IsDir:        true,
			})
		}

		root = findFolder(root, part)
		// Add the size of the folder
		root.Size += size
	}

	root.Items = append(root.Items, &File{
		Name:         fileName,
		Size:         size,
		LastModified: time.Now(),
	})
}

func findFolder(root *Folder, name string) *Folder {
	for _, folder := range root.Items {
		if folder.GetName() == name {
			return folder.(*Folder)
		}
	}
	return nil
}

func containsFolder(folders []FileItem, name string) bool {
	for _, folder := range folders {
		if folder.GetName() == name {
			return true
		}
	}
	return false
}
