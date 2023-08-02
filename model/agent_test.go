// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPrintableKindSingular(t *testing.T) {
	agent := Agent{}
	require.Equal(t, "Agent", agent.PrintableKindSingular())
}

func TestPrintableKindPlural(t *testing.T) {
	agent := Agent{}
	require.Equal(t, "Agents", agent.PrintableKindPlural())
}

func TestPrintableFieldTitles(t *testing.T) {
	agent := Agent{}
	expected := []string{"ID", "Name", "Version", "Status", "Configuration", "Connected", "Disconnected", "Labels", "Reported"}
	require.ElementsMatch(t, expected, agent.PrintableFieldTitles())
}

func TestPrintableFieldValue(t *testing.T) {
	agent := Agent{
		ID:             "1",
		Name:           "Agent1",
		Version:        "v1.0",
		Status:         Connected,
		ConnectedAt:    nil,
		DisconnectedAt: nil,
		ConfigurationStatus: ConfigurationVersions{
			Current: "current",
			Pending: "pending",
			Future:  "future",
		},
	}

	require.Equal(t, "1", agent.PrintableFieldValue("ID"))
	require.Equal(t, "Agent1", agent.PrintableFieldValue("Name"))
	require.Equal(t, "v1.0", agent.PrintableFieldValue("Version"))
	require.Equal(t, agent.StatusDisplayText(), agent.PrintableFieldValue("Status"))
	require.Equal(t, agent.ConnectedDurationDisplayText(), agent.PrintableFieldValue("Connected"))
	require.Equal(t, agent.DisconnectedDurationDisplayText(), agent.PrintableFieldValue("Disconnected"))
	require.Equal(t, "current", agent.PrintableFieldValue("Configuration"))
	require.Equal(t, "current", agent.PrintableFieldValue("Current"))
	require.Equal(t, "pending", agent.PrintableFieldValue("Pending"))
	require.Equal(t, "future", agent.PrintableFieldValue("Future"))
	require.Equal(t, "", agent.PrintableFieldValue("Nonexistent Title"))
}

func TestConnectedDurationDisplayText(t *testing.T) {
	t.Run("Disconnected agent", func(t *testing.T) {
		agent := Agent{
			Status: Disconnected,
		}

		require.Equal(t, "-", agent.ConnectedDurationDisplayText())
	})

	t.Run("Connected agent", func(t *testing.T) {
		now := time.Now()
		agent := Agent{
			Status:      Connected,
			ConnectedAt: &now,
		}

		require.NotEqual(t, "-", agent.ConnectedDurationDisplayText())
	})
}

func TestDisconnectedDurationDisplayText(t *testing.T) {
	agent := Agent{}
	require.Equal(t, "-", agent.DisconnectedDurationDisplayText())
}

func TestDisconnectedSince(t *testing.T) {
	t.Run("Not disconnected", func(t *testing.T) {
		agent := Agent{}
		require.Equal(t, false, agent.DisconnectedSince(time.Now()))
	})

	t.Run("Disconnected before 'since'", func(t *testing.T) {
		disconnectedAt := time.Now().Add(-1 * time.Hour)
		agent := Agent{
			DisconnectedAt: &disconnectedAt,
		}
		require.Equal(t, true, agent.DisconnectedSince(time.Now()))
	})
}

func TestConnect(t *testing.T) {
	t.Run("New version", func(t *testing.T) {
		agent := &Agent{
			Version: "1.0.0",
		}
		agent.Connect("2.0.0")
		require.NotNil(t, agent.ConnectedAt)
		require.Nil(t, agent.DisconnectedAt)
	})

	t.Run("Not new version", func(t *testing.T) {
		connectedAnHourAgo := time.Now().Add(-1 * time.Hour)
		agent := &Agent{
			Version: "1.0.0",
		}
		agent.ConnectedAt = &connectedAnHourAgo
		agent.Connect("1.0.0")
		require.Equal(t, *agent.ConnectedAt, connectedAnHourAgo)
		require.Nil(t, agent.DisconnectedAt)
	})
}

func TestDisconnect(t *testing.T) {
	agent := &Agent{
		Status: Connected,
	}
	agent.Disconnect()
	require.NotNil(t, agent.DisconnectedAt)
	require.Equal(t, Disconnected, agent.Status)
}

