package cloudinary

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Upload file to the service. When using a mongoDB database for storing
// file information (such as checksums), the database is updated after
// any successful upload.
func (s *cloudinaryService) uploadFile(fullPath string, data io.Reader, randomPublicId bool) (string, error) {

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)

	// Write public ID
	var publicId string
	if !randomPublicId {
		// publicId = cleanAssetName(fullPath, s.basePathDir, s.prependPath)
		// make the  publictId looks like a regular file path, such as /banners/1.jpg but actually
		// the publicId is banners/1.jpg
		publicId = CleanExtensionNameWithPrepend(fullPath, s.prependPath)
		pi, err := w.CreateFormField("public_id")
		if err != nil {
			return fullPath, err
		}
		pi.Write([]byte(publicId))
	}
	// Write API key
	ak, err := w.CreateFormField("api_key")
	if err != nil {
		return fullPath, err
	}
	ak.Write([]byte(s.apiKey))

	// Write timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	ts, err := w.CreateFormField("timestamp")
	if err != nil {
		return fullPath, err
	}
	ts.Write([]byte(timestamp))

	// Write signature
	hash := sha1.New()
	part := fmt.Sprintf("timestamp=%s%s", timestamp, s.apiSecret)
	if !randomPublicId {
		part = fmt.Sprintf("public_id=%s&%s", publicId, part)
	}
	io.WriteString(hash, part)
	signature := fmt.Sprintf("%x", hash.Sum(nil))

	si, err := w.CreateFormField("signature")
	if err != nil {
		return fullPath, err
	}
	si.Write([]byte(signature))

	// Write file field
	fw, err := w.CreateFormFile("file", fullPath)
	if err != nil {
		return fullPath, err
	}
	if data != nil { // file descriptor given
		tmp, err := ioutil.ReadAll(data)
		if err != nil {
			return fullPath, err
		}
		fw.Write(tmp)
	} else { // no file descriptor, try opening the file
		fd, err := os.Open(fullPath)
		if err != nil {
			return fullPath, err
		}
		defer fd.Close()

		_, err = io.Copy(fw, fd)
		if err != nil {
			return fullPath, err
		}
		log.Printf("Uploading: %s\n", fullPath)
	}
	// Don't forget to close the multipart writer to get a terminating boundary
	w.Close()
	if s.simulate {
		return fullPath, nil
	}

	upURI := s.uploadURI.String()

	if s.uploadResType == PdfType {
		upURI = strings.Replace(upURI, imageType, pdfType, 1)
	} else if s.uploadResType == VideoType {
		upURI = strings.Replace(upURI, imageType, videoType, 1)
	} else if s.uploadResType == RawType {
		upURI = strings.Replace(upURI, imageType, rawType, 1)
	}
	req, err := http.NewRequest("POST", upURI, buf)
	if err != nil {
		return fullPath, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fullPath, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Body is JSON data and looks like:
		// {"public_id":"Downloads/file","version":1369431906,"format":"png","resource_type":"image"}
		dec := json.NewDecoder(resp.Body)
		upInfo := new(uploadResponse)
		if err := dec.Decode(upInfo); err != nil {
			return fullPath, err
		}
		accessURL := getAccessURL(s.uploadResType, s.cloudName, upInfo.PublicId, upInfo.Format)
		log.Printf("URL: %s\n", accessURL)
		return upInfo.PublicId, nil
	} else {
		return fullPath, errors.New("Request error: " + resp.Status)
	}
}

// helpers
func (s *cloudinaryService) UploadStaticRaw(path string, data io.Reader, prepend string) (string, error) {
	return s.Upload(path, data, prepend, false, RawType)
}

func (s *cloudinaryService) UploadStaticImage(path string, data io.Reader, prepend string) (string, error) {
	return s.Upload(path, data, prepend, false, ImageType)
}

func (s *cloudinaryService) UploadRaw(path string, data io.Reader, prepend string) (string, error) {
	return s.Upload(path, data, prepend, false, RawType)
}

func (s *cloudinaryService) UploadImage(path string, data io.Reader, prepend string) (string, error) {
	return s.Upload(path, data, prepend, false, ImageType)
}

func (s *cloudinaryService) UploadVideo(path string, data io.Reader, prepend string) (string, error) {
	return s.Upload(path, data, prepend, false, VideoType)
}

func (s *cloudinaryService) UploadPdf(path string, data io.Reader, prepend string) (string, error) {
	return s.Upload(path, data, prepend, false, PdfType)
}

// Upload a file or a set of files to the cloud. The path parameter is
// a file location or a directory. If the source path is a directory,
// all files are recursively uploaded to Cloudinary.
//
// In order to upload content, path is always required (used to get the
// directory name or resource name if randomPublicId is false) but data
// can be nil. If data is non-nil the content of the file will be read
// from it. If data is nil, the function will try to open filename(s)
// specified by path.
//
// If ramdomPublicId is true, the service generates a unique random public
// id. Otherwise, the resource's public id is computed using the absolute
// path of the file.
//
// Set rtype to the target resource type, e.g. image or raw file.
//
// For example, a raw file /tmp/css/default.css will be stored with a public
// name of css/default.css (raw file keeps its extension), but an image file
// /tmp/images/logo.png will be stored as images/logo.
//
// The function returns the public identifier of the resource.
func (s *cloudinaryService) Upload(path string, data io.Reader, prepend string, randomPublicId bool, rtype ResourceType) (string, error) {
	s.uploadResType = rtype
	s.basePathDir = ""
	s.prependPath = prepend
	if data == nil {
		info, err := os.Stat(path)
		if err != nil {
			return path, err
		}

		if info.IsDir() {
			s.basePathDir = path
			if err := filepath.Walk(path, s.walkIt); err != nil {
				return path, err
			}
		} else {
			return s.uploadFile(path, nil, randomPublicId)
		}
	} else {
		return s.uploadFile(path, data, randomPublicId)
	}
	return path, nil
}

func (s *cloudinaryService) walkIt(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}
	if _, err := s.uploadFile(path, nil, false); err != nil {
		return err
	}
	return nil
}
