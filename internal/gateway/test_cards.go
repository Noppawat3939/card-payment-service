package gateway

type CardScenario string

const (
	CardScenarioSuccess           CardScenario = "success"
	CardScenarioDeclined          CardScenario = "declined"
	CardScenarioInsufficientFunds CardScenario = "insufficient_funds"
	CardScenarioExpired           CardScenario = "expired"
)

var testCardRegistry = map[string]CardScenario{
	"4242424242424242": CardScenarioSuccess,
	"4000000000000002": CardScenarioDeclined,
	"4000000000009995": CardScenarioInsufficientFunds,
	"4000000000000069": CardScenarioExpired,
}

func GetTestCardScenario(cardNumber string) (CardScenario, bool) {
	scenario, ok := testCardRegistry[cardNumber]
	return scenario, ok
}
