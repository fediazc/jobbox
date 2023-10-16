package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"cazar.fediaz.net/internal/validator"
	"github.com/justinas/nosurf"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	errText := http.StatusText(http.StatusInternalServerError)
	http.Error(w, errText, http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		IsAuthenticated: app.isAuthenticated(r),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		Locations:       getValidLocations(),
		Statuses:        getValidStatuses(),
		CommuteTypes:    getValidCommuteTypes(),
		CSRFToken:       nosurf.Token(r),
	}
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exists", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}

func (app *application) parseJobForm(r *http.Request) (*jobForm, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, ErrUnprocessableForm
	}

	dateApplied, err := time.Parse("2006-01-02", r.PostForm.Get("date-applied"))
	if err != nil {
		return nil, ErrBadDate
	}

	form := &jobForm{
		Company:           r.PostForm.Get("company"),
		Role:              r.PostForm.Get("role"),
		Commute:           r.PostForm.Get("commute"),
		ApplicationStatus: r.PostForm.Get("status"),
		Location:          r.PostForm.Get("location"),
		DateApplied:       dateApplied,
		Notes:             r.PostForm.Get("notes"),
	}

	form.CheckField(validator.NotBlank(form.Company), "company", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Company, 100), "company", "This field cannot be longer than 100 characters")
	form.CheckField(validator.NotBlank(form.Role), "role", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Role, 100), "role", "This field cannot be longer than 100 characters")
	form.CheckField(validator.PermittedValue(form.Commute, getValidCommuteTypes()...), "commute", "This field must be one of the values given here")
	form.CheckField(validator.PermittedValue(form.ApplicationStatus, getValidStatuses()...), "status", "This field must be one of the values given here")
	form.CheckField(validator.PermittedValue(form.Location, getValidLocations()...), "location", "This field must be one of the values given here")
	form.CheckField(validator.MaxChars(form.Notes, 280), "notes", "This field cannot be longer than 280 characters")

	return form, nil
}

func getValidLocations() []string {
	return []string{"Unknown", "Alabama", "Alaska", "American Samoa", "Arizona", "Arkansas", "California", "Colorado", "Connecticut", "Delaware", "District of Columbia", "Florida", "Georgia", "Guam", "Hawaii", "Idaho", "Illinois", "Indiana", "Iowa", "Kansas", "Kentucky", "Louisiana", "Maine", "Maryland", "Massachusetts", "Michigan", "Minnesota", "Minor Outlying Islands", "Mississippi", "Missouri", "Montana", "Nebraska", "Nevada", "New Hampshire", "New Jersey", "New Mexico", "New York", "North Carolina", "North Dakota", "Northern Mariana Islands", "Ohio", "Oklahoma", "Oregon", "Pennsylvania", "Puerto Rico", "Rhode Island", "South Carolina", "South Dakota", "Tennessee", "Texas", "U.S. Virgin Islands", "Utah", "Vermont", "Virginia", "Washington", "West Virginia", "Wisconsin", "Wyoming"}
}

func getValidCommuteTypes() []string {
	return []string{"Unknown", "On-Site", "Remote", "Hybrid"}
}

func getValidStatuses() []string {
	return []string{"Applied", "Heard Back", "Interviewing", "Offer Received", "Rejected"}
}
