package resultlist

import (
	"bgscan/internal/core/result"
)

type UpdateTableMsg struct{}

type SelectResultFileMsg struct{}

type RequestRenameResultFileMsg struct{}

type RequestDeleteResultFileMsg struct{}

type DeleteResultFileMsg struct {
	File result.ResultFile
}
