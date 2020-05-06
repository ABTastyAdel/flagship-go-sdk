package utils

import (
	"errors"
	"testing"

	"github.com/abtasty/flagship-go-sdk/pkg/logging"
)

var testLogger = logging.GetLogger("FS Test")

func TestHandleRecovered(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				HandleRecovered(r, testLogger)
			}
		}()
		panic("test")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				HandleRecovered(r, testLogger)
			}
		}()
		panic(errors.New("test"))
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				HandleRecovered(r, testLogger)
			}
		}()
		panic(false)
	}()
}
