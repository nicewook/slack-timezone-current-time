module main

go 1.12

require (
	cloud.google.com/go v0.41.0 // indirect
	github.com/nicewook/slack-timezone-current-time/api v0.1.0
	github.com/nlopes/slack v0.5.0
)

replace github.com/nicewook/slack-timezone-current-time/api v0.1.0 => ./api
