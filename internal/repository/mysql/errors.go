package mysql

import (
	"errors"
	"strings"

	driver "github.com/go-sql-driver/mysql"
)

const mysqlDuplicateEntry = 1062

func duplicateKey(err error) (key string, ok bool) {
	var me *driver.MySQLError
	if !errors.As(err, &me) || me.Number != mysqlDuplicateEntry {
		return "", false
	}
	const marker = " for key "
	if i := strings.LastIndex(me.Message, marker); i >= 0 {
		return me.Message[i+len(marker):], true
	}
	return "", true
}
