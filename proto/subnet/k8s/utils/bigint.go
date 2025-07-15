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
