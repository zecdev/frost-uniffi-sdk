//
//  Coordinator.swift
//
//
//  Created by Pacu on 27-06-2024.
//

import Foundation
import FrostSwiftFFI

enum FROSTCoordinatorError: Error {
    case repeatedCommitmentFromIdentifier(Identifier)
    case repeatedSignatureShareFromIdentifier(Identifier)
    case incorrectNumberOfCommitments(min: UInt16, max: UInt16, found: UInt16)
    case signingPackageAlreadyCreated
    case signingPackageMissing
    // mapped from FFI
    case failedToCreateSigningPackage
    case invalidSigningCommitment
    case identifierDeserializationError
    case signingPackageSerializationError
    case signatureShareDeserializationError
    case publicKeyPackageDeserializationError
    case signatureShareAggregationFailed(message: String)
    case invalidRandomizer

    /// some error we don't have mapped
    case otherError(CoordinationError)

    init(coordinationError: CoordinationError) {
        switch coordinationError {
        case .FailedToCreateSigningPackage:
            self = .failedToCreateSigningPackage
        case .InvalidSigningCommitment:
            self = .invalidSigningCommitment
        case .IdentifierDeserializationError:
            self = .identifierDeserializationError
        case .SigningPackageSerializationError:
            self = .signingPackageSerializationError
        case .SignatureShareDeserializationError:
            self = .signatureShareDeserializationError
        case .PublicKeyPackageDeserializationError:
            self = .publicKeyPackageDeserializationError
        case let .SignatureShareAggregationFailed(message: message):
            self = .signatureShareAggregationFailed(message: message)
        case .InvalidRandomizer:
            self = .invalidRandomizer
        }
    }
}

/// Protocol that defines the minimum functionality of a FROST signature
/// scheme Coordinator. A coordinator can be part of the signature or not.  For your
/// convenience, ``NonSigningCoordinator`` and ``SigningCoordinator``
/// implementations are provided.
/// - Important: Fallback because of misbehaving participants are not considered
/// within this protocol. Implementors must define their own strategies to deal with such
/// scenarios.
/// - Note: This protocol is bound to actors because it is assumed that several
/// participants may try to interact with it concurrently.
public protocol FROSTCoordinator: Actor {
    /// configuration of the FROST threshold scheme
    var configuration: Configuration { get }
    /// public key tha represent this signature scheme
    var publicKeyPackage: PublicKeyPackage { get }
    /// message to be signed by the _t_ participants
    var message: Message { get }
    /// this is the configuration of Round 2.
    var round2Config: Round2Configuration? { get }
    /// receives a signing commitment from a given participant
    /// - parameter commitment: the SigningCommitment from signature participant
    /// - throws: should throw ``FROSTCoordinatorError.alreadyReceivedCommitmentFromIdentifier`` when
    /// receiving a commitment from an Identifier that was received already independently from whether
    /// it is the same or a different one.
    func receive(commitment: SigningCommitments) async throws
    /// Creates a ``Round2Configuration`` struct with a ``SigningPackage`` and
    /// ``Randomizer`` for the case that Re-Randomized FROST is used.
    /// - throws: ``FROSTCoordinatorError.signingPackageAlreadyCreated`` if this function was
    /// called already or ``FROSTCoordinatorError.incorrectNumberOfCommitments``
    /// when the number of commitments gathered is less or more than the specified by the ``Configuration``
    func createSigningPackage() async throws -> Round2Configuration
    /// Receives a ``SignatureShare`` from a ``SigningParticipant``.
    /// - throws: ``FROSTCoordinatorError.alreadyReceivedSignatureShareFromIdentifier``
    /// when the same participant sends the same share repeatedly
    func receive(signatureShare: SignatureShare) async throws
    /// Aggregates all shares and creates the FROST ``Signature``
    /// - throws: several things can go wrong here. 🥶
    func aggregate() async throws -> Signature
    /// Verify a ``Signature`` with the ``PublicKeyPackage`` of this coordinator.
    func verify(signature: Signature) async throws
}

