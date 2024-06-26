//
//  FFIConversion.swift
//
//
//  Created by Pacu on 26-06-2024.
//

import Foundation
import FrostSwiftFFI

extension Configuration {
    func intoFFIConfiguration() -> FrostSwiftFFI.Configuration {
        FrostSwiftFFI.Configuration(
            minSigners: minSigners,
            maxSigners: maxSigners,
            secret: secret ?? Data()
        )
    }
}

extension ParticipantIdentifier {
    func toIdentifier() -> Identifier {
        Identifier(participant: self)
    }
}

extension Identifier {
    func toParticipantIdentifier() -> ParticipantIdentifier {
        id
    }
}

extension TrustedKeyGeneration {
    func toKeyGeneration() -> TrustedDealerCoordinator.KeyGeneration {
        var keys = [Identifier: SecretShare]()

        for secretShare in secretShares {
            keys[secretShare.key.toIdentifier()] = SecretShare(share: secretShare.value)
        }

        return TrustedDealerCoordinator.KeyGeneration(
            publicKeyPackage: PublicKeyPackage(package: publicKeyPackage),
            secretShares: keys
        )
    }
}
