package dogecoin_test

import (
	"github.com/renprotocol/multichain"
	"github.com/renprotocol/multichain/chain/bitcoin"
	"github.com/renprotocol/multichain/chain/dogecoin"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dogecoin", func() {
	Context("when decoding segwit address", func() {
		Context("when decoding an address from a different network ", func() {
			It("should return an error ", func() {
				// A valid bitcoin segwit address which is not a valid doge address
				addr := multichain.Address("bc1qk6yk2ctcu2pmtxfzhya692h774562vlv2g7dvl")
				decoder := bitcoin.NewAddressDecoder(&dogecoin.MainNetParams)
				_, err := decoder.DecodeAddress(addr)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
