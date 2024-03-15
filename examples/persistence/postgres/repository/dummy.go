package repository

import (
	"context"

	keelpostgres "github.com/foomo/keel/persistence/postgres"
)

type TaskRepository struct {
	persistor *keelpostgres.Persistor
}

// NewTaskRepository constructor
func NewTaskRepository(persistor *keelpostgres.Persistor) *TaskRepository {
	return &TaskRepository{
		persistor: persistor,
	}
}

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
		var id int32
		var description string
		err := rows.Scan(&id, &description)
		if err != nil {
			return nil, err
		}
		ret[id] = description
	}

	return ret, rows.Err()
}

func (r *TaskRepository) Insert(ctx context.Context, description string) error {
	conn, err := r.persistor.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "insert into tasks(description) values($1)", description)
	return err
}

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
