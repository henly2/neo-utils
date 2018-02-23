package smartcontract

import "encoding/hex"

type NativeAsset string
type TradingVersion byte //currently 0

const (
	NEO NativeAsset = "c56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b"
	GAS NativeAsset = "602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7"
)

func (n NativeAsset) ToLittleEndianBytes() []byte {
	b, err := hex.DecodeString(string(n))
	if err != nil {
		return nil
	}
	return reverseBytes(b)
}

const (
	NEOTradingVersion TradingVersion = 0x00
)