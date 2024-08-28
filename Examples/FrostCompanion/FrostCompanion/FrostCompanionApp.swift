//
//  FrostCompanionApp.swift
//  FrostCompanion
//
//  Created by Pacu on 2024-06-03.
//

import SwiftUI
import ComposableArchitecture
@main
struct FrostCompanionApp: App {
    var body: some Scene {
        WindowGroup {
            MainScreenView(
                store: Store(initialState: MainScreenFeature.State()){
                    MainScreenFeature()
                }
            )
        }
    }
}
