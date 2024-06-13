#[cfg(feature = "redpallas")]
type E = reddsa::frost::redpallas::PallasBlake2b512;
#[cfg(not(feature = "redpallas"))]
type E = frost_ed25519::Ed25519Sha512;

use frost_core::Identifier;
#[cfg(not(feature = "redpallas"))]
use frost_uniffi_sdk::coordinator::{aggregate, verify_signature};

#[cfg(feature = "redpallas")]
use frost_uniffi_sdk::{
    coordinator::verify_signature,
    randomized::{coordinator::aggregate, tests::helpers::round_2},
};

use rand::thread_rng;
use std::{collections::HashMap, sync::Arc};

use frost_uniffi_sdk::{
    coordinator::Message, FrostKeyPackage, FrostPublicKeyPackage, ParticipantIdentifier,
};

use frost_uniffi_sdk::dkg::lib::{
    part_1, part_2, part_3, DKGRound1Package, DKGRound1SecretPackage, DKGRound2Package,
    DKGRound2SecretPackage,
};
mod helpers;
use helpers::round_1;

#[cfg(not(feature = "redpallas"))]
use helpers::round_2;

struct Participant {
    pub identifier: ParticipantIdentifier,
    min_signers: u16,
    max_signers: u16,
    secret1: Option<DKGRound1SecretPackage>,
    secret2: Option<DKGRound2SecretPackage>,
    round1_packages: Option<HashMap<ParticipantIdentifier, DKGRound1Package>>,
    key_package: Option<FrostKeyPackage>,
    round2_packages: HashMap<ParticipantIdentifier, DKGRound2Package>,
    public_key_package: Option<FrostPublicKeyPackage>,
}

impl Participant {
    fn do_part1(&mut self) -> DKGRound1Package {
        let part1 = part_1(self.identifier.clone(), self.max_signers, self.min_signers).unwrap();

        self.secret1 = Some(part1.secret.clone());

        part1.package.clone()
    }

    fn do_part2(
        &mut self,
        round1_packages: HashMap<ParticipantIdentifier, DKGRound1Package>,
    ) -> Vec<DKGRound2Package> {
        let mut round1_packages_except_myself = round1_packages.clone();
        round1_packages_except_myself.remove(&self.identifier);

        self.round1_packages = Some(round1_packages_except_myself);

        let r1_pkg = self.round1_packages.clone().unwrap();
        let part2 = part_2(Arc::new(self.secret1.clone().unwrap()), r1_pkg).unwrap();

        // keep secret for later
        self.secret2 = Some(part2.clone().secret.clone());

        part2.clone().packages.clone()
    }

    fn do_part3(&mut self) -> (FrostKeyPackage, FrostPublicKeyPackage) {
        let part3 = part_3(
            Arc::new(self.secret2.clone().unwrap()),
            self.round1_packages.clone().unwrap(),
            self.round2_packages.clone(),
        )
        .unwrap();

        self.key_package = Some(part3.key_package.clone());
        self.public_key_package = Some(part3.public_key_package.clone());

        (part3.key_package, part3.public_key_package)
    }

    fn receive_round2_package(&mut self, sender: ParticipantIdentifier, package: DKGRound2Package) {
        self.round2_packages.insert(sender.clone(), package.clone());
    }
}

