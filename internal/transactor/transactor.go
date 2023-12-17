package transactor

import (
	"context"
	"database/sql"
)

type transactorInstance struct {
	db *sql.DB
}

func New(db *sql.DB) *transactorInstance {
	return &transactorInstance{db}
}

func (t *transactorInstance) Within(ctx context.Context, tFunc func(ctx context.Context, tx *sql.Tx) error) error {

	// Начало глобальной транзакции
	tx, err := t.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			// откат
			tx.Rollback()
			panic(p)
		}
		//tx.Close()
		//tx.cl
	}()

	err = tFunc(ctx, tx)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()

	// тут функцию для работы (могут быть в разных местах)
	// err = updateDataInTransaction(tx)
	// if err != nil {
	// 	return err
	// }

	// err = insertDataInTransaction(tx)
	// if err != nil {
	// 	return err
	// }

	//return nil
}
