package main

import (
	"html/template"
	"io"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Contact struct {
	Id    uuid.UUID
	Name  string
	Email string
}

func newContact(name, email string) Contact {
	return Contact{
		Name:  name,
		Email: email,
		Id:    uuid.New(),
	}
}

type Contacts = []Contact

type Data struct {
	Contacts Contacts
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact("John", "wJw7x@example.com"),
			newContact("Jane", "xJw8x@example.com"),
		},
	}
}

func (d *Data) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
}

func (d *Data) indexOf(id uuid.UUID) int {
	for i, contact := range d.Contacts {
		if contact.Id == id {
			return i
		}
	}
	return -1
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.Static("/images", "images")
	e.Static("/css", "css")

	page := newPage()
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"

			return c.Render(422, "createContact", formData)
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)

		c.Render(200, "createContact", newFormData())
		return c.Render(200, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		id := c.Param("id")

		index := page.Data.indexOf(uuid.MustParse(id))
		if index == -1 {
			return c.String(404, "Contact not found")
		}

		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)
		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
