/*
// ----------------------------------------------------------------------------
// profile.go
// Countertop Profile Microservice

// Created by Paul Pietkiewicz on 10/21/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"
	"fmt"
	"net"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials"

	"github.com/Sirupsen/logrus"
	pb "github.com/theorangechefco/cts/go-protos"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	profileutil "github.com/theorangechefco/cts/profile"
)

var (
	port          = flag.Int("port", 50051, "The server port")
	dbHost        = flag.String("db_host", ":::::::", "Hostname of MySQL server")
	dbPort        = flag.Int("db_port", 3306, "Port of MySQL server")
	dbUser        = flag.String("db_user", "admin", "Username of MySQL server")
	dbPass        = flag.String("db_pass", "admin", "Password of MySQL server user")
	dbName        = flag.String("db_name", "profile", "Database name")
	dbMaxIdleConn = flag.Int("db_max_idle_conn", 25, "Maximum idle connections")
	dbMaxOpenConn = flag.Int("db_max_open_conn", 150, "Maximum open connections")
	dbLog         = flag.Bool("db_log", false, "Log SQL queries")
	tls           = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile      = flag.String("cert_file", "keys/server1.pem", "The TLS cert file")
	keyFile       = flag.String("key_file", "keys/server1.key", "The TLS key file")
	stdErrLog     = flag.Bool("stderr_log", true, "Log to STDERR")
	fluentdHost   = flag.String("fluentd_host", "", "Fluentd agent hostname. If left blank, fluentd logging disabled")
	fluentdPort   = flag.Int("fluentd_port", 24224, "Fluentd agent port")
)

func main() {
	flag.Parse()
	var err error

	profileServerInstance := new(profileutil.Server)
	profileServerInstance.Logger = logger.NewLogger("profilesrv", *fluentdHost, *fluentdPort, *stdErrLog, logrus.DebugLevel)
	connStr := fmt.Sprintf("%s:%s@tcp([%s]:%d)/%s?charset=utf8&parseTime=true", *dbUser, *dbPass, *dbHost, *dbPort, *dbName)

	profileServerInstance.DB, err = gorm.Open("mysql", connStr)
	if err != nil {
		profileServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "connect"},
			fmt.Sprintf("Unable to set up connection to with database: %s Error: %v", connStr, err))
	}
	defer profileServerInstance.DB.Close()

	infoMsg := fmt.Sprintf("Connected to: ([%s]:%d)", *dbHost, *dbPort)
	profileServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "connect"},
		infoMsg)

	profileServerInstance.DB.DB().SetMaxIdleConns(*dbMaxIdleConn)
	profileServerInstance.DB.DB().SetMaxOpenConns(*dbMaxOpenConn)
	profileServerInstance.DB.SingularTable(true)
	if *dbLog {
		profileServerInstance.DB.LogMode(true)
	}

	if err := profileServerInstance.DB.DB().Ping(); err != nil {
		profileServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "connect"},
			fmt.Sprintf("Unable to pinh database %s. Error: %v", connStr, err))
	}

	lis, lisErr := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		profileServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "bind"},
			fmt.Sprintf("Failed to listen: %v", lisErr))
	}

	var opts []grpc.ServerOption
	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			profileServerInstance.Logger.Fatal(logrus.Fields{
				"phase": "startup",
				"event": "setup",
				"tags":  "certificates"},
				fmt.Sprintf("Failed to generate credentials: %v", err))
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterProfileServiceServer(grpcServer, profileServerInstance)

	profileServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "bind"},
		fmt.Sprintf("Countertop Profile service listening on port: %d", *port))

	grpcServer.Serve(lis)
}
