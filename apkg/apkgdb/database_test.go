package apkgdb

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabase_Empty_Write(t *testing.T) {
	db, err := NewDatabase()
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = db.Write(&buf)
	assert.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)

	assert.NotEmpty(t, buf.Bytes())
}
