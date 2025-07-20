package tui

import (
	"crudmonkey/crudui"
	"crudmonkey/models"

	"github.com/gdamore/tcell/v2"
	"github.com/gobuffalo/pop/v6"
	"github.com/rivo/tview"
)

func Tui() {
	app := tview.NewApplication()
	pages := tview.NewPages()

	// --- Menu List ---
	menu := tview.NewList().
		//SetBorder(true).
		//SetTitle(" Main Menu ").
		AddItem("Home", "Go to home screen", 'h', func() {
			pages.SwitchToPage("home")
		}).
		AddItem("Settings", "Configure options", 's', func() {
			pages.SwitchToPage("settings")
		}).
		AddItem("Quit", "Exit the app", 'q', func() {
			app.Stop()
		})

	// --- Home Page ---
	homePage := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("🏠 Welcome to Home!\n\nPress ESC to return.")
	homePage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("menu")
			return nil
		}
		return event
	})

	// --- Settings Page ---
	settingsForm := tview.NewForm().
		AddInputField("Username", "", 20, nil, nil).
		AddPasswordField("Password", "", 20, '*', nil).
		AddButton("Save", func() { pages.SwitchToPage("menu") }).
		AddButton("Cancel", func() { pages.SwitchToPage("menu") })
	settingsForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("menu")
			return nil
		}
		return event
	})

	db, err := pop.Connect("development")
	if err != nil {
		panic(err)
	}

	// --- Assemble Pages ---
	pages.
		AddPage("menu", menu, true, true).
		AddPage("home", homePage, true, false).
		AddPage("settings", settingsForm, true, false)

	crudui.RegisterCRUD(app, pages, menu, db, &models.Blog{}, "blogs")
	crudui.RegisterCRUD(app, pages, menu, db, &models.User{}, "users")
	// --- Run ---
	if err := app.
		SetRoot(pages, true).
		EnableMouse(true).
		SetFocus(menu).
		Run(); err != nil {
		panic(err)
	}
}
