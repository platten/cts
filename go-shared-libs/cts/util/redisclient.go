/*
// ----------------------------------------------------------------------------
// redisutils.go
// Countertop Redis Utility Functions

// Created by Paul Pietkiewicz on 10/1/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package util

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"

	"github.com/fzzy/radix/redis"
	//"github.com/fzzy/radix/extra/sentinel"
)

func NewClient(loggerObject *logger.CtsLogger, host string, pass string) (*redis.Client, error) {
	redisClient, err := redis.DialTimeout("tcp", host, time.Second*60)
	if err != nil {
		loggerObject.Error(logrus.Fields{
			"phase": "connection",
			"event": "dial",
			"tag":   "redis",
			"rpc":   "NewClient"},
			fmt.Sprintf("Cannot connect to Redis server running on: %s. Error: %v", host, err))
		return nil, err
	}

	// NOTE: Code is assuming you know if you need to authenticate with Redis or not.
	// If no password is provided, and Redis expects a password, will get NOAUTH errors
	// when commands executed.
	if pass != "" {
		err = redisClient.Cmd("AUTH", pass).Err
		if err != nil {
			loggerObject.Error(logrus.Fields{
				"phase": "connection",
				"event": "authenticate",
				"tag":   "redis",
				"rpc":   "NewClient"},
				fmt.Sprintf("Cannot authenticate with Redis server running on: %s. Error: %v", host, err))
			redisClient.Close()
			return nil, err
		}
	}
	loggerObject.Info(logrus.Fields{
		"phase": "connection",
		"event": "connected",
		"tag":   "redis",
		"rpc":   "NewClient"},
		fmt.Sprintf("Successfully conencted to Redis server running on: %s", host))
	return redisClient, nil
}

func CreateRedisEntry(authenticatedRedisClient *redis.Client, key string, value string) error {
	err := authenticatedRedisClient.Cmd("SET", key, value, "EX", 86400).Err
	if err != nil {
		errorString := fmt.Sprintf("Could not set value \"%s\" for key \"%s\". Error: %v", value, key, err)
		return errors.New(errorString)
	}
	return nil
}

func UpdateTTL(authenticatedRedisClient *redis.Client, key string, ttl int) error {
	if err := authenticatedRedisClient.Cmd("EXPIRE", key, ttl).Err; err != nil {
		errorString := fmt.Sprintf("Could not update TTL for Redis key %s. Error: %v", key, err)
		return errors.New(errorString)
	}
	return nil
}

func GetRedisStringValue(authenticatedRedisClient *redis.Client, key string) (string, error) {
	value, err := authenticatedRedisClient.Cmd("GET", key).Str()
	if err != nil {
		errorString := fmt.Sprintf("Could not get value for Redis key %s. Error: %v", key, err)
		return "", errors.New(errorString)
	}
	return value, nil
}

func CheckForRedisKey(authenticatedRedisClient *redis.Client, key string) (bool, error) {
	exists, err := authenticatedRedisClient.Cmd("EXISTS", key).Bool()
	if err != nil {
		errorString := fmt.Sprintf("Could not check if redis key %s exists. Error: %v", key, err)
		return false, errors.New(errorString)
	}
	return exists, nil
}

func DelRedisKey(authenticatedRedisClient *redis.Client, keys ...string) (int, error) {
	var interfaceList []interface{}
	for _, key := range keys {
		var interf interface{} = key
		interfaceList = append(interfaceList, interf)
	}
	deleteCount, deleteErr := authenticatedRedisClient.Cmd("DEL", interfaceList...).Int()
	if deleteErr != nil {
		errorString := fmt.Sprintf("Could not delete %v keys from redis. Error: %v", keys, deleteErr)
		return 0, errors.New(errorString)
	}
	return deleteCount, nil
}

func BlowAwayRedis(authenticatedRedisClient *redis.Client) error {
	_, err := authenticatedRedisClient.Cmd("FLUSHDB").Str()
	if err != nil {
		errorString := fmt.Sprintf("Could not flush Redis. Error: %v", err)
		return errors.New(errorString)
	}
	return nil
}
