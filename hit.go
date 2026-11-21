package layout

// HitTest returns the visible node whose box is painted last at the canvas
// point p, together with p expressed in that node's local coordinate space
// (relative to the node's border-box origin, the origin of its Rect).
//
// It reads the Rects produced by Layout / LayoutWithPositioning and never
// computes or modifies layout, so it is safe to call after layout, repeatedly,
// and on a tree that is being inspected for rendering.
//
// # Coordinates
//
// A child's Rect is expressed in its parent's coordinate space, so a node's
// canvas origin is the sum of the Rect origins from the root down to it. p is
// given in the root's coordinate space — for a tree laid out as the root, that
// is canvas space (Layout writes the root's Rect origin at (0, 0)). The
// returned Point is p minus the hit node's canvas origin, so it falls inside
// [0, Width) x [0, Height) of the node's border box.
//
// The one exception is position: fixed. Layout positions fixed boxes against
// the viewport rectangle and writes that viewport-relative position directly
// into Rect rather than into the parent's local space, so a fixed box's Rect is
// already a canvas origin. Its descendants then accumulate from there as usual.
//
// # Which node wins
//
// Painting order is document order: a box is painted before its descendants,
// and children are painted in slice order (CSS 2.1 Appendix E). Hit testing
// therefore walks the tree in pre-order and keeps the last visible box whose
// final border box covers p: a descendant beats its ancestor, and an
// overlapping later sibling beats an earlier one. This is the "topmost" node a
// renderer would draw at p.
//
// The engine does not paint, so Style.ZIndex is not consulted (layout does not
// reorder boxes by it) and Style.Transform is not inverted (transforms do not
// affect layout; GetFinalRect gives the transformed bounds). Hit testing uses
// the untransformed Rect of every node.
//
// # Box coverage
//
// A node is a candidate when its Rect has positive width and height and p lies
// in the half-open box [X, X+Width) x [Y, Y+Height). The half-open convention
// matches pixel coverage and CSSOM View hit testing: a point exactly on an
// edge between two tangent boxes belongs to the box on the right/bottom side
// only, so a shared edge clicks the later-painted sibling rather than both.
// A zero-area box (Width or Height <= 0, or NaN) never covers a point, so
// zero-size nodes and their boxes are transparent to hit testing. Negative
// coordinates are ordinary coordinates and hit normally.
//
// A descendant is tested even when p is outside its ancestor's box, because
// overflow is visible by default and absolutely positioned boxes may lie
// outside their parent. The visible test is display: none: such a node and its
// entire subtree generate no boxes and are skipped (CSS Display Level 3 §2.7).
//
// # No hit
//
// HitTest returns (nil, Point{}) when root is nil, when root is display: none,
// when p is outside every visible box, or when layout has not been computed
// (an unlaid-out tree has zero-area Rects, which cover nothing). A nil return
// is the same unambiguous "nothing under the pointer" result in all of these
// cases.
//
// Based on:
// - CSS 2.1 Appendix E: painting order
// - CSSOM View §5.2: elementFromPoint / hit testing
// - CSS Display Level 3 §2.7: display: none generates no boxes
//
// See:
// - https://www.w3.org/TR/CSS21/zindex.html
// - https://www.w3.org/TR/cssom-view/#dom-document-elementfrompoint
// - https://www.w3.org/TR/css-display-3/#box-generation
//
// Example:
//
//	size := layout.LayoutSimple(root, layout.Loose(800, 600))
//	node, local := layout.HitTest(root, layout.Point{X: mouseX, Y: mouseY})
//	if node != nil {
//	    // local is the click position inside node's border box.
//	    fmt.Printf("hit %p at (%.1f, %.1f)\n", node, local.X, local.Y)
//	}
func HitTest(root *Node, p Point) (*Node, Point) {
	if root == nil {
		return nil, Point{}
	}
	var (
		hit       *Node
		hitOrigin Point
	)
	hitTestNode(root, Point{}, true, p, &hit, &hitOrigin)
	if hit == nil {
		return nil, Point{}
	}
	return hit, Point{X: p.X - hitOrigin.X, Y: p.Y - hitOrigin.Y}
}

// hitTestNode visits n and its descendants in paint order, remembering the
// last visible box that covers p. origin is n's parent's canvas origin.
func hitTestNode(n *Node, origin Point, isRoot bool, p Point, hit **Node, hitOrigin *Point) {
	if n == nil || n.Style.Display == DisplayNone {
		return
	}

	// A child's Rect is parent-relative; fixed boxes are the exception because
	// layout writes their viewport-relative Rect directly.
	canvasOrigin := Point{X: origin.X + n.Rect.X, Y: origin.Y + n.Rect.Y}
	if !isRoot && n.Style.Position == PositionFixed {
		canvasOrigin = Point{X: n.Rect.X, Y: n.Rect.Y}
	}

	if hitTestCovers(canvasOrigin, n.Rect, p) {
		*hit = n
		*hitOrigin = canvasOrigin
	}

	// Descend even when p is outside n's box: overflow is visible and absolute
	// boxes may sit outside their parent. Paint order is pre-order, so a later
	// assignment below replaces n (and an earlier sibling's subtree).
	for _, child := range n.Children {
		hitTestNode(child, canvasOrigin, false, p, hit, hitOrigin)
	}
}

// hitTestCovers reports whether p lies in the half-open border box of a node
// whose canvas origin is origin and whose size is r.Width x r.Height.
//
// The positive-area check also rejects NaN and negative dimensions, so an
// invalid or zero-area box never claims a point.
func hitTestCovers(origin Point, r Rect, p Point) bool {
	if !(r.Width > 0) || !(r.Height > 0) {
		return false
	}
	return p.X >= origin.X && p.X < origin.X+r.Width &&
		p.Y >= origin.Y && p.Y < origin.Y+r.Height
}
