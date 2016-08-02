/*
// ----------------------------------------------------------------------------
// endpoint.go
// Countertop Server Endpoint Microservice Library

// Created by Paul Pietkiewicz on 9/15/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package endpoint

import (
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	pb "github.com/theorangechefco/cts/go-protos"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	"github.com/theorangechefco/cts/go-shared-libs/cts/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Server struct {
	RecipePool  *util.HostConnPool
	IdentityPool  *util.HostConnPool
	ProfilePool   *util.HostConnPool
	Logger         *logger.CtsLogger
}

func (s *Server) GetRecipe(ctx context.Context, recipeReq *pb.RecipeRequest) (*pb.Recipe, error) {
	identConn, identClient, identPoolErr := s.getIdentityClient("GetRecipe")
	defer s.IdentityPool.CarefullyPut(identConn, &identPoolErr)
	if identPoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching identity service connection from pool for recipe request with ID: %s. Error: %v", recipeReq.Recipeid, identPoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}
	if _, _, err := util.Authenticate(s.Logger, identClient, ctx, "GetRecipe", true); err != nil {
		return nil, err
	}

	recipeConn, recipeClient, recipePoolErr := s.getRecipeClient("GetRecipe")
	defer s.RecipePool.CarefullyPut(recipeConn, &recipePoolErr)
	if recipePoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching recipe service connection from pool for recipe request with ID: %s. Error: %v", recipeReq.Recipeid, identPoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}

	recipe, recipeErr := recipeClient.GetRecipe(context.Background(), recipeReq)
	if recipeErr != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "recipestore",
			"rpc":   "GetRecipe"},
			fmt.Sprintf("Cannot fetch recipe with ID %s. Error: %v", recipeReq.Recipeid, recipeErr))
		return nil, grpc.Errorf(codes.NotFound, "Cannot fetch recipe with ID %s.", recipeReq.Recipeid)
	}

	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "fetch",
		"tag":   "recipestore",
		"rpc":   "GetRecipe"},
		fmt.Sprintf("Successfully fetched recipe with ID %s ", recipeReq.Recipeid))

	return recipe, nil
}

func (s *Server) GetRecipePacks(recipePackRequest *pb.RecipePacksRequest, serviceStream pb.EndpointService_GetRecipePacksServer) error {
	ctx := serviceStream.Context()

	identConn, identClient, identPoolErr := s.getIdentityClient("GetRecipePacks")
	defer s.IdentityPool.CarefullyPut(identConn, &identPoolErr)
	if identPoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching identity service connection from pool for recipe pack request: %v. Error: %v", recipePackRequest, identPoolErr)
		return grpc.Errorf(codes.Internal, errorMsg)
	}
	if _, _, err := util.Authenticate(s.Logger, identClient, ctx, "GetRecipePacks", true); err != nil {
		return err
	}

	recipeConn, recipeClient, recipePoolErr := s.getRecipeClient("GetRecipePacks")
	defer s.RecipePool.CarefullyPut(recipeConn, &recipePoolErr)
	if recipePoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching recipe service connection from pool for recipe pack request: %v. Error: %v", recipePackRequest, identPoolErr)
		return grpc.Errorf(codes.Internal, errorMsg)
	}

	clientStream, err := recipeClient.GetRecipePacks(context.Background(), recipePackRequest)
	if err != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "recipestore"},
			fmt.Sprintf("Problem establishing connection to recipe service: %v", err))
		return grpc.Errorf(codes.Internal, "Problem pulling recipes")
	}

	counter := 0
	for {
		recipePack, err := clientStream.Recv()
		if err == io.EOF {
			break
		}
		if serverErr := serviceStream.Send(recipePack); serverErr != nil {
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "recipestore"},
				fmt.Sprintf("Problem sending data to client: %v", serverErr))
			return grpc.Errorf(codes.Internal, "Problem sending data to client.")
		}
		counter++
	}

	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "fetch",
		"tag":   "recipestore"},
		fmt.Sprintf("Returned %d recipes for recipepack request %v", counter, recipePackRequest))

	return nil
}

// Returns the session token based on the identifier
func (s *Server) GetSessionToken(ctx context.Context, identifier *pb.Identifier) (*pb.SessionToken, error) {
	profileConn, profileClient, profilePoolErr := s.getProfileClient("GetSessionToken")
	defer s.ProfilePool.CarefullyPut(profileConn, &profilePoolErr)
	if profilePoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from profile pool when fetching session token for: %v. Error: %v", identifier, profilePoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}

	userID, err := profileClient.GetUUID(context.Background(), identifier)
	if err != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "create",
			"tag":   "session",
			"rpc":   "GetSessionToken"},
			fmt.Sprintf("Cannot fetch for UUID for identifier %v, error: %v", identifier, err))
		return nil, grpc.Errorf(codes.Unauthenticated, "Invalid identifier %v.", identifier)
	}

	identConn, identClient, identPoolErr := s.getIdentityClient("GetSessionToken")
	defer s.IdentityPool.CarefullyPut(identConn, &identPoolErr)
	if identPoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from identity pool when fetching session token for: %v. Error: %v", identifier, identPoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}

	token, tokenGenErr := identClient.GenerateSessionToken(context.Background(), userID)
	if tokenGenErr != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "create",
			"tag":   "session",
			"rpc":   "GetSessionToken"},
			fmt.Sprintf("Cannot generate token for user with  UUID %s. Error: %v", userID.Uuid, tokenGenErr))
		return nil, grpc.Errorf(codes.Internal, "Cannot generate session token for user %s.", userID.Uuid)
	}

	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "create",
		"tag":   "session",
		"rpc":   "GetSessionToken"},
		fmt.Sprintf("Generated session token %s for user with identifier %v", token.Id, identifier))

	return token, nil
}

func (s *Server) CreateProfile(ctx context.Context, profile *pb.Profile) (*pb.SessionToken, error) {
	profileConn, profileClient, profilePoolErr := s.getProfileClient("CreateProfile")
	defer s.ProfilePool.CarefullyPut(profileConn, &profilePoolErr)
	if profilePoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from profile pool when fetching session token for: %v. Error: %v", profile, profilePoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}

	userID, err := profileClient.CreateProfile(context.Background(), profile)
	if err != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "create",
			"tag":   "profile",
			"rpc":   "CreateProfile"},
			fmt.Sprintf("Cannot create profile %v, error: %v", profile.Identifier, err))
		return nil, err
	}
	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "create",
		"tag":   "profile",
		"rpc":   "CreateProfile"},
		fmt.Sprintf("Successfully created new profile with UUID %s for user with identifier %v", userID.Uuid, profile.Identifier))

	identConn, identClient, identPoolErr := s.getIdentityClient("GetSessionToken")
	defer s.IdentityPool.CarefullyPut(identConn, &identPoolErr)
	if identPoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from identity pool when fetching session token for: %v. Error: %v", profile, identPoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}

	token, tokenGenErr := identClient.GenerateSessionToken(context.Background(), userID)
	if tokenGenErr != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "create",
			"tag":   "session",
			"rpc":   "CreateProfile"},
			fmt.Sprintf("Profile created, but cannot generate token for user with UUID %s. Error: %v", userID.Uuid, tokenGenErr))
		return nil, grpc.Errorf(codes.Internal, "Cannot generate session token for user %s.", userID.Uuid)
	}

	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "create",
		"tag":   "session",
		"rpc":   "CreateProfile"},
		fmt.Sprintf("Generatedtoken for user with UUID %s.", token.Id))

	return token, nil
}

func (s *Server) GetProfileInfo(ctx context.Context, null *pb.EmptyRequest) (*pb.Profile, error) {
	identConn, identClient, identPoolErr := s.getIdentityClient("GetProfileInfo")
	defer s.IdentityPool.CarefullyPut(identConn, &identPoolErr)
	if identPoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from identity pool. Error: %v", identPoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}
	_, userID, err := util.Authenticate(s.Logger, identClient, ctx, "GetProfileInfo", true)
	if err != nil {
		return nil, err
	}

	profileConn, profileClient, profilePoolErr := s.getProfileClient("CreateProfile")
	defer s.ProfilePool.CarefullyPut(profileConn, &profilePoolErr)
	if profilePoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from profile pool when fetching profile for user: %s. Error: %v", userID.Uuid, profilePoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}

	profile, profileErr := profileClient.GetProfileInfoByUUID(context.Background(), userID)
	if profileErr != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "profile",
			"rpc":   "GetProfileInfo"},
			fmt.Sprintf("Cannot fetch profile for user with  UUID %s. Error: %v", userID.Uuid, profileErr))
		return nil, grpc.Errorf(codes.Internal, "Cannot fetch profile token for user %s.", userID.Uuid)
	}

	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "fetch",
		"tag":   "profile",
		"rpc":   "GetProfileInfo"},
		fmt.Sprintf("Successfully fetched profile for user with UUID %s ", userID.Uuid))
	return profile, nil
}

func (s *Server) SetProfileInfo(ctx context.Context, profile *pb.Profile) (*pb.Response, error) {
	identConn, identClient, identPoolErr := s.getIdentityClient("SetProfileInfo")
	defer s.IdentityPool.CarefullyPut(identConn, &identPoolErr)
	if identPoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from identity pool. Error: %v", identPoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}
	_, userID, err := util.Authenticate(s.Logger, identClient, ctx, "SetProfileInfo", true)
	if err != nil {
		return nil, err
	}

	profileConn, profileClient, profilePoolErr := s.getProfileClient("SetProfileInfo")
	defer s.ProfilePool.CarefullyPut(profileConn, &profilePoolErr)
	if profilePoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from profile pool when fetching profile for user: %s. Error: %v", userID.Uuid, profilePoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}

	response, updateErr := profileClient.SetProfileInfo(context.Background(), &pb.ProfileUpdateRequest{Profile: profile, Id: userID})
	if updateErr != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "update",
			"tag":   "profile",
			"rpc":   "SetProfileInfo"},
			fmt.Sprintf("Cannot update profile for user with UUID %s. Error: %v", userID.Uuid, updateErr))
		return nil, grpc.Errorf(codes.Internal, "Cannot update profile.")
	}

	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "update",
		"tag":   "profile",
		"rpc":   "SetProfileInfo"},
		"Successfully updated profile.")
	return response, nil
}

// Closes the session by invalidating the token
func (s *Server) CloseSession(ctx context.Context, null *pb.EmptyRequest) (*pb.Response, error) {
	identConn, identClient, identPoolErr := s.getIdentityClient("CloseSession")
	defer s.IdentityPool.CarefullyPut(identConn, &identPoolErr)
	if identPoolErr != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from identity pool. Error: %v", identPoolErr)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}
	token, userID, err := util.Authenticate(s.Logger, identClient, ctx, "CloseSession", true)
	if err != nil {
		return nil, err
	}

	response, err := identClient.CloseSession(context.Background(), &pb.SessionToken{Id: token})
	if err != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "close",
			"tag":   "session",
			"rpc":   "CloseSession"},
			fmt.Sprintf("Cannot close session for user with token %s. Error: %v", token, err))
		return nil, grpc.Errorf(codes.Internal, "Cannot close session.")
	}

	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "close",
		"tag":   "session",
		"rpc":   "CloseSession"},
		fmt.Sprintf("Successfully closed session for user with UUID %s.", userID.Uuid))
	return response, nil
}

func (s *Server) getProfileClient(rpc string) (*grpc.ClientConn, pb.ProfileServiceClient, error) {
	conn, err := s.ProfilePool.Get()
	if err != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from profile pool. Error: %v", err)
		s.Logger.Error(logrus.Fields{
		"phase": "process",
		"event": "getconn",
		"tag":   "profile",
		"rpc":   rpc},
		errorMsg)
		return nil, nil, grpc.Errorf(codes.Internal, fmt.Sprintf("%v", err))
	}
	client := pb.NewProfileServiceClient(conn)
	return conn, client, nil
}

func (s *Server) getIdentityClient(rpc string) (*grpc.ClientConn, pb.IdentityServiceClient, error) {
	conn, err := s.IdentityPool.Get()
	if err != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from identity pool. Error: %v", err)
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "getconn",
			"tag":   "identity",
			"rpc":   rpc},
			errorMsg)
		return nil, nil, grpc.Errorf(codes.Internal, fmt.Sprintf("%v", err))
	}
	client := pb.NewIdentityServiceClient(conn)
	return conn, client, nil
}

func (s *Server) getRecipeClient(rpc string) (*grpc.ClientConn, pb.RecipeServiceClient, error) {
	conn, err := s.RecipePool.Get()
	if err != nil {
		errorMsg := fmt.Sprintf("Problem fetching connection from recipe pool. Error: %v", err)
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "getconn",
			"tag":   "recipe",
			"rpc":   rpc},
			errorMsg)
		return nil, nil, grpc.Errorf(codes.Internal, fmt.Sprintf("%v", err))
	}
	client := pb.NewRecipeServiceClient(conn)
	return conn, client, nil
}
