/*
// ----------------------------------------------------------------------------
// profile.go
// Countertop Profile Microservice Utility Functions

// Created by Paul Pietkiewicz on 9/1/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package profile

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/theorangechefco/cts/go-protos"
	"google.golang.org/grpc/codes"

	"github.com/pborman/uuid"
)

type Server struct {
	DB     gorm.DB
	Logger *logger.CtsLogger
}

func (s *Server) GetUUID(ctx context.Context, identifier *pb.Identifier) (*pb.UserId, error) {
	var user User
	var query *gorm.DB
	var identifierString string

	if identifier.Useridentifier != "" {
		identifierString = identifier.Useridentifier
	} else if identifier.Deviceidentifier != "" {
		identifierString = identifier.Deviceidentifier
	} else {
		errorMsg := fmt.Sprintf("Identifier not specified.")
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "parseparameters",
			"tag":   "invalidparameters",
			"rpc":   "GetUUID"},
			errorMsg)
		return nil, grpc.Errorf(codes.InvalidArgument, errorMsg)
	}

	query = s.DB.Select("uuid").Where(&User{UserId: identifierString}).First(&user)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			errorMsg := fmt.Sprintf("User with userid %s not found.", identifier.Useridentifier)
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "database",
				"rpc":   "GetUUID"},
				errorMsg)
			return nil, grpc.Errorf(codes.NotFound, errorMsg)
		} else {
			errorMsg := fmt.Sprintf("Database query failed: %v", query.Error)
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "database",
				"rpc":   "GetUUID"},
				errorMsg)
			return nil, grpc.Errorf(codes.Unknown, errorMsg)
		}
	}

	infoMsg := fmt.Sprintf("Returning UserID %s for user with identifier %s", user.UUID, identifierString)
	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "fetch",
		"tag":   "database",
		"rpc":   "GetUUID"},
		infoMsg)

	return &pb.UserId{Uuid: user.UUID}, nil
}

func (s *Server) GetProfileInfoByUUID(ctx context.Context, userID *pb.UserId) (*pb.Profile, error) {
	var user User

	if userID.Uuid == "" {
		errorMsg := fmt.Sprintf("Identifier not specified.")
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "parseparameters",
			"tag":   "invalidparameters",
			"rpc":   "GetProfileInfoByUUID"},
			errorMsg)
		return nil, grpc.Errorf(codes.InvalidArgument, errorMsg)
	}

	query := s.DB.Where(&User{UUID: userID.Uuid}).First(&user)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			errorMsg := fmt.Sprintf("Profile with id %s not found.", userID.Uuid)
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "database",
				"rpc":   "GetProfileInfoByUUID"},
				errorMsg)
			return nil, grpc.Errorf(codes.NotFound, errorMsg)
		} else {
			errorMsg := fmt.Sprintf("Database query failed: %v", query.Error)
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "database",
				"rpc":   "GetProfileInfoByUUID"},
				errorMsg)
			return nil, grpc.Errorf(codes.Unknown, errorMsg)
		}
	}
	identifier := pb.Identifier{
		Deviceidentifier: user.DeviceId,
		Useridentifier:   user.UserId,
	}

	infoMsg := fmt.Sprintf("Returning profile for user with UserID %s", user.UUID)
	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "fetch",
		"tag":   "database",
		"rpc":   "GetUUID"},
		infoMsg)

	return &pb.Profile{
		Identifier:    &identifier,
		Firstname:     user.Firstname,
		Birthyear:     user.Birthyear,
		Gender:        pb.Gender(user.Gender),
		Heightcm:      user.Heightcm,
		Weightkg:      user.Weightkg,
		Goalweightkg:  user.Goalweightkg,
		Activitylevel: pb.ActivityLevel(user.Activitylevel),
		Mealplan:      pb.MealPlan(user.Mealplan),
		Weightgoal:    pb.WeightGoal(user.Weightgoal),
		Dietaryprofile: &pb.DietaryProfile{
			Omnivore:   user.Omnivore,
			Vegetarian: user.Vegetarian,
			Vegan:      user.Vegan,
			Raw:        user.Raw,
		},
		Dietaryrestriction: &pb.DietaryRestriction{
			Glutenfree: user.Glutenfree,
			Nutfree:    user.Nutfree,
			Dairyfree:  user.Dairyfree,
			Soyfree:    user.Soyfree,
			Lowsodium:  user.Lowsodium,
		},
	}, nil
}

func (s *Server) CreateProfile(ctx context.Context, Profile *pb.Profile) (*pb.UserId, error) {
	user := User{
		DeviceId:      Profile.Identifier.Deviceidentifier,
		UserId:        Profile.Identifier.Useridentifier,
		UUID:          uuid.NewRandom().String(),
		Firstname:     Profile.Firstname,
		Birthyear:     Profile.Birthyear,
		Gender:        int32(Profile.Gender),
		Heightcm:      Profile.Heightcm,
		Weightkg:      Profile.Weightkg,
		Goalweightkg:  Profile.Goalweightkg,
		Activitylevel: int32(Profile.Activitylevel),
		Mealplan:      int32(Profile.Mealplan),
		Weightgoal:    int32(Profile.Weightgoal),
		Omnivore:      Profile.Dietaryprofile.Omnivore,
		Vegetarian:    Profile.Dietaryprofile.Vegetarian,
		Vegan:         Profile.Dietaryprofile.Vegan,
		Raw:           Profile.Dietaryprofile.Raw,
		Glutenfree:    Profile.Dietaryrestriction.Glutenfree,
		Nutfree:       Profile.Dietaryrestriction.Nutfree,
		Dairyfree:     Profile.Dietaryrestriction.Dairyfree,
		Soyfree:       Profile.Dietaryrestriction.Soyfree,
		Lowsodium:     Profile.Dietaryrestriction.Lowsodium,
	}
	//TODO(ppietkiewicz): Ensure that duplicate records aren't entered and handle error gracefully
	query := s.DB.Create(&user)
	if query.Error != nil {
		errorMsg := fmt.Sprintf("Could not add record for profile with ID: %v. Error: %v", Profile.Identifier, query.Error)
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "createrecord",
			"tag":   "database",
			"rpc":   "CreateProfile"},
			errorMsg)
		return nil, grpc.Errorf(codes.Unknown, errorMsg)
	}

	infoMsg := fmt.Sprintf("Successfuly created profile with UUID: %s", user.UUID)
	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "fetch",
		"tag":   "database",
		"rpc":   "CreateProfile"},
		infoMsg)

	return &pb.UserId{Uuid: user.UUID}, nil
}

func (s *Server) SetProfileInfo(ctx context.Context, profileUpdateReq *pb.ProfileUpdateRequest) (*pb.Response, error) {
	var user User

	query := s.DB.Where(&User{UUID: profileUpdateReq.Id.Uuid}).First(&user)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			errorMsg := fmt.Sprintf("Profile with ID %s not found.", profileUpdateReq.Id.Uuid)
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "database",
				"rpc":   "SetProfileInfo"},
				errorMsg)
			return nil, grpc.Errorf(codes.NotFound, errorMsg)
		} else {
			errorMsg := fmt.Sprintf("Database query failed: %v", query.Error)
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "database",
				"rpc":   "SetProfileInfo"},
				errorMsg)
			return nil, grpc.Errorf(codes.Unknown, errorMsg)
		}
	}

	// Not updating deviceid, useridentifier, id or uuid
	userMap := map[string]interface{}{
		"Firstname":     profileUpdateReq.Profile.Firstname,
		"Birthyear":     profileUpdateReq.Profile.Birthyear,
		"Gender":        int32(profileUpdateReq.Profile.Gender),
		"Heightcm":      profileUpdateReq.Profile.Heightcm,
		"Weightkg":      profileUpdateReq.Profile.Weightkg,
		"Goalweightkg":  profileUpdateReq.Profile.Goalweightkg,
		"Activitylevel": int32(profileUpdateReq.Profile.Activitylevel),
		"Mealplan":      int32(profileUpdateReq.Profile.Mealplan),
		"Weightgoal":    int32(profileUpdateReq.Profile.Weightgoal),
		"Omnivore":      profileUpdateReq.Profile.Dietaryprofile.Omnivore,
		"Vegetarian":    profileUpdateReq.Profile.Dietaryprofile.Vegetarian,
		"Vegan":         profileUpdateReq.Profile.Dietaryprofile.Vegan,
		"Raw":           profileUpdateReq.Profile.Dietaryprofile.Raw,
		"Glutenfree":    profileUpdateReq.Profile.Dietaryrestriction.Glutenfree,
		"Nutfree":       profileUpdateReq.Profile.Dietaryrestriction.Nutfree,
		"Dairyfree":     profileUpdateReq.Profile.Dietaryrestriction.Dairyfree,
		"Soyfree":       profileUpdateReq.Profile.Dietaryrestriction.Soyfree,
		"Lowsodium":     profileUpdateReq.Profile.Dietaryrestriction.Lowsodium,
	}
	query = s.DB.Model(&user).Updates(userMap)
	if query.Error != nil {
		errorMsg := fmt.Sprintf("Could not update profile with ID %s., Error: %v", profileUpdateReq.Id.Uuid, query.Error)
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "update",
			"tag":   "database",
			"rpc":   "SetProfileInfo"},
			errorMsg)
		return nil, grpc.Errorf(codes.Unknown, errorMsg)
	}

	infoMsg := fmt.Sprintf("Successfully updated profile with UserID %s", user.UUID)
	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "update",
		"tag":   "database",
		"rpc":   "SetProfileInfo"},
		infoMsg)

	return &pb.Response{true}, nil
}
