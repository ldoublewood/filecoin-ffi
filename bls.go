package ffi

import (
	"unsafe"
)

// #cgo LDFLAGS: ${SRCDIR}/libfilecoin.a
// #cgo pkg-config: ${SRCDIR}/filecoin.pc
// #include "./filecoin.h"
import "C"

// SignatureBytes is the length of a BLS signature
const SignatureBytes = 96

// PrivateKeyBytes is the length of a BLS private key
const PrivateKeyBytes = 32

// PublicKeyBytes is the length of a BLS public key
const PublicKeyBytes = 48

// DigestBytes is the length of a BLS message hash/digest
const DigestBytes = 96

// Signature is a compressed affine
type Signature [SignatureBytes]byte

// PrivateKey is a compressed affine
type PrivateKey [PrivateKeyBytes]byte

// PublicKey is a compressed affine
type PublicKey [PublicKeyBytes]byte

// Message is a byte slice
type Message []byte

// Digest is a compressed affine
type Digest [DigestBytes]byte

// Hash computes the digest of a message
func Hash(message Message) Digest {
	// prep request
	cMessage := C.CBytes(message)
	defer C.free(cMessage)
	cMessagePtr := (*C.uchar)(cMessage)
	cMessageLen := C.size_t(len(message))

	// call method
	resPtr := (*C.HashResponse)(unsafe.Pointer(C.hash(cMessagePtr, cMessageLen)))
	defer C.destroy_hash_response(resPtr)

	// prep response
	var digest Digest
	digestSlice := C.GoBytes(unsafe.Pointer(&resPtr.digest), DigestBytes) // nolint: staticcheck
	copy(digest[:], digestSlice)

	return digest
}

// Verify verifies that a signature is the aggregated signature of digests - pubkeys
func Verify(signature *Signature, digests []Digest, publicKeys []PublicKey) bool {
	// prep data
	flattenedDigests := make([]byte, DigestBytes*len(digests))
	for idx, digest := range digests {
		copy(flattenedDigests[(DigestBytes*idx):(DigestBytes*(1+idx))], digest[:])
	}

	flattenedPublicKeys := make([]byte, PublicKeyBytes*len(publicKeys))
	for idx, publicKey := range publicKeys {
		copy(flattenedPublicKeys[(PublicKeyBytes*idx):(PublicKeyBytes*(1+idx))], publicKey[:])
	}

	// prep request
	cSignature := C.CBytes(signature[:])
	defer C.free(cSignature)
	cSignaturePtr := (*C.uchar)(cSignature)

	cFlattenedDigests := C.CBytes(flattenedDigests)
	defer C.free(cFlattenedDigests)
	cFlattenedDigestsPtr := (*C.uint8_t)(cFlattenedDigests)
	cFlattenedDigestsLen := C.size_t(len(flattenedDigests))

	cFlattenedPublicKeys := C.CBytes(flattenedPublicKeys)
	defer C.free(cFlattenedPublicKeys)
	cFlattenedPublicKeysPtr := (*C.uint8_t)(cFlattenedPublicKeys)
	cFlattenedPublicKeysLen := C.size_t(len(flattenedPublicKeys))

	// call method
	res := (C.int)(C.verify(cSignaturePtr, cFlattenedDigestsPtr, cFlattenedDigestsLen, cFlattenedPublicKeysPtr, cFlattenedPublicKeysLen))

	return res > 0
}

// Aggregate aggregates signatures together into a new signature
func Aggregate(signatures []Signature) *Signature {
	// prep data
	flattenedSignatures := make([]byte, SignatureBytes*len(signatures))
	for idx, sig := range signatures {
		copy(flattenedSignatures[(SignatureBytes*idx):(SignatureBytes*(1+idx))], sig[:])
	}

	// prep request
	cFlattenedSignatures := C.CBytes(flattenedSignatures)
	defer C.free(cFlattenedSignatures)
	cFlattenedSignaturesPtr := (*C.uint8_t)(cFlattenedSignatures)
	cFlattenedSignaturesLen := C.size_t(len(flattenedSignatures))

	// call method
	resPtr := (*C.AggregateResponse)(unsafe.Pointer(C.aggregate(cFlattenedSignaturesPtr, cFlattenedSignaturesLen)))
	if resPtr == nil {
		return nil
	}
	defer C.destroy_aggregate_response(resPtr)

	// prep response
	var signature Signature
	signatureSlice := C.GoBytes(unsafe.Pointer(&resPtr.signature), SignatureBytes) // nolint: staticcheck
	copy(signature[:], signatureSlice)

	return &signature
}

// PrivateKeyGenerate generates a private key
func PrivateKeyGenerate() PrivateKey {
	// call method
	resPtr := (*C.PrivateKeyGenerateResponse)(unsafe.Pointer(C.private_key_generate()))
	defer C.destroy_private_key_generate_response(resPtr)

	// prep response
	var privateKey PrivateKey
	privateKeySlice := C.GoBytes(unsafe.Pointer(&resPtr.private_key), PrivateKeyBytes) // nolint: staticcheck
	copy(privateKey[:], privateKeySlice)

	return privateKey
}

// PrivateKeySign signs a message
func PrivateKeySign(privateKey PrivateKey, message Message) *Signature {
	// prep request
	cPrivateKey := C.CBytes(privateKey[:])
	defer C.free(cPrivateKey)
	cPrivateKeyPtr := (*C.uchar)(cPrivateKey)

	cMessage := C.CBytes(message)
	defer C.free(cMessage)
	cMessagePtr := (*C.uchar)(cMessage)
	cMessageLen := C.size_t(len(message))

	// call method
	resPtr := (*C.PrivateKeySignResponse)(unsafe.Pointer(C.private_key_sign(cPrivateKeyPtr, cMessagePtr, cMessageLen)))
	if resPtr == nil {
		return nil
	}
	defer C.destroy_private_key_sign_response(resPtr)

	// prep response
	var signature Signature
	signatureSlice := C.GoBytes(unsafe.Pointer(&resPtr.signature), SignatureBytes) // nolint: staticcheck
	copy(signature[:], signatureSlice)

	return &signature
}

// PrivateKeyPublicKey gets the public key for a private key
func PrivateKeyPublicKey(privateKey PrivateKey) PublicKey {
	// prep request
	cPrivateKey := C.CBytes(privateKey[:])
	defer C.free(cPrivateKey)
	cPrivateKeyPtr := (*C.uchar)(cPrivateKey)

	// call method
	resPtr := (*C.PrivateKeyPublicKeyResponse)(unsafe.Pointer(C.private_key_public_key(cPrivateKeyPtr))) // nolint: staticcheck
	defer C.destroy_private_key_public_key_response(resPtr)

	// prep response
	var publicKey PublicKey
	publicKeySlice := C.GoBytes(unsafe.Pointer(&resPtr.public_key), PublicKeyBytes) // nolint: staticcheck
	copy(publicKey[:], publicKeySlice)

	return publicKey
}
