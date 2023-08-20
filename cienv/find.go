package cienv

func FindCiEnv() CiEnv {
	if ghenv := (GithubCiEnv{}); ghenv.IsCI() {
		return ghenv
	}

	return DefaultCiEnv{}
}
