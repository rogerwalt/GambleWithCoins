package masc

func PrisonersDilemma(action1 string, action2 string) (int, int) {
	switch {
	case action1 == "cooperate" && action2 == "cooperate":
		return -1, -1
	case action1 == "cooperate" && action2 == "defect":
		return 0, 3
	case action1 == "defect" && action2 == "cooperate":
		return 3, 0
	case action1 == "defect" && action2 == "defect":
		return 2, 2
	default:
		return 0, 0
	}
}
