module github.com/hermes

go 1.13

require github.com/sirupsen/logrus v1.8.1

require golang.org/x/crypto v0.0.0

replace golang.org/x/crypto => ../../golang.org/x/crypto

require (
	golang.org/x/sys v0.0.0
	golang.org/x/text v0.3.3
)

replace golang.org/x/sys => ../../golang.org/x/sys
