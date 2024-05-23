# frost-uniffi-sdk
Proof-of-Concept of a FROST SDK using Uniffi

FROST is a thresold signature scheme created by Zcash Foundation. 
This Software Development Kit wraps the [RustLang implementation of
FROST](https://github.com/ZcashFoundation/frost) using [Mozilla UniFFI](https://mozilla.github.io/uniffi-rs/).

The creation of this SDK wouldn't have been possible without the reference of the
[Zcash FROST Demo](https://zfnd.org/demo-for-frost-for-zcash-library/)

## Before you start.

**This specific wrapper has not been audited**. 

FROST itself as a signature scheme and the FROST SDK have been audited.
The results of the audit can be found [here](https://research.nccgroup.com/2023/10/23/public-report-zcash-frost-security-assessment/).

## Where to start?

### Read First!
DON'T JUMP INTO THE CODE YET! Please read these resources carefully. They are 
essencial for understanding how to use FROST.

If you are not familiar with FROST threshold signatures, start with the 
basics: [The FROST Book](https://frost.zfnd.org/index.html). A quick, concise
and comprehensive guide to FROSTying your signatures by @ZcashFoundation.

FROST is a creation of Chelsea Komlo (ZF) and Ian Goldberg (University of Waterloo).
The first FROST paper can be found [here](https://eprint.iacr.org/2020/852.pdf)

Zcash uses [Re-Randomized FROST](https://eprint.iacr.org/2024/436) by Conrado Gouvea 
and Chelsea Komlo from @ZcashFoundation.

### Building

**Pre-requisites** 

- Install Rust and Cargo
- Install [Cargo Swift plugin](https://github.com/antoniusnaumann/cargo-swift)
- Install MacOS and iOS targets `aarch64-apple-ios`, `86_64-apple-darwin`, 
`aarch64-apple-darwin`

**Building the mobile SDK**
to build the mobile SDK use the following command on the `frost-uniffi-sdk`
folder.

```
cargo swift package --name frost-swift-sdk
```


## Features
This SDK contains all the moving parts to do 2-round FROST threshold signatures

Dealership
----------
- Trusted Dealer Key Generation
- Trusted Dealer Key Generation with existing identifiers

Round 1:
-------
- Participant Commitment generation
- Key Package generation
- Public Key Package generation
- Signing Commitments and Nonces generation

Round 2:
--------
- Coordinator Signing Package generation
- Participant `round2::sign()` signature
- Coordinator Signature aggregation
- Signature Verification

### Platforms supported
- Swift

To be Supported:
- GoLang
- Kotlin
- Python

## What is not included
- Secure transport to exchange key material
- Server communication
- Distributed Key Generation

# Contributing

Please open issues to request features. You can send PRs for the issues
you open as well 🙏😅

# Acknowledments

This work was possible thanks to the effort of the Zcash Community and 
the funding of Zcash Community Grants commitee. 

