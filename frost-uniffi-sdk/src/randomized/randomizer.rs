use std::sync::Arc;

use rand::thread_rng;
use reddsa::frost::redpallas as frost;

use crate::randomized::randomizer::frost::Randomizer;
use crate::{coordinator::FrostSigningPackage, FrostError, FrostPublicKeyPackage};
use frost::RandomizedParams;
use frost_core::{Ciphersuite, Error};
use uniffi;

type E = reddsa::frost::redpallas::PallasBlake2b512;

#[cfg(feature = "redpallas")]
#[derive(uniffi::Object, Clone)]
pub struct FrostRandomizedParams {
    params: RandomizedParams,
}

impl FrostRandomizedParams {
    fn new(
        public_key_package: FrostPublicKeyPackage,
        signing_package: FrostSigningPackage,
    ) -> Result<FrostRandomizedParams, FrostError> {
        let rng = thread_rng();
        let pallas_signing_package = signing_package.to_signing_package().unwrap();
        let randomized_params = RandomizedParams::new(
            public_key_package
                .into_public_key_package()
                .unwrap()
                .verifying_key(),
            &pallas_signing_package,
            rng,
        )
        .map_err(FrostError::map_err)?;

        let params = FrostRandomizedParams {
            params: randomized_params,
        };

        Ok(params)
    }
}

#[uniffi::export]
pub fn randomized_params_from_public_key_and_signing_package(
    public_key: FrostPublicKeyPackage,
    signing_package: FrostSigningPackage,
) -> Result<Arc<FrostRandomizedParams>, FrostError> {
    let r = FrostRandomizedParams::new(public_key, signing_package)?;

    Ok(Arc::new(r))
}

#[uniffi::export]
pub fn randomizer_from_params(
    randomized_params: Arc<FrostRandomizedParams>,
) -> Result<FrostRandomizer, FrostError> {
    let randomizer = FrostRandomizer::from_randomizer::<E>(*randomized_params.params.randomizer())
        .map_err(FrostError::map_err)?;
    Ok(randomizer)
}

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

    let randomizer = Randomizer::deserialize(&buf).map_err(FrostError::map_err)?;

    FrostRandomizer::from_randomizer::<E>(randomizer).map_err(FrostError::map_err)
}

impl FrostRandomizer {
    pub fn into_randomizer<C: Ciphersuite>(&self) -> Result<Randomizer, Error<E>> {
        Randomizer::deserialize(&self.data)
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
