/*
// ----------------------------------------------------------------------------
// recipestore.go
// Countertop Recipe Microservice

// Created by Paul Pietkiewicz on 8/5/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"gopkg.in/mgo.v2"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	pb "github.com/theorangechefco/cts/go-protos"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	recipestoreutil "github.com/theorangechefco/cts/recipestore"

	"google.golang.org/grpc/credentials"
)

var (
	port         = flag.Int("port", 50051, "The server port")
	mongoHost    = flag.String("mongo_host", "0.0.0.0:27017", "Hostname of MongoDB server")
	mongoUser    = flag.String("mongo_user", "recipestoresrv", "Username of MongoDB server")
	mongoPass    = flag.String("mongo_pass", "abc", "Password  of MongoDB server user")
	tls          = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile     = flag.String("cert_file", "keys/server1.pem", "The TLS cert file")
	keyFile      = flag.String("key_file", "keys/server1.key", "The TLS key file")
	stdErrLog    = flag.Bool("stderr_log", false, "Log to STDERR")
	fluentdHost  = flag.String("fluentd_host", "", "Fluentd agent hostname. If left blank, fluentd logging disabled")
	fluentdPort  = flag.Int("fluentd_port", 24224, "Fluentd agent port")
	mongoTimeout = 60 * time.Second
	mongoDB      = "recipes"
)

func main() {
	flag.Parse()
	var err error

	recipestoreServerInstance := new(recipestoreutil.Server)
	recipestoreServerInstance.Logger = logger.NewLogger("recipestoresrv", *fluentdHost, *fluentdPort, *stdErrLog, logrus.DebugLevel)

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{*mongoHost},
		Timeout:  mongoTimeout,
		Database: mongoDB,
		Username: *mongoUser,
		Password: *mongoPass,
	}

	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	recipestoreServerInstance.MongoSession, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		recipestoreServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "connect"},
			fmt.Sprintf("Unable to set up connection to MongoDB: %v", err))
	}
	defer recipestoreServerInstance.MongoSession.Close()
	recipestoreServerInstance.MongoSession.SetMode(mgo.Monotonic, true)

	recipestoreServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "connect"},
		fmt.Sprintf("Successfully setup connection to MongoDB on host: %s.", *mongoHost))

	lis, lisErr := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if lisErr != nil {
		recipestoreServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "bind"},
			fmt.Sprintf("Failed to listen: %v", lisErr))
	}

	var opts []grpc.ServerOption
	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			recipestoreServerInstance.Logger.Fatal(logrus.Fields{
				"phase": "startup",
				"event": "setup",
				"tags":  "certificates"},
				fmt.Sprintf("Failed to generate credentials: %v", err))
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRecipeServiceServer(grpcServer, recipestoreServerInstance)

	recipestoreServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "bind"},
		fmt.Sprintf("Countertop Recipe service listening on port: %d", *port))

	grpcServer.Serve(lis)
}
