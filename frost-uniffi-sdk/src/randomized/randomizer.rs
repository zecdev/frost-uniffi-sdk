use reddsa::frost::redpallas as frost;

use crate::FrostError;
use frost::round2::Randomizer;
use frost::RandomizedParams;
use frost_core::{Ciphersuite, Error};
use uniffi;

type E = reddsa::frost::redpallas::PallasBlake2b512;

#[cfg(feature = "redpallas")]
#[derive(uniffi::Record, Clone)]
pub struct FrostRandomizer {
    data: Vec<u8>,
}

#[cfg(feature = "redpallas")]
#[uniffi::export]
pub fn from_hex_string(hex_string: String) -> Result<FrostRandomizer, FrostError> {
    let randomizer_hex_bytes =
        hex::decode(hex_string.trim()).map_err(|_| FrostError::DeserializationError)?;

    let buf: [u8; 32] = randomizer_hex_bytes[0..32]
        .try_into()
        .map_err(|_| FrostError::DeserializationError)?;

    let randomizer = frost::round2::Randomizer::deserialize(&buf).map_err(FrostError::map_err)?;

    FrostRandomizer::from_randomizer::<E>(randomizer).map_err(FrostError::map_err)
}

impl FrostRandomizer {
    pub fn into_randomizer<C: Ciphersuite>(&self) -> Result<Randomizer, Error<E>> {
        let raw_randomizer = &self.data[0..32]
            .try_into()
            .map_err(|_| Error::DeserializationError)?;

        Randomizer::deserialize(raw_randomizer)
    }

    pub fn from_randomizer<C: Ciphersuite>(
        randomizer: Randomizer,
    ) -> Result<FrostRandomizer, Error<C>> {
        Ok(FrostRandomizer {
            data: randomizer.serialize().to_vec(),
        })
    }

    pub(crate) fn randomizer_params(
        randomizer: Randomizer,
        public_key_package: &frost::keys::PublicKeyPackage,
    ) -> RandomizedParams {
        RandomizedParams::from_randomizer(public_key_package.verifying_key(), randomizer)
    }
}
