package cienv

func FindCiEnv() CiEnv {
	if fjenv := (ForgejoCiEnv{}); fjenv.IsCI() {
		return fjenv
	}

	if ghenv := (GithubCiEnv{}); ghenv.IsCI() {
		return ghenv
	}

	return DefaultCiEnv{}
}
