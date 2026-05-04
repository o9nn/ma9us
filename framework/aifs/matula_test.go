package aifs

import "testing"

func TestIsPrime(t *testing.T) {
	cases := []struct {
		n    uint64
		want bool
	}{
		{0, false}, {1, false}, {2, true}, {3, true}, {4, false},
		{5, true}, {6, false}, {7, true}, {8, false}, {9, false},
		{11, true}, {97, true}, {99, false}, {100, false}, {101, true},
	}
	for _, tc := range cases {
		if got := IsPrime(tc.n); got != tc.want {
			t.Errorf("IsPrime(%d) = %v, want %v", tc.n, got, tc.want)
		}
	}
}

func TestNthPrime(t *testing.T) {
	// Known values: 2, 3, 5, 7, 11, 13, 17, 19, 23, 29
	expected := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	for i, want := range expected {
		if got := NthPrime(uint64(i + 1)); got != want {
			t.Errorf("NthPrime(%d) = %d, want %d", i+1, got, want)
		}
	}
}

func TestNthPrimeZero(t *testing.T) {
	if NthPrime(0) != 0 {
		t.Error("NthPrime(0) should return 0")
	}
}

func TestNthPrimeLargeIndex(t *testing.T) {
	// Should extend the cache beyond the initial 15 entries.
	p := NthPrime(20)
	if !IsPrime(p) {
		t.Errorf("NthPrime(20) = %d is not prime", p)
	}
	// The 20th prime is 71.
	if p != 71 {
		t.Errorf("NthPrime(20) = %d, want 71", p)
	}
}

func TestNextPrimeAfter(t *testing.T) {
	cases := []struct{ n, want uint64 }{
		{0, 2}, {1, 2}, {2, 3}, {3, 5}, {4, 5}, {10, 11}, {12, 13},
	}
	for _, tc := range cases {
		if got := NextPrimeAfter(tc.n); got != tc.want {
			t.Errorf("NextPrimeAfter(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestMatulaCompositeEmpty(t *testing.T) {
	if got := MatulaComposite(nil); got != 1 {
		t.Errorf("MatulaComposite(nil) = %d, want 1", got)
	}
	if got := MatulaComposite([]uint64{}); got != 1 {
		t.Errorf("MatulaComposite([]) = %d, want 1", got)
	}
}

func TestMatulaCompositeAndFactor(t *testing.T) {
	// MatulaComposite([1]) = NthPrime(1) = 2
	if got := MatulaComposite([]uint64{1}); got != 2 {
		t.Errorf("MatulaComposite([1]) = %d, want 2", got)
	}
	// MatulaComposite([1, 2]) = NthPrime(1)*NthPrime(2) = 2*3 = 6
	if got := MatulaComposite([]uint64{1, 2}); got != 6 {
		t.Errorf("MatulaComposite([1,2]) = %d, want 6", got)
	}
	// MatulaComposite([2, 3]) = 3*5 = 15
	composite := MatulaComposite([]uint64{2, 3})
	if composite != 15 {
		t.Errorf("MatulaComposite([2,3]) = %d, want 15", composite)
	}
	// MatulaFactor(15) should return [2, 3] (indices of 3 and 5).
	indices := MatulaFactor(composite)
	if len(indices) != 2 {
		t.Fatalf("MatulaFactor(15) = %v, want length 2", indices)
	}
	if indices[0] != 2 || indices[1] != 3 {
		t.Errorf("MatulaFactor(15) = %v, want [2 3]", indices)
	}
}

func TestMatulaFactorOne(t *testing.T) {
	if got := MatulaFactor(1); got != nil {
		t.Errorf("MatulaFactor(1) = %v, want nil", got)
	}
}

func TestMatulaFactorPrime(t *testing.T) {
	// Factor a prime: MatulaFactor(2) should return [1] (2 is the 1st prime).
	indices := MatulaFactor(2)
	if len(indices) != 1 || indices[0] != 1 {
		t.Errorf("MatulaFactor(2) = %v, want [1]", indices)
	}
	// Factor 3: 3rd index = [2] (3 is the 2nd prime).
	indices = MatulaFactor(3)
	if len(indices) != 1 || indices[0] != 2 {
		t.Errorf("MatulaFactor(3) = %v, want [2]", indices)
	}
}

func TestMatulaRoundTrip(t *testing.T) {
	// Encode a slice of indices and decode back.
	original := []uint64{1, 1, 2, 3}
	composite := MatulaComposite(original)
	decoded := MatulaFactor(composite)
	if len(decoded) != len(original) {
		t.Fatalf("round-trip length: got %d want %d (composite=%d)", len(decoded), len(original), composite)
	}
	// Both slices should contain the same multiset.
	counts := make(map[uint64]int)
	for _, v := range original {
		counts[v]++
	}
	for _, v := range decoded {
		counts[v]--
	}
	for k, v := range counts {
		if v != 0 {
			t.Errorf("round-trip mismatch at index %d (delta %d)", k, v)
		}
	}
}

func TestAtomMatulaNameAssignment(t *testing.T) {
	g := NewKnowledgeGraph()
	a := g.AddNode(ConceptNode, "Alpha", TruthValue{1.0, 0.9})
	if !IsPrime(a.MatulaName) {
		t.Errorf("first node MatulaName %d should be prime", a.MatulaName)
	}
	b := g.AddNode(ConceptNode, "Beta", TruthValue{1.0, 0.9})
	if !IsPrime(b.MatulaName) {
		t.Errorf("second node MatulaName %d should be prime", b.MatulaName)
	}
	if a.MatulaName == b.MatulaName {
		t.Error("distinct nodes should have distinct MatulaNames")
	}
	// Links also receive distinct primes.
	lnk := g.AddLink(InheritanceLink, []uint64{a.ID, b.ID}, TruthValue{1.0, 0.9})
	if !IsPrime(lnk.MatulaName) {
		t.Errorf("link MatulaName %d should be prime", lnk.MatulaName)
	}
	if lnk.MatulaName == a.MatulaName || lnk.MatulaName == b.MatulaName {
		t.Error("link MatulaName should differ from node MatulaNames")
	}
}

func TestDuplicateNodeRetainsMatulaName(t *testing.T) {
	g := NewKnowledgeGraph()
	first := g.AddNode(ConceptNode, "Dog", TruthValue{1.0, 0.9})
	originalName := first.MatulaName
	second := g.AddNode(ConceptNode, "Dog", TruthValue{0.8, 0.5})
	if second.ID != first.ID {
		t.Error("duplicate node should return existing atom")
	}
	if second.MatulaName != originalName {
		t.Error("duplicate node should retain its original MatulaName")
	}
}
