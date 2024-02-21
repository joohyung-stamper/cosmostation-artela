package custom

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

func AccountExporterFromUndefinedTxMsg(msg *sdktypes.Msg, txHash string) (msgType string, accounts []string) {
	switch msg := (*msg).(type) {

	default:
		// 전체 case에서 이 msg를 찾지 못하였기 때문에 에러 로깅한다.
		msgType = proto.MessageName(msg)
		zap.S().Infof("Undefined msg Type : %T(hash = %s)\n", msg, txHash)
	}

	return
}
