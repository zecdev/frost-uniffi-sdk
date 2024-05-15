use std::collections::BTreeMap;

use frost::{round1::SigningCommitments, Error, Identifier, SigningPackage};
#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

use uniffi;
use rand::thread_rng;

use crate::participant::FrostSigningCommitments;

#[derive(uniffi::Record)]
pub struct FrostSigningPackage {
    data: Vec<u8>
}

impl FrostSigningPackage {
    pub (crate) fn to_signing_package(&self) -> Result<SigningPackage, Error> {
        SigningPackage::deserialize(&self.data)
    }
}

#[derive(uniffi::Record)]
pub struct Message {
    data: Vec<u8>
}

#[derive(Debug, uniffi::Error, thiserror::Error)]
pub enum CoordinationError {
    #[error("Signing Package creation failed")]
    FailedToCreateSigningPackage,
    #[error("one or more of the signing commitments is invalid.")]
    InvalidSigningCommitment,
    #[error("Participant Identifier could not be deserialized.")]
    IdentifierDeserializationError,
    #[error("Signing Package could not be serialized")]
    SigningPackageSerializationError,
}

#[uniffi::export]
pub fn new_signing_package(message: Message, commitments: Vec<FrostSigningCommitments>) -> Result<FrostSigningPackage, CoordinationError> {

    let mut signing_commitments: BTreeMap<Identifier,SigningCommitments> = BTreeMap::new();

    for c in commitments.into_iter() {
        let commitment = c.to_commitments()
            .map_err(|_| CoordinationError::InvalidSigningCommitment)?;
        let identifier = c.identifier.into_identifier()
            .map_err(|_| CoordinationError::IdentifierDeserializationError)?;
        signing_commitments.insert(identifier, commitment);
    }
    
    let serialized_package = SigningPackage::new(signing_commitments, &message.data)
        .serialize()
        .map_err(|_| CoordinationError::SigningPackageSerializationError)?;

    Ok(
        FrostSigningPackage { data: serialized_package }
    )
}