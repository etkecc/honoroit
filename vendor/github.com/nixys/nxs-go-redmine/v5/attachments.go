package redmine

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nixys/nxs-go-redmine/v5/mimereader"
)

/* Get */

// AttachmentObject struct used for attachments get operations
type AttachmentObject struct {
	ID           int64  `json:"id"`
	FileName     string `json:"filename"`
	FileSize     string `json:"filesize"`
	ContentType  string `json:"content_type"`
	Description  string `json:"description"`
	ContentURL   string `json:"content_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Author       IDName `json:"author"`
	CreatedOn    string `json:"created_on"`
}

/* Upload */

// AttachmentUploadObject struct used for attachments upload operations
type AttachmentUploadObject struct {
	ID          *int64 `json:"id,omitempty"`
	Token       string `json:"token"`
	Filename    string `json:"filename"`     // This field fills in AttachmentUpload() function, not by Redmine. User can redefine this value manually
	ContentType string `json:"content_type"` // This field fills in AttachmentUpload() function, not by Redmine. User can redefine this value manually
}

/* Internal types */

type attachmentSingleResult struct {
	Attachment AttachmentObject `json:"attachment"`
}

type attachmentUploadResult struct {
	Upload AttachmentUploadObject `json:"upload"`
}

// AttachmentSingleGet gets single attachment info
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Attachments#GET
func (r *Context) AttachmentSingleGet(id int64) (AttachmentObject, StatusCode, error) {

	var a attachmentSingleResult

	status, err := r.Get(
		&a,
		url.URL{
			Path: "/attachments/" + strconv.FormatInt(id, 10) + ".json",
		},
		http.StatusOK,
	)

	return a.Attachment, status, err
}

// AttachmentUpload uploads file
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_api#Attaching-files
func (r *Context) AttachmentUpload(filePath string) (AttachmentUploadObject, StatusCode, error) {

	var a attachmentUploadResult

	f, err := os.Open(filePath)
	if err != nil {
		return a.Upload, 0, err
	}
	defer f.Close()

	mr := mimereader.New(f)

	status, err := r.uploadFile(
		mr,
		&a,
		url.URL{
			Path: "/uploads.json",
		},
		http.StatusCreated,
	)
	if err != nil {
		return a.Upload, status, err
	}

	a.Upload.ContentType = mr.DetectContentType()
	a.Upload.Filename = filepath.Base(filePath)

	return a.Upload, status, nil
}

// AttachmentUploadStream uploads file as a stream.
func (r *Context) AttachmentUploadStream(f io.Reader, fileName string) (AttachmentUploadObject, StatusCode, error) {

	var a attachmentUploadResult

	mr := mimereader.New(f)

	status, err := r.uploadFile(
		mr,
		&a,
		url.URL{
			Path: "/uploads.json",
		},
		http.StatusCreated,
	)
	if err != nil {
		return a.Upload, status, err
	}

	a.Upload.ContentType = mr.DetectContentType()
	a.Upload.Filename = filepath.Base(fileName)

	return a.Upload, status, nil
}

func (r *Context) AttachmentDownload(id int64, dstPath string) (AttachmentObject, StatusCode, error) {

	s, o, status, err := r.AttachmentDownloadStream(id)
	if err != nil {
		return AttachmentObject{}, status, err
	}

	lf, err := os.Create(dstPath)
	if err != nil {
		return AttachmentObject{}, status, err
	}
	defer lf.Close()

	if _, err := io.Copy(lf, s); err != nil {
		return AttachmentObject{}, status, err
	}

	return o, status, nil
}

func (r *Context) AttachmentDownloadStream(id int64) (io.ReadCloser, AttachmentObject, StatusCode, error) {

	o, status, err := r.AttachmentSingleGet(id)
	if err != nil {
		return nil, AttachmentObject{}, status, err
	}

	s, status, err := r.downloadFile(o.ContentURL, http.StatusOK)
	if err != nil {
		return nil, AttachmentObject{}, status, err
	}

	return s, o, status, nil
}
