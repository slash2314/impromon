package tui

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	spinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	PollResults PollResults
	Urls        []string
	UpSpinner   spinner.Model
	DownSpinner spinner.Model
}

func InitModel(urls []string) Model {
	slowSpinner := spinner.Meter
	slowSpinner.FPS = time.Second / 3
	upSpinner := spinner.New(
		spinner.WithSpinner(slowSpinner),
		spinner.WithStyle(lipgloss.NewStyle().
			BorderBackground(lipgloss.Color("0")).
			Foreground(lipgloss.Color("46"))))
	downSpinner := spinner.New(
		spinner.WithSpinner(slowSpinner),
		spinner.WithStyle(lipgloss.NewStyle().
			BorderBackground(lipgloss.Color("0")).
			Foreground(lipgloss.Color("196"))))
	return Model{
		Urls:        urls,
		UpSpinner:   upSpinner,
		DownSpinner: downSpinner,
		PollResults: NewPollResults(),
	}
}

type pollResult struct {
	url      string
	status   int
	err      error
	lastPoll time.Time
	protocol string
}

type PollResults struct {
	statuses     map[string]int
	errs         map[string]error
	pollingTimes map[string]time.Time
	protocols    map[string]string
}

func NewPollResults() PollResults {
	return PollResults{
		statuses:     make(map[string]int),
		errs:         make(map[string]error),
		pollingTimes: make(map[string]time.Time),
		protocols:    make(map[string]string),
	}
}

func (pr PollResults) AddResult(p pollResult) PollResults {
	pr.statuses[p.url] = p.status
	pr.errs[p.url] = p.err
	pr.pollingTimes[p.url] = p.lastPoll
	pr.protocols[p.url] = p.protocol
	return pr
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.DownSpinner.Tick, m.UpSpinner.Tick, m.checkUrls())
}

// urlPrefix returns the prefix of the url
// e.g. http://example.com:8080 returns http
func urlPrefix(url string) (string, error) {
	colonPos := strings.Index(url, ":")
	if colonPos == -1 {
		return "", fmt.Errorf("invalid url %s", url)
	}
	return url[:colonPos], nil
}

// checkUrls checks the urls every second
// It returns a tea.Cmd that will be executed every second
// It sends a tickMsg to the update function
// The URLs are checked in parallel
func (m *Model) checkUrls() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		pollResultChan := make(chan pollResult)
		wg := sync.WaitGroup{}
		wg.Add(len(m.Urls))
		for _, url := range m.Urls {
			go func(url string) {
				m.processUrl(url, &wg, pollResultChan)
			}(url)
		}
		go func() {
			wg.Wait()
			close(pollResultChan)
		}()
		for pollRes := range pollResultChan {
			m.PollResults.AddResult(pollRes)
		}
		return m.PollResults
	})
}

// processUrl checks the status of the url
// It sends the result to the pollResultChan
func (m *Model) processUrl(url string, wg *sync.WaitGroup, pollResultChan chan pollResult) {
	defer wg.Done()
	pollRes := pollResult{url: url}
	pollRes.lastPoll = time.Now()
	// Get prefix to get the protocols from the url
	prefix, err := urlPrefix(url)
	if err != nil {
		pollRes.err = err
		pollResultChan <- pollRes
		return
	}
	var status int
	switch prefix {
	case "http", "https":
		status, err = httpCheck(url, 1*time.Second)
	case "tcp":
		status, err = tcpCheck(url, 1*time.Second)
	default:
		pollRes.err = fmt.Errorf("unsupported protocol %s", prefix)
		pollResultChan <- pollRes
		return
	}

	pollRes.protocol = prefix
	if err != nil {
		pollRes.err = err
		pollResultChan <- pollRes
		return
	}
	pollRes.status = status
	pollResultChan <- pollRes
}

// httpCheck checks the status of the url
// It returns the status code and an error
func httpCheck(url string, timeout time.Duration) (int, error) {
	c := &http.Client{Timeout: timeout}
	res, err := c.Get(url)
	if err != nil {
		return 0, err
	}
	return res.StatusCode, nil
}

type UriComponents struct {
	Protocol string
	Host     string
	Port     int
}

// parseUri parses the url and returns the protocol, host and port
// e.g. http://example.com:8080 returns UriComponents{Protocol: "http", Host: "example.com", Port: 8080}
func parseUri(url string) (UriComponents, error) {
	endProtocolIndex := strings.Index(url, "://")
	protocol := url[:endProtocolIndex]
	portStartIndex := strings.LastIndex(url, ":") + 1

	portStr := url[portStartIndex:]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return UriComponents{}, fmt.Errorf("error getting port from url %s", url)
	}
	// Get IP from url
	return UriComponents{
		Protocol: protocol,
		Host:     url[endProtocolIndex+3 : portStartIndex-1],
		Port:     port,
	}, nil

}

// tcpCheck checks the status of the url
// It returns 1 if the url is reachable
// It returns an error if the url is not reachable
func tcpCheck(url string, timeout time.Duration) (int, error) {
	uriComps, err := parseUri(url)
	if err != nil {
		return 0, err
	}
	// Get IP from url
	address := fmt.Sprintf("%s:%d", uriComps.Host, uriComps.Port)
	con, err := net.DialTimeout(uriComps.Protocol, address, timeout)
	if err != nil {
		return 0, err
	}
	con.Close()
	return 1, nil
}

type tickMsg time.Time

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case PollResults:
		m.PollResults = msg
		return m, m.checkUrls()
	case tickMsg:
		return m, m.checkUrls()
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	default:
		var okCmd tea.Cmd
		var downCmd tea.Cmd
		m.UpSpinner, okCmd = m.UpSpinner.Update(msg)
		m.DownSpinner, downCmd = m.DownSpinner.Update(msg)
		return m, tea.Batch(okCmd, downCmd)
	}
	return m, nil

}

// View will display the status of the urls in the terminal
func (m *Model) View() string {
	pr := m.PollResults
	var s string
	upSpinView := m.UpSpinner.View()

	if len(pr.statuses) == 0 {
		s += fmt.Sprintf("%s Checking urls ...", upSpinView)
	}
	for _, url := range m.Urls {
		pollingTime := pr.pollingTimes[url].Format(time.RFC3339)
		if urlStatus, ok := pr.statuses[url]; ok && urlStatus != 0 {
			switch pr.protocols[url] {
			case "http", "https":
				status := pr.statuses[url]
				statusText := http.StatusText(status)
				s += fmt.Sprintf("\n%s %s %d %s at %s", upSpinView, url, status, statusText, pollingTime)
			case "tcp":
				s += fmt.Sprintf("\n%s %s at %s", upSpinView, url, pollingTime)
			}
		} else if pr.errs[url] != nil {
			s += fmt.Sprintf("\n%s %s %s at %s", m.DownSpinner.View(), url, pr.errs[url].Error(), pollingTime)
		}
	}

	// Send off whatever we came up with above for rendering.
	return "\n" + s + "\n\n"

}
