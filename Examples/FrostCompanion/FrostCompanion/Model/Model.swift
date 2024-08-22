//
//  Model.swift
//  FrostCompanion
//
//  Created by pacu on 2024-08-21.
//

import Foundation


struct JSONKeyShare: Equatable {
    let raw: String
}

extension JSONKeyShare {
    static let empty = JSONKeyShare(raw: "")
}
