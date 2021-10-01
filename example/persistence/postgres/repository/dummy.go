package repository

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type TaskRepository struct {
	conn *pgx.Conn
}

// NewTaskRepository constructor
func NewTaskRepository(conn *pgx.Conn) *TaskRepository {
	return &TaskRepository{
		conn: conn,
	}
}

func (r *TaskRepository) List(ctx context.Context) (map[int32]string, error) {
	rows, err := r.conn.Query(ctx, "select * from tasks")
	if err != nil {
		return nil, err
	}

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
	_, err := r.conn.Exec(context.Background(), "insert into tasks(description) values($1)", description)
	return err
}

func (r *TaskRepository) Drop(ctx context.Context) error {
	if _, err := r.conn.Exec(ctx, `DROP TABLE IF EXISTS order_numbers;`); err != nil {
		return err
	}
	return nil
}
