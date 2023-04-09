package db

import (
	"context"
	"entriesMicroService/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	YYYYMMDD = "2006-01-02"
)

func CreateRandomUser() User {
	return User{
		Username:       util.RandomString(6),
		HashedPassword: util.RandomString(6),
		FullName:       util.RandomFullName(),
		Email:          util.RandomEmail(),
		TotalExpenses:  util.RandomMoney(),
	}
}

func CreateRandomEntry(user User) Entry {
	date, _ := time.Parse(YYYYMMDD, "2022-12-11")

	return Entry{
		ID:      95,
		Owner:   user.Username,
		Name:    util.RandomString(6),
		DueDate: date,
		Amount:  5,
	}
}

func TestAddEntryTx(t *testing.T) {
	store := NewStore(testDB)

	user := CreateRandomUser()

	date, err := GetMadeUpDate("2022-12-11")
	require.NoError(t, err)
	require.NotEmpty(t, date)

	addEntryTxParams := AddEntryTxParams{
		Username: user.Username,
		Name:     util.RandomString(6),
		DueDate:  date,
		Amount:   util.RandomMoney(),
		Category: util.RandomString(6),
	}

	result, err := store.AddEntryTx(context.Background(), addEntryTxParams)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	require.NotEmpty(t, result.Entry)
	require.Equal(t, addEntryTxParams.Amount, result.Entry.Amount)

	require.NotEmpty(t, result.Entry)
	require.Equal(t, addEntryTxParams.Category, result.Entry.Category.String)
}

func TestUpdateEntryTx(t *testing.T) {
	store := NewStore(testDB)

	user := CreateRandomUser()
	entry := createRandomEntry(t, user)

	amount := int64(10)

	result, err := store.UpdateEntryTx(context.Background(), UpdateEntryTxParams{
		Username: user.Username,
		ID:       entry.ID,
		Name:     entry.Name,
		DueDate:  entry.DueDate,
		Amount:   amount,
		Category: entry.Category.String,
	})
	require.NoError(t, err)
	require.NotEmpty(t, result)

	require.NotEmpty(t, result.Entry)
	require.Equal(t, entry.Name, result.Entry.Name)
	require.NotEqual(t, entry.Amount, result.Entry.Amount)
	require.Equal(t, entry.Category, result.Entry.Category)
}

func TestDeleteEntryTx(t *testing.T) {
	store := NewStore(testDB)

	user := CreateRandomUser()
	entry := createRandomEntry(t, user)

	result, err := store.DeleteEntryTx(context.Background(), DeleteEntryTxParams{
		Username: user.Username,
		ID:       entry.ID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

// func TestConcurrentAddEntryTx(t *testing.T) {
// 	store := NewStore(testDB)

// 	user := CreateRandomUser()

// 	date, err := GetMadeUpDate("2022-12-11")
// 	require.NoError(t, err)
// 	require.NotEmpty(t, date)

// 	// run n concurrent additions
// 	n := 10

// 	errs := make(chan error)
// 	results := make(chan AddEntryTxResult)

// 	for i := 0; i < n; i++ {
// 		go func() {
// 			result, err := store.AddEntryTx(context.Background(), AddEntryTxParams{
// 				Username: user.Username,
// 				Name:     util.RandomString(6),
// 				DueDate:  date,
// 				Amount:   int64(10),
// 			})

// 			errs <- err
// 			results <- result
// 		}()
// 	}

// 	totalExpenses := user.TotalExpenses
// 	// check results
// 	for i := 0; i < n; i++ {
// 		err := <-errs
// 		require.NoError(t, err)

// 		result := <-results
// 		require.NotEmpty(t, result)

// 		entryAmount := int64(10)
// 		totalExpenses = totalExpenses + entryAmount

// 		require.NotEmpty(t, result.User)
// 		// require.Equal(t, totalExpenses, result.User.TotalExpenses)

// 		require.NotEmpty(t, result.Entry)
// 		require.Equal(t, entryAmount, result.Entry.Amount)
// 	}

// 	// updatedUser, err := store.GetUser(context.Background(), user.Username)
// 	// require.NoError(t, err)
// 	// require.NotEmpty(t, updatedUser)
// 	// require.Equal(t, totalExpenses, updatedUser.TotalExpenses)
// }

// func TestConcurrentUpdateEntryTx(t *testing.T) {
// 	store := NewStore(testDB)

// 	user := CreateRandomUser()
// 	entry := createRandomEntry(t, user)

// 	amount := int64(0)

// 	// run n concurrent updates
// 	n := 10

// 	errs := make(chan error)
// 	results := make(chan UpdateEntryTxResult)

// 	for i := 0; i < n; i++ {
// 		amount += 10
// 		go func() {
// 			result, err := store.UpdateEntryTx(context.Background(), UpdateEntryTxParams{
// 				Username: user.Username,
// 				ID:       entry.ID,
// 				Amount:   amount,
// 			})

// 			errs <- err
// 			results <- result
// 		}()
// 	}

// 	// check results
// 	for i := 0; i < n; i++ {
// 		err := <-errs
// 		require.NoError(t, err)

// 		result := <-results
// 		require.NotEmpty(t, result)

// 		totalExpenses := user.TotalExpenses
// 		entryAmount := entry.Amount

// 		changeInAmount := amount - entryAmount
// 		totalExpenses = totalExpenses + changeInAmount

// 		require.NotEmpty(t, result.User)
// 		require.Equal(t, totalExpenses, result.User.TotalExpenses)

// 		require.NotEmpty(t, result.Entry)
// 		require.Equal(t, amount, result.Entry.Amount)
// 	}
// }
