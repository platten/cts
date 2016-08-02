/*
// ----------------------------------------------------------------------------
// endpoint_test.go
// Countertop Server Endpoint Microservice Testing Client

// Created by Paul Pietkiewicz on 9/15/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package endpoint_test

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fzzy/radix/redis"
	"github.com/golang/protobuf/proto"
	pb "github.com/theorangechefco/cts/go-protos"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	"github.com/theorangechefco/cts/go-shared-libs/cts/util"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
)

var (
	serverAddr = flag.String("server_addr", "localhost:50053", "The server address in the format of host:port")
	redisHost  = flag.String("redis_host", "0.0.0.0:6379", "Hostname of Redis server")
	redisPass  = flag.String("redis_pass", "abc", "Password of Redis server")
	dbHost     = flag.String("db_host", ":::::::", "Hostname of MySQL server")
	dbPort     = flag.Int("db_port", 3306, "Port of MySQL server")
	dbUser     = flag.String("db_user", "admin", "Username of MySQL server")
	dbPass     = flag.String("db_pass", "admin", "Password of MySQL server user")
	dbName     = flag.String("db_name", "profile", "Database name")

	redisClient *redis.Client
	client      pb.EndpointServiceClient
)

func TestCreateAccountAndCreateToken(t *testing.T) {
	// Move this to identity service and make this higher levvel
	logObj := logger.NewLogger("test", "", 8080, true, logrus.DebugLevel)
	redisClient, _ := util.NewClient(logObj, *redisHost, *redisPass)
	defer redisClient.Close()

	util.BlowAwayRedis(redisClient)
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
		t.Errorf("%v", err)
	}
	util.BlowAwayRedis(redisClient)

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

func TestCloseSession(t *testing.T) {
	util.BlowAwayRedis(redisClient)

	logObj := logger.NewLogger("test", "", 8080, true, logrus.DebugLevel)
	redisClient, _ := util.NewClient(logObj, *redisHost, *redisPass)
	defer redisClient.Close()

	util.SetupUser(*dbHost, *dbPort, *dbUser, *dbPass, *dbName)
	token, _ := util.SetupRedisAccount(*redisHost, *redisPass)

	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")
	err := checkForRedisKey(*redisClient, *redisPass, sessionTokenKey, true)
	if err != nil {
		t.Errorf("%v", err)
	}
	md := metadata.New(map[string]string{"token": token.Id})
	ctx := metadata.NewContext(context.Background(), md)

	response, closeErr := client.CloseSession(ctx, &pb.EmptyRequest{})

	if closeErr != nil {
		t.Errorf("%v.CloseSession(_) = _, %v: ", client, err)
	}
	if response != nil && !response.Success {
		t.Errorf("response.Success should be true.")
	}
	err = checkForRedisKey(*redisClient, *redisPass, sessionTokenKey, false)
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestCreateProfileAndGetProfile(t *testing.T) {
	util.ClearUserDB(*dbHost, int32(*dbPort), *dbUser, *dbPass, *dbName)
	logObj := logger.NewLogger("test", "", 8080, true, logrus.DebugLevel)
	redisClient, _ := util.NewClient(logObj, *redisHost, *redisPass)
	defer redisClient.Close()
	util.BlowAwayRedis(redisClient)

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

	token, err := client.CreateProfile(context.Background(), &profile)
	if err != nil {
		t.Errorf(err.Error())
	}
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")
	err = checkForRedisKey(*redisClient, *redisPass, sessionTokenKey, true)
	if err != nil {
		t.Errorf(err.Error())
	}

	md := metadata.New(map[string]string{"token": token.Id})
	ctx := metadata.NewContext(context.Background(), md)

	returnedProfile, getProfileError := client.GetProfileInfo(ctx, &pb.EmptyRequest{})
	if getProfileError != nil {
		t.Errorf(getProfileError.Error())
	}
	if !proto.Equal(&profile, returnedProfile) {
		t.Errorf("Profiles are not equal")
	}
}

func TestCreateAndUpdateProfile(t *testing.T) {
	util.ClearUserDB(*dbHost, int32(*dbPort), *dbUser, *dbPass, *dbName)
	logObj := logger.NewLogger("test", "", 8080, true, logrus.DebugLevel)
	redisClient, _ := util.NewClient(logObj, *redisHost, *redisPass)
	defer redisClient.Close()
	util.BlowAwayRedis(redisClient)
	_, err := util.SetupUser(*dbHost, *dbPort, *dbUser, *dbPass, *dbName)
	if err != nil {
		t.Errorf(err.Error())
	}
	token, redisErr := util.SetupRedisAccount(*redisHost, *redisPass)
	if redisErr != nil {
		t.Errorf(redisErr.Error())
	}
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")
	err = checkForRedisKey(*redisClient, *redisPass, sessionTokenKey, true)
	if err != nil {
		t.Errorf(err.Error())
	}

	updatedProfile := pb.Profile{
		Identifier: &pb.Identifier{
			Deviceidentifier: "ecb7381b-566c-4e53-b3ea-add8cd372d6a",
		},
		Firstname:     "Bob",
		Birthyear:     1985,
		Gender:        pb.Gender_MALE,
		Heightcm:      188,
		Weightkg:      90,
		Goalweightkg:  90,
		Activitylevel: pb.ActivityLevel_SEDENTARY,
		Mealplan:      pb.MealPlan_EIGHTEEN_HUNDRED,
		Weightgoal:    pb.WeightGoal_LOSE,
		Dietaryprofile: &pb.DietaryProfile{
			Omnivore:   false,
			Vegetarian: true,
		},
		Dietaryrestriction: &pb.DietaryRestriction{
			Soyfree:   false,
			Lowsodium: true,
		},
	}

	md := metadata.New(map[string]string{"token": token.Id})
	ctx := metadata.NewContext(context.Background(), md)

	response, setProfileError := client.SetProfileInfo(ctx, &updatedProfile)
	if setProfileError != nil {
		t.Errorf(setProfileError.Error())
	}
	if response != nil && response.Success != true {
		t.Errorf("Profile did not update successfully")
	}

	returnedProfile, getProfileError := client.GetProfileInfo(ctx, &pb.EmptyRequest{})
	if getProfileError != nil {
		t.Errorf(getProfileError.Error())
	}

	if !proto.Equal(&updatedProfile, returnedProfile) {
		t.Errorf("Updated profile did not save correctly.")
	}
}

func TestGetRecipe(t *testing.T) {
	util.ClearUserDB(*dbHost, int32(*dbPort), *dbUser, *dbPass, *dbName)
	logObj := logger.NewLogger("test", "", 8080, true, logrus.DebugLevel)
	redisClient, _ := util.NewClient(logObj, *redisHost, *redisPass)
	defer redisClient.Close()
	util.BlowAwayRedis(redisClient)
	_, err := util.SetupUser(*dbHost, *dbPort, *dbUser, *dbPass, *dbName)
	if err != nil {
		t.Errorf(err.Error())
	}
	token, redisErr := util.SetupRedisAccount(*redisHost, *redisPass)
	if redisErr != nil {
		t.Errorf(redisErr.Error())
	}
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")
	err = checkForRedisKey(*redisClient, *redisPass, sessionTokenKey, true)
	if err != nil {
		t.Errorf(err.Error())
	}

	md := metadata.New(map[string]string{"token": token.Id})
	ctx := metadata.NewContext(context.Background(), md)

	recipe, recipeErr := client.GetRecipe(ctx, &pb.RecipeRequest{Recipeid: "142"})
	if recipeErr != nil {
		if grpc.Code(recipeErr) != codes.NotFound {
			t.Errorf("GetRecipe(_) = _, %v", recipeErr)
		} else {
			t.Errorf("Recipe with ID 142 not found.")
		}
	}
	if recipe == nil {
		t.Errorf("Recipe not fetched.")
	}
}

func TestGetMissingRecipe(t *testing.T) {
	util.ClearUserDB(*dbHost, int32(*dbPort), *dbUser, *dbPass, *dbName)
	logObj := logger.NewLogger("test", "", 8080, true, logrus.DebugLevel)
	redisClient, _ := util.NewClient(logObj, *redisHost, *redisPass)
	defer redisClient.Close()
	util.BlowAwayRedis(redisClient)
	_, err := util.SetupUser(*dbHost, *dbPort, *dbUser, *dbPass, *dbName)
	if err != nil {
		t.Errorf(err.Error())
	}
	token, redisErr := util.SetupRedisAccount(*redisHost, *redisPass)
	if redisErr != nil {
		t.Errorf(redisErr.Error())
	}
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")
	err = checkForRedisKey(*redisClient, *redisPass, sessionTokenKey, true)
	if err != nil {
		t.Errorf(err.Error())
	}

	md := metadata.New(map[string]string{"token": token.Id})
	ctx := metadata.NewContext(context.Background(), md)

	recipe, recipeErr := client.GetRecipe(ctx, &pb.RecipeRequest{Recipeid: "141"})
	if recipeErr == nil || grpc.Code(recipeErr) != codes.NotFound {
		t.Errorf("GetRecipe(_) = _, %v", recipeErr)
	}
	if recipe != nil {
		t.Errorf("Should not have returned recipe.")
	}
}

func TestGetMessagePack(t *testing.T) {
	util.ClearUserDB(*dbHost, int32(*dbPort), *dbUser, *dbPass, *dbName)
	logObj := logger.NewLogger("test", "", 8080, true, logrus.DebugLevel)
	redisClient, _ := util.NewClient(logObj, *redisHost, *redisPass)
	defer redisClient.Close()
	util.BlowAwayRedis(redisClient)
	_, err := util.SetupUser(*dbHost, *dbPort, *dbUser, *dbPass, *dbName)
	if err != nil {
		t.Errorf(err.Error())
	}
	token, redisErr := util.SetupRedisAccount(*redisHost, *redisPass)
	if redisErr != nil {
		t.Errorf(redisErr.Error())
	}
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")
	err = checkForRedisKey(*redisClient, *redisPass, sessionTokenKey, true)
	if err != nil {
		t.Errorf(err.Error())
	}

	counter := 0
	recipePackRequst := pb.RecipePacksRequest{
		Dietaryprofile: &pb.DietaryProfile{
			Omnivore:   true,
			Vegetarian: false,
			Vegan:      false,
			Raw:        false},
		Dietaryrestriction: &pb.DietaryRestriction{
			Glutenfree: false,
			Nutfree:    false,
			Dairyfree:  false,
			Soyfree:    false,
			Lowsodium:  false},
		Mealplan: pb.MealPlan_EIGHTEEN_HUNDRED}

	md := metadata.New(map[string]string{"token": token.Id})
	ctx := metadata.NewContext(context.Background(), md)

	stream, recipeErr := client.GetRecipePacks(ctx, &recipePackRequst)
	if recipeErr != nil {
		t.Errorf("GetRecipePacks(_) = _, %v: ", recipeErr)
		t.FailNow()

	}

	for {
		_, streamErr := stream.Recv()
		if streamErr == io.EOF {
			break
		}
		if streamErr != nil {
			t.Errorf("GetRecipePacks(_) = _, %v", err)
		}
		counter++
	}
	if counter == 0 {
		t.Errorf("No recipes returned")
	}
}

func TestGetMissingMessagePack(t *testing.T) {
	util.ClearUserDB(*dbHost, int32(*dbPort), *dbUser, *dbPass, *dbName)
	logObj := logger.NewLogger("test", "", 8080, true, logrus.DebugLevel)
	redisClient, _ := util.NewClient(logObj, *redisHost, *redisPass)
	defer redisClient.Close()
	util.BlowAwayRedis(redisClient)
	_, err := util.SetupUser(*dbHost, *dbPort, *dbUser, *dbPass, *dbName)
	if err != nil {
		t.Errorf(err.Error())
	}
	token, redisErr := util.SetupRedisAccount(*redisHost, *redisPass)
	if redisErr != nil {
		t.Errorf(redisErr.Error())
	}
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")
	err = checkForRedisKey(*redisClient, *redisPass, sessionTokenKey, true)
	if err != nil {
		t.Errorf(err.Error())
	}

	counter := 0
	recipePackRequst := pb.RecipePacksRequest{
		Dietaryprofile: &pb.DietaryProfile{
			Omnivore:   false,
			Vegetarian: false,
			Vegan:      true,
			Raw:        true},
		Dietaryrestriction: &pb.DietaryRestriction{
			Glutenfree: true,
			Nutfree:    true,
			Dairyfree:  true,
			Soyfree:    true,
			Lowsodium:  false},
		Mealplan: pb.MealPlan_EIGHTEEN_HUNDRED}

	md := metadata.New(map[string]string{"token": token.Id})
	ctx := metadata.NewContext(context.Background(), md)

	stream, recipeErr := client.GetRecipePacks(ctx, &recipePackRequst)
	if recipeErr != nil {
		t.Errorf("GetRecipePacks(_) = _, %v: ", recipeErr)
		t.FailNow()

	}

	for {
		_, streamErr := stream.Recv()
		if streamErr == io.EOF {
			break
		}
		if streamErr != nil {
			t.Errorf("GetRecipePacks(_) = _, %v", err)
		}
		counter++
	}
	if counter != 0 {
		t.Errorf("Should not have returned recipes")
	}
}

//
// Testing util functions
//

func createRedisEntry(redisClient redis.Client, pass string, key string, value string) error {
	if pass != "" {
		authErr := redisClient.Cmd("AUTH", pass).Err
		if authErr != nil {
			return errors.New("Invalid Redis password")
		}
	}

	err := redisClient.Cmd("SET", key, value, "EX", 86400).Err
	if err != nil {
		errorString := fmt.Sprintf("Could not set value \"%s\" for key \"%s\". Error: %v", value, key, err)
		return errors.New(errorString)
	}
	return nil
}

func checkForRedisKey(redisClient redis.Client, pass string, key string, shouldExist bool) error {
	if pass != "" {
		authErr := redisClient.Cmd("AUTH", pass).Err
		if authErr != nil {
			return errors.New("Invalid Redis password")
		}
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
	conn, err := grpc.Dial(*serverAddr, []grpc.DialOption{grpc.WithInsecure()}...)
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
