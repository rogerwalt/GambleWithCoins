package masc

func PrisonersDilemma(action1, action2 string, b, E int) (int, int) {
	switch {
	case action1 == "cooperate" && action2 == "cooperate":
		return E / 2, E / 2
	case action1 == "cooperate" && action2 == "defect":
		return -b, b
	case action1 == "defect" && action2 == "cooperate":
		return b, -b
	case action1 == "defect" && action2 == "defect":
		return 0, 0
	default:
		return 0, 0
	}
}
