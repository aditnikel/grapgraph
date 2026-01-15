package replay

import (
	"github.com/aditnikel/grapgraph/src/domain"
	"github.com/aditnikel/grapgraph/src/graph/write"
)

func Replay(w *write.TransferWriter, transfers []domain.Transfer) {
	for _, t := range transfers {
		_ = w.Write(t)
	}
}
