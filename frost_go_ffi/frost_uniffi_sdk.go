package frost_uniffi_sdk

// #include <frost_go_ffi.h>
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync/atomic"
	"unsafe"
)

// This is needed, because as of go 1.24
// type RustBuffer C.RustBuffer cannot have methods,
// RustBuffer is treated as non-local type
type GoRustBuffer struct {
	inner C.RustBuffer
}

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() uint64
	Capacity() uint64
}

func RustBufferFromExternal(b RustBufferI) GoRustBuffer {
	return GoRustBuffer{
		inner: C.RustBuffer{
			capacity: C.uint64_t(b.Capacity()),
			len:      C.uint64_t(b.Len()),
			data:     (*C.uchar)(b.Data()),
		},
	}
}

func (cb GoRustBuffer) Capacity() uint64 {
	return uint64(cb.inner.capacity)
}

func (cb GoRustBuffer) Len() uint64 {
	return uint64(cb.inner.len)
}

func (cb GoRustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.inner.data)
}

func (cb GoRustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.inner.data), C.uint64_t(cb.inner.len))
	return bytes.NewReader(b)
}

func (cb GoRustBuffer) Free() {
	rustCall(func(status *C.RustCallStatus) bool {
		C.ffi_frost_uniffi_sdk_rustbuffer_free(cb.inner, status)
		return false
	})
}

func (cb GoRustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.inner.data), C.int(cb.inner.len))
}

func stringToRustBuffer(str string) C.RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) C.RustBuffer {
	if len(b) == 0 {
		return C.RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes{
		len:  C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}

	return rustCall(func(status *C.RustCallStatus) C.RustBuffer {
		return C.ffi_frost_uniffi_sdk_rustbuffer_from_bytes(foreign, status)
	})
}

type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) C.RustBuffer
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) C.RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}

func rustCallWithError[E any, U any](converter BufReader[*E], callback func(*C.RustCallStatus) U) (U, *E) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)
	return returnValue, err
}

func checkCallStatus[E any](converter BufReader[*E], status C.RustCallStatus) *E {
	switch status.code {
	case 0:
		return nil
	case 1:
		return LiftFromRustBuffer(converter, GoRustBuffer{inner: status.errorBuf})
	case 2:
		// when the rust code sees a panic, it tries to construct a rustBuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(GoRustBuffer{inner: status.errorBuf})))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		panic(fmt.Errorf("unknown status code: %d", status.code))
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a C.RustBuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(GoRustBuffer{
				inner: status.errorBuf,
			})))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError[error](nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}

type NativeError interface {
	AsError() error
}

func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {

	uniffiCheckChecksums()
}

func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 26
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_frost_uniffi_sdk_uniffi_contract_version()
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("frost_uniffi_sdk: UniFFI contract version mismatch")
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_aggregate()
		})
		if checksum != 14107 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_aggregate: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_commitment_to_json()
		})
		if checksum != 37322 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_commitment_to_json: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_from_hex_string()
		})
		if checksum != 6554 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_from_hex_string: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_generate_nonces_and_commitments()
		})
		if checksum != 61549 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_generate_nonces_and_commitments: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_identifier_from_json_string()
		})
		if checksum != 4885 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_identifier_from_json_string: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_identifier_from_string()
		})
		if checksum != 17207 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_identifier_from_string: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_identifier_from_uint16()
		})
		if checksum != 13096 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_identifier_from_uint16: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_json_to_commitment()
		})
		if checksum != 377 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_json_to_commitment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_json_to_key_package()
		})
		if checksum != 50636 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_json_to_key_package: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_json_to_public_key_package()
		})
		if checksum != 47876 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_json_to_public_key_package: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_json_to_randomizer()
		})
		if checksum != 43415 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_json_to_randomizer: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_json_to_signature_share()
		})
		if checksum != 20444 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_json_to_signature_share: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_key_package_to_json()
		})
		if checksum != 27984 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_key_package_to_json: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_new_signing_package()
		})
		if checksum != 59539 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_new_signing_package: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_part_1()
		})
		if checksum != 48695 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_part_1: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_part_2()
		})
		if checksum != 4947 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_part_2: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_part_3()
		})
		if checksum != 39757 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_part_3: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_public_key_package_to_json()
		})
		if checksum != 13971 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_public_key_package_to_json: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_randomized_params_from_public_key_and_signing_package()
		})
		if checksum != 42974 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_randomized_params_from_public_key_and_signing_package: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_randomizer_from_params()
		})
		if checksum != 26841 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_randomizer_from_params: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_randomizer_to_json()
		})
		if checksum != 17475 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_randomizer_to_json: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_sign()
		})
		if checksum != 22743 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_sign: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_signature_share_package_to_json()
		})
		if checksum != 17380 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_signature_share_package_to_json: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_from()
		})
		if checksum != 4367 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_from: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_with_identifiers()
		})
		if checksum != 25579 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_with_identifiers: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_validate_config()
		})
		if checksum != 42309 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_validate_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_verify_and_get_key_package_from()
		})
		if checksum != 52603 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_verify_and_get_key_package_from: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_verify_randomized_signature()
		})
		if checksum != 61115 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_verify_randomized_signature: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_verify_signature()
		})
		if checksum != 31978 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_verify_signature: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_method_orchardaddress_string_encoded()
		})
		if checksum != 17163 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_method_orchardaddress_string_encoded: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_method_orchardcommitivkrandomness_to_bytes()
		})
		if checksum != 45794 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_method_orchardcommitivkrandomness_to_bytes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_ak()
		})
		if checksum != 17920 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_ak: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_derive_address()
		})
		if checksum != 39349 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_derive_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_encode()
		})
		if checksum != 15911 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_encode: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_nk()
		})
		if checksum != 26127 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_nk: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_rivk()
		})
		if checksum != 62140 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_method_orchardfullviewingkey_rivk: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_method_orchardnullifierderivingkey_to_bytes()
		})
		if checksum != 38576 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_method_orchardnullifierderivingkey_to_bytes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_method_orchardspendvalidatingkey_to_bytes()
		})
		if checksum != 32867 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_method_orchardspendvalidatingkey_to_bytes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_constructor_orchardaddress_new_from_string()
		})
		if checksum != 54798 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_constructor_orchardaddress_new_from_string: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_constructor_orchardcommitivkrandomness_new()
		})
		if checksum != 65326 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_constructor_orchardcommitivkrandomness_new: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_constructor_orchardfullviewingkey_decode()
		})
		if checksum != 15654 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_constructor_orchardfullviewingkey_decode: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_constructor_orchardfullviewingkey_new_from_checked_parts()
		})
		if checksum != 32693 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_constructor_orchardfullviewingkey_new_from_checked_parts: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_constructor_orchardfullviewingkey_new_from_validating_key_and_seed()
		})
		if checksum != 29602 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_constructor_orchardfullviewingkey_new_from_validating_key_and_seed: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_constructor_orchardkeyparts_random()
		})
		if checksum != 17995 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_constructor_orchardkeyparts_random: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_constructor_orchardnullifierderivingkey_new()
		})
		if checksum != 3116 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_constructor_orchardnullifierderivingkey_new: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_constructor_orchardspendvalidatingkey_from_bytes()
		})
		if checksum != 52420 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_constructor_orchardspendvalidatingkey_from_bytes: UniFFI API checksum mismatch")
		}
	}
}

type FfiConverterUint16 struct{}

var FfiConverterUint16INSTANCE = FfiConverterUint16{}

func (FfiConverterUint16) Lower(value uint16) C.uint16_t {
	return C.uint16_t(value)
}

func (FfiConverterUint16) Write(writer io.Writer, value uint16) {
	writeUint16(writer, value)
}

func (FfiConverterUint16) Lift(value C.uint16_t) uint16 {
	return uint16(value)
}

func (FfiConverterUint16) Read(reader io.Reader) uint16 {
	return readUint16(reader)
}

type FfiDestroyerUint16 struct{}

func (FfiDestroyerUint16) Destroy(_ uint16) {}

type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) C.RustBuffer {
	return stringToRustBuffer(value)
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct{}

func (FfiDestroyerString) Destroy(_ string) {}

type FfiConverterBytes struct{}

var FfiConverterBytesINSTANCE = FfiConverterBytes{}

func (c FfiConverterBytes) Lower(value []byte) C.RustBuffer {
	return LowerIntoRustBuffer[[]byte](c, value)
}

func (c FfiConverterBytes) Write(writer io.Writer, value []byte) {
	if len(value) > math.MaxInt32 {
		panic("[]byte is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := writer.Write(value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing []byte, expected %d, written %d", len(value), write_length))
	}
}

func (c FfiConverterBytes) Lift(rb RustBufferI) []byte {
	return LiftFromRustBuffer[[]byte](c, rb)
}

func (c FfiConverterBytes) Read(reader io.Reader) []byte {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading []byte, expected %d, read %d", length, read_length))
	}
	return buffer
}

type FfiDestroyerBytes struct{}

func (FfiDestroyerBytes) Destroy(_ []byte) {}

// Below is an implementation of synchronization requirements outlined in the link.
// https://github.com/mozilla/uniffi-rs/blob/0dc031132d9493ca812c3af6e7dd60ad2ea95bf0/uniffi_bindgen/src/bindings/kotlin/templates/ObjectRuntime.kt#L31

type FfiObject struct {
	pointer       unsafe.Pointer
	callCounter   atomic.Int64
	cloneFunction func(unsafe.Pointer, *C.RustCallStatus) unsafe.Pointer
	freeFunction  func(unsafe.Pointer, *C.RustCallStatus)
	destroyed     atomic.Bool
}

func newFfiObject(
	pointer unsafe.Pointer,
	cloneFunction func(unsafe.Pointer, *C.RustCallStatus) unsafe.Pointer,
	freeFunction func(unsafe.Pointer, *C.RustCallStatus),
) FfiObject {
	return FfiObject{
		pointer:       pointer,
		cloneFunction: cloneFunction,
		freeFunction:  freeFunction,
	}
}

func (ffiObject *FfiObject) incrementPointer(debugName string) unsafe.Pointer {
	for {
		counter := ffiObject.callCounter.Load()
		if counter <= -1 {
			panic(fmt.Errorf("%v object has already been destroyed", debugName))
		}
		if counter == math.MaxInt64 {
			panic(fmt.Errorf("%v object call counter would overflow", debugName))
		}
		if ffiObject.callCounter.CompareAndSwap(counter, counter+1) {
			break
		}
	}

	return rustCall(func(status *C.RustCallStatus) unsafe.Pointer {
		return ffiObject.cloneFunction(ffiObject.pointer, status)
	})
}

func (ffiObject *FfiObject) decrementPointer() {
	if ffiObject.callCounter.Add(-1) == -1 {
		ffiObject.freeRustArcPtr()
	}
}

func (ffiObject *FfiObject) destroy() {
	if ffiObject.destroyed.CompareAndSwap(false, true) {
		if ffiObject.callCounter.Add(-1) == -1 {
			ffiObject.freeRustArcPtr()
		}
	}
}

func (ffiObject *FfiObject) freeRustArcPtr() {
	rustCall(func(status *C.RustCallStatus) int32 {
		ffiObject.freeFunction(ffiObject.pointer, status)
		return 0
	})
}

type DkgPart1ResultInterface interface {
}
type DkgPart1Result struct {
	ffiObject FfiObject
}

func (object *DkgPart1Result) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterDkgPart1Result struct{}

var FfiConverterDkgPart1ResultINSTANCE = FfiConverterDkgPart1Result{}

func (c FfiConverterDkgPart1Result) Lift(pointer unsafe.Pointer) *DkgPart1Result {
	result := &DkgPart1Result{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_dkgpart1result(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_dkgpart1result(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*DkgPart1Result).Destroy)
	return result
}

func (c FfiConverterDkgPart1Result) Read(reader io.Reader) *DkgPart1Result {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterDkgPart1Result) Lower(value *DkgPart1Result) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*DkgPart1Result")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterDkgPart1Result) Write(writer io.Writer, value *DkgPart1Result) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerDkgPart1Result struct{}

func (_ FfiDestroyerDkgPart1Result) Destroy(value *DkgPart1Result) {
	value.Destroy()
}

type DkgPart2ResultInterface interface {
}
type DkgPart2Result struct {
	ffiObject FfiObject
}

func (object *DkgPart2Result) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterDkgPart2Result struct{}

var FfiConverterDkgPart2ResultINSTANCE = FfiConverterDkgPart2Result{}

func (c FfiConverterDkgPart2Result) Lift(pointer unsafe.Pointer) *DkgPart2Result {
	result := &DkgPart2Result{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_dkgpart2result(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_dkgpart2result(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*DkgPart2Result).Destroy)
	return result
}

func (c FfiConverterDkgPart2Result) Read(reader io.Reader) *DkgPart2Result {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterDkgPart2Result) Lower(value *DkgPart2Result) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*DkgPart2Result")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterDkgPart2Result) Write(writer io.Writer, value *DkgPart2Result) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerDkgPart2Result struct{}

func (_ FfiDestroyerDkgPart2Result) Destroy(value *DkgPart2Result) {
	value.Destroy()
}

type DkgRound1SecretPackageInterface interface {
}
type DkgRound1SecretPackage struct {
	ffiObject FfiObject
}

func (object *DkgRound1SecretPackage) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterDkgRound1SecretPackage struct{}

var FfiConverterDkgRound1SecretPackageINSTANCE = FfiConverterDkgRound1SecretPackage{}

func (c FfiConverterDkgRound1SecretPackage) Lift(pointer unsafe.Pointer) *DkgRound1SecretPackage {
	result := &DkgRound1SecretPackage{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_dkground1secretpackage(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_dkground1secretpackage(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*DkgRound1SecretPackage).Destroy)
	return result
}

func (c FfiConverterDkgRound1SecretPackage) Read(reader io.Reader) *DkgRound1SecretPackage {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterDkgRound1SecretPackage) Lower(value *DkgRound1SecretPackage) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*DkgRound1SecretPackage")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterDkgRound1SecretPackage) Write(writer io.Writer, value *DkgRound1SecretPackage) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerDkgRound1SecretPackage struct{}

func (_ FfiDestroyerDkgRound1SecretPackage) Destroy(value *DkgRound1SecretPackage) {
	value.Destroy()
}

type DkgRound2SecretPackageInterface interface {
}
type DkgRound2SecretPackage struct {
	ffiObject FfiObject
}

func (object *DkgRound2SecretPackage) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterDkgRound2SecretPackage struct{}

var FfiConverterDkgRound2SecretPackageINSTANCE = FfiConverterDkgRound2SecretPackage{}

func (c FfiConverterDkgRound2SecretPackage) Lift(pointer unsafe.Pointer) *DkgRound2SecretPackage {
	result := &DkgRound2SecretPackage{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_dkground2secretpackage(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_dkground2secretpackage(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*DkgRound2SecretPackage).Destroy)
	return result
}

func (c FfiConverterDkgRound2SecretPackage) Read(reader io.Reader) *DkgRound2SecretPackage {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterDkgRound2SecretPackage) Lower(value *DkgRound2SecretPackage) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*DkgRound2SecretPackage")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterDkgRound2SecretPackage) Write(writer io.Writer, value *DkgRound2SecretPackage) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerDkgRound2SecretPackage struct{}

func (_ FfiDestroyerDkgRound2SecretPackage) Destroy(value *DkgRound2SecretPackage) {
	value.Destroy()
}

type FrostRandomizedParamsInterface interface {
}
type FrostRandomizedParams struct {
	ffiObject FfiObject
}

func (object *FrostRandomizedParams) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterFrostRandomizedParams struct{}

var FfiConverterFrostRandomizedParamsINSTANCE = FfiConverterFrostRandomizedParams{}

func (c FfiConverterFrostRandomizedParams) Lift(pointer unsafe.Pointer) *FrostRandomizedParams {
	result := &FrostRandomizedParams{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_frostrandomizedparams(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_frostrandomizedparams(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*FrostRandomizedParams).Destroy)
	return result
}

func (c FfiConverterFrostRandomizedParams) Read(reader io.Reader) *FrostRandomizedParams {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterFrostRandomizedParams) Lower(value *FrostRandomizedParams) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*FrostRandomizedParams")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterFrostRandomizedParams) Write(writer io.Writer, value *FrostRandomizedParams) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerFrostRandomizedParams struct{}

func (_ FfiDestroyerFrostRandomizedParams) Destroy(value *FrostRandomizedParams) {
	value.Destroy()
}

// An Zcash Orchard Address and its associated network type.
type OrchardAddressInterface interface {
	// Returns the string-encoded form of this Orchard Address (A
	// Unified Address containing only the orchard receiver.)
	StringEncoded() string
}

// An Zcash Orchard Address and its associated network type.
type OrchardAddress struct {
	ffiObject FfiObject
}

// Creates an [`OrchardAddress`] from its string-encoded form
// If the string is invalid `Err(OrchardKeyError::DeserializationError)`
// is returned in the Result.
func OrchardAddressNewFromString(string string) (*OrchardAddress, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_constructor_orchardaddress_new_from_string(FfiConverterStringINSTANCE.Lower(string), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *OrchardAddress
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOrchardAddressINSTANCE.Lift(_uniffiRV), nil
	}
}

// Returns the string-encoded form of this Orchard Address (A
// Unified Address containing only the orchard receiver.)
func (_self *OrchardAddress) StringEncoded() string {
	_pointer := _self.ffiObject.incrementPointer("*OrchardAddress")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_method_orchardaddress_string_encoded(
				_pointer, _uniffiStatus),
		}
	}))
}
func (object *OrchardAddress) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOrchardAddress struct{}

var FfiConverterOrchardAddressINSTANCE = FfiConverterOrchardAddress{}

func (c FfiConverterOrchardAddress) Lift(pointer unsafe.Pointer) *OrchardAddress {
	result := &OrchardAddress{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_orchardaddress(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_orchardaddress(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*OrchardAddress).Destroy)
	return result
}

func (c FfiConverterOrchardAddress) Read(reader io.Reader) *OrchardAddress {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOrchardAddress) Lower(value *OrchardAddress) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OrchardAddress")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterOrchardAddress) Write(writer io.Writer, value *OrchardAddress) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOrchardAddress struct{}

func (_ FfiDestroyerOrchardAddress) Destroy(value *OrchardAddress) {
	value.Destroy()
}

// The `rivk` component of an Orchard Full Viewing Key.
// This is intended for key backup purposes.
// - Note: See [ZF FROST Book - Technical Details](https://frost.zfnd.org/zcash/technical-details.html)
type OrchardCommitIvkRandomnessInterface interface {
	ToBytes() []byte
}

// The `rivk` component of an Orchard Full Viewing Key.
// This is intended for key backup purposes.
// - Note: See [ZF FROST Book - Technical Details](https://frost.zfnd.org/zcash/technical-details.html)
type OrchardCommitIvkRandomness struct {
	ffiObject FfiObject
}

// Creates a `rivk` from a sequence of bytes. Returns [`OrchardKeyError::DeserializationError`]
// if these bytes can't be deserialized into a valid `rivk`
func NewOrchardCommitIvkRandomness(bytes []byte) (*OrchardCommitIvkRandomness, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_constructor_orchardcommitivkrandomness_new(FfiConverterBytesINSTANCE.Lower(bytes), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *OrchardCommitIvkRandomness
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOrchardCommitIvkRandomnessINSTANCE.Lift(_uniffiRV), nil
	}
}

func (_self *OrchardCommitIvkRandomness) ToBytes() []byte {
	_pointer := _self.ffiObject.incrementPointer("*OrchardCommitIvkRandomness")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_method_orchardcommitivkrandomness_to_bytes(
				_pointer, _uniffiStatus),
		}
	}))
}
func (object *OrchardCommitIvkRandomness) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOrchardCommitIvkRandomness struct{}

