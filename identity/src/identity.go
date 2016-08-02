/*
// ----------------------------------------------------------------------------
// identity.go
// Countertop Identity Microservice Utility Functions

// Created by Paul Pietkiewicz on 10/12/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package identity

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	"github.com/theorangechefco/cts/go-shared-libs/cts/util"

	"github.com/fzzy/radix/redis"
	misc "github.com/theorangechefco/cts/identity/misc"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"strings"

	pb "github.com/theorangechefco/cts/go-protos"
)

type Server struct {
	Logger *logger.CtsLogger
	Pool   *util.RedisHandler
	TTL    int
}

func (s *Server) GenerateSessionToken(ctx context.Context, userID *pb.UserId) (*pb.SessionToken, error) {
	var replies []*redis.Reply
	var reply *redis.Reply
	var err error

	userKey := strings.Join([]string{"user", userID.Uuid}, "_")
	reply, err = s.Pool.Get(userKey)
	token, tokenErr := reply.Str()
	if tokenErr != nil || err != nil {
		s.Logger.Info(logrus.Fields{
			"phase": "process",
			"event": "update",
			"tag":   "redis",
			"rpc":   "GenerateSessionToken"},
			fmt.Sprintf("User key %s does not exist in redis.", userKey))
	}

	if token != "" {
		sessionTTL := time.Now().Add(time.Duration(s.TTL) * time.Second).Unix()
		replies, err = s.Pool.Expire(s.TTL, userKey, token)
		if err != nil {
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "update",
				"tag":   "redis",
				"rpc":   "GenerateSessionToken"},
				fmt.Sprintf("Cannot update TTL for keys %s, %s. Error: %v", userKey, token, err))
			return nil, grpc.Errorf(codes.Internal, "Redis connection problem, cannot update TTL user %s.", userID.Uuid)
		}

		return &pb.SessionToken{Id: token, Ttl: &pb.Timestamp{Seconds: sessionTTL}}, nil
	}

	// Token does not exist, lets generate a new one
	sessionTokenValue, _ := misc.GenerateSessionKey()
	sessionTokenKey := strings.Join([]string{"token", sessionTokenValue}, "_")

	// Ensure token is unique
	for {
		existsReply, existsErr := s.Pool.Exists(sessionTokenKey)
		exists, existsReplyErr := existsReply.Bool()
		if existsReplyErr != nil || existsErr != nil {
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "fetch",
				"tag":   "redis",
				"rpc":   "GenerateSessionToken"},
				fmt.Sprintf("Cannot check if key session token %s exists. Error: %v, %v", sessionTokenKey, existsErr, existsReplyErr))
			return nil, grpc.Errorf(codes.Internal, "Redis connection problem, cannot check session token candidate for user %s.", userID.Uuid)
		}
		if exists {
			sessionTokenValue, _ = misc.GenerateSessionKey()
			sessionTokenKey = strings.Join([]string{"token_", sessionTokenValue}, "_")
		} else {
			break
		}
	}

	sessionTTL := time.Now().Add(time.Duration(s.TTL) * time.Second).Unix()
	replies, setErr := s.Pool.SetMany(s.TTL, []string{userKey, sessionTokenValue}, []string{sessionTokenKey, userID.Uuid})
	if setErr != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "update",
			"tag":   "redis",
			"rpc":   "GenerateSessionToken"},
			fmt.Sprintf("Redis problem, could not set value for user %s. Err: %v", sessionTokenKey, setErr))
		return nil, grpc.Errorf(codes.Internal, "Redis problem, could not set value for user %s. Err: %v", userID.Uuid, setErr)
	}

	for _, reply = range replies {
		if reply.Err != nil {
			s.Logger.Error(logrus.Fields{
				"phase": "process",
				"event": "update",
				"tag":   "redis",
				"rpc":   "GenerateSessionToken"},
				fmt.Sprintf("Redis problem, could not set value for user %s. Err: %v", userID.Uuid, reply.Err))
		}
	}

	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "respond",
		"tag":   "identity",
		"rpc":   "GenerateSessionToken"},
		fmt.Sprintf("Token %s successfully generated for user %s", token, userID.Uuid))

	return &pb.SessionToken{Id: sessionTokenValue, Ttl: &pb.Timestamp{Seconds: sessionTTL}}, nil
}

func (s *Server) LookupSessionToken(ctx context.Context, token *pb.SessionToken) (*pb.UserId, error) {
	var reply *redis.Reply
	var err error
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")

	reply, err = s.Pool.Get(sessionTokenKey)
	uuid, uuidErr := reply.Str()
	if uuidErr != nil || err != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "redis",
			"rpc":   "LookupSessionToken"},
			fmt.Sprintf("Cannot fetch user ID for session token %s. Error: %v, %v", sessionTokenKey, uuidErr, err))
		return nil, grpc.Errorf(codes.Internal, "Cannot fetch user ID for session token %s", sessionTokenKey)
	}

	if uuid == "" {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "redis",
			"rpc":   "LookupSessionToken"},
			fmt.Sprintf("Session token %s does not exist", token.Id))
		return nil, grpc.Errorf(codes.Internal, "Session token %s does not exist", token.Id)
	}

	s.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "respond",
		"tag":   "identity",
		"rpc":   "LookupSessionToken"},
		fmt.Sprintf("Returning User ID %s for token %s ", uuid, token.Id))

	return &pb.UserId{Uuid: uuid}, nil
}

func (s *Server) CloseSession(ctx context.Context, token *pb.SessionToken) (*pb.Response, error) {
	var reply *redis.Reply
	var err error
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")

	reply, err = s.Pool.Get(sessionTokenKey)
	uuid, uuidErr := reply.Str()
	if uuidErr != nil || err != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "redis",
			"rpc":   "CloseSession"},
			fmt.Sprintf("Cannot fetch user ID for session token %s. Error: %v, %v", sessionTokenKey, uuidErr, err))
		return nil, grpc.Errorf(codes.Internal, "Cannot fetch user ID for session token %s", sessionTokenKey)
	}

	if uuid == "" {
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "fetch",
			"tag":   "redis",
			"rpc":   "CloseSession"},
			fmt.Sprintf("Session token %s does not exist", token.Id))
		return nil, grpc.Errorf(codes.Unknown, "Session token %s does not exist", token.Id)
	}

	userKey := strings.Join([]string{"user", uuid}, "_")
	reply, err = s.Pool.Del(sessionTokenKey, userKey)

	deletedItems, deleteErr := reply.Int()
	if deleteErr != nil || err != nil {
		errorMsg := fmt.Sprintf("Could not delete session and user keys. Error: %v, %v", deleteErr, err)
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "delete",
			"tag":   "redis",
			"rpc":   "CloseSession"},
			errorMsg)

		return nil, grpc.Errorf(codes.Unknown, errorMsg)
	}
	if deletedItems != 2 {
		errorMsg := fmt.Sprintf("Inconsistant state, %d keys deleted (should be 2)", deletedItems)
		s.Logger.Error(logrus.Fields{
			"phase": "process",
			"event": "delete",
			"tag":   "redis",
			"rpc":   "CloseSession"},
			errorMsg)
		return nil, grpc.Errorf(codes.Internal, errorMsg)
	}
	return &pb.Response{Success: true}, nil
}
