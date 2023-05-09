package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var board_size, num_mins int

var boards = map[string]*MineSweeper{}

type MineSweeper struct {
	board     [][]int
	cell      []*MineSweeperCell
	gameEnded bool
}

type MineSweeperCell struct {
	x      int
	y      int
	isOpen bool
}

func (m *MineSweeper) initialize(board_size int) [][]int {
	var cells []*MineSweeperCell
	board := [][]int{}
	for i := 0; i < board_size; i++ {
		row := []int{}
		for j := 0; j < board_size; j++ {
			row = append(row, 0)
			cells = append(cells, &MineSweeperCell{i, j, false})
		}
		board = append(board, row)
	}
	m.board = board
	m.cell = cells
	return board
}

func (m *MineSweeper) createMines(num_mins int, board_size int) [][]int {
	rand.Seed(time.Now().UnixNano())
	mine_nums := rand.Perm(board_size*board_size - 1)
	mine_coords := [][]int{}
	for _, mi := range mine_nums[:num_mins] {
		mine_coords = append(mine_coords, []int{mi / board_size, mi % board_size})
		m.board[mi/board_size][mi%board_size] = -1
	}
	return mine_coords
}

func (m *MineSweeper) clearEmpty() {
	for _x, x := range m.board {
		for _y, y := range x {
			if y == 0 && m.checkIfOpen(_x, _y) {
				neighs := m.getNeighbours(_x, _y)
				for _, ncoords := range neighs {
					m.openCell(ncoords[0], ncoords[1])
				}
			}
		}
	}
}

func (m *MineSweeper) add_neighbours(mine_coords [][]int) {
	for _, coords := range mine_coords {
		neighs := m.getNeighbours(coords[0], coords[1])
		for _, neighcoord := range neighs {
			m.board[neighcoord[0]][neighcoord[1]] = m.board[neighcoord[0]][neighcoord[1]] + 1
		}
	}
}

func (m *MineSweeper) getNeighbours(x int, y int) [][]int {
	l := []int{-1, 0, 1}
	neighs := [][]int{}
	for _, i := range l {
		for _, j := range l {
			_x := x + i
			_y := y + j
			if 0 <= _x && _x < board_size && 0 <= _y && _y < board_size && m.board[_x][_y] != -1 {
				neighs = append(neighs, []int{_x, _y})
			}
		}
	}
	return neighs
}

func (m *MineSweeper) checkIfOpen(x int, y int) bool {
	for _, cell := range m.cell {
		if cell.x == x && cell.y == y {
			return cell.isOpen
		}
	}
	return false
}

func (m *MineSweeper) openCell(x int, y int) {
	for _, cell := range m.cell {
		if cell.x == x && cell.y == y {
			cell.isOpen = true
		}
	}
}

func (m *MineSweeper) openAll() {
	m.gameEnded = true
	for _, cell := range m.cell {
		cell.isOpen = true
	}
}

func (m *MineSweeper) checkWin() bool {
	for _, cell := range m.cell {
		if !cell.isOpen {
			if m.board[cell.x][cell.y] == -1 {
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func MinesweeperCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var comps []discordgo.MessageComponent
	m := MineSweeper{gameEnded: false}
	m.initialize(5)
	num_mins = 3
	board_size = 5
	mi := m.createMines(num_mins, board_size)
	m.add_neighbours(mi)
	fmt.Print(m.v)
	comps = getComponents(&m, i.Member.User.ID)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf("%s#%s's minesweeper game (3 mines)", i.Member.User.Username, i.Member.User.Discriminator),
			Components: comps,
		},
	})
	if err != nil {
		fmt.Println("Error responsding to minesweeper: ", err)
	}
	boards[i.Member.User.ID] = &m
}

func MineButtonHandle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	split := strings.Split(i.MessageComponentData().CustomID, "_")
	if split[0] == i.Member.User.ID {
		x, _ := strconv.Atoi(split[3])
		y, _ := strconv.Atoi(split[4])
		m := boards[i.Member.User.ID]
		m.openCell(x, y)
		content := fmt.Sprintf("%s#%s's minesweeper game (3 mines)", i.Member.User.Username, i.Member.User.Discriminator)
		switch split[2] {
		case "0":
			m.clearEmpty()
		case "-1":
			m.openAll()
			content = "Kaboom!"
		}
		if split[2] != "-1" {
			if m.checkWin() {
				m.openAll()
				content = "You won!"
			}
		}
		comps := getComponents(m, i.Member.User.ID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    content,
				Components: comps,
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "This isn't your game!", Flags: discordgo.MessageFlagsEphemeral},
		})
	}
}

func getComponents(m *MineSweeper, userID string) []discordgo.MessageComponent {
	comp := []discordgo.MessageComponent{}
	for x, row := range m.board {
		actionrow := []discordgo.MessageComponent{}
		for y, cell := range row {
			var label string
			var disabled bool
			if cell == 0 || cell == -1 {
				label = "\u200b"
			} else {
				if !m.checkIfOpen(x, y) {
					label = "\u200b"
				} else {
					label = fmt.Sprintf("%d", cell)
				}
			}
			if m.checkIfOpen(x, y) {
				disabled = true
			} else {
				disabled = false
			}
			style := discordgo.PrimaryButton
			if m.gameEnded {
				if cell == -1 {
					style = discordgo.DangerButton
				}
			}
			if cell == -1 && m.gameEnded {
				actionrow = append(actionrow, discordgo.Button{
					Label:    label,
					Style:    style,
					CustomID: fmt.Sprintf("%s_mine_%d_%+d_%d", userID, cell, x, y),
					Disabled: disabled,
					Emoji:    discordgo.ComponentEmoji{Name: "💣"},
				})
			} else {
				actionrow = append(actionrow, discordgo.Button{
					Label:    label,
					Style:    style,
					CustomID: fmt.Sprintf("%s_mine_%d_%+d_%d", userID, cell, x, y),
					Disabled: disabled,
				})
			}
		}
		act := discordgo.ActionsRow{Components: actionrow}
		comp = append(comp, act)
	}
	return comp
}
