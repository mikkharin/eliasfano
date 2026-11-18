package eliasfano_test

import (
	"fmt"

	"github.com/mikkharin/eliasfano"
)

func ExampleNew() {
	s, err := eliasfano.New([]uint64{3, 4, 7, 13, 14, 15, 21, 25, 36, 38, 54, 62})
	if err != nil {
		panic(err)
	}
	fmt.Println(s.Len(), s.At(0), s.At(3), s.At(11))

	_, err = eliasfano.New([]uint64{2, 1})
	fmt.Println(err)
	// Output:
	// 12 3 13 62
	// eliasfano: input is not sorted in non-decreasing order: values[1]=1 < values[0]=2
}

func ExampleSequence_LowerBound() {
	s, _ := eliasfano.New([]uint64{10, 20, 20, 30})
	fmt.Println(s.LowerBound(5))  // below the minimum
	fmt.Println(s.LowerBound(20)) // first of the duplicates
	fmt.Println(s.LowerBound(25)) // between values
	fmt.Println(s.LowerBound(31)) // above the maximum: Len()
	// Output:
	// 0
	// 1
	// 3
	// 4
}
