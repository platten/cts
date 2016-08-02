/*
// ----------------------------------------------------------------------------
// grpcconnpool.go
// Countertop gRPC Connection Pool Management Library

// Created by Paul Pietkiewicz on 10/28/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package util

import (
	"fmt"
	//"reflect"
	"github.com/Sirupsen/logrus"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc"
)

type HostConnPool struct {
	logger      *logger.CtsLogger
	serviceName string
	pool        chan *grpc.ClientConn
	poolSize    int
	host        string
	port        int
	tls         bool
	certFile    string
}


func (p *HostConnPool) getSingleConn() (*grpc.ClientConn, error) {
	var serverOpts []grpc.DialOption
	if p.tls {
		var sn string
		var creds credentials.TransportAuthenticator
		if p.certFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(p.certFile, sn)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to create TLS credentials for %s service. Error: %v", p.serviceName, err)
				p.logger.Error(logrus.Fields{
					"phase": "startup",
					"event": "connection",
					"tag":   p.serviceName},
					errorMsg)
				return nil, err
			}
		} else {
			creds = credentials.NewClientTLSFromCert(nil, sn)
		}
		serverOpts = append(serverOpts, grpc.WithTransportCredentials(creds))
	} else {
		serverOpts = append(serverOpts, grpc.WithInsecure())
	}

	client, err := grpc.Dial(fmt.Sprintf("%s:%d", p.host, p.port), serverOpts...)
	if err != nil {
		p.logger.Error(logrus.Fields{
			"phase": "startup",
			"event": "connection",
			"tag":   p.serviceName},
			fmt.Sprintf("Fail to dial %s service: %v", p.serviceName, err))
		return nil, err
	}
	return client, nil

	// TODO(ppietkiewicz): test connection before passing it off. Use simple health checkl
	// NOTE: services need to have health checks setup and loaded (or at least have ping rpc)
}

func NewConnPool(loggerObj *logger.CtsLogger, serviceName string, host string, port int, tls bool, certFile string, maxSize int) (*HostConnPool, error) {
	hostConnPool := new(HostConnPool)
	hostConnPool.logger = loggerObj
	hostConnPool.serviceName = serviceName
	hostConnPool.host = host
	hostConnPool.port = port
	hostConnPool.tls = tls
	hostConnPool.certFile = certFile
	hostConnPool.poolSize = maxSize

	pool := make([]*grpc.ClientConn, 0, maxSize)

	for i := 0; i < maxSize; i++ {
		client, err := hostConnPool.getSingleConn()
		if err != nil {
			hostConnPool.logger.Error(logrus.Fields{
				"phase": "startup",
				"event": "connection",
				"tag":   serviceName},
				fmt.Sprintf("Problem establishing connection to service %s, closing existing pool connections. Error: %v", serviceName, err))

			for _, client = range pool {
				client.Close()
			}
			return nil, err
		}
		if client != nil {
			pool = append(pool, client)
			hostConnPool.logger.Info(logrus.Fields{
				"phase": "startup",
				"event": "connection",
				"tag":   serviceName},
				fmt.Sprintf("Added connection #%d for %s service to pool.", i, serviceName))
		}
	}
	for i := range pool {
		hostConnPool.pool <- pool[i]
	}

	hostConnPool.logger.Info(logrus.Fields{
		"phase": "connection",
		"event": "connected",
		"tag":   serviceName},
		fmt.Sprintf("Successfully created pool with %d connections with %s service running on: %s", len(hostConnPool.pool), serviceName, host))

	return hostConnPool, nil
}

func (p *HostConnPool) Get() (*grpc.ClientConn, error) {
	select {
	case conn := <-p.pool:
		return conn, nil
	default:
		p.logger.Info(logrus.Fields{
			"phase": "connection",
			"event": "poolget",
			"tag":   p.serviceName},
			"Pool full, connection terminated.")
		return p.getSingleConn()
	}
}

func (p *HostConnPool) Put(conn *grpc.ClientConn) {
	select {
	case p.pool <- conn:
		p.logger.Info(logrus.Fields{
			"phase": "connection",
			"event": "poolput",
			"tag":   p.serviceName},
			"Connection successfully returned to pool.")
	default:
		p.logger.Info(logrus.Fields{
			"phase": "connection",
			"event": "poolput",
			"tag":   p.serviceName},
			"Pool full, connection terminated.")
		conn.Close()
	}
}

func (p *HostConnPool) CarefullyPut(conn *grpc.ClientConn, potentialErr *error) {
	if potentialErr != nil && *potentialErr != nil {
		switch *potentialErr {
		case grpc.ErrClientConnClosing, grpc.ErrClientConnTimeout, grpc.ErrCredentialsMisuse, grpc.ErrNoTransportSecurity, grpc.ErrUnspecTarget:
			p.logger.Error(logrus.Fields{
				"phase": "connection",
				"event": "poolput",
				"tag":   p.serviceName},
				fmt.Sprintf("Received connection error: %v Closing connection.", potentialErr))
			conn.Close()
			return
		default:
			p.logger.Error(logrus.Fields{
				"phase": "connection",
				"event": "poolput",
				"tag":   p.serviceName},
				fmt.Sprintf("Recieved other type of error from connection. Returning connection pool on %s. Error: %v", p.host, potentialErr))
		}
	}
	p.Put(conn)
}
