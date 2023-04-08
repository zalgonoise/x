package stream

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zalgonoise/x/audio/wav/fft"
)

const labelUnit = 10

type EQModel struct {
	Data []fft.FrequencyPower
}

func (m EQModel) Init() tea.Cmd {
	return nil
}

func (m EQModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// exit
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case []fft.FrequencyPower:
		m.Data = msg
	}

	return m, nil
}

func normalize(index, graphHeight int, value float64) bool {
	boost := value * 10000.0
	return int(boost)<<3>>(graphHeight-index)/2 > 0
}

func (m EQModel) View() string {
	graphWidth := len(m.Data)*labelUnit + 1
	graphHeight := 16
	sb := new(strings.Builder)

	for i := 0; i < graphHeight; i++ {
		for j := 0; j < graphWidth; j++ {
			if i == 0 || i == graphHeight-1 {
				sb.WriteRune('-')
				continue
			}

			switch j {
			case 0, graphWidth - 1:
				sb.WriteRune('|')
			default:
				switch j % labelUnit {
				case 0:
					sb.WriteRune('|')
				case 3, 4, 5, 6:
					if normalize(i, graphHeight, m.Data[j/labelUnit].Mag) {
						sb.WriteRune('█')
						continue
					}
					sb.WriteRune(' ')
				default:
					sb.WriteRune(' ')
				}
			}
		}
		sb.WriteByte('\n')
	}
	sb.WriteRune('|')
	for i := range m.Data {
		lb := new(strings.Builder)
		label := strconv.Itoa(m.Data[i].Freq)
		if len(label) > labelUnit-3 {
			label = label[:4] + "..."
		}
		leftover := (labelUnit - 1 - len(label)) / 2
		for i := 0; i < leftover; i++ {
			lb.WriteRune(' ')
		}
		lb.WriteString(label)
		for i := 0; i < leftover; i++ {
			lb.WriteRune(' ')
		}
		for lb.Len() < labelUnit-1 {
			lb.WriteRune(' ')
		}
		lb.WriteRune('|')
		sb.WriteString(lb.String())
	}

	return lipgloss.NewStyle().Width(graphWidth).Render(sb.String())
}

func NewEQ(ch <-chan []fft.FrequencyPower) error {
	p := tea.NewProgram(EQModel{})
	go func() {
		for spectrum := range ch {
			p.Send(spectrum)
		}
	}()
	_, err := p.Run()
	if err != nil {
		return err
	}
	return nil
}
