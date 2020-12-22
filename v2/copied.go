package bsonpb

import "math/bits"

func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}
func isASCIIUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// JSONCamelCase converts a snake_case identifier to a camelCase identifier,
// according to the protobuf JSON specification.
func JSONCamelCase(s string) string {
	var b []byte
	var wasUnderscore bool
	for i := 0; i < len(s); i++ { // proto identifiers are always ASCII
		c := s[i]
		if c != '_' {
			if wasUnderscore && isASCIILower(c) {
				c -= 'a' - 'A' // convert to uppercase
			}
			b = append(b, c)
		}
		wasUnderscore = c == '_'
	}
	return string(b)
}

// JSONSnakeCase converts a camelCase identifier to a snake_case identifier,
// according to the protobuf JSON specification.
func JSONSnakeCase(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ { // proto identifiers are always ASCII
		c := s[i]
		if isASCIIUpper(c) {
			b = append(b, '_')
			c += 'a' - 'A' // convert to lowercase
		}
		b = append(b, c)
	}
	return string(b)
}

// sets.go
// int64s represents a set of integers within the range of 0..63.
type int64s uint64

func (bs *int64s) Len() int {
	return bits.OnesCount64(uint64(*bs))
}
func (bs *int64s) Has(n uint64) bool {
	return uint64(*bs)&(uint64(1)<<n) > 0
}
func (bs *int64s) Set(n uint64) {
	*(*uint64)(bs) |= uint64(1) << n
}
func (bs *int64s) Clear(n uint64) {
	*(*uint64)(bs) &^= uint64(1) << n
}

// Ints represents a set of integers within the range of 0..math.MaxUint64.
type Ints struct {
	lo int64s
	hi map[uint64]struct{}
}

func (bs *Ints) Len() int {
	return bs.lo.Len() + len(bs.hi)
}
func (bs *Ints) Has(n uint64) bool {
	if n < 64 {
		return bs.lo.Has(n)
	}
	_, ok := bs.hi[n]
	return ok
}
func (bs *Ints) Set(n uint64) {
	if n < 64 {
		bs.lo.Set(n)
		return
	}
	if bs.hi == nil {
		bs.hi = make(map[uint64]struct{})
	}
	bs.hi[n] = struct{}{}
}
func (bs *Ints) Clear(n uint64) {
	if n < 64 {
		bs.lo.Clear(n)
		return
	}
	delete(bs.hi, n)
}
