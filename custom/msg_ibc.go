package custom

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	mbltypes "github.com/cosmostation/mintscan-backend-library/types"
	"go.uber.org/zap"

	//ibc
	interchainaccountstypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

const (
	// ibc (1)
	IBCTransferMsgTransfer = "ibctransfer/transfer"

	// IBC 02-client (4)
	IBCClientMsgCreateClient       = "ibcclient/create_client"
	IBCClientMsgUpdateClient       = "ibcclient/update_client"
	IBCClientMsgUpgradeClient      = "ibcclient/upgrade_client"
	IBCClientMsgSubmitMisbehaviour = "ibcclient/submit_misbehaviour"

	// IBC 03 connection (4)
	IBCConnectionMsgConnectionOpenInit    = "ibcconnection/connection_open_init"
	IBCConnectionMsgConnectionOpenTry     = "ibcconnection/connection_open_try"
	IBCConnectionMsgConnectionOpenAck     = "ibcconnection/connection_open_ack"
	IBCConnectionMsgConnectionOpenConfirm = "ibcconnection/connection_open_confirm"

	// IBC 04 channel (10)
	IBCChannelMsgChannelOpenInit     = "ibcchannel/channel_open_init"
	IBCChannelMsgChannelOpenTry      = "ibcchannel/channel_open_try"
	IBCChannelMsgChannelOpenAck      = "ibcchannel/channel_open_ack"
	IBCChannelMsgChannelOpenConfirm  = "ibcchannel/channel_open_confirm"
	IBCChannelMsgChannelCloseInit    = "ibcchannel/channel_close_init"
	IBCChannelMsgChannelCloseConfirm = "ibcchannel/channel_close_confirm"
	IBCChannelMsgRecvPacket          = "ibcchannel/recv_packet"
	IBCChannelMsgTimeout             = "ibcchannel/timeout"
	IBCChannelMsgTimeoutOnClose      = "ibcchannel/timeout_on_close"
	IBCChannelMsgAcknowledgement     = "ibcchannel/acknowledgement"
)

type txParser func(msg *sdktypes.Msg, txHash string) (msgType string, accounts []string)

var CustomTxParsers = make([]txParser, 0)

func init() {
	CustomTxParsers = append(CustomTxParsers, AccountExporterFromIBCMsg)
	CustomTxParsers = append(CustomTxParsers, AccountExporterFromUndefinedTxMsg) // <-- 이 파서는 undefined는 마지막에 명세해야 함
}

func AccountExporterFromIBCMsg(msg *sdktypes.Msg, txHash string) (msgType string, accounts []string) {
	switch msg := (*msg).(type) {
	//ibc transfer (1)
	case *ibctransfertypes.MsgTransfer:
		msgType = IBCTransferMsgTransfer

	// ibc 02-client (4)
	case *ibcclienttypes.MsgCreateClient:
		msgType = IBCClientMsgCreateClient
	case *ibcclienttypes.MsgUpdateClient:
		msgType = IBCClientMsgUpdateClient
	case *ibcclienttypes.MsgUpgradeClient:
		msgType = IBCClientMsgUpgradeClient
	case *ibcclienttypes.MsgSubmitMisbehaviour:
		msgType = IBCClientMsgSubmitMisbehaviour

	// ibc 03 connection (4)
	case *ibcconnectiontypes.MsgConnectionOpenInit:
		msgType = IBCConnectionMsgConnectionOpenInit
	case *ibcconnectiontypes.MsgConnectionOpenTry:
		msgType = IBCConnectionMsgConnectionOpenTry
	case *ibcconnectiontypes.MsgConnectionOpenAck:
		msgType = IBCConnectionMsgConnectionOpenAck
	case *ibcconnectiontypes.MsgConnectionOpenConfirm:
		msgType = IBCConnectionMsgConnectionOpenConfirm

	// ibc 04 channel (10)
	case *ibcchanneltypes.MsgChannelOpenInit:
		msgType = IBCChannelMsgChannelOpenInit
	case *ibcchanneltypes.MsgChannelOpenTry:
		msgType = IBCChannelMsgChannelOpenTry
	case *ibcchanneltypes.MsgChannelOpenAck:
		msgType = IBCChannelMsgChannelOpenAck
	case *ibcchanneltypes.MsgChannelOpenConfirm:
		msgType = IBCChannelMsgChannelOpenConfirm
	case *ibcchanneltypes.MsgChannelCloseInit:
		msgType = IBCChannelMsgChannelCloseInit
	case *ibcchanneltypes.MsgChannelCloseConfirm:
		msgType = IBCChannelMsgChannelCloseConfirm
	case *ibcchanneltypes.MsgRecvPacket:
		msgType = IBCChannelMsgRecvPacket
		switch msg.Packet.DestinationPort {
		case "transfer":
			var pd ibctransfertypes.FungibleTokenPacketData
			AppCodec.UnmarshalJSON(msg.Packet.GetData(), &pd)
			accounts = mbltypes.AddNotNullAccount(pd.Receiver)
		case "icahost":
			var pd interchainaccountstypes.InterchainAccountPacketData
			AppCodec.UnmarshalJSON(msg.Packet.GetData(), &pd)
			icaMsgs, err := interchainaccountstypes.DeserializeCosmosTx(EncodingConfig.Codec, pd.GetData())
			if err != nil {
				// TODO :
				// catch error
				zap.S().Errorf("failed to deserialize tx, error : ", err)
			}
			for i := range icaMsgs {
				var icaMsgType string
				// TODO
				// icaMsgType을 수집해야 할지 여부를 결정하지 못함
				icaMsgType, accounts = mbltypes.AccountExporterFromCosmosTxMsg(&icaMsgs[i])
				for _, customTxParser := range CustomTxParsers {
					if icaMsgType != "" {
						break
					}
					customMsgType, account := customTxParser(&icaMsgs[i], txHash)
					_ = customMsgType
					accounts = append(accounts, account...)
				}
			}
		}
	case *ibcchanneltypes.MsgTimeout:
		msgType = IBCChannelMsgTimeout
		var pd ibctransfertypes.FungibleTokenPacketData
		AppCodec.UnmarshalJSON(msg.Packet.GetData(), &pd)
		accounts = mbltypes.AddNotNullAccount(pd.Sender)
	case *ibcchanneltypes.MsgTimeoutOnClose:
		msgType = IBCChannelMsgTimeoutOnClose
		var pd ibctransfertypes.FungibleTokenPacketData
		AppCodec.UnmarshalJSON(msg.Packet.GetData(), &pd)
		accounts = mbltypes.AddNotNullAccount(pd.Sender)
	case *ibcchanneltypes.MsgAcknowledgement:
		msgType = IBCChannelMsgAcknowledgement

	default:
		// AccountExporterFromCustomTxMsg() 에서 처리
	}

	return
}
