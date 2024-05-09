use frost::keys::KeyPackage;
use frost::keys::SigningShare;
use frost::Ciphersuite;
use frost::Identifier;
#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;
use hex::ToHex;
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

pub mod trusted_dealer;
mod utils;
use crate::frost::Error;
use std::collections::HashMap;

use uniffi;
use frost::keys::SecretShare;
use frost::keys::PublicKeyPackage;
uniffi::setup_scaffolding!();


#[derive(uniffi::Record)]
pub struct FrostSecretKeyShare {
    pub identifier: String,
    pub data: Vec<u8>
}

#[derive(uniffi::Record)]
pub struct FrostPublicKeyPackage {
    pub verifying_shares: HashMap<String, String>,
    pub verifying_key: String
}

impl FrostSecretKeyShare {
    fn from_secret_share(secret_share: SecretShare) -> Result<FrostSecretKeyShare, frost::Error> {
        let identifier = hex::encode(secret_share.identifier().serialize());
        let serialized_share = secret_share.serialize()
        .map_err(|_| frost::Error::SerializationError)?;

        Ok(
            FrostSecretKeyShare {
                identifier: identifier,
                data: serialized_share
            }
        )
    }

    fn to_secret_share(&self) -> Result<SecretShare, Error> {
        let hex_identifier = hex::decode(self.identifier.clone())
            .map_err(|_| frost::Error::DeserializationError)?;
        
        let slice_identifier = hex_identifier[0..32]
            .try_into()
            .map_err(|_| frost::Error::DeserializationError)?;

        let identifier = Identifier::deserialize(slice_identifier)?;
    
        let secret_share = SecretShare::deserialize(&self.data)
            .map_err(|_| frost::Error::SerializationError)?;

        if identifier != *secret_share.identifier() {
            Err(frost::Error::UnknownIdentifier)
        } else {
            Ok(secret_share)
        }
    }
}

impl FrostPublicKeyPackage {
    fn from_public_key_package(key_package: PublicKeyPackage) -> Result<FrostPublicKeyPackage, Error> {

        let verifying_shares = key_package.verifying_shares();
        let verifying_key = key_package.verifying_key();

        let mut shares = HashMap::new();

        for (k, v) in verifying_shares {
            shares.insert(
                hex::encode(k.serialize()), 
                hex::encode(v.serialize())
            );
        }

        Ok(Self {
            verifying_shares: shares,
            verifying_key: hex::encode(verifying_key.serialize())
        })
    }
}

#[derive(uniffi::Record)]
pub struct Configuration {
    pub min_signers: u16,
    pub max_signers: u16,
    pub secret: Vec<u8>,
}

#[derive(Debug, uniffi::Error, thiserror::Error)]
pub enum ConfigurationError {
    #[error("Number of maximum signers in invalid.")]
    InvalidMaxSigners,
    #[error("Number of minimum signers in invalid.")]
    InvalidMinSigners,
}

#[uniffi::export]
fn validate_config(config: &Configuration) -> Result<(), ConfigurationError> {
    if config.min_signers < 2 {
        return Err(ConfigurationError::InvalidMinSigners);
    }

    if config.max_signers < 2 {
        return Err(ConfigurationError::InvalidMaxSigners);
    }

    if config.min_signers > config.max_signers {
        return Err(ConfigurationError::InvalidMinSigners);
    }
    
    Ok(())
}

pub struct FrostKeyPackage {
    pub identifier: String,
    pub data: Vec<u8>
}

impl FrostKeyPackage {
    fn from_key_package(key_package: &KeyPackage) -> Result<Self, Error> {
        let serialized_package = utils::json_bytes(key_package);
        let identifier = key_package.identifier();
        Ok(
            FrostKeyPackage {
                identifier: identifier.serialize().encode_hex(),
                data: serialized_package
            }
        )
    }
}
fn verify_and_get_key_package_from(secret_share: FrostSecretKeyShare) -> Result<FrostKeyPackage, Error> {
    
    let secret_share = secret_share.to_secret_share()
        .map_err(|_| frost::Error::InvalidSecretShare)?;
    
    frost::keys::KeyPackage::try_from(secret_share)
        .map_err(|_| frost::Error::IncorrectPackage)
        .map(|p| FrostKeyPackage::from_key_package(&p))?
}
