package stores_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/internal/stores"
)

func TestSystemStore(t *testing.T) {
	scenarios := map[string]func(
		t *testing.T,
		store *stores.SystemStore,
		mock sqlmock.Sqlmock,
	){
		"get user from system store (success)":     testGetUserFromSystemStoreSuccess,
		"get user from system store (no rows err)": testGetUserFromSystemStoreNoRows,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock database: %s", err)
			}
			defer db.Close()

			store := stores.NewSystemStore(db)

			fn(t, store, mock)
		})
	}
}

func testGetUserFromSystemStoreSuccess(t *testing.T, store *stores.SystemStore, mock sqlmock.Sqlmock) {

}

func testGetUserFromSystemStoreNoRows(t *testing.T, store *stores.SystemStore, mock sqlmock.Sqlmock) {

}
