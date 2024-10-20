package model

var languages = []Language{
	{CountryCode: "en", Flag: "us", Name: "English"},
	{CountryCode: "de", Flag: "de", Name: "German"},
	{CountryCode: "ru", Flag: "ru", Name: "Russian"},
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

func NewLanguageSelector(source, target string) LanguageSelector {
	return LanguageSelector{
		SourceCountryCode: source,
		TargetCountryCode: target,
		Languages:         languages,
	}
}

func UpdateLanguageSelector(sourceActive, targetActive, sourceNew, targetNew string) LanguageSelector {
	if targetNew == sourceActive || sourceNew == targetActive {
		// swap source and target (don't allow the same language for source and target)
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
