package d1

import (
	"errors"

	"fana-karaoke/internal/domain"
)

const (
	codeUnreachable     = "storage_unreachable"
	codeAuth            = "storage_auth_failed"
	codeLimit           = "storage_limit_reached"
	codeConfiguration   = "storage_configuration_error"
	codeInvalidResponse = "storage_invalid_response"
	codeOversize        = "storage_oversize"
	codeImportFailed    = "storage_import_failed"
	codeNoError         = ""
)

func errorCode(err error) string {
	if err == nil {
		return codeNoError
	}
	var api *apiError
	if errors.As(err, &api) {
		switch api.code {
		case "unauthorized":
			return codeAuth
		case "d1_limit_reached":
			return codeLimit
		case "invalid_request", "invalid_song", "too_large":
			return codeInvalidResponse
		case codeOversize:
			return codeOversize
		default:
			return codeUnreachable
		}
	}
	if errors.Is(err, domain.ErrLibraryTooLarge) {
		return codeOversize
	}
	if errors.Is(err, domain.ErrLibraryInvalid) {
		return codeImportFailed
	}
	return codeUnreachable
}
