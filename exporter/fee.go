package exporter

import (
	"encoding/json"
	"fmt"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"
	"go.uber.org/zap"
)

type DailyFee map[int64]sdktypes.Coins

func (ex *Exporter) runFeeMaker() {
	// database index_pointer로 부터 시작 위치를 가져옴
	// error : sleep 후 다시 시도
	// case -1 : index_pointer에 해당 마커가 없을 경우 0으로 초기화
	// pointer로 부터 fee 추출
	// fee, index_pointer 데이터베이스 저장 & 업데이트

	indexName := "tx_fee_pointer"

	for {
		ip, err := ex.DB.GetIndexPointer(indexName)
		if err != nil {
			// 실패 시 재시도
			zap.S().Error("failed to get index pointer ", err)
			time.Sleep(time.Second * 2)
			continue
		}

		switch ip.Pointer {
		case -1: // when no rows
			ip.Pointer = 0
			err := ex.DB.InitializeIndexPointer(ip)
			if err != nil {
				// 실패 시 재시도
				zap.S().Error("failed to insert index pointer ", err)
				time.Sleep(time.Second * 2)
				continue
			}
			fallthrough
		default:
			newPointer, fees, dailyFee, err := ex.GetFees(ip.Pointer)
			if err != nil {
				zap.S().Error("failed to get fees ", err)
				// 실패 시 재시도
				time.Sleep(time.Second * 2)
				continue
			}
			_ = dailyFee

			// fee, index_pointer 저장
			err = ex.DB.SaveFees(ip, newPointer, fees)
			if err != nil {
				zap.S().Error("failed to save fees  ", err)
				time.Sleep(time.Second * 2)
				continue
			}
		}
	}
}

func (ex *Exporter) GetFees(beginTxID int64) (int64, []mdschema.Fee, DailyFee, error) {
	feeList := make([]mdschema.Fee, 0)
	dailyFee := make(DailyFee)

	endTxID := beginTxID

	zap.S().Info("start aggregating fee, tx_id : ", beginTxID)
	txs, err := ex.DB.GetTransactionsOrderBy(beginTxID, "ASC", 100)
	if err != nil {
		return endTxID, feeList, dailyFee, fmt.Errorf("failed to get transactions %s", err)
	}

	if len(txs) == 0 {
		time.Sleep(2 * time.Second)
		return endTxID, feeList, dailyFee, fmt.Errorf("no transactions")
	}

	// json -> tx -> value -> fee -> amount[] -> denom, amount // sdk 0.39 below
	// json -> tx -> auth_info -> fee -> amount[] -> denom, amount // sdk 0.40 above
	for i := range txs {
		raw := make(map[string]interface{})
		err := json.Unmarshal(txs[i].Chunk, &raw)
		if err != nil {
			return endTxID, feeList, dailyFee, fmt.Errorf("failed to unmarshal chunk %s", err)
		}

		rawTx, ok := raw["tx"].(map[string]interface{})
		if !ok {
			return endTxID, feeList, dailyFee, fmt.Errorf("failed to assert tx")
		}

		aggregationDate := txs[i].Timestamp.Truncate(time.Hour * 24)
		unixTime := aggregationDate.Unix()

		// unix := time.Unix(unixTime, 0).UTC().Format(time.RFC3339) // daily_fee timestamp format
		// zap.S().Info(txs[i].ID, txs[i].Timestamp, unix)

		rawAuthInfo, ok := rawTx["auth_info"].(map[string]interface{}) // sdk-0.40 above
		if !ok {
			rawAuthInfo, ok = rawTx["value"].(map[string]interface{}) // sdk-0.33 - sdk-0.39
			if !ok {
				return endTxID, feeList, dailyFee, fmt.Errorf("failed to assert value and auth_info")
			}
		}

		fee, ok := rawAuthInfo["fee"].(map[string]interface{})
		if !ok {
			return endTxID, feeList, dailyFee, fmt.Errorf("failed to assert auth_info")
		}
		coins := make([]sdktypes.Coin, 0)
		b, err := json.Marshal(fee["amount"])
		if err != nil {
			return endTxID, feeList, dailyFee, fmt.Errorf("failed to unmarshal amount to []byte %s", err)
		}
		err = json.Unmarshal(b, &coins)
		if err != nil {
			return endTxID, feeList, dailyFee, fmt.Errorf("failed to unmarshal to coins %s", err)
		}

		df, ok := dailyFee[unixTime]
		if !ok {
			df = sdktypes.NewCoins()
		}
		dailyFee[unixTime] = df.Add(coins...)

		uniqueCoins := make(map[string]sdktypes.Int)
		for j := range coins {
			uc, ok := uniqueCoins[coins[j].Denom]
			if !ok {
				uc = sdktypes.ZeroInt()
			}
			sum := uc.Add(coins[j].Amount)
			uniqueCoins[coins[j].Denom] = sum
		}

		for k, v := range uniqueCoins {
			f := mdschema.Fee{
				TxID:      txs[i].ID,
				Denom:     k,
				Amount:    v.String(),
				Timestamp: txs[i].Timestamp,
			}
			feeList = append(feeList, f)
		}
		if endTxID < txs[i].ID {
			endTxID = txs[i].ID
		}
	}

	return endTxID, feeList, dailyFee, nil
}

func ParseDailyFee(df DailyFee) (dfs []mdschema.DailyFee) {
	for unixTime, coins := range df {
		for i := range coins {
			dbdf := mdschema.DailyFee{
				Denom:     coins[i].Denom,
				Amount:    coins[i].Amount.String(),
				Timestamp: time.Unix(unixTime, 0).UTC(),
			}
			dfs = append(dfs, dbdf)
		}
	}
	return dfs
}

func MergeDailyFee(dbdf []mdschema.DailyFee, df DailyFee) (dfs []mdschema.DailyFee) {
	// // db에서 가져와서 map 호출해서 합치고 마지막에 parse후 리턴
	// for i := range dbdf {
	// 	dbdf[i].Denom
	// }

	return ParseDailyFee(df)
}
