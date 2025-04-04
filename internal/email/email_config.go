package email

type EmailConfig struct {
	AppName               string
	AppDescription        string
	AppEmailContact       string
	StaticURL             string
	LegalNoticeURL        string
	TermsAndConditionsURL string
	PrivacyAndCookiesURL  string
}

// TODO: Correct the URLs and email
var DefaultEmailConfig = EmailConfig{
	AppName:               "KTHAIS",
	AppDescription:        "Welcome to KTHAIS",
	AppEmailContact:       "jack@gugolz.se",
	StaticURL:             "https://kthais.com/static",
	LegalNoticeURL:        "https://kthais.com/page/legal/legal-notice/",
	TermsAndConditionsURL: "https://kthais.com/page/legal/terms-and-conditions/",
	PrivacyAndCookiesURL:  "https://kthais.com/page/legal/privacy-and-cookies/",
}
