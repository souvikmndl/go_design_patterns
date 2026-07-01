package payment

import (
	"fmt"
	"os"
)

// Here we will register which payment process to use
// without if=else or switch
// Using a Registry pattern

var registry = make(map[string]Processor)

// Register add a payment processor to the registry
// This would be usually done during start up or Service init()
// thats why we can panic or os.exit
func Register(name string, processor Processor) {
	if _, exists := registry[name]; exists {
		fmt.Printf("processor %s already exists\n", name)
		os.Exit(1)
	}

	registry[name] = processor
}

// Get retrieves a processor by name.
/*
Notice: There is
- no switch
- no if-else
- no factory to modify
*/
func Get(name string) (Processor, error) {
	processor, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("unknown processor: %s", name)
	}

	return processor, nil
}
