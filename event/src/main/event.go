/*
// ----------------------------------------------------------------------------
// main.go
// Countertop Server Event Microservice

// Created by Paul Pietkiewicz on 7/15/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	eventutil "github.com/theorangechefco/cts/event"
	pb "github.com/theorangechefco/cts/go-protos"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"

	"github.com/gocql/gocql"
)

var (
	port               = flag.Int("port", 50056, "Event service server port.")
	identityServerAddr = flag.String("identity_server_addr", "127.0.0.1:50052", "The identity server address in the format of host:port")
	cassandraHost      = flag.String("cassandra_host", "0.0.0.0", "Cassandra hostname")
	cassandraUser      = flag.String("cassandra_user", "eventsrv", "Cassandra username")
	cassandraPass      = flag.String("cassandra_pass", "abc", "Cassandra password")
	stdErrLog          = flag.Bool("stderr_log", true, "Log to STDERR")
	fluentdHost        = flag.String("fluentd_host", "", "Fluentd agent hostname. If left blank, fluentd logging disabled")
	fluentdPort        = flag.Int("fluentd_port", 24224, "Fluentd agent port")
)

func main() {
	flag.Parse()

	eventServerInstance := new(eventutil.Server)
	eventServerInstance.Logger = logger.NewLogger("eventsrv", *fluentdHost, *fluentdPort, *stdErrLog, logrus.DebugLevel)

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
		eventServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "connection",
			"tag":   "cassandra"},
			fmt.Sprintf("Cannot connect to Cassandra: %v", err))
	}
	defer session.Close()
	eventServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "connection",
		"tag":   "cassandra"},
		fmt.Sprintf("Connected to Cassandra server at: %s", *cassandraHost))

	identityConn, err := grpc.Dial(*identityServerAddr, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		eventServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "connection",
			"tag":   "identity"},
			fmt.Sprintf("Fail to dial Identity service: %v", err))
	}
	defer identityConn.Close()
	eventServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "connection",
		"tag":   "identity"},
		fmt.Sprintf("Connected to Identity Service at: %s", *identityServerAddr))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		eventServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "bind"},
			fmt.Sprintf("Failed to listen: %v", err))
	}
	var serverOpts []grpc.ServerOption
	grpcServer := grpc.NewServer(serverOpts...)

	pb.RegisterEventServiceServer(grpcServer, eventServerInstance)

	eventServerInstance.Session = session
	eventServerInstance.IdentityClient = pb.NewIdentityServiceClient(identityConn)

	eventServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "bind"},
		fmt.Sprintf("Countertop Event service listening on port: %d", *port))
	grpcServer.Serve(lis)
}
