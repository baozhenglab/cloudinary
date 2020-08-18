package cloudinary

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Delete deletes a resource uploaded to Cloudinary.
func (s *cloudinaryService) Delete(publicId, prepend string, rtype ResourceType) error {
	// TODO: also delete resource entry from database (if used)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	data := url.Values{
		"api_key":   []string{s.apiKey},
		"public_id": []string{prepend + publicId},
		"timestamp": []string{timestamp},
	}
	if s.keepFilesPattern != nil {
		if s.keepFilesPattern.MatchString(prepend + publicId) {
			fmt.Println("keep")
			return nil
		}
	}
	if s.simulate {
		fmt.Println("ok")
		return nil
	}

	// Signature
	hash := sha1.New()
	part := fmt.Sprintf("public_id=%s&timestamp=%s%s", prepend+publicId, timestamp, s.apiSecret)
	io.WriteString(hash, part)
	data.Set("signature", fmt.Sprintf("%x", hash.Sum(nil)))

	rt := imageType
	if rtype == RawType {
		rt = rawType
	}
	resp, err := http.PostForm(fmt.Sprintf("%s/%s/%s/destroy/", baseUploadURL, s.cloudName, rt), data)
	if err != nil {
		return err
	}

	m, err := handleHttpResponse(resp)
	if err != nil {
		return err
	}
	if e, ok := m["result"]; ok {
		fmt.Println(e.(string))
	}
	return nil
}

func (s *cloudinaryService) Rename(publicID, toPublicID, prepend string, rtype ResourceType) error {
	publicID = strings.TrimPrefix(publicID, "/")
	toPublicID = strings.TrimPrefix(toPublicID, "/")
	timestamp := fmt.Sprintf(`%d`, time.Now().Unix())
	data := url.Values{
		"api_key":        []string{s.apiKey},
		"from_public_id": []string{prepend + publicID},
		"timestamp":      []string{timestamp},
		"to_public_id":   []string{prepend + toPublicID},
	}
	// Signature
	hash := sha1.New()
	part := fmt.Sprintf("from_public_id=%s&timestamp=%s&to_public_id=%s%s", prepend+publicID, timestamp, toPublicID, s.apiSecret)
	io.WriteString(hash, part)
	data.Set("signature", fmt.Sprintf("%x", hash.Sum(nil)))

	rt := imageType
	if rtype == RawType {
		rt = rawType
	}
	resp, err := http.PostForm(fmt.Sprintf("%s/%s/%s/rename", baseUploadURL, s.cloudName, rt), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil
}
