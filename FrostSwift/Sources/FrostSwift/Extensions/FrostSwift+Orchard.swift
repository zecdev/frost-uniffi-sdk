//
//  FrostSwift+Orchard.swift
//
//
//  Created by pacu on 2024-08-17.
//

import FrostSwiftFFI

extension OrchardFullViewingKey: Equatable {
    /// This method checks the equality of the UFVK by verifying that the encoded String is the same
    /// this means that it will return true if the UFVK has the same viewing keys
    /// - Note: Not really efficient.
    /// - Returns: true if the encodings are the same. False if not or if any of the two throw.
    public static func == (lhs: FrostSwiftFFI.OrchardFullViewingKey, rhs: FrostSwiftFFI.OrchardFullViewingKey) -> Bool {
        guard let lhs = try? lhs.encode(), let rhs = try? rhs.encode() else {
            return false
        }
        
        return lhs == rhs
    }
}
