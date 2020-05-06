package utils

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/abtasty/flagship-go-sdk/pkg/logging"
)

// HandleRecovered logs a recovered panic
func HandleRecovered(r interface{}, logger logging.FlagshipLogger) (err error) {
	switch t := r.(type) {
	case error:
		err = t
	case string:
		err = errors.New(t)
	default:
		err = errors.New("Unexpected error")
	}
	errorMessage := fmt.Sprintf("Flagship SDK is panicking with the error:")
	logger.Error(errorMessage, err)
	logger.Debug(string(debug.Stack()))

	return err
}
