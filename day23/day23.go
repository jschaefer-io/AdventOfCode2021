package day23

import (
    "fmt"
    "github.com/jschaefer-io/aoc2021/orchestration"
    "math"
    "strconv"
    "strings"
)

type Position struct {
    X int
    Y int
}

type PathCache struct {
    cache map[Position]map[Position][]Path
}

func (p *PathCache) Lookup(a, b Position) (bool, []Path) {
    if _, ok := p.cache[a]; !ok {
        return false, nil
    }
    path, ok := p.cache[a][b]
    return ok, path
}

func (p *PathCache) AddLookup(a, b Position, paths []Path) {
    if _, ok := p.cache[a]; !ok {
        p.cache[a] = make(map[Position][]Path)
    }
    p.cache[a][b] = paths
}

type Map struct {
    Tiles     map[Position]Tile
    Amphipods map[Position]Amphipod
    Forbidden map[Position]struct{}
    Score     int
    Depth     int
    cache     PathCache
    progress  []string
}

func (m *Map) Finished() bool {
    for pos, tile := range m.Tiles {
        if tile.Room == 0 {
            continue
        }
        amp, ok := m.Amphipods[pos]
        if !ok {
            return false
        }
        if amp.Destination != tile.Room {
            return false
        }
    }
    return true
}

func NewMap(length, depth int) Map {
    m := Map{
        Tiles:     make(map[Position]Tile),
        Amphipods: make(map[Position]Amphipod),
        Forbidden: make(map[Position]struct{}),
        Depth:     depth,
        cache: PathCache{
            cache: make(map[Position]map[Position][]Path),
        },
        progress: make([]string, 0),
    }
    var offset rune = 0
    for i := 0; i < length; i++ {
        m.Tiles[Position{i, 0}] = Tile{}
        if i%2 == 0 && i != 0 && i != length-1 {
            m.Forbidden[Position{i, 0}] = struct{}{}
            for y := 0; y < depth; y++ {
                m.Tiles[Position{i, 1 + y}] = Tile{
                    Room: 'A' + offset,
                }
            }
            offset++
        }
    }
    return m
}

func (m *Map) AddAmphipod(pos Position, amp Amphipod) {
    m.Amphipods[pos] = amp
}

func (m *Map) GetPaths(current, last Position) []Path {
    if ok, lookup := m.cache.Lookup(current, last); ok {
        return lookup
    }
    paths := make([]Path, 0)
    paths = append(paths, Path{})
    for _, n := range m.Neighbors(current) {
        if n == last {
            continue
        }
        paths = append(paths, m.GetPaths(n, current)...)
    }
    for i, path := range paths {
        nPath := make(Path, 0)
        nPath = append(nPath, current)
        nPath = append(nPath, path...)
        paths[i] = nPath
    }
    m.cache.AddLookup(current, last, paths)
    return paths
}

func (m *Map) Neighbors(pos Position) []Position {
    neighbors := make([]Position, 0)
    for xO := -1; xO <= 1; xO += 2 {
        target := Position{pos.X + xO, pos.Y}
        if _, ok := m.Tiles[target]; ok {
            neighbors = append(neighbors, target)
        }
    }
    for yO := -1; yO <= 1; yO += 2 {
        target := Position{pos.X, pos.Y + yO}
        if _, ok := m.Tiles[target]; ok {
            neighbors = append(neighbors, target)
        }
    }
    return neighbors
}

func (m *Map) Copy() Map {
    newMap := Map{
        Tiles:     make(map[Position]Tile),
        Amphipods: make(map[Position]Amphipod),
        Forbidden: make(map[Position]struct{}),
        Depth:     m.Depth,
        Score:     m.Score,
        cache:     m.cache,
        progress:  append(m.progress, m.String()),
    }

    for p, t := range m.Tiles {
        newMap.Tiles[p] = Tile{
            Room: t.Room,
        }
    }

    for p, amp := range m.Amphipods {
        newMap.Amphipods[p] = Amphipod{
            Destination: amp.Destination,
        }
    }

    for p, s := range m.Forbidden {
        newMap.Forbidden[p] = s
    }

    return newMap
}

func (m *Map) Step() []Map {
    variations := make([]Map, 0)
    for pos, amp := range m.Amphipods {
        for _, path := range amp.GetPaths(pos, m) {
            newMap := m.Copy()
            nAmp := newMap.Amphipods[pos]
            delete(newMap.Amphipods, pos)
            newMap.Amphipods[path.destination] = nAmp
            newMap.Score += nAmp.Energy(path.count)
            variations = append(variations, newMap)
        }
    }
    return variations
}

