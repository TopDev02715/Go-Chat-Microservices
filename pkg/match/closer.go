package match

import (
	"github.com/minghsu0107/go-random-chat/pkg/common"
	"github.com/minghsu0107/go-random-chat/pkg/infra"
)

type InfraCloser struct{}

func NewInfraCloser() common.InfraCloser {
	return &InfraCloser{}
}

func (closer *InfraCloser) Close() error {
	if err := ChatConn.Conn.Close(); err != nil {
		return err
	}
	return infra.RedisClient.Close()
}
