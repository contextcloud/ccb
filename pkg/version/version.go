package version

const UserAgent = "Context Cloud CLI"

var (
	Version   string
	GitCommit string
)

func BuildVersion() string {
	if len(Version) == 0 {
		return "dev"
	}
	return Version
}
