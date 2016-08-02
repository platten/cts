/*
// ----------------------------------------------------------------------------
// bootstrap.go
// Countertop Profile Microservice DB Bootstraping & Migration Tool

// Created by Paul Pietkiewicz on 9/2/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	dbHost = flag.String("db_host", "admin:admin@tcp([:::::::]:3306)/profile?charset=utf8", "DB connection string")
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

func main() {
	flag.Parse()

	db, err := gorm.Open("mysql", *dbHost)
	if err != nil {
		fmt.Printf("cannot connect to db: %v\n", err)
		os.Exit(-1)
	}
	fmt.Printf("Connecting to %s\n", *dbHost)
	fmt.Println("Pausing for 10 seconds...")
	time.Sleep(time.Duration(10) * time.Second)
	fmt.Println("Running migration...")
	db.SingularTable(true)
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&User{})
	fmt.Println("Database migration complete!")
}
