/*
// ----------------------------------------------------------------------------
// model.go
// Countertop Profile Models

// Created by Paul Pietkiewicz on 10/21/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package profile

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm"
)

type User struct {
	ID            uint   `gorm:"primary_key"`
	UUID          string `sql:"unique;not null;index"`
	DeviceId      string `sql:"unique;not null;index"`
	UserId        string `sql:"unique;default: null;index"`
	Firstname     string
	Birthyear     int32
	Gender        int32
	Heightcm      float32
	Weightkg      float32
	Goalweightkg  float32
	Activitylevel int32
	Mealplan      int32
	Weightgoal    int32
	Omnivore      bool `sql:"default: 1"`
	Vegetarian    bool `sql:"default: 0"`
	Vegan         bool `sql:"default: 0"`
	Raw           bool `sql:"default: 0"`
	Glutenfree    bool `sql:"default: 0"`
	Nutfree       bool `sql:"default: 0"`
	Dairyfree     bool `sql:"default: 0"`
	Soyfree       bool `sql:"default: 0"`
	Lowsodium     bool `sql:"default: 0"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
