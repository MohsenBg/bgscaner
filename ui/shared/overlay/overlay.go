package overlay

import (
	tea "github.com/charmbracelet/bubbletea"
	o "github.com/rmhubbert/bubbletea-overlay"
)

type Position = o.Position

const (
	Top Position = iota + 1
	Right
	Bottom
	Left
	Center
)

type Overlay interface {
	ID() OverlayID
	Init() tea.Cmd
	Update(tea.Msg) (Overlay, tea.Cmd)
	View() string
}

type OverlayID = string

type AddOverlayMsg struct {
	Overlay Overlay
	XPos    Position
	YPos    Position
	XOffset int
	YOffset int
}

type CloseOverlayMsg struct {
	ID OverlayID
	o.Position
}

func NewAddOverlay(
	ov Overlay,
	xPos, yPos Position,
	xOffset, yOffset int,
) AddOverlayMsg {
	return AddOverlayMsg{
		Overlay: ov,
		XPos:    xPos,
		YPos:    yPos,
		XOffset: xOffset,
		YOffset: yOffset,
	}
}
