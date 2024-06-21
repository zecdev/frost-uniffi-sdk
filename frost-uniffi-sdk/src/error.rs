use crate::{FrostError, ParticipantIdentifier};
use frost_core::{Ciphersuite, Error};

impl FrostError {
    pub(crate) fn map_err<C: Ciphersuite>(e: frost_core::Error<C>) -> FrostError {
        match e {
            Error::InvalidMinSigners => Self::InvalidMinSigners,
            Error::InvalidMaxSigners => Self::InvalidMaxSigners,
            Error::InvalidCoefficients => Self::InvalidCoefficients,
            Error::MalformedIdentifier => Self::MalformedIdentifier,
            Error::DuplicatedIdentifier => Self::DuplicatedIdentifier,
            Error::UnknownIdentifier => Self::UnknownIdentifier,
            Error::IncorrectNumberOfIdentifiers => Self::IncorrectNumberOfIdentifiers,
            Error::MalformedSigningKey => Self::MalformedSigningKey,
            Error::MalformedVerifyingKey => Self::MalformedVerifyingKey,
            Error::MalformedSignature => Self::MalformedSignature,
            Error::InvalidSignature => Self::InvalidSignature,
            Error::DuplicatedShares => Self::DuplicatedShares,
            Error::IncorrectNumberOfShares => Self::IncorrectNumberOfShares,
            Error::IdentityCommitment => Self::IdentityCommitment,
            Error::MissingCommitment => Self::MissingCommitment,
            Error::IncorrectCommitment => Self::IncorrectCommitment,
            Error::IncorrectNumberOfCommitments => Self::IncorrectNumberOfCommitments,
            Error::InvalidSignatureShare { culprit } => {
                match ParticipantIdentifier::from_identifier(culprit) {
                    Ok(p) => Self::InvalidProofOfKnowledge { culprit: p },
                    Err(_) => Self::MalformedIdentifier,
                }
            }
            Error::InvalidSecretShare => Self::InvalidSecretShare,
            Error::PackageNotFound => Self::PackageNotFound,
            Error::IncorrectNumberOfPackages => Self::IncorrectNumberOfPackages,
            Error::IncorrectPackage => Self::IncorrectPackage,
            Error::DKGNotSupported => Self::DKGNotSupported,
            Error::InvalidProofOfKnowledge { culprit } => {
                match ParticipantIdentifier::from_identifier(culprit) {
                    Ok(p) => Self::InvalidProofOfKnowledge { culprit: p },
                    Err(_) => Self::MalformedIdentifier,
                }
            }
            Error::FieldError(error) => Self::FieldError {
                message: error.to_string(),
            },
            Error::GroupError(error) => Self::GroupError {
                message: error.to_string(),
            },
            Error::InvalidCoefficient => Self::InvalidCoefficient,
            Error::IdentifierDerivationNotSupported => Self::IdentifierDerivationNotSupported,
            Error::SerializationError => Self::SerializationError,
            Error::DeserializationError => Self::DeserializationError,
            _ => Self::UnknownIdentifier,
        }
    }
}
