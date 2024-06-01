package database

// func TestDatabase(t *testing.T) {
// 	scenarios := map[string]func(t *testing.T, db *sql.DB, mock sqlmock.Sqlmock){
// 		"test create tables (success)": testDatabaseCreateTablesSuccess,
// 	}
//
// 	for scenario, fn := range scenarios {
// 		t.Run(scenario, func(t *testing.T) {
// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatal("unable to create mock database")
// 			}
//
// 			fn(t, db, mock)
// 		})
// 	}
// }

// func testDatabaseCreateTablesSuccess(t *testing.T, db *sql.DB, mock sqlmock.Sqlmock) {
// 	query := `
// 		create table if not exists variables_ (
// 			id_ integer primary key autoincrement not null,
// 			key_ text not null,
// 			value_ text not null,
// 			secret_ boolean,
// 			project_name_ text,
// 			environment_name_ text,
// 			unique (key_, project_name_, environment_name_)
// 		)
// 	`
//
// 	mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 0))
//
// 	err := CreateTables(db)
//
// 	require.NoError(t, err, "should not return error")
// 	require.NoError(t, mock.ExpectationsWereMet(), "should query database as expected")
// }
