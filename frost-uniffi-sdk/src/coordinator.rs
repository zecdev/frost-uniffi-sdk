use frost_core as frost;

#[cfg(feature = "redpallas")]
type E = reddsa::frost::redpallas::PallasBlake2b512;
#[cfg(not(feature = "redpallas"))]
type E = frost_ed25519::Ed25519Sha512;

#[cfg(not(feature = "redpallas"))]
use crate::participant::FrostSignatureShare;

#[cfg(not(feature = "redpallas"))]
use frost::round2::SignatureShare;

use crate::{participant::FrostSigningCommitments, FrostPublicKeyPackage};

use frost::{
    round1::SigningCommitments, Ciphersuite, Error, Identifier, Signature, SigningPackage,
};
use std::collections::BTreeMap;
use uniffi;

#[derive(uniffi::Record, Clone)]
pub struct FrostSigningPackage {
    data: Vec<u8>,
}

#[derive(uniffi::Record, Clone)]
pub struct Message {
    pub data: Vec<u8>,
}

#[derive(uniffi::Record)]
pub struct FrostSignature {
    data: Vec<u8>,
}

#[derive(Debug, uniffi::Error, thiserror::Error)]
pub enum CoordinationError {
    #[error("Signing Package creation failed")]
    FailedToCreateSigningPackage,
    #[error("one or more of the signing commitments is invalid.")]
    InvalidSigningCommitment,
    #[error("Participant Identifier could not be deserialized.")]
    IdentifierDeserializationError,
    #[error("Signing Package could not be deserialized")]
    SigningPackageSerializationError,
    #[error("Signature Share could not be deserialized")]
    SignatureShareDeserializationError,
    #[error("Public Key Package could not be deserialized")]
    PublicKeyPackageDeserializationError,
    #[error("Signatures shares failed to be aggregated with error {message:?}")]
    SignatureShareAggregationFailed { message: String },
    #[cfg(feature = "redpallas")]
    #[error("An invalid Randomizer was provided.")]
    InvalidRandomizer,
}

#[uniffi::export]
pub fn new_signing_package(
    message: Message,
    commitments: Vec<FrostSigningCommitments>,
) -> Result<FrostSigningPackage, CoordinationError> {
    let mut signing_commitments: BTreeMap<Identifier<E>, SigningCommitments<E>> = BTreeMap::new();

    for c in commitments.into_iter() {
        let commitment = c
            .to_commitments()
            .map_err(|_| CoordinationError::InvalidSigningCommitment)?;
        let identifier = c
            .identifier
            .into_identifier()
            .map_err(|_| CoordinationError::IdentifierDeserializationError)?;
        signing_commitments.insert(identifier, commitment);
    }

    let signing_package = SigningPackage::new(signing_commitments, &message.data);

    let serialized_package = FrostSigningPackage::from_signing_package(signing_package)
        .map_err(|_| CoordinationError::SigningPackageSerializationError)?;

    Ok(serialized_package)
}

#[cfg(not(feature = "redpallas"))]
#[uniffi::export]
pub fn aggregate(
    signing_package: FrostSigningPackage,
    signature_shares: Vec<FrostSignatureShare>,
    pubkey_package: FrostPublicKeyPackage,
) -> Result<FrostSignature, CoordinationError> {
    let signing_package = signing_package
        .to_signing_package()
        .map_err(|_| CoordinationError::FailedToCreateSigningPackage)?;

    let mut shares: BTreeMap<Identifier<E>, SignatureShare<E>> = BTreeMap::new();

    for share in signature_shares {
        shares.insert(
            share
                .identifier
                .into_identifier()
                .map_err(|_| CoordinationError::IdentifierDeserializationError)?,
            share
                .to_signature_share::<E>()
                .map_err(|_| CoordinationError::SignatureShareDeserializationError)?,
        );
    }

    let public_key_package = pubkey_package
        .into_public_key_package()
        .map_err(|_| CoordinationError::PublicKeyPackageDeserializationError)?;

    let signature =
        frost::aggregate(&signing_package, &shares, &public_key_package).map_err(|e| {
            CoordinationError::SignatureShareAggregationFailed {
                message: e.to_string(),
            }
        })?;

    Ok(FrostSignature {
        data: signature.serialize().map_err(|e| {
            CoordinationError::SignatureShareAggregationFailed {
                message: e.to_string(),
            }
        })?,
    })
}

#[derive(Debug, uniffi::Error, thiserror::Error)]
pub enum FrostSignatureVerificationError {
    #[error("Public Key Package is invalid")]
    InvalidPublicKeyPackage,
    #[error("FROST signature is invalid. Reason: {reason:?}")]
    ValidationFailed { reason: String },
}

#[uniffi::export]
pub fn verify_signature(
    message: Message,
    signature: FrostSignature,
    pubkey: FrostPublicKeyPackage,
) -> Result<(), FrostSignatureVerificationError> {
    let signature = signature.to_signature::<E>().map_err(|e| {
        FrostSignatureVerificationError::ValidationFailed {
            reason: e.to_string(),
        }
    })?;

    let pubkey = pubkey
        .into_public_key_package()
        .map_err(|_| FrostSignatureVerificationError::InvalidPublicKeyPackage)?;

    pubkey
        .verifying_key()
        .verify(&message.data, &signature)
        .map_err(|e| FrostSignatureVerificationError::ValidationFailed {
            reason: e.to_string(),
        })
}

impl FrostSignature {
    pub fn to_signature<C: Ciphersuite>(&self) -> Result<Signature<E>, Error<E>> {
        let bytes: [u8; 64] = self.data[0..64]
            .try_into()
            .map_err(|_| Error::DeserializationError)?;
        Signature::<E>::deserialize(&bytes)
    }

    pub fn from_signature<C: Ciphersuite>(
        signature: Signature<C>,
    ) -> Result<FrostSignature, Error<C>> {
        let data = signature.serialize()?;
        Ok(FrostSignature { data })
    }
}

impl FrostSigningPackage {
    pub fn to_signing_package<C: Ciphersuite>(&self) -> Result<SigningPackage<C>, Error<C>> {
        SigningPackage::deserialize(&self.data)
    }

    pub fn from_signing_package<C: Ciphersuite>(
        signing_package: SigningPackage<C>,
    ) -> Result<FrostSigningPackage, Error<C>> {
        let data = signing_package.serialize()?;
        Ok(FrostSigningPackage { data })
    }
}
