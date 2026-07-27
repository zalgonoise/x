package politebot

type Rank uint8

const (
	RankNone Rank = iota
	RankKind
	RankChill
	RankCheeky
	RankCocky
	RankFeral
	RankReeeeee
	RankRagelord
	RankCrashingOut
	RankMalding
	RankMaster
)

func (r Rank) String() string {
	switch r {
	case RankKind:
		return "Kind"
	case RankChill:
		return "Chill"
	case RankCheeky:
		return "Cheeky"
	case RankCocky:
		return "Cocky"
	case RankFeral:
		return "Feral"
	case RankReeeeee:
		return "REEEEEE"
	case RankRagelord:
		return "Ragelord"
	case RankCrashingOut:
		return "Crashing tf out"
	case RankMalding:
		return "ABSOLULY MALDING"
	default:
		return "None"
	}
}

func (r Rank) Descriptor() string {
	switch r {
	case RankMaster:
		return "Necrosympathy"
	default:
		return "Polite"
	}
}

func GetRank(angyPoints int) Rank {
	switch {
	case angyPoints < 0:
		return RankNone
	case angyPoints < 5:
		return RankKind
	case angyPoints < 10:
		return RankChill
	case angyPoints < 15:
		return RankCheeky
	case angyPoints < 20:
		return RankCocky
	case angyPoints < 25:
		return RankFeral
	case angyPoints < 30:
		return RankReeeeee
	case angyPoints < 35:
		return RankRagelord
	case angyPoints < 40:
		return RankCrashingOut
	default:
		return RankMalding
	}
}
