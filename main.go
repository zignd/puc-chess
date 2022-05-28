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

// Constantes que representam argumentos de linha de comando que podem
// customizar a forma como o programa funciona
const (
	AISIDE             = "aiside"
	AGAINST_RANDOM_CPU = "againstRandomCPU"
)

var randomizer *rand.Rand

func init() {
	// Registro dos possíveis argumentos de linha de comando aceitos pelo programa,
	// seus valores padrão e uma breve descrição sobre o que cada um faz
	flag.String(AISIDE, "white", "which side of the game the AI will play")
	flag.Bool(AGAINST_RANDOM_CPU, false, "set to true in order for the AI to play against an automated player choosing random moves")

	// Interpretação dos argumentos de linha de comando informados
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// Inicializa um randomizador utilizado para gerar jogas aleatórias no modo AGAINST_RANDOM_CPU
	randSource := rand.NewSource(time.Now().UnixNano())
	randomizer = rand.New(randSource)
}

func main() {
	// Cria um novo tabuleiro com as peças nas posições iniciais
	game := chess.NewGame()
	PrintBoard(game)

	// Continua o jogo até que ele acabe
	for game.Outcome() == chess.NoOutcome {
		// Trecho referente à vez da peça branca
		fmt.Println("\n# White's turn")
		// Verificando se a IA irá jogar do lado das peças brancas
		if viper.GetString(AISIDE) == "white" {
			// Faz a jogada utilizando a IA
			if err := PlayAI(game); err != nil {
				fmt.Println(err)
				break
			}
		} else { // Caso contrário, se as peças brancas serão controladas pelo modo aleatório ou humano
			// Faz a jogada utilizando o modo aleatório ou humano
			if err := PlayRandomOrHuman(game); err != nil {
				fmt.Println(err)
				break
			}
		}

		// Trecho referente à vez da peça preta
		fmt.Println("\n# Black's turn")
		// Verificando se a IA irá jogar do lado das peças pretas
		if viper.GetString(AISIDE) == "black" {
			// Faz a jogada utilizando a IA
			if err := PlayAI(game); err != nil {
				fmt.Println(err)
				break
			}
		} else { // Caso contrário, se as peças brancas serão controladas pelo modo aleatório ou humano
			// Faz a jogada utilizando o modo aleatório ou humano
			if err := PlayRandomOrHuman(game); err != nil {
				fmt.Println(err)
				break
			}
		}
	}

	// Após sair do loop acima o jogo terá terminado, então será exibido aqui o resultado final do jogo
	fmt.Printf("The game finished. Outcome: %s. Method: %s.\n", game.Outcome(), game.Method())
	fmt.Println("PGN:", game.String())
}

// PlayAI, dado um tabuleiro, faz uma jogada utilizando o algoritmo Alfa-Beta
func PlayAI(game *chess.Game) error {
	// Primeiro criamos um nó inicial para a nossa game tree
	fmt.Println("# AI player")
	gameTreeRootNode := NewGameTreeNode(game)
	t1 := time.Now()
	// Constrói uma game tree a partir do nó inicial
	BuildGameTreeAt(gameTreeRootNode, 1)
	fmt.Println("Time spent building game tree", time.Since(t1))
	t2 := time.Now()
	// Utiliza o algoritmo Alfa-Beta para identificar o melhor jogo
	// dentre os que foram gerados na game tree
	bestGame := AlphaBeta(gameTreeRootNode, 5, -1000000, 1000000, true)
	fmt.Println("Time spent during AlphaBeta", time.Since(t2))
	if bestGame == nil || bestGame.Game == nil {
		return fmt.Errorf("it seems that there is no best game to choose")
	}
	// Extrai o histórico de jogadas do melhor jogo obtido pelo Alfa-Beta
	moveHist := bestGame.Game.MoveHistory()
	offset := 2
	if viper.GetString(AISIDE) == "black" {
		offset = 1
	}
	// Extrai a última jogada que deverá ser feita pela IA
	aiMove := moveHist[len(moveHist)-offset].Move
	// Executa a jogada no tabuleiro como a IA
	game.Move(aiMove)
	PrintBoard(game)
	return nil
}

// PlayRandomOrHuman, dado um tabuleiro, faz uma jogada que pode ser aleatória ou por um humano
func PlayRandomOrHuman(game *chess.Game) error {
	// Identifica se a jogada será feita de forma aleatória
	if viper.GetBool(AGAINST_RANDOM_CPU) {
		fmt.Println("# Random player")
		// Executa a jogada aleatória
		if err := MoveRandom(game); err != nil {
			return err
		}
	} else { // Caso contrário a jogada será com base na entrada de um humano
		fmt.Println("# Human player")
		for {
			// Lê o movimento inserido pelo teclado
			moveStr := ReadMove()
			// Para jogadas mais rápidas, se o usuário digitar "r" iremos fazer uma jogada aleatória
			if moveStr == "r" {
				// Faz um movimento aleatório
				if err := MoveRandom(game); err != nil {
					fmt.Println(err)
					continue
				} else {
					break
				}
			} else {
				// Faz um movimento com base na jogada digitada pelo teclado
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

// PrintBoard exibe o tabuleiro informado
func PrintBoard(game *chess.Game) {
	fmt.Println(game.Position().Board().Draw())
	fmt.Println("Board evaluation: ", EvaluateStrongerSide(game))
	fmt.Println("Current FEN:", game.FEN())
}

// ReadMove lê um movimento a partir do teclado
func ReadMove() string {
	fmt.Print("Enter the move > ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// MoveRandom faz um movimento aleatório no tabuleiro informado
func MoveRandom(game *chess.Game) error {
	// Identifica os movimentos válidos no tabuleiro no momento
	moves := game.ValidMoves()
	lenMoves := len(moves)
	if len(moves) == 0 {
		return fmt.Errorf("there are no valid moves left")
	}
	// Escolhe um movimento aleatório entre os válidos
	move := moves[randomizer.Intn(lenMoves)]
	fmt.Println("Selected random move:", move.String())
	// Executa o movimento aleatório escolhido
	game.Move(move)
	return nil
}

// EvaluateStrongerSide calcula qual lado do tabuleiro está ganhando
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

// GameTreeNode representa um nó da game tree
type GameTreeNode struct {
	Game       *chess.Game
	Evaluation int
	Children   []*GameTreeNode
}

// NewGameTreeNode dado um tabuleiro, calcula quem está
// ganhando neste tabuleiro e retorna novo nó da game tree
func NewGameTreeNode(game *chess.Game) *GameTreeNode {
	return &GameTreeNode{
		Game:       game,
		Evaluation: EvaluateStrongerSide(game),
	}
}

// CloneGameTreeNode faz uma cópia de um nó da game tree, de forma que
// o tabuleiro contido possa ser alterado sem que o tabuleiro de outros
// nós sejam alterados também
func CloneGameTreeNode(gameTreeNode *GameTreeNode) *GameTreeNode {
	return &GameTreeNode{
		Game:       gameTreeNode.Game.Clone(),
		Evaluation: gameTreeNode.Evaluation,
		Children:   gameTreeNode.Children,
	}
}

// BuildGameTreeAt cria uma nova game tree a partir de um nó inicial
// previamente inicializado
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

// MaxNode verifica qual nó possui o maior evaluation, ou seja,
// o nó onde as peças brancas estão ganhando
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

// MinNode verifica qual nó possui o maior evaluation, ou seja,
// o nó onde as peças brancas estão ganhando
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

// AlphaBeta aplica o algoritmo a partir de um nó da game tree para
// encontrar a melhor jogada
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
