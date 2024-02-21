package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"go.uber.org/zap"

	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	sdktypestx "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	mblconfig "github.com/cosmostation/mintscan-backend-library/config"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var cli *Client

func TestMain(m *testing.M) {
	fileBaseName := "chain-exporter"
	cfg := mblconfig.ParseConfig(fileBaseName)

	cli = NewClient(&cfg.Client)

	os.Exit(m.Run())

}

func TestGetAccount(t *testing.T) {
	// address := "cosmos1pvzrncl89w5z9psr8ch90057va9tc23pehpd2t"
	address := "cosmos1x5wgh6vwye60wv3dtshs9dmqggwfx2ldnqvev0"
	sdkaddr, err := sdktypes.AccAddressFromBech32(address)
	if err != nil {
		log.Println(err)
	}
	ar := authtypes.AccountRetriever{}
	log.Println(cli.CliCtx)
	acc, err := ar.GetAccount(cli.GetCLIContext(), sdkaddr)
	if err != nil {
		log.Println(err)
	}

	log.Println(acc.GetAddress())
	log.Println(acc.GetPubKey())
}

func TestGetAccountBalance(t *testing.T) {

	// address := "cosmos1x5wgh6vwye60wv3dtshs9dmqggwfx2ldnqvev0"
	address := "cosmos1emaa7mwgpnpmc7yptm728ytp9quamsvuz92x5u"
	// log.Println(cli.GetAccount("cosmos1x5wgh6vwye60wv3dtshs9dmqggwfx2ldnqvev0"))
	sdkaddr, err := sdktypes.AccAddressFromBech32(address)
	if err != nil {
		log.Println(err)
	}
	b := banktypes.NewQueryBalanceRequest(sdkaddr, "umuon")
	log.Println(b)
	bankClient := banktypes.NewQueryClient(cli.GRPC)
	var header metadata.MD
	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	log.Println("blockHeight :", blockHeight)
	// header.Set(k string, vals ...string)
	// header.Append(grpctypes.GRPCBlockHeightHeader, "1")
	// header.Set(grpctypes.GRPCBlockHeightHeader, "1")
	// bankRes, err := bankClient.Balance(
	// 	metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, "1"), // Add metadata to request
	// 	b,
	// 	grpc.Header(&header),
	// )
	bankRes, err := bankClient.Balance(
		context.Background(),
		b,
		grpc.Header(&header), // Also fetch grpc header
	)
	if err != nil {
		log.Println(err)
	}
	if bankRes.GetBalance() != nil {
		log.Println(*bankRes.GetBalance())
	}
	blockHeight = header.Get(grpctypes.GRPCBlockHeightHeader)
	log.Println("blockHeight :", blockHeight)
}
func TestGetNetworkChainID(t *testing.T) {
	n, err := cli.RPC.GetNetworkChainID()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(n)
}

func TestGetBlock(t *testing.T) {
	log.Println(cli.RPC.GetBlock(11111))
}

func TestTimeoutTx(t *testing.T) {
	hash := "CF2A5718C64D676A86D7EB1EB1121EB79C18EF8D32259395A8636883981FFCFCa"

	txResp, err := cli.CliCtx.GetTx(hash)
	// if txResp.TxHash == "" {
	// 	t.Log(" nil")
	// 	t.Log(txResp)
	// }
	t.Log("txresp :", txResp)
	require.NoError(t, err)

	msgs := txResp.GetTx().GetMsgs()
	for _, m := range msgs {
		switch e := m.(type) {
		case *ibcchanneltypes.MsgTimeout:
			var pd ibctransfertypes.FungibleTokenPacketData
			cli.GetCLIContext().Codec.UnmarshalJSON(e.Packet.GetData(), &pd)
			log.Println(pd.Receiver)
			log.Println(pd.Sender)
		default:
		}
	}

}

