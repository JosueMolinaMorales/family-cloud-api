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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func Routes(r *chi.Mux, logger *config.Logger) {
	r.Route("/s3", func(r chi.Router) {
		r.Get("/list", listObjects)
		r.Get("/folder", listFolder)
		r.Get("/folder/size", getFolderSize)
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
// This method gets a list of all the objects in the bucket, then builds a file tree
// based on the keys of the objects. This method allows for collection of size of folders
func listObjects(w http.ResponseWriter, r *http.Request) {
	cfg, err := aws_config.LoadDefaultConfig(context.TODO(), aws_config.WithSharedConfigProfile("personal"), aws_config.WithRegion("us-east-1"))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)
	bucket := "morales-storage-drive"
	var continuationToken *string
	// Create root folder
	folder := &Folder{
		Name:         "/",
		Size:         0,
		Items:        make([]FileItem, 0),
		LastModified: time.Now(),
		IsDir:        true,
	}

	for {
		res, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            &bucket,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			panic("failed to list objects, " + err.Error())
		}
		fmt.Println(len(res.Contents))
		for _, item := range res.Contents {
			if item.Key == nil {
				continue
			}
			buildFileTree(folder, *item.Key, item.Size, *item.LastModified)
			// add the size of the file
			folder.Size += item.Size
		}
		if !res.IsTruncated {
			break
		} else {
			continuationToken = res.NextContinuationToken
		}
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

func buildFileTree(root *Folder, path string, size int64, lastModified time.Time) {
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
		LastModified: lastModified,
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

// listFolder lists all the items within a prefix in the bucket
// this method returns a file tree of the items within the prefix
// including files and folders. This method does not allow for collection
// of folder sizes
func listFolder(w http.ResponseWriter, r *http.Request) {
	cfg, err := aws_config.LoadDefaultConfig(context.TODO(), aws_config.WithSharedConfigProfile("personal"), aws_config.WithRegion("us-east-1"))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)
	bucket := "morales-storage-drive"

	prefix := r.URL.Query().Get("prefix")

	if prefix != "" {
		prefix = fmt.Sprintf("%s/", prefix)
	}
	res, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:    &bucket,
		Prefix:    &prefix,
		Delimiter: aws.String("/"),
	})
	if err != nil {
		panic("failed to list objects, " + err.Error())
	}

	if prefix == "" {
		prefix = "/"
	}
	root := &Folder{
		Name:         prefix,
		Size:         0,
		Items:        make([]FileItem, 0),
		LastModified: time.Now(),
		IsDir:        true,
	}

	// Get the files in this folder
	fmt.Println(len(res.Contents))
	for _, item := range res.Contents {
		if item.Key == nil {
			continue
		}
		fileParts := strings.Split(*item.Key, "/")
		fileName := fileParts[len(fileParts)-1]
		root.Items = append(root.Items, &File{
			Name:         fileName,
			Size:         item.Size,
			LastModified: *item.LastModified,
			IsDir:        false,
		})
	}
	fmt.Println(len(res.CommonPrefixes))
	// Get the folders in this folder
	for _, item := range res.CommonPrefixes {
		if item.Prefix == nil {
			continue
		}
		folderParts := strings.Split(*item.Prefix, "/")
		folderName := folderParts[len(folderParts)-2]
		root.Items = append(root.Items, &Folder{
			Name:  folderName,
			Size:  0,
			Items: make([]FileItem, 0),
			IsDir: true,
		})
	}

	// Return the results
	render.JSON(w, r, root)
}

// getFolderSize returns the size of a folder given a prefix
func getFolderSize(w http.ResponseWriter, r *http.Request) {
	cfg, err := aws_config.LoadDefaultConfig(context.TODO(), aws_config.WithSharedConfigProfile("personal"), aws_config.WithRegion("us-east-1"))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)
	bucket := "morales-storage-drive"

	prefix := r.URL.Query().Get("prefix")

	if prefix != "" {
		prefix = fmt.Sprintf("%s/", prefix)
	}

	size, err := calculateFolderSize(client, bucket, prefix)
	if err != nil {
		panic("failed to calculate folder size, " + err.Error())
	}

	// Return the results
	render.JSON(w, r, struct {
		Size int64 `json:"size"`
	}{
		Size: size,
	})
}

func calculateFolderSize(client *s3.Client, bucket string, prefix string) (int64, error) {
	var continuationToken *string

	var size int64
	for {
		res, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            &bucket,
			Prefix:            &prefix,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			panic("failed to list objects, " + err.Error())
		}

		// Get the files in this folder
		fmt.Println(len(res.Contents))
		for _, item := range res.Contents {
			if item.Key == nil {
				continue
			}
			size += item.Size
		}

		if !res.IsTruncated {
			break
		} else {
			continuationToken = res.NextContinuationToken
		}
	}

	return size, nil
}
