/*
// ----------------------------------------------------------------------------
// sqldbutils.go
// Countertop SQL DB Utility Functions

// Created by Paul Pietkiewicz on 10/1/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package util

import (
	"fmt"
	"os"
	"strings"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func ClearUserDB(hostAddress string, port int32, username string, password string, dbName string) {
	// Find out if we have IPv6 and IPv4
	var hostTemplate string
	if strings.Contains(hostAddress, ":") {
		hostTemplate = "%s:%s@tcp([%s]:%d)/%s?charset=utf8"
	} else {
		hostTemplate = "%s:%s@tcp(%s:%d)/%s?charset=utf8"
	}

	connStr := fmt.Sprintf(hostTemplate, username, password, hostAddress, port, dbName)
	db, err := sql.Open("mysql", connStr)
	_, err = db.Exec("DELETE FROM user")
	if err != nil {
		fmt.Printf("Could not delete table. Error: %v", err)
		os.Exit(-1)
	}
}
