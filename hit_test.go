package layout

import (
	"math"
	"testing"
)

// hitTestEpsilon is the tolerance used when comparing computed Rects and
// local coordinates in the hit-test fixtures.
const hitTestEpsilon = 1e-6

// hitTestRectEqual reports whether two Rects are equal within hitTestEpsilon.
func hitTestRectEqual(a, b Rect) bool {
	return math.Abs(a.X-b.X) <= hitTestEpsilon &&
		math.Abs(a.Y-b.Y) <= hitTestEpsilon &&
		math.Abs(a.Width-b.Width) <= hitTestEpsilon &&
		math.Abs(a.Height-b.Height) <= hitTestEpsilon
}

// hitTestPointEqual reports whether two Points are equal within hitTestEpsilon.
func hitTestPointEqual(a, b Point) bool {
	return math.Abs(a.X-b.X) <= hitTestEpsilon && math.Abs(a.Y-b.Y) <= hitTestEpsilon
}

// hitTestFixture is the fixed tree used by the hit-testing tests.
//
//	root (block 400x450, paints at canvas 0,0)
//	├─ grid (2x2, 100px tracks, canvas 0..200 x 0..200)
//	│  ├─ cellA (nested 2-column grid)      -> inner1, inner2 (50x100 each)
//	│  └─ cellB, cellC, cellD (100x100 each, canvas 100,0 / 0,100 / 100,100)
//	├─ flexRow (400x50, canvas 0..400 x 200..250) -> flexA, flexB
//	└─ overlayContainer (abs, canvas 150..350 x 50..150)
//	   ├─ overlay   (abs, canvas 150..250 x 50..150)
//	   ├─ overlayB  (abs, canvas 250..300 x 50..150; tangent to overlay)
//	   ├─ overlayC  (abs, canvas 230..290 x 100..150; overlaps both)
//	   ├─ chip      (abs, canvas -20..10 x -10..20; negative coordinates)
//	   ├─ zero      (abs, 0x0 at canvas 150,50)
//	   ├─ zeroLine  (abs, 100x0 at canvas 150..250 x 110)
//	   └─ hidden    (abs 200x100 at canvas 150,50, display:none)
//	      └─ hiddenChild (300x300, must never be hit)
type hitTestFixture struct {
	root             *Node
	grid             *Node
	cellA            *Node // nested grid
	cellB            *Node
	cellC            *Node
	cellD            *Node
	inner1           *Node
	inner2           *Node
	flexRow          *Node
	flexA            *Node
	flexB            *Node
	overlayContainer *Node
	overlay          *Node
	overlayB         *Node
	overlayC         *Node
	chip             *Node
	zero             *Node
	zeroLine         *Node
	hidden           *Node
	hiddenChild      *Node
}

