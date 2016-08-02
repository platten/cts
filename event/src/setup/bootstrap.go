/*
// ----------------------------------------------------------------------------
// bootstrap.go
// Countertop Profile Microservice Cassandra DB Bootstraping Tool

// Created by Paul Pietkiewicz on 9/2/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
)

var (
	cassandraHost = flag.String("cassandra_host", "127.0.0.1", "Cassandra hostname")
	cassandraUser = flag.String("cassandra_user", "cassandra", "Cassandra username")
	cassandraPass = flag.String("cassandra_pass", "abc", "Cassandra password")
)

func main() {
	flag.Parse()

	logger := log.New(os.Stderr, "", log.LstdFlags)

	cluster := gocql.NewCluster(*cassandraHost)
	if *cassandraUser != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: *cassandraUser,
			Password: *cassandraPass,
		}
	}
	cluster.Keyspace = "eventks"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Fatal(err)
	}
	defer session.Close()

	// if err := session.Query("INSERT INTO users (lastname, age, city, email, firstname) VALUES ('Jones', 35, 'Austin', 'bob@example.com', 'Bob')").Exec(); err != nil {
	session.Query("DROP TABLE events;").Exec()

	createString := `CREATE TABLE events (
                      id timeuuid PRIMARY KEY,
                      version int,
                      apprelease ascii,
                      appversion ascii,
                      carrier varchar,
                      city varchar,
                      country varchar,
                      devicemodel ascii,
                      manufacturer ascii,
                      model ascii,
                      osversion ascii,
                      operatingsystem ascii,
                      radio ascii,
                      region ascii,
                      screenheight int,
                      screenwidth int,
                      wifi boolean,
                      createdat timestamp,
                      payload text
                    );`
	if err := session.Query(createString).Exec(); err != nil {
		logger.Fatal(err)
	}
	fmt.Println(err)

	logger.Println("Cassandra setup complete!")
}
