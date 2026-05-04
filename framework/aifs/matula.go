package aifs

import "sync"

// Matula-Goebel numbers provide a lossless integer encoding of rooted forests.
// Every memory atom carries a Matula prime as its eternal name. A collection
// of atoms can be encoded as the product of their MatulaNames, and the
// constituent atoms recovered by prime factorisation — preserving the rooted
// forest structure without loss.
//
// Reference: Matula (1968), Goebel (1980).

var (
	primeCacheMu sync.Mutex
	primeCache   = []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47}
)

// IsPrime reports whether n is a prime number.
func IsPrime(n uint64) bool {
	if n < 2 {
		return false
	}
	if n == 2 || n == 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := uint64(5); i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// NthPrime returns the n-th prime (1-indexed: NthPrime(1) == 2).
// It extends an internal cache as needed.
func NthPrime(n uint64) uint64 {
	if n == 0 {
		return 0
	}
	primeCacheMu.Lock()
	defer primeCacheMu.Unlock()
	for uint64(len(primeCache)) < n {
		candidate := primeCache[len(primeCache)-1] + 1
		for !IsPrime(candidate) {
			candidate++
		}
		primeCache = append(primeCache, candidate)
	}
	return primeCache[n-1]
}

// NextPrimeAfter returns the smallest prime strictly greater than n.
func NextPrimeAfter(n uint64) uint64 {
	c := n + 1
	for !IsPrime(c) {
		c++
	}
	return c
}

// MatulaComposite computes the product of NthPrime(m) for each m in
// childMatulaNames. This encodes a multi-child node in the rooted forest.
// An empty slice returns 1 (the identity for the multiplicative encoding).
func MatulaComposite(childMatulaNames []uint64) uint64 {
	if len(childMatulaNames) == 0 {
		return 1
	}
	product := uint64(1)
	for _, m := range childMatulaNames {
		product *= NthPrime(m)
	}
	return product
}

// MatulaFactor decomposes a Matula composite n into the list of prime-indices
// whose primes multiply to give n. Each returned value k means NthPrime(k)
// was a factor of n. This is the inverse of MatulaComposite.
func MatulaFactor(n uint64) []uint64 {
	if n <= 1 {
		return nil
	}
	var indices []uint64
	for idx := uint64(1); ; idx++ {
		p := NthPrime(idx)
		if p*p > n {
			break
		}
		for n%p == 0 {
			indices = append(indices, idx)
			n /= p
		}
	}
	if n > 1 {
		// n is a prime remainder; find its 1-based index.
		for idx := uint64(1); ; idx++ {
			p := NthPrime(idx)
			if p == n {
				indices = append(indices, idx)
				break
			}
			if p > n {
				break
			}
		}
	}
	return indices
}
