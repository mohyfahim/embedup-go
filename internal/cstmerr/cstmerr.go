package cstmerr

import (
	"fmt"
)

// BaseError provides a base for custom errors, allowing for wrapped errors.
type BaseError struct {
	Msg string
	Err error // Underlying error
}

func (e *BaseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *BaseError) Unwrap() error {
	return e.Err
}

// ConfigError indicates a problem with configuration.
type ConfigError struct{ BaseError }

func NewConfigError(msg string, underlyingErr error) *ConfigError {
	return &ConfigError{BaseError{Msg: msg, Err: underlyingErr}}
}

// VersionReadError indicates a problem reading the version file.
type VersionReadError struct{ BaseError }

func NewVersionReadError(msg string, underlyingErr error) *VersionReadError {
	return &VersionReadError{BaseError{Msg: msg, Err: underlyingErr}}
}

// TokenReadError (if you were reading token from a file, not from config directly)
type TokenReadError struct{ BaseError }

func NewTokenReadError(msg string, underlyingErr error) *TokenReadError {
	return &TokenReadError{BaseError{Msg: msg, Err: underlyingErr}}
}

// VersionFormatError indicates an invalid version format.
type VersionFormatError struct{ BaseError }

func NewVersionFormatError(msg string, underlyingErr error) *VersionFormatError {
	return &VersionFormatError{BaseError{Msg: msg, Err: underlyingErr}}
}

// APIClientError indicates a general problem with the HTTP client or request creation.
type APIClientError struct{ BaseError }

func NewAPIClientError(underlyingErr error) *APIClientError {
	return &APIClientError{BaseError{Msg: "API client error", Err: underlyingErr}}
}

// APIRequestFailedError indicates an API request returned a non-success status.
type APIRequestFailedError struct {
	BaseError
	StatusCode int
	Message    string // Message from API response body
}

func NewAPIRequestFailedError(statusCode int, message string) *APIRequestFailedError {
	return &APIRequestFailedError{
		BaseError:  BaseError{Msg: fmt.Sprintf("API request failed with status %d", statusCode)},
		StatusCode: statusCode,
		Message:    message,
	}
}
func (e *APIRequestFailedError) Error() string {
	return fmt.Sprintf("%s - %s", e.BaseError.Msg, e.Message)
}

// NoUpdateAvailable is used when the service is already up-to-date.
// This might be better handled by returning (nil, nil) from CheckForUpdates if no update.
type NoUpdateAvailableError struct{ BaseError }

func NewNoUpdateAvailableError() *NoUpdateAvailableError {
	return &NoUpdateAvailableError{BaseError{Msg: "No update available or service up-to-date"}}
}

// DownloadError indicates a problem during file download.
type DownloadError struct{ BaseError }

func NewDownloadError(msg string) *DownloadError {
	return &DownloadError{BaseError{Msg: "Download error: " + msg}}
}

// TimeoutError indicates a timeout during an operation.
type TimeoutError struct{ BaseError }

func NewTimeoutError(underlyingErr error) *TimeoutError {
	return &TimeoutError{BaseError{Msg: "Timeout error", Err: underlyingErr}}
}

// HeadError indicates a problem with the HEAD request.
type HeadError struct{ BaseError }

func NewHeadError(msg string) *HeadError {
	return &HeadError{BaseError{Msg: "Head error: " + msg}}
}

// DecryptionError (if used)
// type DecryptionError struct{ BaseError }
// func NewDecryptionError(msg string, underlyingErr error) *DecryptionError { ... }

// ArchiveError indicates a problem with archive extraction.
type ArchiveError struct{ BaseError }

func NewArchiveError(msg string, underlyingErr error) *ArchiveError {
	return &ArchiveError{BaseError{Msg: "Archive extraction error", Err: underlyingErr}}
}

// ScriptError indicates a problem executing an update script.
type ScriptError struct{ BaseError }

func NewScriptError(msg string, underlyingErr error) *ScriptError {
	return &ScriptError{BaseError{Msg: msg, Err: underlyingErr}}
}

