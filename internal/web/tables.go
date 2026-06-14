package web

import (
	"fmt"
	"time"
)

// tableData is the demo dataset for the scrollable tables page. It is generated
// deterministically — a direct port of the Ruby that previously lived in the
// Slim template.
type tableData struct {
	Columns      []string
	HeaderFrozen string
	HeaderScroll []string
	Rows         []tableRow
}

type tableRow struct {
	Cells  []string // the data cells (every column except the trailing action)
	Action string   // label for the action button in the final column
}

func buildTableData() tableData {
	columns := []string{
		"ID", "Company", "Region", "Owner", "Stage", "Health",
		"Revenue", "Growth", "Seats", "Updated", "Priority", "Reset",
	}
	companies := []string{"Northwind Systems", "Fjord Labs", "Aurora Works", "Polar Metrics", "Signal Forge", "Delta Harbor", "Vector House", "Granite Cloud", "Summit Grid", "Orbit Field"}
	regions := []string{"Oslo", "Bergen", "Trondheim", "Tromsø", "Stockholm", "Copenhagen", "Aarhus", "Helsinki", "Reykjavík", "Gothenburg"}
	owners := []string{"M. Åsrud", "S. Nilsen", "K. Dahl", "E. Hansen", "R. Holm", "I. Olsen", "T. Moe", "A. Lind", "D. Solberg", "L. Berg"}
	stages := []string{"Discovery", "Pilot", "Active", "Expansion"}
	healthStates := []string{"Stable", "Watch", "Strong", "Risk"}
	priorities := []string{"High", "Medium", "Low"}

	start := time.Date(2025, time.April, 1, 0, 0, 0, 0, time.UTC)

	rows := make([]tableRow, 0, 100)
	for index := 1; index <= 100; index++ {
		i := index - 1
		cells := []string{
			fmt.Sprintf("AC-%03d", index),
			companies[i%len(companies)],
			regions[i%len(regions)],
			owners[i%len(owners)],
			stages[i%len(stages)],
			healthStates[i%len(healthStates)],
			fmt.Sprintf("$%dk", 12+((index*7)%85)),
			fmt.Sprintf("%+d%%", (index%18)-4),
			fmt.Sprintf("%d", 6+(index%28)),
			start.AddDate(0, 0, i).Format("2006-01-02"),
			priorities[i%len(priorities)],
		}
		rows = append(rows, tableRow{Cells: cells, Action: "Reset"})
	}

	return tableData{
		Columns:      columns,
		HeaderFrozen: columns[0],
		HeaderScroll: columns[1:],
		Rows:         rows,
	}
}
