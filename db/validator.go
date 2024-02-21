package db

import (
	"strings"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmostation/mintscan-database/schema"
	pg "github.com/go-pg/pg/v10"
)

// GetValidatorByAnyAddr returns a validator information by any type of address format
func (db *Database) GetValidatorByAnyAddr(anyAddr string) (schema.Validator, error) {
	var val schema.Validator
	var err error

	switch {
	// jeonghwan
	case strings.HasPrefix(anyAddr, sdktypes.GetConfig().GetBech32ConsensusPubPrefix()): // Bech32 prefix for validator public key
		err = db.Model(&val).
			Where("consensus_pubkey = ?", anyAddr).
			Limit(1).
			Select()
	case strings.HasPrefix(anyAddr, sdktypes.GetConfig().GetBech32ValidatorAddrPrefix()): // Bech32 prefix for validator address
		err = db.Model(&val).
			Where("operator_address = ?", anyAddr).
			Limit(1).
			Select()
	case strings.HasPrefix(anyAddr, sdktypes.GetConfig().GetBech32AccountAddrPrefix()): // Bech32 prefix for account address
		err = db.Model(&val).
			Where("address = ?", anyAddr).
			Limit(1).
			Select()
	case len(anyAddr) == 40: // Validator consensus address in hex
		anyAddr := strings.ToUpper(anyAddr)
		err = db.Model(&val).
			Where("proposer = ?", anyAddr).
			Limit(1).
			Select()
	default:
		err = db.Model(&val).
			Where("moniker = ?", anyAddr). // Validator moniker
			Limit(1).
			Select()
	}

	if err != nil {
		if err == pg.ErrNoRows {
			return schema.Validator{}, nil
		}
		return schema.Validator{}, err
	}

	return val, nil
}
