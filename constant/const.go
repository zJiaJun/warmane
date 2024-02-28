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
	BaseUrl    = "https://www.warmane.com"
	AccountUrl = BaseUrl + "/account"
	LoginUrl   = AccountUrl + "/login"
	TradeUrl   = AccountUrl + "/trade"
	LogoutUrl  = AccountUrl + "/logout"
)
const (
	CsrfTokenSelector = "meta[name='csrf-token']"
	CoinsSelector     = ".myCoins"
	PointsSelector    = ".myPoints"
)

var CookieFileName = func(name string) string {
	return name + ".cookies"
}

var CookieKeys = [6]string{
	"bb_lastvisit",
	"bb_lastactivity",
	"bb_sessionhash",
	"bb_userid",
	"bb_password",
	"_wM",
}
