package custom

import (
	chainapp "github.com/CoreumFoundation/coreum/v3/app"
	"github.com/CoreumFoundation/coreum/v3/pkg/config"
	"github.com/cosmos/cosmos-sdk/codec"
)

// Codec is the application-wide Amino codec and is initialized upon package loading.
var (
	AppCodec       codec.Codec
	AminoCodec     *codec.LegacyAmino
	EncodingConfig config.EncodingConfig
)

func init() {
	EncodingConfig = config.NewEncodingConfig(chainapp.ModuleBasics)
	AppCodec = EncodingConfig.Codec
	AminoCodec = EncodingConfig.Amino
}
