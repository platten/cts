/*
// ----------------------------------------------------------------------------
// event.go
// Countertop Server Event Recording Microservice Library

// Created by Paul Pietkiewicz on 9/29/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package event

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"github.com/golang/blog/content/context/userip"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"

	pb "github.com/theorangechefco/cts/go-protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const insertTemplate string = `INSERT INTO events (id, version, apprelease, appversion,
	 carrier, city, country, devicemodel, manufacturer, model , osversion , operatingsystem , radio,
	 region, screenheight, screenwidth, wifi, createdat, payload) VALUES
	 (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

type Server struct {
	IdentityClient pb.IdentityServiceClient
	Session        *gocql.Session
	Logger         *logger.CtsLogger
}

func (s *Server) WriteEvent(ctx context.Context, event *pb.Event) (*pb.EmptyRequest, error) {
	md, _ := metadata.FromContext(ctx)
	token, ok := md["token"]
	available := false
	var ip net.IP
	if ok == false {
		ip, available = userip.FromContext(ctx)
		if available {
			s.Logger.Error(logrus.Fields{
				"phase": "authorization",
				"event": "connection",
				"tag":   "missingtoken"},
				fmt.Sprintf("Token from %s not provided, access denied", ip.String()))
		} else {
			s.Logger.Error(logrus.Fields{
				"phase": "authorization",
				"event": "connection",
				"tag":   "missingtoken"},
				"Token from not provided, access denied")
		}

		return nil, grpc.Errorf(codes.Unauthenticated, "Valid session token not provided, access denied.")
	}

	userID, err := s.IdentityClient.LookupSessionToken(ctx, &pb.SessionToken{Id: strings.Join(token, "")})
	if err != nil {
		if available {
			s.Logger.Error(logrus.Fields{
				"phase": "authorization",
				"event": "connection",
				"tag":   "validtoken"},
				fmt.Sprintf("Valid token from %s not provided, access denied", ip.String()))
		} else {
			s.Logger.Error(logrus.Fields{
				"phase": "authorization",
				"event": "connection",
				"tag":   "validtoken"},
				"Valid token from not provided, access denied")
		}
		return nil, grpc.Errorf(codes.Unauthenticated, "Valid session token not provided, access denied.")
	}
	ID := gocql.TimeUUID()
	err = s.Session.Query(insertTemplate, ID, event.Version, event.Apprelease,
		event.Appversion, event.Carrier, event.City, event.Country,
		event.Devicemodel, event.Manufacturer, event.Model, event.Osversion,
		event.Operatingsystem, event.Radio, event.Region, event.Screenheight,
		event.Screenwidth, event.Wifi, time.Now().Unix(), event.JsonPayload).Exec()
	if err != nil {
		s.Logger.Error(logrus.Fields{
			"phase": "persist",
			"event": "connection",
			"tag":   "cassandra"},
			fmt.Sprintf("Could not write to Cassandra, error %v", err))
		return nil, grpc.Errorf(codes.Internal, "Unable to write to Cassandra %v", err)
	}
	s.Logger.Info(logrus.Fields{
		"phase": "persist",
		"event": "connection",
		"tag":   "cassandra"},
		fmt.Sprintf("Wrote entry into Cassandra, ID %s for user with ID %s", ID, userID))
	return &pb.EmptyRequest{}, nil
}
