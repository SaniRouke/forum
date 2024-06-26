package main

import (
	"errors"
	"fmt"
	"forum/internal/models"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type snippetCreateForm struct {
	Title       string
	Content     string
	Expires     int
	FieldErrors map[string]string
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets

	app.render(w, http.StatusOK, "home.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.Split(r.URL.Path, "view/")[1])
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = snippet

	// Use the new render helper.
	app.render(w, http.StatusOK, "view.html", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := app.newTemplateData(r)

		// Initialize a new createSnippetForm instance and pass it to the template.
		// Notice how this is also a great opportunity to set any default or
		// 'initial' values for the form --- here we set the initial value for the
		// snippet expiry to 365 days.
		data.Form = snippetCreateForm{
			Expires: 365,
		}

		app.render(w, http.StatusOK, "create.html", data)
	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		// Get the expires value from the form as normal.
		expires, err := strconv.Atoi(r.PostForm.Get("expires"))
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		// Create an instance of the snippetCreateForm struct containing the values
		// from the form and an empty map for any validation errors.
		form := snippetCreateForm{
			Title:       r.PostForm.Get("title"),
			Content:     r.PostForm.Get("content"),
			Expires:     expires,
			FieldErrors: map[string]string{},
		}

		// Update the validation checks so that they operate on the snippetCreateForm
		// instance.
		if strings.TrimSpace(form.Title) == "" {
			form.FieldErrors["title"] = "This field cannot be blank"
		} else if utf8.RuneCountInString(form.Title) > 100 {
			form.FieldErrors["title"] = "This field cannot be more than 100 characters long"
		}

		if strings.TrimSpace(form.Content) == "" {
			form.FieldErrors["content"] = "This field cannot be blank"
		}

		if form.Expires != 1 && form.Expires != 7 && form.Expires != 365 {
			form.FieldErrors["expires"] = "This field must equal 1, 7 or 365"
		}

		// If there are any validation errors re-display the create.tmpl template,
		// passing in the snippetCreateForm instance as dynamic data in the Form
		// field. Note that we use the HTTP status code 422 Unprocessable Entity
		// when sending the response to indicate that there was a validation error.
		if len(form.FieldErrors) > 0 {
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "create.html", data)
			return
		}

		// We also need to update this line to pass the data from the
		// snippetCreateForm instance to our Insert() method.
		id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
		if err != nil {
			app.serverError(w, err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
	}
}
