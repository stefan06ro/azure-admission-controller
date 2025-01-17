package project

var (
	description = "Azure admission controller"
	gitSHA      = "n/a"
	name        = "azure-admission-controller"
	source      = "https://github.com/giantswarm/azure-admission-controller"
	version     = "3.2.1-dev"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
