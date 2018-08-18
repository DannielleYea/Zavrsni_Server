package main

func writeAMove(game *game, move GameMove) {

	row, column := calculateMatrixField(move.Feield - 1)
	game.board[row][column] = move.Player
}

func checkGame(game *game, move *GameMove) {

	if game.board[0][0] == move.Player && game.board[0][1] == move.Player && game.board[0][2] == move.Player {
		move.Winner = move.Player
	} else if game.board[0][0] == move.Player && game.board[1][0] == move.Player && game.board[2][0] == move.Player {
		move.Winner = move.Player
	} else if game.board[0][0] == move.Player && game.board[1][1] == move.Player && game.board[2][2] == move.Player {
		move.Winner = move.Player
	} else if game.board[1][0] == move.Player && game.board[1][1] == move.Player && game.board[1][2] == move.Player {
		move.Winner = move.Player
	} else if game.board[0][2] == move.Player && game.board[1][2] == move.Player && game.board[2][2] == move.Player {
		move.Winner = move.Player
	} else if game.board[2][0] == move.Player && game.board[1][1] == move.Player && game.board[0][2] == move.Player {
		move.Winner = move.Player
	} else if game.board[0][1] == move.Player && game.board[1][1] == move.Player && game.board[2][1] == move.Player {
		move.Winner = move.Player
	} else if game.board[2][0] == move.Player && game.board[2][1] == move.Player && game.board[2][2] == move.Player {
		move.Winner = move.Player
	} else if move.Turn == 9 {
		move.Winner = 3
		db.Query("UPDATE game SET turns=9, draw=1 WHERE game_id='" + move.GameId + "');")
	}
}

func calculateMatrixField(field int) (int, int) {
	row := field % 3
	column := int(field / 3)

	return row, column
}