func TestAgentConfigurationSetting(t *testing.T) {
	config1 := &Configuration{
		ResourceMeta: ResourceMeta{
			APIVersion: "v1",
			Metadata: Metadata{
				ID:      "1",
				Name:    "config1",
				Hash:    "hash1",
				Version: 1,
			},
		},
	}

	config2 := &Configuration{
		ResourceMeta: ResourceMeta{
			APIVersion: "v2",
			Metadata: Metadata{
				ID:      "2",
				Name:    "config2",
				Hash:    "hash2",
				Version: 1,
			},
		},
	}

	config3 := &Configuration{
		ResourceMeta: ResourceMeta{
			APIVersion: "v3",
			Metadata: Metadata{
				ID:      "3",
				Name:    "config3",
				Hash:    "hash3",
				Version: 1,
			},
		},
	}

	t.Run("SetCurrentConfiguration", func(t *testing.T) {
		t.Run("when configuration is nil", func(t *testing.T) {
			agent := &Agent{}
			agent.SetCurrentConfiguration(nil)
			require.Empty(t, agent.ConfigurationStatus.Current)
			require.Empty(t, agent.ConfigurationStatus.Pending)
			require.Empty(t, agent.ConfigurationStatus.Future)
		})

		t.Run("when configuration matches pending or future", func(t *testing.T) {
			agent := &Agent{
				ConfigurationStatus: ConfigurationVersions{
					Pending: config1.NameAndVersion(),
					Future:  config1.NameAndVersion(),
				},
			}
			agent.SetCurrentConfiguration(config1)
			require.Equal(t, config1.NameAndVersion(), agent.ConfigurationStatus.Current)
			require.Empty(t, agent.ConfigurationStatus.Pending)
			require.Empty(t, agent.ConfigurationStatus.Future)
		})
	})

	t.Run("SetPendingConfiguration", func(t *testing.T) {
		t.Run("when configuration is nil", func(t *testing.T) {
			agent := &Agent{}
			agent.SetPendingConfiguration(nil)
			require.Empty(t, agent.ConfigurationStatus.Current)
			require.Empty(t, agent.ConfigurationStatus.Pending)
			require.Empty(t, agent.ConfigurationStatus.Future)
		})

		t.Run("when configuration matches current", func(t *testing.T) {
			agent := &Agent{
				ConfigurationStatus: ConfigurationVersions{
					Current: config2.NameAndVersion(),
				},
			}
			agent.SetPendingConfiguration(config2)
			require.Equal(t, config2.NameAndVersion(), agent.ConfigurationStatus.Current)
			require.Empty(t, agent.ConfigurationStatus.Pending)
			require.Empty(t, agent.ConfigurationStatus.Future)
		})

		t.Run("when configuration matches future", func(t *testing.T) {
			agent := &Agent{
				ConfigurationStatus: ConfigurationVersions{
					Future: config2.NameAndVersion(),
				},
			}
			agent.SetPendingConfiguration(config2)
			require.Equal(t, config2.NameAndVersion(), agent.ConfigurationStatus.Pending)
			require.Empty(t, agent.ConfigurationStatus.Future)
		})
	})

	t.Run("SetFutureConfiguration", func(t *testing.T) {
		t.Run("when configuration is nil", func(t *testing.T) {
			agent := &Agent{}
			agent.SetFutureConfiguration(nil)
			require.Empty(t, agent.ConfigurationStatus.Current)
			require.Empty(t, agent.ConfigurationStatus.Pending)
			require.Empty(t, agent.ConfigurationStatus.Future)
		})

		t.Run("when configuration matches current", func(t *testing.T) {
			agent := &Agent{
				ConfigurationStatus: ConfigurationVersions{
					Current: config3.NameAndVersion(),
				},
			}
			agent.SetFutureConfiguration(config3)
			require.Equal(t, config3.NameAndVersion(), agent.ConfigurationStatus.Current)
			require.Empty(t, agent.ConfigurationStatus.Pending)
			require.Empty(t, agent.ConfigurationStatus.Future)
		})

		t.Run("when configuration matches pending", func(t *testing.T) {
			agent := &Agent{
				ConfigurationStatus: ConfigurationVersions{
					Pending: config3.NameAndVersion(),
				},
			}
			agent.SetFutureConfiguration(config3)
			require.Equal(t, config3.NameAndVersion(), agent.ConfigurationStatus.Pending)
			require.Empty(t, agent.ConfigurationStatus.Future)
		})
	})
}

