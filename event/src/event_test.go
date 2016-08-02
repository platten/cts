/*
// ----------------------------------------------------------------------------
// endpoint_test.go
// Countertop Server Event Recording Microservice Testing Client

// Created by Paul Pietkiewicz on 9/29/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package event_test

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fzzy/radix/redis"
	pb "github.com/theorangechefco/cts/go-protos"
	"github.com/theorangechefco/cts/go-shared-libs/cts/util"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	serverAddr      = flag.String("server_addr", "localhost:50053", "The server address in the format of host:port")
	cassandraHost   = flag.String("cassandra_host", "127.0.0.1:6379", "Hostname of Cassandra server")
	cassandraPass   = flag.String("cassandra_pass", "", "Password of Cassandra server")
	cassandraClient *gocql.Session
)

func TestCreateAccountAndCreateToken(t *testing.T) {
	// Move this to identity service and make this higher levvel
	util.BlowAwayRedis(*redisHost, *redisPass)
	util.ClearUserDB(*dbHost, int32(*dbPort), *dbUser, *dbPass, *dbName)

	identifier := pb.Identifier{
		Deviceidentifier: "abc123",
	}

	profile := pb.Profile{
		Identifier:    &identifier,
		Firstname:     "John",
		Birthyear:     1985,
		Gender:        pb.Gender_MALE,
		Heightcm:      185,
		Weightkg:      90,
		Goalweightkg:  80,
		Activitylevel: pb.ActivityLevel_SEDENTARY,
		Mealplan:      pb.MealPlan_EIGHTEEN_HUNDRED,
		Weightgoal:    pb.WeightGoal_LOSE,
		Dietaryprofile: &pb.DietaryProfile{
			Omnivore: true,
		},
		Dietaryrestriction: &pb.DietaryRestriction{
			Soyfree: true,
		},
	}

	token, err := client.CreateProfile(context.Background(), &profile)
	if err != nil {
		t.Errorf(err.Error())
	}
	util.BlowAwayRedis(*redisHost, *redisPass)

	token, err = client.GetSessionToken(context.Background(), &identifier)
	if err != nil {
		t.Errorf(err.Error())
	}
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")
	err = checkForRedisKey(*redisClient, *redisPass, sessionTokenKey, true)
	if err != nil {
		t.Errorf("%v", err)
	}

}

func createRedisEntry(redisClient redis.Client, pass string, key string, value string) error {
	authErr := redisClient.Cmd("AUTH", pass).Err
	if authErr != nil {
		return errors.New("Invalid Redis password")
	}

	err := redisClient.Cmd("SET", key, value, "EX", 86400).Err
	if err != nil {
		errorString := fmt.Sprintf("Could not set value \"%s\" for key \"%s\". Error: %v", value, key, err)
		return errors.New(errorString)
	}
	return nil
}

func checkForRedisKey(redisClient redis.Client, pass string, key string, shouldExist bool) error {
	authErr := redisClient.Cmd("AUTH", pass).Err
	if authErr != nil {
		return errors.New("Invalid Redis password")
	}
	exists, err := redisClient.Cmd("EXISTS", key).Bool()
	if err != nil {
		errorString := fmt.Sprintf("Could not check if redis key %s exists. Error: %v", key, err)
		return errors.New(errorString)
	}
	if exists != shouldExist {
		var existsString string
		if shouldExist {
			existsString = "not exist"
		} else {
			existsString = "exist"
		}
		errorString := fmt.Sprintf("Key \"%s\" for session token does %s", key, existsString)
		return errors.New(errorString)
	}
	return nil
}

func TestMain(m *testing.M) {
	flag.Parse()
	var opts []grpc.DialOption
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		fmt.Println("fail to dial: %v", err)
		os.Exit(-1)
	}

	redisClient, err = redis.DialTimeout("tcp", *redisHost, time.Second*60)
	if err != nil {
		grpclog.Fatalf("Cannot connect to redis server")
	}

	defer conn.Close()
	client = pb.NewEndpointServiceClient(conn)
	os.Exit(m.Run())
}
