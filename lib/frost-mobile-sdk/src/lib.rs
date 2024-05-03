use std::collections::HashMap;
use thiserror::Error;
use reddsa::frost::redpallas as frost;
use uniffi;
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
pub struct FrostKeyShare {
    pub header: Header,
    pub identifier: String,
    pub signing_share: String,
    pub commitment: Vec<String>,
}

#[derive(uniffi::Record)]
pub struct FrostPublicKeyPackage {
    pub header: Header,
    pub verifying_shares: HashMap<String, String>,
    pub verifying_key: String
}

impl FrostPublicKeyPackage {
   
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