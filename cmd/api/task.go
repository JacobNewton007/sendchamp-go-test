package main

import (
	// "encoding/json"
	"errors"
	"fmt"
	"net/http"

	// "time"

	"github.com/JacobNewton007/sendchamp-go-test/internal/data"
	"github.com/JacobNewton007/sendchamp-go-test/internal/rabbitmq"
	"github.com/JacobNewton007/sendchamp-go-test/internal/validator"
)

func (app *application) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintln(w, "create a new task")

	var input rabbitmq.AddTask
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.rMq.Publisher(input)

	taskJob := app.rMq.Worker()

	// copy the values from the input struct to a new task struct.
	task := &data.Tasks{
		Title:     taskJob.Title,
		CreatedBy: taskJob.CreatedBy,
	}

	// Initialize a new validator
	v := validator.New()

	// Call the ValidateTask() function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateTask(v, task); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Println("JOB ::::::::::::::", &task)
	app.background(func() {
		id, err := app.models.Tasks.Insert(task)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}
		task.ID = id
	})

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new task in the URL.

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/tasks/%d", task.ID))

	// Write a JSON response with a 201 Created status code, the task data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "task is being processed"}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	// fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) GetTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDparam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific task. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	task, err := app.models.Tasks.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"task": task}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the task ID from the URL
	id, err := app.readIDparam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing task record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	task, err := app.models.Tasks.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Declare an input struct to hold the expected data from the client.
	var input struct {
		Title     *string `json:"title"`
		CreatedBy *string `json:"created_by"`
	}

	// Read the JSON request body into the input struct.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the request body to appropriate fields of the task
	// record
	if input.Title != nil {
		task.Title = *input.Title
	}

	if input.CreatedBy != nil {
		task.CreatedBy = *input.CreatedBy
	}

	// Validate the updated task record, sending the client a 422 unprocessable Entity
	// response if any checks fail.

	v := validator.New()

	if data.ValidateTask(v, task); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// pass the updated task record to our Update() method.
	err = app.models.Tasks.Update(task)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write the updated task record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"task": task}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the task ID from the URL
	id, err := app.readIDparam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the task from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.

	err = app.models.Tasks.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 ok status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "task successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
