// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package apitypes

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
)

var typedDataReferenceTypeRegexp = regexp.MustCompile(`^[A-Za-z](\w*)(\[\d*\])*$`)

// TypedData is a type to encapsulate EIP-712 typed messages
type TypedData struct {
	Types       Types            `json:"types"`
	PrimaryType string           `json:"primaryType"`
	Domain      TypedDataDomain  `json:"domain"`
	Message     TypedDataMessage `json:"message"`
}

// Type is the inner type of an EIP-712 message
type Type struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// isArray returns true if the type is a fixed or variable sized array.
// This method may return false positives, in case the Type is not a valid
// expression, e.g. "fooo[[[[".
func (t *Type) isArray() bool {
	return strings.IndexByte(t.Type, '[') > 0
}

// typeName returns the canonical name of the type. If the type is 'Person[]' or 'Person[2]', then
// this method returns 'Person'
func (t *Type) typeName() string {
	return strings.Split(t.Type, "[")[0]
}

type Types map[string][]Type

type TypePriority struct {
	Type  string
	Value uint
}

type TypedDataMessage = map[string]interface{}

// TypedDataDomain represents the domain part of an EIP-712 message.
type TypedDataDomain struct {
	Name              string                `json:"name"`
	Version           string                `json:"version"`
	ChainId           *math.HexOrDecimal256 `json:"chainId"`
	VerifyingContract string                `json:"verifyingContract"`
	Salt              string                `json:"salt"`
}

// TypedDataAndHash is a helper function that calculates a hash for typed data conforming to EIP-712.
// This hash can then be safely used to calculate a signature.
//
// See https://eips.ethereum.org/EIPS/eip-712 for the full specification.
//
// This gives context to the signed typed data and prevents signing of transactions.
func TypedDataAndHash(typedData TypedData) ([]byte, string, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, "", err
	}
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, "", err
	}
	rawData := fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash))
	return crypto.Keccak256([]byte(rawData)), rawData, nil
}

// HashStruct generates a keccak256 hash of the encoding of the provided data
func (typedData *TypedData) HashStruct(primaryType string, data TypedDataMessage) (hexutil.Bytes, error) {
	encodedData, err := typedData.EncodeData(primaryType, data, 1)
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(encodedData), nil
}

// Dependencies returns an array of custom types ordered by their hierarchical reference tree
func (typedData *TypedData) Dependencies(primaryType string, found []string) []string {
	primaryType = strings.Split(primaryType, "[")[0]

	if slices.Contains(found, primaryType) {
		return found
	}
	if typedData.Types[primaryType] == nil {
		return found
	}
	found = append(found, primaryType)
	for _, field := range typedData.Types[primaryType] {
		for _, dep := range typedData.Dependencies(field.Type, found) {
			if !slices.Contains(found, dep) {
				found = append(found, dep)
			}
		}
	}
	return found
}

// EncodeType generates the following encoding:
// `name ‖ "(" ‖ member₁ ‖ "," ‖ member₂ ‖ "," ‖ … ‖ memberₙ ")"`
//
// each member is written as `type ‖ " " ‖ name` encodings cascade down and are sorted by name
func (typedData *TypedData) EncodeType(primaryType string) hexutil.Bytes {
	// Get dependencies primary first, then alphabetical
	deps := typedData.Dependencies(primaryType, []string{})
	if len(deps) > 0 {
		slicedDeps := deps[1:]
		sort.Strings(slicedDeps)
		deps = append([]string{primaryType}, slicedDeps...)
	}

	// Format as a string with fields
	var buffer bytes.Buffer
	for _, dep := range deps {
		buffer.WriteString(dep)
		buffer.WriteString("(")
		for _, obj := range typedData.Types[dep] {
			buffer.WriteString(obj.Type)
			buffer.WriteString(" ")
			buffer.WriteString(obj.Name)
			buffer.WriteString(",")
		}
		buffer.Truncate(buffer.Len() - 1)
		buffer.WriteString(")")
	}
	return buffer.Bytes()
}

// TypeHash creates the keccak256 hash  of the data
func (typedData *TypedData) TypeHash(primaryType string) hexutil.Bytes {
	return crypto.Keccak256(typedData.EncodeType(primaryType))
}