func TestAgentStatusDisplayText(t *testing.T) {
	t.Run("Disconnected", func(t *testing.T) {
		a := &Agent{
			Status: Disconnected,
		}

		require.Equal(t, "Disconnected", a.StatusDisplayText())
	})

	t.Run("Connected", func(t *testing.T) {
		a := &Agent{
			Status: Connected,
		}

		require.Equal(t, "Connected", a.StatusDisplayText())
	})

	t.Run("Error", func(t *testing.T) {
		a := &Agent{
			Status: Error,
		}

		require.Equal(t, "Error", a.StatusDisplayText())
	})

	t.Run("ComponentFailed", func(t *testing.T) {
		a := &Agent{
			Status: ComponentFailed,
		}

		require.Equal(t, "Component Failed", a.StatusDisplayText())
	})

	t.Run("Deleted", func(t *testing.T) {
		a := &Agent{
			Status: Deleted,
		}

		require.Equal(t, "Deleted", a.StatusDisplayText())
	})

	t.Run("Configuring", func(t *testing.T) {
		a := &Agent{
			Status: Configuring,
		}

		require.Equal(t, "Configuring", a.StatusDisplayText())
	})

	t.Run("Upgrading", func(t *testing.T) {
		a := &Agent{
			Status: Upgrading,
		}

		require.Equal(t, "Upgrading", a.StatusDisplayText())
	})

	t.Run("Default", func(t *testing.T) {
		a := &Agent{
			Status: 123,
		}

		require.Equal(t, "Unknown", a.StatusDisplayText())
	})
}

func TestConfigurationVersions(t *testing.T) {
	t.Run("UniqueKey", func(t *testing.T) {
		cv := &ConfigurationVersions{
			Current: "current-1.0.0",
			Pending: "pending-2.0.0",
			Future:  "future-3.0.0",
		}

		key := cv.UniqueKey()

		expectedKey := "current-1.0.0|pending-2.0.0|future-3.0.0"
		require.Equal(t, expectedKey, key)
	})

	t.Run("Clear", func(t *testing.T) {
		cv := &ConfigurationVersions{
			Current: "current-1.0.0",
			Pending: "pending-2.0.0",
			Future:  "future-3.0.0",
		}

		cv.Clear()

		require.Equal(t, "", cv.Current)
		require.Equal(t, "", cv.Pending)
		require.Equal(t, "", cv.Future)
	})

	t.Run("Set", func(t *testing.T) {
		cv := &ConfigurationVersions{}

		cv.Set("newcurrent-1.0.0", "newpending-2.0.0", "newfuture-3.0.0")

		require.Equal(t, "newcurrent-1.0.0", cv.Current)
		require.Equal(t, "newpending-2.0.0", cv.Pending)
		require.Equal(t, "newfuture-3.0.0", cv.Future)
	})
}

func TestAgentUpgrade(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		upgrade := AgentUpgrade{
			Status:          UpgradeStarted,
			Version:         "1.0.0",
			AllPackagesHash: []byte("somehash"),
			Error:           "None",
		}

		value, err := upgrade.Value()
		require.NoError(t, err)

		// Convert the driver.Value to []byte and then to string
		valueStr := string(value.([]byte))

		expectedJSON := `{"status":1,"version":"1.0.0","allPackagesHash":"c29tZWhhc2g=","error":"None"}`
		require.JSONEq(t, expectedJSON, valueStr)
	})

	t.Run("Scan", func(t *testing.T) {
		upgrade := &AgentUpgrade{}

		json := []byte(`{"status":1,"version":"1.0.0","allPackagesHash":"c29tZWhhc2g=","error":"None"}`)

		err := upgrade.Scan(json)
		require.NoError(t, err)

		expectedUpgrade := &AgentUpgrade{
			Status:          UpgradeStarted,
			Version:         "1.0.0",
			AllPackagesHash: []byte("somehash"),
			Error:           "None",
		}

		require.Equal(t, expectedUpgrade, upgrade)
	})
}

