/*
// ----------------------------------------------------------------------------
// main.go
// Countertop Recipe Microservice

// Created by Paul Pietkiewicz on 8/5/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package recipestore

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/theorangechefco/cts/go-protos"

	"github.com/Sirupsen/logrus"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	"google.golang.org/grpc/codes"
)

type Server struct {
	Logger       *logger.CtsLogger
	MongoSession *mgo.Session
}

func (r *Server) GetRecipe(ctx context.Context, recipeRequest *pb.RecipeRequest) (*pb.Recipe, error) {
	recipe := new(pb.Recipe)
	session := r.MongoSession.Copy()
	defer session.Close()

	if err := session.Ping(); err != nil {
		r.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "ping",
			"tag":   "mongodb",
			"rpc":   "GetRecipe"},
			fmt.Sprintf("Cannot ping MongoDB server while fetching Recipe %s Error: %v", recipeRequest.Recipeid, err))
		return nil, grpc.Errorf(codes.Unknown, "Cannot ping MongDB server when fetching Recipe %s. Error: %v", recipeRequest.Recipeid, err)
	}

	c := session.DB("recipes").C("recipes")

	err := c.Find(bson.M{"id": recipeRequest.Recipeid}).One(&recipe)
	if err != nil {
		if err == mgo.ErrNotFound {
			r.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "mongodb",
				"rpc":   "GetRecipe"},
				fmt.Sprintf("Recipe %s not found", recipeRequest.Recipeid))
			return nil, grpc.Errorf(codes.NotFound, "Recipe not found.")
		}
		errMsg := fmt.Sprintf("Unknown recipe lookup error: %v", err)
		r.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "mongodb",
			"rpc":   "GetRecipe"},
			errMsg)
		return nil, grpc.Errorf(codes.Unknown, errMsg)
	}

	infoMsg := fmt.Sprintf("Returning recipe %s with ID: %s", recipe.Name, recipe.Id)
	r.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "fetch",
		"tag":   "mongodb",
		"rpc":   "GetRecipe"},
		infoMsg)

	return recipe, nil
}

// TODO(ppietkiewicz): Convert following RPC method to stream
// func (r *Server) GetRecipesForMealCourse(ctx context.Context, course *pb.MealCourseRequest) (*pb.Recipes, error) {
// 	recipes := new(pb.Recipes)
// 	c := r.mongoSession.DB("recipes").C("recipes")
//
// 	err := c.Find(bson.M{"$and": []bson.M{
// 		bson.M{"mealcourses.breakfast": course.Breakfast},
// 		bson.M{"mealcourses.lunchanddinner": course.Lunchanddinner},
// 		bson.M{"mealcourses.snacks": course.Snacks},
// 		bson.M{"mealcourses.dessert": course.Dessert}}}).All(&recipes.Recipe)
// 	if err != nil {
// 		if err == mgo.ErrNotFound {
// 			grpclog.Println(err)
// 			return nil, grpc.Errorf(codes.NotFound, "Recipe not found.")
// 		}
// 		errMsg := fmt.Sprintf("Unknown recipe lookup error: %v", err)
// 		grpclog.Println(errMsg)
// 		return nil, grpc.Errorf(codes.Unknown, errMsg)
// 	}
// 	grpclog.Printf("INFO: Returning", len(recipes.Recipe), "recipes for: %v", course.Mealcourse)
// 	return recipes, nil
// }

func (r *Server) GetRecipePacks(recipePackRequest *pb.RecipePacksRequest, stream pb.RecipeService_GetRecipePacksServer) error {
	recipePacks := new([]pb.RecipePack)
	session := r.MongoSession.Copy()
	defer session.Close()

	if err := session.Ping(); err != nil {
		r.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "ping",
			"tag":   "mongodb",
			"rpc":   "GetRecipePacks"},
			fmt.Sprintf("Cannot ping MongoDB server while fetching Recipepack %v Error: %v", recipePackRequest, err))
		return grpc.Errorf(codes.Unknown, "Cannot ping MongDB server when fetching Recipepacks %v. Error: %v", recipePackRequest, err)
	}

	c := session.DB("recipes").C("recipepacks")
	// TODO(ppietkiewicz): add dietary restrictions to responses!
	var err error
	switch {
	case recipePackRequest.Dietaryprofile.Omnivore && recipePackRequest.Dietaryprofile.Raw:
		err = c.Find(bson.M{"$and": []bson.M{
			bson.M{"dietaryprofile.omnivore": recipePackRequest.Dietaryprofile.Omnivore},
			bson.M{"dietaryprofile.raw": recipePackRequest.Dietaryprofile.Raw},
			bson.M{"mealplan": recipePackRequest.Mealplan}}}).All(recipePacks)
	case recipePackRequest.Dietaryprofile.Omnivore == true:
		err = c.Find(bson.M{"mealplan": recipePackRequest.Mealplan}).All(recipePacks)
	case recipePackRequest.Dietaryprofile.Vegetarian && recipePackRequest.Dietaryprofile.Raw:
		err = c.Find(bson.M{"$and": []bson.M{
			bson.M{"dietaryprofile.vegetarian": recipePackRequest.Dietaryprofile.Vegetarian},
			bson.M{"dietaryprofile.raw": recipePackRequest.Dietaryprofile.Raw},
			bson.M{"mealplan": recipePackRequest.Mealplan}}}).All(recipePacks)
	case recipePackRequest.Dietaryprofile.Vegetarian:
		err = c.Find(bson.M{"$and": []bson.M{
			bson.M{"dietaryprofile.vegetarian": recipePackRequest.Dietaryprofile.Vegetarian},
			bson.M{"mealplan": recipePackRequest.Mealplan}}}).All(recipePacks)
	case recipePackRequest.Dietaryprofile.Vegan && recipePackRequest.Dietaryprofile.Raw:
		err = c.Find(bson.M{"$and": []bson.M{
			bson.M{"dietaryprofile.vegan": recipePackRequest.Dietaryprofile.Vegan},
			bson.M{"dietaryprofile.raw": recipePackRequest.Dietaryprofile.Raw},
			bson.M{"mealplan": recipePackRequest.Mealplan}}}).All(recipePacks)
	case recipePackRequest.Dietaryprofile.Vegan:
		err = c.Find(bson.M{"$and": []bson.M{
			bson.M{"dietaryprofile.vegan": recipePackRequest.Dietaryprofile.Vegan},
			bson.M{"mealplan": recipePackRequest.Mealplan}}}).All(recipePacks)
	default:
		r.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "mongodb",
			"rpc":   "GetRecipePacks"},
			"Dietary Profile not provided")
		return grpc.Errorf(codes.InvalidArgument, "Dietary Profile not provided.")
	}

	// err := c.Find(bson.M{"$and": []bson.M{
	// 	bson.M{"dietaryprofile.omnivore": recipePackRequest.Dietaryprofile.Omnivore},
	// 	bson.M{"dietaryprofile.vegetarian": recipePackRequest.Dietaryprofile.Vegetarian},
	// 	bson.M{"dietaryprofile.vegan": recipePackRequest.Dietaryprofile.Vegan},
	// 	bson.M{"dietaryprofile.raw": recipePackRequest.Dietaryprofile.Raw},
	// 	bson.M{"dietaryrestriction.glutenfree": recipePackRequest.Dietaryrestriction.Glutenfree},
	// 	bson.M{"dietaryrestriction.nutfree": recipePackRequest.Dietaryrestriction.Nutfree},
	// 	bson.M{"dietaryrestriction.dairyfree": recipePackRequest.Dietaryrestriction.Dairyfree},
	// 	bson.M{"dietaryrestriction.soyfree": recipePackRequest.Dietaryrestriction.Soyfree},
	// 	bson.M{"dietaryrestriction.lowsodium": recipePackRequest.Dietaryrestriction.Lowsodium},
	// 	bson.M{"mealplan": recipePackRequest.Mealplan}}}).All(recipePacks)
	if err != nil {
		if err == mgo.ErrNotFound {
			r.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "mongodb",
				"rpc":   "GetRecipePacks"},
				fmt.Sprintf("Recipe Pack not found: %v", err))
			return grpc.Errorf(codes.NotFound, "Recipepacks not found.")
		}
		errMsg := fmt.Sprintf("Unknown recipepack lookup error: %v", err)
		r.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "mongodb",
			"rpc":   "GetRecipePacks"},
			errMsg)
		return grpc.Errorf(codes.Unknown, errMsg)
	}

	for _, recipePack := range *recipePacks {
		if err := stream.Send(&recipePack); err != nil {
			return err
		}
		infoMsg := fmt.Sprintf("Returning recipepack %s with ID: %s", recipePack.Name, recipePack.Id)
		r.Logger.Info(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "mongodb",
			"rpc":   "GetRecipePacks"},
			infoMsg)
	}
	return err
}
