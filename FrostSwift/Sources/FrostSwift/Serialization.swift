//
//  Serialization.swift
//
//
//  Created by Pacu in 2024.
//    
   

import Foundation
import FrostSwiftFFI

/// Allows Models that are meant to be shared with other actors of the FROST
/// signature scheme protocol to be serialized into JSON strings
public protocol JSONSerializable {
    func toJSONString() throws -> String
}

/// Allows Models that are meant to be shared with other actors of the FROST
/// signature scheme protocol to be deserialized from JSON strings
public protocol JSONDeserializable {
    static func fromJSONString(_ jsonString: String) throws -> Self
}

/// Allows Models that are meant to be shared with other actors of the FROST
/// signature scheme protocol and are tied to their ``Identifier``  to be deserialized
/// from JSON strings associating them to the corresponding Identifier
public protocol JSONIdentifiable {
    static func fromJSONString(_ jsonString: String, identifier: Identifier) throws -> Self
}

extension KeyPackage: JSONSerializable, JSONDeserializable {
    public func toJSONString() throws -> String {
        try keyPackageToJson(keyPackage: self.package)
    }
    
    public static func fromJSONString(_ jsonString: String) throws -> KeyPackage {
        let keyPackage = try jsonToKeyPackage(keyPackageJson: jsonString)
        return KeyPackage(package: keyPackage)
    }
}

extension PublicKeyPackage: JSONSerializable, JSONDeserializable {
    public func toJSONString() throws -> String {
        try publicKeyPackageToJson(publicKeyPackage: self.package)
    }
    
    public static func fromJSONString(_ jsonString: String) throws -> PublicKeyPackage {
        let publicKeyPackage = try jsonToPublicKeyPackage(publicKeyPackageJson: jsonString)

        return PublicKeyPackage(package: publicKeyPackage)
    }
}

extension SignatureShare: JSONSerializable, JSONIdentifiable  {
    public func toJSONString() throws -> String {
        try signatureSharePackageToJson(signatureShare: self.share)
    }
    
    public static func fromJSONString(_ jsonString: String, identifier: Identifier) throws -> SignatureShare {
        let signatureShare = try jsonToSignatureShare(
            signatureShareJson: jsonString,
            identifier: identifier.toParticipantIdentifier()
        )
        return SignatureShare(share: signatureShare)
    }
}

extension SigningCommitments: JSONSerializable, JSONIdentifiable {
    public func toJSONString() throws -> String {
        try commitmentToJson(commitment: self.commitment)
    }

    public static func fromJSONString(_ jsonString: String, identifier: Identifier) throws -> SigningCommitments {
        let commitments = try jsonToCommitment(commitmentJson: jsonString, identifier: identifier.toParticipantIdentifier())

        return SigningCommitments(commitment: commitments)
    }
}

extension Randomizer: JSONSerializable, JSONDeserializable {
    public func toJSONString() throws -> String {
        try randomizerToJson(randomizer: self.randomizer)
    }
    
    public static func fromJSONString(_ jsonString: String) throws -> Randomizer {
        let randomizer = try jsonToRandomizer(randomizerJson: jsonString)

        return Randomizer(randomizer: randomizer)
    }
}
