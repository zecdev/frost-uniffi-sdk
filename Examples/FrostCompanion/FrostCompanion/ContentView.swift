//
//  ContentView.swift
//  FrostCompanion
//
//  Created by Pacu on 2024-06-03.
//

import SwiftUI
import ComposableArchitecture
struct ContentView: View {
    let store: StoreOf<MainScreenFeature>
    var body: some View {
        NavigationStack {
            VStack {
                Image(systemName: "snow")
                    .imageScale(.large)
                    .foregroundStyle(.tint)
                
                Text("Who are you?")
                VStack {
                    Button("Participant") {
                        store.send(.participantTapped)
                    }
                    
                    Button("Coordinator") {
                        store.send(.coordinatorTapped)
                    }
                }
            }
            .padding()
        }
        .navigationTitle("Hello, FROST! ❄️")
    }
    
}

#Preview {
    ContentView(
        store: Store(initialState: MainScreenFeature.State()){
            MainScreenFeature()
        }
    )
}
