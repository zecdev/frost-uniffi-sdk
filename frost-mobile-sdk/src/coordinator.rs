#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;

#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

#[cfg(feature = "redpallas")]
use crate::randomizer::FrostRandomizer;
use crate::{
    participant::{FrostSignatureShare, FrostSigningCommitments},
    FrostPublicKeyPackage,
};
use frost::{
    round1::SigningCommitments, round2::SignatureShare, Error, Identifier, Signature,
    SigningPackage,
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
    let mut signing_commitments: BTreeMap<Identifier, SigningCommitments> = BTreeMap::new();

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

#[uniffi::export]
pub fn aggregate(
    signing_package: FrostSigningPackage,
    signature_shares: Vec<FrostSignatureShare>,
    pubkey_package: FrostPublicKeyPackage,
    #[cfg(feature = "redpallas")] randomizer: FrostRandomizer,
) -> Result<FrostSignature, CoordinationError> {
    let signing_package = signing_package
        .to_signing_package()
        .map_err(|_| CoordinationError::FailedToCreateSigningPackage)?;

    let mut shares: BTreeMap<Identifier, SignatureShare> = BTreeMap::new();

    for share in signature_shares {
        shares.insert(
            share
                .identifier
                .into_identifier()
                .map_err(|_| CoordinationError::IdentifierDeserializationError)?,
            share
                .to_signature_share()
                .map_err(|_| CoordinationError::SignatureShareDeserializationError)?,
        );
    }

    let public_key_package = pubkey_package
        .into_public_key_package()
        .map_err(|_| CoordinationError::PublicKeyPackageDeserializationError)?;

    #[cfg(feature = "redpallas")]
    let randomizer = randomizer
        .into_randomizer()
        .map_err(|_| CoordinationError::InvalidRandomizer)?;

    frost::aggregate(
        &signing_package,
        &shares,
        &public_key_package,
        #[cfg(feature = "redpallas")]
        &FrostRandomizer::randomizer_params(randomizer, &public_key_package),
    )
    .map_err(|e| CoordinationError::SignatureShareAggregationFailed {
        message: e.to_string(),
    })
    .map(|signature| FrostSignature::from_signature(signature))
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
    let signature = signature.to_signature().map_err(|e| {
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
    pub fn to_signature(&self) -> Result<Signature, Error> {
        Signature::deserialize(
            self.data[0..64]
                .try_into()
                .map_err(|_| Error::DeserializationError)?,
        )
    }

    fn from_signature(signature: Signature) -> FrostSignature {
        FrostSignature {
            data: signature.serialize().to_vec(),
        }
    }
}

impl FrostSigningPackage {
    pub fn to_signing_package(&self) -> Result<SigningPackage, Error> {
        SigningPackage::deserialize(&self.data)
    }

    pub fn from_signing_package(
        signing_package: SigningPackage,
    ) -> Result<FrostSigningPackage, Error> {
        let data = signing_package.serialize()?;
        Ok(FrostSigningPackage { data: data })
    }
}
