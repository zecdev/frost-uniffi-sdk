use std::collections::HashMap;

use reddsa::frost::redpallas as frost;
use uniffi;
use frost::keys::PublicKeyPackage;
uniffi::setup_scaffolding!();

#[derive(uniffi::Enum)]
pub enum Ciphersuite {
    RedPallas
}
#[derive(uniffi::Record)]
pub struct Header {
    pub version: u8, 
    pub suite: Ciphersuite, 
}

#[derive(uniffi::Record)]
pub struct FrostPublicKeyPackage {
    pub header: Header,
    pub verifying_shares: HashMap<String, String>,
    pub verifying_key: String
}

impl FrostPublicKeyPackage {
   
}

#[derive(uniffi::Object)]
pub struct Test {
    pub value: i64
}

#[uniffi::export]
impl Test { 
    fn do_something(&self) {
        print!("hello");
    }
}