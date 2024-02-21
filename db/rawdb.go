package db

import (

	//mbl
	mblconfig "github.com/cosmostation/mintscan-backend-library/config"
	mdrawdb "github.com/cosmostation/mintscan-database/rawdb"
	mdschema "github.com/cosmostation/mintscan-database/schema"
)

// Database implements a wrapper of golang ORM with focus on PostgreSQL.
type RawDatabase struct {
	*mdrawdb.Database
}

// Connect opens a database connections with the given database connection info from config.
func RawDBConnect(dbcfg *mblconfig.DatabaseConfig) *RawDatabase {
	db := mdrawdb.Connect(dbcfg.Host, dbcfg.Port, dbcfg.User, dbcfg.Password, dbcfg.DBName, dbcfg.CommonSchema, dbcfg.ChainSchema, dbcfg.Timeout)
	mdschema.SetCommonSchema(dbcfg.CommonSchema)
	mdschema.SetChainSchema(dbcfg.ChainSchema)
	return &RawDatabase{db}
}

// CreateTables creates database tables using ORM (Object Relational Mapper).
func (db *RawDatabase) CreateTablesAndIndexes() {
	// 생성 오류 시 패닉
	db.CreateTables()
}
