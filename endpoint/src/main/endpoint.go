/*
// ----------------------------------------------------------------------------
// endpoint.go
// Countertop Server Endpoint Microservice

// Created by Paul Pietkiewicz on 7/15/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/Sirupsen/logrus"
	endpointutil "github.com/theorangechefco/cts/endpoint"
	pb "github.com/theorangechefco/cts/go-protos"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials"
)

var (
	port               = flag.Int("port", 50053, "Endpoint service server port")
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "keys/server1.pem", "The TLS cert file")
	keyFile  = flag.String("key_file", "keys/server1.key", "The TLS key file")

	recipeServerAddr   = flag.String("recipe_server_addr", "127.0.0.1", "The recipe server address")
	recipeServerPort   = flag.Int("recipe_server_port", 50051, "The recipe server port")
	recipeTLS      = flag.Bool("recipe_tls", false, "Connection to recipe service uses TLS if true, else plain TCP")
	recipeCertFile = flag.String("recipe_cert_file", "keys/server1.pem", "The recipe service TLS cert file")
	recipePoolSize = flag.Int("recipe_service_pool_size", 50, "The recipe service pool size")

	identityServerAddr = flag.String("identity_server_addr", "127.0.0.1", "The identity server address")
	identityServerPort   = flag.Int("identity_server_port", 50052, "The identity server port")
	identityTLS      = flag.Bool("identity_tls", false, "Connection to identity service uses TLS if true, else plain TCP")
	identityCertFile = flag.String("identity_cert_file", "keys/server1.pem", "The identity service TLS cert file")
	identityPoolSize = flag.Int("identity_service_pool_size", 50, "The identity service pool size")


	profileServerAddr  = flag.String("profile_server_addr", "127.0.0.1:50055", "The profile server address")
	profileServerPort  = flag.Int("profile_server_port", 50055, "The profile server port")
	profileTLS      = flag.Bool("profile_tls", false, "Connection to profile service uses TLS if true, else plain TCP")
	profileCertFile = flag.String("profile_cert_file", "keys/server1.pem", "The profile service TLS cert file")
	profilePoolSize = flag.Int("profile_service_pool_size", 50, "The profile service pool size")


	stdErrLog   = flag.Bool("stderr_log", true, "Log to STDERR")
	fluentdHost = flag.String("fluentd_host", "", "Fluentd agent hostname. If left blank, fluentd logging disabled")
	fluentdPort = flag.Int("fluentd_port", 24224, "Fluentd agent port")
)

func getDialOpts(tls bool, caFile string, serviceName string, loggerObj *logger.CtsLogger) []grpc.DialOption {
	var serverOpts []grpc.DialOption

	if tls {
		var sn string
		var creds credentials.TransportAuthenticator
		if caFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(caFile, sn)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to create TLS credentials for %s service. Error: %v", serviceName, err)
				loggerObj.Fatal(logrus.Fields{
					"phase": "startup",
					"event": "connection",
					"tag":   serviceName},
					errorMsg)
			}
		} else {
			creds = credentials.NewClientTLSFromCert(nil, sn)
		}
		serverOpts = append(serverOpts, grpc.WithTransportCredentials(creds))
	} else {
		serverOpts = append(serverOpts, grpc.WithInsecure())
	}
	return serverOpts
}

func main() {
	flag.Parse()
	var err error

	endpointServerInstance := new(endpointutil.Server)
	endpointServerInstance.Logger = logger.NewLogger("endpointsrv", *fluentdHost, *fluentdPort, *stdErrLog, logrus.DebugLevel)

	endpointServerInstance.RecipePool, err = NewConnPool(endpointServerInstance.Logger, "recipe", *recipeServerAddr, *recipeServerPort, *recipeTLS, *recipeCertFile, *recipePoolSize) (*HostConnPool, error) {
	if err != nil {
		endpointServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "connection",
			"tag":   "identity"},
			fmt.Sprintf("Fail to dial Recipe service: %v", err))
	}

	endpointServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "connection",
		"tag":   "recipe"},
		fmt.Sprintf("Connected to Recipe Service at: %s", *recipeServerAddr))

	identityServerOpts := getDialOpts(*identityTLS, *identityCertFile, "identity", endpointServerInstance.Logger)
	identityConn, err := grpc.Dial(*identityServerAddr, identityServerOpts...)
	if err != nil {
		endpointServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "connection",
			"tag":   "identity"},
			fmt.Sprintf("Fail to dial Identity service: %v", err))
	}
	defer identityConn.Close()
	endpointServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "connection",
		"tag":   "identity"},
		fmt.Sprintf("Connected to Identity Service at: %s", *identityServerAddr))

	profileServerOpts := getDialOpts(*profileTLS, *profileCertFile, "profile", endpointServerInstance.Logger)
	profileConn, err := grpc.Dial(*profileServerAddr, profileServerOpts...)
	if err != nil {
		endpointServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "connection",
			"tag":   "profile"},
			fmt.Sprintf("Fail to dial Profile service: %v", err))
	}
	defer profileConn.Close()
	endpointServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "connection",
		"tag":   "profile"},
		fmt.Sprintf("Connected to Profile Service at:", *profileServerAddr))

	var serverOpts []grpc.ServerOption
	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			endpointServerInstance.Logger.Fatal(logrus.Fields{
				"phase": "startup",
				"event": "setup",
				"tags":  "certificates"},
				fmt.Sprintf("Failed to generate credentials: %v", err))
		}
		serverOpts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		endpointServerInstance.Logger.Fatal(logrus.Fields{
			"phase": "startup",
			"event": "bind"},
			fmt.Sprintf("Failed to listen: %v", err))
	}
	grpcServer := grpc.NewServer(serverOpts...)
	pb.RegisterEndpointServiceServer(grpcServer, endpointServerInstance)

	endpointServerInstance.RecipeClient = pb.NewRecipeServiceClient(recipeConn)
	endpointServerInstance.IdentityClient = pb.NewIdentityServiceClient(identityConn)
	endpointServerInstance.ProfileClient = pb.NewProfileServiceClient(profileConn)

	endpointServerInstance.Logger.Info(logrus.Fields{
		"phase": "startup",
		"event": "bind"},
		fmt.Sprintf("Countertop Endpoint service listening on port: %d", *port))

	grpcServer.Serve(lis)
}
