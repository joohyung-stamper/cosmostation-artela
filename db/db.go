package db

import (
	"context"
	"fmt"

	mblconfig "github.com/cosmostation/mintscan-backend-library/config"
	mddb "github.com/cosmostation/mintscan-database/db"
	"github.com/cosmostation/mintscan-database/schema"
	mdschema "github.com/cosmostation/mintscan-database/schema"
	"go.uber.org/zap"

	pg "github.com/go-pg/pg/v10"
)

var (
	// columnLength is the column length of varchar type in every table.
	// This needs to be considered again to set it to what specific length is needed, but right now set it to 99999.
	columnLength = 99999
)

// Database implements a wrapper of golang ORM with focus on PostgreSQL.
type Database struct {
	*mddb.Database
}

// Connect opens a database connections with the given database connection info from config.
func Connect(dbcfg *mblconfig.DatabaseConfig) *Database {
	db := mddb.Connect(dbcfg.Host, dbcfg.Port, dbcfg.User, dbcfg.Password, dbcfg.DBName, dbcfg.CommonSchema, dbcfg.ChainSchema, dbcfg.Timeout)
	zap.S().Info("db package :", schema.GetCommonSchema())
	zap.S().Info("db package :", schema.GetChainSchema())
	mdschema.SetCommonSchema(dbcfg.CommonSchema)
	mdschema.SetChainSchema(dbcfg.ChainSchema)
	return &Database{db}
}

// CreateTables creates database tables using ORM (Object Relational Mapper).
func (db *Database) CreateTablesAndIndexes() {
	// 생성 오류 시 패닉
	db.CreateTables()
}

func (db *Database) InsertRefineRealTimeData(e *schema.BasicData) error {
	err := db.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		err := db.InsertBlock(tx, e.Block)
		if err != nil {
			return err
		}

		if len(e.Transactions) > 0 {
			for i := range e.Transactions {
				if e.Block.ID != 0 {
					e.Transactions[i].BlockID = e.Block.ID
				} else {
					return fmt.Errorf("failed to insert result txs, can not get block.id")
				}
			}
			err := db.InsertTransaction(tx, e.Transactions, e.TMAs)
			if err != nil {
				return err
			}
		}

		return nil
	})

	// Roll back if any insertion fails.
	if err != nil {
		return err
	}

	return nil
}

// QueryAccountMobile queries account information
func (db *Database) QueryAccountMobile(address string) (*mdschema.AccountMobile, error) {
	var account *mdschema.AccountMobile
	_ = db.Model(&account).
		Where("address = ?", address).
		Select()

	return account, nil
}

// QueryAlarmTokens queries user's alarm tokens
func (db *Database) QueryAlarmTokens(address string) ([]string, error) {
	var accounts []mdschema.AccountMobile
	_ = db.Model(&accounts).
		Column("alarm_token").
		Where("address = ?", address).
		Select()

	var result []string
	if len(accounts) > 0 {
		for _, account := range accounts {
			result = append(result, account.AlarmToken)
		}
	}

	return result, nil
}

func (db *Database) TestGetRecvPacketTransaction(beginTxID int64) (txs []mdschema.Transaction, err error) {
	// var txs []mdschema.Transaction
	// var err error

	limit := 100

	_, err = db.Query(&txs, "select id, hash, chunk from public.transaction where id in ( "+
		"select distinct(tx_id) from public.transaction_message where msg_id = 34 and tx_id > ? order by tx_id asc limit ?)", beginTxID, limit)

	if err != nil {
		if err == pg.ErrNoRows {
			return txs, nil
		}
		return txs, err
	}

	return txs, nil
}

// GetTransactionsByMsgType returns transactions with txid ascending order satifying txid >= beginTxID and msg == msgType
func (db *Database) GetTransactionsByMsgType(beginTxID int64, msgType string, limit int) (txs []schema.Transaction, err error) {
	// _, err = db.Query(&txs, "select * from transaction where id in (select tx_id from transaction_message where tx_id >= ? and msg_id = (select id from message_info where type = ?)) order by id asc limit ?; ", beginTxID, msgType, limit)
	_, err = db.Query(&txs, "select * from transaction where id in (select tx_id from transaction_message where tx_id >= ? and msg_id = (select id from message_info where type = ?) order by tx_id asc limit ?) order by id asc", beginTxID, msgType, limit)

	if err != nil {
		if err == pg.ErrNoRows {
			return txs, nil
		}
		return txs, err
	}

	return txs, nil
}
