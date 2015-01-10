package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/sethgrid/curse"
)

var (
	Width        int
	Height       int
	RoomFillRate int
	Animate      bool
)

type Stage struct {
	width, height int
	cell          map[int]map[int]Tile
}

type Tile struct {
	empty bool
	x, y  int
}

type Region struct{}

type Room struct {
	width, height, x, y int
}

func init() {
	flag.IntVar(&Width, "width", 79, "Total maze width (default 79)")
	flag.IntVar(&Height, "height", 21, "Total maze height (default 21)")
	flag.IntVar(&RoomFillRate, "room_fill_rate", 20, "Minimum percent space given to rooms (default 20)")
	flag.BoolVar(&Animate, "animate", false, "Set to watch animation in terminal")

	Width = roundUpToEven(Width) - 1
	Height = roundUpToEven(Height) - 1

	rand.Seed(time.Now().Unix())
}

func main() {
	flag.Parse()

	s := NewStage(Width, Height)
	s.AddRooms()
	s.FillMaze()

	s.PrintUnicode()
}

func NewStage(w, h int) *Stage {
	s := &Stage{width: w, height: h, cell: make(map[int]map[int]Tile)}

	// init all the cells with a new filled tile (empty defaults to false)
	for ; w >= 1; w-- {
		s.cell[w] = make(map[int]Tile)
		for ; h >= 1; h-- {
			s.cell[w][h] = Tile{x: w, y: h}
		}
		h = s.height
	}

	return s
}

// cellMask returns four character string "0000" ... "1111"
// each digit represents if the cell above, below, left, and right are walls (1)
func (s *Stage) cellMask(x, y int) string {
	top, right, bottom, left := "0", "0", "0", "0"

	if s.cellExists(x, y-1) && !s.cell[x][y-1].empty {
		top = "1"
	}
	if s.cellExists(x+1, y) && !s.cell[x+1][y].empty {
		right = "1"
	}
	if s.cellExists(x, y+1) && !s.cell[x][y+1].empty {
		bottom = "1"
	}
	if s.cellExists(x-1, y) && !s.cell[x-1][y].empty {
		left = "1"
	}

	return top + right + bottom + left
}

// cellExists is a DRY way to check into the multidimensional map
func (s *Stage) cellExists(x, y int) bool {
	if _, ok := s.cell[x]; !ok {
		return false
	}
	if _, ok := s.cell[x][y]; !ok {
		return false
	}
	return true
}

// PrintUnicode inspects surrounding cells and determines the correct
// unicode box drawing character to use
// If Animate is set to true, it will re-paint by clearing the terminal
func (s *Stage) PrintUnicode() {
	if Animate {
		// erase the terminal and start at the top
		c, _ := curse.New()
		c.Move(1, 1)
		c.EraseAll()
	}
	for y := 1; y <= s.height; y++ {
		for x := 1; x <= s.width; x++ {
			r := ' '
			if s.cell[x][y].empty {
				r = ' '
			} else {
				switch s.cellMask(x, y) {
				case "1111":
					r = '╋'
				case "0111":
					r = '┳'
				case "1011":
					r = '┫'
				case "1101":
					r = '┻'
				case "1110":
					r = '┣'
				case "0011":
					r = '┓'
				case "0110":
					r = '┏'
				case "0101":
					r = '━'
				case "1010":
					r = '┃'
				case "1001":
					r = '┛'
				case "1100":
					r = '┗'
				case "0010":
					r = '╻'
				case "0100":
					r = '╺'
				case "0001":
					r = '╸'
				case "1000":
					r = '╹'
				case "0000":
					r = '╋'
				default:
					r = '?'
				}
			}
			fmt.Printf("%c", r)
		}
		fmt.Printf("\n")
	}
}

// Print prints a non-unicode maze. Boring.
func (s *Stage) Print() {
	for y := 1; y <= s.height; y++ {
		for x := 1; x <= s.width; x++ {
			if s.cell[x][y].empty {
				fmt.Print(" ")
			} else {
				fmt.Printf("%c", '#')
			}
		}
		fmt.Printf("\n")
	}
}

// FillMaze changes s.cell values to be empty or not empty and forms a maze
func (s *Stage) FillMaze() {
	/*
		Growing Tree Algorythm - http://www.astrolog.org/labyrnth/algrithm.htm
		Each time you carve a cell, add that cell to a list.
		Proceed by picking a cell from the list, and carving into an unmade cell next to it.
		If there are no unmade cells next to the current cell, remove the current cell from the list.
		The Maze is done when the list becomes empty
	*/

	// get init cell
	var x, y int
	for x == 0 && y == 0 {
		x1, y1 := roundUpToEven(rand.Intn(s.width)), roundUpToEven(rand.Intn(s.height))
		// start on a border
		if rand.Intn(1) == 1 {
			x = 0
		} else {
			y = 0
		}
		// cells start filled as walls. empty cells are carved out already
		if !s.cellExists(x1, y1) || s.cell[x1][y1].empty {
			continue
		}

		x, y = x1, y1
	}

	// clear out the init cell
	tmpTile := s.cell[x][y]
	tmpTile.empty = true
	s.cell[x][y] = tmpTile

	tiles := make([]Tile, 0)
	tiles = append(tiles, s.cell[x][y])
	i := 0
	for len(tiles) > 0 {
		if Animate {
			s.PrintUnicode()
			time.Sleep(time.Millisecond * 20)
		}
		// pick a random cell
		i = rand.Intn(len(tiles))

		// find the next cell to carve out
		nextX, nextY, middleX, middleY := s.getNextMove(tiles, i)
		if nextX == 0 || nextY == 0 || middleX == 0 || middleY == 0 {
			// no new move found, remove this tile from the list
			tiles = append(tiles[:i], tiles[i+1:]...)
			continue
		}
		//fmt.Println(tiles)
		// carve out this cell and add it to the list, and start over
		if s.cellExists(nextX, nextY) {
			tmpTile := s.cell[nextX][nextY]
			tmpTile.empty = true
			s.cell[nextX][nextY] = tmpTile
		} else {
			//fmt.Printf("Now you fucked up.")
			// tiles = append(tiles[:i], tiles[i+1:]...)
			continue
		}
		// and clear the cell in the middle
		if s.cellExists(middleX, middleY) {
			tmpTile := s.cell[middleX][middleY]
			tmpTile.empty = true
			s.cell[middleX][middleY] = tmpTile
		}

		tiles = append(tiles, s.cell[nextX][nextY])
	}
}

