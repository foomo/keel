package keel

var (
	// Version usage -ldflags "-X github.com/foomo/keel/server.Version=$VERSION"
	Version string
	// GitCommit usage -ldflags "-X github.com/foomo/keel/server.GitCommit=$GIT_COMMIT"
	GitCommit string
	// BuildTime usage -ldflags "-X 'github.com/foomo/keel/server.BuildTime=$(date -u '+%Y-%m-%d %H:%M:%S')'"
	BuildTime string
)
