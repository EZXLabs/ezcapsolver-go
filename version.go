package ezcapsolver

// Version is the release this build corresponds to.
//
// It ends up in the User-Agent header, so it has to stay in step with the
// published git tag. TestVersionMatchesUserAgent guards the format.
const Version = "0.1.0"

// defaultUserAgent identifies the SDK and its version to the service, in the
// `ezcapsolver-<language>/<version>` form every language SDK uses.
const defaultUserAgent = "ezcapsolver-go/" + Version
