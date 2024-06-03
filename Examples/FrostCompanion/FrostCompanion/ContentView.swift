//
//  ContentView.swift
//  FrostCompanion
//
//  Created by Pacu on 2024-06-03.
//

import SwiftUI

struct ContentView: View {
    var body: some View {
        VStack {
            Image(systemName: "snow")
                .imageScale(.large)
                .foregroundStyle(.tint)
            Text("Hello, FROST!")
        }
        .padding()
    }
}

#Preview {
    ContentView()
}
