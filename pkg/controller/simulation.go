package controller

import (
	"encoding/json"
	"os"

	kp "k8s.io/utils/path"
)

var SimulationConfigPath = "/tmp/mc-simulation-config.json"

type SimulationConfig struct {
	ScaleErrorForMachineDeployments []string
}

func CreateLoadSimulationConfig() (simConfig SimulationConfig, err error) {
	var data []byte

	exists, err := kp.Exists(kp.CheckFollowSymlink, SimulationConfigPath)
	if err != nil {
		return
	}

	if !exists {
		data, err = json.MarshalIndent(simConfig, "", " ")
		if err != nil {
			return
		}
		err = os.WriteFile(SimulationConfigPath, data, 0o666)
		return
	}

	data, err = os.ReadFile(SimulationConfigPath)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &simConfig)
	return
}
