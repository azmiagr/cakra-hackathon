package main

import (
	"github.com/azmiagr/cakra-hackathon/internal/handler/rest"
	"github.com/azmiagr/cakra-hackathon/internal/repository"
	"github.com/azmiagr/cakra-hackathon/internal/service"
	"github.com/azmiagr/cakra-hackathon/pkg/bcrypt"
	"github.com/azmiagr/cakra-hackathon/pkg/config"
	"github.com/azmiagr/cakra-hackathon/pkg/database/mariadb"
	"github.com/azmiagr/cakra-hackathon/pkg/jwt"
	"github.com/azmiagr/cakra-hackathon/pkg/mail"
	"github.com/azmiagr/cakra-hackathon/pkg/middleware"
	"log"
)

func main() {
	config.LoadEnvironment()

	db, err := mariadb.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = mariadb.Migrate(db)
	if err != nil {
		log.Fatal(err)
	}
	if err := mariadb.Seed(db); err != nil {
		log.Fatal(err)
	}

	registrationConfig, err := config.LoadRegistrationConfig()
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	bcryptAuth := bcrypt.Init()
	jwtAuth := jwt.Init()
	mailer := mail.Init()
	svc := service.NewService(db, repo, bcryptAuth, jwtAuth, mailer, registrationConfig)

	middleware := middleware.Init(svc, jwtAuth)
	r := rest.NewRest(svc, middleware)
	r.MountEndpoint()

	r.Run()
}
