package repository

import (
	"context"

	keelpostgres "github.com/foomo/keel/persistence/postgres"
)

// TaskRepository stores task descriptions in the tasks table.
type TaskRepository struct {
	persistor *keelpostgres.Persistor
}

// NewTaskRepository returns a TaskRepository backed by persistor.
func NewTaskRepository(persistor *keelpostgres.Persistor) *TaskRepository {
	return &TaskRepository{
		persistor: persistor,
	}
}

// List returns all task descriptions keyed by ID.
func (r *TaskRepository) List(ctx context.Context) (map[int32]string, error) {
	conn, err := r.persistor.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx, "select * from tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := map[int32]string{}

	for rows.Next() {
		var (
			id          int32
			description string
		)

		err := rows.Scan(&id, &description)
		if err != nil {
			return nil, err
		}

		ret[id] = description
	}

	return ret, rows.Err()
}

// Insert adds a task with the given description.
func (r *TaskRepository) Insert(ctx context.Context, description string) error {
	conn, err := r.persistor.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "insert into tasks(description) values($1)", description)

	return err
}

// Drop drops the order_numbers table if it exists.
func (r *TaskRepository) Drop(ctx context.Context) error {
	conn, err := r.persistor.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `DROP TABLE IF EXISTS order_numbers;`); err != nil {
		return err
	}

	return nil
}
