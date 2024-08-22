//
//  MainScreenReducer.swift
//  FrostCompanion
//
//  Created by pacu on 21/08/2024.
//

import Foundation
import ComposableArchitecture


@Reducer
struct MainScreenFeature {
    @ObservableState
    struct State: Equatable {
        @Presents var participant: ParticipantReducer.State?
    }
    
    enum Action {
        case coordinatorTapped
        case participantTapped
        case importKeyShare(PresentationAction<ParticipantReducer.Action>)
    }
    
    var body: some ReducerOf<Self> {
        Reduce { state, action in
            switch action {
            case .coordinatorTapped:
                return .none
                
            case .participantTapped:
                state.participant = ParticipantReducer.State(
                    keyShare: JSONKeyShare.empty
                )
                return .none
        
            case .importKeyShare(.presented(.delegate(.keyShareImported(let keyShare)))):
                debugPrint(keyShare)
                return .none

            case .importKeyShare:
                return .none
       
            }
        }
        .ifLet(\.$participant, action: \.importKeyShare) {
            ParticipantReducer()
        }
    }
}