// getNextMove finds the next cell's x and y. Because we have to clear out
// two cells, it returns two sets of x,y (the destination cell, and the
// cell in between)
//
//   Ex: we start at cell N, target cell M, and will need to clear O
//   ####    ####    ####
//   #N## => #N#M => #NOM
//   ####    ####    ####
//   In this way, we eat through the maze. nom nom nom
func (s *Stage) getNextMove(tiles []Tile, i int) (int, int, int, int) {
	// pick random order (1 up, 2 right, 3 down, 4 left)
	directions := getRandomIntList(1, 5)
	nextX, nextY, middleX, middleY := 0, 0, 0, 0

	for _, direction := range directions {
		curXAdj, curYAdj := 0, 0

		switch direction {
		case 1:
			curYAdj += 2
		case 2:
			curXAdj += 2
		case 3:
			curYAdj -= 2
		case 4:
			curXAdj -= 2
		default:
			log.Printf("error - direction list gave unexpected result ", direction)
		}
		nextX = tiles[i].x + curXAdj
		nextY = tiles[i].y + curYAdj

		if s.isEdge(nextX, nextY) {
			continue
		}

		if (nextX > s.width || nextX <= 0) || (nextY > s.height || nextY <= 0) {
			continue
		}

		if s.cell[nextX][nextY].empty {
			// todo - is this where logic goes to not collide with rooms?
			continue
		}

		if !s.cellExists(nextX, nextY) {
			continue
		}

		middleX, middleY = tiles[i].x+curXAdj/2, tiles[i].y+curYAdj/2
		break
	}
	return nextX, nextY, middleX, middleY
}

func (s *Stage) isEdge(x, y int) bool {
	if x == 1 || x == s.width || y == 1 || y == s.height {
		return true
	}
	return false
}

// getRandomIntList returns [start, end) random sorted list of ints
func getRandomIntList(start, end int) []int {
	r := make([]int, end-start)
	// populate it with our starting numbers
	for i, _ := range r {
		r[i] = start + i
	}
	// do some swaps
	for i, _ := range r {
		j := rand.Intn(i + 1)
		r[i], r[j] = r[j], r[i]
	}
	return r
}

func roundUpToEven(n int) int {
	if n%2 == 0 {
		return n
	}
	return n + 1
}

func (s *Stage) AddRooms() {
	roomVolumeLeft := s.width * s.height * RoomFillRate / 100
	if roomVolumeLeft == 0 {
		return
	}

	// pick some big max just to avoid infinate looping
	for maxIterations := 10000; maxIterations >= 0; maxIterations-- {
		room := Room{
			width:  roundUpToEven(rand.Intn(12) + 3),
			height: roundUpToEven(rand.Intn(8) + 3),
			x:      roundUpToEven(rand.Intn(s.width) + 1),
			y:      roundUpToEven(rand.Intn(s.height) + 1),
		}

		validRoom := true
		// +/- 1 as padding
		for x := room.x - 1; x <= room.x+room.width+1; x++ {
			for y := room.y - 1; y <= room.y+room.height+1; y++ {
				// cells start filled
				if x > s.width || y > s.height || s.cell[x][y].empty {
					validRoom = false
					continue
				}
			}
		}

		if !validRoom {
			continue
		}

		// place room
		for x := room.x; x <= room.x+room.width; x++ {
			for y := room.y; y <= room.y+room.height; y++ {
				// cells start filled
				tmpTile := s.cell[x][y]
				tmpTile.empty = true
				s.cell[x][y] = tmpTile
			}
		}
		roomVolumeLeft -= room.height * room.width
		if roomVolumeLeft <= 0 {
			break
		}
	}
}

/// The random dungeon generator.
///
/// Starting with a stage of solid walls, it works like so:
///
/// 1. Place a number of randomly sized and positioned rooms. If a room
///    overlaps an existing room, it is discarded. Any remaining rooms are
///    carved out.
/// 2. Any remaining solid areas are filled in with mazes. The maze generator
///    will grow and fill in even odd-shaped areas, but will not touch any
///    rooms.
/// 3. The result of the previous two steps is a series of unconnected rooms
///    and mazes. We walk the stage and find every tile that can be a
///    "connector". This is a solid tile that is adjacent to two unconnected
///    regions.
/// 4. We randomly choose connectors and open them or place a door there until
///    all of the unconnected regions have been joined. There is also a slight
///    chance to carve a connector between two already-joined regions, so that
///    the dungeon isn't single connected.
/// 5. The mazes will have a lot of dead ends. Finally, we remove those by
///    repeatedly filling in any open tile that's closed on three sides. When
///    this is done, every corridor in a maze actually leads somewhere.
///
/// The end result of this is a multiply-connected dungeon with rooms and lots
/// of winding corridors.