// newHitTestFixture builds and lays out the fixture. TestHitTestFixtureLayout
// pins every Rect the hit expectations depend on.
func newHitTestFixture(t *testing.T) *hitTestFixture {
	t.Helper()

	inner1 := &Node{Style: Style{Width: Px(50), Height: Px(100)}}
	inner2 := &Node{Style: Style{Width: Px(50), Height: Px(100)}}
	cellA := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
		},
		Children: []*Node{inner1, inner2},
	}
	cellB := &Node{Style: Style{Width: Px(100), Height: Px(100)}}
	cellC := &Node{Style: Style{Width: Px(100), Height: Px(100)}}
	cellD := &Node{Style: Style{Width: Px(100), Height: Px(100)}}
	grid := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
		},
		Children: []*Node{cellA, cellB, cellC, cellD},
	}

	flexA := &Node{Style: Style{Width: Px(100), Height: Px(50)}}
	flexB := &Node{Style: Style{Width: Px(100), Height: Px(50)}}
	flexRow := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			Width:         Px(400),
			Height:        Px(50),
		},
		Children: []*Node{flexA, flexB},
	}

	overlay := &Node{Style: Style{
		Position: PositionAbsolute, Left: Px(0), Top: Px(0), Width: Px(100), Height: Px(100),
	}}
	overlayB := &Node{Style: Style{
		Position: PositionAbsolute, Left: Px(100), Top: Px(0), Width: Px(50), Height: Px(100),
	}}
	overlayC := &Node{Style: Style{
		Position: PositionAbsolute, Left: Px(80), Top: Px(50), Width: Px(60), Height: Px(50),
	}}
	// The chip's canvas box is -20..10 x -10..20: it straddles the canvas
	// origin, which the grid also occupies, so the later-painted chip wins
	// where the two overlap.
	chip := &Node{Style: Style{
		Position: PositionAbsolute, Left: Px(-170), Top: Px(-60), Width: Px(30), Height: Px(30),
	}}
	zero := &Node{Style: Style{
		Position: PositionAbsolute, Left: Px(0), Top: Px(0), Width: Px(0), Height: Px(0),
	}}
	zeroLine := &Node{Style: Style{
		Position: PositionAbsolute, Left: Px(0), Top: Px(60), Width: Px(100), Height: Px(0),
	}}
	hiddenChild := &Node{Style: Style{Width: Px(300), Height: Px(300)}}
	hidden := &Node{
		Style: Style{
			Position: PositionAbsolute, Left: Px(0), Top: Px(0), Width: Px(200), Height: Px(100),
			Display: DisplayNone,
		},
		Children: []*Node{hiddenChild},
	}
	overlayContainer := &Node{
		Style: Style{
			Display: DisplayBlock, Position: PositionAbsolute,
			Left: Px(150), Top: Px(50), Width: Px(200), Height: Px(100),
		},
		Children: []*Node{overlay, overlayB, overlayC, chip, zero, zeroLine, hidden},
	}

	root := &Node{
		Style: Style{Display: DisplayBlock, Width: Px(400), Height: Px(450)},
		Children: []*Node{
			grid,
			flexRow,
			overlayContainer,
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutWithPositioning(root, Loose(400, 450), Rect{Width: 800, Height: 600}, ctx)

	return &hitTestFixture{
		root: root, grid: grid,
		cellA: cellA, cellB: cellB, cellC: cellC, cellD: cellD,
		inner1: inner1, inner2: inner2,
		flexRow: flexRow, flexA: flexA, flexB: flexB,
		overlayContainer: overlayContainer,
		overlay:          overlay, overlayB: overlayB, overlayC: overlayC,
		chip: chip, zero: zero, zeroLine: zeroLine,
		hidden: hidden, hiddenChild: hiddenChild,
	}
}

// TestHitTestFixtureLayout pins the geometry the hit-test assertions rely on.
// Child Rects are parent-relative; absolute children are positioned against
// the nearest positioned ancestor's padding box (CSS 2.1 §10.1).
func TestHitTestFixtureLayout(t *testing.T) {
	f := newHitTestFixture(t)

	rects := []struct {
		name string
		node *Node
		want Rect
	}{
		{"root", f.root, Rect{0, 0, 400, 450}},
		{"grid", f.grid, Rect{0, 0, 200, 200}},
		{"cellA", f.cellA, Rect{0, 0, 100, 100}},
		{"cellB", f.cellB, Rect{100, 0, 100, 100}},
		{"cellC", f.cellC, Rect{0, 100, 100, 100}},
		{"cellD", f.cellD, Rect{100, 100, 100, 100}},
		{"inner1", f.inner1, Rect{0, 0, 50, 100}},
		{"inner2", f.inner2, Rect{50, 0, 50, 100}},
		{"flexRow", f.flexRow, Rect{0, 200, 400, 50}},
		{"flexB", f.flexB, Rect{100, 0, 100, 50}},
		{"overlayContainer", f.overlayContainer, Rect{150, 50, 200, 100}},
		{"overlay", f.overlay, Rect{0, 0, 100, 100}},
		{"overlayB", f.overlayB, Rect{100, 0, 50, 100}},
		{"overlayC", f.overlayC, Rect{80, 50, 60, 50}},
		{"chip", f.chip, Rect{-170, -60, 30, 30}},
		{"zero", f.zero, Rect{0, 0, 0, 0}},
		{"zeroLine", f.zeroLine, Rect{0, 60, 100, 0}},
	}
	for _, tc := range rects {
		if !hitTestRectEqual(tc.node.Rect, tc.want) {
			t.Errorf("%s.Rect = %+v, want %+v", tc.name, tc.node.Rect, tc.want)
		}
	}
}

// TestHitTestPaintOrderAndLocalCoordinates is the main table-driven check: the
// hit node is the deepest, last-painted visible box covering the point, and
// the local coordinate is the canvas point minus that node's accumulated
// canvas origin.
func TestHitTestPaintOrderAndLocalCoordinates(t *testing.T) {
	f := newHitTestFixture(t)

	tests := []struct {
		name      string
		p         Point
		wantNode  *Node
		wantLocal Point
	}{
		// Nested grid: a descendant beats its ancestor; within the grid, the
		// later-painted sibling owns a shared edge.
		{"inner grid first cell", Point{10, 10}, f.inner1, Point{10, 10}},
		{"nested grid second cell beats its grid ancestor", Point{60, 10}, f.inner2, Point{10, 10}},
		{"grid sibling right at shared edge x=100", Point{100, 10}, f.cellB, Point{0, 10}},
		{"grid sibling below at shared edge y=100", Point{10, 100}, f.cellC, Point{10, 0}},
		{"grid cell D offsets accumulate grid+cell", Point{110, 160}, f.cellD, Point{10, 60}},

		// Absolute overlay container paints after the grid; its children paint
		// after it. The overlay occupies canvas 150..350 x 50..150.
		{"overlay child beats overlay container", Point{160, 60}, f.overlay, Point{10, 10}},
		{"overlay origin is inclusive", Point{150, 50}, f.overlay, Point{0, 0}},
		{"tangent overlay sibling owns the shared edge", Point{250, 60}, f.overlayB, Point{0, 10}},
		{"overlayB exclusive right edge falls through to container", Point{300, 60}, f.overlayContainer, Point{150, 10}},
		{"overlapping later sibling wins over overlay", Point{240, 120}, f.overlayC, Point{10, 20}},
		{"overlapping later sibling wins over overlayB", Point{280, 120}, f.overlayC, Point{50, 20}},
		{"absolute chip beats the grid cell it covers", Point{5, 5}, f.chip, Point{25, 15}},

		// Flex and block containers.
		{"flex item second", Point{110, 210}, f.flexB, Point{10, 10}},
		{"root block covers empty area", Point{390, 440}, f.root, Point{390, 440}},

		// Negative canvas coordinates are ordinary coordinates.
		{"negative-coordinate chip", Point{-5, 0}, f.chip, Point{15, 10}},

		// Zero-area boxes never cover a point, even when painted last.
		{"zero-area node is transparent", Point{150, 50}, f.overlay, Point{0, 0}},
		{"zero-height line is transparent", Point{160, 110}, f.overlay, Point{10, 60}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node, local := HitTest(f.root, tc.p)
			if node != tc.wantNode {
				t.Fatalf("HitTest(%v) node = %p, want %p", tc.p, node, tc.wantNode)
			}
			if !hitTestPointEqual(local, tc.wantLocal) {
				t.Errorf("HitTest(%v) local = %+v, want %+v", tc.p, local, tc.wantLocal)
			}
			// The local coordinate must fall inside the node's half-open box.
			if local.X < 0 || local.X >= node.Rect.Width || local.Y < 0 || local.Y >= node.Rect.Height {
				t.Errorf("local %+v outside node %+v", local, node.Rect)
			}
		})
	}
}

