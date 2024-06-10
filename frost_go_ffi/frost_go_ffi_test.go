package frost_uniffi_sdk

import "testing"

func TestTrustedDealerFromConfigWithSecret(t *testing.T) {
	secret := []byte{
		123, 28, 51, 211, 245, 41, 29, 133, 222, 102, 72, 51, 190, 177, 173, 70, 159, 127, 182, 2,
		90, 14, 199, 139, 58, 121, 12, 110, 19, 169, 131, 4,
	}

	secretConfig := Configuration{
		MinSigners: 2,
		MaxSigners: 3,
		Secret:     secret,
	}

	message := Message{
		Data: []byte("i am a message"),
	}

	keygen, err := TrustedDealerKeygenFrom(secretConfig)
	if err != nil {
		t.Fatalf("Failed to generate keygen: %v", err)
	}

	publicKey := keygen.PublicKeyPackage
	shares := keygen.SecretShares

	keyPackages := make(map[ParticipantIdentifier]FrostKeyPackage)
	for identifier, value := range shares {
		keyPackage, err := VerifyAndGetKeyPackageFrom(value)
		if err != nil {
			t.Fatalf("Failed to get key package: %v", err)
		}
		keyPackages[identifier] = keyPackage
	}

	if len(shares) != 3 {
		t.Fatalf("Expected 3 shares, got %d", len(shares))
	}

	if len(publicKey.VerifyingShares) != 3 {
		t.Fatalf("Expected 3 verifying shares, got %d", len(publicKey.VerifyingShares))
	}

	nonces := make(map[ParticipantIdentifier]FrostSigningNonces)
	var commitments []FrostSigningCommitments

	for participant, secretShare := range shares {
		firstRoundCommitment, err := GenerateNoncesAndCommitments(secretShare)
		if err != nil {
			t.Fatalf("Failed to generate nonces and commitments: %v", err)
		}
		nonces[participant] = firstRoundCommitment.Nonces
		commitments = append(commitments, firstRoundCommitment.Commitments)
	}

	signingPackage, err := NewSigningPackage(message, commitments)
	if err != nil {
		t.Fatalf("Failed to create signing package: %v", err)
	}

	var signatureShares []FrostSignatureShare
	for participant, keyPackage := range keyPackages {
		signatureShare, err := Sign(signingPackage, nonces[participant], keyPackage)
		if err != nil {
			t.Fatalf("Failed to sign: %v", err)
		}
		signatureShares = append(signatureShares, signatureShare)
	}

	signature, err := Aggregate(signingPackage, signatureShares, publicKey)
	if err != nil {
		t.Fatalf("Failed to aggregate signature: %v", err)
	}

	if err := VerifySignature(message, signature, publicKey); err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
}
