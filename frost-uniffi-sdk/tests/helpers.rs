#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

use frost_uniffi_sdk::{
    coordinator::{new_signing_package, FrostSigningPackage, Message},
    participant::{sign, FrostSignatureShare, FrostSigningCommitments, FrostSigningNonces},
    FrostKeyPackage, FrostSecretKeyShare, ParticipantIdentifier,
};
use rand::rngs::ThreadRng;
use std::collections::HashMap;

pub fn key_package(
    shares: &HashMap<ParticipantIdentifier, FrostSecretKeyShare>,
) -> HashMap<ParticipantIdentifier, FrostKeyPackage> {
    let mut key_packages: HashMap<_, _> = HashMap::new();

    for (identifier, secret_share) in shares {
        let key_package = secret_share.into_key_package().unwrap();
        key_packages.insert(identifier.clone(), key_package);
    }

    key_packages
}

pub fn round_1(
    mut rng: &mut ThreadRng,
    key_packages: &HashMap<ParticipantIdentifier, FrostKeyPackage>,
) -> (
    HashMap<ParticipantIdentifier, FrostSigningNonces>,
    HashMap<ParticipantIdentifier, FrostSigningCommitments>,
) {
    // Participant Round 1

    let mut nonces_map = HashMap::new();
    let mut commitments_map = HashMap::new();

    for (participant, key_package) in key_packages {
        let (nonces, commitments) = frost::round1::commit(
            key_package.into_key_package().unwrap().signing_share(),
            &mut rng,
        );
        nonces_map.insert(
            participant.clone(),
            FrostSigningNonces::from_nonces(nonces).unwrap(),
        );
        commitments_map.insert(
            participant.clone(),
            FrostSigningCommitments::with_identifier_and_commitments(
                participant.into_identifier().unwrap(),
                commitments,
            )
            .unwrap(),
        );
    }

    (nonces_map, commitments_map)
}

#[cfg(not(feature = "redpallas"))]
pub fn round_2(
    nonces_map: &HashMap<ParticipantIdentifier, FrostSigningNonces>,
    key_packages: &HashMap<ParticipantIdentifier, FrostKeyPackage>,
    commitments_map: HashMap<ParticipantIdentifier, FrostSigningCommitments>,
    message: Message,
) -> (
    FrostSigningPackage,
    HashMap<ParticipantIdentifier, FrostSignatureShare>,
) {
    let commitments = commitments_map.into_iter().map(|c| c.1).collect();
    let signing_package = new_signing_package(message, commitments).unwrap();
    let mut signature_shares = HashMap::new();

    for participant_identifier in nonces_map.keys() {
        let key_package = key_packages[participant_identifier].clone();

        let nonces = nonces_map[participant_identifier].clone();

        let signature_share = sign(signing_package.clone(), nonces, key_package).unwrap();

        signature_shares.insert(participant_identifier.clone(), signature_share);
    }

    (signing_package, signature_shares)
}
