package model

var languages = []Language{
	{CountryCode: "en", Flag: "us", Name: "English"},
	{CountryCode: "de", Flag: "de", Name: "German"},
	{CountryCode: "ru", Flag: "ru", Name: "Russian"},
}

var DefaultLanguageSelector = LanguageSelector{
	SourceCountryCode: "de",
	TargetCountryCode: "en",
	Languages:         languages,
}

type LanguageSelector struct {
	SourceCountryCode string
	TargetCountryCode string
	Languages         []Language
}

type Language struct {
	CountryCode string
	Flag        string
	Name        string
}

func UpdateLanguageSelector(sourceActive, targetActive, sourceNew, targetNew string) LanguageSelector {
	if targetNew == sourceActive || sourceNew == targetActive {
		// swap source and target; do not allow the same language for source and target
		return LanguageSelector{
			SourceCountryCode: targetActive,
			TargetCountryCode: sourceActive,
			Languages:         languages,
		}
	}

	if sourceNew != "" {
		sourceActive = sourceNew
	}
	if targetNew != "" {
		targetActive = targetNew
	}

	return LanguageSelector{
		SourceCountryCode: sourceActive,
		TargetCountryCode: targetActive,
		Languages:         languages,
	}
}
