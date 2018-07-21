package guid

import (
	"math/big"
	"math/rand"
	"time"
)

// Random source to generate random GUIDs
var randomSource = rand.NewSource(time.Now().Unix())

// Used to allow only one initialization of the GUID module
var isGuidInitialized = false

// 160-bits default (To maintain compatibility with used chord overlay implementation)
var guidSizeBits = 160

// Represents a Global Unique Identifier (GUID) for a system's node
type GUID struct {
	id *big.Int
}

// Initializes the GUID package with the size of the GUID.
func Init(guidBitsSize int) {
	if !isGuidInitialized {
		guidSizeBits = guidBitsSize
		isGuidInitialized = true
	}
}

// Size of the GUID (in bits).
func SizeBits() int {
	return guidSizeBits
}

// Size of the GUID (in bytes).
func SizeBytes() int {
	return guidSizeBits / 8
}

// Maximum GUID available for the current defined number of bits.
func MaximumGUID() *GUID {
	maxId := big.NewInt(0)
	maxId.Exp(big.NewInt(2), big.NewInt(int64(guidSizeBits)), nil)
	maxId = maxId.Sub(maxId, big.NewInt(1))
	return newGUIDBigInt(maxId)
}

// Generate a random GUID in the range [0,MaxGUID).
func NewGUIDRandom() *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.Rand(rand.New(randomSource), MaximumGUID().id)

	return guid
}

// Creates a new GUID based on a string representation (in base 10) of the identifier.
func NewGUIDString(stringID string) *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.SetString(stringID, 10)

	return guid
}

// Creates a new GUID based on an integer64 representation of the identifier.
func NewGUIDInteger(intId int64) *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.SetInt64(intId)

	return guid
}

// Creates a new GUID based on an array of bytes representation of the identifier.
// Array of bytes is a representation of the number using the minimum number of bits.
func NewGUIDBytes(bytesID []byte) *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.SetBytes(bytesID)
	return guid
}

// Creates a new GUID based on Golang big.Int representation.
func newGUIDBigInt(bytesID *big.Int) *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.Set(bytesID)
	return guid
}

// Generates a random GUID that belongs to the interval [this, topGUID).
func (guid *GUID) GenerateInnerRandomGUID(topGUID GUID) (*GUID, error) {
	dif := big.NewInt(0)
	randOffset := big.NewInt(0)
	res := big.NewInt(0)

	dif.Sub(topGUID.id, guid.id)

	randOffset.Rand(rand.New(randomSource), dif)

	res.Add(guid.id, randOffset)

	return NewGUIDString(res.String()), nil
}

// Returns the number of ids (as a string with an integer in base 10) using % offset to higher GUID
func (guid *GUID) PercentageOffset(offsetPercentage int, nextGuid GUID) string {
	offset := big.NewInt(int64(offsetPercentage))
	dif := big.NewInt(0)
	dif.Sub(nextGuid.id, guid.id) // Dif between nextGuid and receiver

	offset.Mul(offset, dif)
	offset.Div(offset, big.NewInt(100))
	return offset.String()
}

// Adds an offset (as a string in base 10) of ids to the GUID.
func (guid *GUID) AddOffset(offset string) {
	toAdd := big.NewInt(0)
	toAdd.SetString(offset, 10)

	guid.id.Add(guid.id, toAdd)
}

// Cmp used to check what if the guid is higher, lower or equal than the given guid.
func (guid *GUID) Cmp(guid2 GUID) int {
	return guid.id.Cmp(guid2.id)
}

// Higher returns true if guid is higher than the given guid and false otherwise.
func (guid *GUID) Higher(guid2 GUID) bool {
	return guid.id.Cmp(guid2.id) > 0
}

// Greater returns true if guid is lower than the given guid and false otherwise.
func (guid *GUID) Lower(guid2 GUID) bool {
	return guid.id.Cmp(guid2.id) < 0
}

// Compare if two GUIDs are equal or not.
func (guid *GUID) Equals(guid2 GUID) bool {
	return guid.id.Cmp(guid2.id) == 0
}

// Returns an array of bytes (with size of guidSizeBits) with the value of the GUID
func (guid *GUID) Bytes() []byte {
	numOfBytes := guidSizeBits / 8
	res := make([]byte, numOfBytes)
	idBytes := guid.id.Bytes()
	index := 0
	for ; index < numOfBytes-cap(idBytes); index++ { // Padding the higher bytes with 0
		res[index] = 0
	}
	for k := 0; index < numOfBytes; k++ {
		res[index] = idBytes[k]
		index++
	}
	return res
}

// Returns an int64 that represents the GUID
func (guid *GUID) Int64() int64 {
	return guid.id.Int64()
}

// Creates a copy of the GUID object.
func (guid *GUID) Copy() *GUID {
	return NewGUIDString(guid.String())
}

// String returns the value of the GUID in a string representation (as an integer in base 10)
func (guid *GUID) String() string {
	return guid.id.String()
}

// Short returns the first digits of the GUID in a string representation
func (guid *GUID) Short() string {
	return guid.id.String()[0:12]
}
