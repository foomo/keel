package keel

// Build information, intended to be set at link time.
var (
	// Version is the build version, set with:
	//
	//	-ldflags "-X github.com/foomo/keel.Version=$VERSION"
	Version string
	// GitCommit is the build commit, set with:
	//
	//	-ldflags "-X github.com/foomo/keel.GitCommit=$GIT_COMMIT"
	GitCommit string
	// BuildTime is the build time, set with:
	//
	//	-ldflags "-X 'github.com/foomo/keel.BuildTime=$(date -u '+%Y-%m-%d %H:%M:%S')'"
	BuildTime string
)
