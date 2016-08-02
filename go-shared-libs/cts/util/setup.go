package util

import (
	"fmt"
	"os"
	"strings"

	//"github.com/fzzy/radix/extra/sentinel"
	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	pb "github.com/theorangechefco/cts/go-protos"
	loglib "github.com/theorangechefco/cts/go-shared-libs/cts/logger"
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
}

func SetupRedisAccount(redisHost string, redisPass string) (*pb.SessionToken, error) {
	token := pb.SessionToken{
		Id: "goXycIUTGQyKJXSxbOJY",
	}

	userID := pb.UserId{
		Uuid: "d17eaf65-244a-4913-83d8-1583bb3cbbfd",
	}

	userKey := strings.Join([]string{"user", userID.Uuid}, "_")
	sessionTokenKey := strings.Join([]string{"token", token.Id}, "_")
	logger := loglib.NewLogger("setup", "", 8080, true, logrus.DebugLevel)

	redisClient, err := NewClient(logger, redisHost, redisPass)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	exists, err := CheckForRedisKey(redisClient, token.Id)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	if !exists {
		err = CreateRedisEntry(redisClient, userKey, token.Id)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		err = CreateRedisEntry(redisClient, sessionTokenKey, userID.Uuid)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	return &token, nil
}

func SetupUser(dbHost string, dbPort int, dbUser string, dbPass string, dbName string) (*pb.UserId, error) {

	identifier := pb.Identifier{
		Deviceidentifier: "ecb7381b-566c-4e53-b3ea-add8cd372d6a",
	}
	profile := pb.Profile{
		Identifier:    &identifier,
		Firstname:     "John Doe",
		Birthyear:     1985,
		Gender:        pb.Gender_MALE,
		Heightcm:      185,
		Weightkg:      85,
		Goalweightkg:  80,
		Activitylevel: pb.ActivityLevel_SEDENTARY,
		Mealplan:      pb.MealPlan_EIGHTEEN_HUNDRED,
		Weightgoal:    pb.WeightGoal_LOSE,
		Dietaryprofile: &pb.DietaryProfile{
			Omnivore: true,
		},
		Dietaryrestriction: &pb.DietaryRestriction{},
	}

	userID := pb.UserId{
		Uuid: "d17eaf65-244a-4913-83d8-1583bb3cbbfd",
	}

	user := User{
		DeviceId:      profile.Identifier.Deviceidentifier,
		UserId:        profile.Identifier.Useridentifier,
		UUID:          userID.Uuid,
		Firstname:     profile.Firstname,
		Birthyear:     profile.Birthyear,
		Gender:        int32(profile.Gender),
		Heightcm:      profile.Heightcm,
		Weightkg:      profile.Weightkg,
		Goalweightkg:  profile.Goalweightkg,
		Activitylevel: int32(profile.Activitylevel),
		Mealplan:      int32(profile.Mealplan),
		Weightgoal:    int32(profile.Weightgoal),
		Omnivore:      profile.Dietaryprofile.Omnivore,
		Vegetarian:    profile.Dietaryprofile.Vegetarian,
		Vegan:         profile.Dietaryprofile.Vegan,
		Raw:           profile.Dietaryprofile.Raw,
		Glutenfree:    profile.Dietaryrestriction.Glutenfree,
		Nutfree:       profile.Dietaryrestriction.Nutfree,
		Dairyfree:     profile.Dietaryrestriction.Dairyfree,
		Soyfree:       profile.Dietaryrestriction.Soyfree,
		Lowsodium:     profile.Dietaryrestriction.Lowsodium,
	}

	connStr := fmt.Sprintf("%s:%s@tcp([%s]:%d)/%s?charset=utf8", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := gorm.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	err = db.DB().Ping()
	if err != nil {
		return nil, err
	}

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.SingularTable(true)

	query := db.Create(&user)
	if query.Error != nil {
		fmt.Println(err)
	}

	return &userID, nil
}
