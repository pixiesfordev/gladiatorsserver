module crontasker

go 1.22.3

replace gladiatorsGoModule => ../gladiatorsGoModule // for local

// replace gladiatorsGoModule => /home/gladiatorsGoModule // for docker

require github.com/sirupsen/logrus v1.9.3

require (
	github.com/robfig/cron/v3 v3.0.1 // indirect
	gladiatorsGoModule v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.12.1
	golang.org/x/sys v0.13.0 // indirect
)
