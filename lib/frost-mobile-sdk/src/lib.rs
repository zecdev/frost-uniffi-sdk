use frost::keys::SigningShare;
use frost::Ciphersuite;
use frost::Identifier;
#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;
use reddsa::frost::redpallas::frost::keys::CoefficientCommitment;
use reddsa::frost::redpallas::frost::keys::VerifiableSecretSharingCommitment;
use reddsa::frost::redpallas::frost::Identifier;
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

pub mod trusted_dealer;
use crate::frost::Error;
use std::collections::HashMap;

use uniffi;
use frost::keys::SecretShare;
use frost::keys::PublicKeyPackage;
uniffi::setup_scaffolding!();


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
            .serialize()
            .into_iter()
            .map(|c| hex::encode(c))
            .collect();
        FrostSecretKeyShare {
            identifier: hex::encode(identifier.serialize()),
            signing_share: hex::encode(signing_share.serialize()),
            commitment: commitment,
        }
    }

    fn to_secret_share(&self) -> Result<SecretShare, Error> {
        let hex_identifier = hex::decode(&mut self.identifier)
            .map_err(|_| frost::Error::DeserializationError)?;
        
        let slice_identifier = hex_identifier[0..32]
            .try_into()
            .map_err(|_| frost::Error::DeserializationError)?;

        let identifier = Identifier::deserialize(slice_identifier)?;

        let hex_signing_share = hex::decode(self.signing_share)
        .map_err(|_| frost::Error::DeserializationError)?;

        let signing_share = hex_signing_share[0..32]
            .try_into()
            .map_err(|_| frost::Error::DeserializationError)
            .map(|v| 
                SigningShare::deserialize(v)
                    .map_err(|_| frost::Error::DeserializationError)?
                )?;

        for s in self.commitment {
            let hex_commitment = hex::decode(s)
                    .map_err(|_| frost::Error::SerializationError)?;
        }
        let v = self.commitment
            .into_iter()
            .map(|s| 
                
            )
            .collect();

    
        let secret_share = SecretShare::new(identifier, signing_share, commitments);

        Ok(secret_share)
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