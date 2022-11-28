package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/JacobNewton007/sendchamp-go-test/internal/validator"
)

type Tasks struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	CreatedBy string    `json:"created_by,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateTask(v *validator.Validator, task *Tasks) {
	// Use the Check() method to execute our validation checks. This will add the
	// provided key and error message to the errors map if the check does not evaluate
	// to true. For example, in the first line here we "check that the title is not
	// equal to the empty string". In the second, we "check that the length of the title
	// is less than or equal to 500 bytes" and so on.
	v.Check(task.Title != "", "title", "must be provided")
	v.Check(len(task.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(task.CreatedBy != "", "title", "must be provided")
	v.Check(len(task.CreatedBy) <= 500, "title", "must not be more than 500 bytes long")

}

// Define a movieModel struct type which wraps a sql.DB connection pool.
type TaskModel struct {
	DB *sql.DB
}

// Add a placeholder method for inserting a new record in the movie table
func (m TaskModel) Insert(task *Tasks) (int64, error) {
	// Define the SQL query for inserting a new record in
	// the system-generated data.
	query := `
		INSERT INTO tasks (title, created_by)
		VALUES (?, ?)`

	// Create an args slice containing the values for the placeholder parameters from
	args := []interface{}{task.Title, task.CreatedBy}

	// Create a context with a 3-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameter and scanning the system-
	// generated id, created_at and version values into the task struct.
	result, err := m.DB.ExecContext(ctx, query, args...)

	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Add a placeholder method for fetching a specfic record in the movie table
func (m TaskModel) Get(id int64) (*Tasks, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving the movie data.
	query := `
					SELECT id, created_at, title, created_by, version
					FROM tasks
					WHERE id = ?
					`
	// Declare a Task struct to hold the data returned by the query.
	var task Tasks

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder parameter, and scan the response data into the fields of the
	// Task struct. Importantly, notice that we need to convert the scan target for the

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.CreatedAt,
		&task.Title,
		&task.CreatedBy,
		&task.Version,
	)

	// Handle any errors. If there was no matching task found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &task, nil
}

// Add a placeholder method for updating a specific record in the task table
func (m TaskModel) Update(task *Tasks) error {
	// Declare the SQL query for updating the record and returning the new version numer
	query := `
					UPDATE tasks
					SET title = ?, created_by = ?, version = version + 1
					WHERE id = ? AND version = ?
					RETURNING version
					`
	// Create an args slice containing the value for the placeholder parameters.
	args := []interface{}{
		task.Title,
		task.CreatedBy,
		task.ID,
		task.Version,
	}

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the task struct.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&task.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// Add a placeholder method for deleting a specific record in the task table
func (m TaskModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the task ID is less than 1
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record
	query := `
					DELETE FROM tasks 
					WHERE id = ?
					`
	// Create a context with a 3-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Execute the SQL query using the Exec() method, passing in the id variable as
	// the valur for the placeholder parameter. The Exec method returns a sql.Result
	// object.
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected method on the sql.Result object to get the number of rows
	// affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// If no rows were affected, we know that the task table didn't contain a record // with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
