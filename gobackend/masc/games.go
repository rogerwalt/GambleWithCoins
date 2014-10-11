package masc

func PrisonersDilemma(action1 string, action2 string) (int, int) {
	switch {
	case action1 == "C" && action2 == "C":
		return -1, -1
	case action1 == "C" && action2 == "D":
		return 0, 3
	case action1 == "D" && action2 == "C":
		return 3, 0
	case action1 == "D" && action2 == "D":
		return 2, 2
	default:
		return 0, 0
	}
}
