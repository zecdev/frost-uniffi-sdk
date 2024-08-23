//
//  OrchardSwiftTests.swift
//  
//
//  Created by Pacu in  2024.
//    
   

import XCTest
import Foundation
import CryptoSwift
@testable import FrostSwiftFFI
import FrostSwift
final class OrchardSwiftTests: XCTestCase {
    /// This test verifies that the APIs succeed at creating a  Full viewing key from a validating key
    /// and a ZIP 32 seed. this should be done other than for viewing key creation that is tossed.
    func testUFVKandAddressAreDerivedfromSeed() throws {
        
        let hexStringAk = "d2bf40ca860fb97e9d6d15d7d25e4f17d2e8ba5dd7069188cbf30b023910a71b"
        let hexAk = [UInt8](hex: hexStringAk)
        let ak = try OrchardSpendValidatingKey.fromBytes(bytes: Data(hexAk))

        let randomSeedBytesHexString = "659ce2e5362b515f30c38807942a10c18a3a2f7584e7135b3523d5e72bb796cc64c366a8a6bfb54a5b32c41720bdb135758c1afacac3e72fd5974be0846bf7a5"

        let randomSeedbytes = [UInt8](hex: randomSeedBytesHexString)
        let zcashNetwork = ZcashNetwork.testnet

        let fvk = try OrchardFullViewingKey.newFromValidatingKeyAndSeed(
            validatingKey: ak,
            zip32Seed: Data(randomSeedbytes),
            network: zcashNetwork
        )

        XCTAssertEqual(
            "uviewtest1jd7ucm0fdh9s0gqk9cse9xtqcyycj2k06krm3l9r6snakdzqz5tdp3ua4nerj8uttfepzjxrhp9a4c3wl7h508fmjwqgmqgvslcgvc8htqzm8gg5h9sygqt76un40xvzyyk7fvlestphmmz9emyqhjkl60u4dx25t86lhs30jreghq40cfnw9nqh858z4",
            try fvk.encode()
        )
        
        let address = try fvk.deriveAddress()

        XCTAssertEqual(
            "utest1fqasmz9zpaq3qlg4ghy6r5cf6u3qsvdrty9q6e4jh4sxd2ztryy0nvp59jpu5npaqwrgf7sgqu9z7hz9sdxw22vdpay4v4mm8vv2hlg4",
            address.stringEncoded()
        )
    }
    
    /// This tests that an Orchard UFVK can actually be decomposed into its parts.
    func testUFVKIsDecomposedOnParts() throws {
        let ufvkString = "uviewtest1jd7ucm0fdh9s0gqk9cse9xtqcyycj2k06krm3l9r6snakdzqz5tdp3ua4nerj8uttfepzjxrhp9a4c3wl7h508fmjwqgmqgvslcgvc8htqzm8gg5h9sygqt76un40xvzyyk7fvlestphmmz9emyqhjkl60u4dx25t86lhs30jreghq40cfnw9nqh858z4"
        
        let ufvk = try OrchardFullViewingKey.decode(stringEnconded: ufvkString, network: .testnet)
        
        let nk = ufvk.nk()
        let ak = ufvk.ak()
        let rivk = ufvk.rivk()
        
        let roundtripUFVK = try OrchardFullViewingKey.newFromCheckedParts(
            ak: ak,
            nk: nk,
            rivk: rivk,
            network: .testnet
        )
        
        XCTAssertEqual(ufvk, roundtripUFVK)
    }
}
