#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

#[cfg(feature = "redpallas")]
use crate::randomized::randomizer::FrostRandomizer;

#[cfg(feature = "redpallas")]
type E = reddsa::frost::redpallas::PallasBlake2b512;
#[cfg(not(feature = "redpallas"))]
type E = frost_ed25519::Ed25519Sha512;

use crate::{
    coordinator::{CoordinationError, FrostSignature, FrostSigningPackage},
    participant::FrostSignatureShare,
    FrostPublicKeyPackage,
};

use frost::{round2::SignatureShare, Identifier};
use std::collections::BTreeMap;
use uniffi;

#[cfg(feature = "redpallas")]
#[uniffi::export]
pub fn aggregate(
    signing_package: FrostSigningPackage,
    signature_shares: Vec<FrostSignatureShare>,
    pubkey_package: FrostPublicKeyPackage,
    randomizer: FrostRandomizer,
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
                .to_signature_share::<E>()
                .map_err(|_| CoordinationError::SignatureShareDeserializationError)?,
        );
    }

    let public_key_package = pubkey_package
        .into_public_key_package()
        .map_err(|_| CoordinationError::PublicKeyPackageDeserializationError)?;

    #[cfg(feature = "redpallas")]
    let randomizer = randomizer
        .into_randomizer::<E>()
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
    .map(FrostSignature::from_signature)
}
