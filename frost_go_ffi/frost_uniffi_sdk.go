package frost_uniffi_sdk

// #include <frost_go_ffi.h>
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"
)

type RustBuffer = C.RustBuffer

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() int
	Capacity() int
}

func RustBufferFromExternal(b RustBufferI) RustBuffer {
	return RustBuffer{
		capacity: C.int(b.Capacity()),
		len:      C.int(b.Len()),
		data:     (*C.uchar)(b.Data()),
	}
}

func (cb RustBuffer) Capacity() int {
	return int(cb.capacity)
}

func (cb RustBuffer) Len() int {
	return int(cb.len)
}

func (cb RustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.data)
}

func (cb RustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.data), C.int(cb.len))
	return bytes.NewReader(b)
}

func (cb RustBuffer) Free() {
	rustCall(func(status *C.RustCallStatus) bool {
		C.ffi_frost_uniffi_sdk_rustbuffer_free(cb, status)
		return false
	})
}

func (cb RustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.data), C.int(cb.len))
}

func stringToRustBuffer(str string) RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) RustBuffer {
	if len(b) == 0 {
		return RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes{
		len:  C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}

	return rustCall(func(status *C.RustCallStatus) RustBuffer {
		return C.ffi_frost_uniffi_sdk_rustbuffer_from_bytes(foreign, status)
	})
}

type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) RustBuffer
}

type FfiConverter[GoType any, FfiType any] interface {
	Lift(value FfiType) GoType
	Lower(value GoType) FfiType
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

type FfiRustBufConverter[GoType any, FfiType any] interface {
	FfiConverter[GoType, FfiType]
	BufReader[GoType]
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) RustBuffer {
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

func rustCallWithError[U any](converter BufLifter[error], callback func(*C.RustCallStatus) U) (U, error) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)

	return returnValue, err
}

func checkCallStatus(converter BufLifter[error], status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		return converter.Lift(status.errorBuf)
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError(nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
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
	bindingsContractVersion := 24
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_frost_uniffi_sdk_uniffi_contract_version(uniffiStatus)
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("frost_uniffi_sdk: UniFFI contract version mismatch")
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_aggregate(uniffiStatus)
		})
		if checksum != 46119 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_aggregate: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_generate_nonces_and_commitments(uniffiStatus)
		})
		if checksum != 47101 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_generate_nonces_and_commitments: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_new_signing_package(uniffiStatus)
		})
		if checksum != 50111 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_new_signing_package: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_sign(uniffiStatus)
		})
		if checksum != 48101 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_sign: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_from(uniffiStatus)
		})
		if checksum != 27691 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_from: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_with_identifiers(uniffiStatus)
		})
		if checksum != 46297 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_trusted_dealer_keygen_with_identifiers: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_validate_config(uniffiStatus)
		})
		if checksum != 26688 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_validate_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_verify_and_get_key_package_from(uniffiStatus)
		})
		if checksum != 16387 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_verify_and_get_key_package_from: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_frost_uniffi_sdk_checksum_func_verify_signature(uniffiStatus)
		})
		if checksum != 13620 {
			// If this happens try cleaning and rebuilding your project
			panic("frost_uniffi_sdk: uniffi_frost_uniffi_sdk_checksum_func_verify_signature: UniFFI API checksum mismatch")
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
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) RustBuffer {
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

func (c FfiConverterBytes) Lower(value []byte) RustBuffer {
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
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading []byte, expected %d, read %d", length, read_length))
	}
	return buffer
}

type FfiDestroyerBytes struct{}

func (FfiDestroyerBytes) Destroy(_ []byte) {}

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

type FfiConverterTypeConfiguration struct{}

var FfiConverterTypeConfigurationINSTANCE = FfiConverterTypeConfiguration{}

func (c FfiConverterTypeConfiguration) Lift(rb RustBufferI) Configuration {
	return LiftFromRustBuffer[Configuration](c, rb)
}

