package main

import "github.com/open-spaced-repetition/go-fsrs/v3"

func StateToString(state fsrs.State) string {
	switch state {
	case fsrs.New:
		return "New"
	case fsrs.Learning:
		return "Learning"
	case fsrs.Review:
		return "Review"
	case fsrs.Relearning:
		return "Relearning"
	default:
		return "Unknown"
	}
}

func StringToState(s string) fsrs.State {
	switch s {
	case "New":
		return fsrs.New
	case "Learning":
		return fsrs.Learning
	case "Review":
		return fsrs.Review
	case "Relearning":
		return fsrs.Relearning
	default:
		return fsrs.New
	}
}