/*
// ----------------------------------------------------------------------------
// redispool.go
// Countertop Redis Redis Pool Management Utility Libraries

// Created by Paul Pietkiewicz on 10/1/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/fzzy/radix/extra/pool"
	"github.com/fzzy/radix/redis"
	"github.com/theorangechefco/cts/go-shared-libs/cts/logger"
)

//"github.com/fzzy/radix/extra/sentinel"

const SLEEPMILLISECS = 500

type RedisHandler struct {
	Logger     *logger.CtsLogger
	pool       *pool.Pool
	client     *redis.Client
	retryTimes int
	retrySleep int
	host       string
}

func NewPool(loggerObject *logger.CtsLogger, host string, pass string, size int, retryTimes int, retrySleep int) (*RedisHandler, error) {
	redisHandler := new(RedisHandler)
	redisHandler.Logger = loggerObject
	redisHandler.retryTimes = retryTimes
	redisHandler.retrySleep = retrySleep
	redisHandler.host = host

	var err error

	df := func(network, addr string) (*redis.Client, error) {
		client, err := redis.Dial(network, addr)
		if err != nil {
			redisHandler.Logger.Error(logrus.Fields{
				"phase": "connection",
				"event": "pooldial",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Cannot connect with Redis server running on: %s. Error: %v", host, err))
			return nil, err
		}
		if err = client.Cmd("AUTH", pass).Err; err != nil {
			redisHandler.Logger.Error(logrus.Fields{
				"phase": "connection",
				"event": "poolauthenticate",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Cannot authenticate with Redis server running on: %s. Error: %v", host, err))
			client.Close()
			return nil, err
		}
		redisHandler.Logger.Info(logrus.Fields{
			"phase": "connection",
			"event": "poolconnected",
			"tag":   "redis",
			"rpc":   "NewPool"},
			"Successfully added connection with Redis server to pool.")
		return client, nil
	}

	redisHandler.pool, err = pool.NewCustomPool("tcp", host, size, df)
	if err != nil {
		redisHandler.Logger.Error(logrus.Fields{
			"phase": "connection",
			"event": "createpool",
			"tag":   "redis",
			"rpc":   "NewPool"},
			fmt.Sprintf("Could not create Redis pool. Error: %v", err))
	}

	// Test pool
	redisHandler.Logger.Info(logrus.Fields{
		"phase": "connection",
		"event": "pooltest",
		"tag":   "redis",
		"rpc":   "NewPool"},
		"Testing pool connection when ping")

	conn, redisErr := redisHandler.pool.Get()
	defer redisHandler.pool.CarefullyPut(conn, &redisErr)
	if redisErr == nil {
		reply := conn.Cmd("PING")
		if reply.Err != nil {
			redisHandler.Logger.Error(logrus.Fields{
				"phase": "connection",
				"event": "poolconnect",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Could not PING connection to pool: %s Error: %v", host, reply.Err))
			return nil, err

		}
	} else {
		redisHandler.Logger.Error(logrus.Fields{
			"phase": "connection",
			"event": "poolconnect",
			"tag":   "redis",
			"rpc":   "NewPool"},
			fmt.Sprintf("Could not PING connection to pool: %s Error: %v", host, redisErr))
		return nil, err

	}

	redisHandler.Logger.Info(logrus.Fields{
		"phase": "connection",
		"event": "connected",
		"tag":   "redis",
		"rpc":   "NewPool"},
		fmt.Sprintf("Successfully created pool with %d connections with Redis server running on: %s", size, host))

	return redisHandler, nil
}

func (c *RedisHandler) runCommand(command string, args ...interface{}) (*redis.Reply, error) {
	var i int
	var err error

	conn, redisErr := c.pool.Get()
	defer c.pool.CarefullyPut(conn, &redisErr)
	for i = 1; i <= c.retryTimes; i++ {
		if redisErr != nil || conn == nil {
			c.Logger.Error(logrus.Fields{
				"phase": "connection",
				"event": "poolerr",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Could not get connection from pool %s, retry attempt %d. Error: %v", c.host, i, redisErr))
			c.Logger.Info(logrus.Fields{
				"phase": "connection",
				"event": "sleep",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Sleeping %d milliseconds before requesting new connection from pool: %s", c.retrySleep, c.host))
			time.Sleep(time.Duration(c.retrySleep) * time.Millisecond)
			conn, redisErr = c.pool.Get()
			defer c.pool.CarefullyPut(conn, &redisErr)
		} else {
			break
		}
	}
	if i == c.retryTimes {
		c.Logger.Error(logrus.Fields{
			"phase": "connection",
			"event": "poolerr",
			"tag":   "redis",
			"rpc":   "NewPool"},
			fmt.Sprintf("Could not get connection from pool upon retrys attempt %d. Exiting. Error: %v", i, redisErr))
		return nil, redisErr
	}

	reply := conn.Cmd(command, args...)
	err = reply.Err
	for i = 1; i <= c.retryTimes; i++ {
		if _, ok := (err).(*redis.CmdError); err != nil && !ok {
			if i >= c.retryTimes {
				c.Logger.Error(logrus.Fields{
					"phase": "connection",
					"event": "poolerr",
					"tag":   "redis",
					"rpc":   "NewPool"},
					fmt.Sprintf("Could not get connection from pool upon %d retry attempts. Exiting. Error: %v", i, err))
				return nil, reply.Err
			}
			c.Logger.Error(logrus.Fields{
				"phase": "connection",
				"event": "poolerr",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Problem running command %v, retry attempt %d for pool: %s Error: %v", command, i, c.host, err))
			c.Logger.Info(logrus.Fields{
				"phase": "connection",
				"event": "sleep",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Sleeping %d milliseconds before requesting new connection from pool: %s", c.retrySleep, c.host))
			time.Sleep(time.Duration(c.retrySleep) * time.Millisecond)
			conn, err = c.pool.Get()
			defer c.pool.CarefullyPut(conn, &err)
			reply := conn.Cmd(command, args...)
			err = reply.Err
		} else {
			break
		}
	}
	c.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "command",
		"tag":   "redis",
		"rpc":   "runCommandPipeline"},
		"Redis command successfully executed.")
	return reply, err
}

func (c *RedisHandler) runCommandPipeline(commands ...string) ([]*redis.Reply, error) {
	var i int
	var err error
	var replies []*redis.Reply

	done := false

	conn, redisErr := c.pool.Get()
	defer c.pool.CarefullyPut(conn, &redisErr)
	for i = 1; i <= c.retryTimes; i++ {
		if redisErr != nil || conn == nil {
			c.Logger.Error(logrus.Fields{
				"phase": "connection",
				"event": "poolerr",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Could not get connection from pool %s, retry attempt %d. Error: %v", c.host, i, redisErr))
			c.Logger.Info(logrus.Fields{
				"phase": "connection",
				"event": "sleep",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Sleeping %d milliseconds before requesting new connection from pool: %s", c.retrySleep, c.host))
			time.Sleep(time.Duration(c.retrySleep) * time.Millisecond)
			conn, redisErr := c.pool.Get()
			defer c.pool.CarefullyPut(conn, &redisErr)
		} else {
			break
		}
		if i == c.retryTimes {
			c.Logger.Error(logrus.Fields{
				"phase": "connection",
				"event": "poolerr",
				"tag":   "redis",
				"rpc":   "NewPool"},
				fmt.Sprintf("Could not get connection from pool %s upon %d retry attempts. Exiting. Error: %v", c.host, i, redisErr))
			return nil, redisErr
		}
	}

	for _, command := range commands {
		var interfaceList []interface{}
		splitString := strings.Split(command, " ")

		for _, arg := range splitString[1:] {
			var interf interface{} = arg
			interfaceList = append(interfaceList, interf)
		}
		conn.Append(splitString[0], interfaceList...)
	}

	for done == false {
		reply := conn.GetReply()
		for i = 1; i <= c.retryTimes; i++ {
			err = reply.Err
			if _, ok := (err).(*redis.CmdError); err != nil && !ok {
				if reply.Err == redis.PipelineQueueEmptyError {
					done = true
					break
				}
				if i >= c.retryTimes {
					c.Logger.Error(logrus.Fields{
						"phase": "connection",
						"event": "poolerr",
						"tag":   "redis",
						"rpc":   "NewPool"},
						fmt.Sprintf("Could not get connection from pool %s upon %d retry attempts. Exiting. Error: %v", c.host, i, err))
					return nil, err
				}
				c.Logger.Error(logrus.Fields{
					"phase": "connection",
					"event": "poolerr",
					"tag":   "redis",
					"rpc":   "NewPool"},
					fmt.Sprintf("Problem running command, retry attempt %d for pool: %s Error: %v", i, c.host, err))
				c.Logger.Info(logrus.Fields{
					"phase": "connection",
					"event": "sleep",
					"tag":   "redis",
					"rpc":   "NewPool"},
					fmt.Sprintf("Sleeping %d milliseconds before requesting new connection from pool: %s", c.retrySleep, c.host))
				time.Sleep(time.Duration(c.retrySleep) * time.Millisecond)
				conn, err := c.pool.Get()
				defer c.pool.CarefullyPut(conn, &err)
			} else {
				replies = append(replies, reply)
				break
			}
		}
	}
	c.Logger.Info(logrus.Fields{
		"phase": "process",
		"event": "command",
		"tag":   "redis",
		"rpc":   "runCommandPipeline"},
		fmt.Sprintf("%d Redis commands successfully executed.", len(commands)))

	return replies, nil
}

func (c *RedisHandler) Exists(keys ...string) (*redis.Reply, error) {
	var interfaceList []interface{}
	for _, key := range keys {
		var interf interface{} = key
		interfaceList = append(interfaceList, interf)
	}
	return c.runCommand("EXISTS", interfaceList...)
}

func (c *RedisHandler) Del(keys ...string) (*redis.Reply, error) {
	var interfaceList []interface{}
	for _, key := range keys {
		var interf interface{} = key
		interfaceList = append(interfaceList, interf)
	}
	return c.runCommand("DEL", interfaceList...)
}

func (c *RedisHandler) Expire(ttl int, keys ...string) ([]*redis.Reply, error) {
	var commandList []string
	for _, key := range keys {
		command := strings.Join([]string{"EXPIRE", strconv.Itoa(ttl), key}, " ")
		commandList = append(commandList, command)
	}
	return c.runCommandPipeline(commandList...)
}

func (c *RedisHandler) Get(keys ...string) (*redis.Reply, error) {
	var interfaceList []interface{}
	for _, key := range keys {
		var interf interface{} = key
		interfaceList = append(interfaceList, interf)
	}

	if len(keys) == 1 {
		return c.runCommand("GET", interfaceList...)
	}
	return c.runCommand("MGET", interfaceList...)
}

func (c *RedisHandler) SetMany(ttl int, tuples ...[]string) ([]*redis.Reply, error) {
	var commandList []string
	for _, tuple := range tuples {
		command := strings.Join([]string{"SET", tuple[0], tuple[1], "EX", strconv.Itoa(ttl)}, " ")
		commandList = append(commandList, command)
	}
	return c.runCommandPipeline(commandList...)
}

func (c *RedisHandler) Set(ttl int, key string, value interface{}) (*redis.Reply, error) {
	return c.runCommand("SET", key, value, "EX", ttl)
}