// EncodeData generates the following encoding:
// `enc(value₁) ‖ enc(value₂) ‖ … ‖ enc(valueₙ)`
//
// each encoded member is 32-byte long
func (typedData *TypedData) EncodeData(primaryType string, data map[string]interface{}, depth int) (hexutil.Bytes, error) {
	if err := typedData.validate(); err != nil {
		return nil, err
	}

	buffer := bytes.Buffer{}

	// Verify extra data
	if exp, got := len(typedData.Types[primaryType]), len(data); exp < got {
		return nil, fmt.Errorf("there is extra data provided in the message (%d < %d)", exp, got)
	}

	// Add typehash
	buffer.Write(typedData.TypeHash(primaryType))

	// Add field contents. Structs and arrays have special handlers.
	for _, field := range typedData.Types[primaryType] {
		encType := field.Type
		encValue := data[field.Name]
		if encType[len(encType)-1:] == "]" {
			encodedData, err := typedData.encodeArrayValue(encValue, encType, depth)
			if err != nil {
				return nil, err
			}
			buffer.Write(encodedData)
		} else if typedData.Types[field.Type] != nil {
			mapValue, ok := encValue.(map[string]interface{})
			if !ok {
				return nil, dataMismatchError(encType, encValue)
			}
			encodedData, err := typedData.EncodeData(field.Type, mapValue, depth+1)
			if err != nil {
				return nil, err
			}
			buffer.Write(crypto.Keccak256(encodedData))
		} else {
			byteValue, err := typedData.EncodePrimitiveValue(encType, encValue, depth)
			if err != nil {
				return nil, err
			}
			buffer.Write(byteValue)
		}
	}
	return buffer.Bytes(), nil
}

func (typedData *TypedData) encodeArrayValue(encValue interface{}, encType string, depth int) (hexutil.Bytes, error) {
	arrayValue, err := convertDataToSlice(encValue)
	if err != nil {
		return nil, dataMismatchError(encType, encValue)
	}

	arrayBuffer := new(bytes.Buffer)
	parsedType := strings.Split(encType, "[")[0]
	for _, item := range arrayValue {
		if reflect.TypeOf(item).Kind() == reflect.Slice ||
			reflect.TypeOf(item).Kind() == reflect.Array {
			encodedData, err := typedData.encodeArrayValue(item, parsedType, depth+1)
			if err != nil {
				return nil, err
			}
			arrayBuffer.Write(encodedData)
		} else {
			if typedData.Types[parsedType] != nil {
				mapValue, ok := item.(map[string]interface{})
				if !ok {
					return nil, dataMismatchError(parsedType, item)
				}
				encodedData, err := typedData.EncodeData(parsedType, mapValue, depth+1)
				if err != nil {
					return nil, err
				}
				digest := crypto.Keccak256(encodedData)
				arrayBuffer.Write(digest)
			} else {
				bytesValue, err := typedData.EncodePrimitiveValue(parsedType, item, depth)
				if err != nil {
					return nil, err
				}
				arrayBuffer.Write(bytesValue)
			}
		}
	}
	return crypto.Keccak256(arrayBuffer.Bytes()), nil
}

// Attempt to parse bytes in different formats: byte array, hex string, hexutil.Bytes.
func parseBytes(encType interface{}) ([]byte, bool) {
	// Handle array types.
	val := reflect.ValueOf(encType)
	if val.Kind() == reflect.Array && val.Type().Elem().Kind() == reflect.Uint8 {
		v := reflect.MakeSlice(reflect.TypeOf([]byte{}), val.Len(), val.Len())
		reflect.Copy(v, val)
		return v.Bytes(), true
	}

	switch v := encType.(type) {
	case []byte:
		return v, true
	case hexutil.Bytes:
		return v, true
	case string:
		bytes, err := hexutil.Decode(v)
		if err != nil {
			return nil, false
		}
		return bytes, true
	default:
		return nil, false
	}
}

