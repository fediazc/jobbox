package main

import (
	"net/http"

	"cazar.fediaz.net/ui"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

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

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
