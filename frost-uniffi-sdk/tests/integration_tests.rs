mod helpers;

use frost_uniffi_sdk::{
    coordinator::{verify_signature, Message},
    trusted_dealer::trusted_dealer_keygen_from_configuration,
    Configuration,
};
use helpers::{key_package, round_1};
use rand::thread_rng;

#[cfg(not(feature = "redpallas"))]
use frost_uniffi_sdk::coordinator::aggregate;
#[cfg(feature = "redpallas")]
use frost_uniffi_sdk::{randomized::coordinator::aggregate, randomized::tests::helpers::round_2};
#[cfg(not(feature = "redpallas"))]
use helpers::round_2;


#[cfg(feature = "redpallas")]
type E = reddsa::frost::redpallas::PallasBlake2b512;
#[cfg(not(feature = "redpallas"))]
type E = frost_ed25519::Ed25519Sha512;

fn test_signing_key() -> Vec<u8> {
    #[cfg(feature = "redpallas")]
    return hex::decode("f500df73b2b416bec6a2b6bbb44e97164e05520b63aa27554cfc7ba82f5ba215")
        .unwrap();

    #[cfg(not(feature = "redpallas"))]
    return vec![
        123, 28, 51, 211, 245, 41, 29, 133, 222, 102, 72, 51, 190, 177, 173, 70, 159, 127, 182, 2,
        90, 14, 199, 139, 58, 121, 12, 110, 19, 169, 131, 4,
    ];
}
#[test]
fn test_trusted_from_configuration_with_secret() {
    let mut rng = thread_rng();

    let secret: Vec<u8> = test_signing_key();
    let secret_config = Configuration {
        min_signers: 2,
        max_signers: 3,
        secret,
    };

    let (pubkeys, shares) = trusted_dealer_keygen_from_configuration::<E>(&secret_config).unwrap();
    let key_packages = key_package::<E>(&shares);
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

    let group_signature = aggregate(
        signing_package,
        signature_shares.into_iter().map(|s| s.1).collect(),
        pubkeys.clone(),
        #[cfg(feature = "redpallas")]
        randomizer.unwrap(),
    )
    .unwrap();

    let verify_signature = verify_signature(message, group_signature, pubkeys);

    match verify_signature {
        Ok(()) => assert!(true),
        Err(e) => {
            assert!(false, "signature verification failed with error: {e:?}")
        }
    }
}

#[test]
fn check_keygen_with_dealer_with_secret_with_large_num_of_signers() {
    let mut rng = thread_rng();
    let secret: Vec<u8> = test_signing_key();
    let secret_config = Configuration {
        min_signers: 14,
        max_signers: 20,
        secret,
    };

    let (pubkeys, shares) = trusted_dealer_keygen_from_configuration::<E>(&secret_config).unwrap();

    let key_packages = key_package::<E>(&shares);
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

    #[cfg(feature = "redpallas")]
    let frost_randomizer = randomizer.unwrap();

    let group_signature = aggregate(
        signing_package,
        signature_shares.into_iter().map(|s| s.1).collect(),
        pubkeys.clone(),
        #[cfg(feature = "redpallas")]
        frost_randomizer,
    )
    .unwrap();

    assert!(verify_signature(message, group_signature, pubkeys).is_ok())
}
