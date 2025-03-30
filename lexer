package main

type ItemType int

const (
	ItemError ItemType = iota

	ItemEOF

	ItemArtifact       // e.g: `file.ext`, `location/file.ext`
	ItemDirectory      // e.g: `.../location`
	ItemDo             // do keyword
	ItemIdentifier     // namespace, procedure
	ItemIn             // in keyword
	ItemLeftCurly      // { symbol
	ItemOut            // out keyword
	ItemTransformArrow // > operator
	ItemRightCurly     // } symbol
	ItemUse            // use keyword
)

const eof = -1

type Item struct {
	typ ItemType
	val string
}

type Lexer struct {
	name  string
	input string // the string being scanned
	start int    // start position of this item
	pos   int    // current position in the input
	width int    // width of the next rune
	items chan Item
}

// StateFn represents the state of the scanner as a function that
// return the next state.
type StateFn func(*Lexer) StateFn

func lexStatement(l *Lexer) StateFn {
	for {

	}
	return nil // stop the run loop
}

// returns the next rune in the input
func (l *Lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRunInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// skips over the pending input before this point
func (l *Lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune
func (l *Lexer) backup() {
	l.pos -= l.width
}

// returns the next rune in the input but does not consume it
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *Lexer) emit(t ItemType) {
	l.items <- Item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *Lexer) run() {
	for state := lexStatement; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

func lex(name, input string) (*Lexer, chan Item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan Item),
	}
	go l.run()
	return l, l.items
}
