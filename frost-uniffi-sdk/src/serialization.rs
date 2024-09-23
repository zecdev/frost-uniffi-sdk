#[cfg(not(feature = "redpallas"))]
type E = frost_ed25519::Ed25519Sha512;
#[cfg(feature = "redpallas")]
type E = reddsa::frost::redpallas::PallasBlake2b512;

use frost_core::{
    keys::{KeyPackage, PublicKeyPackage},
    round1::SigningCommitments,
    round2::SignatureShare,
};

use uniffi;

use crate::{
    participant::{FrostSignatureShare, FrostSigningCommitments},
    FrostError, FrostKeyPackage, FrostPublicKeyPackage, ParticipantIdentifier,
};

#[cfg(feature = "redpallas")]
use crate::randomized::randomizer::FrostRandomizer;

#[uniffi::export]
pub fn key_package_to_json(key_package: FrostKeyPackage) -> Result<String, FrostError> {
    let key_package = key_package
        .into_key_package::<E>()
        .map_err(FrostError::map_err)?;

    match serde_json::to_string(&key_package) {
        Ok(json) => Ok(json),
        Err(_) => Err(FrostError::SerializationError),
    }
}

#[uniffi::export]
pub fn json_to_key_package(key_package_json: String) -> Result<FrostKeyPackage, FrostError> {
    let key_package: KeyPackage<E> =
        serde_json::from_str(&key_package_json).map_err(|_| FrostError::DeserializationError)?;

    let frost_key_package =
        FrostKeyPackage::from_key_package::<E>(&key_package).map_err(FrostError::map_err)?;

    Ok(frost_key_package)
}

#[uniffi::export]
pub fn json_to_commitment(
    commitment_json: String,
    identifier: ParticipantIdentifier,
) -> Result<FrostSigningCommitments, FrostError> {
    let identifier = identifier
        .into_identifier::<E>()
        .map_err(FrostError::map_err)?;

    let commitments: SigningCommitments<E> =
        serde_json::from_str(&commitment_json).map_err(|_| FrostError::DeserializationError)?;

    let frost_commitments =
        FrostSigningCommitments::with_identifier_and_commitments::<E>(identifier, commitments)
            .map_err(FrostError::map_err)?;

    Ok(frost_commitments)
}
/// returns Raw Signing commitnments using serde_json
/// WARNING: The identifier you have in the `FrostSigningCommitments`
/// is not an original field of `SigningCommitments`, we've included
/// them as a nice-to-have.
#[uniffi::export]
pub fn commitment_to_json(commitment: FrostSigningCommitments) -> Result<String, FrostError> {
    let commitment = commitment
        .to_commitments::<E>()
        .map_err(FrostError::map_err)?;

    match serde_json::to_string(&commitment) {
        Ok(json) => Ok(json),
        Err(_) => Err(FrostError::SerializationError),
    }
}
#[cfg(feature = "redpallas")]
#[uniffi::export]
pub fn randomizer_to_json(randomizer: FrostRandomizer) -> Result<String, FrostError> {
    let randomizer = randomizer
        .into_randomizer::<E>()
        .map_err(FrostError::map_err)?;

    match serde_json::to_string(&randomizer) {
        Ok(json) => Ok(json),
        Err(_) => Err(FrostError::SerializationError),
    }
}

#[cfg(feature = "redpallas")]
#[uniffi::export]
pub fn json_to_randomizer(randomizer_json: String) -> Result<FrostRandomizer, FrostError> {
    let randomizer: String =
        serde_json::from_str(&randomizer_json).map_err(|_| FrostError::DeserializationError)?;

    super::randomized::randomizer::from_hex_string(randomizer)
}

#[uniffi::export]
pub fn public_key_package_to_json(
    public_key_package: FrostPublicKeyPackage,
) -> Result<String, FrostError> {
    let public_key_package = public_key_package
        .into_public_key_package()
        .map_err(FrostError::map_err)?;

    let string =
        serde_json::to_string(&public_key_package).map_err(|_| FrostError::SerializationError)?;

    Ok(string)
}

#[uniffi::export]
pub fn json_to_public_key_package(
    public_key_package_json: String,
) -> Result<FrostPublicKeyPackage, FrostError> {
    let public_key_package_json: PublicKeyPackage<E> =
        serde_json::from_str(&public_key_package_json)
            .map_err(|_| FrostError::DeserializationError)?;

    let frost_public_key_package =
        FrostPublicKeyPackage::from_public_key_package(public_key_package_json)
            .map_err(FrostError::map_err)?;

    Ok(frost_public_key_package)
}

#[uniffi::export]
pub fn signature_share_package_to_json(
    signature_share: FrostSignatureShare,
) -> Result<String, FrostError> {
    let signature_share = signature_share
        .to_signature_share::<E>()
        .map_err(FrostError::map_err)?;

    let json_share =
        serde_json::to_string(&signature_share).map_err(|_| FrostError::SerializationError)?;

    Ok(json_share)
}

