#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;

#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

#[cfg(feature = "redpallas")]
pub mod randomized;

pub mod coordinator;
pub mod participant;
pub mod trusted_dealer;
use crate::frost::Error;
use crate::trusted_dealer::{trusted_dealer_keygen, trusted_dealer_keygen_from_configuration};
use frost::keys::{KeyPackage, PublicKeyPackage, SecretShare};
use frost::{
    keys::{IdentifierList, VerifyingShare},
    Identifier, VerifyingKey,
};
use hex::ToHex;
use rand::thread_rng;
use std::{
    collections::{BTreeMap, HashMap},
    hash::Hash,
};

uniffi::setup_scaffolding!();

#[derive(uniffi::Record)]
pub struct ParticipantList {
    pub identifiers: Vec<ParticipantIdentifier>,
}

#[derive(uniffi::Record, Hash, Eq, PartialEq, Clone)]
pub struct ParticipantIdentifier {
    pub data: Vec<u8>,
}

#[derive(uniffi::Record)]
pub struct TrustedKeyGeneration {
    pub secret_shares: HashMap<ParticipantIdentifier, FrostSecretKeyShare>,
    pub public_key_package: FrostPublicKeyPackage,
}

#[derive(uniffi::Record)]
pub struct FrostSecretKeyShare {
    pub identifier: ParticipantIdentifier,
    pub data: Vec<u8>,
}

#[derive(uniffi::Record, Clone)]
pub struct FrostPublicKeyPackage {
    pub verifying_shares: HashMap<ParticipantIdentifier, String>,
    pub verifying_key: String,
}

#[derive(uniffi::Record, Clone)]
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
    #[error("One or more of the custom Identifiers provided are invalid.")]
    InvalidIdentifier,
    #[error("There's a problem with this configuration.")]
    UnknownError,
}

