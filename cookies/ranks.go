package cookies

type Rank uint8

const (
	RankNone Rank = iota
	RankNovice
	RankAdept
	RankPro
	RankMaster
)

func (r Rank) String() string {
	switch r {
	case RankNovice:
		return "Novice"
	case RankAdept:
		return "Adept"
	case RankPro:
		return "Pro"
	case RankMaster:
		return "Master"
	default:
		return "None"
	}
}

func (r Rank) Descriptor() string {
	switch r {
	case RankMaster:
		return "Cookie Factory"
	default:
		return "User"
	}
}

func GetRank(cookies int) Rank {
	switch {
	case cookies < 20:
		return RankNone
	case cookies < 1000:
		return RankNovice
	case cookies < 5000:
		return RankAdept
	case cookies > 9001:
		return RankPro
	default:
		return RankNone
	}
}