var FfiConverterOrchardCommitIvkRandomnessINSTANCE = FfiConverterOrchardCommitIvkRandomness{}

func (c FfiConverterOrchardCommitIvkRandomness) Lift(pointer unsafe.Pointer) *OrchardCommitIvkRandomness {
	result := &OrchardCommitIvkRandomness{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_orchardcommitivkrandomness(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_orchardcommitivkrandomness(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*OrchardCommitIvkRandomness).Destroy)
	return result
}

func (c FfiConverterOrchardCommitIvkRandomness) Read(reader io.Reader) *OrchardCommitIvkRandomness {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOrchardCommitIvkRandomness) Lower(value *OrchardCommitIvkRandomness) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OrchardCommitIvkRandomness")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterOrchardCommitIvkRandomness) Write(writer io.Writer, value *OrchardCommitIvkRandomness) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOrchardCommitIvkRandomness struct{}

func (_ FfiDestroyerOrchardCommitIvkRandomness) Destroy(value *OrchardCommitIvkRandomness) {
	value.Destroy()
}

// A UnifiedViewingKey containing only an Orchard component and
// its associated network constant.
type OrchardFullViewingKeyInterface interface {
	// Returns the Spend Validating Key component of this Orchard FVK
	Ak() *OrchardSpendValidatingKey
	// derives external address 0 of this Orchard Full viewing key.
	DeriveAddress() (*OrchardAddress, error)
	// Encodes a [`OrchardFullViewingKey`] to its Unified Full Viewing Key
	// string-encoded format. If this operation fails, it returns
	// `Err(OrchardKeyError::DeserializationError)`. This should be straight
	// forward and an error thrown could indicate another kind of issue like a
	// PEBKAC.
	Encode() (string, error)
	Nk() *OrchardNullifierDerivingKey
	// Returns the External Scope of this FVK
	Rivk() *OrchardCommitIvkRandomness
}

// A UnifiedViewingKey containing only an Orchard component and
// its associated network constant.
type OrchardFullViewingKey struct {
	ffiObject FfiObject
}

// Decodes a [`OrchardFullViewingKey`] from its Unified Full Viewing Key
// string-encoded format. If this operation fails, it returns
// `Err(OrchardKeyError::DeserializationError)`
func OrchardFullViewingKeyDecode(stringEnconded string, network ZcashNetwork) (*OrchardFullViewingKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_constructor_orchardfullviewingkey_decode(FfiConverterStringINSTANCE.Lower(stringEnconded), FfiConverterZcashNetworkINSTANCE.Lower(network), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *OrchardFullViewingKey
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOrchardFullViewingKeyINSTANCE.Lift(_uniffiRV), nil
	}
}

// Creates an [`OrchardFullViewingKey`] from its checked composing parts
// and its associated Network constant.
func OrchardFullViewingKeyNewFromCheckedParts(ak *OrchardSpendValidatingKey, nk *OrchardNullifierDerivingKey, rivk *OrchardCommitIvkRandomness, network ZcashNetwork) (*OrchardFullViewingKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_constructor_orchardfullviewingkey_new_from_checked_parts(FfiConverterOrchardSpendValidatingKeyINSTANCE.Lower(ak), FfiConverterOrchardNullifierDerivingKeyINSTANCE.Lower(nk), FfiConverterOrchardCommitIvkRandomnessINSTANCE.Lower(rivk), FfiConverterZcashNetworkINSTANCE.Lower(network), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *OrchardFullViewingKey
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOrchardFullViewingKeyINSTANCE.Lift(_uniffiRV), nil
	}
}

// Creates a new FullViewingKey from a ZIP-32 Seed and validating key
// using the `Network` coin type on `AccountId(0u32)`
// see https://frost.zfnd.org/zcash/technical-details.html for more
// information.
func OrchardFullViewingKeyNewFromValidatingKeyAndSeed(validatingKey *OrchardSpendValidatingKey, zip32Seed []byte, network ZcashNetwork) (*OrchardFullViewingKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_constructor_orchardfullviewingkey_new_from_validating_key_and_seed(FfiConverterOrchardSpendValidatingKeyINSTANCE.Lower(validatingKey), FfiConverterBytesINSTANCE.Lower(zip32Seed), FfiConverterZcashNetworkINSTANCE.Lower(network), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *OrchardFullViewingKey
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOrchardFullViewingKeyINSTANCE.Lift(_uniffiRV), nil
	}
}

// Returns the Spend Validating Key component of this Orchard FVK
func (_self *OrchardFullViewingKey) Ak() *OrchardSpendValidatingKey {
	_pointer := _self.ffiObject.incrementPointer("*OrchardFullViewingKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOrchardSpendValidatingKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_ak(
			_pointer, _uniffiStatus)
	}))
}

// derives external address 0 of this Orchard Full viewing key.
func (_self *OrchardFullViewingKey) DeriveAddress() (*OrchardAddress, error) {
	_pointer := _self.ffiObject.incrementPointer("*OrchardFullViewingKey")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_derive_address(
			_pointer, _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *OrchardAddress
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOrchardAddressINSTANCE.Lift(_uniffiRV), nil
	}
}

// Encodes a [`OrchardFullViewingKey`] to its Unified Full Viewing Key
// string-encoded format. If this operation fails, it returns
// `Err(OrchardKeyError::DeserializationError)`. This should be straight
// forward and an error thrown could indicate another kind of issue like a
// PEBKAC.
func (_self *OrchardFullViewingKey) Encode() (string, error) {
	_pointer := _self.ffiObject.incrementPointer("*OrchardFullViewingKey")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_encode(
				_pointer, _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), nil
	}
}

func (_self *OrchardFullViewingKey) Nk() *OrchardNullifierDerivingKey {
	_pointer := _self.ffiObject.incrementPointer("*OrchardFullViewingKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOrchardNullifierDerivingKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_nk(
			_pointer, _uniffiStatus)
	}))
}

// Returns the External Scope of this FVK
func (_self *OrchardFullViewingKey) Rivk() *OrchardCommitIvkRandomness {
	_pointer := _self.ffiObject.incrementPointer("*OrchardFullViewingKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOrchardCommitIvkRandomnessINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_method_orchardfullviewingkey_rivk(
			_pointer, _uniffiStatus)
	}))
}
func (object *OrchardFullViewingKey) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOrchardFullViewingKey struct{}

var FfiConverterOrchardFullViewingKeyINSTANCE = FfiConverterOrchardFullViewingKey{}

func (c FfiConverterOrchardFullViewingKey) Lift(pointer unsafe.Pointer) *OrchardFullViewingKey {
	result := &OrchardFullViewingKey{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_orchardfullviewingkey(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_orchardfullviewingkey(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*OrchardFullViewingKey).Destroy)
	return result
}

func (c FfiConverterOrchardFullViewingKey) Read(reader io.Reader) *OrchardFullViewingKey {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOrchardFullViewingKey) Lower(value *OrchardFullViewingKey) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OrchardFullViewingKey")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterOrchardFullViewingKey) Write(writer io.Writer, value *OrchardFullViewingKey) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOrchardFullViewingKey struct{}

func (_ FfiDestroyerOrchardFullViewingKey) Destroy(value *OrchardFullViewingKey) {
	value.Destroy()
}

// This responds to Backup and DKG requirements
// for FROST.
//
// - Note: See [FROST Book backup section](https://frost.zfnd.org/zcash/technical-details.html#backing-up-key-shares)
type OrchardKeyPartsInterface interface {
}

// This responds to Backup and DKG requirements
// for FROST.
//
// - Note: See [FROST Book backup section](https://frost.zfnd.org/zcash/technical-details.html#backing-up-key-shares)
type OrchardKeyParts struct {
	ffiObject FfiObject
}

// Creates a Random `nk` and `rivk` from a random Spending Key
// originated from a random 24-word Mnemonic seed which is tossed
// away.
// This responds to Backup and DKG requirements
// for FROST.
//
// - Note: See [FROST Book backup section](https://frost.zfnd.org/zcash/technical-details.html#backing-up-key-shares)
func OrchardKeyPartsRandom(network ZcashNetwork) (*OrchardKeyParts, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_constructor_orchardkeyparts_random(FfiConverterZcashNetworkINSTANCE.Lower(network), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *OrchardKeyParts
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOrchardKeyPartsINSTANCE.Lift(_uniffiRV), nil
	}
}

func (object *OrchardKeyParts) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOrchardKeyParts struct{}

var FfiConverterOrchardKeyPartsINSTANCE = FfiConverterOrchardKeyParts{}

func (c FfiConverterOrchardKeyParts) Lift(pointer unsafe.Pointer) *OrchardKeyParts {
	result := &OrchardKeyParts{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_orchardkeyparts(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_orchardkeyparts(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*OrchardKeyParts).Destroy)
	return result
}

func (c FfiConverterOrchardKeyParts) Read(reader io.Reader) *OrchardKeyParts {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOrchardKeyParts) Lower(value *OrchardKeyParts) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OrchardKeyParts")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterOrchardKeyParts) Write(writer io.Writer, value *OrchardKeyParts) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOrchardKeyParts struct{}

func (_ FfiDestroyerOrchardKeyParts) Destroy(value *OrchardKeyParts) {
	value.Destroy()
}

// The Orchard Nullifier Deriving Key component of an
// Orchard full viewing key. This is intended for key backup
// purposes.
// - Note: See [ZF FROST Book - Technical Details](https://frost.zfnd.org/zcash/technical-details.html)
type OrchardNullifierDerivingKeyInterface interface {
	// Serializes [`OrchardNullifierDerivingKey`] to a sequence of bytes.
	ToBytes() []byte
}

// The Orchard Nullifier Deriving Key component of an
// Orchard full viewing key. This is intended for key backup
// purposes.
// - Note: See [ZF FROST Book - Technical Details](https://frost.zfnd.org/zcash/technical-details.html)
type OrchardNullifierDerivingKey struct {
	ffiObject FfiObject
}

// Creates an [`OrchardNullifierDerivingKey`] from a sequence of bytes.
// If the byte sequence is not suitable for doing so, it will return an
// [`Err(OrchardKeyError::DeserializationError)`]
func NewOrchardNullifierDerivingKey(bytes []byte) (*OrchardNullifierDerivingKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_constructor_orchardnullifierderivingkey_new(FfiConverterBytesINSTANCE.Lower(bytes), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *OrchardNullifierDerivingKey
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOrchardNullifierDerivingKeyINSTANCE.Lift(_uniffiRV), nil
	}
}

// Serializes [`OrchardNullifierDerivingKey`] to a sequence of bytes.
func (_self *OrchardNullifierDerivingKey) ToBytes() []byte {
	_pointer := _self.ffiObject.incrementPointer("*OrchardNullifierDerivingKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_method_orchardnullifierderivingkey_to_bytes(
				_pointer, _uniffiStatus),
		}
	}))
}
func (object *OrchardNullifierDerivingKey) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOrchardNullifierDerivingKey struct{}

var FfiConverterOrchardNullifierDerivingKeyINSTANCE = FfiConverterOrchardNullifierDerivingKey{}

func (c FfiConverterOrchardNullifierDerivingKey) Lift(pointer unsafe.Pointer) *OrchardNullifierDerivingKey {
	result := &OrchardNullifierDerivingKey{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_orchardnullifierderivingkey(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_orchardnullifierderivingkey(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*OrchardNullifierDerivingKey).Destroy)
	return result
}

func (c FfiConverterOrchardNullifierDerivingKey) Read(reader io.Reader) *OrchardNullifierDerivingKey {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOrchardNullifierDerivingKey) Lower(value *OrchardNullifierDerivingKey) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OrchardNullifierDerivingKey")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterOrchardNullifierDerivingKey) Write(writer io.Writer, value *OrchardNullifierDerivingKey) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOrchardNullifierDerivingKey struct{}

func (_ FfiDestroyerOrchardNullifierDerivingKey) Destroy(value *OrchardNullifierDerivingKey) {
	value.Destroy()
}

// The `ak` component of an Orchard Full Viewing key. This shall be
// derived from the Spend Authorizing Key `ask`
type OrchardSpendValidatingKeyInterface interface {
	// Serialized the [`OrchardSpendValidatingKey`] into bytes for
	// backup purposes.
	// - Note: See [ZF FROST Book - Technical Details](https://frost.zfnd.org/zcash/technical-details.html)
	// to deserialize use the `OrchardSpendValidatingKey::from_bytes`
	// constructor
	ToBytes() []byte
}

// The `ak` component of an Orchard Full Viewing key. This shall be
// derived from the Spend Authorizing Key `ask`
type OrchardSpendValidatingKey struct {
	ffiObject FfiObject
}

// Deserialized the [`OrchardSpendValidatingKey`] into bytes for
// backup purposes.
// - Note: See [ZF FROST Book - Technical Details](https://frost.zfnd.org/zcash/technical-details.html)
// to serialize use the `OrchardSpendValidatingKey::to_bytes`
// constructor
func OrchardSpendValidatingKeyFromBytes(bytes []byte) (*OrchardSpendValidatingKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[OrchardKeyError](FfiConverterOrchardKeyError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_constructor_orchardspendvalidatingkey_from_bytes(FfiConverterBytesINSTANCE.Lower(bytes), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *OrchardSpendValidatingKey
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOrchardSpendValidatingKeyINSTANCE.Lift(_uniffiRV), nil
	}
}

// Serialized the [`OrchardSpendValidatingKey`] into bytes for
// backup purposes.
// - Note: See [ZF FROST Book - Technical Details](https://frost.zfnd.org/zcash/technical-details.html)
// to deserialize use the `OrchardSpendValidatingKey::from_bytes`
// constructor
func (_self *OrchardSpendValidatingKey) ToBytes() []byte {
	_pointer := _self.ffiObject.incrementPointer("*OrchardSpendValidatingKey")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBytesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_method_orchardspendvalidatingkey_to_bytes(
				_pointer, _uniffiStatus),
		}
	}))
}
func (object *OrchardSpendValidatingKey) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOrchardSpendValidatingKey struct{}

var FfiConverterOrchardSpendValidatingKeyINSTANCE = FfiConverterOrchardSpendValidatingKey{}

