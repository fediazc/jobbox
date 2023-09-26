package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cazar.fediaz.net/internal/models"
	"cazar.fediaz.net/internal/validator"

	"github.com/julienschmidt/httprouter"
)

type signupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type loginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type jobForm struct {
	Company           string
	Role              string
	Commute           string
	ApplicationStatus string
	Location          string
	DateApplied       time.Time
	Notes             string
	validator.Validator
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, http.StatusOK, "home.gohtml", data)
}

func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = signupForm{}

	app.render(w, http.StatusOK, "signup.gohtml", data)
}

func (app *application) signupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := signupForm{
		Name:     r.PostForm.Get("name"),
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "Password must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.gohtml", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.gohtml", data)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	app.sessionManager.Put(r.Context(), "flash", "Your sign up was successful! Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = signupForm{}

	app.render(w, http.StatusOK, "login.gohtml", data)
}

func (app *application) loginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	form := &loginForm{
		Email:    email,
		Password: password,
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.gohtml", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Invalid email or password")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.gohtml", data)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) dashboard(w http.ResponseWriter, r *http.Request) {
	id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	jobs, err := app.jobs.GetLatest(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Jobs = jobs

	app.render(w, http.StatusOK, "dashboard.gohtml", data)
}

func (app *application) viewJob(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	job, err := app.jobs.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Job = job

	app.render(w, http.StatusOK, "view.gohtml", data)
}

func (app *application) addJob(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = jobForm{}

	app.render(w, http.StatusOK, "add.gohtml", data)
}

func (app *application) addJobPost(w http.ResponseWriter, r *http.Request) {
	form, err := app.parseJobForm(r)
	if err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "add.gohtml", data)
		return
	}

	uid := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	jid, err := app.jobs.Insert(uid, form.Company, form.Role, form.Commute, form.ApplicationStatus, form.Location, form.Notes, form.DateApplied)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Application added successfully!")

	http.Redirect(w, r, fmt.Sprintf("/application/view/%d", jid), http.StatusSeeOther)
}

func (app *application) updateJob(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	job, err := app.jobs.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	form := &jobForm{
		Company:           job.Company,
		Role:              job.Role,
		Commute:           job.Commute,
		ApplicationStatus: job.ApplicationStatus,
		Location:          job.Location,
		Notes:             job.Notes,
		DateApplied:       job.DateApplied,
	}

	data := app.newTemplateData(r)
	data.Form = form
	data.Job = job

	app.render(w, http.StatusOK, "update.gohtml", data)
}

func (app *application) updateJobPost(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	form, err := app.parseJobForm(r)
	if err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}

	if !form.Valid() {
		job := &models.Job{ID: id}

		data := app.newTemplateData(r)
		data.Form = form
		data.Job = job
		app.render(w, http.StatusUnprocessableEntity, "update.gohtml", data)
		return
	}

	err = app.jobs.Update(id, form.Company, form.Role, form.Commute, form.ApplicationStatus, form.Location, form.Notes, form.DateApplied)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Application updated successfully!")

	http.Redirect(w, r, fmt.Sprintf("/application/view/%d", id), http.StatusSeeOther)
}

func (app *application) deleteJob(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	job, err := app.jobs.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Job = job

	app.render(w, http.StatusOK, "delete.gohtml", data)
}

func (app *application) deleteJobPost(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	err = app.jobs.Delete(id)
	if err != nil {
		app.serverError(w, err)
	}

	app.sessionManager.Put(r.Context(), "flash", "Application deleted successfully!")

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
