mod helpers;
use frost_mobile_sdk::{
    coordinator::{aggregate, verify_signature, Message},
    trusted_dealer::trusted_dealer_keygen_from_configuration,
    Configuration,
};
use helpers::{key_package, round_1, round_2};
use rand::thread_rng;

#[test]
fn test_trusted_from_configuration_with_secret() {
    let mut rng = thread_rng();
    let secret: Vec<u8> = vec![
        123, 28, 51, 211, 245, 41, 29, 133, 222, 102, 72, 51, 190, 177, 173, 70, 159, 127, 182, 2,
        90, 14, 199, 139, 58, 121, 12, 110, 19, 169, 131, 4,
    ];
    let secret_config = Configuration {
        min_signers: 2,
        max_signers: 3,
        secret,
    };

    let (pubkeys, shares) = trusted_dealer_keygen_from_configuration(&secret_config).unwrap();
    let key_packages = key_package(&shares);
    let (nonces, commitments) = round_1(&mut rng, &key_packages);
    let message = Message {
        data: "i am a message".as_bytes().to_vec(),
    };

    let (signing_package, signature_shares, randomizer) = round_2(
        &mut rng,
        &nonces,
        &key_packages,
        commitments,
        message.clone(),
        #[cfg(feature = "redpallas")]
        None,
    );

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
    let secret: Vec<u8> = vec![
        123, 28, 51, 211, 245, 41, 29, 133, 222, 102, 72, 51, 190, 177, 173, 70, 159, 127, 182, 2,
        90, 14, 199, 139, 58, 121, 12, 110, 19, 169, 131, 4,
    ];
    let secret_config = Configuration {
        min_signers: 14,
        max_signers: 20,
        secret,
    };
    let (pubkeys, shares) = trusted_dealer_keygen_from_configuration(&secret_config).unwrap();
    let key_packages = key_package(&shares);
    let (nonces, commitments) = round_1(&mut rng, &key_packages);
    let message = Message {
        data: "i am a message".as_bytes().to_vec(),
    };
    let (signing_package, signature_shares, randomizer) = round_2(
        &mut rng,
        &nonces,
        &key_packages,
        commitments,
        message.clone(),
        #[cfg(feature = "redpallas")]
        None,
    );

    #[cfg(feature = "redpallas")]
    let frost_randomizer = randomizer.unwrap();
    let group_signature = aggregate(
        signing_package,
        signature_shares.into_iter().map(|s| s.1).collect(),
        pubkeys.clone(),
        #[cfg(feature = "redpallas")]
        frost_randomizer,
    );

    match group_signature {
        Ok(s) => {
            let group = s.to_signature().unwrap();

            let is_signature_valid = pubkeys
                .into_public_key_package()
                .unwrap()
                .verifying_key()
                .verify("test".as_bytes(), &group)
                .is_ok();
            assert!(is_signature_valid);

            let verify_signature = verify_signature(message, s, pubkeys);

            assert!(verify_signature.is_ok());
        }
        Err(e) => assert!(false, "verify  signature failed with error {e}"),
    }
}
