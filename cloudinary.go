package cloudinary

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	goservice "github.com/baozhenglab/go-sdk/v2"
)

//from https://github.com/rootsongjc/cloudinary-go/blob/master/service.go

type cloudinaryService struct {
	// format cloudinary://api_key:api_secret@cloud_name
	uri              string
	cloudName        string
	apiKey           string
	apiSecret        string
	uploadURI        *url.URL     // To upload resources
	adminURI         *url.URL     // To use the admin API
	uploadResType    ResourceType // Upload resource type
	basePathDir      string       // Base path directory
	prependPath      string       // Remote prepend path
	verbose          bool
	simulate         bool // Dry run (NOP)
	keepFilesPattern *regexp.Regexp
}

type service struct {
	uri       string
	cloudName string
	apiKey    string
	apiSecret string
}

// Resource holds information about an image or a raw file.
type Resource struct {
	PublicId     string `json:"public_id"`
	Version      int    `json:"version"`
	ResourceType string `json:"resource_type"` // image or raw
	Size         int    `json:"bytes"`         // In bytes
	Url          string `json:"url"`           // Remote url
	SecureUrl    string `json:"secure_url"`    // Over https
}

type pagination struct {
	NextCursor int64 `json: "next_cursor"`
}

type resourceList struct {
	pagination
	Resources []*Resource `json: "resources"`
}

type ResourceDetails struct {
	PublicId     string     `json:"public_id"`
	Format       string     `json:"format"`
	Version      int        `json:"version"`
	ResourceType string     `json:"resource_type"` // image or raw
	Size         int        `json:"bytes"`         // In bytes
	Width        int        `json:"width"`         // Width
	Height       int        `json:"height"`        // Height
	Url          string     `json:"url"`           // Remote url
	SecureUrl    string     `json:"secure_url"`    // Over https
	Derived      []*Derived `json:"derived"`       // Derived
}

type Derived struct {
	Transformation string `json:"transformation"` // Transformation
	Size           int    `json:"bytes"`          // In bytes
	Url            string `json:"url"`            // Remote url
}

// Upload response after uploading a file.
type uploadResponse struct {
	Id           string `bson:"_id"`
	PublicId     string `json:"public_id"`
	Version      uint   `json:"version"`
	Format       string `json:"format"`
	ResourceType string `json:"resource_type"` // "image" or "raw"
	Size         int    `json:"bytes"`         // In bytes
	Checksum     string // SHA1 Checksum
}

type ResourceType int

const (
	KeyCloudinary   = "cloudinary"
	baseUploadURL   = "https://api.cloudinary.com/v1_1"
	baseResourceURL = "https://res.cloudinary.com"
	imageType       = "image"
	videoType       = "video"
	pdfType         = "image"
	rawType         = "raw"
)

const (
	ImageType ResourceType = iota
	PdfType
	VideoType
	RawType
)

func (s *service) InitFlags() {
	prefix := fmt.Sprintf("%s-", s.Name())
	flag.StringVar(&s.uri, prefix+"uri", "", "URI connect to cloudinary service,require cloudinary:// scheme in URI")
}

func (s *service) Name() string {
	return KeyCloudinary
}

func (s *service) GetPrefix() string {
	return KeyCloudinary
}

func (s *service) Run() error {
	return s.Configure()
}

func (s *service) Get() interface{} {
	cs := &cloudinaryService{
		cloudName:     s.cloudName,
		apiKey:        s.apiKey,
		apiSecret:     s.apiSecret,
		uploadResType: ImageType,
		simulate:      false,
		verbose:       false,
	}

	// Default upload URI to the service. Can change at runtime in the
	// Upload() function for raw file uploading.
	up, err := url.Parse(fmt.Sprintf("%s/%s/image/upload/", baseUploadURL, s.cloudName))
	if err != nil {
		return err
	}
	cs.uploadURI = up
	return cs
}

func (s *service) Configure() error {
	u, err := url.Parse(s.uri)
	if err != nil {
		return err
	}
	if u.Scheme != "cloudinary" {
		return errors.New("Missing cloudinary:// scheme in URI")
	}
	secret, exists := u.User.Password()
	if !exists {
		return errors.New("No API secret provided in URI.")
	}

	s.cloudName = u.Host
	s.apiKey = u.User.Username()
	s.apiSecret = secret
	return nil
}

func (s *service) Stop() <-chan bool {
	c := make(chan bool)
	go func() { c <- true }()
	return c
}

// Verbose activate/desactivate debugging information on standard output.
func (s *cloudinaryService) Verbose(v bool) {
	s.verbose = v
}

// Simulate show what would occur but actualy don't do anything. This is a dry-run.
func (s *cloudinaryService) Simulate(v bool) {
	s.simulate = v
}

// KeepFiles sets a regex pattern of remote public ids that won't be deleted
// by any Delete() command. This can be useful to forbid deletion of some
// remote resources. This regexp pattern applies to both image and raw data
// types.
func (s *cloudinaryService) KeepFiles(pattern string) error {
	if len(strings.TrimSpace(pattern)) == 0 {
		return nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	s.keepFilesPattern = re
	return nil
}

// CloudName returns the cloud name used to access the Cloudinary service.
func (s *cloudinaryService) CloudName() string {
	return s.cloudName
}

// ApiKey returns the API key used to access the Cloudinary service.
func (s *cloudinaryService) ApiKey() string {
	return s.apiKey
}

// DefaultUploadURI returns the default URI used to upload images to the Cloudinary service.
func (s *cloudinaryService) DefaultUploadURI() *url.URL {
	return s.uploadURI
}

// Url returns the complete access path in the cloud to the
// resource designed by publicId or the empty string if
// no match.
func (s *cloudinaryService) URL(publicId string, rtype ResourceType) string {
	path := imageType
	if rtype == PdfType {
		path = pdfType
	} else if rtype == VideoType {
		path = videoType
	} else if rtype == RawType {
		path = rawType
	}
	return fmt.Sprintf("%s/%s/%s/upload/%s", baseResourceURL, s.cloudName, path, publicId)
}

func NewCloudinaryService() goservice.PrefixRunnable {
	return new(service)
}
