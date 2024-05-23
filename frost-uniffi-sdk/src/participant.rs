#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

use frost::{
    round1::{SigningCommitments, SigningNonces},
    round2::SignatureShare,
    Error, Identifier,
};
use rand::thread_rng;
use uniffi;

use crate::{FrostSecretKeyShare, ParticipantIdentifier};

#[cfg(not(feature = "redpallas"))]
use crate::{coordinator::FrostSigningPackage, FrostKeyPackage};

#[derive(uniffi::Record, Clone)]
pub struct FrostSigningNonces {
    pub data: Vec<u8>,
}

impl FrostSigningNonces {
    pub(super) fn to_signing_nonces(&self) -> Result<SigningNonces, Error> {
        SigningNonces::deserialize(&self.data)
    }

    pub fn from_nonces(nonces: SigningNonces) -> Result<FrostSigningNonces, Error> {
        let data = nonces.serialize()?;
        Ok(FrostSigningNonces { data })
    }
}

#[derive(uniffi::Record)]
pub struct FrostSigningCommitments {
    pub identifier: ParticipantIdentifier,
    pub data: Vec<u8>,
}

impl FrostSigningCommitments {
    pub(crate) fn to_commitments(&self) -> Result<SigningCommitments, Error> {
        SigningCommitments::deserialize(&self.data)
    }

    pub fn with_identifier_and_commitments(
        identifier: Identifier,
        commitments: SigningCommitments,
    ) -> Result<FrostSigningCommitments, Error> {
        Ok(FrostSigningCommitments {
            identifier: ParticipantIdentifier::from_identifier(identifier)?,
            data: commitments.serialize()?,
        })
    }
}

#[derive(Debug, uniffi::Error, thiserror::Error)]
pub enum Round1Error {
    #[error("Provided Key Package is invalid.")]
    InvalidKeyPackage,
    #[error("Nonce could not be serialized.")]
    NonceSerializationError,
    #[error("Commitment could not be serialized.")]
    CommitmentSerializationError,
}

#[derive(Debug, uniffi::Error, thiserror::Error)]
pub enum Round2Error {
    #[error("Provided Key Package is invalid.")]
    InvalidKeyPackage,
    #[error("Nonce could not be serialized.")]
    NonceSerializationError,
    #[error("Commitment could not be serialized.")]
    CommitmentSerializationError,
    #[error("Could not deserialize Signing Package.")]
    SigningPackageDeserializationError,
    #[error("Failed to sign message with error: {message:?}")]
    SigningFailed { message: String },
    #[cfg(feature = "redpallas")]
    #[error("Invalid Randomizer.")]
    InvalidRandomizer,
}

#[derive(uniffi::Record)]
pub struct FirstRoundCommitment {
    pub nonces: FrostSigningNonces,
    pub commitments: FrostSigningCommitments,
}

#[uniffi::export]
pub fn generate_nonces_and_commitments(
    secret_share: FrostSecretKeyShare,
) -> Result<FirstRoundCommitment, Round1Error> {
    let mut rng = thread_rng();

    let secret_share = secret_share
        .to_secret_share()
        .map_err(|_| Round1Error::InvalidKeyPackage)?;

    let _ = secret_share
        .verify()
        .map_err(|_| Round1Error::InvalidKeyPackage)?;

    let signing_share = secret_share.signing_share();
    let (nonces, commitments) = frost::round1::commit(signing_share, &mut rng);

    Ok(FirstRoundCommitment {
        nonces: FrostSigningNonces::from_nonces(nonces)
            .map_err(|_| Round1Error::NonceSerializationError)?,
        commitments: FrostSigningCommitments::with_identifier_and_commitments(
            *secret_share.identifier(),
            commitments,
        )
        .map_err(|_| Round1Error::CommitmentSerializationError)?,
    })
}

#[derive(uniffi::Record)]
pub struct FrostSignatureShare {
    pub identifier: ParticipantIdentifier,
    pub data: Vec<u8>,
}

impl FrostSignatureShare {
    pub fn to_signature_share(&self) -> Result<SignatureShare, Error> {
        let bytes = self.data[0..32]
            .try_into()
            .map_err(|_| Error::DeserializationError)?;

        SignatureShare::deserialize(bytes)
    }

    pub fn from_signature_share(
        identifier: Identifier,
        share: SignatureShare,
    ) -> Result<FrostSignatureShare, Error> {
        Ok(FrostSignatureShare {
            identifier: ParticipantIdentifier::from_identifier(identifier)?,
            data: share.serialize().to_vec(),
        })
    }
}

#[cfg(not(feature = "redpallas"))]
#[uniffi::export]
pub fn sign(
    signing_package: FrostSigningPackage,
    nonces: FrostSigningNonces,
    key_package: FrostKeyPackage,
) -> Result<FrostSignatureShare, Round2Error> {
    let signing_package = signing_package
        .to_signing_package()
        .map_err(|_| Round2Error::SigningPackageDeserializationError)?;

    let nonces = nonces
        .to_signing_nonces()
        .map_err(|_| Round2Error::NonceSerializationError)?;

    let key_package = key_package
        .into_key_package()
        .map_err(|_| Round2Error::InvalidKeyPackage)?;

    let identifier = *key_package.identifier();

    let share = frost::round2::sign(&signing_package, &nonces, &key_package).map_err(|e| {
        Round2Error::SigningFailed {
            message: e.to_string(),
        }
    })?;

    FrostSignatureShare::from_signature_share(identifier, share).map_err(|e| {
        Round2Error::SigningFailed {
            message: e.to_string(),
        }
    })
}