/// A signature scheme coordinator that does not participate in the signature scheme
public actor NonSigningCoordinator: FROSTCoordinator {
    public let configuration: Configuration
    public let publicKeyPackage: PublicKeyPackage
    public let message: Message
    public var round2Config: Round2Configuration?
    var commitments: [Identifier: SigningCommitments] = [:]
    var signatureShares: [Identifier: SignatureShare] = [:]

    public init(configuration: Configuration, publicKeyPackage: PublicKeyPackage, message: Message) throws {
        self.configuration = configuration
        self.publicKeyPackage = publicKeyPackage
        self.message = message
    }

    public func receive(commitment: SigningCommitments) throws {
        // TODO: validate that the commitment belongs to a known identifier
        guard commitments[commitment.identifier] == nil else {
            throw FROSTCoordinatorError.repeatedCommitmentFromIdentifier(commitment.identifier)
        }

        commitments[commitment.identifier] = commitment
    }

    public func createSigningPackage() throws -> Round2Configuration {
        guard round2Config?.signingPackage == nil else {
            throw FROSTCoordinatorError.signingPackageAlreadyCreated
        }

        try validateNumberOfCommitments()

        let package = try SigningPackage(
            package: newSigningPackage(
                message: message,
                commitments: commitments.values.map { $0.commitment }
            )
        )

        let randomizedParams = try RandomizedParams(
            publicKey: publicKeyPackage,
            signingPackage: package
        )

        let randomizer = try randomizedParams.randomizer()

        let config = Round2Configuration(signingPackage: package, randomizer: randomizer)
        round2Config = config

        return config
    }

    /// receives the signature share from a partipicant
    public func receive(signatureShare: SignatureShare) throws {
        // TODO: validate that the commitment belongs to a known identifier
        guard signatureShares[signatureShare.identifier] == nil else {
            throw FROSTCoordinatorError.repeatedSignatureShareFromIdentifier(signatureShare.identifier)
        }

        signatureShares[signatureShare.identifier] = signatureShare
    }

    public func aggregate() throws -> Signature {
        try validateNumberOfCommitments()
        let round2config = try round2ConfigPresent()

        guard let randomizer = round2config.randomizer?.randomizer else {
            throw FROSTCoordinatorError.invalidRandomizer
        }

        let signature = try FrostSwiftFFI.aggregate(
            signingPackage: round2config.signingPackage.package,
            signatureShares: signatureShares.values.map { $0.share },
            pubkeyPackage: publicKeyPackage.package,
            randomizer: randomizer
        )

        return Signature(signature: signature)
    }

    public func verify(signature _: Signature) throws {
        throw FrostError.invalidSignature
    }

    func round2ConfigPresent() throws -> Round2Configuration {
        guard let config = round2Config else {
            throw FROSTCoordinatorError.signingPackageMissing
        }

        return config
    }

    func validateNumberOfCommitments() throws {
        guard commitments.count >= configuration.minSigners &&
            commitments.count <= configuration.maxSigners
        else {
            throw FROSTCoordinatorError.incorrectNumberOfCommitments(
                min: configuration.minSigners,
                max: configuration.maxSigners,
                found: UInt16(commitments.count)
            )
        }
    }
}

/// A Coordinator that also participates in the signature production.
public actor SigningCoordinator: FROSTCoordinator {
    public let configuration: Configuration
    public let publicKeyPackage: PublicKeyPackage
    public let message: Message
    public var round2Config: Round2Configuration?
    let nonSigningCoordinator: NonSigningCoordinator
    let signingParticipant: SigningParticipant

    public init(
        configuration: Configuration,
        publicKeyPackage: PublicKeyPackage,
        signingParticipant: SigningParticipant,
        message: Message
    ) throws {
        self.configuration = configuration
        self.publicKeyPackage = publicKeyPackage
        self.message = message
        nonSigningCoordinator = try NonSigningCoordinator(
            configuration: configuration,
            publicKeyPackage: publicKeyPackage,
            message: message
        )
        self.signingParticipant = signingParticipant
    }

    public init(configuration: Configuration, publicKeyPackage: PublicKeyPackage, keyPackage: KeyPackage, message: Message) throws {
        let signingParticipant = SigningParticipant(
            keyPackage: keyPackage,
            publicKey: publicKeyPackage
        )

        try self.init(
            configuration: configuration,
            publicKeyPackage: publicKeyPackage,
            signingParticipant: signingParticipant,
            message: message
        )
    }

    public func receive(commitment: SigningCommitments) async throws {
        try await nonSigningCoordinator.receive(commitment: commitment)
    }

    public func createSigningPackage() async throws -> Round2Configuration {
        // sends its own commitment before creating the signign package
        let commitment = try signingParticipant.commit()
        try await nonSigningCoordinator.receive(commitment: commitment)

        // create the signing package
        let round2Config = try await nonSigningCoordinator.createSigningPackage()
        self.round2Config = round2Config
        return round2Config
    }

    public func receive(signatureShare: SignatureShare) async throws {
        try await nonSigningCoordinator.receive(signatureShare: signatureShare)
    }

    public func aggregate() async throws -> Signature {
        let round2Config = try await nonSigningCoordinator.round2ConfigPresent()
        // create own signature share
        signingParticipant.receive(round2Config: round2Config)
        let signatureShare = try signingParticipant.sign()

        // sends its own share before creating the signature
        try await nonSigningCoordinator.receive(signatureShare: signatureShare)

        // produce signature by aggregating all shares.
        let signature = try await nonSigningCoordinator.aggregate()

        return signature
    }

    public func verify(signature: Signature) async throws {
        try await nonSigningCoordinator.verify(signature: signature)
    }
}
