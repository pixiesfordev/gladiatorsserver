module gladiatorsClient

go 1.22.0

replace gladiatorsGoModule => ../gladiatorsGoModule // for local

// replace gladiatorsGoModule => /home/gladiatorsGoModule // for docker

require (
	github.com/sirupsen/logrus v1.9.3
	gladiatorsGoModule v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.13.0 // indirect
