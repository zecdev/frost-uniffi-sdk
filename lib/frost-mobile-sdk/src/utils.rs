#[cfg(not(feature = "redpallas"))]
use frost_ed25519 as frost;
#[cfg(feature = "redpallas")]
use reddsa::frost::redpallas as frost;
use serde_json;
use frost::serde::Serialize;

pub (crate) fn json_bytes<T>(structure: T) -> Vec<u8> where T: Serialize {
    let mut bytes: Vec<u8> = Vec::new();
    serde_json::to_writer(&mut bytes, &structure).unwrap();
    bytes
}