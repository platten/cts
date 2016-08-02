/*
// ----------------------------------------------------------------------------
// client.go
// Countertop Server Profile Microservice Testing Client

// Created by Paul Pietkiewicz on 9/4/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
	pb "github.com/theorangechefco/cts/go-protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:50051", "The server address in the format of host:port")
	dbHost     = flag.String("db_host", ":::::::", "Hostname of MySQL server")
	dbPort     = flag.Int("db_port", 3306, "Port of MySQL server")
	dbUser     = flag.String("db_user", "admin", "Username of MySQL server")
	dbPass     = flag.String("db_pass", "admin", "Password of MySQL server user")
	dbName     = flag.String("db_name", "profile", "Database name")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*serverAddr, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		grpclog.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewProfileServiceClient(conn)

	profile := pb.Profile{
		Identifier: &pb.Identifier{
			Deviceidentifier: "abc123",
		},
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

	connStr := fmt.Sprintf("%s:%s@tcp([%s]:%d)/%s?charset=utf8", *dbUser, *dbPass, *dbHost, *dbPort, *dbName)
	db, err := gorm.Open("mysql", connStr)
	if err != nil {
		grpclog.Fatalf("Cannot establish connection with database %s. Error: %v", connStr, err)
	} else {
		grpclog.Printf("Connected to: ([%s]:%d)", *dbHost, *dbPort)
	}

	err = db.DB().Ping()
	if err != nil {
		grpclog.Fatalf("Cannot ping database: %v", err)
	}
	db.Exec("DELETE FROM profile.user;")

	userId, error := client.CreateProfile(context.Background(), &profile)
	pp.Print(userId)
	pp.Print(error)

	lookedUpID, lookupError := client.GetUUID(context.Background(), &pb.Identifier{Deviceidentifier: "abc123"})
	pp.Print(lookedUpID)
	pp.Print(lookupError)

	if lookedUpID != nil && userId.Uuid == lookedUpID.Uuid {
		fmt.Print("UUIDs match")
	} else {
		fmt.Print("ERROR: UUIDs don't match")
		pp.Print(userId)
		pp.Print(lookedUpID)
	}

	notFoundID, notFoundError := client.GetUUID(context.Background(), &pb.Identifier{Deviceidentifier: "abc"})
	pp.Print(notFoundID)
	pp.Print(notFoundError)

	badUserId, notFoundError := client.CreateProfile(context.Background(), &profile)
	pp.Print(badUserId)
	pp.Print(notFoundError)

	profileResponse, profileError := client.GetProfileInfoByUUID(context.Background(), userId)
	pp.Print(profileResponse)
	pp.Print(profileError)

	altUserId := pb.UserId{Uuid: "abc"}
	profileResponse, profileError = client.GetProfileInfoByUUID(context.Background(), &altUserId)
	pp.Print(profileResponse)
	pp.Print(profileError)

	profile = pb.Profile{
		Firstname:    "John",
		Birthyear:    1988,
		Gender:       pb.Gender_MALE,
		Heightcm:     185,
		Weightkg:     85,
		Goalweightkg: 75,
		Dietaryprofile: &pb.DietaryProfile{
			Omnivore: true,
		},
		Dietaryrestriction: &pb.DietaryRestriction{
			Soyfree: false,
		},
	}

	profileUpdateRequest := pb.ProfileUpdateRequest{
		Id:      userId,
		Profile: &profile,
	}

	response, updateError := client.SetProfileInfo(context.Background(), &profileUpdateRequest)
	pp.Print(response)
	pp.Print(updateError)

	profileResponse, profileError = client.GetProfileInfoByUUID(context.Background(), userId)
	pp.Print(profileResponse)
	pp.Print(profileError)

	profileUpdateRequest = pb.ProfileUpdateRequest{
		Id:      &altUserId,
		Profile: &profile,
	}

	response, updateError = client.SetProfileInfo(context.Background(), &profileUpdateRequest)
	pp.Print(response)
	pp.Print(updateError)
}