func parseInteger(encType string, encValue interface{}) (*big.Int, error) {
	var (
		length int
		signed = strings.HasPrefix(encType, "int")
		b      *big.Int
	)
	if encType == "int" || encType == "uint" {
		length = 256
	} else {
		lengthStr := ""
		if strings.HasPrefix(encType, "uint") {
			lengthStr = strings.TrimPrefix(encType, "uint")
		} else {
			lengthStr = strings.TrimPrefix(encType, "int")
		}
		atoiSize, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid size on integer: %v", lengthStr)
		}
		length = atoiSize
	}
	switch v := encValue.(type) {
	case *math.HexOrDecimal256:
		b = (*big.Int)(v)
	case *big.Int:
		b = v
	case string:
		var hexIntValue math.HexOrDecimal256
		if err := hexIntValue.UnmarshalText([]byte(v)); err != nil {
			return nil, err
		}
		b = (*big.Int)(&hexIntValue)
	case float64:
		// JSON parses non-strings as float64. Fail if we cannot
		// convert it losslessly
		if float64(int64(v)) == v {
			b = big.NewInt(int64(v))
		} else {
			return nil, fmt.Errorf("invalid float value %v for type %v", v, encType)
		}
	}
	if b == nil {
		return nil, fmt.Errorf("invalid integer value %v/%v for type %v", encValue, reflect.TypeOf(encValue), encType)
	}
	if b.BitLen() > length {
		return nil, fmt.Errorf("integer larger than '%v'", encType)
	}
	if !signed && b.Sign() == -1 {
		return nil, fmt.Errorf("invalid negative value for unsigned type %v", encType)
	}
	return b, nil
}

// EncodePrimitiveValue deals with the primitive values found
// while searching through the typed data
func (typedData *TypedData) EncodePrimitiveValue(encType string, encValue interface{}, depth int) ([]byte, error) {
	switch encType {
	case "address":
		retval := make([]byte, 32)
		switch val := encValue.(type) {
		case string:
			if common.IsHexAddress(val) {
				copy(retval[12:], common.HexToAddress(val).Bytes())
				return retval, nil
			}
		case []byte:
			if len(val) == 20 {
				copy(retval[12:], val)
				return retval, nil
			}
		case [20]byte:
			copy(retval[12:], val[:])
			return retval, nil
		}
		return nil, dataMismatchError(encType, encValue)
	case "bool":
		boolValue, ok := encValue.(bool)
		if !ok {
			return nil, dataMismatchError(encType, encValue)
		}
		if boolValue {
			return math.PaddedBigBytes(common.Big1, 32), nil
		}
		return math.PaddedBigBytes(common.Big0, 32), nil
	case "string":
		strVal, ok := encValue.(string)
		if !ok {
			return nil, dataMismatchError(encType, encValue)
		}
		return crypto.Keccak256([]byte(strVal)), nil
	case "bytes":
		bytesValue, ok := parseBytes(encValue)
		if !ok {
			return nil, dataMismatchError(encType, encValue)
		}
		return crypto.Keccak256(bytesValue), nil
	}
	if strings.HasPrefix(encType, "bytes") {
		lengthStr := strings.TrimPrefix(encType, "bytes")
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid size on bytes: %v", lengthStr)
		}
		if length < 0 || length > 32 {
			return nil, fmt.Errorf("invalid size on bytes: %d", length)
		}
		if byteValue, ok := parseBytes(encValue); !ok || len(byteValue) != length {
			return nil, dataMismatchError(encType, encValue)
		} else {
			// Right-pad the bits
			dst := make([]byte, 32)
			copy(dst, byteValue)
			return dst, nil
		}
	}
	if strings.HasPrefix(encType, "int") || strings.HasPrefix(encType, "uint") {
		b, err := parseInteger(encType, encValue)
		if err != nil {
			return nil, err
		}
		return math.U256Bytes(new(big.Int).Set(b)), nil
	}
	return nil, fmt.Errorf("unrecognized type '%s'", encType)
}

// dataMismatchError generates an error for a mismatch between
// the provided type and data
func dataMismatchError(encType string, encValue interface{}) error {
	return fmt.Errorf("provided data '%v' doesn't match type '%s'", encValue, encType)
}

func convertDataToSlice(encValue interface{}) ([]interface{}, error) {
	var outEncValue []interface{}
	rv := reflect.ValueOf(encValue)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			outEncValue = append(outEncValue, rv.Index(i).Interface())
		}
	} else {
		return outEncValue, fmt.Errorf("provided data '%v' is not slice", encValue)
	}
	return outEncValue, nil
}