// TestHitTestDisplayNoneSkipsSubtree: display:none generates no boxes, so both
// the hidden node and its entire subtree are invisible to hit testing even
// though the positioned pass may still have written a Rect (CSS Display 3
// §2.7). A display:none root has no hit at all.
func TestHitTestDisplayNoneSkipsSubtree(t *testing.T) {
	f := newHitTestFixture(t)

	// The hidden panel covers canvas 150..350 x 50..150 and paints last; the
	// overlay underneath it in paint order must still be the hit.
	if !hitTestRectEqual(f.hidden.Rect, Rect{0, 0, 200, 100}) {
		t.Fatalf("hidden.Rect = %+v, want (0,0,200,100)", f.hidden.Rect)
	}
	node, local := HitTest(f.root, Point{160, 60})
	if node != f.overlay {
		t.Fatalf("display:none panel must not be hit; got %p want overlay %p", node, f.overlay)
	}
	if !hitTestPointEqual(local, Point{10, 10}) {
		t.Errorf("overlay local = %+v, want (10,10)", local)
	}
	// The hidden subtree is never reached either.
	if node, _ := HitTest(f.root, Point{300, 140}); node == f.hiddenChild {
		t.Error("display:none descendant was hit")
	}

	// A tree whose only box is display:none has no hit.
	hiddenOnly := &Node{
		Style:    Style{Display: DisplayBlock, Width: Px(100), Height: Px(100)},
		Children: []*Node{{Style: Style{Display: DisplayNone, Width: Px(100), Height: Px(100)}}},
	}
	LayoutSimple(hiddenOnly, Loose(100, 100))
	if node, local := HitTest(hiddenOnly.Children[0], Point{10, 10}); node != nil || local != (Point{}) {
		t.Errorf("display:none root hit = (%p, %+v), want (nil, zero)", node, local)
	}

	// A nil root has no hit.
	if node, local := HitTest(nil, Point{0, 0}); node != nil || local != (Point{}) {
		t.Errorf("nil root hit = (%p, %+v), want (nil, zero)", node, local)
	}
}

