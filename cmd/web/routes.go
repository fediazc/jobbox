package main

import (
	"net/http"

	"cazar.fediaz.net/ui"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	dynamic := alice.New(app.sessionManager.LoadAndSave, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.signup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.signupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.login))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.loginPost))

	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.logout))
	router.Handler(http.MethodGet, "/dashboard", protected.ThenFunc(app.dashboard))
	router.Handler(http.MethodGet, "/application/add", protected.ThenFunc(app.addJob))
	router.Handler(http.MethodPost, "/application/add", protected.ThenFunc(app.addJobPost))

	validated := protected.Append(app.validateAccess)

	router.Handler(http.MethodGet, "/application/view/:id", validated.ThenFunc(app.viewJob))
	router.Handler(http.MethodGet, "/application/update/:id", validated.ThenFunc(app.updateJob))
	router.Handler(http.MethodPost, "/application/update/:id", validated.ThenFunc(app.updateJobPost))
	router.Handler(http.MethodGet, "/application/delete/:id", validated.ThenFunc(app.deleteJob))
	router.Handler(http.MethodPost, "/application/delete/:id", validated.ThenFunc(app.deleteJobPost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
