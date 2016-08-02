/*
// ----------------------------------------------------------------------------
// logger.go
// Countertop Logging Framework

// Created by Kaushal Kantawala on 9/10/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package logger

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_fluent"
	//https://github.com/aybabtme/grpclogrus.git
)

type CtsLogger struct {
	logger *logrus.Logger
}

func NewLogger(serviceName string, fluentdHost string, fluentdPort int, toStdErr bool, level logrus.Level) *CtsLogger {
	l := new(CtsLogger)
	l.logger = logrus.New()

	if fluentdHost != "" {
		hook := logrus_fluent.NewHook("localhost", 24224)
		hook.SetLevels([]logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
		})
		l.logger.Hooks.Add(hook)
	}

	if toStdErr {
		l.logger.Out = os.Stderr
	}
	l.logger.Level = level
	return l
}

func (l *CtsLogger) Panic(fields logrus.Fields, message string) {
	l.logger.WithFields(fields).Error(message)
}

func (l *CtsLogger) Fatal(fields logrus.Fields, message string) {
	l.logger.WithFields(fields).Fatal(message)
}

func (l *CtsLogger) Error(fields logrus.Fields, message string) {
	l.logger.WithFields(fields).Error(message)
}

func (l *CtsLogger) Warn(fields logrus.Fields, message string) {
	l.logger.WithFields(fields).Warn(message)
}

func (l *CtsLogger) Info(fields logrus.Fields, message string) {
	l.logger.WithFields(fields).Info(message)
}

func (l *CtsLogger) Debug(fields logrus.Fields, message string) {
	l.logger.WithFields(fields).Debug(message)
}