func (c FfiConverterTypeConfiguration) Read(reader io.Reader) Configuration {
	return Configuration{
		FfiConverterUint16INSTANCE.Read(reader),
		FfiConverterUint16INSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeConfiguration) Lower(value Configuration) RustBuffer {
	return LowerIntoRustBuffer[Configuration](c, value)
}

func (c FfiConverterTypeConfiguration) Write(writer io.Writer, value Configuration) {
	FfiConverterUint16INSTANCE.Write(writer, value.MinSigners)
	FfiConverterUint16INSTANCE.Write(writer, value.MaxSigners)
	FfiConverterBytesINSTANCE.Write(writer, value.Secret)
}

type FfiDestroyerTypeConfiguration struct{}

func (_ FfiDestroyerTypeConfiguration) Destroy(value Configuration) {
	value.Destroy()
}

type FirstRoundCommitment struct {
	Nonces      FrostSigningNonces
	Commitments FrostSigningCommitments
}

func (r *FirstRoundCommitment) Destroy() {
	FfiDestroyerTypeFrostSigningNonces{}.Destroy(r.Nonces)
	FfiDestroyerTypeFrostSigningCommitments{}.Destroy(r.Commitments)
}

type FfiConverterTypeFirstRoundCommitment struct{}

var FfiConverterTypeFirstRoundCommitmentINSTANCE = FfiConverterTypeFirstRoundCommitment{}

func (c FfiConverterTypeFirstRoundCommitment) Lift(rb RustBufferI) FirstRoundCommitment {
	return LiftFromRustBuffer[FirstRoundCommitment](c, rb)
}

func (c FfiConverterTypeFirstRoundCommitment) Read(reader io.Reader) FirstRoundCommitment {
	return FirstRoundCommitment{
		FfiConverterTypeFrostSigningNoncesINSTANCE.Read(reader),
		FfiConverterTypeFrostSigningCommitmentsINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFirstRoundCommitment) Lower(value FirstRoundCommitment) RustBuffer {
	return LowerIntoRustBuffer[FirstRoundCommitment](c, value)
}

func (c FfiConverterTypeFirstRoundCommitment) Write(writer io.Writer, value FirstRoundCommitment) {
	FfiConverterTypeFrostSigningNoncesINSTANCE.Write(writer, value.Nonces)
	FfiConverterTypeFrostSigningCommitmentsINSTANCE.Write(writer, value.Commitments)
}

type FfiDestroyerTypeFirstRoundCommitment struct{}

func (_ FfiDestroyerTypeFirstRoundCommitment) Destroy(value FirstRoundCommitment) {
	value.Destroy()
}

type FrostKeyPackage struct {
	Identifier string
	Data       []byte
}

func (r *FrostKeyPackage) Destroy() {
	FfiDestroyerString{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterTypeFrostKeyPackage struct{}

var FfiConverterTypeFrostKeyPackageINSTANCE = FfiConverterTypeFrostKeyPackage{}

func (c FfiConverterTypeFrostKeyPackage) Lift(rb RustBufferI) FrostKeyPackage {
	return LiftFromRustBuffer[FrostKeyPackage](c, rb)
}

func (c FfiConverterTypeFrostKeyPackage) Read(reader io.Reader) FrostKeyPackage {
	return FrostKeyPackage{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFrostKeyPackage) Lower(value FrostKeyPackage) RustBuffer {
	return LowerIntoRustBuffer[FrostKeyPackage](c, value)
}

func (c FfiConverterTypeFrostKeyPackage) Write(writer io.Writer, value FrostKeyPackage) {
	FfiConverterStringINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeFrostKeyPackage struct{}

func (_ FfiDestroyerTypeFrostKeyPackage) Destroy(value FrostKeyPackage) {
	value.Destroy()
}

type FrostPublicKeyPackage struct {
	VerifyingShares map[ParticipantIdentifier]string
	VerifyingKey    string
}

func (r *FrostPublicKeyPackage) Destroy() {
	FfiDestroyerMapTypeParticipantIdentifierString{}.Destroy(r.VerifyingShares)
	FfiDestroyerString{}.Destroy(r.VerifyingKey)
}

type FfiConverterTypeFrostPublicKeyPackage struct{}

var FfiConverterTypeFrostPublicKeyPackageINSTANCE = FfiConverterTypeFrostPublicKeyPackage{}

func (c FfiConverterTypeFrostPublicKeyPackage) Lift(rb RustBufferI) FrostPublicKeyPackage {
	return LiftFromRustBuffer[FrostPublicKeyPackage](c, rb)
}

func (c FfiConverterTypeFrostPublicKeyPackage) Read(reader io.Reader) FrostPublicKeyPackage {
	return FrostPublicKeyPackage{
		FfiConverterMapTypeParticipantIdentifierStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFrostPublicKeyPackage) Lower(value FrostPublicKeyPackage) RustBuffer {
	return LowerIntoRustBuffer[FrostPublicKeyPackage](c, value)
}

func (c FfiConverterTypeFrostPublicKeyPackage) Write(writer io.Writer, value FrostPublicKeyPackage) {
	FfiConverterMapTypeParticipantIdentifierStringINSTANCE.Write(writer, value.VerifyingShares)
	FfiConverterStringINSTANCE.Write(writer, value.VerifyingKey)
}

type FfiDestroyerTypeFrostPublicKeyPackage struct{}

func (_ FfiDestroyerTypeFrostPublicKeyPackage) Destroy(value FrostPublicKeyPackage) {
	value.Destroy()
}

type FrostSecretKeyShare struct {
	Identifier ParticipantIdentifier
	Data       []byte
}

func (r *FrostSecretKeyShare) Destroy() {
	FfiDestroyerTypeParticipantIdentifier{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterTypeFrostSecretKeyShare struct{}

var FfiConverterTypeFrostSecretKeyShareINSTANCE = FfiConverterTypeFrostSecretKeyShare{}

func (c FfiConverterTypeFrostSecretKeyShare) Lift(rb RustBufferI) FrostSecretKeyShare {
	return LiftFromRustBuffer[FrostSecretKeyShare](c, rb)
}

func (c FfiConverterTypeFrostSecretKeyShare) Read(reader io.Reader) FrostSecretKeyShare {
	return FrostSecretKeyShare{
		FfiConverterTypeParticipantIdentifierINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFrostSecretKeyShare) Lower(value FrostSecretKeyShare) RustBuffer {
	return LowerIntoRustBuffer[FrostSecretKeyShare](c, value)
}

func (c FfiConverterTypeFrostSecretKeyShare) Write(writer io.Writer, value FrostSecretKeyShare) {
	FfiConverterTypeParticipantIdentifierINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeFrostSecretKeyShare struct{}

func (_ FfiDestroyerTypeFrostSecretKeyShare) Destroy(value FrostSecretKeyShare) {
	value.Destroy()
}

type FrostSignature struct {
	Data []byte
}

func (r *FrostSignature) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterTypeFrostSignature struct{}

var FfiConverterTypeFrostSignatureINSTANCE = FfiConverterTypeFrostSignature{}

func (c FfiConverterTypeFrostSignature) Lift(rb RustBufferI) FrostSignature {
	return LiftFromRustBuffer[FrostSignature](c, rb)
}

func (c FfiConverterTypeFrostSignature) Read(reader io.Reader) FrostSignature {
	return FrostSignature{
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFrostSignature) Lower(value FrostSignature) RustBuffer {
	return LowerIntoRustBuffer[FrostSignature](c, value)
}

func (c FfiConverterTypeFrostSignature) Write(writer io.Writer, value FrostSignature) {
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeFrostSignature struct{}

func (_ FfiDestroyerTypeFrostSignature) Destroy(value FrostSignature) {
	value.Destroy()
}

type FrostSignatureShare struct {
	Identifier ParticipantIdentifier
	Data       []byte
}

func (r *FrostSignatureShare) Destroy() {
	FfiDestroyerTypeParticipantIdentifier{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterTypeFrostSignatureShare struct{}

var FfiConverterTypeFrostSignatureShareINSTANCE = FfiConverterTypeFrostSignatureShare{}

func (c FfiConverterTypeFrostSignatureShare) Lift(rb RustBufferI) FrostSignatureShare {
	return LiftFromRustBuffer[FrostSignatureShare](c, rb)
}

func (c FfiConverterTypeFrostSignatureShare) Read(reader io.Reader) FrostSignatureShare {
	return FrostSignatureShare{
		FfiConverterTypeParticipantIdentifierINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFrostSignatureShare) Lower(value FrostSignatureShare) RustBuffer {
	return LowerIntoRustBuffer[FrostSignatureShare](c, value)
}

func (c FfiConverterTypeFrostSignatureShare) Write(writer io.Writer, value FrostSignatureShare) {
	FfiConverterTypeParticipantIdentifierINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeFrostSignatureShare struct{}

func (_ FfiDestroyerTypeFrostSignatureShare) Destroy(value FrostSignatureShare) {
	value.Destroy()
}

type FrostSigningCommitments struct {
	Identifier ParticipantIdentifier
	Data       []byte
}

func (r *FrostSigningCommitments) Destroy() {
	FfiDestroyerTypeParticipantIdentifier{}.Destroy(r.Identifier)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterTypeFrostSigningCommitments struct{}

var FfiConverterTypeFrostSigningCommitmentsINSTANCE = FfiConverterTypeFrostSigningCommitments{}

func (c FfiConverterTypeFrostSigningCommitments) Lift(rb RustBufferI) FrostSigningCommitments {
	return LiftFromRustBuffer[FrostSigningCommitments](c, rb)
}

func (c FfiConverterTypeFrostSigningCommitments) Read(reader io.Reader) FrostSigningCommitments {
	return FrostSigningCommitments{
		FfiConverterTypeParticipantIdentifierINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFrostSigningCommitments) Lower(value FrostSigningCommitments) RustBuffer {
	return LowerIntoRustBuffer[FrostSigningCommitments](c, value)
}

func (c FfiConverterTypeFrostSigningCommitments) Write(writer io.Writer, value FrostSigningCommitments) {
	FfiConverterTypeParticipantIdentifierINSTANCE.Write(writer, value.Identifier)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeFrostSigningCommitments struct{}

func (_ FfiDestroyerTypeFrostSigningCommitments) Destroy(value FrostSigningCommitments) {
	value.Destroy()
}

type FrostSigningNonces struct {
	Data []byte
}

func (r *FrostSigningNonces) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterTypeFrostSigningNonces struct{}

var FfiConverterTypeFrostSigningNoncesINSTANCE = FfiConverterTypeFrostSigningNonces{}

func (c FfiConverterTypeFrostSigningNonces) Lift(rb RustBufferI) FrostSigningNonces {
	return LiftFromRustBuffer[FrostSigningNonces](c, rb)
}

func (c FfiConverterTypeFrostSigningNonces) Read(reader io.Reader) FrostSigningNonces {
	return FrostSigningNonces{
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFrostSigningNonces) Lower(value FrostSigningNonces) RustBuffer {
	return LowerIntoRustBuffer[FrostSigningNonces](c, value)
}

func (c FfiConverterTypeFrostSigningNonces) Write(writer io.Writer, value FrostSigningNonces) {
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeFrostSigningNonces struct{}

func (_ FfiDestroyerTypeFrostSigningNonces) Destroy(value FrostSigningNonces) {
	value.Destroy()
}

type FrostSigningPackage struct {
	Data []byte
}

func (r *FrostSigningPackage) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterTypeFrostSigningPackage struct{}

var FfiConverterTypeFrostSigningPackageINSTANCE = FfiConverterTypeFrostSigningPackage{}

func (c FfiConverterTypeFrostSigningPackage) Lift(rb RustBufferI) FrostSigningPackage {
	return LiftFromRustBuffer[FrostSigningPackage](c, rb)
}

func (c FfiConverterTypeFrostSigningPackage) Read(reader io.Reader) FrostSigningPackage {
	return FrostSigningPackage{
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFrostSigningPackage) Lower(value FrostSigningPackage) RustBuffer {
	return LowerIntoRustBuffer[FrostSigningPackage](c, value)
}

func (c FfiConverterTypeFrostSigningPackage) Write(writer io.Writer, value FrostSigningPackage) {
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeFrostSigningPackage struct{}

func (_ FfiDestroyerTypeFrostSigningPackage) Destroy(value FrostSigningPackage) {
	value.Destroy()
}

type Message struct {
	Data []byte
}

func (r *Message) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterTypeMessage struct{}

var FfiConverterTypeMessageINSTANCE = FfiConverterTypeMessage{}

func (c FfiConverterTypeMessage) Lift(rb RustBufferI) Message {
	return LiftFromRustBuffer[Message](c, rb)
}

func (c FfiConverterTypeMessage) Read(reader io.Reader) Message {
	return Message{
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMessage) Lower(value Message) RustBuffer {
	return LowerIntoRustBuffer[Message](c, value)
}

func (c FfiConverterTypeMessage) Write(writer io.Writer, value Message) {
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeMessage struct{}

func (_ FfiDestroyerTypeMessage) Destroy(value Message) {
	value.Destroy()
}

type ParticipantIdentifier struct {
	Data string
}

func (r *ParticipantIdentifier) Destroy() {
	FfiDestroyerString{}.Destroy(r.Data)
}

type FfiConverterTypeParticipantIdentifier struct{}

var FfiConverterTypeParticipantIdentifierINSTANCE = FfiConverterTypeParticipantIdentifier{}

func (c FfiConverterTypeParticipantIdentifier) Lift(rb RustBufferI) ParticipantIdentifier {
	return LiftFromRustBuffer[ParticipantIdentifier](c, rb)
}

func (c FfiConverterTypeParticipantIdentifier) Read(reader io.Reader) ParticipantIdentifier {
	return ParticipantIdentifier{
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeParticipantIdentifier) Lower(value ParticipantIdentifier) RustBuffer {
	return LowerIntoRustBuffer[ParticipantIdentifier](c, value)
}

func (c FfiConverterTypeParticipantIdentifier) Write(writer io.Writer, value ParticipantIdentifier) {
	FfiConverterStringINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeParticipantIdentifier struct{}

func (_ FfiDestroyerTypeParticipantIdentifier) Destroy(value ParticipantIdentifier) {
	value.Destroy()
}

type ParticipantList struct {
	Identifiers []ParticipantIdentifier
}

func (r *ParticipantList) Destroy() {
	FfiDestroyerSequenceTypeParticipantIdentifier{}.Destroy(r.Identifiers)
}

type FfiConverterTypeParticipantList struct{}

var FfiConverterTypeParticipantListINSTANCE = FfiConverterTypeParticipantList{}

func (c FfiConverterTypeParticipantList) Lift(rb RustBufferI) ParticipantList {
	return LiftFromRustBuffer[ParticipantList](c, rb)
}

func (c FfiConverterTypeParticipantList) Read(reader io.Reader) ParticipantList {
	return ParticipantList{
		FfiConverterSequenceTypeParticipantIdentifierINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeParticipantList) Lower(value ParticipantList) RustBuffer {
	return LowerIntoRustBuffer[ParticipantList](c, value)
}

func (c FfiConverterTypeParticipantList) Write(writer io.Writer, value ParticipantList) {
	FfiConverterSequenceTypeParticipantIdentifierINSTANCE.Write(writer, value.Identifiers)
}

type FfiDestroyerTypeParticipantList struct{}

func (_ FfiDestroyerTypeParticipantList) Destroy(value ParticipantList) {
	value.Destroy()
}

type TrustedKeyGeneration struct {
	SecretShares     map[ParticipantIdentifier]FrostSecretKeyShare
	PublicKeyPackage FrostPublicKeyPackage
}

func (r *TrustedKeyGeneration) Destroy() {
	FfiDestroyerMapTypeParticipantIdentifierTypeFrostSecretKeyShare{}.Destroy(r.SecretShares)
	FfiDestroyerTypeFrostPublicKeyPackage{}.Destroy(r.PublicKeyPackage)
}

type FfiConverterTypeTrustedKeyGeneration struct{}

var FfiConverterTypeTrustedKeyGenerationINSTANCE = FfiConverterTypeTrustedKeyGeneration{}

func (c FfiConverterTypeTrustedKeyGeneration) Lift(rb RustBufferI) TrustedKeyGeneration {
	return LiftFromRustBuffer[TrustedKeyGeneration](c, rb)
}

func (c FfiConverterTypeTrustedKeyGeneration) Read(reader io.Reader) TrustedKeyGeneration {
	return TrustedKeyGeneration{
		FfiConverterMapTypeParticipantIdentifierTypeFrostSecretKeyShareINSTANCE.Read(reader),
		FfiConverterTypeFrostPublicKeyPackageINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTrustedKeyGeneration) Lower(value TrustedKeyGeneration) RustBuffer {
	return LowerIntoRustBuffer[TrustedKeyGeneration](c, value)
}

func (c FfiConverterTypeTrustedKeyGeneration) Write(writer io.Writer, value TrustedKeyGeneration) {
	FfiConverterMapTypeParticipantIdentifierTypeFrostSecretKeyShareINSTANCE.Write(writer, value.SecretShares)
	FfiConverterTypeFrostPublicKeyPackageINSTANCE.Write(writer, value.PublicKeyPackage)
}

type FfiDestroyerTypeTrustedKeyGeneration struct{}

func (_ FfiDestroyerTypeTrustedKeyGeneration) Destroy(value TrustedKeyGeneration) {
	value.Destroy()
}

type ConfigurationError struct {
	err error
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
	return &ConfigurationError{
		err: &ConfigurationErrorInvalidMaxSigners{},
	}
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
	return &ConfigurationError{
		err: &ConfigurationErrorInvalidMinSigners{},
	}
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
	return &ConfigurationError{
		err: &ConfigurationErrorInvalidIdentifier{},
	}
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
	return &ConfigurationError{
		err: &ConfigurationErrorUnknownError{},
	}
}

func (err ConfigurationErrorUnknownError) Error() string {
	return fmt.Sprint("UnknownError")
}

func (self ConfigurationErrorUnknownError) Is(target error) bool {
	return target == ErrConfigurationErrorUnknownError
}

type FfiConverterTypeConfigurationError struct{}

var FfiConverterTypeConfigurationErrorINSTANCE = FfiConverterTypeConfigurationError{}

func (c FfiConverterTypeConfigurationError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeConfigurationError) Lower(value *ConfigurationError) RustBuffer {
	return LowerIntoRustBuffer[*ConfigurationError](c, value)
}

func (c FfiConverterTypeConfigurationError) Read(reader io.Reader) error {
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
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeConfigurationError.Read()", errorID))
	}
}

func (c FfiConverterTypeConfigurationError) Write(writer io.Writer, value *ConfigurationError) {
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
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeConfigurationError.Write", value))
	}
}

type CoordinationError struct {
	err error
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

// Variant structs
type CoordinationErrorFailedToCreateSigningPackage struct {
}

func NewCoordinationErrorFailedToCreateSigningPackage() *CoordinationError {
	return &CoordinationError{
		err: &CoordinationErrorFailedToCreateSigningPackage{},
	}
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
	return &CoordinationError{
		err: &CoordinationErrorInvalidSigningCommitment{},
	}
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
	return &CoordinationError{
		err: &CoordinationErrorIdentifierDeserializationError{},
	}
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
	return &CoordinationError{
		err: &CoordinationErrorSigningPackageSerializationError{},
	}
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
	return &CoordinationError{
		err: &CoordinationErrorSignatureShareDeserializationError{},
	}
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
	return &CoordinationError{
		err: &CoordinationErrorPublicKeyPackageDeserializationError{},
	}
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
	return &CoordinationError{
		err: &CoordinationErrorSignatureShareAggregationFailed{
			Message: message,
		},
	}
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

type FfiConverterTypeCoordinationError struct{}

var FfiConverterTypeCoordinationErrorINSTANCE = FfiConverterTypeCoordinationError{}

func (c FfiConverterTypeCoordinationError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeCoordinationError) Lower(value *CoordinationError) RustBuffer {
	return LowerIntoRustBuffer[*CoordinationError](c, value)
}

func (c FfiConverterTypeCoordinationError) Read(reader io.Reader) error {
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
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeCoordinationError.Read()", errorID))
	}
}

func (c FfiConverterTypeCoordinationError) Write(writer io.Writer, value *CoordinationError) {
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
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeCoordinationError.Write", value))
	}
}

type FrostError struct {
	err error
}

func (err FrostError) Error() string {
	return fmt.Sprintf("FrostError: %s", err.err.Error())
}

func (err FrostError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrFrostErrorSerializationError = fmt.Errorf("FrostErrorSerializationError")
var ErrFrostErrorDeserializationError = fmt.Errorf("FrostErrorDeserializationError")
var ErrFrostErrorInvalidKeyPackage = fmt.Errorf("FrostErrorInvalidKeyPackage")
var ErrFrostErrorInvalidSecretKey = fmt.Errorf("FrostErrorInvalidSecretKey")
var ErrFrostErrorUnknownIdentifier = fmt.Errorf("FrostErrorUnknownIdentifier")

// Variant structs
type FrostErrorSerializationError struct {
}

func NewFrostErrorSerializationError() *FrostError {
	return &FrostError{
		err: &FrostErrorSerializationError{},
	}
}

func (err FrostErrorSerializationError) Error() string {
	return fmt.Sprint("SerializationError")
}

func (self FrostErrorSerializationError) Is(target error) bool {
	return target == ErrFrostErrorSerializationError
}

type FrostErrorDeserializationError struct {
}

func NewFrostErrorDeserializationError() *FrostError {
	return &FrostError{
		err: &FrostErrorDeserializationError{},
	}
}

func (err FrostErrorDeserializationError) Error() string {
	return fmt.Sprint("DeserializationError")
}

func (self FrostErrorDeserializationError) Is(target error) bool {
	return target == ErrFrostErrorDeserializationError
}

type FrostErrorInvalidKeyPackage struct {
}

func NewFrostErrorInvalidKeyPackage() *FrostError {
	return &FrostError{
		err: &FrostErrorInvalidKeyPackage{},
	}
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
	return &FrostError{
		err: &FrostErrorInvalidSecretKey{},
	}
}

func (err FrostErrorInvalidSecretKey) Error() string {
	return fmt.Sprint("InvalidSecretKey")
}

func (self FrostErrorInvalidSecretKey) Is(target error) bool {
	return target == ErrFrostErrorInvalidSecretKey
}

type FrostErrorUnknownIdentifier struct {
}

func NewFrostErrorUnknownIdentifier() *FrostError {
	return &FrostError{
		err: &FrostErrorUnknownIdentifier{},
	}
}

func (err FrostErrorUnknownIdentifier) Error() string {
	return fmt.Sprint("UnknownIdentifier")
}

func (self FrostErrorUnknownIdentifier) Is(target error) bool {
	return target == ErrFrostErrorUnknownIdentifier
}

type FfiConverterTypeFrostError struct{}

var FfiConverterTypeFrostErrorINSTANCE = FfiConverterTypeFrostError{}

func (c FfiConverterTypeFrostError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeFrostError) Lower(value *FrostError) RustBuffer {
	return LowerIntoRustBuffer[*FrostError](c, value)
}

func (c FfiConverterTypeFrostError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &FrostError{&FrostErrorSerializationError{}}
	case 2:
		return &FrostError{&FrostErrorDeserializationError{}}
	case 3:
		return &FrostError{&FrostErrorInvalidKeyPackage{}}
	case 4:
		return &FrostError{&FrostErrorInvalidSecretKey{}}
	case 5:
		return &FrostError{&FrostErrorUnknownIdentifier{}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeFrostError.Read()", errorID))
	}
}

func (c FfiConverterTypeFrostError) Write(writer io.Writer, value *FrostError) {
	switch variantValue := value.err.(type) {
	case *FrostErrorSerializationError:
		writeInt32(writer, 1)
	case *FrostErrorDeserializationError:
		writeInt32(writer, 2)
	case *FrostErrorInvalidKeyPackage:
		writeInt32(writer, 3)
	case *FrostErrorInvalidSecretKey:
		writeInt32(writer, 4)
	case *FrostErrorUnknownIdentifier:
		writeInt32(writer, 5)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeFrostError.Write", value))
	}
}

type FrostSignatureVerificationError struct {
	err error
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
	return &FrostSignatureVerificationError{
		err: &FrostSignatureVerificationErrorInvalidPublicKeyPackage{},
	}
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
	return &FrostSignatureVerificationError{
		err: &FrostSignatureVerificationErrorValidationFailed{
			Reason: reason,
		},
	}
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

type FfiConverterTypeFrostSignatureVerificationError struct{}

var FfiConverterTypeFrostSignatureVerificationErrorINSTANCE = FfiConverterTypeFrostSignatureVerificationError{}

func (c FfiConverterTypeFrostSignatureVerificationError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeFrostSignatureVerificationError) Lower(value *FrostSignatureVerificationError) RustBuffer {
	return LowerIntoRustBuffer[*FrostSignatureVerificationError](c, value)
}

func (c FfiConverterTypeFrostSignatureVerificationError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &FrostSignatureVerificationError{&FrostSignatureVerificationErrorInvalidPublicKeyPackage{}}
	case 2:
		return &FrostSignatureVerificationError{&FrostSignatureVerificationErrorValidationFailed{
			Reason: FfiConverterStringINSTANCE.Read(reader),
		}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeFrostSignatureVerificationError.Read()", errorID))
	}
}

func (c FfiConverterTypeFrostSignatureVerificationError) Write(writer io.Writer, value *FrostSignatureVerificationError) {
	switch variantValue := value.err.(type) {
	case *FrostSignatureVerificationErrorInvalidPublicKeyPackage:
		writeInt32(writer, 1)
	case *FrostSignatureVerificationErrorValidationFailed:
		writeInt32(writer, 2)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Reason)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeFrostSignatureVerificationError.Write", value))
	}
}

type Round1Error struct {
	err error
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
	return &Round1Error{
		err: &Round1ErrorInvalidKeyPackage{},
	}
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
	return &Round1Error{
		err: &Round1ErrorNonceSerializationError{},
	}
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
	return &Round1Error{
		err: &Round1ErrorCommitmentSerializationError{},
	}
}

func (err Round1ErrorCommitmentSerializationError) Error() string {
	return fmt.Sprint("CommitmentSerializationError")
}

func (self Round1ErrorCommitmentSerializationError) Is(target error) bool {
	return target == ErrRound1ErrorCommitmentSerializationError
}

type FfiConverterTypeRound1Error struct{}

var FfiConverterTypeRound1ErrorINSTANCE = FfiConverterTypeRound1Error{}

func (c FfiConverterTypeRound1Error) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeRound1Error) Lower(value *Round1Error) RustBuffer {
	return LowerIntoRustBuffer[*Round1Error](c, value)
}

func (c FfiConverterTypeRound1Error) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &Round1Error{&Round1ErrorInvalidKeyPackage{}}
	case 2:
		return &Round1Error{&Round1ErrorNonceSerializationError{}}
	case 3:
		return &Round1Error{&Round1ErrorCommitmentSerializationError{}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeRound1Error.Read()", errorID))
	}
}

func (c FfiConverterTypeRound1Error) Write(writer io.Writer, value *Round1Error) {
	switch variantValue := value.err.(type) {
	case *Round1ErrorInvalidKeyPackage:
		writeInt32(writer, 1)
	case *Round1ErrorNonceSerializationError:
		writeInt32(writer, 2)
	case *Round1ErrorCommitmentSerializationError:
		writeInt32(writer, 3)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeRound1Error.Write", value))
	}
}

type Round2Error struct {
	err error
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

// Variant structs
type Round2ErrorInvalidKeyPackage struct {
}

func NewRound2ErrorInvalidKeyPackage() *Round2Error {
	return &Round2Error{
		err: &Round2ErrorInvalidKeyPackage{},
	}
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
	return &Round2Error{
		err: &Round2ErrorNonceSerializationError{},
	}
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
	return &Round2Error{
		err: &Round2ErrorCommitmentSerializationError{},
	}
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
	return &Round2Error{
		err: &Round2ErrorSigningPackageDeserializationError{},
	}
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
	return &Round2Error{
		err: &Round2ErrorSigningFailed{
			Message: message,
		},
	}
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

type FfiConverterTypeRound2Error struct{}

var FfiConverterTypeRound2ErrorINSTANCE = FfiConverterTypeRound2Error{}

func (c FfiConverterTypeRound2Error) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeRound2Error) Lower(value *Round2Error) RustBuffer {
	return LowerIntoRustBuffer[*Round2Error](c, value)
}

func (c FfiConverterTypeRound2Error) Read(reader io.Reader) error {
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
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeRound2Error.Read()", errorID))
	}
}

func (c FfiConverterTypeRound2Error) Write(writer io.Writer, value *Round2Error) {
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
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeRound2Error.Write", value))
	}
}

type FfiConverterSequenceTypeFrostSignatureShare struct{}

var FfiConverterSequenceTypeFrostSignatureShareINSTANCE = FfiConverterSequenceTypeFrostSignatureShare{}

func (c FfiConverterSequenceTypeFrostSignatureShare) Lift(rb RustBufferI) []FrostSignatureShare {
	return LiftFromRustBuffer[[]FrostSignatureShare](c, rb)
}

func (c FfiConverterSequenceTypeFrostSignatureShare) Read(reader io.Reader) []FrostSignatureShare {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]FrostSignatureShare, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeFrostSignatureShareINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeFrostSignatureShare) Lower(value []FrostSignatureShare) RustBuffer {
	return LowerIntoRustBuffer[[]FrostSignatureShare](c, value)
}

func (c FfiConverterSequenceTypeFrostSignatureShare) Write(writer io.Writer, value []FrostSignatureShare) {
	if len(value) > math.MaxInt32 {
		panic("[]FrostSignatureShare is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeFrostSignatureShareINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeFrostSignatureShare struct{}

func (FfiDestroyerSequenceTypeFrostSignatureShare) Destroy(sequence []FrostSignatureShare) {
	for _, value := range sequence {
		FfiDestroyerTypeFrostSignatureShare{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeFrostSigningCommitments struct{}

var FfiConverterSequenceTypeFrostSigningCommitmentsINSTANCE = FfiConverterSequenceTypeFrostSigningCommitments{}

func (c FfiConverterSequenceTypeFrostSigningCommitments) Lift(rb RustBufferI) []FrostSigningCommitments {
	return LiftFromRustBuffer[[]FrostSigningCommitments](c, rb)
}

func (c FfiConverterSequenceTypeFrostSigningCommitments) Read(reader io.Reader) []FrostSigningCommitments {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]FrostSigningCommitments, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeFrostSigningCommitmentsINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeFrostSigningCommitments) Lower(value []FrostSigningCommitments) RustBuffer {
	return LowerIntoRustBuffer[[]FrostSigningCommitments](c, value)
}

func (c FfiConverterSequenceTypeFrostSigningCommitments) Write(writer io.Writer, value []FrostSigningCommitments) {
	if len(value) > math.MaxInt32 {
		panic("[]FrostSigningCommitments is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeFrostSigningCommitmentsINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeFrostSigningCommitments struct{}

func (FfiDestroyerSequenceTypeFrostSigningCommitments) Destroy(sequence []FrostSigningCommitments) {
	for _, value := range sequence {
		FfiDestroyerTypeFrostSigningCommitments{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeParticipantIdentifier struct{}

var FfiConverterSequenceTypeParticipantIdentifierINSTANCE = FfiConverterSequenceTypeParticipantIdentifier{}

func (c FfiConverterSequenceTypeParticipantIdentifier) Lift(rb RustBufferI) []ParticipantIdentifier {
	return LiftFromRustBuffer[[]ParticipantIdentifier](c, rb)
}

func (c FfiConverterSequenceTypeParticipantIdentifier) Read(reader io.Reader) []ParticipantIdentifier {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ParticipantIdentifier, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeParticipantIdentifierINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeParticipantIdentifier) Lower(value []ParticipantIdentifier) RustBuffer {
	return LowerIntoRustBuffer[[]ParticipantIdentifier](c, value)
}

func (c FfiConverterSequenceTypeParticipantIdentifier) Write(writer io.Writer, value []ParticipantIdentifier) {
	if len(value) > math.MaxInt32 {
		panic("[]ParticipantIdentifier is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeParticipantIdentifierINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeParticipantIdentifier struct{}

func (FfiDestroyerSequenceTypeParticipantIdentifier) Destroy(sequence []ParticipantIdentifier) {
	for _, value := range sequence {
		FfiDestroyerTypeParticipantIdentifier{}.Destroy(value)
	}
}

type FfiConverterMapTypeParticipantIdentifierString struct{}

var FfiConverterMapTypeParticipantIdentifierStringINSTANCE = FfiConverterMapTypeParticipantIdentifierString{}

func (c FfiConverterMapTypeParticipantIdentifierString) Lift(rb RustBufferI) map[ParticipantIdentifier]string {
	return LiftFromRustBuffer[map[ParticipantIdentifier]string](c, rb)
}

func (_ FfiConverterMapTypeParticipantIdentifierString) Read(reader io.Reader) map[ParticipantIdentifier]string {
	result := make(map[ParticipantIdentifier]string)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterTypeParticipantIdentifierINSTANCE.Read(reader)
		value := FfiConverterStringINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapTypeParticipantIdentifierString) Lower(value map[ParticipantIdentifier]string) RustBuffer {
	return LowerIntoRustBuffer[map[ParticipantIdentifier]string](c, value)
}

func (_ FfiConverterMapTypeParticipantIdentifierString) Write(writer io.Writer, mapValue map[ParticipantIdentifier]string) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[ParticipantIdentifier]string is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterTypeParticipantIdentifierINSTANCE.Write(writer, key)
		FfiConverterStringINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapTypeParticipantIdentifierString struct{}

func (_ FfiDestroyerMapTypeParticipantIdentifierString) Destroy(mapValue map[ParticipantIdentifier]string) {
	for key, value := range mapValue {
		FfiDestroyerTypeParticipantIdentifier{}.Destroy(key)
		FfiDestroyerString{}.Destroy(value)
	}
}

type FfiConverterMapTypeParticipantIdentifierTypeFrostSecretKeyShare struct{}

var FfiConverterMapTypeParticipantIdentifierTypeFrostSecretKeyShareINSTANCE = FfiConverterMapTypeParticipantIdentifierTypeFrostSecretKeyShare{}

func (c FfiConverterMapTypeParticipantIdentifierTypeFrostSecretKeyShare) Lift(rb RustBufferI) map[ParticipantIdentifier]FrostSecretKeyShare {
	return LiftFromRustBuffer[map[ParticipantIdentifier]FrostSecretKeyShare](c, rb)
}

func (_ FfiConverterMapTypeParticipantIdentifierTypeFrostSecretKeyShare) Read(reader io.Reader) map[ParticipantIdentifier]FrostSecretKeyShare {
	result := make(map[ParticipantIdentifier]FrostSecretKeyShare)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterTypeParticipantIdentifierINSTANCE.Read(reader)
		value := FfiConverterTypeFrostSecretKeyShareINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapTypeParticipantIdentifierTypeFrostSecretKeyShare) Lower(value map[ParticipantIdentifier]FrostSecretKeyShare) RustBuffer {
	return LowerIntoRustBuffer[map[ParticipantIdentifier]FrostSecretKeyShare](c, value)
}

func (_ FfiConverterMapTypeParticipantIdentifierTypeFrostSecretKeyShare) Write(writer io.Writer, mapValue map[ParticipantIdentifier]FrostSecretKeyShare) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[ParticipantIdentifier]FrostSecretKeyShare is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterTypeParticipantIdentifierINSTANCE.Write(writer, key)
		FfiConverterTypeFrostSecretKeyShareINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapTypeParticipantIdentifierTypeFrostSecretKeyShare struct{}

func (_ FfiDestroyerMapTypeParticipantIdentifierTypeFrostSecretKeyShare) Destroy(mapValue map[ParticipantIdentifier]FrostSecretKeyShare) {
	for key, value := range mapValue {
		FfiDestroyerTypeParticipantIdentifier{}.Destroy(key)
		FfiDestroyerTypeFrostSecretKeyShare{}.Destroy(value)
	}
}

func Aggregate(signingPackage FrostSigningPackage, signatureShares []FrostSignatureShare, pubkeyPackage FrostPublicKeyPackage) (FrostSignature, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeCoordinationError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_frost_uniffi_sdk_fn_func_aggregate(FfiConverterTypeFrostSigningPackageINSTANCE.Lower(signingPackage), FfiConverterSequenceTypeFrostSignatureShareINSTANCE.Lower(signatureShares), FfiConverterTypeFrostPublicKeyPackageINSTANCE.Lower(pubkeyPackage), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostSignature
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeFrostSignatureINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func GenerateNoncesAndCommitments(secretShare FrostSecretKeyShare) (FirstRoundCommitment, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRound1Error{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_frost_uniffi_sdk_fn_func_generate_nonces_and_commitments(FfiConverterTypeFrostSecretKeyShareINSTANCE.Lower(secretShare), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FirstRoundCommitment
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeFirstRoundCommitmentINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func NewSigningPackage(message Message, commitments []FrostSigningCommitments) (FrostSigningPackage, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeCoordinationError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_frost_uniffi_sdk_fn_func_new_signing_package(FfiConverterTypeMessageINSTANCE.Lower(message), FfiConverterSequenceTypeFrostSigningCommitmentsINSTANCE.Lower(commitments), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostSigningPackage
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeFrostSigningPackageINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func Sign(signingPackage FrostSigningPackage, nonces FrostSigningNonces, keyPackage FrostKeyPackage) (FrostSignatureShare, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeRound2Error{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_frost_uniffi_sdk_fn_func_sign(FfiConverterTypeFrostSigningPackageINSTANCE.Lower(signingPackage), FfiConverterTypeFrostSigningNoncesINSTANCE.Lower(nonces), FfiConverterTypeFrostKeyPackageINSTANCE.Lower(keyPackage), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostSignatureShare
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeFrostSignatureShareINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func TrustedDealerKeygenFrom(configuration Configuration) (TrustedKeyGeneration, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeConfigurationError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_frost_uniffi_sdk_fn_func_trusted_dealer_keygen_from(FfiConverterTypeConfigurationINSTANCE.Lower(configuration), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue TrustedKeyGeneration
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeTrustedKeyGenerationINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func TrustedDealerKeygenWithIdentifiers(configuration Configuration, participants ParticipantList) (TrustedKeyGeneration, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeConfigurationError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_frost_uniffi_sdk_fn_func_trusted_dealer_keygen_with_identifiers(FfiConverterTypeConfigurationINSTANCE.Lower(configuration), FfiConverterTypeParticipantListINSTANCE.Lower(participants), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue TrustedKeyGeneration
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeTrustedKeyGenerationINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func ValidateConfig(config Configuration) error {
	_, _uniffiErr := rustCallWithError(FfiConverterTypeConfigurationError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_frost_uniffi_sdk_fn_func_validate_config(FfiConverterTypeConfigurationINSTANCE.Lower(config), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func VerifyAndGetKeyPackageFrom(secretShare FrostSecretKeyShare) (FrostKeyPackage, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeFrostError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_frost_uniffi_sdk_fn_func_verify_and_get_key_package_from(FfiConverterTypeFrostSecretKeyShareINSTANCE.Lower(secretShare), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue FrostKeyPackage
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeFrostKeyPackageINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func VerifySignature(message Message, signature FrostSignature, pubkey FrostPublicKeyPackage) error {
	_, _uniffiErr := rustCallWithError(FfiConverterTypeFrostSignatureVerificationError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_frost_uniffi_sdk_fn_func_verify_signature(FfiConverterTypeMessageINSTANCE.Lower(message), FfiConverterTypeFrostSignatureINSTANCE.Lower(signature), FfiConverterTypeFrostPublicKeyPackageINSTANCE.Lower(pubkey), _uniffiStatus)
		return false
	})
	return _uniffiErr
}
