#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

#[cfg(feature = "redpallas")]
use crate::randomized::randomizer::FrostRandomizer;


use crate::{
    participant::{FrostSignatureShare, FrostSigningCommitments}, FrostPublicKeyPackage
};
use frost::{
    round1::SigningCommitments, round2::SignatureShare, Error, Identifier, Signature,
    SigningPackage,
};
use std::collections::BTreeMap;
use uniffi;
#[uniffi::export]
pub fn aggregate(
    signing_package: FrostSigningPackage,
    signature_shares: Vec<FrostSignatureShare>,
    pubkey_package: FrostPublicKeyPackage,
    #[cfg(feature = "redpallas")] randomizer: FrostRandomizer,
) -> Result<FrostSignature, CoordinationError> {
    let signing_package = signing_package
        .to_signing_package()
        .map_err(|_| CoordinationError::FailedToCreateSigningPackage)?;

    let mut shares: BTreeMap<Identifier, SignatureShare> = BTreeMap::new();

    for share in signature_shares {
        shares.insert(
            share
                .identifier
                .into_identifier()
                .map_err(|_| CoordinationError::IdentifierDeserializationError)?,
            share
                .to_signature_share()
                .map_err(|_| CoordinationError::SignatureShareDeserializationError)?,
        );
    }

    let public_key_package = pubkey_package
        .into_public_key_package()
        .map_err(|_| CoordinationError::PublicKeyPackageDeserializationError)?;

    #[cfg(feature = "redpallas")]
    let randomizer = randomizer
        .into_randomizer()
        .map_err(|_| CoordinationError::InvalidRandomizer)?;

    frost::aggregate(
        &signing_package,
        &shares,
        &public_key_package,
        #[cfg(feature = "redpallas")]
        &FrostRandomizer::randomizer_params(randomizer, &public_key_package),
    )
    .map_err(|e| CoordinationError::SignatureShareAggregationFailed {
        message: e.to_string(),
    })
    .map(|signature| FrostSignature::from_signature(signature))
}