package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/tasks/:id", app.requireActivatedUser(app.GetTaskHandler))
	router.HandlerFunc(http.MethodPost, "/v1/tasks", app.requireActivatedUser(app.createTaskHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/tasks/:id", app.requireActivatedUser(app.updateTaskHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/tasks/:id", app.requireActivatedUser(app.deleteTaskHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
