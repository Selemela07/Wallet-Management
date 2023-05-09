/*
package main

import (
	"fmt"
	"log"
	"your_project_path/database"
)

func main() {
	db, err := database.New()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	// Create a new transaction
	newTransaction := &database.Transaction{
		Description: "Sample transaction",
		Amount:      100.0,
	}
	err = db.CreateTransaction(newTransaction)
	if err != nil {
		log.Fatalf("Error creating transaction: %v", err)
	}
	fmt.Printf("New transaction created with ID: %d\n", newTransaction.ID)

	// Get a transaction by ID
	transaction, err := db.GetTransaction(newTransaction.ID)
	if err != nil {
		log.Fatalf("Error getting transaction: %v", err)
	}
	fmt.Printf("Transaction: %+v\n", transaction)

	// Update a transaction
	transaction.Description = "Updated transaction"
	transaction.Amount = 150.0
	err = db.UpdateTransaction(transaction)
	if err != nil {
		log.Fatalf("Error updating transaction: %v", err)
	}
	fmt.Println("Transaction updated")

	// Delete a transaction
	err = db.DeleteTransaction(transaction.ID)
	if err != nil {
		log.Fatalf("Error deleting transaction: %v", err)
	}
	fmt.Println("Transaction deleted")
}
*/