// validate makes sure the types are sound
func (typedData *TypedData) validate() error {
	if err := typedData.Types.validate(); err != nil {
		return err
	}
	if err := typedData.Domain.validate(); err != nil {
		return err
	}
	return nil
}

// Map generates a map version of the typed data
func (typedData *TypedData) Map() map[string]interface{} {
	dataMap := map[string]interface{}{
		"types":       typedData.Types,
		"domain":      typedData.Domain.Map(),
		"primaryType": typedData.PrimaryType,
		"message":     typedData.Message,
	}
	return dataMap
}

// Format returns a representation of typedData, which can be easily displayed by a user-interface
// without in-depth knowledge about 712 rules
func (typedData *TypedData) Format() ([]*NameValueType, error) {
	domain, err := typedData.formatData("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, err
	}
	ptype, err := typedData.formatData(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}
	var nvts []*NameValueType
	nvts = append(nvts, &NameValueType{
		Name:  "EIP712Domain",
		Value: domain,
		Typ:   "domain",
	})
	nvts = append(nvts, &NameValueType{
		Name:  typedData.PrimaryType,
		Value: ptype,
		Typ:   "primary type",
	})
	return nvts, nil
}

func (typedData *TypedData) formatData(primaryType string, data map[string]interface{}) ([]*NameValueType, error) {
	var output []*NameValueType

	// Add field contents. Structs and arrays have special handlers.
	for _, field := range typedData.Types[primaryType] {
		encName := field.Name
		encValue := data[encName]
		item := &NameValueType{
			Name: encName,
			Typ:  field.Type,
		}
		if field.isArray() {
			arrayValue, _ := convertDataToSlice(encValue)
			parsedType := field.typeName()
			for _, v := range arrayValue {
				if typedData.Types[parsedType] != nil {
					mapValue, _ := v.(map[string]interface{})
					mapOutput, err := typedData.formatData(parsedType, mapValue)
					if err != nil {
						return nil, err
					}
					item.Value = mapOutput
				} else {
					primitiveOutput, err := formatPrimitiveValue(field.Type, encValue)
					if err != nil {
						return nil, err
					}
					item.Value = primitiveOutput
				}
			}
		} else if typedData.Types[field.Type] != nil {
			if mapValue, ok := encValue.(map[string]interface{}); ok {
				mapOutput, err := typedData.formatData(field.Type, mapValue)
				if err != nil {
					return nil, err
				}
				item.Value = mapOutput
			} else {
				item.Value = "<nil>"
			}
		} else {
			primitiveOutput, err := formatPrimitiveValue(field.Type, encValue)
			if err != nil {
				return nil, err
			}
			item.Value = primitiveOutput
		}
		output = append(output, item)
	}
	return output, nil
}

func formatPrimitiveValue(encType string, encValue interface{}) (string, error) {
	switch encType {
	case "address":
		if stringValue, ok := encValue.(string); !ok {
			return "", fmt.Errorf("could not format value %v as address", encValue)
		} else {
			return common.HexToAddress(stringValue).String(), nil
		}
	case "bool":
		if boolValue, ok := encValue.(bool); !ok {
			return "", fmt.Errorf("could not format value %v as bool", encValue)
		} else {
			return fmt.Sprintf("%t", boolValue), nil
		}
	case "bytes", "string":
		return fmt.Sprintf("%s", encValue), nil
	}
	if strings.HasPrefix(encType, "bytes") {
		return fmt.Sprintf("%s", encValue), nil
	}
	if strings.HasPrefix(encType, "uint") || strings.HasPrefix(encType, "int") {
		if b, err := parseInteger(encType, encValue); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%d (%#x)", b, b), nil
		}
	}
	return "", fmt.Errorf("unhandled type %v", encType)
}

