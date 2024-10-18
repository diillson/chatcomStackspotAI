// llm/utils.go

package llm

import (
	"net"
)

func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}
	netErr, ok := err.(net.Error)
	return ok && (netErr.Timeout() || netErr.Temporary())
}
