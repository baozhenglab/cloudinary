package cloudinary

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// cleanAssetName returns an asset name from the parent dirname and
// the file name without extension.
// The combination
//   path=/tmp/css/default.css
//   basePath=/tmp/
//   prependPath=new/
// will return
//   new/css/default
func cleanAssetName(path, basePath, prependPath string) string {
	var name string
	path, basePath, prependPath = strings.TrimSpace(path), strings.TrimSpace(basePath), strings.TrimSpace(prependPath)
	basePath, err := filepath.Abs(basePath)
	if err != nil {
		basePath = ""
	}
	apath, err := filepath.Abs(path)
	if err == nil {
		path = apath
	}
	if basePath == "" {
		idx := strings.LastIndex(path, string(os.PathSeparator))
		if idx != -1 {
			idx = strings.LastIndex(path[:idx], string(os.PathSeparator))
		}
		name = path[idx+1:]
	} else {
		// Directory
		name = strings.Replace(path, basePath, "", 1)
		if name[0] == os.PathSeparator {
			name = name[1:]
		}
	}
	if prependPath != "" {
		if prependPath[0] == os.PathSeparator {
			prependPath = prependPath[1:]
		}
		prependPath = EnsureTrailingSlash(prependPath)
	}
	r := prependPath + name[:len(name)-len(filepath.Ext(name))]
	return strings.Replace(r, string(os.PathSeparator), "/", -1)
}

// EnsureTrailingSlash adds a missing trailing / at the end
// of a directory name.
func EnsureTrailingSlash(dirname string) string {
	if !strings.HasSuffix(dirname, "/") {
		dirname += "/"
	}
	return dirname
}

// cleanAssetName returns an asset name from the parent dirname and
// the file name without extension.
// The combination
//   path=/tmp/css/default.css
//   prependPath=new/
// will return
//   new/default
func CleanExtensionNameWithPrepend(path, prependPath string) string {
	var name string
	path, prependPath = strings.TrimSpace(path), strings.TrimSpace(prependPath)
	apath, err := filepath.Abs(path)
	if err == nil {
		path = apath
	}
	idx := strings.LastIndex(path, string(os.PathSeparator))
	if idx == -1 {
		idx = strings.LastIndex(path[:idx], string(os.PathSeparator))
	}
	name = path[idx+1:]
	if prependPath != "" {
		if prependPath[0] == os.PathSeparator {
			prependPath = prependPath[1:]
		}
		prependPath = EnsureTrailingSlash(prependPath)
	}
	r := prependPath + name[:len(name)-len(filepath.Ext(name))]
	return strings.Replace(r, string(os.PathSeparator), "/", -1)
}

func setPublicID(prependPath, fileName string) string {
	idx := strings.LastIndex(fileName, string(os.PathSeparator))
	if idx != -1 {
		idx = strings.LastIndex(fileName[:idx], string(os.PathSeparator))
	}
	return fileName[idx+1:]
}

// setPath to set the PublicID
func setPath(prependPath, fileName string) string {
	return prependPath[1:] + fileName
}

// getAccessURL to get the file URL
func getAccessURL(resType ResourceType, cloudName, publicId, extensionName string) string {
	var t string
	switch resType {
	case PdfType:
		t = "pdf"
	case VideoType:
		t = "video"
	case RawType:
		t = "raw"
	default:
		t = "image"
	}
	// non-image resource PublicID remain extension
	if t != "image" {
		return baseResourceURL + "/" + cloudName + "/" + t + "/" + "upload/" + publicId
	}
	return baseResourceURL + "/" + cloudName + "/" + t + "/" + "upload/" + publicId + "." + extensionName
}

func handleHttpResponse(resp *http.Response) (map[string]interface{}, error) {
	if resp == nil {
		return nil, errors.New("nil http response")
	}
	dec := json.NewDecoder(resp.Body)
	var msg interface{}
	if err := dec.Decode(&msg); err != nil {
		return nil, err
	}
	m := msg.(map[string]interface{})
	if resp.StatusCode != http.StatusOK {
		// JSON error looks like {"error":{"message":"Missing required parameter - public_id"}}
		if e, ok := m["error"]; ok {
			return nil, errors.New(e.(map[string]interface{})["message"].(string))
		}
		return nil, errors.New(resp.Status)
	}
	return m, nil
}