// FileSystemError indicates a general filesystem problem.
type FileSystemError struct{ BaseError }

func NewFileSystemError(msg string) *FileSystemError {
	return &FileSystemError{BaseError{Msg: "Filesystem error: " + msg}}
}

// HexError (if used for decryption key)
// type HexError struct{ BaseError }
// func NewHexError(msg string, underlyingErr error) *HexError { ... }

// FileIOError indicates an I/O problem during file operations.
type FileIOError struct{ BaseError }

func NewFileIOError(msg string, underlyingErr error) *FileIOError {
	return &FileIOError{BaseError{Msg: "I/O error during file operation: " + msg, Err: underlyingErr}}
}

type FileDeleteError struct{ BaseError }

func NewFileDeleteError(msg string, underlyingErr error) *FileDeleteError {
	return &FileDeleteError{BaseError{Msg: "File delete error: " + msg, Err: underlyingErr}}
}

type DBError struct{ BaseError }

func NewDBError(msg string, underlyingErr error) *DBError {
	return &DBError{BaseError{Msg: "Database error: " + msg, Err: underlyingErr}}
}

// DBConnectionError indicates a problem connecting to the database.
type DBConnectionError struct{ BaseError }

func NewDBConnectionError(msg string, underlyingErr error) *DBConnectionError {
	return &DBConnectionError{BaseError{Msg: "DB connection error: " + msg, Err: underlyingErr}}
}

// DBQueryError indicates a problem executing a database query.
type DBQueryError struct{ BaseError }

func NewDBQueryError(msg string, underlyingErr error) *DBQueryError {
	return &DBQueryError{BaseError{Msg: "DB query error: " + msg, Err: underlyingErr}}
}

// DBNotFoundError indicates that a query returned no results when at least one was expected.
type DBNotFoundError struct{ BaseError }

func NewDBNotFoundError(msg string, underlyingErr error) *DBNotFoundError {
	return &DBNotFoundError{BaseError{Msg: "DB not found error: " + msg, Err: underlyingErr}}
}

// DBTransactionError indicates an issue with a database transaction.
type DBTransactionError struct{ BaseError }

func NewDBTransactionError(msg string, underlyingErr error) *DBTransactionError {
	return &DBTransactionError{BaseError{Msg: "DB transaction error: " + msg, Err: underlyingErr}}
}

// TempFileError (if you use temporary files)
// type TempFileError struct{ BaseError }
// func NewTempFileError(msg string, underlyingErr error) *TempFileError { ... }

// You can then use type assertions or `errors.As` to check for specific error types:
// if _, ok := err.(*customerrors.TimeoutError); ok { ... }
// var timeoutErr *customerrors.TimeoutError
// if errors.As(err, &timeoutErr) { ... }

type LinkParseError struct{ BaseError }

func NewLinkParseError(msg string) *LinkParseError {
	return &LinkParseError{BaseError{Msg: "Link Parse error: " + msg}}
}

type RetryError struct{ BaseError }

func NewRetryError(msg string, underlyingError error) *RetryError {
	return &RetryError{BaseError{Msg: "Retry error: " + msg, Err: underlyingError}}
}

type ProcessError struct{ BaseError }

func NewProcessError(msg string, underlyingError error) *ProcessError {
	return &ProcessError{BaseError{Msg: "Process error: " + msg, Err: underlyingError}}
}

const (
	PROCESS_DOWNLOAD_ERROR     = "Process Error in downloading %s"
	PROCESS_HASH_ERROR         = "unable to calculate md5 hash"
	PROCESS_DELETE_ENTITY      = "unable to delete entity"
	PROCESS_DELETE_FILE        = "unable to delete file"
	PROCESS_FIND_ENTITY        = "unable to find entity"
	PROCESS_CREATE_ERROR       = "failed to create entity"
	PROCESS_FIND_DIRECTORY     = "unable to find directories inside of %s"
	PROCESS_FIND_SUB_DIRECTORY = "unable to find subdirectory inside"
	PROCESS_HASH_FIND          = "unable to get hash of file from server"
)
