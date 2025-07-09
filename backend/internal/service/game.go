package service

func checkWin(Board [3][3]string) string {
	// Check rows and columns
	for i := 0; i < 3; i++ {
		if Board[i][0] != "" && Board[i][0] == Board[i][1] && Board[i][1] == Board[i][2] {
			return Board[i][0]
		}
		if Board[0][i] != "" && Board[0][i] == Board[1][i] && Board[1][i] == Board[2][i] {
			return Board[0][i]
		}
	}

	// Check diagonals
	if Board[0][0] != "" && Board[0][0] == Board[1][1] && Board[1][1] == Board[2][2] {
		return Board[0][0]
	}
	if Board[0][2] != "" && Board[0][2] == Board[1][1] && Board[1][1] == Board[2][0] {
		return Board[0][2]
	}

	// Check if there are any empty cells
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if Board[i][j] == "" {
				return "" // Game not finished
			}
		}
	}

	return "draw" // All cells filled, no winner
}

func ChangeTurn(role string) string{
	println(role,"x")
	if role == "x"{
		return "o"
	}
	return "x"
}

