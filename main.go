package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/notnil/chess"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	AISIDE             = "aiside"
	AGAINST_RANDOM_CPU = "againstRandomCPU"
)

var randomizer *rand.Rand

func init() {
	flag.String(AISIDE, "white", "which side of the game the AI will play")
	flag.Bool(AGAINST_RANDOM_CPU, false, "set to true in order for the AI to play against an automated player choosing random moves")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	randSource := rand.NewSource(time.Now().UnixNano())
	randomizer = rand.New(randSource)
}

func main() {
	game := chess.NewGame()
	PrintBoard(game)

	// generate moves until game is over
	for game.Outcome() == chess.NoOutcome {
		fmt.Println("\n# White's turn")
		if viper.GetString(AISIDE) == "white" {
			if err := PlayAI(game); err != nil {
				fmt.Println(err)
				break
			}
		} else {
			if err := PlayRandomOrHuman(game); err != nil {
				fmt.Println(err)
				break
			}
		}

		fmt.Println("\n# Black's turn")
		if viper.GetString(AISIDE) == "black" {
			if err := PlayAI(game); err != nil {
				fmt.Println(err)
				break
			}
		} else {
			if err := PlayRandomOrHuman(game); err != nil {
				fmt.Println(err)
				break
			}
		}
	}

	fmt.Printf("The game finished. Outcome: %s. Method: %s.\n", game.Outcome(), game.Method())
	fmt.Println("PGN:", game.String())
}

func PlayAI(game *chess.Game) error {
	fmt.Println("# AI player")
	gameTreeRootNode := NewGameTreeNode(game)
	t1 := time.Now()
	BuildGameTreeAt(gameTreeRootNode, 1)
	fmt.Println("Time spent building game tree", time.Since(t1))
	t2 := time.Now()
	bestGame := AlphaBeta(gameTreeRootNode, 5, -1000000, 1000000, true)
	fmt.Println("Time spent during AlphaBeta", time.Since(t2))
	if bestGame == nil || bestGame.Game == nil {
		return fmt.Errorf("it seems that there is no best game to choose")
	}
	moveHist := bestGame.Game.MoveHistory()
	offset := 2
	if viper.GetString(AISIDE) == "black" {
		offset = 1
	}
	aiMove := moveHist[len(moveHist)-offset].Move
	game.Move(aiMove)
	PrintBoard(game)
	return nil
}

func PlayRandomOrHuman(game *chess.Game) error {
	if viper.GetBool(AGAINST_RANDOM_CPU) {
		fmt.Println("# Random player")
		if err := MoveRandom(game); err != nil {
			return err
		}
	} else {
		fmt.Println("# Human player")
		for {
			moveStr := ReadMove()
			if moveStr == "r" {
				if err := MoveRandom(game); err != nil {
					fmt.Println(err)
					continue
				} else {
					break
				}
			} else {
				if err := game.MoveStr(moveStr); err != nil {
					fmt.Printf("Invalid move provided, %s. It should be like, 'd3f5' or 'Qf5': %s\n", moveStr, err)
					continue
				}
				break
			}
		}
	}
	PrintBoard(game)
	return nil
}

func PrintBoard(game *chess.Game) {
	fmt.Println(game.Position().Board().Draw())
	fmt.Println("Board evaluation: ", EvaluateStrongerSide(game))
	fmt.Println("Current FEN:", game.FEN())
}

func ReadMove() string {
	fmt.Print("Enter the move > ")
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
	move := moves[randomizer.Intn(lenMoves)]
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
