package stores_test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/internal/stores"
	"github.com/stretchr/testify/require"
)

const (
	setItemQuery         = `insert into store_ (key_, value_) values ($key, $value) on conflict(key_) do update set value_ = $value`
	getItemByKeyQuery    = `select id_, key_, value_ from store_ where key_ = $key`
	listItemsQuery       = `select id_, key_, value_ from store_`
	removeItemByKeyQuery = `delete from store_ where key_ = $key`
)

func TestTenantStore(t *testing.T) {
	scenarios := map[string]func(
		t *testing.T,
		store *stores.TenantStore,
		mock sqlmock.Sqlmock,
	){
		"set item in tenant store (success)":              testSetItemInTenantStoreSuccess,
		"set item in tenant store (db error)":             testSetItemInTenantStoreDBErr,
		"get item by key from tenant store (success)":     testGetItemByKeyFromTenantStoreSuccess,
		"get item by key from tenant store (no rows err)": testGetItemByKeyFromTenantStoreNoRowsErr,
		"list items in tenant store (multiple items)":     testListItemsInTenantStoreMultipleItemsSuccess,
		"list items in tenant store (single item)":        testListItemsInTenantStoreSingleItemSuccess,
		"list items in tenant store (no items)":           testListItemsInTenantStoreNoItems,
		"list items in tenant store (db error)":           testListItemsInTenantStoreDBErr,
		"remove item by key from tenant store (success)":  testRemoveItemByKeyFromTenantStoreSuccess,
		"remove item by key from tenant store (db error)": testRemoveItemByKeyFromTenantStoreDBErr,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer db.Close()

			store := stores.NewTenantStore(db)

			fn(t, store, mock)
		})
	}
}

func testSetItemInTenantStoreSuccess(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectExec(
		regexp.QuoteMeta(setItemQuery),
	).WithArgs(
		sql.Named("key", "foo"),
		sql.Named("value", "bar"),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.SetItem(&stores.Item{
		Key:   "foo",
		Value: "bar",
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testSetItemInTenantStoreDBErr(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectExec(
		regexp.QuoteMeta(setItemQuery),
	).WithArgs(
		sql.Named("key", "foo"),
		sql.Named("value", "bar"),
	).WillReturnError(fmt.Errorf("db_err"))

	err := store.SetItem(&stores.Item{
		Key:   "foo",
		Value: "bar",
	})

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testGetItemByKeyFromTenantStoreSuccess(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectQuery(
		regexp.QuoteMeta(getItemByKeyQuery),
	).WithArgs(
		sql.Named("key", "foo"),
	).WillReturnRows(
		sqlmock.
			NewRows([]string{"id_", "key_", "value_"}).
			AddRow(1, "foo", "bar"),
	)

	item, err := store.GetItemByKey("foo")

	require.NoError(t, err)
	require.Equal(t, &stores.Item{
		ID:    1,
		Key:   "foo",
		Value: "bar",
	}, item)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testGetItemByKeyFromTenantStoreNoRowsErr(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectQuery(
		regexp.QuoteMeta(getItemByKeyQuery),
	).WithArgs(
		sql.Named("key", "foo"),
	).WillReturnRows(
		sqlmock.
			NewRows(
				[]string{"id_", "key_", "value_"},
			),
	)

	item, err := store.GetItemByKey("foo")

	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, item)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testListItemsInTenantStoreMultipleItemsSuccess(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectQuery(
		listItemsQuery,
	).WillReturnRows(
		sqlmock.
			NewRows(
				[]string{"id_", "key_", "value_"},
			).AddRows([][]driver.Value{
			{1, "foo", "bar"},
			{2, "baz", "qux"},
			{3, "ned", "dur"},
		}...),
	)

	items, err := store.ListItems()

	require.NoError(t, err)
	require.Equal(t, []stores.Item{
		{ID: 1, Key: "foo", Value: "bar"},
		{ID: 2, Key: "baz", Value: "qux"},
		{ID: 3, Key: "ned", Value: "dur"},
	}, items)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testListItemsInTenantStoreSingleItemSuccess(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectQuery(
		listItemsQuery,
	).WillReturnRows(
		sqlmock.
			NewRows([]string{"id_", "key_", "value_"}).
			AddRow(1, "foo", "bar"),
	)

	items, err := store.ListItems()

	require.NoError(t, err)
	require.Equal(t, []stores.Item{
		{ID: 1, Key: "foo", Value: "bar"},
	}, items)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testListItemsInTenantStoreNoItems(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectQuery(
		listItemsQuery,
	).WillReturnRows(sqlmock.NewRows([]string{"id_", "key_", "value_"}))

	items, err := store.ListItems()

	var expect []stores.Item

	require.NoError(t, err)
	require.Equal(t, expect, items)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testListItemsInTenantStoreDBErr(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectQuery(
		listItemsQuery,
	).WillReturnError(fmt.Errorf("db_err"))

	items, err := store.ListItems()

	require.Error(t, err)
	require.Nil(t, items)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testRemoveItemByKeyFromTenantStoreSuccess(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectExec(
		regexp.QuoteMeta(removeItemByKeyQuery),
	).WithArgs(
		sql.Named("key", "foo"),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.RemoveItemByKey("foo")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testRemoveItemByKeyFromTenantStoreDBErr(t *testing.T, store *stores.TenantStore, mock sqlmock.Sqlmock) {
	mock.ExpectExec(
		regexp.QuoteMeta(removeItemByKeyQuery),
	).WithArgs(
		sql.Named("key", "foo"),
	).WillReturnError(fmt.Errorf("db_err"))

	err := store.RemoveItemByKey("foo")

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
