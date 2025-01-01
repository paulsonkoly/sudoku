package main

import (
	"fmt"
	"iter"
	"math/bits"
	"strings"
)

// Size of the sudoku. This can't really be changed easily because the 3x3 box
// logic assumes a 9 grid sudoku.
const Size = 9

// Cell is a one hot encoding of pencil marks. Lowest 9 bits used to indicate
// what values a cell can take.
type Cell uint16

// All is cell that has no information - can take all values.
func All() Cell {
	return 1<<Size - 1
}

// Clear removes all pencilmarks from a cell. It returns self reference for chaining.
func (c *Cell) Clear() *Cell {
	*c = 0
	return c
}

// Set adds a pencilmark to a cell. It returns self reference for chaining.
func (c *Cell) Set(i uint) *Cell {
	*c |= (1 << (i - 1))
	return c
}

// Drop removes a pencilmark from a cell. It returns self reference for chaining.
func (c *Cell) Drop(i uint) *Cell {
	*c &= ^(1 << (i - 1))
	return c
}

// IsSet determines wheter a pencilmark is set in cell.
func (c *Cell) IsSet(i uint) bool {
	return (*c)&(1<<(i-1)) != 0
}

// Single determines if a cell has 1 value left.
func (c Cell) Single() bool {
	return c&(c-1) == 0 && c != 0
}

// Digit returns the value corresponding to the first (lowest value) pencil mark.
func (c Cell) Digit() uint {
	return 1 + uint(bits.TrailingZeros(uint(c)))
}

// Digits iterate the possible (pencilmarked) digits in a cell.
func (c Cell) Digits() iter.Seq[uint] {
	return func(yield func(uint) bool) {
		for c != 0 {
			if !yield(c.Digit()) {
				return
			}
			c ^= c & -c // flip lowest 1 bit
		}
	}
}

// Count is the number of possible digits (pencilmarks) in a cell.
func (c Cell) Count() int {
	return bits.OnesCount(uint(c))
}

// String is the string representation of a cell.
func (c Cell) String() string {
	b := strings.Builder{}
	for i := uint(Size); i > 0; i-- {
		if c.IsSet(i) {
			b.WriteString(fmt.Sprintf("%d", i))
		} else {
			b.WriteString("_")
		}
	}
	return b.String()
}

// Bpard is a sudoku board.
type Board [Size * Size]Cell

// EmptyBoard is a sudoku board where all cells are [All].
func EmptyBoard() *Board {
	b := Board{}
	for i := range Size * Size {
		b[i] = All()
	}
	return &b
}

// At returns a cell pointer to the ith row jth column.
func (b *Board) At(i, j int) *Cell {
	return &b[i*Size+j]
}

// Row iteraters the column indices along with the corresponding cell from the ith row.
func (b *Board) Row(i int) iter.Seq2[int, *Cell] {
	return func(yield func(int, *Cell) bool) {
		for j := range Size {
			if !yield(j, b.At(i, j)) {
				return
			}
		}
	}
}

// Col iteraters the row indices along with the corresponding cell from the jth column.
func (b *Board) Col(j int) iter.Seq2[int, *Cell] {
	return func(yield func(int, *Cell) bool) {
		for i := range Size {
			if !yield(i, b.At(i, j)) {
				return
			}
		}
	}
}

// Box iterates the indices along with the corresponding cell from the 3x3 box
// that the cell with coordinates i, j fall into. It does not include the cell
// itself, skipping i and j.
func (b *Board) Box(i, j int) iter.Seq2[[]int, *Cell] {
	return func(yield func([]int, *Cell) bool) {
		for x := (i / 3) * 3; x < (i/3)*3+3; x++ {
			for y := (j / 3) * 3; y < (j/3)*3+3; y++ {
				if x == i && y == j {
					continue
				}
				if !yield([]int{x, y}, b.At(x, y)) {
					return
				}
			}
		}
	}
}

// Set sets (resolves) the cell digit to be d for the ith row jth column. It
// removes the pencilmarks from affected cells following sudoku rules, then
// recursively sets any cell that becomes a single digit pencilmark.
func (b *Board) Set(i, j int, d uint) {
	b.At(i, j).Clear().Set(d)

	for jj, c := range b.Row(i) {
		if j != jj && c.IsSet(d) && c.Drop(d).Single() {
			b.Set(i, jj, c.Digit())
		}
	}
	for ii, c := range b.Col(j) {
		if i != ii && c.IsSet(d) && c.Drop(d).Single() {
			b.Set(ii, j, c.Digit())
		}
	}
	for xy, c := range b.Box(i, j) {
		if c.IsSet(d) && c.Drop(d).Single() {
			b.Set(xy[0], xy[1], c.Digit())
		}
	}
}

// Lowest is the coordinates of the lowest bitcount (fewest pencilmark) cell
// that is not a single digit - if any. If none found it returns ok false.
func (b *Board) Lowest() (i, j int, ok bool) {
	lowest := Size + 1
	for ii := range Size {
		for jj := range Size {
			c := b.At(ii, jj).Count()
			if 1 < c && c < lowest {
				lowest = c
				i = ii
				j = jj
				ok = true
			}
		}
	}
	return i, j, ok
}

// Solved decides if the board contains only single digit cells.
func (b *Board) Solved() bool {
	for i := range Size {
		for j := range Size {
			if !b.At(i, j).Single() {
				return false
			}
		}
	}
	return true
}

// Solve solves the sudoku by guessing the [Lowest] cell recursively. If the
// board is not solvable it returns false.
func (b *Board) Solve() bool {
	i, j, ok := b.Lowest()
	if !ok {
		return b.Solved()
	}

	for d := range b.At(i, j).Digits() {
		cpy := Board{}
		copy(cpy[:], b[:])

		b.Set(i, j, d)

		if b.Solve() {
			return true
		}

		copy(b[:], cpy[:])
	}
	return false
}

// Print prints the sudoku board. (all pencilmarks for all cells.)
func (b *Board) Print() {
	for i := range Size {
		if i%3 == 0 {
			fmt.Printf("|-------------------------------|-------------------------------|-------------------------------|\n")
		}
		for j := range Size {
			if j%3 == 0 {
				fmt.Printf("| ")
			}
			fmt.Printf("%s ", b[i*Size+j])
		}
		fmt.Printf("|\n")
	}
	fmt.Printf("|-------------------------------|-------------------------------|-------------------------------|\n")
}

func main() {
	b := EmptyBoard()

	// https://sudoku2.com/play-the-hardest-sudoku-in-the-world/
	b.Set(0, 0, 8)
	b.Set(1, 2, 3)
	b.Set(1, 3, 6)
	b.Set(2, 1, 7)
	b.Set(2, 4, 9)
	b.Set(2, 6, 2)
	b.Set(3, 1, 5)
	b.Set(3, 5, 7)
	b.Set(4, 4, 4)
	b.Set(4, 5, 5)
	b.Set(4, 6, 7)
	b.Set(5, 3, 1)
	b.Set(5, 7, 3)
	b.Set(6, 2, 1)
	b.Set(6, 7, 6)
	b.Set(6, 8, 8)
	b.Set(7, 2, 8)
	b.Set(7, 3, 5)
	b.Set(7, 7, 1)
	b.Set(8, 1, 9)
	b.Set(8, 6, 4)

	b.Solve()
	b.Print()
}
