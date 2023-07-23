package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate this module. Any variable of this type will have access
// to all the methods with the receiver *Tools
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

// RandomString generates a random string of length n.
//
// It takes an integer parameter n, which specifies the length of the string to be generated.
// The function returns a string.
func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}

	return string(s)
}

// UploadedFile is a struct used to save information about the uploaded file
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

// UploadOneFile uploads a single file from the given http.Request to the specified upload directory.
//
// Parameters:
// - r: The http.Request containing the file to upload.
// - uploadDir: The directory where the uploaded file will be stored.
// - rename: An optional boolean flag indicating whether to rename the uploaded file.
//
// Returns:
// - *UploadedFile: The uploaded file information.
// - error: An error if the upload fails.
func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	files, err := t.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}

	return files[0], nil
}

// UploadFiles uploads files from a request to the specified upload directory.
// It allows for optional file renaming.
//
// Parameters:
// - r: the http.Request object containing the files
// - uploadDir: the directory to upload the files to
// - rename: optional boolean value indicating whether to rename the files
//
// Return:
// - []*UploadedFile: a slice of UploadedFile pointers representing the uploaded files
// - error: an error if any occurred during the file upload process
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}

	err := r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("max file size exceeded!!!")
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				buffer := make([]byte, 512)
				_, err = infile.Read(buffer)
				if err != nil {
					return nil, err
				}

				// check to see if file type is permitted
				allowed := false
				fileType := http.DetectContentType(buffer)

				if len(t.AllowedFileTypes) > 0 {
					for _, t := range t.AllowedFileTypes {
						if strings.EqualFold(t, fileType) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("file type not permitted")
				}

				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				uploadedFile.OriginalFileName = hdr.Filename

				var outFile *os.File
				defer outFile.Close()

				if outFile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outFile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = fileSize
				}

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil

			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	return uploadedFiles, nil
}
