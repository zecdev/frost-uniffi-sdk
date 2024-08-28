//
//  TrustedDealerFeature.swift
//  FrostCompanion
//
//  Created by Pacu in  2024.
//    
   

import Foundation
import ComposableArchitecture

@Reducer
struct TrustedDealerFeature {
    @ObservableState
    struct State {
        var maxParticipants: UInt = 3
        var minParticipants: UInt = 2
        var path: StackState<NewTrustedDealerSchemeFeature.State>
    }

    enum Action {
        case createScheme
        case setMaxParticipants(Int)
        case setMinParticipants(Int)
    }
}
