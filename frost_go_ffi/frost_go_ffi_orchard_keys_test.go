package frost_uniffi_sdk

import (
	"encoding/hex"
	"testing"
)

func TestUFVKAndAddressAreDerivedFromSeed(t *testing.T) {
	// Define the expected values
	expectedFVK := "uviewtest1jd7ucm0fdh9s0gqk9cse9xtqcyycj2k06krm3l9r6snakdzqz5tdp3ua4nerj8uttfepzjxrhp9a4c3wl7h508fmjwqgmqgvslcgvc8htqzm8gg5h9sygqt76un40xvzyyk7fvlestphmmz9emyqhjkl60u4dx25t86lhs30jreghq40cfnw9nqh858z4"
	expectedAddress := "utest1fqasmz9zpaq3qlg4ghy6r5cf6u3qsvdrty9q6e4jh4sxd2ztryy0nvp59jpu5npaqwrgf7sgqu9z7hz9sdxw22vdpay4v4mm8vv2hlg4"

	// Hex-encoded strings
	hexStringAk := "d2bf40ca860fb97e9d6d15d7d25e4f17d2e8ba5dd7069188cbf30b023910a71b"
	hexAk, err := hex.DecodeString(hexStringAk)
	if err != nil {
		t.Fatalf("failed to decode hex string for Ak: %v", err)
	}

	randomSeedBytesHexString := "659ce2e5362b515f30c38807942a10c18a3a2f7584e7135b3523d5e72bb796cc64c366a8a6bfb54a5b32c41720bdb135758c1afacac3e72fd5974be0846bf7a5"
	randomSeedBytes, err := hex.DecodeString(randomSeedBytesHexString)
	if err != nil {
		t.Fatalf("failed to decode hex string for random seed: %v", err)
	}

	zcashNetwork := ZcashNetworkTestnet

	ak, err := OrchardSpendValidatingKeyFromBytes(hexAk)
	if err != nil {
		t.Fatalf("failed to create OrchardSpendValidatingKey: %v", err)
	}

	fvk, err := OrchardFullViewingKeyNewFromValidatingKeyAndSeed(ak, randomSeedBytes, zcashNetwork)
	if err != nil {
		t.Fatalf("failed to create OrchardFullViewingKey: %v", err)
	}

	encodedFVK, err := fvk.Encode()
	if err != nil {
		t.Fatalf("failed to create encode OrchardFullViewingKey: %v", err)
	}

	if encodedFVK != expectedFVK {
		t.Errorf("expected FVK %s, got %s", expectedFVK, encodedFVK)
	}

	address, err := fvk.DeriveAddress()
	if err != nil {
		t.Fatalf("failed to derive address: %v", err)
	}
	stringEncodedAddress := address.StringEncoded()
	if stringEncodedAddress != expectedAddress {
		t.Errorf("expected address %s, got %s", expectedAddress, stringEncodedAddress)
	}
}

func TestUFVKIsDecomposedOnParts(t *testing.T) {
	// Define the UFVK string to be tested
	ufvkString := "uviewtest1jd7ucm0fdh9s0gqk9cse9xtqcyycj2k06krm3l9r6snakdzqz5tdp3ua4nerj8uttfepzjxrhp9a4c3wl7h508fmjwqgmqgvslcgvc8htqzm8gg5h9sygqt76un40xvzyyk7fvlestphmmz9emyqhjkl60u4dx25t86lhs30jreghq40cfnw9nqh858z4"
	// Decode the UFVK string
	zcashNetwork := ZcashNetworkTestnet
	ufvk, err := OrchardFullViewingKeyDecode(ufvkString, zcashNetwork)

	if err != nil {
		t.Fatalf("failed to decode UFVK: %v", err)
	}

	// Decompose UFVK into parts
	nk := ufvk.Nk()

	ak := ufvk.Ak()

	rivk := ufvk.Rivk()

	// Reconstruct the UFVK from the decomposed parts
	roundtripUFVK, err := OrchardFullViewingKeyNewFromCheckedParts(ak, nk, rivk, zcashNetwork)
	if err != nil {
		t.Fatalf("failed to reconstruct UFVK from parts: %v", err)
	}

	roundtripUFVKEncoded, err := roundtripUFVK.Encode()
	if err != nil {
		t.Fatalf("failed to decode Roundtrip UFVK: %v", err)
	}
	// Verify that the original UFVK and the round-trip UFVK are equal
	if roundtripUFVKEncoded != ufvkString {
		t.Errorf("UFVK mismatch: expected %v, got %v", ufvk, roundtripUFVK)
	}
}
