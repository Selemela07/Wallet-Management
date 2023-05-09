package database

type Transaction struct {
	ID          int
	Description string
	Amount      float64
}

func (db *DB) CreateTransaction(t *Transaction) error {
	sqlStatement := `
		INSERT INTO transactions (description, amount)
		VALUES ($1, $2)
		RETURNING id`
	err := db.QueryRow(sqlStatement, t.Description, t.Amount).Scan(&t.ID)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetTransaction(id int) (*Transaction, error) {
	t := &Transaction{}
	sqlStatement := `SELECT id, description, amount FROM transactions WHERE id=$1;`
	err := db.QueryRow(sqlStatement, id).Scan(&t.ID, &t.Description, &t.Amount)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (db *DB) UpdateTransaction(t *Transaction) error {
	sqlStatement := `
		UPDATE transactions
		SET description=$2, amount=$3
		WHERE id=$1;`
	_, err := db.Exec(sqlStatement, t.ID, t.Description, t.Amount)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) DeleteTransaction(id int) error {
	sqlStatement := `DELETE FROM transactions WHERE id=$1;`
	_, err := db.Exec(sqlStatement, id)
	if err != nil {
		return err
	}
	return nil
}