#[test]
fn test_dkg_from_3_participants() {
    let mut participants: HashMap<ParticipantIdentifier, Participant> = HashMap::new();

    let p1_id: Identifier<E> = Identifier::try_from(1).unwrap();
    let p2_id: Identifier<E> = Identifier::try_from(2).unwrap();
    let p3_id: Identifier<E> = Identifier::try_from(3).unwrap();

    let p1_identifier = ParticipantIdentifier::from_identifier(p1_id).unwrap();
    let p2_identifier = ParticipantIdentifier::from_identifier(p2_id).unwrap();
    let p3_identifier = ParticipantIdentifier::from_identifier(p3_id).unwrap();

    participants.insert(
        p1_identifier.clone(),
        Participant {
            identifier: p1_identifier.clone(),
            min_signers: 2,
            max_signers: 3,
            secret1: None,
            secret2: None,
            round1_packages: None,
            key_package: None,
            round2_packages: HashMap::new(),
            public_key_package: None,
        },
    );

    participants.insert(
        p2_identifier.clone(),
        Participant {
            identifier: p2_identifier.clone(),
            min_signers: 2,
            max_signers: 3,
            secret1: None,
            secret2: None,
            round1_packages: None,
            key_package: None,
            round2_packages: HashMap::new(),
            public_key_package: None,
        },
    );

    participants.insert(
        p3_identifier.clone(),
        Participant {
            identifier: p3_identifier.clone(),
            min_signers: 2,
            max_signers: 3,
            secret1: None,
            secret2: None,
            round1_packages: None,
            key_package: None,
            round2_packages: HashMap::new(),
            public_key_package: None,
        },
    );

    // gather part1 from all participants
    let mut round1_packages: HashMap<ParticipantIdentifier, DKGRound1Package> = HashMap::new();

    for (identifier, participant) in participants.iter_mut() {
        round1_packages.insert(identifier.clone(), participant.do_part1());
    }

    // do part2. this HashMap key will contain who generated the packages and the values will
    // be the recipients of them.
    let mut round2_packages: HashMap<ParticipantIdentifier, Vec<DKGRound2Package>> = HashMap::new();

    for (identifier, participant) in participants.iter_mut() {
        let packages = participant.do_part2(round1_packages.clone());
        round2_packages.insert(identifier.clone(), packages);
    }

    // now each participant has to receive the round::package from their peers
    // We need to give every the their round 2 packages generated on by the other n-1 participants
    for (originating_participant, generated_packages) in round2_packages {
        for pkg in generated_packages {
            let recipient = participants.get_mut(&pkg.identifier).unwrap();

            recipient.receive_round2_package(originating_participant.clone(), pkg);
        }
    }

    let mut keys: HashMap<ParticipantIdentifier, (FrostKeyPackage, FrostPublicKeyPackage)> =
        HashMap::new();

    // do part 3.
    for (identifier, participant) in participants.iter_mut() {
        keys.insert(identifier.clone(), participant.do_part3());
    }

    assert_eq!(keys.len(), 3);

    // sign

    let mut key_packages: HashMap<ParticipantIdentifier, FrostKeyPackage> = HashMap::new();
    let mut pubkeys: HashMap<ParticipantIdentifier, FrostPublicKeyPackage> = HashMap::new();

    for k in keys.into_iter() {
        key_packages.insert(k.0.clone(), k.1 .0.clone());
        pubkeys.insert(k.0, k.1 .1.clone());
    }

    let mut rng = thread_rng();
    let (nonces, commitments) = round_1::<E>(&mut rng, &key_packages);
    let message = Message {
        data: "i am a message".as_bytes().to_vec(),
    };

    #[cfg(feature = "redpallas")]
    let (signing_package, signature_shares, randomizer) = round_2(
        &mut rng,
        &nonces,
        &key_packages,
        commitments,
        message.clone(),
        None,
    );

    #[cfg(not(feature = "redpallas"))]
    let (signing_package, signature_shares) =
        round_2(&nonces, &key_packages, commitments, message.clone());

    let p1identifier = p1_identifier.clone();
    let pubkey = pubkeys.get(&p1identifier).unwrap().clone();
    let group_signature = aggregate(
        signing_package,
        signature_shares.into_iter().map(|s| s.1).collect(),
        pubkey.clone(),
        #[cfg(feature = "redpallas")]
        randomizer.unwrap(),
    )
    .unwrap();

    let verify_signature = verify_signature(message, group_signature, pubkey.clone());

    match verify_signature {
        Ok(()) => assert!(true),
        Err(e) => {
            assert!(false, "signature verification failed with error: {e:?}")
        }
    }
}
