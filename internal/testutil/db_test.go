package testutil

import (
	"testing"
)

func TestSetupDB(t *testing.T) {
	db, cleanup, dsn := SetupDB()
	defer cleanup()
	if db == nil {
		t.Fatal()
	}

	if err := CleanDB(dsn); err != nil {
		t.Fatal(err)
	}
}
