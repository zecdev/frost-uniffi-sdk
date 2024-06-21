#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

use crate::{
    coordinator::FrostSigningPackage,
    participant::{FrostSignatureShare, FrostSigningNonces, Round2Error},
    randomized::randomizer::FrostRandomizer,
    FrostKeyPackage,
};

#[cfg(feature = "redpallas")]
#[uniffi::export]
pub fn sign(
    signing_package: FrostSigningPackage,
    nonces: FrostSigningNonces,
    key_package: FrostKeyPackage,
    randomizer: &FrostRandomizer,
) -> Result<FrostSignatureShare, Round2Error> {
    let signing_package = signing_package
        .to_signing_package()
        .map_err(|_| Round2Error::SigningPackageDeserializationError)?;

    let nonces = nonces
        .to_signing_nonces()
        .map_err(|_| Round2Error::NonceSerializationError)?;

    let key_package = key_package
        .into_key_package()
        .map_err(|_| Round2Error::InvalidKeyPackage)?;

    let identifier = *key_package.identifier();

    let randomizer = randomizer
        .into_randomizer::<frost::PallasBlake2b512>()
        .map_err(|_| Round2Error::InvalidRandomizer)?;

    let share =
        frost::round2::sign(&signing_package, &nonces, &key_package, randomizer).map_err(|e| {
            Round2Error::SigningFailed {
                message: e.to_string(),
            }
        })?;

    FrostSignatureShare::from_signature_share(identifier, share).map_err(|e| {
        Round2Error::SigningFailed {
            message: e.to_string(),
        }
    })
}
