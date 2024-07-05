//
//  SerializationTests.swift
//  
//
//  Created by Pacu in 2024.
//    
   

import XCTest
import FrostSwift
/// JSON Serialization and Deserialization integration tests
/// - Important: These tests failing indicate that something is has changed on the
/// FROST crate.
final class SerializationTests: XCTestCase {
    /// ```
    /// let json_package = r#"
    ///    {
    ///        "header": {
    ///            "version": 0,
    ///            "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
    ///        },
    ///        "identifier": "0300000000000000000000000000000000000000000000000000000000000000",
    ///        "signing_share": "112911acbfec15db78da8c4b1027b3ac75ce342447111226e1f15c93ca062d12",
    ///        "verifying_share": "60a1623d2a419d6a007177d13b75458b148e8ef1ba74e0fbc1d837d07b5f9706",
    ///        "verifying_key": "1fa942a303acbc3185dce72b2909ba838bb1efb16d500986f4afb7e04d43de85",
    ///        "min_signers": 2
    ///    }
    ///"#;
    /// ```
    func testKeyPackageRoundTripRoundTrip() throws {
        let jsonKeyPackage =
        """
        {"header":{"version":0,"ciphersuite":"FROST(Pallas, BLAKE2b-512)"},"identifier":"0300000000000000000000000000000000000000000000000000000000000000","signing_share":"112911acbfec15db78da8c4b1027b3ac75ce342447111226e1f15c93ca062d12","verifying_share":"60a1623d2a419d6a007177d13b75458b148e8ef1ba74e0fbc1d837d07b5f9706","verifying_key":"1fa942a303acbc3185dce72b2909ba838bb1efb16d500986f4afb7e04d43de85","min_signers":2}
        """

        let keyPackage = try KeyPackage.fromJSONString(jsonKeyPackage)

        let keyPackageJSON = try keyPackage.toJSONString()

        XCTAssertEqual(jsonKeyPackage, keyPackageJSON)
    }

    /// ```
    /// let public_key_package = r#"
    ///    {
    ///        "header": {
    ///            "version": 0,
    ///            "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
    ///        },
    ///        "verifying_shares": {
    ///            "0100000000000000000000000000000000000000000000000000000000000000": "61a199916a3c2b64c5e566deb1ab18997282f9559f5b328f6ae50ca24b349f9d",
    ///            "0200000000000000000000000000000000000000000000000000000000000000": "389656dbe50a0b260c5b4e7ee953e8d81b0814cbdc112a6cd773d55de4202c0e",
    ///            "0300000000000000000000000000000000000000000000000000000000000000": "c0d94a637e113a82942bd0b886fa7d0e2256010bd42a9893c81df1a58e34ff8d"
    ///        },
    ///        "verifying_key": "93c3d1dca3634e26c7068342175b7dd5b3e3f3654494f6f6a3b77f96f3cb0a39"
    ///    }
    ///   "#;
    /// ```
    func testPublicKeyPackageRoundTrip() throws {
        let jsonPublicKeyPackage =
        """
        {"header":{"version":0,"ciphersuite":"FROST(Pallas, BLAKE2b-512)"},"verifying_shares":{"0100000000000000000000000000000000000000000000000000000000000000":"61a199916a3c2b64c5e566deb1ab18997282f9559f5b328f6ae50ca24b349f9d","0200000000000000000000000000000000000000000000000000000000000000":"389656dbe50a0b260c5b4e7ee953e8d81b0814cbdc112a6cd773d55de4202c0e","0300000000000000000000000000000000000000000000000000000000000000":"c0d94a637e113a82942bd0b886fa7d0e2256010bd42a9893c81df1a58e34ff8d"},"verifying_key":"93c3d1dca3634e26c7068342175b7dd5b3e3f3654494f6f6a3b77f96f3cb0a39"}
        """

        let publicKeyPackage = try PublicKeyPackage.fromJSONString(jsonPublicKeyPackage)
        let publicKeyPackageJSON = try publicKeyPackage.toJSONString()

        XCTAssertEqual(jsonPublicKeyPackage, publicKeyPackageJSON)
    }

    func testRandomizerRoundTrip() throws {
        let jsonRandomizer =
        """
        "6fe2e6f26bca5f3a4bc1cd811327cdfc6a4581dc3fe1c101b0c5115a21697510"
        """

        let randomizer = try Randomizer.fromJSONString(jsonRandomizer)

        let randomizerJSON = try randomizer.toJSONString()

        XCTAssertEqual(jsonRandomizer, randomizerJSON)
    }

    ///
    /// ```
    /// let share_json = r#"
    /// {
    ///   "header": {
    ///     "version": 0,
    ///     "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
    ///   },
    ///   "share": "d202ad8525dd0b238bdc969141ebe9b33402b71694fb6caffa78439634ee320d"
    /// }"#;"
    /// ```
    func testSignatureShareRoundTrip() throws {
        let jsonSignatureShare =
        """
        {"header":{"version":0,"ciphersuite":"FROST(Pallas, BLAKE2b-512)"},"share":"307ebf4d5b7125407f359fa010cdca940a83e942fd389ecd67c6683ecee78f3e"}
        """

        let identifier = Identifier(with: 1)!

        let signatureShare = try SignatureShare.fromJSONString(jsonSignatureShare, identifier: identifier)

        let signatureShareJSON = try signatureShare.toJSONString()

        XCTAssertEqual(jsonSignatureShare, signatureShareJSON)
    }

    /// ```
    /// let commitment_json = r#"
    ///    {
    ///        "header": {
    ///            "version": 0,
    ///            "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
    ///        },
    ///        "hiding": "ad737cac6f8e9ae3ae21a0de51556c8ea86c8e483b2418cf58300c036ebc100c",
    ///        "binding": "d36a016645420728b278f33fcaa45781840b4960625e3c7cf189cebb76f9a08c"
    ///    }
    ///    "#;
    /// ```
    func testCommitmentsRoundTrip() throws {
        let jsonCommitment =
        """
        {"header":{"version":0,"ciphersuite":"FROST(Pallas, BLAKE2b-512)"},"hiding":"ad737cac6f8e9ae3ae21a0de51556c8ea86c8e483b2418cf58300c036ebc100c","binding":"d36a016645420728b278f33fcaa45781840b4960625e3c7cf189cebb76f9a08c"}
        """

        let identifier = Identifier(with: 1)!

        let commitment = try SigningCommitments.fromJSONString(jsonCommitment, identifier: identifier)

        let commitmentJSON = try commitment.toJSONString()

    }

}
