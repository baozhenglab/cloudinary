package cloudinary

import "io"

type CloudinaryService interface {
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
	Upload(path string, data io.Reader, prepend string, randomPublicID bool, rtype ResourceType) (string, error)

	UploadStaticRaw(path string, data io.Reader, prepend string) (string, error)

	UploadStaticImage(path string, data io.Reader, prepend string) (string, error)

	UploadRaw(path string, data io.Reader, prepend string) (string, error)

	UploadImage(path string, data io.Reader, prepend string) (string, error)

	UploadVideo(path string, data io.Reader, prepend string) (string, error)

	UploadPdf(path string, data io.Reader, prepend string) (string, error)

	// Url returns the complete access path in the cloud to the
	// resource designed by publicId or the empty string if
	// no match.
	URL(publicID string, rtype ResourceType) string
}
