#[cfg(feature = "redpallas")]
type E = reddsa::frost::redpallas::PallasBlake2b512;
#[cfg(not(feature = "redpallas"))]
type E = frost_ed25519::Ed25519Sha512;

use crate::coordinator::new_signing_package;
use crate::randomized::participant::sign;
use crate::randomized::randomizer::FrostRandomizer;
use crate::FrostPublicKeyPackage;
use crate::{
    coordinator::{FrostSigningPackage, Message},
    participant::{FrostSignatureShare, FrostSigningCommitments, FrostSigningNonces},
    FrostKeyPackage, ParticipantIdentifier,
};
use frost::RandomizedParams;
use rand::rngs::ThreadRng;
use reddsa::frost::redpallas as frost;
use std::collections::HashMap;

pub fn round_2(
    rng: &mut ThreadRng,
    nonces_map: &HashMap<ParticipantIdentifier, FrostSigningNonces>,
    key_packages: &HashMap<ParticipantIdentifier, FrostKeyPackage>,
    commitments_map: HashMap<ParticipantIdentifier, FrostSigningCommitments>,
    pub_key: FrostPublicKeyPackage,
    message: Message,
) -> (
    FrostSigningPackage,
    HashMap<ParticipantIdentifier, FrostSignatureShare>,
    RandomizedParams,
) {
    let commitments = commitments_map.into_iter().map(|c| c.1).collect();
    let signing_package = new_signing_package(message, commitments).unwrap();
    let mut signature_shares = HashMap::new();

    let pub_keys = pub_key.into_public_key_package().unwrap();
    let pallas_signing_package = signing_package.to_signing_package().unwrap();
    let randomized_params =
        RandomizedParams::new(pub_keys.verifying_key(), &pallas_signing_package, rng).unwrap();
    let randomizer = randomized_params.randomizer();

    let frost_randomizer = FrostRandomizer::from_randomizer::<E>(*randomizer).unwrap();

    for participant_identifier in nonces_map.keys() {
        let key_package = key_packages[participant_identifier].clone();

        let nonces = nonces_map[participant_identifier].clone();

        let signature_share = sign(
            signing_package.clone(),
            nonces,
            key_package,
            &frost_randomizer,
        )
        .unwrap();

        signature_shares.insert(participant_identifier.clone(), signature_share);
    }

    (signing_package, signature_shares, randomized_params)
}
