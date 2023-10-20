package types

import "time"

// FileItem is an interface for a file or folder
type FileItem interface {
	GetName() string
	GetSize() int64
	GetItems() []FileItem
	GetLastModified() time.Time
	IsDirectory() bool
}

// Folder is a folder
type Folder struct {
	Name         string     `json:"name"`
	Size         int64      `json:"size"`
	Items        []FileItem `json:"items"`
	LastModified time.Time  `json:"lastModified"`
	IsDir        bool       `json:"isDir"`
}

// GetName returns the name of the folder
func (f *Folder) GetName() string {
	return f.Name
}

// GetSize returns the size of the folder
func (f *Folder) GetSize() int64 {
	// Calculate the size of the folder
	var size int64
	for _, item := range f.Items {
		size += item.GetSize()
	}
	return size
}

// GetItems returns the items within the folder
func (f *Folder) GetItems() []FileItem {
	return f.Items
}

// GetLastModified returns the last modified date of the folder
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

// IsDirectory returns true since this is a folder
func (f *Folder) IsDirectory() bool {
	return true
}

// File is a file
type File struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	IsDir        bool      `json:"isDir"`
}

// GetName returns the name of the file
func (f *File) GetName() string {
	return f.Name
}

// GetSize returns the size of the file
func (f *File) GetSize() int64 {
	return f.Size
}

// GetItems returns nil since this is a file
func (f *File) GetItems() []FileItem {
	return nil
}

// GetLastModified returns the last modified date of the file
func (f *File) GetLastModified() time.Time {
	return f.LastModified
}

// IsDirectory returns false since this is a file
func (f *File) IsDirectory() bool {
	return false
}
