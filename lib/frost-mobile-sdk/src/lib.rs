#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

pub mod trusted_dealer;
use crate::frost::Error;
use std::collections::HashMap;

use uniffi;
use frost::keys::SecretShare;
use frost::keys::PublicKeyPackage;
uniffi::setup_scaffolding!();

#[derive(uniffi::Enum)]
pub enum Ciphersuite {
    RedPallas
}
#[derive(uniffi::Record)]
pub struct Header {
    pub version: u8, 
    pub suite: Ciphersuite, 
}
#[derive(uniffi::Record)]
pub struct FrostSecretKeyShare {
    pub identifier: String,
    pub signing_share: String,
    pub commitment: Vec<String>
}

#[derive(uniffi::Record)]
pub struct FrostPublicKeyPackage {
    pub verifying_shares: HashMap<String, String>,
    pub verifying_key: String
}

impl FrostSecretKeyShare {
    fn from_secret_share(secret_share: SecretShare) -> FrostSecretKeyShare {
        let identifier = secret_share.identifier();
        let signing_share = secret_share.signing_share();
        let commitment = secret_share.commitment()
        .coefficients()
        .iter()
        .map(|c| hex::encode(c.serialize()))
        .collect();


        FrostSecretKeyShare {
            identifier: hex::encode(identifier.serialize()),
            signing_share: hex::encode(signing_share.serialize()),
            commitment: commitment,
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
    #[error("The Secret can't be empty")]
    InvalidEmptySecret,
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

    if config.secret.is_empty() {
        return Err(ConfigurationError::InvalidEmptySecret)
    }
    
    Ok(())
}

#[derive(uniffi::Object)]
pub struct Test {
    pub value: i64
}

#[uniffi::export]
impl Test { 
    fn do_something(&self) {
        print!("hello");
    }
}