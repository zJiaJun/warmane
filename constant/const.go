package constant

const (
	/*
			{"messages":{"errors":["Incorrect account name or password."]}}
			{"messages":{"errors":["You have already collected your points today."]}}
		 	{"messages":{"errors":["You have not logged in-game today."]}}
			{"messages":{"success":["Daily points collected."]},"points":[10.4]}
	*/
	LoginSuccessBody = "{\"redirect\":[\"\\/account\"]}"
)

const (
	HOST              = "www.warmane.com"
	BaseUrl           = "https://" + HOST
	AccountUrl        = BaseUrl + "/account"
	AuthenticationUrl = AccountUrl + "/authentication"
	LoginUrl          = AccountUrl + "/login"
	TradeUrl          = AccountUrl + "/trade"
	LogoutUrl         = AccountUrl + "/logout"
)

const WarmaneSiteKey = "6LfXRRsUAAAAAEApnVwrtQ7aFprn4naEcc05AZUR"
const captchaApiKey = ""

const (
	ONLINE_STATUS  = "online"
	OFFLINE_STATUS = "offline"
)

var CookieFileName = func(name string) string {
	return name + ".cookies"
}

var CookieKeys = [7]string{
	"PHPSESSID",
	"bb_lastvisit",
	"bb_lastactivity",
	"bb_sessionhash",
	"bb_userid",
	"bb_password",
	"_wM",
}