#[uniffi::export]
pub fn json_to_signature_share(
    signature_share_json: String,
    identifier: ParticipantIdentifier,
) -> Result<FrostSignatureShare, FrostError> {
    let signature_share: SignatureShare<E> = serde_json::from_str(&signature_share_json)
        .map_err(|_| FrostError::DeserializationError)?;

    let identifier = identifier.into_identifier().map_err(FrostError::map_err)?;

    let share = FrostSignatureShare::from_signature_share(identifier, signature_share)
        .map_err(FrostError::map_err)?;

    Ok(share)
}
#[cfg(feature = "redpallas")]
#[cfg(test)]
mod test {
    use frost_core::{keys::KeyPackage, round2::SignatureShare, Identifier};
    use reddsa::frost::redpallas::PallasBlake2b512;

    use crate::{
        participant::{FrostSignatureShare, FrostSigningCommitments},
        serialization::{
            json_to_key_package, key_package_to_json, signature_share_package_to_json,
        },
        FrostKeyPackage, ParticipantIdentifier,
    };

    use super::{
        commitment_to_json, json_to_commitment, json_to_public_key_package,
        json_to_signature_share, public_key_package_to_json,
    };

    #[cfg(feature = "redpallas")]
    use super::{json_to_randomizer, randomizer_to_json};

    /// ```
    /// let json_package = r#"
    ///    {
    ///        "header": {
    ///            "version": 0,
    ///            "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
    ///        },
    ///        "identifier": "0300000000000000000000000000000000000000000000000000000000000000",
    ///        "signing_share": "112911acbfec15db78da8c4b1027b3ac75ce342447111226e1f15c93ca062d12",
    ///        "verifying_share": "60a1623d2a419d6a007177d13b75458b148e8ef1ba74e0fbc1d837d07b5f9706",
    ///        "verifying_key": "1fa942a303acbc3185dce72b2909ba838bb1efb16d500986f4afb7e04d43de85",
    ///        "min_signers": 2
    ///    }
    ///"#;
    /// ```
    #[cfg(feature = "redpallas")]
    #[test]
    fn test_key_package_roundtrip() {
        #[cfg(feature = "redpallas")]
        let json_package = r#"{"header":{"version":0,"ciphersuite":"FROST(Pallas, BLAKE2b-512)"},"identifier":"0300000000000000000000000000000000000000000000000000000000000000","signing_share":"112911acbfec15db78da8c4b1027b3ac75ce342447111226e1f15c93ca062d12","verifying_share":"60a1623d2a419d6a007177d13b75458b148e8ef1ba74e0fbc1d837d07b5f9706","verifying_key":"1fa942a303acbc3185dce72b2909ba838bb1efb16d500986f4afb7e04d43de85","min_signers":2}"#;

        let package: KeyPackage<PallasBlake2b512> = json_to_key_package(json_package.to_string())
            .unwrap()
            .into_key_package()
            .unwrap();

        let expected_identifier = Identifier::<PallasBlake2b512>::try_from(3).unwrap();

        assert_eq!(*package.identifier(), expected_identifier);

        let resulting_json = key_package_to_json(
            FrostKeyPackage::from_key_package::<PallasBlake2b512>(&package).unwrap(),
        )
        .unwrap();

        assert_eq!(resulting_json, json_package);
    }
    ///
    /// ```
    /// let share_json = r#"
    /// {
    ///   "header": {
    ///     "version": 0,
    ///     "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
    ///   },
    ///   "share": "d202ad8525dd0b238bdc969141ebe9b33402b71694fb6caffa78439634ee320d"
    /// }"#;"
    /// ```
    ///
    ///
    #[cfg(feature = "redpallas")]
    #[test]
    fn test_signature_share_serialization_round_trip() {
        let share_json = r#"{"header":{"version":0,"ciphersuite":"FROST(Pallas, BLAKE2b-512)"},"share":"d202ad8525dd0b238bdc969141ebe9b33402b71694fb6caffa78439634ee320d"}"#;

        let identifier = ParticipantIdentifier::from_identifier(
            Identifier::<PallasBlake2b512>::try_from(1).unwrap(),
        )
        .unwrap();
        let signature_share: SignatureShare<PallasBlake2b512> =
            json_to_signature_share(share_json.to_string(), identifier)
                .unwrap()
                .to_signature_share::<PallasBlake2b512>()
                .unwrap();

        let resulting_json = signature_share_package_to_json(
            FrostSignatureShare::from_signature_share(
                Identifier::<PallasBlake2b512>::try_from(3).unwrap(),
                signature_share,
            )
            .unwrap(),
        )
        .unwrap();

        assert_eq!(share_json, resulting_json);
    }

