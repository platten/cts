/*
// ----------------------------------------------------------------------------
// client.go
// Countertop Server RecipeStore Microservice Testing Client

// Created by Paul Pietkiewicz on 8/5/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"
	"io"

	"github.com/k0kubun/pp"
	pb "github.com/theorangechefco/cts/go-protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:50051", "The server address in the format of host:port")
)

// func printRecipes(client pb.RecipeServiceClient, course *pb.MealCourseRequest) {
// 	grpclog.Printf("Getting recipes for Mealcourse (%v)", course)
// 	recipes, err := client.GetRecipesForMealCourse(context.Background(), course)
// 	if err != nil {
// 		grpclog.Fatalf("%v.GetRecipesForMealCourse(_) = _, %v: ", client, err)
// 	}
// 	pp.Println(recipes)
// }

func printRecipe(client pb.RecipeServiceClient, recipeRequst *pb.RecipeRequest) {
	grpclog.Printf("Getting recipe with id (%v)", recipeRequst)
	recipe, err := client.GetRecipe(context.Background(), recipeRequst)
	if err != nil {
		grpclog.Fatalf("%v.GetRecipe(_) = _, %v: ", client, err)
	}
	pp.Println(recipe)
}

func printRecipePack(client pb.RecipeServiceClient, recipePackRequest *pb.RecipePacksRequest) {
	pp.Printf("Getting recipepacks for RecipePacksRequest (%v)", recipePackRequest)
	stream, err := client.GetRecipePacks(context.Background(), recipePackRequest)
	if err != nil {
		grpclog.Fatalf("%v.GetRecipePacks(_) = _, %v: ", client, err)
	}
	for {
		recipepack, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			grpclog.Fatalf("%v.GetRecipePacks(_) = _, %v", client, err)
		}
		pp.Println(recipepack)
	}
}

func main() {
	flag.Parse()
	// var opts []grpc.DialOption
	conn, err := grpc.Dial(*serverAddr, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewRecipeServiceClient(conn)
	// course := pb.MealCourseRequest{
	// 	pb.MealCourse_SNACK, //BREAKFAST, //SNACK,
	// }
	recipeRequest := pb.RecipeRequest{
		Recipeid: "41",
	}
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

	// printRecipes(client, &course)
	printRecipe(client, &recipeRequest)
	printRecipePack(client, &recipePackRequst)
}
