package gui

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func (gui *Gui) SetKeybindings() {
	gui.InputPathKeybinding()
	gui.EntryManagerKeybinding()
}

// globalKeybinding
func (gui *Gui) GlobalKeybinding(event *tcell.EventKey) {
	switch {
	// go to input view
	case event.Key() == tcell.KeyTab:
		gui.App.SetFocus(gui.InputPath)

	// go to previous history
	case event.Key() == tcell.KeyCtrlH:
		history := gui.HistoryManager.Previous()
		if history != nil {
			gui.InputPath.SetText(history.Path)
			gui.EntryManager.SetEntries(history.Path)
			gui.EntryManager.Select(history.RowIdx, 0)
		}

	// go to next history
	case event.Key() == tcell.KeyCtrlL:
		history := gui.HistoryManager.Next()
		if history != nil {
			gui.InputPath.SetText(history.Path)
			gui.EntryManager.SetEntries(history.Path)
			gui.EntryManager.Select(history.RowIdx, 0)
		}

	// go to previous dir
	case event.Rune() == 'h':
		path := filepath.Dir(gui.InputPath.GetText())
		entry := gui.EntryManager.GetSelectEntry()
		if path != "" {
			gui.InputPath.SetText(path)
			gui.EntryManager.SetEntries(path)
			gui.EntryManager.Select(1, 0)
			gui.EntryManager.SetOffset(0, 0)
			entry = gui.EntryManager.GetSelectEntry()
			gui.Preview.UpdateView(gui, entry)
		}

	// go to parent dir
	case event.Rune() == 'l':
		if !hasEntry(gui) {
			return
		}

		entry := gui.EntryManager.GetSelectEntry()
		row, _ := gui.EntryManager.GetSelection()

		if entry.IsDir {
			if len(gui.EntryManager.SetEntries(entry.PathName)) == 0 {
				return
			}
			gui.HistoryManager.Save(row, filepath.Join(filepath.Dir(gui.InputPath.GetText()), entry.Path))
			gui.InputPath.SetText(entry.PathName)
			gui.EntryManager.Select(1, 0)
			gui.EntryManager.SetOffset(1, 0)
			entry := gui.EntryManager.GetSelectEntry()
			gui.Preview.UpdateView(gui, entry)
		}
	}
}

func (gui *Gui) EntryManagerKeybinding() {
	gui.EntryManager.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			gui.App.Stop()
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		gui.GlobalKeybinding(event)
		switch {
		// cut entry
		case event.Rune() == 'd':
			if !hasEntry(gui) {
				return event
			}

			gui.Confirm("do you want to remove this?", "remove", gui.EntryManager, func() {
				if err := gui.RemoveFile(); err != nil {
					modal := tview.NewModal().SetText(err.Error()).AddButtons([]string{"yes"})
					gui.Modal(modal, 50, 50)
				}
			})

		// copy entry
		case event.Rune() == 'y':
			if !hasEntry(gui) {
				return event
			}
			gui.Register.CopySources = append(gui.Register.CopySources, gui.EntryManager.GetSelectEntry())

		// paset entry
		case event.Rune() == 'p':
			//for _, source := range gui.Register.MoveSources {
			//	dest := filepath.Join(gui.InputPath.GetText(), filepath.Base(source))
			//	if err := os.Rename(source, dest); err != nil {
			//		log.Printf("cannot copy or move the file: %s", err)
			//	}
			//}

			// TODO implement file copy
			//for _, source := range gui.Register.CopyResources {
			//dest := filepath.Join(gui.InputPath.GetText(), filepath.Base(source))
			//}

			//gui.EntryManager.SetEntries(gui.InputPath.GetText())

		// edit file with $EDITOR
		case event.Rune() == 'e':
			editor := os.Getenv("EDITOR")
			if editor == "" {
				log.Println("$EDITOR is empty, please set $EDITOR")
				return event
			}

			entry := gui.EntryManager.GetSelectEntry()
			if entry == nil {
				log.Println("cannot get entry")
				return event
			}

			gui.App.Suspend(func() {
				if err := gui.ExecCmd(true, editor, entry.PathName); err != nil {
					log.Printf("%s: %s\n", ErrEdit, err)
				}
			})
		case event.Rune() == 'q':
			gui.Stop()
		}

		return event
	})

	gui.EntryManager.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 {
			f := gui.EntryManager.Entries()[row-1]
			gui.Preview.UpdateView(gui, f)
		}
	})

}

func (gui *Gui) RemoveFile() error {
	entry := gui.EntryManager.GetSelectEntry()
	if entry == nil {
		return nil
	}

	if entry.IsDir {
		return nil
	}

	_, err := os.Stat(entry.PathName)
	if os.IsNotExist(err) {
		log.Println(err)
		return err
	}

	if err := os.Remove(entry.PathName); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (gui *Gui) InputPathKeybinding() {
	gui.InputPath.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			gui.App.Stop()
		}

		if key == tcell.KeyEnter {
			path := gui.InputPath.GetText()
			path = os.ExpandEnv(path)
			gui.InputPath.SetText(path)
			row, _ := gui.EntryManager.GetSelection()
			gui.HistoryManager.Save(row, path)
			gui.EntryManager.SetEntries(path)
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			gui.App.SetFocus(gui.EntryManager)
		}

		return event
	})
}
