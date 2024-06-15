module lobby

go 1.22.0

replace gladiatorsGoModule => ../gladiatorsGoModule // for local

// replace gladiatorsGoModule => /home/gladiatorsGoModule // for docker

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
)
