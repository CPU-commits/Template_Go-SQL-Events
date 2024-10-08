// Code generated by SQLBoiler 4.16.2 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import "testing"

// This test suite runs each operation test in parallel.
// Example, if your database has 3 tables, the suite will run:
// table1, table2 and table3 Delete in parallel
// table1, table2 and table3 Insert in parallel, and so forth.
// It does NOT run each operation group in parallel.
// Separating the tests thusly grants avoidance of Postgres deadlocks.
func TestParent(t *testing.T) {
	t.Run("Dogs", testDogs)
}

func TestDelete(t *testing.T) {
	t.Run("Dogs", testDogsDelete)
}

func TestQueryDeleteAll(t *testing.T) {
	t.Run("Dogs", testDogsQueryDeleteAll)
}

func TestSliceDeleteAll(t *testing.T) {
	t.Run("Dogs", testDogsSliceDeleteAll)
}

func TestExists(t *testing.T) {
	t.Run("Dogs", testDogsExists)
}

func TestFind(t *testing.T) {
	t.Run("Dogs", testDogsFind)
}

func TestBind(t *testing.T) {
	t.Run("Dogs", testDogsBind)
}

func TestOne(t *testing.T) {
	t.Run("Dogs", testDogsOne)
}

func TestAll(t *testing.T) {
	t.Run("Dogs", testDogsAll)
}

func TestCount(t *testing.T) {
	t.Run("Dogs", testDogsCount)
}

func TestHooks(t *testing.T) {
	t.Run("Dogs", testDogsHooks)
}

func TestInsert(t *testing.T) {
	t.Run("Dogs", testDogsInsert)
	t.Run("Dogs", testDogsInsertWhitelist)
}

func TestReload(t *testing.T) {
	t.Run("Dogs", testDogsReload)
}

func TestReloadAll(t *testing.T) {
	t.Run("Dogs", testDogsReloadAll)
}

func TestSelect(t *testing.T) {
	t.Run("Dogs", testDogsSelect)
}

func TestUpdate(t *testing.T) {
	t.Run("Dogs", testDogsUpdate)
}

func TestSliceUpdateAll(t *testing.T) {
	t.Run("Dogs", testDogsSliceUpdateAll)
}
