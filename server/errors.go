package server

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/pkg/errors"
)

const (
	pdTypeInvalidCommand    = "invalid_command"
	pdTypeInvalidOperation  = "invalid_operation"
	pdTypeTopicDoesNotExist = "topic_does_not_exist"
)

type problemDetail struct {
	PDType string `json:"type"`
	Detail string `json:"detail"`
}

func (pd problemDetail) Error() string {
	return fmt.Sprintf("%s: %s", pd.PDType, pd.Detail)
}

func (svr *Server) writeErr(conn net.Conn, e error) {

	cause := errors.Cause(e)

	if _, ok := cause.(problemDetail); !ok {
		svr.logger.Error().Stack().Err(e).Msg("Internal Server Error")
		cause = problemDetail{"internal_server_error", ""}
	}

	if err := json.NewEncoder(conn).Encode(cause); err != nil {
		svr.logger.Error().Stack().Err(err).Msg("Cannot write to connection")
	}
}