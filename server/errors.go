package server

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/pkg/errors"
)

const (
	pdTypeInvalidCommand    = "invalid_command"
	pdTypeTopicDoesNotExist = "topic_does_not_exist"
)

type problemDetail struct {
	PDType string `json:"type,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func (pd problemDetail) Error() string {
	return fmt.Sprintf("%s: %s", pd.PDType, pd.Detail)
}

// writeErr writes the error's cause to the connection. If the error isn't a
// problemDetail, it's treated as an internal-server-error.
func (svr *Server) writeErr(conn net.Conn, e error) {

	// Get the errors cause.
	cause := errors.Cause(e)
	if cause == nil {
		return
	}

	// If the cause isn't a problemDetail, log it and override the cause.
	if _, ok := cause.(problemDetail); !ok {
		svr.logger.Error().Stack().Err(e).Msg("Internal Server Error")
		cause = problemDetail{"internal_server_error", ""}
	}

	// Write the cause.
	if err := json.NewEncoder(conn).Encode(cause); err != nil {
		svr.logger.Error().Stack().Err(err).Msg("Cannot write to connection")
	}
}
