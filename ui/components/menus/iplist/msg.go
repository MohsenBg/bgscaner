package iplist

import "bgscan/core/ipmanager"

type ResultFilesLoadedMsg struct {
	Files []ipmanager.ResultFile
}
