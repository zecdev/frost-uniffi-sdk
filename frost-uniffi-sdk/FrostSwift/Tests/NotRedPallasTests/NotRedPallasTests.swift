//
//  NotRedPallasTests.swift
//
//
//  Created by Francisco Gindre on 5/23/24.
//

import Foundation
import XCTest
@testable import FrostSwift

class NotRedPallasTests: XCTestCase {
    func testTrustedDealerFromConfigWithSecret() throws {
        let secret: [UInt8] = [
            123, 28, 51, 211, 245, 41, 29, 133, 222, 102, 72, 51, 190, 177, 173, 70, 159, 127, 182, 2,
            90, 14, 199, 139, 58, 121, 12, 110, 19, 169, 131, 4,
        ]

        let secretConfig = Configuration(minSigners: 2, maxSigners: 3, secret: Data(secret))

        let keygen = trustedDealerKeygenFrom(configuration: secret)

        let publicKey = keygen.publicKeyPackage
        let shares = keygen.secretShares

        XCTAssertEqual(keygen.secretShares.count, 3)
        XCTAssertEqual(publicKey.verifyingShares.count,3)
    }
}
