package reference

import (
	"fmt"
	"testing"
)

const ID = "Pf3D7_01_v3"
const ID2 = "Pf3D7_02_v3"

func TestIndexHomopolymers(t *testing.T) {
	index := index(3, []string{"A", "C", "G", "T"}, t)

	if len(index) != 2 {
		t.Fatalf("Expected index of size 2, got %d", len(index))
	}

	fmt.Println(index)
	if len(index[ID]) != 61 {
		t.Fatalf("Expected 61 homopolymers got %d", len(index[ID]))
	}

}

func TestIndexHomopolymers_min_5(t *testing.T) {
	index := index(5, []string{"A", "C", "G", "T"}, t)

	if len(index[ID]) != 16 {
		t.Fatalf("Expected 16 homopolymers of len 5 or more, got %d", len(index[ID]))
	}

}

func TestIndexFewerBases(t *testing.T) {
	index := index(5, []string{"A"}, t)

	if len(index[ID]) != 3 {
		t.Fatalf("Expected 5 Poly A of len 3 or more, got %d", len(index[ID]))
	}
}

func TestChrTwo(t *testing.T) {
	index := index(3, []string{"A", "C", "G", "T"}, t)

	if len(index[ID2]) != 4 {
		t.Fatalf("Expected 4 hps on chrome 2 : %d", len(index[ID2]))
	}

}

func index(size int, bases []string, t *testing.T) map[string][]Range {
	index, err := IndexHomopolymers("testdata/ref_test.fasta", size, bases)
	if err != nil {
		t.Fatalf("Error indexing %v", err)
	}
	return index.index
}

func TestSearch(t *testing.T) {

	index, err := IndexHomopolymers("testdata/ref_test.fasta", 5, []string{"A"})
	if err != nil {
		t.Fatalf("Error indexing %v", err)
	}

	query := NewRange(149, 152)
	result, found := index.Search(ID, query)
	if !found {
		t.Fatalf("Could not find %s in %v", query.String(), index)
	}
	if result.Start != 149 {
		t.Fatalf("Should have found %s at 0 instead got %d - %v", query.String(), result.Start, index)
	}

}
