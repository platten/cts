/*
// ----------------------------------------------------------------------------
// client.go
// Countertop Server Event Microservice Testing Client

// Created by Paul Pietkiewicz on 10/1/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"
	"fmt"

	"github.com/k0kubun/pp"
	pb "github.com/theorangechefco/cts/go-protos"
	"github.com/theorangechefco/cts/go-shared-libs/cts/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
)

var (
	eventServerAddr = flag.String("server_addr", "localhost:50056", "The server address in the format of host:port")
	redisHost       = flag.String("redis_host", "0.0.0.0:6379", "Hostname of Redis server")
	redisPass       = flag.String("redis_pass", "abc", "Password of Redis server")
)

func hustleAndFlow(client pb.EventServiceClient) {
	fmt.Println("Create account")
	token, err := util.SetupRedisAccount(*redisHost, *redisPass)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Created account")
	event := pb.Event{
		Version:         0,
		Apprelease:      "testrelease",
		Appversion:      "testversion",
		Carrier:         "Tmo",
		City:            "SF",
		Country:         "USA",
		Devicemodel:     "iPizzle",
		Manufacturer:    "Appizlle",
		Model:           "sixy",
		Osversion:       "1",
		Operatingsystem: "iosy",
		Radio:           "KQED",
		Region:          "somewhere",
		Screenheight:    800,
		Screenwidth:     600,
		Wifi:            true,
		JsonPayload:     `{ "bob" : "here" }`,
	}

	md := metadata.New(map[string]string{"token": token.Id})
	ctx := metadata.NewContext(context.Background(), md)
	_, err = client.WriteEvent(ctx, &event)
	if err != nil {
		pp.Print(err)
	} else {
		fmt.Print("Successfully wrote sample data to Cassandra")
	}
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*eventServerAddr, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewEventServiceClient(conn)
	hustleAndFlow(client)
}
