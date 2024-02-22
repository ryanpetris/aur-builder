package pacman

import (
	"errors"
	"path"
	"regexp"
	"strings"
)

type Source struct {
	Folder        string
	VcsType       string
	Url           string
	FragmentType  string
	FragmentValue string
	Original      string
}

func (source *Source) String() string {
	builder := strings.Builder{}

	if source.Folder != "" {
		builder.WriteString(source.Folder)
		builder.WriteString("::")
	}

	if source.VcsType != "" {
		builder.WriteString(source.VcsType)
		builder.WriteString("+")
	}

	builder.WriteString(source.Url)

	if source.FragmentType != "" || source.FragmentValue != "" {
		builder.WriteString("#")
		builder.WriteString(source.FragmentType)
		builder.WriteString("=")
		builder.WriteString(source.FragmentValue)
	}

	return builder.String()
}

func (source *Source) GetFolder() string {
	if source.Folder != "" {
		return source.Folder
	}

	return strings.SplitN(path.Base(source.Url), ".", 2)[0]
}

var (
	sourceRegex = regexp.MustCompile(`^((?P<folder>[^:]+)::)?((?P<vcs>[^+]+)\+)?(?P<url>[^#]+)(#(?P<ftype>[^=]+)=(?P<fvalue>.*))?$`)
)

func ParseSource(source string) (*Source, error) {
	groupNames := sourceRegex.SubexpNames()
	result := &Source{}

	for id, value := range sourceRegex.FindStringSubmatch(source) {
		if name := groupNames[id]; name != "" {
			switch name {
			case "folder":
				result.Folder = value

			case "vcs":
				result.VcsType = value

			case "url":
				result.Url = value

			case "ftype":
				result.FragmentType = value

			case "fvalue":
				result.FragmentValue = value

			}
		}
	}

	if result.Url == "" {
		return nil, errors.New("could not parse source properly")
	}

	result.Original = source

	return result, nil
}
