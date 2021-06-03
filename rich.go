// ## Imports and globals
package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

const (
	// Number of agents in the market
	numOfAgents = 10
	// Initial amount of money that each agent owns
	initialWealth = 100.0
	// How many trades to simulate
	rounds = 10000
	// If the poorer agent gains wealth it is this percentage of their total wealth.
	percentGain = 0.20
	// If the poorer agent loses wealth, it is this percentage of their total wealth.
	percentLoss = 0.17
)

// Agents are defined by the amount of money, or wealth, they have.
type agents []float64

// pickTwoRandomAgents generates two random numbers `sender` and `receiver` between 0 and numOfAgents-1
// and ensures that `sender` and `receiver` are not equal. (After all, agents would not trade with themselves.)
// Note the use of named return values that saves an extra declaration of `receiver` outside the loop
// (to avoid that `receiver` exists only in the scope of the loop).
func pickTwoRandomAgents() (sender, receiver int) {
	sender = rand.Intn(numOfAgents)
	receiver = sender

	// Generate a random`receiver`. Repeat until `receiver` != `sender`
	for receiver == sender {
		receiver = rand.Intn(numOfAgents)
	}
	return sender, receiver
}

// The trading formula assumes that agents sometimes pay either more or less than the traded good is worth.
// Because of this, wealth flows from one agent to another.
// As both agents `sender`, `receiver` were already chosen randomly, we can decide at this point that agent `sender` always loses
// wealth, and agent `receiver` always gains wealth in this transaction.
// Note: the agents
func trade(a agents, sender, receiver int) {
	// Wealth flows from sender to `receiver` in this transaction.
	// The amount that flows from sender to `receiver` is always a given percentage of the poorer agent.

	// If`receiver` is the poorer agent, the gain is `percentGain` of `receiver`'s total wealth.
	transfer := a[receiver] * percentGain

	// If `sender` is the poorer agent, the loss is `percentLoss` of `sender`'s total wealth.
	if a[sender] < a[receiver] {
		transfer = a[sender] * percentLoss
	}
	// It's a deal!
	a[sender] -= transfer
	a[receiver] += transfer
}

// Draw a bar chart of the current wealth of all agents
func drawChart(a agents, bc *widgets.BarChart) {
	bc.Data = a
	// Scale the bar chart dynamically, to better see
	// the distribution when the current maximum wealth is
	// much smaller than the maximum possible wealth.
	maxPossibleWealth := initialWealth * numOfAgents
	currentMaxWealth, _ := ui.GetMaxFloat64FromSlice(a)
	bc.MaxVal = currentMaxWealth + (maxPossibleWealth-currentMaxWealth)*0.05
	ui.Render(bc)
}

// Run the simulation
func run(a agents, bc *widgets.BarChart, done <-chan struct{}) {
	for n := 0; n < rounds; n++ {
		// Pick two different agents.
		sender, receiver := pickTwoRandomAgents()
		// Have them do a trade.
		trade(a, sender, receiver)
		// Update the chart
		drawChart(a, bc)
		// Try to read a value from channel `done`.
		// The read shall not block, hence it is enclosed in a
		// select block with a default clause.
		select {
		case <-done:
			// At this point, the done channel has unblocked and emitted a zero value. Leave the simulation loop.
			return
		default:
		}
	}
}

func main() {
	// Setup

	// Pre-allocate the slice, to avoid allocations during the simulation
	a := make(agents, numOfAgents)

	// Set a random seed
	rand.Seed(time.Now().UnixNano())

	for i := range a {
		// All agents start with the same amount of money.
		a[i] = initialWealth
	}

	// UI setup. `gizak/termui` makes rendering a bar chart in a terminal super easy.
	err := ui.Init()
	if err != nil {
		log.Fatalln(err)
	}
	defer ui.Close()
	bc := widgets.NewBarChart()
	bc.Title = "Agents' Wealth"
	bc.BarWidth = 5
	bc.SetRect(5, 5, 10+(bc.BarWidth+1)*numOfAgents, 25)
	bc.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorBlue)}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorBlack)}
	bc.NumFormatter = func(n float64) string {
		return fmt.Sprintf("%3.1f", n)
	}
	// Start rendering.
	ui.Render(bc)

	// `termui` has its own event polling.
	// We use this here to watch for a key press
	// to end the simulation
	done := make(chan struct{})
	go func(done chan<- struct{}) {
		for e := range ui.PollEvents() {
			if e.Type == ui.KeyboardEvent {
				// Unblock the channel by closing it
				// After closing the channel, it emits zero values upon reading.
				close(done)
				return
			}
		}
	}(done)

	// Start the simulation!
	run(a, bc, done)

	// After the simulation, wait for a key press
	// so that the final chart remains visible.
	<-done
}