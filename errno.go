package rtm2_sdk

import (
	"fmt"
)

type RTMError struct {
	msg   string
	errno int
}

func (e RTMError) Error() string {
	return fmt.Sprintf("%d: %s", e.errno, e.msg)
}

func newSDKError(errno int, msg string) error {
	return RTMError{errno: errno, msg: msg}
}

var (
	ERR_DISCONNECTED = newSDKError(998, "ERR_SDK_DISCONNECTED")
	ERR_TIMEOUT      = newSDKError(999, "ERR_SDK_TIMEOUT")
)
