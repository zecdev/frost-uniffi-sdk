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
        @Presents var destination: Destination.State?
        var path = StackState<ParticipantImportFeature.State>()
    }
    
    enum Action {
        case coordinatorTapped
        case participantTapped
        case destination(PresentationAction<Destination.Action>)
        case path(StackAction<ParticipantImportFeature.State, ParticipantImportFeature.Action>)
    }
    
    var body: some ReducerOf<Self> {
        Reduce { state, action in
            switch action {
            case .coordinatorTapped:
                return .none
                
            case .participantTapped:
                state.destination = .participant(ParticipantImportFeature.State(
                    keyShare: JSONKeyShare.empty
                ))
                return .none
        
            case .destination(.presented(.participant(.delegate(.keyShareImported(let keyShare))))):
                debugPrint(keyShare)
                return .none

            case .destination:
                return .none
            case .path:
                return .none

            }
        }
//        .ifLet(\.$destination, action: \.destination) {
//            Destination.participant(ParticipantReducer())
//        }
        .forEach(\.path, action: \.path)  {
            ParticipantImportFeature()
        }
    }
}


extension MainScreenFeature {

    @Reducer(state: .equatable)
    enum Destination {
        case participant(ParticipantImportFeature)
    }
}
