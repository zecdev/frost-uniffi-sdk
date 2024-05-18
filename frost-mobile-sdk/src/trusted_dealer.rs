#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

use std::collections::HashMap;
use crate::{Configuration, FrostPublicKeyPackage, FrostSecretKeyShare, ParticipantIdentifier};
use rand::thread_rng;


use frost::keys::{IdentifierList, PublicKeyPackage, SecretShare};
use frost::{Error, Identifier, SigningKey};
use rand::rngs::ThreadRng;
use std::collections::BTreeMap;

pub (crate) fn trusted_dealer_keygen_from_configuration(
    config: &Configuration
) -> Result<(FrostPublicKeyPackage, HashMap<ParticipantIdentifier, FrostSecretKeyShare>), Error> {
    let mut rng = thread_rng();

    let keygen = if config.secret.is_empty() {
        trusted_dealer_keygen(&config, IdentifierList::Default, &mut rng)
    } else {
        split_secret(&config, IdentifierList::Default, &mut rng)
    };

    let (shares, pubkeys) = keygen?;

    let pubkey = FrostPublicKeyPackage::from_public_key_package(pubkeys)?;

    let mut hash_map: HashMap<ParticipantIdentifier, FrostSecretKeyShare> = HashMap::new();

    for (k,v) in shares {
        hash_map.insert(
            ParticipantIdentifier::from_identifier(k)?,
            FrostSecretKeyShare::from_secret_share(v)?
        );
    }
    
    Ok((pubkey, hash_map))
} 

pub(crate) fn trusted_dealer_keygen(
    config: &Configuration,
    identifiers: IdentifierList,
    rng: &mut ThreadRng,
) -> Result<(BTreeMap<Identifier, SecretShare>, PublicKeyPackage), Error> {
    let (shares, pubkeys) = frost::keys::generate_with_dealer(
        config.max_signers,
        config.min_signers,
        identifiers,
        rng,
    )?;

    for (_k, v) in shares.clone() {
        frost::keys::KeyPackage::try_from(v)?;
    }

    Ok((shares, pubkeys))
}

fn split_secret(
    config: &Configuration,
    identifiers: IdentifierList,
    rng: &mut ThreadRng,
) -> Result<(BTreeMap<Identifier, SecretShare>, PublicKeyPackage), Error> {
    let secret_key = SigningKey::deserialize(
        config
            .secret
            .clone()
            .try_into()
            .map_err(|_| Error::MalformedSigningKey)?,
    )?;
    let (shares, pubkeys) = frost::keys::split(
        &secret_key,
        config.max_signers,
        config.min_signers,
        identifiers,
        rng,
    )?;

    for (_k, v) in shares.clone() {
        frost::keys::KeyPackage::try_from(v)?;
    }

    Ok((shares, pubkeys))
}


mod tests {
    #[cfg(not(feature = "redpallas"))]
    use frost_ed25519 as frost;
    #[cfg(feature = "redpallas")]
    use reddsa::frost::redpallas as frost;
    use reddsa::Randomizer;
    use uniffi::deps::static_assertions::assert_trait_sub_all;

    use crate::frost::keys::IdentifierList;
    use crate::helpers::{key_package, round_1, round_2};
    #[cfg(feature = "redpallas")]
    use crate::randomizer::FrostRandomizer;
    use rand::thread_rng;
    use crate::{coordinator::{aggregate, verify_signature, Message}, trusted_dealer::{split_secret, trusted_dealer_keygen_from_configuration}, Configuration};

    #[test]
    fn return_malformed_signing_key_error_if_secret_is_invalid() {
        let mut rng = thread_rng();
        let secret_config = Configuration {
            min_signers: 2,
            max_signers: 3,
            secret: b"helloIamaninvalidsecret111111111".to_vec(),
        };

        let out = split_secret(&secret_config, IdentifierList::Default, &mut rng);

        assert!(out.is_err());
    }

    #[test]
    fn return_malformed_signing_key_error_if_secret_is_invalid_type() {
        let mut rng = thread_rng();
        let secret: Vec<u8> = vec![
            123, 28, 51, 211, 245, 41, 29, 133, 222, 102, 72, 51, 190, 177, 173, 70, 159, 127, 182,
            2, 90, 14, 199, 139, 58, 121, 12, 110, 19, 169, 131,
        ];
        let secret_config = Configuration {
            min_signers: 2,
            max_signers: 3,
            secret,
        };

        let out = split_secret(&secret_config, IdentifierList::Default, &mut rng);

        assert!(out.is_err());
    }

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

    let (pubkeys, shares) =
        trusted_dealer_keygen_from_configuration(&secret_config).unwrap();
    let key_packages = key_package(&shares);
    let (nonces, commitments) = round_1(&mut rng, &key_packages);
    let message = Message {
        data: "i am a message".as_bytes().to_vec()
    };

    let (signing_package, signature_shares, randomizer) = round_2
    (
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
        signature_shares.into_iter()
            .map(|s| s.1)
            .collect(), 
        pubkeys.clone(),
        #[cfg(feature = "redpallas")]
        randomizer.unwrap()
    ).unwrap();

    let verify_signature = verify_signature(
        message, 
        group_signature, 
        pubkeys,
    );

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
            data: "i am a message".as_bytes().to_vec()
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
            signature_shares.into_iter()
                .map(|s| s.1)
                .collect(), 
            pubkeys.clone(),
            #[cfg(feature = "redpallas")]
            frost_randomizer,
        );

        match group_signature {
            Ok(s) => {
                let verify_signature = verify_signature(
                    message, 
                    s, 
                    pubkeys
                );
            
                assert!(verify_signature.is_ok());
            },
            Err(e) => assert!(false, "verify  signature failed with error {e}")
            
        }   
    }
}