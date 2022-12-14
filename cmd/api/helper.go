package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	// "net/url"
	"strconv"
	// "strings"

	// "github.com/JacobNewton007/sendchamp-go-test/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDparam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// Define an envelope type
type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use http.maxBytesReader() to limit the size of the request body to 1MB

	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the target destination.
	err := dec.Decode(dst)
	if err != nil {
		// if there is an error during decoding, start the triager
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var InvalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case errors.As(err, &InvalidUnmarshalError):
			panic(err)

		default:
			return err
		}

	}
	return nil
}

// func (app *application) readString(qs url.Values, key string, defaultValue string) string {
// 	// Extract the value for a given key from the query string.
// 	// if no key exists this will return empty string

// 	s := qs.Get(key)

// 	// if no key exists (or the value is empty) then return the default value.
// 	if s == "" {
// 		return defaultValue
// 	}

// 	// Otherwise return the string.
// 	return s
// }

// The readCSV() helper reads a string value from the query string and then splits it
// into a slice on the comma character. If no matching key could be found, it returns
// the provided default value.
// func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
// 	// Extract the value from the query string.
// 	csv := qs.Get(key)

// 	// if no key exists (or the value is empty) then return the default value.
// 	if csv == "" {
// 		return defaultValue
// 	}

// 	// Otherwise parse the value into a []string slice and return it.
// 	return strings.Split(csv, ",")
// }

// The readInt() helper reads a string value from the query string and converts it to an
// integer before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we record an
// error message in the provided Validator instance.
// func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
// 	// Extract the value from the query string.
// 	s := qs.Get(key)

// 	// if no key exists (or the value is empty) then return the default value.
// 	if s == "" {
// 		return defaultValue
// 	}

// 	// Try to convert the value to an int. If this fails, add an error message to the
// 	// validator instance and return the default value.

// 	i, err := strconv.Atoi(s)
// 	if err != nil {
// 		v.AddError(key, "must be an integer value")
// 		return defaultValue
// 	}

// 	// Otherwise, return the converted integer value.
// 	return i
// }

// The background() helper accepts an arbitrary function as a parameter.
func (app *application) background(fn func()) {

	app.wg.Add(1)
	// Launch a background goroutine.
	go func() {
		// recover any panic
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		// Execute the arbitrary function that we passed as the parameter.
		fn()
	}()
}