func (m Map) String() string {
    //var sb strings.Builder
    //for pos, amp := range m.Amphipods {
    //    sb.WriteString(fmt.Sprintf("_%d-%d%d", amp.Destination, pos.X, pos.Y))
    //}
    //return sb.String()
    minX := math.MaxInt32
    minY := math.MaxInt32
    maxX := 0
    maxY := 0
    for pos, _ := range m.Tiles {
        if pos.X < minX {
            minX = pos.X
        }
        if pos.X > maxX {
            maxX = pos.X
        }
        if pos.Y < minY {
            minY = pos.Y
        }
        if pos.Y > maxY {
            maxY = pos.Y
        }
    }
    var sb strings.Builder
    for y := minY - 1; y <= maxY+1; y++ {
        for x := minX - 1; x <= maxX+1; x++ {
            pos := Position{x, y}
            if _, ok := m.Tiles[pos]; ok {
                if amp, ok := m.Amphipods[pos]; ok {
                    sb.WriteString(string(amp.Destination))
                } else {
                    sb.WriteString(".")
                }
            } else {
                sb.WriteString("#")
            }
        }
        sb.WriteString("\n")
    }
    return sb.String()
}

type Path []Position

type Amphipod struct {
    Destination rune
}

type AmphipodPath struct {
    destination Position
    count       int
}

func (amp *Amphipod) GetPaths(current Position, m *Map) []AmphipodPath {
    paths := make([]AmphipodPath, 0)
    neighbors := m.Neighbors(current)
    for _, n := range neighbors {
        for _, path := range m.GetPaths(n, current) {
            if amp.ValidatePath(current, path, m) {
                count := len(path)
                paths = append(paths, AmphipodPath{
                    destination: path[count-1],
                    count:       count,
                })
            }
        }
    }
    return paths
}

func (amp *Amphipod) Energy(count int) int {
    switch amp.Destination {
    case 'A':
        return count
    case 'B':
        return count * 10
    case 'C':
        return count * 100
    case 'D':
        return count * 1000
    }
    panic("undefined amphipod type")
}

func (amp *Amphipod) ValidatePath(current Position, p Path, m *Map) bool {
    destination := p[len(p)-1]
    destinationTile := m.Tiles[destination]
    currentTile := m.Tiles[current]

    if _, ok := m.Forbidden[destination]; ok {
        return false
    }

    if _, ok := m.Amphipods[destination]; ok {
        return false
    }

    if destinationTile.Room != 0 && destinationTile.Room != amp.Destination {
        return false
    }

    if currentTile.Room == 0 && destinationTile.Room != amp.Destination {
        return false
    }

    if currentTile.Room == amp.Destination {
        if current.Y == m.Depth {
            return false
        }

        done := true
        for y := current.Y; y <= m.Depth; y++ {
            nPos := Position{current.X, y}
            nAmp, _ := m.Amphipods[nPos]
            if nAmp.Destination != amp.Destination {
                done = false
            }
        }
        if done {
            return false
        }
    }

    if destinationTile.Room == amp.Destination {
        for y := destination.Y + 1; y <= m.Depth; y++ {
            nPos := Position{destination.X, y}
            nAmp, ok := m.Amphipods[nPos]
            if !ok {
                return false
            }
            if nAmp.Destination != destinationTile.Room {
                return false
            }
        }
    }

    // cant cross other amp
    for _, pos := range p {
        if _, ok := m.Amphipods[pos]; ok {
            return false
        }
    }

    return true
}

type Tile struct {
    Room rune
}

func FromString(str string) Map {
    depth := 2
    if len(str) == 90 {
        depth = 4
    }
    m := NewMap(11, depth)
    for y, line := range strings.Split(str, "\n") {
        for x, c := range strings.Split(line, "") {
            switch c[0] {
            case '#':
            case '.':
            case ' ':
                continue
            default:
                m.AddAmphipod(Position{x - 1, y - 1}, Amphipod{rune(c[0])})
            }
        }
    }
    return m
}

func FindMin(m Map) int {
    mem := make(map[string]int)
    mem[m.String()] = 0
    var winner *Map = nil
    activeMaps := []Map{m}
    iterations := 0
    for len(activeMaps) > 0 {
        fmt.Println(iterations, len(activeMaps))
        newList := make([]Map, 0)
        for _, current := range activeMaps {
            for _, nMap := range current.Step() {
                checksum := nMap.String()
                if v, ok := mem[checksum]; ok {
                    if nMap.Score >= v {
                        continue
                    }
                }
                if winner != nil && winner.Score < nMap.Score {
                    continue
                }

                if nMap.Finished() {
                    c := nMap
                    winner = &c
                    fmt.Println("SCORE:", nMap.Score)
                } else {
                    newList = append(newList, nMap)
                    mem[checksum] = nMap.Score
                }
            }
        }
        activeMaps = newList
        iterations++
    }
    return winner.Score
}

func Solve(data string, result *orchestration.Result) error {
    a := FindMin(FromString(data))
    result.AddResult(strconv.Itoa(a))

    bInput := make([]string, 0)
    lineRaw := strings.Split(data, "\n")
    bInput = append(bInput, lineRaw[:3]...)
    bInput = append(bInput, []string{"  #D#C#B#A#", "  #D#B#A#C#"}...)
    bInput = append(bInput, lineRaw[3:]...)
    b := FindMin(FromString(strings.Join(bInput, "\n")))
    result.AddResult(strconv.Itoa(b))

    return nil
}

func init() {
    orchestration.MainDispatcher.AddSolver("Day23", orchestration.NewSolver(Solve))
}