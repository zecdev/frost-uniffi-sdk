//
//  ContentView.swift
//  FrostCompanion
//
//  Created by Pacu on 2024-06-03.
//

import SwiftUI
import ComposableArchitecture
struct MainScreenView: View {
    @Bindable var store: StoreOf<MainScreenFeature>
    var body: some View {
        NavigationStack(path: $store.scope(state: \.path, action:\.path)) {
            VStack {
                Image(systemName: "snow")
                    .imageScale(.large)
                    .foregroundStyle(.tint)
                
                Text("Who are you?")
                VStack {
                    NavigationLink(state: ParticipantImportFeature.State(keyShare: .empty)){
                        Text("Participant")
//                        Button("Participant") {
//                            store.send(.participantTapped)
//                        }
                    }
                    

                    Button("Coordinator") {
                        store.send(.coordinatorTapped)
                    }
                }
            }
            .padding()
        } destination: { store in
            ParticipantImportView(store: store)
        }

        .navigationTitle("Hello, FROST! ❄️")
    }

}

#Preview {
    MainScreenView(
        store: Store(initialState: MainScreenFeature.State()){
            MainScreenFeature()
        }
    )
}
