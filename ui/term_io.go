package termui

import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"os"
)

//func ListenByte1(out <-chan InputByteCallback, quit <-chan struct{}) {
//	go func(out <-chan InputByteCallback, quit <-chan struct{}) {
//		var err error
//		b := make([]byte, 1)
//		for {
//			select {
//			case <-quit:
//				return
//			default:
//				openRawInput()
//				_, err = os.Stdin.Read(b)
//				if err != nil {
//					panic(err)
//				}
//				callback := <-out
//				closeRawInput()
//
//				callback(b[0])
//			}
//		}
//	}(out, quit)
//}

type InputModeCallback func()
type NormalModeCallback func()
type CommandModeCallback func()

type TermMode int

const (
	TermModeNormal TermMode = iota
	TermModeInput
	TermModeCommand
)

type TermIO struct {
	width               int
	height              int
	x                   int
	y                   int
	data                [][]rune
	mode                TermMode
	inputModeCallback   InputModeCallback
	normalModeCallback  NormalModeCallback
	commandModeCallback CommandModeCallback
}

func NewTermIO() *TermIO {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	t := &TermIO{}
	t.resetSize()
	t.setCursor(0, 0)
	return t
}

func (t *TermIO) Close() {
	termbox.Close()
}

func (t *TermIO) BindNormalMode(f NormalModeCallback) {
	t.normalModeCallback = f
}

func (t *TermIO) BindInputMode(f InputModeCallback) {
	t.inputModeCallback = f
}

func (t *TermIO) Listen() {
	go func() {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				t.doEventKey(ev)
			case termbox.EventResize:
				t.resetSize()
			}
		}
	}()
}

func (t *TermIO) resetSize() {
	_ = termbox.Flush()
	t.width, t.height = termbox.Size()
}

func (t *TermIO) setCursor(x int, y int) {
	t.x, t.y = x, y
	termbox.SetCursor(x, y)
}

func (t *TermIO) setNormalMode() {
	t.mode = TermModeNormal
	t.normalModeCallback()
}

func (t *TermIO) setInputMode() {
	t.mode = TermModeInput
	t.inputModeCallback()
}

func (t *TermIO) setCommandMode() {
	t.mode = TermModeCommand
	t.setCursor(0, t.height-1)
}

func (t *TermIO) doEventKey(ev termbox.Event) {
	switch ev.Key {
	case termbox.KeyCtrlC:
		os.Exit(0)
	case termbox.KeyEsc:
		t.setNormalMode()
	default:
		switch t.mode {
		case TermModeNormal:
			t.doInNormalMode(ev)
		case TermModeInput:
			t.doInInputMode(ev)
		case TermModeCommand:
			t.doInCommandMode(ev)
		}
	}
}

func (t *TermIO) doInInputMode(ev termbox.Event) {
	t.print(ev.Ch)
}

func (t *TermIO) doInNormalMode(ev termbox.Event) {
	t.print(ev.Ch)
	switch ev.Ch {
	case 'i':
		t.setInputMode()
	case ':':
		t.setCommandMode()
	}
}

func (t *TermIO) doInCommandMode(ev termbox.Event) {
	if ev.Key == termbox.KeyEsc {
		t.setNormalMode()
		return
	}
	t.print(ev.Ch)
}

func (t *TermIO) print(r rune) {
	termbox.SetCell(t.x, t.y, r, termbox.ColorWhite, termbox.ColorBlack)
	w := runewidth.RuneWidth(r)
	t.x += w
	if t.x >= t.width-1 {
		t.x = 0
		t.y++
	}
	if t.y >= t.height {
		t.y--
		t.Redraw(0, 1)
	}
	t.setCursor(t.x, t.y)
	_ = termbox.Flush()
}

func (t *TermIO) Redraw(startX int, startY int) {
	buf := termbox.CellBuffer()
	for y := startY; y < t.height; y++ {
		for x := startX; x < t.width; x++ {
			srcIdx := y*t.height + x
			srcCell := buf[srcIdx]
			termbox.SetCell(x-startX, y-startY, srcCell.Ch, srcCell.Fg, srcCell.Bg)
		}
	}
	for x := 0; x < t.width; x++ {
		termbox.SetCell(x, t.height-1, ' ', termbox.ColorWhite, termbox.ColorBlack)
	}
	_ = termbox.Flush()
}