#[derive(Debug, thiserror::Error, uniffi::Error)]
pub enum FrostError {
    #[error("Value could not be serialized.")]
    SerializationError,
    #[error("Value could not be deserialized.")]
    DeserializationError,
    #[error("Key Package is invalid")]
    InvalidKeyPackage,
    #[error("Secret Key couldn't be verified")]
    InvalidSecretKey,
    #[error("Unknown Identifier")]
    UnknownIdentifier,
}
#[uniffi::export]
pub fn validate_config(config: &Configuration) -> Result<(), ConfigurationError> {
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

#[derive(uniffi::Record, Clone)]
pub struct FrostKeyPackage {
    pub identifier: String,
    pub data: Vec<u8>,
}

#[uniffi::export]
pub fn verify_and_get_key_package_from(
    secret_share: FrostSecretKeyShare,
) -> Result<FrostKeyPackage, FrostError> {
    secret_share
        .into_key_package()
        .map_err(|_| FrostError::InvalidSecretKey)
}

#[uniffi::export]
pub fn trusted_dealer_keygen_from(
    configuration: Configuration,
) -> Result<TrustedKeyGeneration, ConfigurationError> {
    let (pubkey, secret_shares) = trusted_dealer_keygen_from_configuration(&configuration)
        .map_err(|e| match e {
            Error::InvalidMaxSigners => ConfigurationError::InvalidMaxSigners,
            Error::InvalidMinSigners => ConfigurationError::InvalidMinSigners,
            _ => ConfigurationError::UnknownError,
        })?;

    Ok(TrustedKeyGeneration {
        public_key_package: pubkey,
        secret_shares,
    })
}

#[uniffi::export]
pub fn trusted_dealer_keygen_with_identifiers(
    configuration: Configuration,
    participants: ParticipantList,
) -> Result<TrustedKeyGeneration, ConfigurationError> {
    if configuration.max_signers as usize != participants.identifiers.len() {
        return Err(ConfigurationError::InvalidMaxSigners);
    }

    let mut custom_identifiers: Vec<Identifier> =
        Vec::with_capacity(participants.identifiers.capacity());

    for identifier in participants.identifiers.clone().into_iter() {
        let identifier = identifier
            .into_identifier()
            .map_err(|_| ConfigurationError::InvalidIdentifier)?;
        custom_identifiers.push(identifier);
    }

    let list = IdentifierList::Custom(&custom_identifiers);

    let mut rng = thread_rng();

    let (shares, pubkey) =
        trusted_dealer_keygen(&configuration, list, &mut rng).map_err(|e| match e {
            Error::InvalidMaxSigners => ConfigurationError::InvalidMaxSigners,
            Error::InvalidMinSigners => ConfigurationError::InvalidMinSigners,
            _ => ConfigurationError::UnknownError,
        })?;

    let pubkey = FrostPublicKeyPackage::from_public_key_package(pubkey)
        .map_err(|_| ConfigurationError::UnknownError)?;

    let mut hash_map: HashMap<ParticipantIdentifier, FrostSecretKeyShare> = HashMap::new();

    for (k, v) in shares {
        hash_map.insert(
            ParticipantIdentifier::from_identifier(k)
                .map_err(|_| ConfigurationError::InvalidIdentifier)?,
            FrostSecretKeyShare::from_secret_share(v)
                .map_err(|_| ConfigurationError::UnknownError)?,
        );
    }

    Ok(TrustedKeyGeneration {
        public_key_package: pubkey,
        secret_shares: hash_map,
    })
}

impl FrostKeyPackage {
    pub fn from_key_package(key_package: &KeyPackage) -> Result<Self, Error> {
        let serialized_package = key_package.serialize()?;
        let identifier = key_package.identifier();
        Ok(FrostKeyPackage {
            identifier: identifier.serialize().encode_hex(),
            data: serialized_package,
        })
    }

    pub fn into_key_package(&self) -> Result<KeyPackage, Error> {
        KeyPackage::deserialize(&self.data)
    }
}

impl ParticipantIdentifier {
    pub fn from_identifier(identifier: Identifier) -> Result<ParticipantIdentifier, Error> {
        Ok(ParticipantIdentifier {
            data: identifier.clone().serialize().to_vec(),
        })
    }

    pub fn into_identifier(&self) -> Result<Identifier, Error> {
        let raw_bytes = self.data[0..32]
            .try_into()
            .map_err(|_| Error::DeserializationError)?;

        Identifier::deserialize(&raw_bytes)
    }
}

impl FrostSecretKeyShare {
    pub fn from_secret_share(
        secret_share: SecretShare,
    ) -> Result<FrostSecretKeyShare, frost::Error> {
        let identifier = ParticipantIdentifier::from_identifier(*secret_share.identifier())?;
        let serialized_share = secret_share
            .serialize()
            .map_err(|_| frost::Error::SerializationError)?;

        Ok(FrostSecretKeyShare {
            identifier,
            data: serialized_share,
        })
    }

    pub fn to_secret_share(&self) -> Result<SecretShare, Error> {
        let identifier = self.identifier.into_identifier()?;

        let secret_share =
            SecretShare::deserialize(&self.data).map_err(|_| frost::Error::SerializationError)?;

        if identifier != *secret_share.identifier() {
            Err(frost::Error::UnknownIdentifier)
        } else {
            Ok(secret_share)
        }
    }

    pub fn into_key_package(&self) -> Result<FrostKeyPackage, FrostError> {
        let secret_share = self
            .to_secret_share()
            .map_err(|_| FrostError::InvalidSecretKey)?;

        let key_package = frost::keys::KeyPackage::try_from(secret_share)
            .map_err(|_| FrostError::InvalidSecretKey)?;

        FrostKeyPackage::from_key_package(&key_package).map_err(|_| FrostError::SerializationError)
    }
}

impl FrostPublicKeyPackage {
    pub fn from_public_key_package(
        key_package: PublicKeyPackage,
    ) -> Result<FrostPublicKeyPackage, Error> {
        let verifying_shares = key_package.verifying_shares();
        let verifying_key = key_package.verifying_key();

        let mut shares: HashMap<ParticipantIdentifier, String> = HashMap::new();

        for (k, v) in verifying_shares {
            shares.insert(
                ParticipantIdentifier::from_identifier(*k)?,
                hex::encode(v.serialize()),
            );
        }

        Ok(Self {
            verifying_shares: shares,
            verifying_key: hex::encode(verifying_key.serialize()),
        })
    }

    pub fn into_public_key_package(&self) -> Result<PublicKeyPackage, Error> {
        let raw_verifying_key =
            hex::decode(self.verifying_key.clone()).map_err(|_| Error::DeserializationError)?;

        let verifying_key_bytes = raw_verifying_key[0..32]
            .try_into()
            .map_err(|_| Error::DeserializationError)?;

        let verifying_key = VerifyingKey::deserialize(verifying_key_bytes)
            .map_err(|_| Error::DeserializationError)?;

        let mut btree_map: BTreeMap<Identifier, VerifyingShare> = BTreeMap::new();
        for (k, v) in self.verifying_shares.clone() {
            let identifier = k.into_identifier()?;

            let raw_verifying_share = hex::decode(v).map_err(|_| Error::DeserializationError)?;

            let verifying_share_bytes: [u8; 32] = raw_verifying_share[0..32]
                .try_into()
                .map_err(|_| Error::DeserializationError)?;

            let verifying_share = VerifyingShare::deserialize(verifying_share_bytes)?;

            btree_map.insert(identifier, verifying_share);
        }

        Ok(PublicKeyPackage::new(btree_map, verifying_key))
    }
}
