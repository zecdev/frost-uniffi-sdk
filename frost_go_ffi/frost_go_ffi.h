

// This file was autogenerated by some hot garbage in the `uniffi` crate.
// Trust me, you don't want to mess with it!



#include <stdbool.h>
#include <stdint.h>

// The following structs are used to implement the lowest level
// of the FFI, and thus useful to multiple uniffied crates.
// We ensure they are declared exactly once, with a header guard, UNIFFI_SHARED_H.
#ifdef UNIFFI_SHARED_H
	// We also try to prevent mixing versions of shared uniffi header structs.
	// If you add anything to the #else block, you must increment the version suffix in UNIFFI_SHARED_HEADER_V6
	#ifndef UNIFFI_SHARED_HEADER_V6
		#error Combining helper code from multiple versions of uniffi is not supported
	#endif // ndef UNIFFI_SHARED_HEADER_V6
#else
#define UNIFFI_SHARED_H
#define UNIFFI_SHARED_HEADER_V6
// ⚠️ Attention: If you change this #else block (ending in `#endif // def UNIFFI_SHARED_H`) you *must* ⚠️
// ⚠️ increment the version suffix in all instances of UNIFFI_SHARED_HEADER_V6 in this file.           ⚠️

typedef struct RustBuffer {
	int32_t capacity;
	int32_t len;
	uint8_t *data;
} RustBuffer;

typedef int32_t (*ForeignCallback)(uint64_t, int32_t, uint8_t *, int32_t, RustBuffer *);

// Task defined in Rust that Go executes
typedef void (*RustTaskCallback)(const void *, int8_t);

// Callback to execute Rust tasks using a Go routine
//
// Args:
//   executor: ForeignExecutor lowered into a uint64_t value
//   delay: Delay in MS
//   task: RustTaskCallback to call
//   task_data: data to pass the task callback
typedef int8_t (*ForeignExecutorCallback)(uint64_t, uint32_t, RustTaskCallback, void *);

typedef struct ForeignBytes {
	int32_t len;
	const uint8_t *data;
} ForeignBytes;

// Error definitions
typedef struct RustCallStatus {
	int8_t code;
	RustBuffer errorBuf;
} RustCallStatus;

// Continuation callback for UniFFI Futures
typedef void (*RustFutureContinuation)(void * , int8_t);

// ⚠️ Attention: If you change this #else block (ending in `#endif // def UNIFFI_SHARED_H`) you *must* ⚠️
// ⚠️ increment the version suffix in all instances of UNIFFI_SHARED_HEADER_V6 in this file.           ⚠️
#endif // def UNIFFI_SHARED_H

// Needed because we can't execute the callback directly from go.
void cgo_rust_task_callback_bridge_frost_go_ffi(RustTaskCallback, const void *, int8_t);

int8_t uniffiForeignExecutorCallbackfrost_go_ffi(uint64_t, uint32_t, RustTaskCallback, void*);

void uniffiFutureContinuationCallbackfrost_go_ffi(void*, int8_t);

