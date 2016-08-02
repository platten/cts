/*
// ----------------------------------------------------------------------------
// auth.go
// Countertop Server Authentication Utility Library

// Created by Paul Pietkiewicz on 10/12/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package util

import (
	"fmt"
	"net"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/golang/blog/content/context/userip"
	pb "github.com/theorangechefco/cts/go-protos"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func Authenticate(log *logger.CtsLogger, identityClient pb.IdentityServiceClient, ctx context.Context, rpcName string, lookupUser bool) (string, *pb.UserId, error) {
	md, _ := metadata.FromContext(ctx)
	token, ok := md["token"]
	available := false
	var ip net.IP
	if ok == false {
		ip, available = userip.FromContext(ctx)
		if available {
			log.Error(logrus.Fields{
				"phase": "authorization",
				"event": "connection",
				"tag":   "missingtoken",
				"rpc":   rpcName},
				fmt.Sprintf("Token from %s not provided, access denied", ip.String()))
		} else {
			log.Error(logrus.Fields{
				"phase": "authorization",
				"event": "connection",
				"tag":   "missingtoken",
				"rpc":   rpcName},
				"Token from not provided, access denied")
		}

		return "", nil, grpc.Errorf(codes.Unauthenticated, "Valid session token not provided, access denied.")
	}

	tokenString := strings.Join(token, "")

	if available {
		log.Info(logrus.Fields{
			"phase": "authorization",
			"event": "connection",
			"tag":   "validtoken",
			"rpc":   rpcName},
			fmt.Sprintf("Token %s provided from %s, continuing", tokenString, ip.String()))
	} else {
		log.Info(logrus.Fields{
			"phase": "authorization",
			"event": "connection",
			"tag":   "validtoken",
			"rpc":   rpcName},
			fmt.Sprintf("Token %s provided, continuing", tokenString))
	}

	if lookupUser {
		userID, err := identityClient.LookupSessionToken(ctx, &pb.SessionToken{Id: tokenString})
		if err != nil {
			// TODO(ppietkiewicz): diferentiate between returned error types and retry on timeouts.
			if available {
				log.Error(logrus.Fields{
					"phase": "authorization",
					"event": "connection",
					"tag":   "validtoken",
					"rpc":   rpcName},
					fmt.Sprintf("Valid token from %s not provided, access denied", ip.String()))
			} else {
				log.Error(logrus.Fields{
					"phase": "authorization",
					"event": "connection",
					"tag":   "validtoken",
					"rpc":   rpcName},
					"Valid token from not provided, access denied")
			}
			return tokenString, nil, grpc.Errorf(codes.Unauthenticated, "Valid session token not provided, access denied.")
		}

		if available {
			log.Info(logrus.Fields{
				"phase": "authorization",
				"event": "connection",
				"tag":   "validtoken",
				"rpc":   rpcName},
				fmt.Sprintf("Valid token from %s for user %s provided, continuing", ip.String(), userID.Uuid))
		} else {
			log.Info(logrus.Fields{
				"phase": "authorization",
				"event": "connection",
				"tag":   "validtoken",
				"rpc":   rpcName},
				fmt.Sprintf("Valid token for user %s provided, continuing", userID.Uuid))
		}
		return tokenString, userID, nil
	}
	return tokenString, nil, nil
}
