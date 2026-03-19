package ui

import (
	bubbleTeaOverlay "github.com/rmhubbert/bubbletea-overlay"
)

// OverlayPosition represents a positioning option for overlays.
type OverlayPosition = bubbleTeaOverlay.Position

const (
	Top    = bubbleTeaOverlay.Top
	Right  = bubbleTeaOverlay.Right
	Bottom = bubbleTeaOverlay.Bottom
	Left   = bubbleTeaOverlay.Left
	Center = bubbleTeaOverlay.Center
)

// AddOverlayMsg requests the UI manager to mount a component
// as an overlay on top of the current interface.
type AddOverlayMsg struct {
	Component Component

	XPos OverlayPosition
	YPos OverlayPosition

	XOffset int
	YOffset int
}

// AddNewOverlay creates a message requesting a new overlay
// component to be displayed.
func AddNewOverlay(
	component Component,
	xPos, yPos OverlayPosition,
	xOffset, yOffset int,
) AddOverlayMsg {
	return AddOverlayMsg{
		Component: component,
		XPos:      xPos,
		YPos:      yPos,
		XOffset:   xOffset,
		YOffset:   yOffset,
	}
}