func TestAgentApplyLabels(t *testing.T) {
	agent := Agent{}

	tests := []struct {
		selector string
		success  bool
		expect   Labels
	}{
		{
			selector: "app=mindplane",
			success:  true,
			expect: LabelsFromValidatedMap(map[string]string{
				"app": "mindplane",
			}),
		},
		{
			selector: "app=mindplane,env=production",
			success:  true,
			expect: LabelsFromValidatedMap(map[string]string{
				"app": "mindplane",
				"env": "production",
			}),
		},
		{
			selector: "app=mindplane, env = production",
			success:  true,
			expect: LabelsFromValidatedMap(map[string]string{
				"app": "mindplane",
				"env": "production",
			}),
		},
		{
			selector: "app=====",
			success:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.selector, func(t *testing.T) {
			labels, err := LabelsFromSelector(test.selector)
			agent.Labels = labels
			if test.success {
				require.NoError(t, err)
				require.Equal(t, test.expect, agent.Labels)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAgentMatchesSelector(t *testing.T) {
	tests := []struct {
		labels   map[string]string
		selector string
		matches  bool
	}{
		{
			labels: map[string]string{
				"app":     "mindplane",
				"os":      "Darwin",
				"version": "2.0.6",
			},
			selector: "app=mindplane",
			matches:  true,
		},
		{
			labels: map[string]string{
				"app":     "mindplane",
				"os":      "Darwin",
				"version": "2.0.6",
			},
			selector: "app=mindplane,version=2",
			matches:  false,
		},
		{
			labels: map[string]string{
				"app":     "mindplane",
				"os":      "Darwin",
				"version": "2.0.6",
			},
			selector: "os=Darwin,app=mindplane",
			matches:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.selector, func(t *testing.T) {
			selector, err := SelectorFromString(test.selector)
			require.NoError(t, err)
			require.Equal(t, test.matches, selector.Matches(LabelsFromValidatedMap(test.labels)))
		})
	}
}

func TestAgentSupportsUpgrade(t *testing.T) {
	tests := []struct {
		version string
		expect  bool
	}{
		{
			version: "v1.5.0",
			expect:  false,
		},
		{
			version: "v1.6.0",
			expect:  true,
		},
		{
			version: "v1.6.1",
			expect:  true,
		},
		{
			version: "v2.0.0",
			expect:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			agent := &Agent{
				Version: test.version,
			}
			require.Equal(t, test.expect, agent.SupportsUpgrade())
		})
	}
}

func TestAgentUpgradeComplete(t *testing.T) {
	tests := []struct {
		name          string
		prepareAgent  func(a *Agent)
		errorMessage  string
		expectStatus  AgentStatus
		expectUpgrade *AgentUpgrade
	}{
		{
			name: "success",
			prepareAgent: func(a *Agent) {
				a.UpgradeTo("v1.1")
			},
			errorMessage:  "",
			expectStatus:  Connected,
			expectUpgrade: nil,
		},
		{
			name: "success with existing error",
			prepareAgent: func(a *Agent) {
				a.UpgradeStarted("v1.1", []byte{1})
				a.ErrorMessage = "error"
			},
			errorMessage:  "",
			expectStatus:  Error,
			expectUpgrade: nil,
		},
		{
			name: "fail",
			prepareAgent: func(a *Agent) {
				a.Status = Connected
			},
			errorMessage: "upgrade error",
			expectStatus: Connected,
			expectUpgrade: &AgentUpgrade{
				Status:  UpgradeFailed,
				Version: "v1.2",
				Error:   "upgrade error",
			},
		},
		{
			name: "fail with upgrade",
			prepareAgent: func(a *Agent) {
				a.UpgradeTo("v1.1")
				a.Status = Connected
			},
			errorMessage: "upgrade error",
			expectStatus: Connected,
			expectUpgrade: &AgentUpgrade{
				Status:  UpgradeFailed,
				Version: "v1.2",
				Error:   "upgrade error",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent := &Agent{}
			test.prepareAgent(agent)
			agent.UpgradeComplete("v1.2", test.errorMessage)
			require.Equal(t, test.expectStatus, agent.Status)
			require.Equal(t, test.expectUpgrade, agent.Upgrade)
		})
	}
}

func TestFeatures(t *testing.T) {
	tests := []struct {
		version        string
		expectFeatures AgentFeatures
	}{
		{
			version:        "",
			expectFeatures: 0,
		},
		{
			version:        "1.5.0",
			expectFeatures: 0,
		},
		{
			version:        "1.6.0",
			expectFeatures: AgentSupportsUpgrade,
		},
		{
			version:        "1.7.0",
			expectFeatures: AgentSupportsUpgrade,
		},
		{
			version:        "1.8.0",
			expectFeatures: AgentSupportsUpgrade | AgentSupportsSnapshots,
		},
		{
			version:        "1.9.2",
			expectFeatures: AgentSupportsUpgrade | AgentSupportsSnapshots | AgentSupportsMeasurements,
		},
		{
			version:        "1.14.0",
			expectFeatures: AgentSupportsUpgrade | AgentSupportsSnapshots | AgentSupportsMeasurements | AgentSupportsLogBasedMetrics,
		},
		{
			version:        "2.0.0",
			expectFeatures: AgentSupportsUpgrade | AgentSupportsSnapshots | AgentSupportsMeasurements | AgentSupportsLogBasedMetrics,
		},
	}
	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			agent := Agent{
				Version: test.version,
			}
			require.Equal(t, test.expectFeatures, agent.Features())
		})
	}
}

func TestAgentIndexID(t *testing.T) {
	agent := &Agent{
		ID: "test_id",
	}

	require.Equal(t, "test_id", agent.IndexID())
}

func TestAgentIndexFields(t *testing.T) {
	agent := &Agent{
		ID:              "test_id",
		Name:            "test_name",
		Type:            "test_type",
		Architecture:    "test_architecture",
		HostName:        "test_hostname",
		Version:         "test_version",
		Home:            "test_home",
		Platform:        "test_platform",
		OperatingSystem: "test_os",
		MacAddress:      "test_mac_address",
		Status:          Connected, // Assume Connected status for test
		ConfigurationStatus: ConfigurationVersions{
			Current: "test_current_config",
			Pending: "test_pending_config",
			Future:  "test_future_config",
		},
	}

	// Create a map to collect the index fields
	indexFields := make(map[string]string)
	indexFunc := func(name string, value string) {
		indexFields[name] = value
	}

	agent.IndexFields(indexFunc)

	// Check if the correct fields were indexed
	require.Equal(t, "test_id", indexFields["id"])
	require.Equal(t, "test_architecture", indexFields["arch"])
	require.Equal(t, "test_hostname", indexFields["hostname"])
	require.Equal(t, "test_platform", indexFields["platform"])
	require.Equal(t, "test_version", indexFields["version"])
	require.Equal(t, "test_name", indexFields["name"])
	require.Equal(t, "test_home", indexFields["home"])
	require.Equal(t, "test_os", indexFields["os"])
	require.Equal(t, "test_mac_address", indexFields["macAddress"])
	require.Equal(t, "test_type", indexFields["type"])
	require.Equal(t, agent.StatusDisplayText(), indexFields["status"])
	require.Equal(t, "test_current_config", indexFields[FieldConfigurationCurrent])
	require.Equal(t, "test_pending_config", indexFields[FieldConfigurationPending])
	require.Equal(t, "test_future_config", indexFields[FieldConfigurationFuture])
	require.Equal(t, "test_current_config", indexFields[FieldRolloutComplete])
	require.Equal(t, "test_pending_config", indexFields[FieldRolloutPending])
	require.Equal(t, "test_future_config", indexFields[FieldRolloutWaiting])
}

func TestSortAgentsByName(t *testing.T) {
	agents := []*Agent{
		{Name: "agent_c"},
		{Name: "agent_a"},
		{Name: "agent_b"},
	}

	SortAgentsByName(agents)

	// After sorting by name, the agents should be in order "agent_a", "agent_b", "agent_c"
	require.Equal(t, "agent_a", agents[0].Name)
	require.Equal(t, "agent_b", agents[1].Name)
	require.Equal(t, "agent_c", agents[2].Name)
}
