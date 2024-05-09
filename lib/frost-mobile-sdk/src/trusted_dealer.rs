#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;

use std::collections::HashMap;
use crate::{Configuration, FrostPublicKeyPackage, FrostSecretKeyShare};
use rand::thread_rng;


use frost::keys::{IdentifierList, PublicKeyPackage, SecretShare};
use frost::{Error, Identifier, SigningKey};
use rand::rngs::ThreadRng;
use std::collections::BTreeMap;

fn trusted_dealer_keygen_from_configuration(
    config: Configuration
) -> Result<(FrostPublicKeyPackage, HashMap<String, FrostSecretKeyShare>), Error> {
    let mut rng = thread_rng();

    let keygen = if config.secret.is_empty() {
        trusted_dealer_keygen(&config, IdentifierList::Default, &mut rng)
    } else {
        split_secret(&config, IdentifierList::Default, &mut rng)
    };

    let (shares, pubkeys) = keygen?;

    let pubkey = FrostPublicKeyPackage::from_public_key_package(pubkeys)?;

    let mut hash_map: HashMap<String, FrostSecretKeyShare> = HashMap::new();

    for (k,v) in shares {
        hash_map.insert(
            hex::encode(k.serialize()),
            FrostSecretKeyShare::from_secret_share(v)?
        );
    }
    
    Ok((pubkey, hash_map))
} 

fn trusted_dealer_keygen(
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