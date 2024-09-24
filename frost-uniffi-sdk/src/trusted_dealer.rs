use frost_core::{self as frost, Ciphersuite};

use crate::{Configuration, FrostPublicKeyPackage, FrostSecretKeyShare, ParticipantIdentifier};
use rand::thread_rng;
use std::collections::HashMap;

use frost::keys::{IdentifierList, PublicKeyPackage, SecretShare};
use frost::{Error, Identifier, SigningKey};
use rand::rngs::ThreadRng;
use std::collections::BTreeMap;

pub fn trusted_dealer_keygen_from_configuration<C: Ciphersuite>(
    config: &Configuration,
) -> Result<
    (
        FrostPublicKeyPackage,
        HashMap<ParticipantIdentifier, FrostSecretKeyShare>,
    ),
    frost_core::Error<C>,
> {
    let mut rng = thread_rng();

    let keygen = if config.secret.is_empty() {
        trusted_dealer_keygen(config, IdentifierList::Default, &mut rng)
    } else {
        split_secret(config, IdentifierList::Default, &mut rng)
    };

    let trusted_dealt_keys = keygen?;

    let pubkey =
        FrostPublicKeyPackage::from_public_key_package::<C>(trusted_dealt_keys.public_keys)?;

    let mut hash_map: HashMap<ParticipantIdentifier, FrostSecretKeyShare> = HashMap::new();

    for (k, v) in trusted_dealt_keys.secret_shares {
        hash_map.insert(
            ParticipantIdentifier::from_identifier(k)?,
            FrostSecretKeyShare::from_secret_share(v)?,
        );
    }

    Ok((pubkey, hash_map))
}

pub struct TrustDealtKeys<C: Ciphersuite> {
    pub secret_shares: BTreeMap<Identifier<C>, SecretShare<C>>,
    pub public_keys: PublicKeyPackage<C>,
}

pub fn trusted_dealer_keygen<C: Ciphersuite>(
    config: &Configuration,
    identifiers: IdentifierList<C>,
    rng: &mut ThreadRng,
) -> Result<TrustDealtKeys<C>, Error<C>> {
    let (shares, pubkeys) = frost::keys::generate_with_dealer(
        config.max_signers,
        config.min_signers,
        identifiers,
        rng,
    )?;

    for (_k, v) in shares.clone() {
        frost::keys::KeyPackage::try_from(v)?;
    }

    Ok(TrustDealtKeys {
        secret_shares: shares,
        public_keys: pubkeys,
    })
}

fn split_secret<C: Ciphersuite>(
    config: &Configuration,
    identifiers: IdentifierList<C>,
    rng: &mut ThreadRng,
) -> Result<TrustDealtKeys<C>, Error<C>> {
    let secret_key = SigningKey::deserialize(&config.secret)?;
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

    Ok(TrustDealtKeys {
        secret_shares: shares,
        public_keys: pubkeys,
    })
}
#[cfg(test)]
mod tests {

    #[cfg(not(feature = "redpallas"))]
    type E = frost_ed25519::Ed25519Sha512;

    #[cfg(feature = "redpallas")]
    type E = reddsa::frost::redpallas::PallasBlake2b512;

    use crate::{trusted_dealer::split_secret, Configuration};
    use frost_core::keys::IdentifierList;
    use rand::thread_rng;

    #[test]
    fn return_malformed_signing_key_error_if_secret_is_invalid() {
        let mut rng = thread_rng();
        let secret_config = Configuration {
            min_signers: 2,
            max_signers: 3,
            #[cfg(feature = "redpallas")]
            secret: b"invalidsecret".to_vec(),
            #[cfg(not(feature = "redpallas"))]
            secret: b"helloIamaninvalidsecret111111111".to_vec(),
        };

        let out = split_secret(&secret_config, IdentifierList::<E>::Default, &mut rng);

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

        let out = split_secret(&secret_config, IdentifierList::<E>::Default, &mut rng);

        assert!(out.is_err());
    }
}
