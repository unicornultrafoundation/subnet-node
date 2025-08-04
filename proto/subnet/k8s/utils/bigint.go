package utils

import "math/big"

// BigInt is a wrapper for math/big.Int that implements Equal for proto compatibility
// and can be used as a gogoproto.customtype in proto definitions.
type BigInt struct {
	big.Int
}

// Equal implements equality for proto compatibility
func (i *BigInt) Equal(that interface{}) bool {
	switch other := that.(type) {
	case *BigInt:
		return i.Cmp(&other.Int) == 0
	case BigInt:
		return i.Cmp(&other.Int) == 0
	default:
		return false
	}
}

func (i *BigInt) Size() int {
	bz, _ := i.Marshal()
	return len(bz)
}

// Marshal implements the gogo proto custom type interface.
func (i *BigInt) Marshal() ([]byte, error) {
	if i == nil {
		i = new(BigInt)
	}
	return i.Int.MarshalText()
}

// MarshalTo implements the gogo proto custom type interface.
func (i *BigInt) MarshalTo(data []byte) (n int, err error) {
	if i == nil {
		i = new(BigInt)
	}
	if i.Int.BitLen() == 0 { // The value 0
		copy(data, []byte{0x30})
		return 1, nil
	}

	bz, err := i.Marshal()
	if err != nil {
		return 0, err
	}

	copy(data, bz)
	return len(bz), nil
}

func (i *BigInt) Unmarshal(data []byte) error {
	return i.Int.UnmarshalText(data)
}

// GT returns true if i is greater than j
func (i *BigInt) GT(j *BigInt) bool {
	return i.Int.Cmp(&j.Int) > 0
}

// IsZero returns true if i is zero
func (i *BigInt) IsZero() bool {
	return i.Int.Cmp(big.NewInt(0)) == 0
}

// LTE returns true if i is less than or equal to j
func (i *BigInt) LTE(j *BigInt) bool {
	return i.Int.Cmp(&j.Int) <= 0
}

func NewInt(i int64) BigInt {
	return BigInt{
		Int: *big.NewInt(i),
	}
}

func (i BigInt) Uint64() uint64 {
	return i.Int.Uint64()
}

func (i BigInt) Sub(j BigInt) BigInt {
	return BigInt{
		Int: *i.Int.Sub(&i.Int, &j.Int),
	}
}

func (i BigInt) IsNegative() bool {
	return i.Int.Cmp(big.NewInt(0)) < 0
}

func (i BigInt) Add(j BigInt) BigInt {
	return BigInt{
		Int: *i.Int.Add(&i.Int, &j.Int),
	}
}

func (i BigInt) Mul(j BigInt) BigInt {
	return BigInt{
		Int: *i.Int.Mul(&i.Int, &j.Int),
	}
}

func (i BigInt) MulRaw(j int64) BigInt {
	return i.Mul(NewInt(j))
}

func NewIntFromUint64(i uint64) BigInt {
	return BigInt{
		Int: *big.NewInt(0).SetUint64(i),
	}
}
