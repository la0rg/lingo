package model

var Languages = LanguageSelector{
	Source: []Language{
		{CountryCode: "en", Flag: "us", Name: "English", Active: true},
		{CountryCode: "de", Flag: "de", Name: "German"},
		{CountryCode: "ru", Flag: "ru", Name: "Russian"},
	},
	Target: []Language{
		{CountryCode: "en", Flag: "us", Name: "English"},
		{CountryCode: "de", Flag: "de", Name: "German", Active: true},
		{CountryCode: "ru", Flag: "ru", Name: "Russian"},
	},
}

type LanguageSelector struct {
	Source []Language
	Target []Language
}

type Language struct {
	CountryCode string
	Flag        string
	Name        string
	Active      bool
}
