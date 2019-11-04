package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	bimg "gopkg.in/h2non/bimg.v1"
)

const (
	Unavailable uint8 = iota
	BadRequest
	NotAllowed
	Unsupported
	Unauthorized
	InternalError
	NotFound
)

var (
	ErrNotFound                = NewError("Not found", NotFound)
	ErrEntityTooLarge          = NewError("Entity too large", BadRequest)
	ErrInvalidApiKey           = NewError("Invalid or missing Authorization key", Unauthorized)
	ErrMethodNotAllowed        = NewError("Method not allowed", NotAllowed)
	ErrRoleUserNotAllowed      = NewError("Role of user not allowed", NotAllowed)
	ErrUnsupportedMedia        = NewError("Unsupported media type", Unsupported)
	ErrOutputFormat            = NewError("Unsupported output image format", BadRequest)
	ErrEmptyBody               = NewError("Empty image", BadRequest)
	ErrMissingParamFile        = NewError("Missing required param: file", BadRequest)
	ErrMissingFilterGroupId    = NewError("Missing required filter: groupId", BadRequest)
	ErrFilterGroupId           = NewError("Filter groupId not the same groupId of  Token User", BadRequest)
	ErrInvalidFilePath         = NewError("Invalid file path", BadRequest)
	ErrInvalidImageURL         = NewError("Invalid image URL", BadRequest)
	ErrMissingImageSource      = NewError("Cannot process the image due to missing or invalid params", BadRequest)
	ErrNotPublicKey            = NewError("Key is not a valid public key", Unauthorized)
	Internal_Server_Error      = NewError("sorry something went wrong", InternalError)
	ErrForbidden               = NewError("Your access was denied", NotAllowed)
	ErrMissingParamEmail       = NewError("Missing required param: email", BadRequest)
	ErrEmailAddressIsInvalid   = NewError("Email address is invalid", BadRequest)
	ErrMissingParamTimeExpires = NewError("Missing required param: timeExpires", BadRequest)
	ErrTimeExpiresIsInvalid    = NewError("timeExpires must be data type int", BadRequest)
	ErrTimeExpiresCannotBeZero = NewError("timeExpires is cannot be zero", BadRequest)
)

type Error struct {
	Message string `json:"message,omitempty"`
	Code    uint8  `json:"code"`
}

func (e Error) JSON() []byte {
	buf, _ := json.Marshal(e)
	return buf
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) HTTPCode() int {
	if e.Code == BadRequest {
		return http.StatusBadRequest
	}
	if e.Code == NotAllowed {
		return http.StatusMethodNotAllowed
	}
	if e.Code == Unsupported {
		return http.StatusUnsupportedMediaType
	}
	if e.Code == InternalError {
		return http.StatusInternalServerError
	}
	if e.Code == Unauthorized {
		return http.StatusUnauthorized
	}
	if e.Code == NotFound {
		return http.StatusNotFound
	}
	return http.StatusServiceUnavailable
}

func NewError(err string, code uint8) Error {
	err = strings.Replace(err, "\n", "", -1)
	return Error{err, code}
}

func replyWithPlaceholder(req *http.Request, w http.ResponseWriter, err Error, o ServerOptions) error {
	image := o.PlaceholderImage

	// Resize placeholder to expected output
	buf, _err := bimg.Resize(o.PlaceholderImage, bimg.Options{
		Force:   true,
		Crop:    true,
		Enlarge: true,
		Width:   parseInt(req.URL.Query().Get("width")),
		Height:  parseInt(req.URL.Query().Get("height")),
		Type:    ImageType(req.URL.Query().Get("type")),
	})

	if _err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\", \"code\": %d}", _err.Error(), BadRequest)))
		return _err
	}

	// Use final response body image
	image = buf

	// Placeholder image response
	w.Header().Set("Content-Type", GetImageMimeType(bimg.DetermineImageType(image)))
	w.Header().Set("Error", string(err.JSON()))
	w.WriteHeader(err.HTTPCode())
	w.Write(image)
	return err
}

func ErrorReply(req *http.Request, w http.ResponseWriter, err Error, o ServerOptions) error {
	// Reply with placeholder if required
	if o.EnablePlaceholder || o.Placeholder != "" {
		return replyWithPlaceholder(req, w, err, o)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPCode())
	w.Write(err.JSON())
	return err
}
