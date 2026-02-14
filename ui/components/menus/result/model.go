package result

import "bgscan/ui/shared/layout"

type Model struct {
	layout *layout.Layout
}

func New(layout *layout.Layout) Model {
	return Model{
		layout: layout,
	}
}