func (c FfiConverterOrchardSpendValidatingKey) Lift(pointer unsafe.Pointer) *OrchardSpendValidatingKey {
	result := &OrchardSpendValidatingKey{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_frost_uniffi_sdk_fn_clone_orchardspendvalidatingkey(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_frost_uniffi_sdk_fn_free_orchardspendvalidatingkey(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*OrchardSpendValidatingKey).Destroy)
	return result
}

func (c FfiConverterOrchardSpendValidatingKey) Read(reader io.Reader) *OrchardSpendValidatingKey {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOrchardSpendValidatingKey) Lower(value *OrchardSpendValidatingKey) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OrchardSpendValidatingKey")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterOrchardSpendValidatingKey) Write(writer io.Writer, value *OrchardSpendValidatingKey) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOrchardSpendValidatingKey struct{}

func (_ FfiDestroyerOrchardSpendValidatingKey) Destroy(value *OrchardSpendValidatingKey) {
	value.Destroy()
}

type Configuration struct {
	MinSigners uint16
	MaxSigners uint16
	Secret     []byte
}

func (r *Configuration) Destroy() {
	FfiDestroyerUint16{}.Destroy(r.MinSigners)
	FfiDestroyerUint16{}.Destroy(r.MaxSigners)
	FfiDestroyerBytes{}.Destroy(r.Secret)
}

type FfiConverterConfiguration struct{}

var FfiConverterConfigurationINSTANCE = FfiConverterConfiguration{}

func (c FfiConverterConfiguration) Lift(rb RustBufferI) Configuration {
	return LiftFromRustBuffer[Configuration](c, rb)
}

func (c FfiConverterConfiguration) Read(reader io.Reader) Configuration {
	return Configuration{
		FfiConverterUint16INSTANCE.Read(reader),
		FfiConverterUint16INSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterConfiguration) Lower(value Configuration) C.RustBuffer {
	return LowerIntoRustBuffer[Configuration](c, value)
}

func (c FfiConverterConfiguration) Write(writer io.Writer, value Configuration) {
	FfiConverterUint16INSTANCE.Write(writer, value.MinSigners)
	FfiConverterUint16INSTANCE.Write(writer, value.MaxSigners)
	FfiConverterBytesINSTANCE.Write(writer, value.Secret)
}

type FfiDestroyerConfiguration struct{}

func (_ FfiDestroyerConfiguration) Destroy(value Configuration) {
	value.Destroy()
}

type DkgPart3Result struct {
	PublicKeyPackage FrostPublicKeyPackage
	KeyPackage       FrostKeyPackage
}

func (r *DkgPart3Result) Destroy() {
	FfiDestroyerFrostPublicKeyPackage{}.Destroy(r.PublicKeyPackage)
	FfiDestroyerFrostKeyPackage{}.Destroy(r.KeyPackage)
}

type FfiConverterDkgPart3Result struct{}

var FfiConverterDkgPart3ResultINSTANCE = FfiConverterDkgPart3Result{}

func (c FfiConverterDkgPart3Result) Lift(rb RustBufferI) DkgPart3Result {
	return LiftFromRustBuffer[DkgPart3Result](c, rb)
}

func (c FfiConverterDkgPart3Result) Read(reader io.Reader) DkgPart3Result {
	return DkgPart3Result{
		FfiConverterFrostPublicKeyPackageINSTANCE.Read(reader),
		FfiConverterFrostKeyPackageINSTANCE.Read(reader),
	}
}

func (c FfiConverterDkgPart3Result) Lower(value DkgPart3Result) C.RustBuffer {
	return LowerIntoRustBuffer[DkgPart3Result](c, value)
}

func (c FfiConverterDkgPart3Result) Write(writer io.Writer, value DkgPart3Result) {
	FfiConverterFrostPublicKeyPackageINSTANCE.Write(writer, value.PublicKeyPackage)
	FfiConverterFrostKeyPackageINSTANCE.Write(writer, value.KeyPackage)
}

type FfiDestroyerDkgPart3Result struct{}

func (_ FfiDestroyerDkgPart3Result) Destroy(value DkgPart3Result) {
	value.Destroy()
}

type DkgRound1Package struct {
	Identifier ParticipantIdentifier
	Data       []byte
}

func (r *DkgRound1Package) Destroy() {
	FfiDestroyerParticipantIdentifier{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterDkgRound1Package struct{}

var FfiConverterDkgRound1PackageINSTANCE = FfiConverterDkgRound1Package{}

func (c FfiConverterDkgRound1Package) Lift(rb RustBufferI) DkgRound1Package {
	return LiftFromRustBuffer[DkgRound1Package](c, rb)
}

func (c FfiConverterDkgRound1Package) Read(reader io.Reader) DkgRound1Package {
	return DkgRound1Package{
		FfiConverterParticipantIdentifierINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterDkgRound1Package) Lower(value DkgRound1Package) C.RustBuffer {
	return LowerIntoRustBuffer[DkgRound1Package](c, value)
}

func (c FfiConverterDkgRound1Package) Write(writer io.Writer, value DkgRound1Package) {
	FfiConverterParticipantIdentifierINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerDkgRound1Package struct{}

func (_ FfiDestroyerDkgRound1Package) Destroy(value DkgRound1Package) {
	value.Destroy()
}

type DkgRound2Package struct {
	Identifier ParticipantIdentifier
	Data       []byte
}

func (r *DkgRound2Package) Destroy() {
	FfiDestroyerParticipantIdentifier{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterDkgRound2Package struct{}

var FfiConverterDkgRound2PackageINSTANCE = FfiConverterDkgRound2Package{}

func (c FfiConverterDkgRound2Package) Lift(rb RustBufferI) DkgRound2Package {
	return LiftFromRustBuffer[DkgRound2Package](c, rb)
}

func (c FfiConverterDkgRound2Package) Read(reader io.Reader) DkgRound2Package {
	return DkgRound2Package{
		FfiConverterParticipantIdentifierINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterDkgRound2Package) Lower(value DkgRound2Package) C.RustBuffer {
	return LowerIntoRustBuffer[DkgRound2Package](c, value)
}

func (c FfiConverterDkgRound2Package) Write(writer io.Writer, value DkgRound2Package) {
	FfiConverterParticipantIdentifierINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerDkgRound2Package struct{}

func (_ FfiDestroyerDkgRound2Package) Destroy(value DkgRound2Package) {
	value.Destroy()
}

type FirstRoundCommitment struct {
	Nonces      FrostSigningNonces
	Commitments FrostSigningCommitments
}

func (r *FirstRoundCommitment) Destroy() {
	FfiDestroyerFrostSigningNonces{}.Destroy(r.Nonces)
	FfiDestroyerFrostSigningCommitments{}.Destroy(r.Commitments)
}

type FfiConverterFirstRoundCommitment struct{}

var FfiConverterFirstRoundCommitmentINSTANCE = FfiConverterFirstRoundCommitment{}

func (c FfiConverterFirstRoundCommitment) Lift(rb RustBufferI) FirstRoundCommitment {
	return LiftFromRustBuffer[FirstRoundCommitment](c, rb)
}

func (c FfiConverterFirstRoundCommitment) Read(reader io.Reader) FirstRoundCommitment {
	return FirstRoundCommitment{
		FfiConverterFrostSigningNoncesINSTANCE.Read(reader),
		FfiConverterFrostSigningCommitmentsINSTANCE.Read(reader),
	}
}

func (c FfiConverterFirstRoundCommitment) Lower(value FirstRoundCommitment) C.RustBuffer {
	return LowerIntoRustBuffer[FirstRoundCommitment](c, value)
}

func (c FfiConverterFirstRoundCommitment) Write(writer io.Writer, value FirstRoundCommitment) {
	FfiConverterFrostSigningNoncesINSTANCE.Write(writer, value.Nonces)
	FfiConverterFrostSigningCommitmentsINSTANCE.Write(writer, value.Commitments)
}

type FfiDestroyerFirstRoundCommitment struct{}

func (_ FfiDestroyerFirstRoundCommitment) Destroy(value FirstRoundCommitment) {
	value.Destroy()
}

type FrostKeyPackage struct {
	Identifier ParticipantIdentifier
	Data       []byte
}

func (r *FrostKeyPackage) Destroy() {
	FfiDestroyerParticipantIdentifier{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterFrostKeyPackage struct{}

var FfiConverterFrostKeyPackageINSTANCE = FfiConverterFrostKeyPackage{}

func (c FfiConverterFrostKeyPackage) Lift(rb RustBufferI) FrostKeyPackage {
	return LiftFromRustBuffer[FrostKeyPackage](c, rb)
}

func (c FfiConverterFrostKeyPackage) Read(reader io.Reader) FrostKeyPackage {
	return FrostKeyPackage{
		FfiConverterParticipantIdentifierINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterFrostKeyPackage) Lower(value FrostKeyPackage) C.RustBuffer {
	return LowerIntoRustBuffer[FrostKeyPackage](c, value)
}

func (c FfiConverterFrostKeyPackage) Write(writer io.Writer, value FrostKeyPackage) {
	FfiConverterParticipantIdentifierINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerFrostKeyPackage struct{}

func (_ FfiDestroyerFrostKeyPackage) Destroy(value FrostKeyPackage) {
	value.Destroy()
}

type FrostPublicKeyPackage struct {
	VerifyingShares map[ParticipantIdentifier]string
	VerifyingKey    string
}

func (r *FrostPublicKeyPackage) Destroy() {
	FfiDestroyerMapParticipantIdentifierString{}.Destroy(r.VerifyingShares)
	FfiDestroyerString{}.Destroy(r.VerifyingKey)
}

type FfiConverterFrostPublicKeyPackage struct{}

var FfiConverterFrostPublicKeyPackageINSTANCE = FfiConverterFrostPublicKeyPackage{}

func (c FfiConverterFrostPublicKeyPackage) Lift(rb RustBufferI) FrostPublicKeyPackage {
	return LiftFromRustBuffer[FrostPublicKeyPackage](c, rb)
}

func (c FfiConverterFrostPublicKeyPackage) Read(reader io.Reader) FrostPublicKeyPackage {
	return FrostPublicKeyPackage{
		FfiConverterMapParticipantIdentifierStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterFrostPublicKeyPackage) Lower(value FrostPublicKeyPackage) C.RustBuffer {
	return LowerIntoRustBuffer[FrostPublicKeyPackage](c, value)
}

func (c FfiConverterFrostPublicKeyPackage) Write(writer io.Writer, value FrostPublicKeyPackage) {
	FfiConverterMapParticipantIdentifierStringINSTANCE.Write(writer, value.VerifyingShares)
	FfiConverterStringINSTANCE.Write(writer, value.VerifyingKey)
}

type FfiDestroyerFrostPublicKeyPackage struct{}

func (_ FfiDestroyerFrostPublicKeyPackage) Destroy(value FrostPublicKeyPackage) {
	value.Destroy()
}

type FrostRandomizer struct {
	Data []byte
}

func (r *FrostRandomizer) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterFrostRandomizer struct{}

var FfiConverterFrostRandomizerINSTANCE = FfiConverterFrostRandomizer{}

func (c FfiConverterFrostRandomizer) Lift(rb RustBufferI) FrostRandomizer {
	return LiftFromRustBuffer[FrostRandomizer](c, rb)
}

func (c FfiConverterFrostRandomizer) Read(reader io.Reader) FrostRandomizer {
	return FrostRandomizer{
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterFrostRandomizer) Lower(value FrostRandomizer) C.RustBuffer {
	return LowerIntoRustBuffer[FrostRandomizer](c, value)
}

func (c FfiConverterFrostRandomizer) Write(writer io.Writer, value FrostRandomizer) {
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerFrostRandomizer struct{}

func (_ FfiDestroyerFrostRandomizer) Destroy(value FrostRandomizer) {
	value.Destroy()
}

type FrostSecretKeyShare struct {
	Identifier ParticipantIdentifier
	Data       []byte
}

func (r *FrostSecretKeyShare) Destroy() {
	FfiDestroyerParticipantIdentifier{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterFrostSecretKeyShare struct{}

var FfiConverterFrostSecretKeyShareINSTANCE = FfiConverterFrostSecretKeyShare{}

func (c FfiConverterFrostSecretKeyShare) Lift(rb RustBufferI) FrostSecretKeyShare {
	return LiftFromRustBuffer[FrostSecretKeyShare](c, rb)
}

func (c FfiConverterFrostSecretKeyShare) Read(reader io.Reader) FrostSecretKeyShare {
	return FrostSecretKeyShare{
		FfiConverterParticipantIdentifierINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterFrostSecretKeyShare) Lower(value FrostSecretKeyShare) C.RustBuffer {
	return LowerIntoRustBuffer[FrostSecretKeyShare](c, value)
}

func (c FfiConverterFrostSecretKeyShare) Write(writer io.Writer, value FrostSecretKeyShare) {
	FfiConverterParticipantIdentifierINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerFrostSecretKeyShare struct{}

func (_ FfiDestroyerFrostSecretKeyShare) Destroy(value FrostSecretKeyShare) {
	value.Destroy()
}

type FrostSignature struct {
	Data []byte
}

func (r *FrostSignature) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterFrostSignature struct{}

var FfiConverterFrostSignatureINSTANCE = FfiConverterFrostSignature{}

func (c FfiConverterFrostSignature) Lift(rb RustBufferI) FrostSignature {
	return LiftFromRustBuffer[FrostSignature](c, rb)
}

func (c FfiConverterFrostSignature) Read(reader io.Reader) FrostSignature {
	return FrostSignature{
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterFrostSignature) Lower(value FrostSignature) C.RustBuffer {
	return LowerIntoRustBuffer[FrostSignature](c, value)
}

func (c FfiConverterFrostSignature) Write(writer io.Writer, value FrostSignature) {
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerFrostSignature struct{}

func (_ FfiDestroyerFrostSignature) Destroy(value FrostSignature) {
	value.Destroy()
}

type FrostSignatureShare struct {
	Identifier ParticipantIdentifier
	Data       []byte
}

func (r *FrostSignatureShare) Destroy() {
	FfiDestroyerParticipantIdentifier{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterFrostSignatureShare struct{}

var FfiConverterFrostSignatureShareINSTANCE = FfiConverterFrostSignatureShare{}

func (c FfiConverterFrostSignatureShare) Lift(rb RustBufferI) FrostSignatureShare {
	return LiftFromRustBuffer[FrostSignatureShare](c, rb)
}

func (c FfiConverterFrostSignatureShare) Read(reader io.Reader) FrostSignatureShare {
	return FrostSignatureShare{
		FfiConverterParticipantIdentifierINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterFrostSignatureShare) Lower(value FrostSignatureShare) C.RustBuffer {
	return LowerIntoRustBuffer[FrostSignatureShare](c, value)
}

func (c FfiConverterFrostSignatureShare) Write(writer io.Writer, value FrostSignatureShare) {
	FfiConverterParticipantIdentifierINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerFrostSignatureShare struct{}

func (_ FfiDestroyerFrostSignatureShare) Destroy(value FrostSignatureShare) {
	value.Destroy()
}

type FrostSigningCommitments struct {
	Identifier ParticipantIdentifier
	Data       []byte
}

func (r *FrostSigningCommitments) Destroy() {
	FfiDestroyerParticipantIdentifier{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterFrostSigningCommitments struct{}

var FfiConverterFrostSigningCommitmentsINSTANCE = FfiConverterFrostSigningCommitments{}

func (c FfiConverterFrostSigningCommitments) Lift(rb RustBufferI) FrostSigningCommitments {
	return LiftFromRustBuffer[FrostSigningCommitments](c, rb)
}

func (c FfiConverterFrostSigningCommitments) Read(reader io.Reader) FrostSigningCommitments {
	return FrostSigningCommitments{
		FfiConverterParticipantIdentifierINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterFrostSigningCommitments) Lower(value FrostSigningCommitments) C.RustBuffer {
	return LowerIntoRustBuffer[FrostSigningCommitments](c, value)
}

func (c FfiConverterFrostSigningCommitments) Write(writer io.Writer, value FrostSigningCommitments) {
	FfiConverterParticipantIdentifierINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerFrostSigningCommitments struct{}

func (_ FfiDestroyerFrostSigningCommitments) Destroy(value FrostSigningCommitments) {
	value.Destroy()
}

type FrostSigningNonces struct {
	Data []byte
}

func (r *FrostSigningNonces) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterFrostSigningNonces struct{}

var FfiConverterFrostSigningNoncesINSTANCE = FfiConverterFrostSigningNonces{}

func (c FfiConverterFrostSigningNonces) Lift(rb RustBufferI) FrostSigningNonces {
	return LiftFromRustBuffer[FrostSigningNonces](c, rb)
}

func (c FfiConverterFrostSigningNonces) Read(reader io.Reader) FrostSigningNonces {
	return FrostSigningNonces{
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterFrostSigningNonces) Lower(value FrostSigningNonces) C.RustBuffer {
	return LowerIntoRustBuffer[FrostSigningNonces](c, value)
}

func (c FfiConverterFrostSigningNonces) Write(writer io.Writer, value FrostSigningNonces) {
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerFrostSigningNonces struct{}

func (_ FfiDestroyerFrostSigningNonces) Destroy(value FrostSigningNonces) {
	value.Destroy()
}

type FrostSigningPackage struct {
	Data []byte
}

func (r *FrostSigningPackage) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterFrostSigningPackage struct{}

var FfiConverterFrostSigningPackageINSTANCE = FfiConverterFrostSigningPackage{}

func (c FfiConverterFrostSigningPackage) Lift(rb RustBufferI) FrostSigningPackage {
	return LiftFromRustBuffer[FrostSigningPackage](c, rb)
}

func (c FfiConverterFrostSigningPackage) Read(reader io.Reader) FrostSigningPackage {
	return FrostSigningPackage{
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterFrostSigningPackage) Lower(value FrostSigningPackage) C.RustBuffer {
	return LowerIntoRustBuffer[FrostSigningPackage](c, value)
}

func (c FfiConverterFrostSigningPackage) Write(writer io.Writer, value FrostSigningPackage) {
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerFrostSigningPackage struct{}

func (_ FfiDestroyerFrostSigningPackage) Destroy(value FrostSigningPackage) {
	value.Destroy()
}

type Message struct {
	Data []byte
}

func (r *Message) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterMessage struct{}

var FfiConverterMessageINSTANCE = FfiConverterMessage{}

func (c FfiConverterMessage) Lift(rb RustBufferI) Message {
	return LiftFromRustBuffer[Message](c, rb)
}

func (c FfiConverterMessage) Read(reader io.Reader) Message {
	return Message{
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterMessage) Lower(value Message) C.RustBuffer {
	return LowerIntoRustBuffer[Message](c, value)
}

func (c FfiConverterMessage) Write(writer io.Writer, value Message) {
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerMessage struct{}

func (_ FfiDestroyerMessage) Destroy(value Message) {
	value.Destroy()
}

type ParticipantIdentifier struct {
	Data string
}

func (r *ParticipantIdentifier) Destroy() {
	FfiDestroyerString{}.Destroy(r.Data)
}

type FfiConverterParticipantIdentifier struct{}

var FfiConverterParticipantIdentifierINSTANCE = FfiConverterParticipantIdentifier{}

func (c FfiConverterParticipantIdentifier) Lift(rb RustBufferI) ParticipantIdentifier {
	return LiftFromRustBuffer[ParticipantIdentifier](c, rb)
}

func (c FfiConverterParticipantIdentifier) Read(reader io.Reader) ParticipantIdentifier {
	return ParticipantIdentifier{
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterParticipantIdentifier) Lower(value ParticipantIdentifier) C.RustBuffer {
	return LowerIntoRustBuffer[ParticipantIdentifier](c, value)
}

func (c FfiConverterParticipantIdentifier) Write(writer io.Writer, value ParticipantIdentifier) {
	FfiConverterStringINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerParticipantIdentifier struct{}

func (_ FfiDestroyerParticipantIdentifier) Destroy(value ParticipantIdentifier) {
	value.Destroy()
}

type ParticipantList struct {
	Identifiers []ParticipantIdentifier
}

func (r *ParticipantList) Destroy() {
	FfiDestroyerSequenceParticipantIdentifier{}.Destroy(r.Identifiers)
}

type FfiConverterParticipantList struct{}

var FfiConverterParticipantListINSTANCE = FfiConverterParticipantList{}

func (c FfiConverterParticipantList) Lift(rb RustBufferI) ParticipantList {
	return LiftFromRustBuffer[ParticipantList](c, rb)
}

func (c FfiConverterParticipantList) Read(reader io.Reader) ParticipantList {
	return ParticipantList{
		FfiConverterSequenceParticipantIdentifierINSTANCE.Read(reader),
	}
}

func (c FfiConverterParticipantList) Lower(value ParticipantList) C.RustBuffer {
	return LowerIntoRustBuffer[ParticipantList](c, value)
}

func (c FfiConverterParticipantList) Write(writer io.Writer, value ParticipantList) {
	FfiConverterSequenceParticipantIdentifierINSTANCE.Write(writer, value.Identifiers)
}

type FfiDestroyerParticipantList struct{}

func (_ FfiDestroyerParticipantList) Destroy(value ParticipantList) {
	value.Destroy()
}

type TrustedKeyGeneration struct {
	SecretShares     map[ParticipantIdentifier]FrostSecretKeyShare
	PublicKeyPackage FrostPublicKeyPackage
}

func (r *TrustedKeyGeneration) Destroy() {
	FfiDestroyerMapParticipantIdentifierFrostSecretKeyShare{}.Destroy(r.SecretShares)
	FfiDestroyerFrostPublicKeyPackage{}.Destroy(r.PublicKeyPackage)
}

type FfiConverterTrustedKeyGeneration struct{}

var FfiConverterTrustedKeyGenerationINSTANCE = FfiConverterTrustedKeyGeneration{}

func (c FfiConverterTrustedKeyGeneration) Lift(rb RustBufferI) TrustedKeyGeneration {
	return LiftFromRustBuffer[TrustedKeyGeneration](c, rb)
}

func (c FfiConverterTrustedKeyGeneration) Read(reader io.Reader) TrustedKeyGeneration {
	return TrustedKeyGeneration{
		FfiConverterMapParticipantIdentifierFrostSecretKeyShareINSTANCE.Read(reader),
		FfiConverterFrostPublicKeyPackageINSTANCE.Read(reader),
	}
}

func (c FfiConverterTrustedKeyGeneration) Lower(value TrustedKeyGeneration) C.RustBuffer {
	return LowerIntoRustBuffer[TrustedKeyGeneration](c, value)
}

func (c FfiConverterTrustedKeyGeneration) Write(writer io.Writer, value TrustedKeyGeneration) {
	FfiConverterMapParticipantIdentifierFrostSecretKeyShareINSTANCE.Write(writer, value.SecretShares)
	FfiConverterFrostPublicKeyPackageINSTANCE.Write(writer, value.PublicKeyPackage)
}

type FfiDestroyerTrustedKeyGeneration struct{}

func (_ FfiDestroyerTrustedKeyGeneration) Destroy(value TrustedKeyGeneration) {
	value.Destroy()
}

type ConfigurationError struct {
	err error
}

// Convience method to turn *ConfigurationError into error
// Avoiding treating nil pointer as non nil error interface
func (err *ConfigurationError) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err ConfigurationError) Error() string {
	return fmt.Sprintf("ConfigurationError: %s", err.err.Error())
}

func (err ConfigurationError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrConfigurationErrorInvalidMaxSigners = fmt.Errorf("ConfigurationErrorInvalidMaxSigners")
var ErrConfigurationErrorInvalidMinSigners = fmt.Errorf("ConfigurationErrorInvalidMinSigners")
var ErrConfigurationErrorInvalidIdentifier = fmt.Errorf("ConfigurationErrorInvalidIdentifier")
var ErrConfigurationErrorUnknownError = fmt.Errorf("ConfigurationErrorUnknownError")

// Variant structs
type ConfigurationErrorInvalidMaxSigners struct {
}

func NewConfigurationErrorInvalidMaxSigners() *ConfigurationError {
	return &ConfigurationError{err: &ConfigurationErrorInvalidMaxSigners{}}
}

func (e ConfigurationErrorInvalidMaxSigners) destroy() {
}

func (err ConfigurationErrorInvalidMaxSigners) Error() string {
	return fmt.Sprint("InvalidMaxSigners")
}

func (self ConfigurationErrorInvalidMaxSigners) Is(target error) bool {
	return target == ErrConfigurationErrorInvalidMaxSigners
}

type ConfigurationErrorInvalidMinSigners struct {
}

func NewConfigurationErrorInvalidMinSigners() *ConfigurationError {
	return &ConfigurationError{err: &ConfigurationErrorInvalidMinSigners{}}
}

func (e ConfigurationErrorInvalidMinSigners) destroy() {
}

func (err ConfigurationErrorInvalidMinSigners) Error() string {
	return fmt.Sprint("InvalidMinSigners")
}

func (self ConfigurationErrorInvalidMinSigners) Is(target error) bool {
	return target == ErrConfigurationErrorInvalidMinSigners
}

type ConfigurationErrorInvalidIdentifier struct {
}

func NewConfigurationErrorInvalidIdentifier() *ConfigurationError {
	return &ConfigurationError{err: &ConfigurationErrorInvalidIdentifier{}}
}

func (e ConfigurationErrorInvalidIdentifier) destroy() {
}

func (err ConfigurationErrorInvalidIdentifier) Error() string {
	return fmt.Sprint("InvalidIdentifier")
}

func (self ConfigurationErrorInvalidIdentifier) Is(target error) bool {
	return target == ErrConfigurationErrorInvalidIdentifier
}

type ConfigurationErrorUnknownError struct {
}

func NewConfigurationErrorUnknownError() *ConfigurationError {
	return &ConfigurationError{err: &ConfigurationErrorUnknownError{}}
}

func (e ConfigurationErrorUnknownError) destroy() {
}

func (err ConfigurationErrorUnknownError) Error() string {
	return fmt.Sprint("UnknownError")
}

func (self ConfigurationErrorUnknownError) Is(target error) bool {
	return target == ErrConfigurationErrorUnknownError
}

type FfiConverterConfigurationError struct{}

var FfiConverterConfigurationErrorINSTANCE = FfiConverterConfigurationError{}

func (c FfiConverterConfigurationError) Lift(eb RustBufferI) *ConfigurationError {
	return LiftFromRustBuffer[*ConfigurationError](c, eb)
}

func (c FfiConverterConfigurationError) Lower(value *ConfigurationError) C.RustBuffer {
	return LowerIntoRustBuffer[*ConfigurationError](c, value)
}

func (c FfiConverterConfigurationError) Read(reader io.Reader) *ConfigurationError {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &ConfigurationError{&ConfigurationErrorInvalidMaxSigners{}}
	case 2:
		return &ConfigurationError{&ConfigurationErrorInvalidMinSigners{}}
	case 3:
		return &ConfigurationError{&ConfigurationErrorInvalidIdentifier{}}
	case 4:
		return &ConfigurationError{&ConfigurationErrorUnknownError{}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterConfigurationError.Read()", errorID))
	}
}

func (c FfiConverterConfigurationError) Write(writer io.Writer, value *ConfigurationError) {
	switch variantValue := value.err.(type) {
	case *ConfigurationErrorInvalidMaxSigners:
		writeInt32(writer, 1)
	case *ConfigurationErrorInvalidMinSigners:
		writeInt32(writer, 2)
	case *ConfigurationErrorInvalidIdentifier:
		writeInt32(writer, 3)
	case *ConfigurationErrorUnknownError:
		writeInt32(writer, 4)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterConfigurationError.Write", value))
	}
}

type FfiDestroyerConfigurationError struct{}

func (_ FfiDestroyerConfigurationError) Destroy(value *ConfigurationError) {
	switch variantValue := value.err.(type) {
	case ConfigurationErrorInvalidMaxSigners:
		variantValue.destroy()
	case ConfigurationErrorInvalidMinSigners:
		variantValue.destroy()
	case ConfigurationErrorInvalidIdentifier:
		variantValue.destroy()
	case ConfigurationErrorUnknownError:
		variantValue.destroy()
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiDestroyerConfigurationError.Destroy", value))
	}
}

type CoordinationError struct {
	err error
}

// Convience method to turn *CoordinationError into error
// Avoiding treating nil pointer as non nil error interface
func (err *CoordinationError) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err CoordinationError) Error() string {
	return fmt.Sprintf("CoordinationError: %s", err.err.Error())
}

func (err CoordinationError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrCoordinationErrorFailedToCreateSigningPackage = fmt.Errorf("CoordinationErrorFailedToCreateSigningPackage")
var ErrCoordinationErrorInvalidSigningCommitment = fmt.Errorf("CoordinationErrorInvalidSigningCommitment")
var ErrCoordinationErrorIdentifierDeserializationError = fmt.Errorf("CoordinationErrorIdentifierDeserializationError")
var ErrCoordinationErrorSigningPackageSerializationError = fmt.Errorf("CoordinationErrorSigningPackageSerializationError")
var ErrCoordinationErrorSignatureShareDeserializationError = fmt.Errorf("CoordinationErrorSignatureShareDeserializationError")
var ErrCoordinationErrorPublicKeyPackageDeserializationError = fmt.Errorf("CoordinationErrorPublicKeyPackageDeserializationError")
var ErrCoordinationErrorSignatureShareAggregationFailed = fmt.Errorf("CoordinationErrorSignatureShareAggregationFailed")
var ErrCoordinationErrorInvalidRandomizer = fmt.Errorf("CoordinationErrorInvalidRandomizer")

// Variant structs
type CoordinationErrorFailedToCreateSigningPackage struct {
}

func NewCoordinationErrorFailedToCreateSigningPackage() *CoordinationError {
	return &CoordinationError{err: &CoordinationErrorFailedToCreateSigningPackage{}}
}

func (e CoordinationErrorFailedToCreateSigningPackage) destroy() {
}

func (err CoordinationErrorFailedToCreateSigningPackage) Error() string {
	return fmt.Sprint("FailedToCreateSigningPackage")
}

func (self CoordinationErrorFailedToCreateSigningPackage) Is(target error) bool {
	return target == ErrCoordinationErrorFailedToCreateSigningPackage
}

type CoordinationErrorInvalidSigningCommitment struct {
}

func NewCoordinationErrorInvalidSigningCommitment() *CoordinationError {
	return &CoordinationError{err: &CoordinationErrorInvalidSigningCommitment{}}
}

func (e CoordinationErrorInvalidSigningCommitment) destroy() {
}

func (err CoordinationErrorInvalidSigningCommitment) Error() string {
	return fmt.Sprint("InvalidSigningCommitment")
}

func (self CoordinationErrorInvalidSigningCommitment) Is(target error) bool {
	return target == ErrCoordinationErrorInvalidSigningCommitment
}

type CoordinationErrorIdentifierDeserializationError struct {
}

func NewCoordinationErrorIdentifierDeserializationError() *CoordinationError {
	return &CoordinationError{err: &CoordinationErrorIdentifierDeserializationError{}}
}

func (e CoordinationErrorIdentifierDeserializationError) destroy() {
}

func (err CoordinationErrorIdentifierDeserializationError) Error() string {
	return fmt.Sprint("IdentifierDeserializationError")
}

func (self CoordinationErrorIdentifierDeserializationError) Is(target error) bool {
	return target == ErrCoordinationErrorIdentifierDeserializationError
}

type CoordinationErrorSigningPackageSerializationError struct {
}

func NewCoordinationErrorSigningPackageSerializationError() *CoordinationError {
	return &CoordinationError{err: &CoordinationErrorSigningPackageSerializationError{}}
}

func (e CoordinationErrorSigningPackageSerializationError) destroy() {
}

func (err CoordinationErrorSigningPackageSerializationError) Error() string {
	return fmt.Sprint("SigningPackageSerializationError")
}

func (self CoordinationErrorSigningPackageSerializationError) Is(target error) bool {
	return target == ErrCoordinationErrorSigningPackageSerializationError
}

type CoordinationErrorSignatureShareDeserializationError struct {
}

func NewCoordinationErrorSignatureShareDeserializationError() *CoordinationError {
	return &CoordinationError{err: &CoordinationErrorSignatureShareDeserializationError{}}
}

func (e CoordinationErrorSignatureShareDeserializationError) destroy() {
}

func (err CoordinationErrorSignatureShareDeserializationError) Error() string {
	return fmt.Sprint("SignatureShareDeserializationError")
}

func (self CoordinationErrorSignatureShareDeserializationError) Is(target error) bool {
	return target == ErrCoordinationErrorSignatureShareDeserializationError
}

type CoordinationErrorPublicKeyPackageDeserializationError struct {
}

func NewCoordinationErrorPublicKeyPackageDeserializationError() *CoordinationError {
	return &CoordinationError{err: &CoordinationErrorPublicKeyPackageDeserializationError{}}
}

func (e CoordinationErrorPublicKeyPackageDeserializationError) destroy() {
}

func (err CoordinationErrorPublicKeyPackageDeserializationError) Error() string {
	return fmt.Sprint("PublicKeyPackageDeserializationError")
}

func (self CoordinationErrorPublicKeyPackageDeserializationError) Is(target error) bool {
	return target == ErrCoordinationErrorPublicKeyPackageDeserializationError
}

type CoordinationErrorSignatureShareAggregationFailed struct {
	Message string
}

func NewCoordinationErrorSignatureShareAggregationFailed(
	message string,
) *CoordinationError {
	return &CoordinationError{err: &CoordinationErrorSignatureShareAggregationFailed{
		Message: message}}
}

func (e CoordinationErrorSignatureShareAggregationFailed) destroy() {
	FfiDestroyerString{}.Destroy(e.Message)
}

func (err CoordinationErrorSignatureShareAggregationFailed) Error() string {
	return fmt.Sprint("SignatureShareAggregationFailed",
		": ",

		"Message=",
		err.Message,
	)
}

func (self CoordinationErrorSignatureShareAggregationFailed) Is(target error) bool {
	return target == ErrCoordinationErrorSignatureShareAggregationFailed
}

type CoordinationErrorInvalidRandomizer struct {
}

func NewCoordinationErrorInvalidRandomizer() *CoordinationError {
	return &CoordinationError{err: &CoordinationErrorInvalidRandomizer{}}
}

func (e CoordinationErrorInvalidRandomizer) destroy() {
}

func (err CoordinationErrorInvalidRandomizer) Error() string {
	return fmt.Sprint("InvalidRandomizer")
}

func (self CoordinationErrorInvalidRandomizer) Is(target error) bool {
	return target == ErrCoordinationErrorInvalidRandomizer
}

type FfiConverterCoordinationError struct{}

var FfiConverterCoordinationErrorINSTANCE = FfiConverterCoordinationError{}

func (c FfiConverterCoordinationError) Lift(eb RustBufferI) *CoordinationError {
	return LiftFromRustBuffer[*CoordinationError](c, eb)
}

func (c FfiConverterCoordinationError) Lower(value *CoordinationError) C.RustBuffer {
	return LowerIntoRustBuffer[*CoordinationError](c, value)
}

func (c FfiConverterCoordinationError) Read(reader io.Reader) *CoordinationError {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &CoordinationError{&CoordinationErrorFailedToCreateSigningPackage{}}
	case 2:
		return &CoordinationError{&CoordinationErrorInvalidSigningCommitment{}}
	case 3:
		return &CoordinationError{&CoordinationErrorIdentifierDeserializationError{}}
	case 4:
		return &CoordinationError{&CoordinationErrorSigningPackageSerializationError{}}
	case 5:
		return &CoordinationError{&CoordinationErrorSignatureShareDeserializationError{}}
	case 6:
		return &CoordinationError{&CoordinationErrorPublicKeyPackageDeserializationError{}}
	case 7:
		return &CoordinationError{&CoordinationErrorSignatureShareAggregationFailed{
			Message: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 8:
		return &CoordinationError{&CoordinationErrorInvalidRandomizer{}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterCoordinationError.Read()", errorID))
	}
}

func (c FfiConverterCoordinationError) Write(writer io.Writer, value *CoordinationError) {
	switch variantValue := value.err.(type) {
	case *CoordinationErrorFailedToCreateSigningPackage:
		writeInt32(writer, 1)
	case *CoordinationErrorInvalidSigningCommitment:
		writeInt32(writer, 2)
	case *CoordinationErrorIdentifierDeserializationError:
		writeInt32(writer, 3)
	case *CoordinationErrorSigningPackageSerializationError:
		writeInt32(writer, 4)
	case *CoordinationErrorSignatureShareDeserializationError:
		writeInt32(writer, 5)
	case *CoordinationErrorPublicKeyPackageDeserializationError:
		writeInt32(writer, 6)
	case *CoordinationErrorSignatureShareAggregationFailed:
		writeInt32(writer, 7)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Message)
	case *CoordinationErrorInvalidRandomizer:
		writeInt32(writer, 8)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterCoordinationError.Write", value))
	}
}

type FfiDestroyerCoordinationError struct{}

func (_ FfiDestroyerCoordinationError) Destroy(value *CoordinationError) {
	switch variantValue := value.err.(type) {
	case CoordinationErrorFailedToCreateSigningPackage:
		variantValue.destroy()
	case CoordinationErrorInvalidSigningCommitment:
		variantValue.destroy()
	case CoordinationErrorIdentifierDeserializationError:
		variantValue.destroy()
	case CoordinationErrorSigningPackageSerializationError:
		variantValue.destroy()
	case CoordinationErrorSignatureShareDeserializationError:
		variantValue.destroy()
	case CoordinationErrorPublicKeyPackageDeserializationError:
		variantValue.destroy()
	case CoordinationErrorSignatureShareAggregationFailed:
		variantValue.destroy()
	case CoordinationErrorInvalidRandomizer:
		variantValue.destroy()
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiDestroyerCoordinationError.Destroy", value))
	}
}

type FrostError struct {
	err error
}

// Convience method to turn *FrostError into error
// Avoiding treating nil pointer as non nil error interface
func (err *FrostError) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err FrostError) Error() string {
	return fmt.Sprintf("FrostError: %s", err.err.Error())
}

func (err FrostError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrFrostErrorInvalidMinSigners = fmt.Errorf("FrostErrorInvalidMinSigners")
var ErrFrostErrorInvalidMaxSigners = fmt.Errorf("FrostErrorInvalidMaxSigners")
var ErrFrostErrorInvalidCoefficients = fmt.Errorf("FrostErrorInvalidCoefficients")
var ErrFrostErrorMalformedIdentifier = fmt.Errorf("FrostErrorMalformedIdentifier")
var ErrFrostErrorDuplicatedIdentifier = fmt.Errorf("FrostErrorDuplicatedIdentifier")
var ErrFrostErrorUnknownIdentifier = fmt.Errorf("FrostErrorUnknownIdentifier")
var ErrFrostErrorIncorrectNumberOfIdentifiers = fmt.Errorf("FrostErrorIncorrectNumberOfIdentifiers")
var ErrFrostErrorMalformedSigningKey = fmt.Errorf("FrostErrorMalformedSigningKey")
var ErrFrostErrorMalformedVerifyingKey = fmt.Errorf("FrostErrorMalformedVerifyingKey")
var ErrFrostErrorMalformedSignature = fmt.Errorf("FrostErrorMalformedSignature")
var ErrFrostErrorInvalidSignature = fmt.Errorf("FrostErrorInvalidSignature")
var ErrFrostErrorDuplicatedShares = fmt.Errorf("FrostErrorDuplicatedShares")
var ErrFrostErrorIncorrectNumberOfShares = fmt.Errorf("FrostErrorIncorrectNumberOfShares")
var ErrFrostErrorIdentityCommitment = fmt.Errorf("FrostErrorIdentityCommitment")
var ErrFrostErrorMissingCommitment = fmt.Errorf("FrostErrorMissingCommitment")
var ErrFrostErrorIncorrectCommitment = fmt.Errorf("FrostErrorIncorrectCommitment")
var ErrFrostErrorIncorrectNumberOfCommitments = fmt.Errorf("FrostErrorIncorrectNumberOfCommitments")
var ErrFrostErrorInvalidSignatureShare = fmt.Errorf("FrostErrorInvalidSignatureShare")
var ErrFrostErrorInvalidSecretShare = fmt.Errorf("FrostErrorInvalidSecretShare")
var ErrFrostErrorPackageNotFound = fmt.Errorf("FrostErrorPackageNotFound")
var ErrFrostErrorIncorrectNumberOfPackages = fmt.Errorf("FrostErrorIncorrectNumberOfPackages")
var ErrFrostErrorIncorrectPackage = fmt.Errorf("FrostErrorIncorrectPackage")
var ErrFrostErrorDkgNotSupported = fmt.Errorf("FrostErrorDkgNotSupported")
var ErrFrostErrorInvalidProofOfKnowledge = fmt.Errorf("FrostErrorInvalidProofOfKnowledge")
var ErrFrostErrorFieldError = fmt.Errorf("FrostErrorFieldError")
var ErrFrostErrorGroupError = fmt.Errorf("FrostErrorGroupError")
var ErrFrostErrorInvalidCoefficient = fmt.Errorf("FrostErrorInvalidCoefficient")
var ErrFrostErrorIdentifierDerivationNotSupported = fmt.Errorf("FrostErrorIdentifierDerivationNotSupported")
var ErrFrostErrorSerializationError = fmt.Errorf("FrostErrorSerializationError")
var ErrFrostErrorDeserializationError = fmt.Errorf("FrostErrorDeserializationError")
var ErrFrostErrorDkgPart2IncorrectNumberOfCommitments = fmt.Errorf("FrostErrorDkgPart2IncorrectNumberOfCommitments")
var ErrFrostErrorDkgPart2IncorrectNumberOfPackages = fmt.Errorf("FrostErrorDkgPart2IncorrectNumberOfPackages")
var ErrFrostErrorDkgPart3IncorrectRound1Packages = fmt.Errorf("FrostErrorDkgPart3IncorrectRound1Packages")
var ErrFrostErrorDkgPart3IncorrectNumberOfPackages = fmt.Errorf("FrostErrorDkgPart3IncorrectNumberOfPackages")
var ErrFrostErrorDkgPart3PackageSendersMismatch = fmt.Errorf("FrostErrorDkgPart3PackageSendersMismatch")
var ErrFrostErrorInvalidKeyPackage = fmt.Errorf("FrostErrorInvalidKeyPackage")
var ErrFrostErrorInvalidSecretKey = fmt.Errorf("FrostErrorInvalidSecretKey")
var ErrFrostErrorInvalidConfiguration = fmt.Errorf("FrostErrorInvalidConfiguration")
var ErrFrostErrorUnexpectedError = fmt.Errorf("FrostErrorUnexpectedError")

// Variant structs
// min_signers is invalid
type FrostErrorInvalidMinSigners struct {
}

// min_signers is invalid
func NewFrostErrorInvalidMinSigners() *FrostError {
	return &FrostError{err: &FrostErrorInvalidMinSigners{}}
}

func (e FrostErrorInvalidMinSigners) destroy() {
}

func (err FrostErrorInvalidMinSigners) Error() string {
	return fmt.Sprint("InvalidMinSigners")
}

func (self FrostErrorInvalidMinSigners) Is(target error) bool {
	return target == ErrFrostErrorInvalidMinSigners
}

// max_signers is invalid
type FrostErrorInvalidMaxSigners struct {
}

// max_signers is invalid
func NewFrostErrorInvalidMaxSigners() *FrostError {
	return &FrostError{err: &FrostErrorInvalidMaxSigners{}}
}

func (e FrostErrorInvalidMaxSigners) destroy() {
}

func (err FrostErrorInvalidMaxSigners) Error() string {
	return fmt.Sprint("InvalidMaxSigners")
}

func (self FrostErrorInvalidMaxSigners) Is(target error) bool {
	return target == ErrFrostErrorInvalidMaxSigners
}

// max_signers is invalid
type FrostErrorInvalidCoefficients struct {
}

// max_signers is invalid
func NewFrostErrorInvalidCoefficients() *FrostError {
	return &FrostError{err: &FrostErrorInvalidCoefficients{}}
}

func (e FrostErrorInvalidCoefficients) destroy() {
}

func (err FrostErrorInvalidCoefficients) Error() string {
	return fmt.Sprint("InvalidCoefficients")
}

func (self FrostErrorInvalidCoefficients) Is(target error) bool {
	return target == ErrFrostErrorInvalidCoefficients
}

// This identifier is unserializable.
type FrostErrorMalformedIdentifier struct {
}

// This identifier is unserializable.
func NewFrostErrorMalformedIdentifier() *FrostError {
	return &FrostError{err: &FrostErrorMalformedIdentifier{}}
}

func (e FrostErrorMalformedIdentifier) destroy() {
}

func (err FrostErrorMalformedIdentifier) Error() string {
	return fmt.Sprint("MalformedIdentifier")
}

func (self FrostErrorMalformedIdentifier) Is(target error) bool {
	return target == ErrFrostErrorMalformedIdentifier
}

// This identifier is duplicated.
type FrostErrorDuplicatedIdentifier struct {
}

// This identifier is duplicated.
func NewFrostErrorDuplicatedIdentifier() *FrostError {
	return &FrostError{err: &FrostErrorDuplicatedIdentifier{}}
}

func (e FrostErrorDuplicatedIdentifier) destroy() {
}

func (err FrostErrorDuplicatedIdentifier) Error() string {
	return fmt.Sprint("DuplicatedIdentifier")
}

func (self FrostErrorDuplicatedIdentifier) Is(target error) bool {
	return target == ErrFrostErrorDuplicatedIdentifier
}

// This identifier does not belong to a participant in the signing process.
type FrostErrorUnknownIdentifier struct {
}

// This identifier does not belong to a participant in the signing process.
func NewFrostErrorUnknownIdentifier() *FrostError {
	return &FrostError{err: &FrostErrorUnknownIdentifier{}}
}

func (e FrostErrorUnknownIdentifier) destroy() {
}

func (err FrostErrorUnknownIdentifier) Error() string {
	return fmt.Sprint("UnknownIdentifier")
}

func (self FrostErrorUnknownIdentifier) Is(target error) bool {
	return target == ErrFrostErrorUnknownIdentifier
}

// Incorrect number of identifiers.
type FrostErrorIncorrectNumberOfIdentifiers struct {
}

// Incorrect number of identifiers.
func NewFrostErrorIncorrectNumberOfIdentifiers() *FrostError {
	return &FrostError{err: &FrostErrorIncorrectNumberOfIdentifiers{}}
}

func (e FrostErrorIncorrectNumberOfIdentifiers) destroy() {
}

func (err FrostErrorIncorrectNumberOfIdentifiers) Error() string {
	return fmt.Sprint("IncorrectNumberOfIdentifiers")
}

func (self FrostErrorIncorrectNumberOfIdentifiers) Is(target error) bool {
	return target == ErrFrostErrorIncorrectNumberOfIdentifiers
}

// The encoding of a signing key was malformed.
type FrostErrorMalformedSigningKey struct {
}

// The encoding of a signing key was malformed.
func NewFrostErrorMalformedSigningKey() *FrostError {
	return &FrostError{err: &FrostErrorMalformedSigningKey{}}
}

func (e FrostErrorMalformedSigningKey) destroy() {
}

func (err FrostErrorMalformedSigningKey) Error() string {
	return fmt.Sprint("MalformedSigningKey")
}

func (self FrostErrorMalformedSigningKey) Is(target error) bool {
	return target == ErrFrostErrorMalformedSigningKey
}

// The encoding of a verifying key was malformed.
type FrostErrorMalformedVerifyingKey struct {
}

// The encoding of a verifying key was malformed.
func NewFrostErrorMalformedVerifyingKey() *FrostError {
	return &FrostError{err: &FrostErrorMalformedVerifyingKey{}}
}

func (e FrostErrorMalformedVerifyingKey) destroy() {
}

func (err FrostErrorMalformedVerifyingKey) Error() string {
	return fmt.Sprint("MalformedVerifyingKey")
}

func (self FrostErrorMalformedVerifyingKey) Is(target error) bool {
	return target == ErrFrostErrorMalformedVerifyingKey
}

// The encoding of a signature was malformed.
type FrostErrorMalformedSignature struct {
}

// The encoding of a signature was malformed.
func NewFrostErrorMalformedSignature() *FrostError {
	return &FrostError{err: &FrostErrorMalformedSignature{}}
}

func (e FrostErrorMalformedSignature) destroy() {
}

func (err FrostErrorMalformedSignature) Error() string {
	return fmt.Sprint("MalformedSignature")
}

func (self FrostErrorMalformedSignature) Is(target error) bool {
	return target == ErrFrostErrorMalformedSignature
}

// Signature verification failed.
type FrostErrorInvalidSignature struct {
}

// Signature verification failed.
func NewFrostErrorInvalidSignature() *FrostError {
	return &FrostError{err: &FrostErrorInvalidSignature{}}
}

func (e FrostErrorInvalidSignature) destroy() {
}

func (err FrostErrorInvalidSignature) Error() string {
	return fmt.Sprint("InvalidSignature")
}

func (self FrostErrorInvalidSignature) Is(target error) bool {
	return target == ErrFrostErrorInvalidSignature
}

// Duplicated shares provided
type FrostErrorDuplicatedShares struct {
}

// Duplicated shares provided
func NewFrostErrorDuplicatedShares() *FrostError {
	return &FrostError{err: &FrostErrorDuplicatedShares{}}
}

func (e FrostErrorDuplicatedShares) destroy() {
}

func (err FrostErrorDuplicatedShares) Error() string {
	return fmt.Sprint("DuplicatedShares")
}

func (self FrostErrorDuplicatedShares) Is(target error) bool {
	return target == ErrFrostErrorDuplicatedShares
}

// Incorrect number of shares.
type FrostErrorIncorrectNumberOfShares struct {
}

// Incorrect number of shares.
func NewFrostErrorIncorrectNumberOfShares() *FrostError {
	return &FrostError{err: &FrostErrorIncorrectNumberOfShares{}}
}

func (e FrostErrorIncorrectNumberOfShares) destroy() {
}

func (err FrostErrorIncorrectNumberOfShares) Error() string {
	return fmt.Sprint("IncorrectNumberOfShares")
}

func (self FrostErrorIncorrectNumberOfShares) Is(target error) bool {
	return target == ErrFrostErrorIncorrectNumberOfShares
}

// Commitment equals the identity
type FrostErrorIdentityCommitment struct {
}

// Commitment equals the identity
func NewFrostErrorIdentityCommitment() *FrostError {
	return &FrostError{err: &FrostErrorIdentityCommitment{}}
}

func (e FrostErrorIdentityCommitment) destroy() {
}

func (err FrostErrorIdentityCommitment) Error() string {
	return fmt.Sprint("IdentityCommitment")
}

func (self FrostErrorIdentityCommitment) Is(target error) bool {
	return target == ErrFrostErrorIdentityCommitment
}

// The participant's commitment is missing from the Signing Package
type FrostErrorMissingCommitment struct {
}

// The participant's commitment is missing from the Signing Package
func NewFrostErrorMissingCommitment() *FrostError {
	return &FrostError{err: &FrostErrorMissingCommitment{}}
}

func (e FrostErrorMissingCommitment) destroy() {
}

func (err FrostErrorMissingCommitment) Error() string {
	return fmt.Sprint("MissingCommitment")
}

func (self FrostErrorMissingCommitment) Is(target error) bool {
	return target == ErrFrostErrorMissingCommitment
}

// The participant's commitment is incorrect
type FrostErrorIncorrectCommitment struct {
}

// The participant's commitment is incorrect
func NewFrostErrorIncorrectCommitment() *FrostError {
	return &FrostError{err: &FrostErrorIncorrectCommitment{}}
}

func (e FrostErrorIncorrectCommitment) destroy() {
}

func (err FrostErrorIncorrectCommitment) Error() string {
	return fmt.Sprint("IncorrectCommitment")
}

func (self FrostErrorIncorrectCommitment) Is(target error) bool {
	return target == ErrFrostErrorIncorrectCommitment
}

// Incorrect number of commitments.
type FrostErrorIncorrectNumberOfCommitments struct {
}

// Incorrect number of commitments.
func NewFrostErrorIncorrectNumberOfCommitments() *FrostError {
	return &FrostError{err: &FrostErrorIncorrectNumberOfCommitments{}}
}

func (e FrostErrorIncorrectNumberOfCommitments) destroy() {
}

func (err FrostErrorIncorrectNumberOfCommitments) Error() string {
	return fmt.Sprint("IncorrectNumberOfCommitments")
}

func (self FrostErrorIncorrectNumberOfCommitments) Is(target error) bool {
	return target == ErrFrostErrorIncorrectNumberOfCommitments
}

type FrostErrorInvalidSignatureShare struct {
	Culprit ParticipantIdentifier
}

func NewFrostErrorInvalidSignatureShare(
	culprit ParticipantIdentifier,
) *FrostError {
	return &FrostError{err: &FrostErrorInvalidSignatureShare{
		Culprit: culprit}}
}

func (e FrostErrorInvalidSignatureShare) destroy() {
	FfiDestroyerParticipantIdentifier{}.Destroy(e.Culprit)
}

func (err FrostErrorInvalidSignatureShare) Error() string {
	return fmt.Sprint("InvalidSignatureShare",
		": ",

		"Culprit=",
		err.Culprit,
	)
}

func (self FrostErrorInvalidSignatureShare) Is(target error) bool {
	return target == ErrFrostErrorInvalidSignatureShare
}

// Secret share verification failed.
type FrostErrorInvalidSecretShare struct {
	Culprit *ParticipantIdentifier
}

// Secret share verification failed.
func NewFrostErrorInvalidSecretShare(
	culprit *ParticipantIdentifier,
) *FrostError {
	return &FrostError{err: &FrostErrorInvalidSecretShare{
		Culprit: culprit}}
}

func (e FrostErrorInvalidSecretShare) destroy() {
	FfiDestroyerOptionalParticipantIdentifier{}.Destroy(e.Culprit)
}

func (err FrostErrorInvalidSecretShare) Error() string {
	return fmt.Sprint("InvalidSecretShare",
		": ",

		"Culprit=",
		err.Culprit,
	)
}

func (self FrostErrorInvalidSecretShare) Is(target error) bool {
	return target == ErrFrostErrorInvalidSecretShare
}

// Round 1 package not found for Round 2 participant.
type FrostErrorPackageNotFound struct {
}

// Round 1 package not found for Round 2 participant.
func NewFrostErrorPackageNotFound() *FrostError {
	return &FrostError{err: &FrostErrorPackageNotFound{}}
}

func (e FrostErrorPackageNotFound) destroy() {
}

func (err FrostErrorPackageNotFound) Error() string {
	return fmt.Sprint("PackageNotFound")
}

func (self FrostErrorPackageNotFound) Is(target error) bool {
	return target == ErrFrostErrorPackageNotFound
}

// Incorrect number of packages.
type FrostErrorIncorrectNumberOfPackages struct {
}

// Incorrect number of packages.
func NewFrostErrorIncorrectNumberOfPackages() *FrostError {
	return &FrostError{err: &FrostErrorIncorrectNumberOfPackages{}}
}

func (e FrostErrorIncorrectNumberOfPackages) destroy() {
}

func (err FrostErrorIncorrectNumberOfPackages) Error() string {
	return fmt.Sprint("IncorrectNumberOfPackages")
}

func (self FrostErrorIncorrectNumberOfPackages) Is(target error) bool {
	return target == ErrFrostErrorIncorrectNumberOfPackages
}

// The incorrect package was specified.
type FrostErrorIncorrectPackage struct {
}

// The incorrect package was specified.
func NewFrostErrorIncorrectPackage() *FrostError {
	return &FrostError{err: &FrostErrorIncorrectPackage{}}
}

func (e FrostErrorIncorrectPackage) destroy() {
}

func (err FrostErrorIncorrectPackage) Error() string {
	return fmt.Sprint("IncorrectPackage")
}

func (self FrostErrorIncorrectPackage) Is(target error) bool {
	return target == ErrFrostErrorIncorrectPackage
}

// The ciphersuite does not support DKG.
type FrostErrorDkgNotSupported struct {
}

// The ciphersuite does not support DKG.
func NewFrostErrorDkgNotSupported() *FrostError {
	return &FrostError{err: &FrostErrorDkgNotSupported{}}
}

func (e FrostErrorDkgNotSupported) destroy() {
}

func (err FrostErrorDkgNotSupported) Error() string {
	return fmt.Sprint("DkgNotSupported")
}

func (self FrostErrorDkgNotSupported) Is(target error) bool {
	return target == ErrFrostErrorDkgNotSupported
}

// The proof of knowledge is not valid.
type FrostErrorInvalidProofOfKnowledge struct {
	Culprit ParticipantIdentifier
}

// The proof of knowledge is not valid.
func NewFrostErrorInvalidProofOfKnowledge(
	culprit ParticipantIdentifier,
) *FrostError {
	return &FrostError{err: &FrostErrorInvalidProofOfKnowledge{
		Culprit: culprit}}
}

func (e FrostErrorInvalidProofOfKnowledge) destroy() {
	FfiDestroyerParticipantIdentifier{}.Destroy(e.Culprit)
}

func (err FrostErrorInvalidProofOfKnowledge) Error() string {
	return fmt.Sprint("InvalidProofOfKnowledge",
		": ",

		"Culprit=",
		err.Culprit,
	)
}

func (self FrostErrorInvalidProofOfKnowledge) Is(target error) bool {
	return target == ErrFrostErrorInvalidProofOfKnowledge
}

// Error in scalar Field.
type FrostErrorFieldError struct {
	Message string
}

// Error in scalar Field.
func NewFrostErrorFieldError(
	message string,
) *FrostError {
	return &FrostError{err: &FrostErrorFieldError{
		Message: message}}
}

func (e FrostErrorFieldError) destroy() {
	FfiDestroyerString{}.Destroy(e.Message)
}

func (err FrostErrorFieldError) Error() string {
	return fmt.Sprint("FieldError",
		": ",

		"Message=",
		err.Message,
	)
}

func (self FrostErrorFieldError) Is(target error) bool {
	return target == ErrFrostErrorFieldError
}

// Error in elliptic curve Group.
type FrostErrorGroupError struct {
	Message string
}

// Error in elliptic curve Group.
func NewFrostErrorGroupError(
	message string,
) *FrostError {
	return &FrostError{err: &FrostErrorGroupError{
		Message: message}}
}

func (e FrostErrorGroupError) destroy() {
	FfiDestroyerString{}.Destroy(e.Message)
}

func (err FrostErrorGroupError) Error() string {
	return fmt.Sprint("GroupError",
		": ",

		"Message=",
		err.Message,
	)
}

func (self FrostErrorGroupError) Is(target error) bool {
	return target == ErrFrostErrorGroupError
}

// Error in coefficient commitment deserialization.
type FrostErrorInvalidCoefficient struct {
}

// Error in coefficient commitment deserialization.
func NewFrostErrorInvalidCoefficient() *FrostError {
	return &FrostError{err: &FrostErrorInvalidCoefficient{}}
}

func (e FrostErrorInvalidCoefficient) destroy() {
}

func (err FrostErrorInvalidCoefficient) Error() string {
	return fmt.Sprint("InvalidCoefficient")
}

func (self FrostErrorInvalidCoefficient) Is(target error) bool {
	return target == ErrFrostErrorInvalidCoefficient
}

// The ciphersuite does not support deriving identifiers from strings.
type FrostErrorIdentifierDerivationNotSupported struct {
}

// The ciphersuite does not support deriving identifiers from strings.
func NewFrostErrorIdentifierDerivationNotSupported() *FrostError {
	return &FrostError{err: &FrostErrorIdentifierDerivationNotSupported{}}
}

func (e FrostErrorIdentifierDerivationNotSupported) destroy() {
}

func (err FrostErrorIdentifierDerivationNotSupported) Error() string {
	return fmt.Sprint("IdentifierDerivationNotSupported")
}

func (self FrostErrorIdentifierDerivationNotSupported) Is(target error) bool {
	return target == ErrFrostErrorIdentifierDerivationNotSupported
}

// Error serializing value.
type FrostErrorSerializationError struct {
}

// Error serializing value.
func NewFrostErrorSerializationError() *FrostError {
	return &FrostError{err: &FrostErrorSerializationError{}}
}

func (e FrostErrorSerializationError) destroy() {
}

func (err FrostErrorSerializationError) Error() string {
	return fmt.Sprint("SerializationError")
}

func (self FrostErrorSerializationError) Is(target error) bool {
	return target == ErrFrostErrorSerializationError
}

// Error deserializing value.
type FrostErrorDeserializationError struct {
}

// Error deserializing value.
func NewFrostErrorDeserializationError() *FrostError {
	return &FrostError{err: &FrostErrorDeserializationError{}}
}

func (e FrostErrorDeserializationError) destroy() {
}

func (err FrostErrorDeserializationError) Error() string {
	return fmt.Sprint("DeserializationError")
}

func (self FrostErrorDeserializationError) Is(target error) bool {
	return target == ErrFrostErrorDeserializationError
}

type FrostErrorDkgPart2IncorrectNumberOfCommitments struct {
}

func NewFrostErrorDkgPart2IncorrectNumberOfCommitments() *FrostError {
	return &FrostError{err: &FrostErrorDkgPart2IncorrectNumberOfCommitments{}}
}

func (e FrostErrorDkgPart2IncorrectNumberOfCommitments) destroy() {
}

func (err FrostErrorDkgPart2IncorrectNumberOfCommitments) Error() string {
	return fmt.Sprint("DkgPart2IncorrectNumberOfCommitments")
}

func (self FrostErrorDkgPart2IncorrectNumberOfCommitments) Is(target error) bool {
	return target == ErrFrostErrorDkgPart2IncorrectNumberOfCommitments
}

type FrostErrorDkgPart2IncorrectNumberOfPackages struct {
}

func NewFrostErrorDkgPart2IncorrectNumberOfPackages() *FrostError {
	return &FrostError{err: &FrostErrorDkgPart2IncorrectNumberOfPackages{}}
}

func (e FrostErrorDkgPart2IncorrectNumberOfPackages) destroy() {
}

func (err FrostErrorDkgPart2IncorrectNumberOfPackages) Error() string {
	return fmt.Sprint("DkgPart2IncorrectNumberOfPackages")
}

func (self FrostErrorDkgPart2IncorrectNumberOfPackages) Is(target error) bool {
	return target == ErrFrostErrorDkgPart2IncorrectNumberOfPackages
}

type FrostErrorDkgPart3IncorrectRound1Packages struct {
}

func NewFrostErrorDkgPart3IncorrectRound1Packages() *FrostError {
	return &FrostError{err: &FrostErrorDkgPart3IncorrectRound1Packages{}}
}

func (e FrostErrorDkgPart3IncorrectRound1Packages) destroy() {
}

func (err FrostErrorDkgPart3IncorrectRound1Packages) Error() string {
	return fmt.Sprint("DkgPart3IncorrectRound1Packages")
}

func (self FrostErrorDkgPart3IncorrectRound1Packages) Is(target error) bool {
	return target == ErrFrostErrorDkgPart3IncorrectRound1Packages
}

type FrostErrorDkgPart3IncorrectNumberOfPackages struct {
}

func NewFrostErrorDkgPart3IncorrectNumberOfPackages() *FrostError {
	return &FrostError{err: &FrostErrorDkgPart3IncorrectNumberOfPackages{}}
}

func (e FrostErrorDkgPart3IncorrectNumberOfPackages) destroy() {
}

func (err FrostErrorDkgPart3IncorrectNumberOfPackages) Error() string {
	return fmt.Sprint("DkgPart3IncorrectNumberOfPackages")
}

func (self FrostErrorDkgPart3IncorrectNumberOfPackages) Is(target error) bool {
	return target == ErrFrostErrorDkgPart3IncorrectNumberOfPackages
}

type FrostErrorDkgPart3PackageSendersMismatch struct {
}

func NewFrostErrorDkgPart3PackageSendersMismatch() *FrostError {
	return &FrostError{err: &FrostErrorDkgPart3PackageSendersMismatch{}}
}

func (e FrostErrorDkgPart3PackageSendersMismatch) destroy() {
}

func (err FrostErrorDkgPart3PackageSendersMismatch) Error() string {
	return fmt.Sprint("DkgPart3PackageSendersMismatch")
}

func (self FrostErrorDkgPart3PackageSendersMismatch) Is(target error) bool {
	return target == ErrFrostErrorDkgPart3PackageSendersMismatch
}

type FrostErrorInvalidKeyPackage struct {
}

func NewFrostErrorInvalidKeyPackage() *FrostError {
	return &FrostError{err: &FrostErrorInvalidKeyPackage{}}
}

func (e FrostErrorInvalidKeyPackage) destroy() {
}

func (err FrostErrorInvalidKeyPackage) Error() string {
	return fmt.Sprint("InvalidKeyPackage")
}

func (self FrostErrorInvalidKeyPackage) Is(target error) bool {
	return target == ErrFrostErrorInvalidKeyPackage
}

type FrostErrorInvalidSecretKey struct {
}

func NewFrostErrorInvalidSecretKey() *FrostError {
	return &FrostError{err: &FrostErrorInvalidSecretKey{}}
}

func (e FrostErrorInvalidSecretKey) destroy() {
}

func (err FrostErrorInvalidSecretKey) Error() string {
	return fmt.Sprint("InvalidSecretKey")
}

func (self FrostErrorInvalidSecretKey) Is(target error) bool {
	return target == ErrFrostErrorInvalidSecretKey
}

type FrostErrorInvalidConfiguration struct {
}

func NewFrostErrorInvalidConfiguration() *FrostError {
	return &FrostError{err: &FrostErrorInvalidConfiguration{}}
}

func (e FrostErrorInvalidConfiguration) destroy() {
}

func (err FrostErrorInvalidConfiguration) Error() string {
	return fmt.Sprint("InvalidConfiguration")
}

func (self FrostErrorInvalidConfiguration) Is(target error) bool {
	return target == ErrFrostErrorInvalidConfiguration
}

type FrostErrorUnexpectedError struct {
}

func NewFrostErrorUnexpectedError() *FrostError {
	return &FrostError{err: &FrostErrorUnexpectedError{}}
}

func (e FrostErrorUnexpectedError) destroy() {
}

func (err FrostErrorUnexpectedError) Error() string {
	return fmt.Sprint("UnexpectedError")
}

func (self FrostErrorUnexpectedError) Is(target error) bool {
	return target == ErrFrostErrorUnexpectedError
}

type FfiConverterFrostError struct{}

var FfiConverterFrostErrorINSTANCE = FfiConverterFrostError{}

func (c FfiConverterFrostError) Lift(eb RustBufferI) *FrostError {
	return LiftFromRustBuffer[*FrostError](c, eb)
}

func (c FfiConverterFrostError) Lower(value *FrostError) C.RustBuffer {
	return LowerIntoRustBuffer[*FrostError](c, value)
}

func (c FfiConverterFrostError) Read(reader io.Reader) *FrostError {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &FrostError{&FrostErrorInvalidMinSigners{}}
	case 2:
		return &FrostError{&FrostErrorInvalidMaxSigners{}}
	case 3:
		return &FrostError{&FrostErrorInvalidCoefficients{}}
	case 4:
		return &FrostError{&FrostErrorMalformedIdentifier{}}
	case 5:
		return &FrostError{&FrostErrorDuplicatedIdentifier{}}
	case 6:
		return &FrostError{&FrostErrorUnknownIdentifier{}}
	case 7:
		return &FrostError{&FrostErrorIncorrectNumberOfIdentifiers{}}
	case 8:
		return &FrostError{&FrostErrorMalformedSigningKey{}}
	case 9:
		return &FrostError{&FrostErrorMalformedVerifyingKey{}}
	case 10:
		return &FrostError{&FrostErrorMalformedSignature{}}
	case 11:
		return &FrostError{&FrostErrorInvalidSignature{}}
	case 12:
		return &FrostError{&FrostErrorDuplicatedShares{}}
	case 13:
		return &FrostError{&FrostErrorIncorrectNumberOfShares{}}
	case 14:
		return &FrostError{&FrostErrorIdentityCommitment{}}
	case 15:
		return &FrostError{&FrostErrorMissingCommitment{}}
	case 16:
		return &FrostError{&FrostErrorIncorrectCommitment{}}
	case 17:
		return &FrostError{&FrostErrorIncorrectNumberOfCommitments{}}
	case 18:
		return &FrostError{&FrostErrorInvalidSignatureShare{
			Culprit: FfiConverterParticipantIdentifierINSTANCE.Read(reader),
		}}
	case 19:
		return &FrostError{&FrostErrorInvalidSecretShare{
			Culprit: FfiConverterOptionalParticipantIdentifierINSTANCE.Read(reader),
		}}
	case 20:
		return &FrostError{&FrostErrorPackageNotFound{}}
	case 21:
		return &FrostError{&FrostErrorIncorrectNumberOfPackages{}}
	case 22:
		return &FrostError{&FrostErrorIncorrectPackage{}}
	case 23:
		return &FrostError{&FrostErrorDkgNotSupported{}}
	case 24:
		return &FrostError{&FrostErrorInvalidProofOfKnowledge{
			Culprit: FfiConverterParticipantIdentifierINSTANCE.Read(reader),
		}}
	case 25:
		return &FrostError{&FrostErrorFieldError{
			Message: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 26:
		return &FrostError{&FrostErrorGroupError{
			Message: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 27:
		return &FrostError{&FrostErrorInvalidCoefficient{}}
	case 28:
		return &FrostError{&FrostErrorIdentifierDerivationNotSupported{}}
	case 29:
		return &FrostError{&FrostErrorSerializationError{}}
	case 30:
		return &FrostError{&FrostErrorDeserializationError{}}
	case 31:
		return &FrostError{&FrostErrorDkgPart2IncorrectNumberOfCommitments{}}
	case 32:
		return &FrostError{&FrostErrorDkgPart2IncorrectNumberOfPackages{}}
	case 33:
		return &FrostError{&FrostErrorDkgPart3IncorrectRound1Packages{}}
	case 34:
		return &FrostError{&FrostErrorDkgPart3IncorrectNumberOfPackages{}}
	case 35:
		return &FrostError{&FrostErrorDkgPart3PackageSendersMismatch{}}
	case 36:
		return &FrostError{&FrostErrorInvalidKeyPackage{}}
	case 37:
		return &FrostError{&FrostErrorInvalidSecretKey{}}
	case 38:
		return &FrostError{&FrostErrorInvalidConfiguration{}}
	case 39:
		return &FrostError{&FrostErrorUnexpectedError{}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterFrostError.Read()", errorID))
	}
}

func (c FfiConverterFrostError) Write(writer io.Writer, value *FrostError) {
	switch variantValue := value.err.(type) {
	case *FrostErrorInvalidMinSigners:
		writeInt32(writer, 1)
	case *FrostErrorInvalidMaxSigners:
		writeInt32(writer, 2)
	case *FrostErrorInvalidCoefficients:
		writeInt32(writer, 3)
	case *FrostErrorMalformedIdentifier:
		writeInt32(writer, 4)
	case *FrostErrorDuplicatedIdentifier:
		writeInt32(writer, 5)
	case *FrostErrorUnknownIdentifier:
		writeInt32(writer, 6)
	case *FrostErrorIncorrectNumberOfIdentifiers:
		writeInt32(writer, 7)
	case *FrostErrorMalformedSigningKey:
		writeInt32(writer, 8)
	case *FrostErrorMalformedVerifyingKey:
		writeInt32(writer, 9)
	case *FrostErrorMalformedSignature:
		writeInt32(writer, 10)
	case *FrostErrorInvalidSignature:
		writeInt32(writer, 11)
	case *FrostErrorDuplicatedShares:
		writeInt32(writer, 12)
	case *FrostErrorIncorrectNumberOfShares:
		writeInt32(writer, 13)
	case *FrostErrorIdentityCommitment:
		writeInt32(writer, 14)
	case *FrostErrorMissingCommitment:
		writeInt32(writer, 15)
	case *FrostErrorIncorrectCommitment:
		writeInt32(writer, 16)
	case *FrostErrorIncorrectNumberOfCommitments:
		writeInt32(writer, 17)
	case *FrostErrorInvalidSignatureShare:
		writeInt32(writer, 18)
		FfiConverterParticipantIdentifierINSTANCE.Write(writer, variantValue.Culprit)
	case *FrostErrorInvalidSecretShare:
		writeInt32(writer, 19)
		FfiConverterOptionalParticipantIdentifierINSTANCE.Write(writer, variantValue.Culprit)
	case *FrostErrorPackageNotFound:
		writeInt32(writer, 20)
	case *FrostErrorIncorrectNumberOfPackages:
		writeInt32(writer, 21)
	case *FrostErrorIncorrectPackage:
		writeInt32(writer, 22)
	case *FrostErrorDkgNotSupported:
		writeInt32(writer, 23)
	case *FrostErrorInvalidProofOfKnowledge:
		writeInt32(writer, 24)
		FfiConverterParticipantIdentifierINSTANCE.Write(writer, variantValue.Culprit)
	case *FrostErrorFieldError:
		writeInt32(writer, 25)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Message)
	case *FrostErrorGroupError:
		writeInt32(writer, 26)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Message)
	case *FrostErrorInvalidCoefficient:
		writeInt32(writer, 27)
	case *FrostErrorIdentifierDerivationNotSupported:
		writeInt32(writer, 28)
	case *FrostErrorSerializationError:
		writeInt32(writer, 29)
	case *FrostErrorDeserializationError:
		writeInt32(writer, 30)
	case *FrostErrorDkgPart2IncorrectNumberOfCommitments:
		writeInt32(writer, 31)
	case *FrostErrorDkgPart2IncorrectNumberOfPackages:
		writeInt32(writer, 32)
	case *FrostErrorDkgPart3IncorrectRound1Packages:
		writeInt32(writer, 33)
	case *FrostErrorDkgPart3IncorrectNumberOfPackages:
		writeInt32(writer, 34)
	case *FrostErrorDkgPart3PackageSendersMismatch:
		writeInt32(writer, 35)
	case *FrostErrorInvalidKeyPackage:
		writeInt32(writer, 36)
	case *FrostErrorInvalidSecretKey:
		writeInt32(writer, 37)
	case *FrostErrorInvalidConfiguration:
		writeInt32(writer, 38)
	case *FrostErrorUnexpectedError:
		writeInt32(writer, 39)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterFrostError.Write", value))
	}
}

type FfiDestroyerFrostError struct{}

func (_ FfiDestroyerFrostError) Destroy(value *FrostError) {
	switch variantValue := value.err.(type) {
	case FrostErrorInvalidMinSigners:
		variantValue.destroy()
	case FrostErrorInvalidMaxSigners:
		variantValue.destroy()
	case FrostErrorInvalidCoefficients:
		variantValue.destroy()
	case FrostErrorMalformedIdentifier:
		variantValue.destroy()
	case FrostErrorDuplicatedIdentifier:
		variantValue.destroy()
	case FrostErrorUnknownIdentifier:
		variantValue.destroy()
	case FrostErrorIncorrectNumberOfIdentifiers:
		variantValue.destroy()
	case FrostErrorMalformedSigningKey:
		variantValue.destroy()
	case FrostErrorMalformedVerifyingKey:
		variantValue.destroy()
	case FrostErrorMalformedSignature:
		variantValue.destroy()
	case FrostErrorInvalidSignature:
		variantValue.destroy()
	case FrostErrorDuplicatedShares:
		variantValue.destroy()
	case FrostErrorIncorrectNumberOfShares:
		variantValue.destroy()
	case FrostErrorIdentityCommitment:
		variantValue.destroy()
	case FrostErrorMissingCommitment:
		variantValue.destroy()
	case FrostErrorIncorrectCommitment:
		variantValue.destroy()
	case FrostErrorIncorrectNumberOfCommitments:
		variantValue.destroy()
	case FrostErrorInvalidSignatureShare:
		variantValue.destroy()
	case FrostErrorInvalidSecretShare:
		variantValue.destroy()
	case FrostErrorPackageNotFound:
		variantValue.destroy()
	case FrostErrorIncorrectNumberOfPackages:
		variantValue.destroy()
	case FrostErrorIncorrectPackage:
		variantValue.destroy()
	case FrostErrorDkgNotSupported:
		variantValue.destroy()
	case FrostErrorInvalidProofOfKnowledge:
		variantValue.destroy()
	case FrostErrorFieldError:
		variantValue.destroy()
	case FrostErrorGroupError:
		variantValue.destroy()
	case FrostErrorInvalidCoefficient:
		variantValue.destroy()
	case FrostErrorIdentifierDerivationNotSupported:
		variantValue.destroy()
	case FrostErrorSerializationError:
		variantValue.destroy()
	case FrostErrorDeserializationError:
		variantValue.destroy()
	case FrostErrorDkgPart2IncorrectNumberOfCommitments:
		variantValue.destroy()
	case FrostErrorDkgPart2IncorrectNumberOfPackages:
		variantValue.destroy()
	case FrostErrorDkgPart3IncorrectRound1Packages:
		variantValue.destroy()
	case FrostErrorDkgPart3IncorrectNumberOfPackages:
		variantValue.destroy()
	case FrostErrorDkgPart3PackageSendersMismatch:
		variantValue.destroy()
	case FrostErrorInvalidKeyPackage:
		variantValue.destroy()
	case FrostErrorInvalidSecretKey:
		variantValue.destroy()
	case FrostErrorInvalidConfiguration:
		variantValue.destroy()
	case FrostErrorUnexpectedError:
		variantValue.destroy()
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiDestroyerFrostError.Destroy", value))
	}
}

type FrostSignatureVerificationError struct {
	err error
}

// Convience method to turn *FrostSignatureVerificationError into error
// Avoiding treating nil pointer as non nil error interface
func (err *FrostSignatureVerificationError) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err FrostSignatureVerificationError) Error() string {
	return fmt.Sprintf("FrostSignatureVerificationError: %s", err.err.Error())
}

func (err FrostSignatureVerificationError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrFrostSignatureVerificationErrorInvalidPublicKeyPackage = fmt.Errorf("FrostSignatureVerificationErrorInvalidPublicKeyPackage")
var ErrFrostSignatureVerificationErrorValidationFailed = fmt.Errorf("FrostSignatureVerificationErrorValidationFailed")

// Variant structs
type FrostSignatureVerificationErrorInvalidPublicKeyPackage struct {
}

func NewFrostSignatureVerificationErrorInvalidPublicKeyPackage() *FrostSignatureVerificationError {
	return &FrostSignatureVerificationError{err: &FrostSignatureVerificationErrorInvalidPublicKeyPackage{}}
}

func (e FrostSignatureVerificationErrorInvalidPublicKeyPackage) destroy() {
}

func (err FrostSignatureVerificationErrorInvalidPublicKeyPackage) Error() string {
	return fmt.Sprint("InvalidPublicKeyPackage")
}

func (self FrostSignatureVerificationErrorInvalidPublicKeyPackage) Is(target error) bool {
	return target == ErrFrostSignatureVerificationErrorInvalidPublicKeyPackage
}

type FrostSignatureVerificationErrorValidationFailed struct {
	Reason string
}

func NewFrostSignatureVerificationErrorValidationFailed(
	reason string,
) *FrostSignatureVerificationError {
	return &FrostSignatureVerificationError{err: &FrostSignatureVerificationErrorValidationFailed{
		Reason: reason}}
}

func (e FrostSignatureVerificationErrorValidationFailed) destroy() {
	FfiDestroyerString{}.Destroy(e.Reason)
}

func (err FrostSignatureVerificationErrorValidationFailed) Error() string {
	return fmt.Sprint("ValidationFailed",
		": ",

		"Reason=",
		err.Reason,
	)
}

func (self FrostSignatureVerificationErrorValidationFailed) Is(target error) bool {
	return target == ErrFrostSignatureVerificationErrorValidationFailed
}

type FfiConverterFrostSignatureVerificationError struct{}

var FfiConverterFrostSignatureVerificationErrorINSTANCE = FfiConverterFrostSignatureVerificationError{}

func (c FfiConverterFrostSignatureVerificationError) Lift(eb RustBufferI) *FrostSignatureVerificationError {
	return LiftFromRustBuffer[*FrostSignatureVerificationError](c, eb)
}

func (c FfiConverterFrostSignatureVerificationError) Lower(value *FrostSignatureVerificationError) C.RustBuffer {
	return LowerIntoRustBuffer[*FrostSignatureVerificationError](c, value)
}

func (c FfiConverterFrostSignatureVerificationError) Read(reader io.Reader) *FrostSignatureVerificationError {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &FrostSignatureVerificationError{&FrostSignatureVerificationErrorInvalidPublicKeyPackage{}}
	case 2:
		return &FrostSignatureVerificationError{&FrostSignatureVerificationErrorValidationFailed{
			Reason: FfiConverterStringINSTANCE.Read(reader),
		}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterFrostSignatureVerificationError.Read()", errorID))
	}
}

func (c FfiConverterFrostSignatureVerificationError) Write(writer io.Writer, value *FrostSignatureVerificationError) {
	switch variantValue := value.err.(type) {
	case *FrostSignatureVerificationErrorInvalidPublicKeyPackage:
		writeInt32(writer, 1)
	case *FrostSignatureVerificationErrorValidationFailed:
		writeInt32(writer, 2)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Reason)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterFrostSignatureVerificationError.Write", value))
	}
}

type FfiDestroyerFrostSignatureVerificationError struct{}

func (_ FfiDestroyerFrostSignatureVerificationError) Destroy(value *FrostSignatureVerificationError) {
	switch variantValue := value.err.(type) {
	case FrostSignatureVerificationErrorInvalidPublicKeyPackage:
		variantValue.destroy()
	case FrostSignatureVerificationErrorValidationFailed:
		variantValue.destroy()
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiDestroyerFrostSignatureVerificationError.Destroy", value))
	}
}

type OrchardKeyError struct {
	err error
}

// Convience method to turn *OrchardKeyError into error
// Avoiding treating nil pointer as non nil error interface
func (err *OrchardKeyError) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err OrchardKeyError) Error() string {
	return fmt.Sprintf("OrchardKeyError: %s", err.err.Error())
}

func (err OrchardKeyError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrOrchardKeyErrorKeyDerivationError = fmt.Errorf("OrchardKeyErrorKeyDerivationError")
var ErrOrchardKeyErrorSerializationError = fmt.Errorf("OrchardKeyErrorSerializationError")
var ErrOrchardKeyErrorDeserializationError = fmt.Errorf("OrchardKeyErrorDeserializationError")
var ErrOrchardKeyErrorOtherError = fmt.Errorf("OrchardKeyErrorOtherError")

// Variant structs
type OrchardKeyErrorKeyDerivationError struct {
	Message string
}

func NewOrchardKeyErrorKeyDerivationError(
	message string,
) *OrchardKeyError {
	return &OrchardKeyError{err: &OrchardKeyErrorKeyDerivationError{
		Message: message}}
}

func (e OrchardKeyErrorKeyDerivationError) destroy() {
	FfiDestroyerString{}.Destroy(e.Message)
}

func (err OrchardKeyErrorKeyDerivationError) Error() string {
	return fmt.Sprint("KeyDerivationError",
		": ",

		"Message=",
		err.Message,
	)
}

func (self OrchardKeyErrorKeyDerivationError) Is(target error) bool {
	return target == ErrOrchardKeyErrorKeyDerivationError
}

type OrchardKeyErrorSerializationError struct {
}

func NewOrchardKeyErrorSerializationError() *OrchardKeyError {
	return &OrchardKeyError{err: &OrchardKeyErrorSerializationError{}}
}

func (e OrchardKeyErrorSerializationError) destroy() {
}

func (err OrchardKeyErrorSerializationError) Error() string {
	return fmt.Sprint("SerializationError")
}

func (self OrchardKeyErrorSerializationError) Is(target error) bool {
	return target == ErrOrchardKeyErrorSerializationError
}

type OrchardKeyErrorDeserializationError struct {
}

func NewOrchardKeyErrorDeserializationError() *OrchardKeyError {
	return &OrchardKeyError{err: &OrchardKeyErrorDeserializationError{}}
}

func (e OrchardKeyErrorDeserializationError) destroy() {
}

func (err OrchardKeyErrorDeserializationError) Error() string {
	return fmt.Sprint("DeserializationError")
}

func (self OrchardKeyErrorDeserializationError) Is(target error) bool {
	return target == ErrOrchardKeyErrorDeserializationError
}

type OrchardKeyErrorOtherError struct {
	ErrorMessage string
}

func NewOrchardKeyErrorOtherError(
	errorMessage string,
) *OrchardKeyError {
	return &OrchardKeyError{err: &OrchardKeyErrorOtherError{
		ErrorMessage: errorMessage}}
}

func (e OrchardKeyErrorOtherError) destroy() {
	FfiDestroyerString{}.Destroy(e.ErrorMessage)
}

func (err OrchardKeyErrorOtherError) Error() string {
	return fmt.Sprint("OtherError",
		": ",

		"ErrorMessage=",
		err.ErrorMessage,
	)
}

func (self OrchardKeyErrorOtherError) Is(target error) bool {
	return target == ErrOrchardKeyErrorOtherError
}

type FfiConverterOrchardKeyError struct{}

var FfiConverterOrchardKeyErrorINSTANCE = FfiConverterOrchardKeyError{}

func (c FfiConverterOrchardKeyError) Lift(eb RustBufferI) *OrchardKeyError {
	return LiftFromRustBuffer[*OrchardKeyError](c, eb)
}

func (c FfiConverterOrchardKeyError) Lower(value *OrchardKeyError) C.RustBuffer {
	return LowerIntoRustBuffer[*OrchardKeyError](c, value)
}

func (c FfiConverterOrchardKeyError) Read(reader io.Reader) *OrchardKeyError {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &OrchardKeyError{&OrchardKeyErrorKeyDerivationError{
			Message: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 2:
		return &OrchardKeyError{&OrchardKeyErrorSerializationError{}}
	case 3:
		return &OrchardKeyError{&OrchardKeyErrorDeserializationError{}}
	case 4:
		return &OrchardKeyError{&OrchardKeyErrorOtherError{
			ErrorMessage: FfiConverterStringINSTANCE.Read(reader),
		}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterOrchardKeyError.Read()", errorID))
	}
}

func (c FfiConverterOrchardKeyError) Write(writer io.Writer, value *OrchardKeyError) {
	switch variantValue := value.err.(type) {
	case *OrchardKeyErrorKeyDerivationError:
		writeInt32(writer, 1)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Message)
	case *OrchardKeyErrorSerializationError:
		writeInt32(writer, 2)
	case *OrchardKeyErrorDeserializationError:
		writeInt32(writer, 3)
	case *OrchardKeyErrorOtherError:
		writeInt32(writer, 4)
		FfiConverterStringINSTANCE.Write(writer, variantValue.ErrorMessage)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterOrchardKeyError.Write", value))
	}
}

type FfiDestroyerOrchardKeyError struct{}

func (_ FfiDestroyerOrchardKeyError) Destroy(value *OrchardKeyError) {
	switch variantValue := value.err.(type) {
	case OrchardKeyErrorKeyDerivationError:
		variantValue.destroy()
	case OrchardKeyErrorSerializationError:
		variantValue.destroy()
	case OrchardKeyErrorDeserializationError:
		variantValue.destroy()
	case OrchardKeyErrorOtherError:
		variantValue.destroy()
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiDestroyerOrchardKeyError.Destroy", value))
	}
}

type Round1Error struct {
	err error
}

// Convience method to turn *Round1Error into error
// Avoiding treating nil pointer as non nil error interface
func (err *Round1Error) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err Round1Error) Error() string {
	return fmt.Sprintf("Round1Error: %s", err.err.Error())
}

func (err Round1Error) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrRound1ErrorInvalidKeyPackage = fmt.Errorf("Round1ErrorInvalidKeyPackage")
var ErrRound1ErrorNonceSerializationError = fmt.Errorf("Round1ErrorNonceSerializationError")
var ErrRound1ErrorCommitmentSerializationError = fmt.Errorf("Round1ErrorCommitmentSerializationError")

// Variant structs
type Round1ErrorInvalidKeyPackage struct {
}

func NewRound1ErrorInvalidKeyPackage() *Round1Error {
	return &Round1Error{err: &Round1ErrorInvalidKeyPackage{}}
}

func (e Round1ErrorInvalidKeyPackage) destroy() {
}

func (err Round1ErrorInvalidKeyPackage) Error() string {
	return fmt.Sprint("InvalidKeyPackage")
}

func (self Round1ErrorInvalidKeyPackage) Is(target error) bool {
	return target == ErrRound1ErrorInvalidKeyPackage
}

type Round1ErrorNonceSerializationError struct {
}

func NewRound1ErrorNonceSerializationError() *Round1Error {
	return &Round1Error{err: &Round1ErrorNonceSerializationError{}}
}

func (e Round1ErrorNonceSerializationError) destroy() {
}

func (err Round1ErrorNonceSerializationError) Error() string {
	return fmt.Sprint("NonceSerializationError")
}

func (self Round1ErrorNonceSerializationError) Is(target error) bool {
	return target == ErrRound1ErrorNonceSerializationError
}

type Round1ErrorCommitmentSerializationError struct {
}

func NewRound1ErrorCommitmentSerializationError() *Round1Error {
	return &Round1Error{err: &Round1ErrorCommitmentSerializationError{}}
}

func (e Round1ErrorCommitmentSerializationError) destroy() {
}

func (err Round1ErrorCommitmentSerializationError) Error() string {
	return fmt.Sprint("CommitmentSerializationError")
}

func (self Round1ErrorCommitmentSerializationError) Is(target error) bool {
	return target == ErrRound1ErrorCommitmentSerializationError
}

type FfiConverterRound1Error struct{}

var FfiConverterRound1ErrorINSTANCE = FfiConverterRound1Error{}

func (c FfiConverterRound1Error) Lift(eb RustBufferI) *Round1Error {
	return LiftFromRustBuffer[*Round1Error](c, eb)
}

func (c FfiConverterRound1Error) Lower(value *Round1Error) C.RustBuffer {
	return LowerIntoRustBuffer[*Round1Error](c, value)
}

func (c FfiConverterRound1Error) Read(reader io.Reader) *Round1Error {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &Round1Error{&Round1ErrorInvalidKeyPackage{}}
	case 2:
		return &Round1Error{&Round1ErrorNonceSerializationError{}}
	case 3:
		return &Round1Error{&Round1ErrorCommitmentSerializationError{}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterRound1Error.Read()", errorID))
	}
}

func (c FfiConverterRound1Error) Write(writer io.Writer, value *Round1Error) {
	switch variantValue := value.err.(type) {
	case *Round1ErrorInvalidKeyPackage:
		writeInt32(writer, 1)
	case *Round1ErrorNonceSerializationError:
		writeInt32(writer, 2)
	case *Round1ErrorCommitmentSerializationError:
		writeInt32(writer, 3)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterRound1Error.Write", value))
	}
}

type FfiDestroyerRound1Error struct{}

func (_ FfiDestroyerRound1Error) Destroy(value *Round1Error) {
	switch variantValue := value.err.(type) {
	case Round1ErrorInvalidKeyPackage:
		variantValue.destroy()
	case Round1ErrorNonceSerializationError:
		variantValue.destroy()
	case Round1ErrorCommitmentSerializationError:
		variantValue.destroy()
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiDestroyerRound1Error.Destroy", value))
	}
}

type Round2Error struct {
	err error
}

// Convience method to turn *Round2Error into error
// Avoiding treating nil pointer as non nil error interface
func (err *Round2Error) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err Round2Error) Error() string {
	return fmt.Sprintf("Round2Error: %s", err.err.Error())
}

func (err Round2Error) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrRound2ErrorInvalidKeyPackage = fmt.Errorf("Round2ErrorInvalidKeyPackage")
var ErrRound2ErrorNonceSerializationError = fmt.Errorf("Round2ErrorNonceSerializationError")
var ErrRound2ErrorCommitmentSerializationError = fmt.Errorf("Round2ErrorCommitmentSerializationError")
var ErrRound2ErrorSigningPackageDeserializationError = fmt.Errorf("Round2ErrorSigningPackageDeserializationError")
var ErrRound2ErrorSigningFailed = fmt.Errorf("Round2ErrorSigningFailed")
var ErrRound2ErrorInvalidRandomizer = fmt.Errorf("Round2ErrorInvalidRandomizer")

// Variant structs
type Round2ErrorInvalidKeyPackage struct {
}

func NewRound2ErrorInvalidKeyPackage() *Round2Error {
	return &Round2Error{err: &Round2ErrorInvalidKeyPackage{}}
}

func (e Round2ErrorInvalidKeyPackage) destroy() {
}

func (err Round2ErrorInvalidKeyPackage) Error() string {
	return fmt.Sprint("InvalidKeyPackage")
}

func (self Round2ErrorInvalidKeyPackage) Is(target error) bool {
	return target == ErrRound2ErrorInvalidKeyPackage
}

type Round2ErrorNonceSerializationError struct {
}

func NewRound2ErrorNonceSerializationError() *Round2Error {
	return &Round2Error{err: &Round2ErrorNonceSerializationError{}}
}

func (e Round2ErrorNonceSerializationError) destroy() {
}

func (err Round2ErrorNonceSerializationError) Error() string {
	return fmt.Sprint("NonceSerializationError")
}

func (self Round2ErrorNonceSerializationError) Is(target error) bool {
	return target == ErrRound2ErrorNonceSerializationError
}

type Round2ErrorCommitmentSerializationError struct {
}

func NewRound2ErrorCommitmentSerializationError() *Round2Error {
	return &Round2Error{err: &Round2ErrorCommitmentSerializationError{}}
}

func (e Round2ErrorCommitmentSerializationError) destroy() {
}

func (err Round2ErrorCommitmentSerializationError) Error() string {
	return fmt.Sprint("CommitmentSerializationError")
}

func (self Round2ErrorCommitmentSerializationError) Is(target error) bool {
	return target == ErrRound2ErrorCommitmentSerializationError
}

type Round2ErrorSigningPackageDeserializationError struct {
}

func NewRound2ErrorSigningPackageDeserializationError() *Round2Error {
	return &Round2Error{err: &Round2ErrorSigningPackageDeserializationError{}}
}

func (e Round2ErrorSigningPackageDeserializationError) destroy() {
}

func (err Round2ErrorSigningPackageDeserializationError) Error() string {
	return fmt.Sprint("SigningPackageDeserializationError")
}

func (self Round2ErrorSigningPackageDeserializationError) Is(target error) bool {
	return target == ErrRound2ErrorSigningPackageDeserializationError
}

type Round2ErrorSigningFailed struct {
	Message string
}

func NewRound2ErrorSigningFailed(
	message string,
) *Round2Error {
	return &Round2Error{err: &Round2ErrorSigningFailed{
		Message: message}}
}

func (e Round2ErrorSigningFailed) destroy() {
	FfiDestroyerString{}.Destroy(e.Message)
}

func (err Round2ErrorSigningFailed) Error() string {
	return fmt.Sprint("SigningFailed",
		": ",

		"Message=",
		err.Message,
	)
}

func (self Round2ErrorSigningFailed) Is(target error) bool {
	return target == ErrRound2ErrorSigningFailed
}

type Round2ErrorInvalidRandomizer struct {
}

func NewRound2ErrorInvalidRandomizer() *Round2Error {
	return &Round2Error{err: &Round2ErrorInvalidRandomizer{}}
}

func (e Round2ErrorInvalidRandomizer) destroy() {
}

func (err Round2ErrorInvalidRandomizer) Error() string {
	return fmt.Sprint("InvalidRandomizer")
}

func (self Round2ErrorInvalidRandomizer) Is(target error) bool {
	return target == ErrRound2ErrorInvalidRandomizer
}

type FfiConverterRound2Error struct{}

var FfiConverterRound2ErrorINSTANCE = FfiConverterRound2Error{}

func (c FfiConverterRound2Error) Lift(eb RustBufferI) *Round2Error {
	return LiftFromRustBuffer[*Round2Error](c, eb)
}

func (c FfiConverterRound2Error) Lower(value *Round2Error) C.RustBuffer {
	return LowerIntoRustBuffer[*Round2Error](c, value)
}

func (c FfiConverterRound2Error) Read(reader io.Reader) *Round2Error {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &Round2Error{&Round2ErrorInvalidKeyPackage{}}
	case 2:
		return &Round2Error{&Round2ErrorNonceSerializationError{}}
	case 3:
		return &Round2Error{&Round2ErrorCommitmentSerializationError{}}
	case 4:
		return &Round2Error{&Round2ErrorSigningPackageDeserializationError{}}
	case 5:
		return &Round2Error{&Round2ErrorSigningFailed{
			Message: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 6:
		return &Round2Error{&Round2ErrorInvalidRandomizer{}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterRound2Error.Read()", errorID))
	}
}

func (c FfiConverterRound2Error) Write(writer io.Writer, value *Round2Error) {
	switch variantValue := value.err.(type) {
	case *Round2ErrorInvalidKeyPackage:
		writeInt32(writer, 1)
	case *Round2ErrorNonceSerializationError:
		writeInt32(writer, 2)
	case *Round2ErrorCommitmentSerializationError:
		writeInt32(writer, 3)
	case *Round2ErrorSigningPackageDeserializationError:
		writeInt32(writer, 4)
	case *Round2ErrorSigningFailed:
		writeInt32(writer, 5)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Message)
	case *Round2ErrorInvalidRandomizer:
		writeInt32(writer, 6)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterRound2Error.Write", value))
	}
}

type FfiDestroyerRound2Error struct{}

func (_ FfiDestroyerRound2Error) Destroy(value *Round2Error) {
	switch variantValue := value.err.(type) {
	case Round2ErrorInvalidKeyPackage:
		variantValue.destroy()
	case Round2ErrorNonceSerializationError:
		variantValue.destroy()
	case Round2ErrorCommitmentSerializationError:
		variantValue.destroy()
	case Round2ErrorSigningPackageDeserializationError:
		variantValue.destroy()
	case Round2ErrorSigningFailed:
		variantValue.destroy()
	case Round2ErrorInvalidRandomizer:
		variantValue.destroy()
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiDestroyerRound2Error.Destroy", value))
	}
}

type ZcashNetwork uint

const (
	ZcashNetworkMainnet ZcashNetwork = 1
	ZcashNetworkTestnet ZcashNetwork = 2
)

type FfiConverterZcashNetwork struct{}

var FfiConverterZcashNetworkINSTANCE = FfiConverterZcashNetwork{}

func (c FfiConverterZcashNetwork) Lift(rb RustBufferI) ZcashNetwork {
	return LiftFromRustBuffer[ZcashNetwork](c, rb)
}

func (c FfiConverterZcashNetwork) Lower(value ZcashNetwork) C.RustBuffer {
	return LowerIntoRustBuffer[ZcashNetwork](c, value)
}
func (FfiConverterZcashNetwork) Read(reader io.Reader) ZcashNetwork {
	id := readInt32(reader)
	return ZcashNetwork(id)
}

func (FfiConverterZcashNetwork) Write(writer io.Writer, value ZcashNetwork) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerZcashNetwork struct{}

func (_ FfiDestroyerZcashNetwork) Destroy(value ZcashNetwork) {
}

type FfiConverterOptionalParticipantIdentifier struct{}

var FfiConverterOptionalParticipantIdentifierINSTANCE = FfiConverterOptionalParticipantIdentifier{}

func (c FfiConverterOptionalParticipantIdentifier) Lift(rb RustBufferI) *ParticipantIdentifier {
	return LiftFromRustBuffer[*ParticipantIdentifier](c, rb)
}

func (_ FfiConverterOptionalParticipantIdentifier) Read(reader io.Reader) *ParticipantIdentifier {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterParticipantIdentifierINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalParticipantIdentifier) Lower(value *ParticipantIdentifier) C.RustBuffer {
	return LowerIntoRustBuffer[*ParticipantIdentifier](c, value)
}

func (_ FfiConverterOptionalParticipantIdentifier) Write(writer io.Writer, value *ParticipantIdentifier) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterParticipantIdentifierINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalParticipantIdentifier struct{}

func (_ FfiDestroyerOptionalParticipantIdentifier) Destroy(value *ParticipantIdentifier) {
	if value != nil {
		FfiDestroyerParticipantIdentifier{}.Destroy(*value)
	}
}

type FfiConverterSequenceFrostSignatureShare struct{}

var FfiConverterSequenceFrostSignatureShareINSTANCE = FfiConverterSequenceFrostSignatureShare{}

func (c FfiConverterSequenceFrostSignatureShare) Lift(rb RustBufferI) []FrostSignatureShare {
	return LiftFromRustBuffer[[]FrostSignatureShare](c, rb)
}

func (c FfiConverterSequenceFrostSignatureShare) Read(reader io.Reader) []FrostSignatureShare {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]FrostSignatureShare, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterFrostSignatureShareINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceFrostSignatureShare) Lower(value []FrostSignatureShare) C.RustBuffer {
	return LowerIntoRustBuffer[[]FrostSignatureShare](c, value)
}

func (c FfiConverterSequenceFrostSignatureShare) Write(writer io.Writer, value []FrostSignatureShare) {
	if len(value) > math.MaxInt32 {
		panic("[]FrostSignatureShare is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterFrostSignatureShareINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceFrostSignatureShare struct{}

func (FfiDestroyerSequenceFrostSignatureShare) Destroy(sequence []FrostSignatureShare) {
	for _, value := range sequence {
		FfiDestroyerFrostSignatureShare{}.Destroy(value)
	}
}

type FfiConverterSequenceFrostSigningCommitments struct{}

var FfiConverterSequenceFrostSigningCommitmentsINSTANCE = FfiConverterSequenceFrostSigningCommitments{}

func (c FfiConverterSequenceFrostSigningCommitments) Lift(rb RustBufferI) []FrostSigningCommitments {
	return LiftFromRustBuffer[[]FrostSigningCommitments](c, rb)
}

func (c FfiConverterSequenceFrostSigningCommitments) Read(reader io.Reader) []FrostSigningCommitments {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]FrostSigningCommitments, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterFrostSigningCommitmentsINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceFrostSigningCommitments) Lower(value []FrostSigningCommitments) C.RustBuffer {
	return LowerIntoRustBuffer[[]FrostSigningCommitments](c, value)
}

func (c FfiConverterSequenceFrostSigningCommitments) Write(writer io.Writer, value []FrostSigningCommitments) {
	if len(value) > math.MaxInt32 {
		panic("[]FrostSigningCommitments is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterFrostSigningCommitmentsINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceFrostSigningCommitments struct{}

func (FfiDestroyerSequenceFrostSigningCommitments) Destroy(sequence []FrostSigningCommitments) {
	for _, value := range sequence {
		FfiDestroyerFrostSigningCommitments{}.Destroy(value)
	}
}

type FfiConverterSequenceParticipantIdentifier struct{}

var FfiConverterSequenceParticipantIdentifierINSTANCE = FfiConverterSequenceParticipantIdentifier{}

func (c FfiConverterSequenceParticipantIdentifier) Lift(rb RustBufferI) []ParticipantIdentifier {
	return LiftFromRustBuffer[[]ParticipantIdentifier](c, rb)
}

func (c FfiConverterSequenceParticipantIdentifier) Read(reader io.Reader) []ParticipantIdentifier {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ParticipantIdentifier, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterParticipantIdentifierINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceParticipantIdentifier) Lower(value []ParticipantIdentifier) C.RustBuffer {
	return LowerIntoRustBuffer[[]ParticipantIdentifier](c, value)
}

func (c FfiConverterSequenceParticipantIdentifier) Write(writer io.Writer, value []ParticipantIdentifier) {
	if len(value) > math.MaxInt32 {
		panic("[]ParticipantIdentifier is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterParticipantIdentifierINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceParticipantIdentifier struct{}

func (FfiDestroyerSequenceParticipantIdentifier) Destroy(sequence []ParticipantIdentifier) {
	for _, value := range sequence {
		FfiDestroyerParticipantIdentifier{}.Destroy(value)
	}
}

type FfiConverterMapParticipantIdentifierString struct{}

var FfiConverterMapParticipantIdentifierStringINSTANCE = FfiConverterMapParticipantIdentifierString{}

func (c FfiConverterMapParticipantIdentifierString) Lift(rb RustBufferI) map[ParticipantIdentifier]string {
	return LiftFromRustBuffer[map[ParticipantIdentifier]string](c, rb)
}

func (_ FfiConverterMapParticipantIdentifierString) Read(reader io.Reader) map[ParticipantIdentifier]string {
	result := make(map[ParticipantIdentifier]string)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterParticipantIdentifierINSTANCE.Read(reader)
		value := FfiConverterStringINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapParticipantIdentifierString) Lower(value map[ParticipantIdentifier]string) C.RustBuffer {
	return LowerIntoRustBuffer[map[ParticipantIdentifier]string](c, value)
}

func (_ FfiConverterMapParticipantIdentifierString) Write(writer io.Writer, mapValue map[ParticipantIdentifier]string) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[ParticipantIdentifier]string is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterParticipantIdentifierINSTANCE.Write(writer, key)
		FfiConverterStringINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapParticipantIdentifierString struct{}

func (_ FfiDestroyerMapParticipantIdentifierString) Destroy(mapValue map[ParticipantIdentifier]string) {
	for key, value := range mapValue {
		FfiDestroyerParticipantIdentifier{}.Destroy(key)
		FfiDestroyerString{}.Destroy(value)
	}
}

type FfiConverterMapParticipantIdentifierDkgRound1Package struct{}

var FfiConverterMapParticipantIdentifierDkgRound1PackageINSTANCE = FfiConverterMapParticipantIdentifierDkgRound1Package{}

func (c FfiConverterMapParticipantIdentifierDkgRound1Package) Lift(rb RustBufferI) map[ParticipantIdentifier]DkgRound1Package {
	return LiftFromRustBuffer[map[ParticipantIdentifier]DkgRound1Package](c, rb)
}

func (_ FfiConverterMapParticipantIdentifierDkgRound1Package) Read(reader io.Reader) map[ParticipantIdentifier]DkgRound1Package {
	result := make(map[ParticipantIdentifier]DkgRound1Package)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterParticipantIdentifierINSTANCE.Read(reader)
		value := FfiConverterDkgRound1PackageINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapParticipantIdentifierDkgRound1Package) Lower(value map[ParticipantIdentifier]DkgRound1Package) C.RustBuffer {
	return LowerIntoRustBuffer[map[ParticipantIdentifier]DkgRound1Package](c, value)
}

func (_ FfiConverterMapParticipantIdentifierDkgRound1Package) Write(writer io.Writer, mapValue map[ParticipantIdentifier]DkgRound1Package) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[ParticipantIdentifier]DkgRound1Package is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterParticipantIdentifierINSTANCE.Write(writer, key)
		FfiConverterDkgRound1PackageINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapParticipantIdentifierDkgRound1Package struct{}

func (_ FfiDestroyerMapParticipantIdentifierDkgRound1Package) Destroy(mapValue map[ParticipantIdentifier]DkgRound1Package) {
	for key, value := range mapValue {
		FfiDestroyerParticipantIdentifier{}.Destroy(key)
		FfiDestroyerDkgRound1Package{}.Destroy(value)
	}
}

type FfiConverterMapParticipantIdentifierDkgRound2Package struct{}

var FfiConverterMapParticipantIdentifierDkgRound2PackageINSTANCE = FfiConverterMapParticipantIdentifierDkgRound2Package{}

func (c FfiConverterMapParticipantIdentifierDkgRound2Package) Lift(rb RustBufferI) map[ParticipantIdentifier]DkgRound2Package {
	return LiftFromRustBuffer[map[ParticipantIdentifier]DkgRound2Package](c, rb)
}

func (_ FfiConverterMapParticipantIdentifierDkgRound2Package) Read(reader io.Reader) map[ParticipantIdentifier]DkgRound2Package {
	result := make(map[ParticipantIdentifier]DkgRound2Package)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterParticipantIdentifierINSTANCE.Read(reader)
		value := FfiConverterDkgRound2PackageINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapParticipantIdentifierDkgRound2Package) Lower(value map[ParticipantIdentifier]DkgRound2Package) C.RustBuffer {
	return LowerIntoRustBuffer[map[ParticipantIdentifier]DkgRound2Package](c, value)
}

func (_ FfiConverterMapParticipantIdentifierDkgRound2Package) Write(writer io.Writer, mapValue map[ParticipantIdentifier]DkgRound2Package) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[ParticipantIdentifier]DkgRound2Package is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterParticipantIdentifierINSTANCE.Write(writer, key)
		FfiConverterDkgRound2PackageINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapParticipantIdentifierDkgRound2Package struct{}

func (_ FfiDestroyerMapParticipantIdentifierDkgRound2Package) Destroy(mapValue map[ParticipantIdentifier]DkgRound2Package) {
	for key, value := range mapValue {
		FfiDestroyerParticipantIdentifier{}.Destroy(key)
		FfiDestroyerDkgRound2Package{}.Destroy(value)
	}
}

type FfiConverterMapParticipantIdentifierFrostSecretKeyShare struct{}

var FfiConverterMapParticipantIdentifierFrostSecretKeyShareINSTANCE = FfiConverterMapParticipantIdentifierFrostSecretKeyShare{}

func (c FfiConverterMapParticipantIdentifierFrostSecretKeyShare) Lift(rb RustBufferI) map[ParticipantIdentifier]FrostSecretKeyShare {
	return LiftFromRustBuffer[map[ParticipantIdentifier]FrostSecretKeyShare](c, rb)
}

func (_ FfiConverterMapParticipantIdentifierFrostSecretKeyShare) Read(reader io.Reader) map[ParticipantIdentifier]FrostSecretKeyShare {
	result := make(map[ParticipantIdentifier]FrostSecretKeyShare)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterParticipantIdentifierINSTANCE.Read(reader)
		value := FfiConverterFrostSecretKeyShareINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapParticipantIdentifierFrostSecretKeyShare) Lower(value map[ParticipantIdentifier]FrostSecretKeyShare) C.RustBuffer {
	return LowerIntoRustBuffer[map[ParticipantIdentifier]FrostSecretKeyShare](c, value)
}

func (_ FfiConverterMapParticipantIdentifierFrostSecretKeyShare) Write(writer io.Writer, mapValue map[ParticipantIdentifier]FrostSecretKeyShare) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[ParticipantIdentifier]FrostSecretKeyShare is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterParticipantIdentifierINSTANCE.Write(writer, key)
		FfiConverterFrostSecretKeyShareINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapParticipantIdentifierFrostSecretKeyShare struct{}

func (_ FfiDestroyerMapParticipantIdentifierFrostSecretKeyShare) Destroy(mapValue map[ParticipantIdentifier]FrostSecretKeyShare) {
	for key, value := range mapValue {
		FfiDestroyerParticipantIdentifier{}.Destroy(key)
		FfiDestroyerFrostSecretKeyShare{}.Destroy(value)
	}
}

func Aggregate(signingPackage FrostSigningPackage, signatureShares []FrostSignatureShare, pubkeyPackage FrostPublicKeyPackage, randomizer FrostRandomizer) (FrostSignature, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[CoordinationError](FfiConverterCoordinationError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_aggregate(FfiConverterFrostSigningPackageINSTANCE.Lower(signingPackage), FfiConverterSequenceFrostSignatureShareINSTANCE.Lower(signatureShares), FfiConverterFrostPublicKeyPackageINSTANCE.Lower(pubkeyPackage), FfiConverterFrostRandomizerINSTANCE.Lower(randomizer), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostSignature
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostSignatureINSTANCE.Lift(_uniffiRV), nil
	}
}

// returns Raw Signing commitnments using serde_json
// WARNING: The identifier you have in the `FrostSigningCommitments`
// is not an original field of `SigningCommitments`, we've included
// them as a nice-to-have.
func CommitmentToJson(commitment FrostSigningCommitments) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_commitment_to_json(FfiConverterFrostSigningCommitmentsINSTANCE.Lower(commitment), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), nil
	}
}

func FromHexString(hexString string) (FrostRandomizer, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_from_hex_string(FfiConverterStringINSTANCE.Lower(hexString), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostRandomizer
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostRandomizerINSTANCE.Lift(_uniffiRV), nil
	}
}

func GenerateNoncesAndCommitments(keyPackage FrostKeyPackage) (FirstRoundCommitment, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[Round1Error](FfiConverterRound1Error{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_generate_nonces_and_commitments(FfiConverterFrostKeyPackageINSTANCE.Lower(keyPackage), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FirstRoundCommitment
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFirstRoundCommitmentINSTANCE.Lift(_uniffiRV), nil
	}
}

func IdentifierFromJsonString(string string) *ParticipantIdentifier {
	return FfiConverterOptionalParticipantIdentifierINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_identifier_from_json_string(FfiConverterStringINSTANCE.Lower(string), _uniffiStatus),
		}
	}))
}

func IdentifierFromString(string string) (ParticipantIdentifier, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_identifier_from_string(FfiConverterStringINSTANCE.Lower(string), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue ParticipantIdentifier
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterParticipantIdentifierINSTANCE.Lift(_uniffiRV), nil
	}
}

func IdentifierFromUint16(unsignedUint uint16) (ParticipantIdentifier, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_identifier_from_uint16(FfiConverterUint16INSTANCE.Lower(unsignedUint), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue ParticipantIdentifier
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterParticipantIdentifierINSTANCE.Lift(_uniffiRV), nil
	}
}

func JsonToCommitment(commitmentJson string, identifier ParticipantIdentifier) (FrostSigningCommitments, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_json_to_commitment(FfiConverterStringINSTANCE.Lower(commitmentJson), FfiConverterParticipantIdentifierINSTANCE.Lower(identifier), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostSigningCommitments
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostSigningCommitmentsINSTANCE.Lift(_uniffiRV), nil
	}
}

func JsonToKeyPackage(keyPackageJson string) (FrostKeyPackage, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_json_to_key_package(FfiConverterStringINSTANCE.Lower(keyPackageJson), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostKeyPackage
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostKeyPackageINSTANCE.Lift(_uniffiRV), nil
	}
}

func JsonToPublicKeyPackage(publicKeyPackageJson string) (FrostPublicKeyPackage, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_json_to_public_key_package(FfiConverterStringINSTANCE.Lower(publicKeyPackageJson), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostPublicKeyPackage
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostPublicKeyPackageINSTANCE.Lift(_uniffiRV), nil
	}
}

func JsonToRandomizer(randomizerJson string) (FrostRandomizer, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_json_to_randomizer(FfiConverterStringINSTANCE.Lower(randomizerJson), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostRandomizer
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostRandomizerINSTANCE.Lift(_uniffiRV), nil
	}
}

func JsonToSignatureShare(signatureShareJson string, identifier ParticipantIdentifier) (FrostSignatureShare, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_json_to_signature_share(FfiConverterStringINSTANCE.Lower(signatureShareJson), FfiConverterParticipantIdentifierINSTANCE.Lower(identifier), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostSignatureShare
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostSignatureShareINSTANCE.Lift(_uniffiRV), nil
	}
}

func KeyPackageToJson(keyPackage FrostKeyPackage) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_key_package_to_json(FfiConverterFrostKeyPackageINSTANCE.Lower(keyPackage), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), nil
	}
}

func NewSigningPackage(message Message, commitments []FrostSigningCommitments) (FrostSigningPackage, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[CoordinationError](FfiConverterCoordinationError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_new_signing_package(FfiConverterMessageINSTANCE.Lower(message), FfiConverterSequenceFrostSigningCommitmentsINSTANCE.Lower(commitments), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostSigningPackage
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostSigningPackageINSTANCE.Lift(_uniffiRV), nil
	}
}

func Part1(participantIdentifier ParticipantIdentifier, maxSigners uint16, minSigners uint16) (*DkgPart1Result, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_func_part_1(FfiConverterParticipantIdentifierINSTANCE.Lower(participantIdentifier), FfiConverterUint16INSTANCE.Lower(maxSigners), FfiConverterUint16INSTANCE.Lower(minSigners), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *DkgPart1Result
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterDkgPart1ResultINSTANCE.Lift(_uniffiRV), nil
	}
}

// DKG Part 2
// receives a SecretPackage from round one generated by the same
// participant and kept in-memory (and secretly) until now.
// It also receives the round 1 packages corresponding to all the
// other participants **except** itself.
//
// Example: if P1, P2 and P3 are doing DKG, then when P1 runs part_2
// this will receive a secret generated by P1 in part_1 and the
// round 1 packages from P2 and P3. Everyone else has to do the same.
//
// For part_3 the P1 will send the round 2 packages generated here to
// the other participants P2 and P3 and should receive packages from
// P2 and P3.
func Part2(secretPackage *DkgRound1SecretPackage, round1Packages map[ParticipantIdentifier]DkgRound1Package) (*DkgPart2Result, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_func_part_2(FfiConverterDkgRound1SecretPackageINSTANCE.Lower(secretPackage), FfiConverterMapParticipantIdentifierDkgRound1PackageINSTANCE.Lower(round1Packages), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *DkgPart2Result
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterDkgPart2ResultINSTANCE.Lift(_uniffiRV), nil
	}
}

func Part3(secretPackage *DkgRound2SecretPackage, round1Packages map[ParticipantIdentifier]DkgRound1Package, round2Packages map[ParticipantIdentifier]DkgRound2Package) (DkgPart3Result, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_part_3(FfiConverterDkgRound2SecretPackageINSTANCE.Lower(secretPackage), FfiConverterMapParticipantIdentifierDkgRound1PackageINSTANCE.Lower(round1Packages), FfiConverterMapParticipantIdentifierDkgRound2PackageINSTANCE.Lower(round2Packages), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue DkgPart3Result
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterDkgPart3ResultINSTANCE.Lift(_uniffiRV), nil
	}
}

func PublicKeyPackageToJson(publicKeyPackage FrostPublicKeyPackage) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_public_key_package_to_json(FfiConverterFrostPublicKeyPackageINSTANCE.Lower(publicKeyPackage), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), nil
	}
}

func RandomizedParamsFromPublicKeyAndSigningPackage(publicKey FrostPublicKeyPackage, signingPackage FrostSigningPackage) (*FrostRandomizedParams, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_frost_uniffi_sdk_fn_func_randomized_params_from_public_key_and_signing_package(FfiConverterFrostPublicKeyPackageINSTANCE.Lower(publicKey), FfiConverterFrostSigningPackageINSTANCE.Lower(signingPackage), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *FrostRandomizedParams
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostRandomizedParamsINSTANCE.Lift(_uniffiRV), nil
	}
}

func RandomizerFromParams(randomizedParams *FrostRandomizedParams) (FrostRandomizer, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_randomizer_from_params(FfiConverterFrostRandomizedParamsINSTANCE.Lower(randomizedParams), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostRandomizer
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostRandomizerINSTANCE.Lift(_uniffiRV), nil
	}
}

func RandomizerToJson(randomizer FrostRandomizer) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_randomizer_to_json(FfiConverterFrostRandomizerINSTANCE.Lower(randomizer), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), nil
	}
}

func Sign(signingPackage FrostSigningPackage, nonces FrostSigningNonces, keyPackage FrostKeyPackage, randomizer FrostRandomizer) (FrostSignatureShare, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[Round2Error](FfiConverterRound2Error{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_sign(FfiConverterFrostSigningPackageINSTANCE.Lower(signingPackage), FfiConverterFrostSigningNoncesINSTANCE.Lower(nonces), FfiConverterFrostKeyPackageINSTANCE.Lower(keyPackage), FfiConverterFrostRandomizerINSTANCE.Lower(randomizer), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostSignatureShare
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostSignatureShareINSTANCE.Lift(_uniffiRV), nil
	}
}

func SignatureSharePackageToJson(signatureShare FrostSignatureShare) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_signature_share_package_to_json(FfiConverterFrostSignatureShareINSTANCE.Lower(signatureShare), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), nil
	}
}

func TrustedDealerKeygenFrom(configuration Configuration) (TrustedKeyGeneration, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_trusted_dealer_keygen_from(FfiConverterConfigurationINSTANCE.Lower(configuration), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue TrustedKeyGeneration
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTrustedKeyGenerationINSTANCE.Lift(_uniffiRV), nil
	}
}

func TrustedDealerKeygenWithIdentifiers(configuration Configuration, participants ParticipantList) (TrustedKeyGeneration, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_trusted_dealer_keygen_with_identifiers(FfiConverterConfigurationINSTANCE.Lower(configuration), FfiConverterParticipantListINSTANCE.Lower(participants), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue TrustedKeyGeneration
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTrustedKeyGenerationINSTANCE.Lift(_uniffiRV), nil
	}
}

func ValidateConfig(config Configuration) error {
	_, _uniffiErr := rustCallWithError[ConfigurationError](FfiConverterConfigurationError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_frost_uniffi_sdk_fn_func_validate_config(FfiConverterConfigurationINSTANCE.Lower(config), _uniffiStatus)
		return false
	})
	return _uniffiErr.AsError()
}

func VerifyAndGetKeyPackageFrom(secretShare FrostSecretKeyShare) (FrostKeyPackage, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[FrostError](FfiConverterFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer{
			inner: C.uniffi_frost_uniffi_sdk_fn_func_verify_and_get_key_package_from(FfiConverterFrostSecretKeyShareINSTANCE.Lower(secretShare), _uniffiStatus),
		}
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostKeyPackage
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterFrostKeyPackageINSTANCE.Lift(_uniffiRV), nil
	}
}

func VerifyRandomizedSignature(randomizer FrostRandomizer, message Message, signature FrostSignature, pubkey FrostPublicKeyPackage) error {
	_, _uniffiErr := rustCallWithError[FrostSignatureVerificationError](FfiConverterFrostSignatureVerificationError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_frost_uniffi_sdk_fn_func_verify_randomized_signature(FfiConverterFrostRandomizerINSTANCE.Lower(randomizer), FfiConverterMessageINSTANCE.Lower(message), FfiConverterFrostSignatureINSTANCE.Lower(signature), FfiConverterFrostPublicKeyPackageINSTANCE.Lower(pubkey), _uniffiStatus)
		return false
	})
	return _uniffiErr.AsError()
}

func VerifySignature(message Message, signature FrostSignature, pubkey FrostPublicKeyPackage) error {
	_, _uniffiErr := rustCallWithError[FrostSignatureVerificationError](FfiConverterFrostSignatureVerificationError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_frost_uniffi_sdk_fn_func_verify_signature(FfiConverterMessageINSTANCE.Lower(message), FfiConverterFrostSignatureINSTANCE.Lower(signature), FfiConverterFrostPublicKeyPackageINSTANCE.Lower(pubkey), _uniffiStatus)
		return false
	})
	return _uniffiErr.AsError()
}
