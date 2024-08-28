//
//  ParticipantImportView.swift
//  FrostCompanion
//
//  Created by pacu on 2024-08-21
//

import SwiftUI
import ComposableArchitecture

struct ParticipantImportView: View {
    @Bindable var store: StoreOf<ParticipantImportFeature>
    var body: some View {
        Form {
            Text("Paste your key-package.json contents")
            TextEditor(text: $store.keyShare.raw.sending(
                \.setKeyShare
            ))
            

            Button(
                "Import"
            ) {
                store.send(
                    .importButtonTapped
                )
            }
        }
        .toolbar {
            ToolbarItem {
                Button(
                    "Cancel"
                )  {
                    store.send(
                        .cancelButtonTapped
                    )
                }
            }
        }
        .navigationTitle(
            "Import your Key-package JSON"
        )
    }
}

#Preview {
    ParticipantImportView(store: Store(
        initialState: ParticipantImportFeature.State(
            keyShare: JSONKeyShare(
                raw: 
                        """
                        {
                            "header": {
                                "version": 0,
                                "ciphersuite": "FROST(Pallas, BLAKE2b-512)"
                            },
                            "identifier": "0100000000000000000000000000000000000000000000000000000000000000",
                            "signing_share": "b02e5a1e43a7f305177682574ac63c1a5f7f57db644c992635d09f699e56f41e",
                            "commitment": [
                                "4141ac3d66ff87c4d14eb14f4262b69de15f7093dfd1f411a02ea70644f0d41f",
                                "2eb4cd3ace283ba6bb9058ff08d0561ff6d87057ecc87b0701123979291fb082"
                            ]
                        }
                        """
            )
        )
    ) {
        ParticipantImportFeature()
    }
    )
    .padding()
}