// TestHitTestBoundariesAndNoLayout documents the edge cases that callers rely
// on: points outside every box, unlaid-out trees, and points exactly on an
// outer edge.
func TestHitTestBoundariesAndNoLayout(t *testing.T) {
	f := newHitTestFixture(t)

	// Outside the root and every overlay.
	if node, local := HitTest(f.root, Point{-200, -200}); node != nil || local != (Point{}) {
		t.Errorf("outside hit = (%p, %+v), want (nil, zero)", node, local)
	}

	// A point on the root's bottom/right edge is outside its half-open box.
	if node, _ := HitTest(f.root, Point{400, 450}); node != nil {
		t.Errorf("point on exclusive root edge hit %p, want nil", node)
	}
	if node, _ := HitTest(f.root, Point{400, 100}); node != nil {
		t.Errorf("point on exclusive right edge hit %p, want nil", node)
	}
	// Just inside the root's box belongs to the root (no child there).
	if node, local := HitTest(f.root, Point{399, 449}); node != f.root || !hitTestPointEqual(local, Point{399, 449}) {
		t.Errorf("point inside root = (%p, %+v), want (root, (399,449))", node, local)
	}

	// Layout has not run: every Rect is zero-area, so nothing is hit.
	fresh := &Node{
		Style:    Style{Display: DisplayBlock, Width: Px(100), Height: Px(100)},
		Children: []*Node{{Style: Style{Width: Px(50), Height: Px(50)}}},
	}
	if node, local := HitTest(fresh, Point{10, 10}); node != nil || local != (Point{}) {
		t.Errorf("unlaid-out tree hit = (%p, %+v), want (nil, zero)", node, local)
	}

	// A zero-size root never covers a point even after layout.
	zeroRoot := &Node{Style: Style{Display: DisplayBlock}}
	LayoutSimple(zeroRoot, Loose(100, 100))
	if node, _ := HitTest(zeroRoot, Point{0, 0}); node != nil {
		t.Errorf("zero-size root hit %p, want nil", node)
	}
}

// TestHitTestFixedUsesViewportOrigin: layout stores position:fixed Rects in
// viewport space rather than the parent's local space, so the hit test must
// not accumulate ancestor offsets for a fixed box. Its own descendants still
// accumulate from that viewport origin.
func TestHitTestFixedUsesViewportOrigin(t *testing.T) {
	child := &Node{Style: Style{Width: Px(4), Height: Px(4)}}
	fixed := &Node{
		Style: Style{
			Position: PositionFixed, Left: Px(50), Top: Px(40), Width: Px(10), Height: Px(10),
		},
		Children: []*Node{child},
	}
	rel := &Node{
		Style: Style{
			Display: DisplayBlock, Position: PositionRelative,
			Left: Px(10), Top: Px(10), Width: Px(50), Height: Px(50),
		},
		Children: []*Node{fixed},
	}
	root := &Node{
		Style:    Style{Display: DisplayBlock, Width: Px(200), Height: Px(200)},
		Children: []*Node{rel},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutWithPositioning(root, Loose(200, 200), Rect{Width: 200, Height: 200}, ctx)

	if !hitTestRectEqual(fixed.Rect, Rect{50, 40, 10, 10}) {
		t.Fatalf("fixed.Rect = %+v, want (50,40,10,10) in viewport space", fixed.Rect)
	}

	node, local := HitTest(root, Point{51, 41})
	if node != child {
		t.Fatalf("HitTest inside fixed child = %p, want child %p", node, child)
	}
	if !hitTestPointEqual(local, Point{1, 1}) {
		t.Errorf("fixed child local = %+v, want (1,1)", local)
	}

	// A point in the fixed box outside the 4x4 child hits the fixed box; the
	// relative ancestor's (10,10) offset must not be applied.
	node, local = HitTest(root, Point{55, 45})
	if node != fixed {
		t.Fatalf("HitTest inside fixed = %p, want fixed %p", node, fixed)
	}
	if !hitTestPointEqual(local, Point{5, 5}) {
		t.Errorf("fixed local = %+v, want (5,5)", local)
	}
}

// TestHitTestDoesNotModifyLayout: hit testing is read-only. Every Rect in the
// tree must be unchanged after a batch of hit tests, and the tree shape and
// node order must be untouched.
func TestHitTestDoesNotModifyLayout(t *testing.T) {
	f := newHitTestFixture(t)

	nodes := f.root.DescendantsAndSelf()
	before := make([]Rect, len(nodes))
	for i, n := range nodes {
		before[i] = n.Rect
	}

	points := []Point{
		{10, 10}, {60, 10}, {160, 60}, {250, 60}, {240, 120},
		{-5, 0}, {300, 60}, {400, 450}, {-200, -200}, {110, 210},
	}
	for _, p := range points {
		HitTest(f.root, p)
	}

	after := f.root.DescendantsAndSelf()
	if len(after) != len(nodes) {
		t.Fatalf("tree shape changed: %d nodes, want %d", len(after), len(nodes))
	}
	for i, n := range after {
		if n != nodes[i] {
			t.Fatalf("node order changed at %d", i)
		}
		if n.Rect != before[i] {
			t.Errorf("HitTest modified Rect: node %d %+v -> %+v", i, before[i], n.Rect)
		}
	}
}