void uniffi_frost_uniffi_sdk_fn_free_dkgpart1result(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_dkgpart2result(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_dkground1secretpackage(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_dkground2secretpackage(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_frostrandomizedparams(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_orchardaddress(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_constructor_orchardaddress_new_from_string(
	RustBuffer string,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_method_orchardaddress_string_encoded(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_orchardcommitivkrandomness(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_constructor_orchardcommitivkrandomness_new(
	RustBuffer bytes,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_method_orchardcommitivkrandomness_to_bytes(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_orchardfullviewingkey(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_constructor_orchardfullviewingkey_decode(
	RustBuffer string_enconded,
	RustBuffer network,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_constructor_orchardfullviewingkey_new_from_checked_parts(
	void* ak,
	void* nk,
	void* rivk,
	RustBuffer network,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_constructor_orchardfullviewingkey_new_from_validating_key_and_seed(
	void* validating_key,
	RustBuffer zip_32_seed,
	RustBuffer network,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_ak(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_derive_address(
	void* ptr,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_encode(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_nk(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_rivk(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_orchardkeyparts(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_constructor_orchardkeyparts_random(
	RustBuffer network,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_orchardnullifierderivingkey(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_constructor_orchardnullifierderivingkey_new(
	RustBuffer bytes,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_method_orchardnullifierderivingkey_to_bytes(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_free_orchardspendvalidatingkey(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_constructor_orchardspendvalidatingkey_from_bytes(
	RustBuffer bytes,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_method_orchardspendvalidatingkey_to_bytes(
	void* ptr,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_aggregate(
	RustBuffer signing_package,
	RustBuffer signature_shares,
	RustBuffer pubkey_package,
	RustBuffer randomizer,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_commitment_to_json(
	RustBuffer commitment,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_from_hex_string(
	RustBuffer hex_string,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_generate_nonces_and_commitments(
	RustBuffer key_package,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_identifier_from_json_string(
	RustBuffer string,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_identifier_from_string(
	RustBuffer string,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_identifier_from_uint16(
	uint16_t unsigned_uint,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_json_to_commitment(
	RustBuffer commitment_json,
	RustBuffer identifier,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_json_to_key_package(
	RustBuffer key_package_json,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_json_to_public_key_package(
	RustBuffer public_key_package_json,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_json_to_randomizer(
	RustBuffer randomizer_json,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_json_to_signature_share(
	RustBuffer signature_share_json,
	RustBuffer identifier,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_key_package_to_json(
	RustBuffer key_package,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_new_signing_package(
	RustBuffer message,
	RustBuffer commitments,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_func_part_1(
	RustBuffer participant_identifier,
	uint16_t max_signers,
	uint16_t min_signers,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_func_part_2(
	void* secret_package,
	RustBuffer round1_packages,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_part_3(
	void* secret_package,
	RustBuffer round1_packages,
	RustBuffer round2_packages,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_public_key_package_to_json(
	RustBuffer public_key_package,
	RustCallStatus* out_status
);

void* uniffi_frost_uniffi_sdk_fn_func_randomized_params_from_public_key_and_signing_package(
	RustBuffer public_key,
	RustBuffer signing_package,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_randomizer_from_params(
	void* randomized_params,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_randomizer_to_json(
	RustBuffer randomizer,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_sign(
	RustBuffer signing_package,
	RustBuffer nonces,
	RustBuffer key_package,
	RustBuffer randomizer,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_signature_share_package_to_json(
	RustBuffer signature_share,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_trusted_dealer_keygen_from(
	RustBuffer configuration,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_trusted_dealer_keygen_with_identifiers(
	RustBuffer configuration,
	RustBuffer participants,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_func_validate_config(
	RustBuffer config,
	RustCallStatus* out_status
);

RustBuffer uniffi_frost_uniffi_sdk_fn_func_verify_and_get_key_package_from(
	RustBuffer secret_share,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_func_verify_randomized_signature(
	RustBuffer randomizer,
	RustBuffer message,
	RustBuffer signature,
	RustBuffer pubkey,
	RustCallStatus* out_status
);

void uniffi_frost_uniffi_sdk_fn_func_verify_signature(
	RustBuffer message,
	RustBuffer signature,
	RustBuffer pubkey,
	RustCallStatus* out_status
);

RustBuffer ffi_frost_uniffi_sdk_rustbuffer_alloc(
	int32_t size,
	RustCallStatus* out_status
);

RustBuffer ffi_frost_uniffi_sdk_rustbuffer_from_bytes(
	ForeignBytes bytes,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rustbuffer_free(
	RustBuffer buf,
	RustCallStatus* out_status
);

RustBuffer ffi_frost_uniffi_sdk_rustbuffer_reserve(
	RustBuffer buf,
	int32_t additional,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_continuation_callback_set(
	RustFutureContinuation callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_u8(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_u8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_u8(
	void* handle,
	RustCallStatus* out_status
);

uint8_t ffi_frost_uniffi_sdk_rust_future_complete_u8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_i8(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_i8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_i8(
	void* handle,
	RustCallStatus* out_status
);

int8_t ffi_frost_uniffi_sdk_rust_future_complete_i8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_u16(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_u16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_u16(
	void* handle,
	RustCallStatus* out_status
);

uint16_t ffi_frost_uniffi_sdk_rust_future_complete_u16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_i16(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_i16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_i16(
	void* handle,
	RustCallStatus* out_status
);

int16_t ffi_frost_uniffi_sdk_rust_future_complete_i16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_u32(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_u32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_u32(
	void* handle,
	RustCallStatus* out_status
);

uint32_t ffi_frost_uniffi_sdk_rust_future_complete_u32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_i32(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_i32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_i32(
	void* handle,
	RustCallStatus* out_status
);

int32_t ffi_frost_uniffi_sdk_rust_future_complete_i32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_u64(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_u64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_u64(
	void* handle,
	RustCallStatus* out_status
);

uint64_t ffi_frost_uniffi_sdk_rust_future_complete_u64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_i64(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_i64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_i64(
	void* handle,
	RustCallStatus* out_status
);

int64_t ffi_frost_uniffi_sdk_rust_future_complete_i64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_f32(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_f32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_f32(
	void* handle,
	RustCallStatus* out_status
);

float ffi_frost_uniffi_sdk_rust_future_complete_f32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_f64(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_f64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_f64(
	void* handle,
	RustCallStatus* out_status
);

double ffi_frost_uniffi_sdk_rust_future_complete_f64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_pointer(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_pointer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_pointer(
	void* handle,
	RustCallStatus* out_status
);

void* ffi_frost_uniffi_sdk_rust_future_complete_pointer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_rust_buffer(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_rust_buffer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_rust_buffer(
	void* handle,
	RustCallStatus* out_status
);

RustBuffer ffi_frost_uniffi_sdk_rust_future_complete_rust_buffer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_poll_void(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_cancel_void(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_free_void(
	void* handle,
	RustCallStatus* out_status
);

void ffi_frost_uniffi_sdk_rust_future_complete_void(
	void* handle,
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_aggregate(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_commitment_to_json(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_from_hex_string(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_generate_nonces_and_commitments(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_identifier_from_json_string(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_identifier_from_string(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_identifier_from_uint16(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_json_to_commitment(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_json_to_key_package(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_json_to_public_key_package(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_json_to_randomizer(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_json_to_signature_share(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_key_package_to_json(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_new_signing_package(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_part_1(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_part_2(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_part_3(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_public_key_package_to_json(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_randomized_params_from_public_key_and_signing_package(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_randomizer_from_params(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_randomizer_to_json(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_sign(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_signature_share_package_to_json(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_from(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_with_identifiers(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_validate_config(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_verify_and_get_key_package_from(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_verify_randomized_signature(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_func_verify_signature(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_method_orchardaddress_string_encoded(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_method_orchardcommitivkrandomness_to_bytes(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_ak(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_derive_address(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_encode(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_nk(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_rivk(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_method_orchardnullifierderivingkey_to_bytes(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_method_orchardspendvalidatingkey_to_bytes(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_constructor_orchardaddress_new_from_string(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_constructor_orchardcommitivkrandomness_new(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_constructor_orchardfullviewingkey_decode(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_constructor_orchardfullviewingkey_new_from_checked_parts(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_constructor_orchardfullviewingkey_new_from_validating_key_and_seed(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_constructor_orchardkeyparts_random(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_constructor_orchardnullifierderivingkey_new(
	RustCallStatus* out_status
);

uint16_t uniffi_frost_uniffi_sdk_checksum_constructor_orchardspendvalidatingkey_from_bytes(
	RustCallStatus* out_status
);

uint32_t ffi_frost_uniffi_sdk_uniffi_contract_version(
	RustCallStatus* out_status
);