// validate checks if the types object is conformant to the specs
func (t Types) validate() error {
	for typeKey, typeArr := range t {
		if len(typeKey) == 0 {
			return errors.New("empty type key")
		}
		for i, typeObj := range typeArr {
			if len(typeObj.Type) == 0 {
				return fmt.Errorf("type %q:%d: empty Type", typeKey, i)
			}
			if len(typeObj.Name) == 0 {
				return fmt.Errorf("type %q:%d: empty Name", typeKey, i)
			}
			if typeKey == typeObj.Type {
				return fmt.Errorf("type %q cannot reference itself", typeObj.Type)
			}
			if isPrimitiveTypeValid(typeObj.Type) {
				continue
			}
			// Must be reference type
			if _, exist := t[typeObj.typeName()]; !exist {
				return fmt.Errorf("reference type %q is undefined", typeObj.Type)
			}
			if !typedDataReferenceTypeRegexp.MatchString(typeObj.Type) {
				return fmt.Errorf("unknown reference type %q", typeObj.Type)
			}
		}
	}
	return nil
}

var validPrimitiveTypes = map[string]struct{}{}

// build the set of valid primitive types
func init() {
	// Types those are trivially valid
	for _, t := range []string{
		"address", "address[]", "bool", "bool[]", "string", "string[]",
		"bytes", "bytes[]", "int", "int[]", "uint", "uint[]",
	} {
		validPrimitiveTypes[t] = struct{}{}
	}
	// For 'bytesN', 'bytesN[]', we allow N from 1 to 32
	for n := 1; n <= 32; n++ {
		validPrimitiveTypes[fmt.Sprintf("bytes%d", n)] = struct{}{}
		validPrimitiveTypes[fmt.Sprintf("bytes%d[]", n)] = struct{}{}
	}
	// For 'intN','intN[]' and 'uintN','uintN[]' we allow N in increments of 8, from 8 up to 256
	for n := 8; n <= 256; n += 8 {
		validPrimitiveTypes[fmt.Sprintf("int%d", n)] = struct{}{}
		validPrimitiveTypes[fmt.Sprintf("int%d[]", n)] = struct{}{}
		validPrimitiveTypes[fmt.Sprintf("uint%d", n)] = struct{}{}
		validPrimitiveTypes[fmt.Sprintf("uint%d[]", n)] = struct{}{}
	}
}

// Checks if the primitive value is valid
func isPrimitiveTypeValid(primitiveType string) bool {
	input := strings.Split(primitiveType, "[")[0]
	_, ok := validPrimitiveTypes[input]
	return ok
}

// validate checks if the given domain is valid, i.e. contains at least
// the minimum viable keys and values
func (domain *TypedDataDomain) validate() error {
	if domain.ChainId == nil && len(domain.Name) == 0 && len(domain.Version) == 0 && len(domain.VerifyingContract) == 0 && len(domain.Salt) == 0 {
		return errors.New("domain is undefined")
	}

	return nil
}

// Map is a helper function to generate a map version of the domain
func (domain *TypedDataDomain) Map() map[string]interface{} {
	dataMap := map[string]interface{}{}

	if domain.ChainId != nil {
		dataMap["chainId"] = domain.ChainId
	}

	if len(domain.Name) > 0 {
		dataMap["name"] = domain.Name
	}

	if len(domain.Version) > 0 {
		dataMap["version"] = domain.Version
	}

	if len(domain.VerifyingContract) > 0 {
		dataMap["verifyingContract"] = domain.VerifyingContract
	}

	if len(domain.Salt) > 0 {
		dataMap["salt"] = domain.Salt
	}
	return dataMap
}

// NameValueType is a very simple struct with Name, Value and Type. It's meant for simple
// json structures used to communicate signing-info about typed data with the UI
type NameValueType struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Typ   string      `json:"type"`
}

// Pprint returns a pretty-printed version of nvt
func (nvt *NameValueType) Pprint(depth int) string {
	output := bytes.Buffer{}
	output.WriteString(strings.Repeat("\u00a0", depth*2))
	output.WriteString(fmt.Sprintf("%s [%s]: ", nvt.Name, nvt.Typ))
	if nvts, ok := nvt.Value.([]*NameValueType); ok {
		output.WriteString("\n")
		for _, next := range nvts {
			sublevel := next.Pprint(depth + 1)
			output.WriteString(sublevel)
		}
	} else {
		if nvt.Value != nil {
			output.WriteString(fmt.Sprintf("%q\n", nvt.Value))
		} else {
			output.WriteString("\n")
		}
	}
	return output.String()
}
