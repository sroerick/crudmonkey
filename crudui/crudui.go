package crudui

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gobuffalo/pop/v6"
	"github.com/rivo/tview"
)

// FieldDescriptor holds info about a struct field for form generation
type FieldDescriptor struct {
	Name   string       // Go struct field name
	DBName string       // db tag name
	Type   reflect.Type // Go type of the field
}

// RegisterCRUD introspects your model struct and adds CRUD pages to the TUI
func RegisterCRUD(
	app *tview.Application,
	pages *tview.Pages,
	menu *tview.List,
	db *pop.Connection,
	model interface{},
	resourceName string,
) {
	// Derive field metadata
	desc := describeModel(model)
	// Add "Manage" menu entry
	menu.AddItem(
		strings.Title(resourceName),
		fmt.Sprintf("Manage %s", resourceName),
		0,
		func() { pages.SwitchToPage(resourceName) },
	)
	// Add "New" menu entry
	menu.AddItem(
		"New "+strings.Title(resourceName),
		"Create a new "+resourceName,
		'n',
		func() {
			form := buildFormPage(app, pages, db, model, desc, resourceName, "")
			pages.AddPage(resourceName+"_new", form, true, true)
		},
	)
	// Build list page
	list := buildListPage(app, pages, db, model, desc, resourceName)
	pages.AddPage(resourceName, list, true, false)
}

// buildListPage creates a table view of all records
func buildListPage(
	app *tview.Application,
	pages *tview.Pages,
	db *pop.Connection,
	model interface{},
	fields []FieldDescriptor,
	resourceName string,
) tview.Primitive {
	tbl := tview.NewTable()
	tbl.SetBorder(true)
	tbl.SetTitle(strings.Title(resourceName))

	// header row
	tbl.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
	for i, f := range fields {
		tbl.SetCell(0, i+1,
			tview.NewTableCell(strings.Title(f.DBName)).SetSelectable(false))
	}

	// dynamic loading of []T
	tType := reflect.TypeOf(model)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}
	sliceType := reflect.SliceOf(tType)
	slicePtr := reflect.New(sliceType).Interface() // *[]T
	if err := db.All(slicePtr); err != nil {
		panic(fmt.Errorf("failed to load %s: %w", resourceName, err))
	}
	sliceVal := reflect.ValueOf(slicePtr).Elem()
	for r := 0; r < sliceVal.Len(); r++ {
		rec := sliceVal.Index(r)
		if rec.Kind() == reflect.Ptr {
			rec = rec.Elem()
		}

		// ID column
		idField := rec.FieldByName("ID")
		tbl.SetCell(r+1, 0,
			tview.NewTableCell(fmt.Sprint(idField.Interface())))

		// other columns
		for c, f := range fields {
			fv := rec.FieldByName(f.Name)
			tbl.SetCell(r+1, c+1,
				tview.NewTableCell(fmt.Sprint(fv.Interface())))
		}
	}

	// key handling: ESC to go back; Enter to edit
	tbl.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			pages.SwitchToPage("menu")
		}
	})
	tbl.SetSelectedFunc(func(row, col int) {
		id := tbl.GetCell(row, 0).Text
		ed := buildFormPage(app, pages, db, model, fields, resourceName, id)
		pages.AddPage(resourceName+"_edit_"+id, ed, true, true)
	})
	// add "n" key to create new
	tbl.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'n' {
			form := buildFormPage(app, pages, db, model, fields, resourceName, "")
			pages.AddPage(resourceName+"_new", form, true, true)
			return nil
		}
		return event
	})

	return tbl
}

// describeModel reads `db` tags and returns fields to include in forms
func describeModel(model interface{}) []FieldDescriptor {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var fields []FieldDescriptor
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}
		if tag == "id" || tag == "created_at" || tag == "updated_at" {
			continue
		}
		fields = append(fields, FieldDescriptor{Name: f.Name, DBName: tag, Type: f.Type})
	}
	return fields
}

// buildFormPage generates a create/edit form
func buildFormPage(
	app *tview.Application,
	pages *tview.Pages,
	db *pop.Connection,
	model interface{},
	fields []FieldDescriptor,
	resourceName, id string,
) tview.Primitive {
	form := tview.NewForm()
	// input fields
	for _, f := range fields {
		initial := ""
		if id != "" {
			// load record to populate initial (omitted)
		}
		form.AddInputField(strings.Title(f.DBName), initial, 20, nil, nil)
	}
	// Save button: differentiate new vs edit
	form.AddButton("Save", func() {
		instVal := reflect.New(reflect.TypeOf(model).Elem())
		inst := instVal.Interface()
		// populate inst via form.GetFormItem
		switch len(fields) {
		}
		for i, f := range fields {
			input := form.GetFormItem(i).(*tview.InputField)
			raw := input.GetText()
			field := instVal.Elem().FieldByName(f.Name)
			switch field.Kind() {
			case reflect.String:
				field.SetString(raw)
				// TODO: parse other kinds (int, time.Time, UUID)
			}
		}
		if id == "" {
			db.Create(inst)
		} else {
			instVal.Elem().FieldByName("ID").SetString(id)
			db.Update(inst)
		}
		pages.SwitchToPage(resourceName)
	})
	form.AddButton("Cancel", func() {
		pages.SwitchToPage(resourceName)
	})
	form.SetCancelFunc(func() {
		pages.SwitchToPage(resourceName)
	})

	return form
}