    /// ```
    /// let commitment_json = r#"
    ///    {
    ///        "header": {
    ///            "version": 0,
    ///            "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
    ///        },
    ///        "hiding": "ad737cac6f8e9ae3ae21a0de51556c8ea86c8e483b2418cf58300c036ebc100c",
    ///        "binding": "d36a016645420728b278f33fcaa45781840b4960625e3c7cf189cebb76f9a08c"
    ///    }
    ///    "#;
    /// ```
    #[cfg(feature = "redpallas")]
    #[test]
    fn test_commitment_serialization() {
        let commitment_json = r#"{"header":{"version":0,"ciphersuite":"FROST(Pallas, BLAKE2b-512)"},"hiding":"ad737cac6f8e9ae3ae21a0de51556c8ea86c8e483b2418cf58300c036ebc100c","binding":"d36a016645420728b278f33fcaa45781840b4960625e3c7cf189cebb76f9a08c"}"#;

        let participant = ParticipantIdentifier::from_identifier(
            Identifier::<PallasBlake2b512>::try_from(1).unwrap(),
        )
        .unwrap();
        let commitment =
            json_to_commitment(commitment_json.to_string(), participant.clone()).unwrap();

        let identifier = participant.into_identifier::<PallasBlake2b512>().unwrap();
        let json_commitment = commitment_to_json(
            FrostSigningCommitments::with_identifier_and_commitments(
                identifier,
                commitment.to_commitments::<PallasBlake2b512>().unwrap(),
            )
            .unwrap(),
        )
        .unwrap();

        assert_eq!(json_commitment, commitment_json);
    }

    /// ```
    ///
    ///
    /// let signature_share_json = r#"{
    ///    "header": {
    ///        "version": 0,
    ///        "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
    ///    },
    ///    "share": "307ebf4d5b7125407f359fa010cdca940a83e942fd389ecd67c6683ecee78f3e"
    /// }"#;
    /// ```
    #[cfg(feature = "redpallas")]
    #[test]
    fn test_signature_share_serialization() {
        let signature_share_json = r#"{"header":{"version":0,"ciphersuite":"FROST(Pallas, BLAKE2b-512)"},"share":"307ebf4d5b7125407f359fa010cdca940a83e942fd389ecd67c6683ecee78f3e"}"#;
        let identifier = ParticipantIdentifier::from_identifier(
            Identifier::<PallasBlake2b512>::try_from(1).unwrap(),
        )
        .unwrap();
        let signature_share =
            json_to_signature_share(signature_share_json.to_string(), identifier).unwrap();

        let json_signature_share = signature_share_package_to_json(signature_share).unwrap();

        assert_eq!(signature_share_json, json_signature_share);
    }

    #[cfg(feature = "redpallas")]
    #[test]
    fn test_randomizer_serialization() {
        let randomizer_json =
            r#""6fe2e6f26bca5f3a4bc1cd811327cdfc6a4581dc3fe1c101b0c5115a21697510""#;

        let randomizer = json_to_randomizer(randomizer_json.to_string()).unwrap();

        let json_randomizer = randomizer_to_json(randomizer).unwrap();

        assert_eq!(randomizer_json, json_randomizer);
    }

    /// ```
    /// let public_key_package = r#"
    ///    {
    ///        "header": {
    ///            "version": 0,
    ///            "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
    ///        },
    ///        "verifying_shares": {
    ///            "0100000000000000000000000000000000000000000000000000000000000000": "61a199916a3c2b64c5e566deb1ab18997282f9559f5b328f6ae50ca24b349f9d",
    ///            "0200000000000000000000000000000000000000000000000000000000000000": "389656dbe50a0b260c5b4e7ee953e8d81b0814cbdc112a6cd773d55de4202c0e",
    ///            "0300000000000000000000000000000000000000000000000000000000000000": "c0d94a637e113a82942bd0b886fa7d0e2256010bd42a9893c81df1a58e34ff8d"
    ///        },
    ///        "verifying_key": "93c3d1dca3634e26c7068342175b7dd5b3e3f3654494f6f6a3b77f96f3cb0a39"
    ///    }
    ///   "#;
    /// ```
    #[cfg(feature = "redpallas")]
    #[test]
    fn test_public_key_package_serialization() {
        let public_key_package_json = r#"{"header":{"version":0,"ciphersuite":"FROST(Pallas, BLAKE2b-512)"},"verifying_shares":{"0100000000000000000000000000000000000000000000000000000000000000":"61a199916a3c2b64c5e566deb1ab18997282f9559f5b328f6ae50ca24b349f9d","0200000000000000000000000000000000000000000000000000000000000000":"389656dbe50a0b260c5b4e7ee953e8d81b0814cbdc112a6cd773d55de4202c0e","0300000000000000000000000000000000000000000000000000000000000000":"c0d94a637e113a82942bd0b886fa7d0e2256010bd42a9893c81df1a58e34ff8d"},"verifying_key":"93c3d1dca3634e26c7068342175b7dd5b3e3f3654494f6f6a3b77f96f3cb0a39"}"#;

        let public_key_package =
            json_to_public_key_package(public_key_package_json.to_string()).unwrap();

        let json_public_key_package = public_key_package_to_json(public_key_package).unwrap();

        assert_eq!(public_key_package_json, json_public_key_package);
    }
}
