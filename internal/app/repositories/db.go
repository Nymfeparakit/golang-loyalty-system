package repositories

import (
	"context"
	"github.com/jmoiron/sqlx"
)

func openConnection(connStr string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createSchema(db *sqlx.DB) error {
	queries := []string{
		`create table if not exists auth_user(
		id serial primary key not null,
		login varchar(64) not null,
		password varchar(128) not null,
		constraint login_unique unique (login)
		);`,
		`create table if not exists user_order(
		 number varchar(64) primary key,
		 uploaded_at timestamptz not null,
		 user_id int not null,
		 status varchar not null default 'NEW',
    	 balance double precision not null default 0,
		 constraint fk_user foreign key(user_id) references auth_user(id),
		 constraint status_values check (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED'))
		);`,
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, q := range queries {
		_, err := tx.Exec(q)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func InitDB(connStr string) (*sqlx.DB, error) {
	db, err := openConnection(connStr)
	if err != nil {
		return nil, err
	}

	if err = createSchema(db); err != nil {
		return nil, err
	}

	return db, nil
}
