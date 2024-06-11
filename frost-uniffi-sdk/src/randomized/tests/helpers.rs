use std::collections::HashMap;

use rand::rngs::ThreadRng;
use reddsa::frost::redpallas as frost;

use crate::coordinator::new_signing_package;
use crate::randomized::participant::sign;
use crate::randomized::randomizer::FrostRandomizer;
use crate::{
    coordinator::{FrostSigningPackage, Message},
    participant::{FrostSignatureShare, FrostSigningCommitments, FrostSigningNonces},
    FrostKeyPackage, ParticipantIdentifier,
};

pub fn round_2(
    rng: &mut ThreadRng,
    nonces_map: &HashMap<ParticipantIdentifier, FrostSigningNonces>,
    key_packages: &HashMap<ParticipantIdentifier, FrostKeyPackage>,
    commitments_map: HashMap<ParticipantIdentifier, FrostSigningCommitments>,
    message: Message,
    randomizer: Option<FrostRandomizer>,
) -> (
    FrostSigningPackage,
    HashMap<ParticipantIdentifier, FrostSignatureShare>,
    Option<FrostRandomizer>,
) {
    let commitments = commitments_map.into_iter().map(|c| c.1).collect();
    let signing_package = new_signing_package(message, commitments).unwrap();
    let mut signature_shares = HashMap::new();

    let randomizer = match randomizer {
        Some(r) => r,
        None => {
            let randomizer =
                frost::round2::Randomizer::new(rng, &signing_package.to_signing_package().unwrap())
                    .unwrap();

            FrostRandomizer::from_randomizer(randomizer).unwrap()
        }
    };

    for participant_identifier in nonces_map.keys() {
        let key_package = key_packages[participant_identifier].clone();

        let nonces = nonces_map[participant_identifier].clone();

        let signature_share =
            sign(signing_package.clone(), nonces, key_package, &randomizer).unwrap();

        signature_shares.insert(participant_identifier.clone(), signature_share);
    }

    (signing_package, signature_shares, Some(randomizer))
}
