package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/notnil/chess"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	SIDE               = "side"
	FEN                = "fen"
	AGAINST_RANDOM_CPU = "againstRandomCPU"
)

func init() {
	flag.String(SIDE, "white", "which side of the game the AI will play")
	flag.String(FEN, "", "a FEN string used to initialize the game")
	flag.Bool(AGAINST_RANDOM_CPU, false, "set to true in order for the AI to play against an automated player choosing random moves")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func main() {

	// return

	// fenFn, err := chess.FEN("B5R1/1P2k1P1/P4R2/8/4P2N/2N5/P1P1RP2/K7 w - - 11 70")
	// if err != nil {
	// 	panic(err)
	// }
	// game := chess.NewGame(fenFn)
	game := chess.NewGame()
	PrintBoard(game)

	// generate moves until game is over
	for game.Outcome() == chess.NoOutcome {
		// White's turn
		gameTreeRootNode := NewGameTreeNode(game)
		BuildGameTreeAt(gameTreeRootNode, 1)
		bestGame := AlphaBeta(gameTreeRootNode, 5, -1000000, 1000000, true)
		if bestGame == nil || bestGame.Game == nil {
			fmt.Println("There's no best game?")
			fmt.Println(bestGame)
			return
		}
		moveHist := bestGame.Game.MoveHistory()
		// bestGame.Game.Positions()
		aiMove := moveHist[len(moveHist)-2].Move
		game.Move(aiMove)
		PrintBoard(game)

		// Black's turn
		if viper.GetBool(AGAINST_RANDOM_CPU) {
			if err := MoveRandom(game); err != nil {
				fmt.Println(err)
				break
			}
		} else {
			moveStr := ReadMove()
			if moveStr == "r" {
				if err := MoveRandom(game); err != nil {
					fmt.Println(err)
					break
				}
			} else {
				if game.MoveStr(moveStr) != nil {
					fmt.Println("Invalid move provided. It should be in the following format for example, 'd3f5'.")
					break
				}
			}
		}
		PrintBoard(game)
	}

	fmt.Printf("The game finished. Outcome: %s. Method: %s.\n", game.Outcome(), game.Method())
	fmt.Println("PGN:", game.String())
}

func PrintBoard(game *chess.Game) {
	fmt.Println(game.Position().Board().Draw())
	fmt.Println("Board evaluation: ", EvaluateStrongerSide(game))
	fmt.Println("Current FEN:", game.FEN())
}

func ReadMove() string {
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func MoveRandom(game *chess.Game) error {
	moves := game.ValidMoves()
	lenMoves := len(moves)
	if len(moves) == 0 {
		return fmt.Errorf("there are no valid moves left")
	}
	move := moves[rand.Intn(lenMoves)]
	fmt.Println("Selected random move:", move.String())
	game.Move(move)
	return nil
}

func EvaluateStrongerSide(game *chess.Game) int {
	sm := game.Position().Board().SquareMap()
	score := 0

	for _, piece := range sm {
		strength := 0
		switch piece.Type() {
		case chess.King:
			strength = 900
		case chess.Queen:
			strength = 90
		case chess.Rook:
			strength = 50
		case chess.Bishop:
			strength = 30
		case chess.Knight:
			strength = 30
		case chess.Pawn:
			strength = 10
		}
		if piece.Color() == chess.White {
			score += strength
		} else {
			score -= strength
		}
	}

	return score
}

type GameTreeNode struct {
	Game       *chess.Game
	Evaluation int
	Children   []*GameTreeNode
}

func NewGameTreeNode(game *chess.Game) *GameTreeNode {
	return &GameTreeNode{
		Game:       game,
		Evaluation: EvaluateStrongerSide(game),
	}
}

func CloneGameTreeNode(gameTreeNode *GameTreeNode) *GameTreeNode {
	return &GameTreeNode{
		Game:       gameTreeNode.Game.Clone(),
		Evaluation: gameTreeNode.Evaluation,
		Children:   gameTreeNode.Children,
	}
}

func BuildGameTreeAt(gameTreeRootNode *GameTreeNode, depth int) {
	possibleMoves := gameTreeRootNode.Game.ValidMoves()
	gameTreeRootNode.Children = []*GameTreeNode{}
	// println(">>> possible moves", len(possibleMoves), depth)
	for _, possibleMove := range possibleMoves {
		possibleGame := CloneGameTreeNode(gameTreeRootNode)
		possibleGame.Game.Move(possibleMove)
		possibleGame.Evaluation = EvaluateStrongerSide(possibleGame.Game)
		gameTreeRootNode.Children = append(gameTreeRootNode.Children, possibleGame)
		if depth > 0 {
			BuildGameTreeAt(possibleGame, depth-1)
		}
	}
}

func MaxNode(node1, node2 *GameTreeNode) *GameTreeNode {
	if node1 == nil {
		return node2
	} else if node2 == nil {
		return node1
	}

	if node1.Evaluation > node2.Evaluation {
		return node1
	} else {
		return node2
	}
}

func MinNode(node1, node2 *GameTreeNode) *GameTreeNode {
	if node1 == nil {
		return node2
	} else if node2 == nil {
		return node1
	}

	if node1.Evaluation > node2.Evaluation {
		return node2
	} else {
		return node1
	}
}

func AlphaBeta(node *GameTreeNode, depth, a, b int, maximizingPlayer bool) *GameTreeNode {
	if depth == 0 || node.Children == nil || len(node.Children) == 0 {
		return node
	}

	nodeA := &GameTreeNode{Evaluation: a}
	nodeB := &GameTreeNode{Evaluation: b}

	var value *GameTreeNode
	if maximizingPlayer {
		for _, child := range node.Children {
			value2 := AlphaBeta(child, depth-1, a, b, false)
			value = MaxNode(value, value2)
			if value != nil && value.Evaluation >= b {
				break
			}
			nodeA = MaxNode(nodeA, value)
		}
		return value
	} else {
		for _, child := range node.Children {
			value = MinNode(value, AlphaBeta(child, depth-1, a, b, true))
			if value != nil && value.Evaluation <= a {
				break
			}
			nodeB = MinNode(nodeB, value)
		}
		return value
	}
}
