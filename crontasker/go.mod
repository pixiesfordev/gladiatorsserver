module crontasker

go 1.21.1

replace gladiatorsGoModule => ../gladiatorsGoModule // for local

// replace gladiatorsGoModule => /home/gladiatorsGoModule // for docker

require github.com/robfig/cron/v3 v3.0.1 // indirect
