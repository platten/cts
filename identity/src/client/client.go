/*
// ----------------------------------------------------------------------------
// client.go
// Countertop Server Identity Microservice Testing Client

// Created by Paul Pietkiewicz on 8/27/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"

	"google.golang.org/grpc/grpclog"
	//	"github.com/k0kubun/pp"
	"github.com/k0kubun/pp"
	pb "github.com/theorangechefco/cts/go-protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:50051", "The server address in the format of host:port")
)

func lookupToken(client pb.IdentityServiceClient, token *pb.SessionToken) (*pb.UserId, error) {
	grpclog.Printf("Getting userid for session token with id (%v)", token)
	userid, err := client.LookupSessionToken(context.Background(), token)
	pp.Println(userid)
	pp.Println(err)
	if err != nil {
		grpclog.Printf("%v.LookupSessionToken(_) = _, %v: ", client, err)
		return nil, err
	}
	return userid, nil
}

func createToken(client pb.IdentityServiceClient, userId *pb.UserId) (*pb.SessionToken, error) {
	grpclog.Printf("Getting session token for user with id (%v)", userId)
	token, err := client.GenerateSessionToken(context.Background(), userId)
	if err != nil {
		grpclog.Printf("%v.RecipeServiceClient(_) = _, %v: ", client, err)
		return nil, err
	}
	return token, nil
}

func closeSession(client pb.IdentityServiceClient, token *pb.SessionToken) (*pb.Response, error) {
	grpclog.Printf("Closing session for token (%v)", token)
	response, err := client.CloseSession(context.Background(), token)
	if err != nil {
		grpclog.Fatalf("%v.CloseSession(_) = _, %v: ", client, err)
		return nil, err
	}
	return response, nil
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*serverAddr, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		grpclog.Fatal("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewIdentityServiceClient(conn)
	// userId := pb.UserId{
	// 	"abc123",
	// }
	pp.Println(client)
	token := &pb.SessionToken{Id: "abc"}

	grpclog.Print("Check for non-existing token.")
	lookupToken(client, token)

	// grpclog.Print("Create a token")
	// token, _ = createToken(client, &userId)
	//
	// grpclog.Print("Looking up token for existing session")
	// _, _ = lookupToken(client, token)
	//
	// grpclog.Print("Closing session")
	// closeSession(client, token)
	//
	// grpclog.Print("Looking up token which should be deleted.")
	// lookupToken(client, token)
}
