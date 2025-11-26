package id

import (
	"fmt"
	"ridash/utils/config"
	"time"

	"github.com/sony/sonyflake/v2"
	"go.uber.org/zap"
)

var sf *sonyflake.Sonyflake

// Init initializes the Sonyflake ID generator
// Must be called before using GetID()
func Init() error {
	var machineIDFunc func() (int, error)

	appID := config.Env().AppMachineID
	if appID > 0 {
		if appID >= 1024 {
			zap.L().Fatal("Machine ID must be between 0 and 1023", zap.Int("machine_id", int(appID)))
		}
		machineIDFunc = func() (int, error) {
			return int(appID), nil
		}
	}

	settings := sonyflake.Settings{
		StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID: machineIDFunc,
		CheckMachineID: nil,
		BitsSequence:   10,                   // 1024 id per ms per machine
		BitsMachineID:  10,                   // 1024 machines
		TimeUnit:       1 * time.Millisecond, // Time unit
	}

	var err error
	sf, err = sonyflake.New(settings)
	if err != nil {
		return fmt.Errorf("failed to initialize sonyflake: %w", err)
	}

	return nil
}

// GetID generates a new unique ID
// Returns the ID as uint64 or an error if generation fails
func GetID() (int64, error) {
	if sf == nil {
		return 0, fmt.Errorf("sonyflake not initialized, call Init() first")
	}

	id, err := sf.NextID()
	if err != nil {
		return 0, fmt.Errorf("failed to generate ID: %w", err)
	}
	
	return id, nil
}
