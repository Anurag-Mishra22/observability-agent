package logging

import (
	"encoding/json"
	"fmt"
)

// Sink writes logs somewhere (stdout for now)
func Sink(event LogEvent) {
	data, _ := json.Marshal(event)
	fmt.Println(string(data))
}
