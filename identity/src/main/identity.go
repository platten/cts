/*
// ----------------------------------------------------------------------------
// identity.go
// Countertop Identity Microservice

// Created by Paul Pietkiewicz on 10/12/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

// TODO(ppietkiewicz): Add sentinel based connection pooling & load balancing

package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	"github.com/theorangechefco/cts/go-shared-libs/cts/util"
	identityutil "github.com/theorangechefco/cts/identity"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/theorangechefco/cts/go-protos"
)

var (
	port          = flag.Int("port", 50051, "The server port")
	redisHost     = flag.String("redis_host", "127.0.0.1:6379", "Hostname of Redis server")
	redisPass     = flag.String("redis_pass", "abc", "Redis password (optional)")
	redisPoolSize = flag.Int("redis_pool_size", 150, "Redis pool size")
	redisTTL      = flag.Int("redis_ttl", 86400, "Redis TTL in seconds")
	tls           = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile      = flag.String("cert_file", "keys/server1.pem", "The TLS cert file")
	keyFile       = flag.String("key_file", "keys/server1.key", "The TLS key file")
	stdErrLog     = flag.Bool("stderr_log", false, "Log to STDERR")
	fluentdHost   = flag.String("fluentd_host", "", "Fluentd agent hostname. If left blank, fluentd logging disabled")
	fluentdPort   = flag.Int("fluentd_port", 24224, "Fluentd agent port")
)

func main() {
	flag.Parse()

	identityServerInstance := new(identityutil.Server)
	identityServerInstance.Logger = logger.NewLogger("identitysrv", *fluentdHost, *fluentdPort, *stdErrLog, logrus.DebugLevel)
	identityServerInstance.TTL = *redisTTL
	pool, err := util.NewPool(identityServerInstance.Logger, *redisHost, *redisPass, *redisPoolSize, 3, 500)
	if err != nil {
		identityServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "connect"},
			fmt.Sprintf("Unable to set up Redis pool: %v", err))
	}
	identityServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "connect"},
		fmt.Sprintf("Successfully setup pool with %d connections.", *redisPoolSize))
	identityServerInstance.Pool = pool

	lis, lisErr := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if lisErr != nil {
		identityServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "bind"},
			fmt.Sprintf("Failed to listen: %v", lisErr))
	}

	var opts []grpc.ServerOption
	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			identityServerInstance.Logger.Fatal(logrus.Fields{
				"phase": "startup",
				"event": "setup",
				"tags":  "certificates"},
				fmt.Sprintf("Failed to generate credentials: %v", err))
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterIdentityServiceServer(grpcServer, identityServerInstance)

	identityServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "bind"},
		fmt.Sprintf("Countertop Identity service listening on port: %d", *port))

	grpcServer.Serve(lis)
}
