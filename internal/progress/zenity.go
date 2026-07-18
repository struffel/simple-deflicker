package progress

import (
	"fmt"

	"github.com/ncruces/zenity"
)

// ZenityUpdater is an implementation of the Updater interface that displays
// progress using a native zenity progress dialog.
type ZenityUpdater struct {
	title string
	dlg   zenity.ProgressDialog
}

// NewZenityUpdater creates a ZenityUpdater that will show a progress dialog
// with the given title once Start is called.
func NewZenityUpdater(title string) *ZenityUpdater {
	return &ZenityUpdater{title: title}
}

func (z *ZenityUpdater) Start() {
	dlg, err := zenity.Progress(
		zenity.Title(z.title),
		zenity.MaxValue(100),
		zenity.NoCancel(),
	)
	if err != nil {
		return
	}
	z.dlg = dlg
}

func (z *ZenityUpdater) Increment(msg string, phase string, completed int, ofTotal int) {
	if z.dlg == nil {
		return
	}
	if ofTotal > 0 {
		z.dlg.Value(completed * 100 / ofTotal)
	}
	z.dlg.Text(fmt.Sprintf("%s: %s (%d/%d)", phase, msg, completed, ofTotal))
}

func (z *ZenityUpdater) Finish() {
	if z.dlg == nil {
		return
	}
	z.dlg.Complete()
	z.dlg.Close()
}
