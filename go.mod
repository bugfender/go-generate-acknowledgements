module github.com/bugfender/go-generate-acknowledgements

go 1.22.1

require (
	github.com/google/licenseclassifier v0.0.0-20221004142553-c1ed8fcf4bab
	github.com/rogpeppe/go-internal v1.12.0
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/mod v0.17.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/sergi/go-diff v1.3.1 // indirect
	golang.org/x/sys v0.19.0 // indirect
)

// licenseclassifier fork contains changes embed the license files from the repo
replace github.com/google/licenseclassifier => github.com/bugfender-contrib/licenseclassifier v0.0.0-20240405154127-80acd0b8f5ee
