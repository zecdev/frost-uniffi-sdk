
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;
use uniffi;
use frost::RandomizedParams;
use crate::{frost::Error, FrostError};
use frost::round2::Randomizer;

#[cfg(feature = "redpallas")]
#[derive(uniffi::Record)]
pub struct FrostRandomizer {
    data: Vec<u8>
}

impl FrostRandomizer {
    pub(crate) fn into_randomizer(&self) -> Result<Randomizer, Error> {
        let raw_randomizer = &self.data[0..32].try_into()
            .map_err(|_| Error::DeserializationError)?;
        
        Randomizer::deserialize(raw_randomizer)
    }

    pub (crate) fn from_randomizer(randomizer: Randomizer) -> Result<FrostRandomizer, Error> {
        Ok(
            FrostRandomizer {
                data: randomizer.serialize().to_vec()
            }
        )
    }

    pub (crate) fn randomizer_params(randomizer: Randomizer, public_key_package: &frost::keys::PublicKeyPackage) -> RandomizedParams {
        RandomizedParams::from_randomizer(
            public_key_package.verifying_key(),
            randomizer,
        )
    }
}

#[uniffi::export]
pub fn from_hex_string(hex_string: String) -> Result<FrostRandomizer, FrostError> {
    let randomizer_hex_bytes = hex::decode(hex_string.trim())
        .map_err(|_| FrostError::DeserializationError)?;
    
    let randomizer = frost::round2::Randomizer::deserialize(
        &randomizer_hex_bytes
            .try_into()
            .map_err(|_| FrostError::UnknownIdentifier)?
    )
    .map_err(|_| FrostError::DeserializationError)?;

    FrostRandomizer::from_randomizer(
        randomizer
    )
    .map_err(|_| FrostError::DeserializationError)
}
