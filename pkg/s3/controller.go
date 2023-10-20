package s3

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config"
	"github.com/JosueMolinaMorales/family-cloud-api/pkg/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

// Controller is the interface for the s3 controller
type Controller interface {
	ListObjects() (*types.Folder, error)
	ListFolder(prefix string) (*types.Folder, error)
	GetFolderSize(prefix string) (int64, error)
	GetObject()
	UploadObject()
	DeleteObject()
}

// NewController creates a new controller
func NewController(logger config.Logger, s3Client config.AwsDriver) Controller {
	return &controller{
		logger:   logger,
		s3Client: s3Client,
	}
}

type controller struct {
	logger   config.Logger
	s3Client config.AwsDriver
}

func (c *controller) ListObjects() (*types.Folder, error) {
	bucket := "morales-storage-drive"
	var continuationToken *string
	// Create root folder
	folder := &types.Folder{
		Name:         "/",
		Size:         0,
		Items:        make([]types.FileItem, 0),
		LastModified: time.Now(),
		IsDir:        true,
	}

	for {
		res, err := c.s3Client.ListObjects(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            &bucket,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			panic("failed to list objects, " + err.Error())
		}

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

	return folder, nil
}

func (c *controller) ListFolder(prefix string) (*types.Folder, error) {
	bucket := "morales-storage-drive"

	if prefix != "" {
		prefix = fmt.Sprintf("%s/", prefix)
	}
	res, err := c.s3Client.ListObjects(context.TODO(), &s3.ListObjectsV2Input{
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
	root := &types.Folder{
		Name:         prefix,
		Size:         0,
		Items:        make([]types.FileItem, 0),
		LastModified: time.Now(),
		IsDir:        true,
	}

	// Get the files in this folder
	for _, item := range res.Contents {
		if item.Key == nil {
			continue
		}
		fileParts := strings.Split(*item.Key, "/")
		fileName := fileParts[len(fileParts)-1]
		root.Items = append(root.Items, &types.File{
			Name:         fileName,
			Size:         item.Size,
			LastModified: *item.LastModified,
			IsDir:        false,
		})
	}
	// Get the folders in this folder
	for _, item := range res.CommonPrefixes {
		if item.Prefix == nil {
			continue
		}
		folderParts := strings.Split(*item.Prefix, "/")
		folderName := folderParts[len(folderParts)-2]
		root.Items = append(root.Items, &types.Folder{
			Name:  folderName,
			Size:  0,
			Items: make([]types.FileItem, 0),
			IsDir: true,
		})
	}

	return root, nil
}

func (c *controller) GetObject() {
	panic("not implemented")
}

func (c *controller) GetFolderSize(prefix string) (int64, error) {
	bucket := "morales-storage-drive"

	if prefix != "" {
		prefix = fmt.Sprintf("%s/", prefix)
	}

	size, err := c.calculateFolderSize(bucket, prefix)
	if err != nil {
		panic("failed to calculate folder size, " + err.Error())
	}

	return size, nil
}

func (c *controller) UploadObject() {
	panic("not implemented")
}

func (c *controller) DeleteObject() {
	panic("not implemented")
}

func (c *controller) calculateFolderSize(bucket string, prefix string) (int64, error) {
	var continuationToken *string

	var size int64
	for {
		res, err := c.s3Client.ListObjects(context.TODO(), &s3.ListObjectsV2Input{
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

func buildFileTree(root *types.Folder, path string, size int64, lastModified time.Time) {
	pathParts := strings.Split(path, "/")
	fileName := pathParts[len(pathParts)-1]

	for i, part := range pathParts {
		// Skip the last part since it's the file name
		if i == len(pathParts)-1 {
			continue
		}
		if !containsFolder(root.Items, part) {
			root.Items = append(root.Items, &types.Folder{
				Name:         part,
				Size:         0,
				Items:        make([]types.FileItem, 0),
				LastModified: time.Now(),
				IsDir:        true,
			})
		}

		root = findFolder(root, part)
		// Add the size of the folder
		root.Size += size
	}

	root.Items = append(root.Items, &types.File{
		Name:         fileName,
		Size:         size,
		LastModified: lastModified,
	})
}

func findFolder(root *types.Folder, name string) *types.Folder {
	for _, folder := range root.Items {
		if folder.GetName() == name {
			return folder.(*types.Folder)
		}
	}
	return nil
}

func containsFolder(folders []types.FileItem, name string) bool {
	for _, folder := range folders {
		if folder.GetName() == name {
			return true
		}
	}
	return false
}