func TestGetTxInBlock(t *testing.T) {

	controler := make(chan struct{}, 60)
	wg := new(sync.WaitGroup)
	h := int64(2665758)

	block, err := cli.RPC.GetBlock(h)
	require.NoError(t, err)

	zap.S().Infof("number of Transactions : %d", len(block.Block.Txs))
	txList := block.Block.Txs
	// txs := make([]*sdktypes.TxResponse, len(block.Block.Txs))

	for idx, tx := range txList {
		hex := fmt.Sprintf("%X", tx.Hash())
		controler <- struct{}{}
		wg.Add(1)

		go func(i int, gHex string) {
			t.Logf(" [%04d] : [%s]", i, gHex)
			defer func() {
				<-controler
				wg.Done()
			}()

			// txs[i], err = ex.Client.CliCtx.GetTx(hex)
			cnt := 0
		RETRY:
			var err error
			if i == 33 {
				t.Logf("[cnt %d] [%04d] : [%s]", cnt, i, gHex)
				t.Log("retry")
				err = fmt.Errorf("abcd")
				cnt++

			}

			if err != nil {
				t.Logf("failed to get tx height=%d, hash=%s", h, gHex)
				time.Sleep(1 * time.Second)
				if cnt > 5 {
					return
				}
				goto RETRY
			}
		}(idx, hex)

	}
	wg.Wait()

}

func TestGetTx(t *testing.T) {
	sendTx := "A80ADDA7929801AF3B1E6957BE9C63C30B5A0B9F903E760C555CAC19D2FC0DFC"
	withdrawAllRewardsTx := "53A036CC53FD3AD8C4B66C11BBB20DC63A5B606144F6655EC9D9E327AB9BA3D9"
	delegateTx := "DDA04447F569B402D96E7CCCC9ACF0C76D3581EC9B056818CED7913DECA6F10A"
	unknownTx := "30B43BB887FA6F56E5302B6CCB9C439A6C2AF29CFADA1465C0174EE6C62E3D28"
	_, _, _, _ = sendTx, withdrawAllRewardsTx, delegateTx, unknownTx
	txhash := unknownTx
	txResp, err := cli.CliCtx.GetTx(txhash)
	if err != nil {
		log.Fatal(err)
	}

	tx := txResp.GetTx()
	ta, ok := tx.(*sdktypestx.Tx)
	log.Println(ok)
	if ok {
		a, err := cli.CliCtx.Codec.MarshalJSON(ta.GetBody())
		if err != nil {
			log.Println(err)
		}
		log.Println("message :", string(a))
		a, err = cli.CliCtx.Codec.MarshalJSON(ta.GetAuthInfo().GetFee())
		if err != nil {
			log.Println(err)
		}
		log.Println("fee :", string(a))
		a, err = cli.CliCtx.Codec.MarshalJSON(ta.GetAuthInfo())
		if err != nil {
			log.Println(err)
		}
		log.Println("authinfo :", string(a))
		a, err = cli.CliCtx.Codec.MarshalJSON(ta.GetAuthInfo().GetSignerInfos()[0])
		if err != nil {
			log.Println(err)
		}
		log.Println("signerinfo[0] :", string(a))
		for _, addr := range ta.GetSigners() {
			log.Println("getsigners addr :", addr)
		}
		a, err = cli.CliCtx.Codec.MarshalJSON(ta.GetAuthInfo().GetSignerInfos()[0].GetPublicKey())
		if err != nil {
			log.Println(err)
		}
		log.Println("pubkey[0] :", ta.GetAuthInfo().GetSignerInfos()[0].GetPublicKey().GetValue())
		log.Println("pubkey[0] :", string(a))
		sig := ta.GetSignatures()
		// json.Unmarshal(sig[0], &i)
		log.Println("signatures :", sig[0])
		log.Println("memo:", ta.GetBody().Memo)
	}

	msgs := txResp.GetTx().GetMsgs()
	for i, m := range msgs {
		switch t := m.(type) {
		case *banktypes.MsgSend:
			log.Println("banktypes :", t)
			log.Println(t.FromAddress)
			log.Println(t.ToAddress)
			log.Println(t.Amount)
			log.Println(t.Type())
		default:
			log.Println(i, t)
		}
	}
}
