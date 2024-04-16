package main

import (
	"errors"
	"fmt"
	"forum/internal/models"
	"net/http"
	"strconv"
	"strings"
)

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

		app.render(w, http.StatusOK, "create.html", data)
	} else if r.Method == http.MethodPost {
		// First we call r.ParseForm() which adds any data in POST request bodies
		// to the r.PostForm map. This also works in the same way for PUT and PATCH
		// requests. If there are any errors, we use our app.ClientError() helper to
		// send a 400 Bad Request response to the user.
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		// Use the r.PostForm.Get() method to retrieve the title and content
		// from the r.PostForm map.
		title := r.PostForm.Get("title")
		content := r.PostForm.Get("content")

		// The r.PostForm.Get() method always returns the form data as a *string*.
		// However, we're expecting our expires value to be a number, and want to
		// represent it in our Go code as an integer. So we need to manually covert
		// the form data to an integer using strconv.Atoi(), and we send a 400 Bad
		// Request response if the conversion fails.
		expires, err := strconv.Atoi(r.PostForm.Get("expires"))
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		id, err := app.snippets.Insert(title, content, expires)
		if err != nil {
			app.serverError(w, err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
	}
}
