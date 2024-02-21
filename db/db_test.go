package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	//mbl
	"github.com/cosmostation/cosmostation-coreum/custom"
	mblconfig "github.com/cosmostation/mintscan-backend-library/config"
	"github.com/cosmostation/mintscan-database/schema"
	mdschema "github.com/cosmostation/mintscan-database/schema"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	pg "github.com/go-pg/pg/v10"

	"github.com/stretchr/testify/require"
)

var db *Database

func TestMain(m *testing.M) {
	// types.SetAppConfig()

	fileBaseName := "chain-exporter"
	cfg := mblconfig.ParseConfig(fileBaseName)
	db = Connect(&cfg.DB)

	os.Exit(m.Run())
}

func TestInsertOrUpdate(t *testing.T) {
	err := db.Ping()
	require.NoError(t, err)

}

func TestUpdate_Validator(t *testing.T) {
	err := db.Ping()
	require.NoError(t, err)

	val := &mdschema.Validator{
		Address: "kava1ulzzxuvghlv04sglkzyxv94rvl7c2llhs098ju",
		Rank:    5,
	}

	validator, err := db.GetValidatorByAnyAddr(val.Address)
	require.NoError(t, err)

	result, err := db.Model(&validator).
		Set("rank = ?", val.Rank).
		Where("id = ?", validator.ID).
		Update()

	require.NoError(t, err)
	require.Equal(t, 1, result.RowsAffected())
}

func TestConnection(t *testing.T) {
	var n int
	_, err := db.QueryOne(pg.Scan(&n), "SELECT 1")
	require.NoError(t, err)

	require.Equal(t, n, 1, "failed to ping database")
}

func TestGetTx(t *testing.T) {
	// id := int64(4318956)
	// txs, err := db.GetTransactionByID(id)
	// require.NoError(t, err)
	var begin int64
	var counter int64
	for {

		txs, err := db.TestGetRecvPacketTransaction(begin)
		require.NoError(t, err)

		if len(txs) <= 0 {
			break
		}

		tmas := make([]mdschema.TMA, 0)
		for i := range txs {
			// t.Log(txs[i].Hash)
			// unmarshaler := custom.EncodingConfig.Marshaler.UnmarshalJSON
			unmarshaler := custom.EncodingConfig.Codec.UnmarshalJSON
			var txResp sdktypes.TxResponse
			err = unmarshaler(txs[i].Chunk, &txResp)
			require.NoError(t, err)

			tx := txResp.GetTx()
			msgs := tx.GetMsgs()

			type IBCRecvPacketData struct {
				Denom    string       `json:"denom,omitempty"`
				Amount   sdktypes.Int `json:"amount"`
				Sender   string       `json:"sender,omitempty"`
				Receiver string       `json:"receiver,omitempty"`
			}

			var pd IBCRecvPacketData

			for _, msg := range msgs {
				switch m := msg.(type) {
				case *ibcchanneltypes.MsgRecvPacket:
					json.Unmarshal(m.Packet.GetData(), &pd)
					// t.Log("sender :", pd.Sender)
					// t.Log("receiver :", pd.Receiver)
					// t.Log("denom :", pd.Denom)
					// t.Log("amount :", pd.Amount)
					tma := mdschema.TMA{
						TxHash:         txs[i].Hash,
						MsgType:        "ibcchannel/recv_packet",
						AccountAddress: pd.Receiver,
					}
					tmas = append(tmas, tma)
					break
				}
			} // end msgs
			begin = txs[i].ID
			counter++
		} // end txs
		t.Log("next begin", begin)
		t.Log("len of tmas :", len(tmas))
		// err = insert(tmas)
		// if err != nil {
		// 	t.Log("stopped id ", begin)
		// }

	}
	t.Log("count of parsed :", counter)
	t.Log("parse complete")
}

func insert(tmas []mdschema.TMA) error {
	err := db.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		lenTMA := len(tmas)
		if lenTMA > 0 {
			limit := 100
			args := make([]string, 0)
			for i := 0; i < lenTMA; i += limit {
				if i+limit > lenTMA {
					limit = lenTMA - i
				}
				args = append(args, parseTMAToArg(tmas[i:i+limit]))
			}
			for i := range args {
				query := "select public.f_insert_tx_msg_acc" + args[i]
				_, err := tx.Exec(query)
				if err != nil {
					return fmt.Errorf("failed to insert public.f_insert_tx_msg_acc : %s", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func parseTMAToArg(tma []schema.TMA) string {
	arg := "("
	fmt.Println("len of partial tma", len(tma))
	for i := range tma {
		if i != 0 {
			arg += ","
		}
		arg += fmt.Sprintf("('%s', '%s', '%s')", tma[i].TxHash, tma[i].MsgType, tma[i].AccountAddress)
	}
	arg += ")"
	return arg
}

func TestGetLiveProposal(t *testing.T) {
	ps, err := db.QueryGetLiveProposal()
	require.NoError(t, err)

	t.Log(ps)
}
