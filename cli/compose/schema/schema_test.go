// FIXME(thaJeztah): remove once we are a module; the go:build directive prevents go from downgrading language version to go1.16:
//go:build go1.23

package schema

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

type dict map[string]any

func TestValidate(t *testing.T) {
	config := dict{
		"version": "3.0",
		"services": dict{
			"foo": dict{
				"image": "busybox",
			},
		},
	}

	assert.NilError(t, Validate(config, "3.0"))
	assert.NilError(t, Validate(config, "3"))
	assert.NilError(t, Validate(config, ""))
	assert.ErrorContains(t, Validate(config, "1.0"), "unsupported Compose file version: 1.0")
	assert.ErrorContains(t, Validate(config, "12345"), "unsupported Compose file version: 12345")
}

func TestValidatePorts(t *testing.T) {
	testcases := []struct {
		ports    any
		hasError bool
	}{
		{
			ports:    []int{8000},
			hasError: false,
		},
		{
			ports:    []string{"8000:8000"},
			hasError: false,
		},
		{
			ports:    []string{"8001-8005"},
			hasError: false,
		},
		{
			ports:    []string{"8001-8005:8001-8005"},
			hasError: false,
		},
		{
			ports:    []string{"8000"},
			hasError: false,
		},
		{
			ports:    []string{"8000-9000:80"},
			hasError: false,
		},
		{
			ports:    []string{"[::1]:8080:8000"},
			hasError: false,
		},
		{
			ports:    []string{"[::1]:8080-8085:8000"},
			hasError: false,
		},
		{
			ports:    []string{"127.0.0.1:8000:8000"},
			hasError: false,
		},
		{
			ports:    []string{"127.0.0.1:8000-8005:8000-8005"},
			hasError: false,
		},
		{
			ports:    []string{"127.0.0.1:8000:8000/udp"},
			hasError: false,
		},
		{
			ports:    []string{"8000:8000/udp"},
			hasError: false,
		},
		{
			ports:    []string{"8000:8000/http"},
			hasError: true,
		},
		{
			ports:    []string{"-1"},
			hasError: true,
		},
		{
			ports:    []string{"65536"},
			hasError: true,
		},
		{
			ports:    []string{"-1:65536/http"},
			hasError: true,
		},
		{
			ports:    []string{"invalid"},
			hasError: true,
		},
		{
			ports:    []string{"12345678:8000:8000/tcp"},
			hasError: true,
		},
		{
			ports:    []string{"8005-8000:8005-8000"},
			hasError: true,
		},
		{
			ports:    []string{"8006-8000:8005-8000"},
			hasError: true,
		},
	}

	for _, tc := range testcases {
		config := dict{
			"version": "3.0",
			"services": dict{
				"foo": dict{
					"image": "busybox",
					"ports": tc.ports,
				},
			},
		}
		if tc.hasError {
			assert.ErrorContains(t, Validate(config, "3"), "services.foo.ports.0 Does not match format 'ports'")
		} else {
			assert.NilError(t, Validate(config, "3"))
		}
	}
}

func TestValidateUndefinedTopLevelOption(t *testing.T) {
	config := dict{
		"version": "3.0",
		"helicopters": dict{
			"foo": dict{
				"image": "busybox",
			},
		},
	}

	err := Validate(config, "3.0")
	assert.ErrorContains(t, err, "Additional property helicopters is not allowed")
}

func TestValidateAllowsXTopLevelFields(t *testing.T) {
	config := dict{
		"version":       "3.4",
		"x-extra-stuff": dict{},
	}

	err := Validate(config, "3.4")
	assert.NilError(t, err)
}

func TestValidateAllowsXFields(t *testing.T) {
	config := dict{
		"version": "3.7",
		"services": dict{
			"bar": dict{
				"x-extra-stuff": dict{},
			},
		},
		"volumes": dict{
			"bar": dict{
				"x-extra-stuff": dict{},
			},
		},
		"networks": dict{
			"bar": dict{
				"x-extra-stuff": dict{},
			},
		},
		"configs": dict{
			"bar": dict{
				"x-extra-stuff": dict{},
			},
		},
		"secrets": dict{
			"bar": dict{
				"x-extra-stuff": dict{},
			},
		},
	}
	err := Validate(config, "3.7")
	assert.NilError(t, err)
}

func TestValidateCredentialSpecs(t *testing.T) {
	tests := []struct {
		version     string
		expectedErr string
	}{
		{version: "3.0", expectedErr: "credential_spec"},
		{version: "3.1", expectedErr: "credential_spec"},
		{version: "3.2", expectedErr: "credential_spec"},
		{version: "3.3", expectedErr: "config"},
		{version: "3.4", expectedErr: "config"},
		{version: "3.5", expectedErr: "config"},
		{version: "3.6", expectedErr: "config"},
		{version: "3.7", expectedErr: "config"},
		{version: "3.8"},
		{version: "3.9"},
		{version: "3.10"},
		{version: "3.11"},
		{version: "3.12"},
		{version: "3.13"},
		{version: "3"},
		{version: ""},
	}

	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			config := dict{
				"version": "99.99",
				"services": dict{
					"foo": dict{
						"image": "busybox",
						"credential_spec": dict{
							"config": "foobar",
						},
					},
				},
			}
			err := Validate(config, tc.version)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, fmt.Sprintf("Additional property %s is not allowed", tc.expectedErr))
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestValidateSecretConfigNames(t *testing.T) {
	config := dict{
		"version": "3.5",
		"configs": dict{
			"bar": dict{
				"name": "foobar",
			},
		},
		"secrets": dict{
			"baz": dict{
				"name": "foobaz",
			},
		},
	}

	err := Validate(config, "3.5")
	assert.NilError(t, err)
}

func TestValidateInvalidVersion(t *testing.T) {
	config := dict{
		"version": "2.1",
		"services": dict{
			"foo": dict{
				"image": "busybox",
			},
		},
	}

	err := Validate(config, "2.1")
	assert.ErrorContains(t, err, "unsupported Compose file version: 2.1")
}

type array []any

func TestValidatePlacement(t *testing.T) {
	config := dict{
		"version": "3.3",
		"services": dict{
			"foo": dict{
				"image": "busybox",
				"deploy": dict{
					"placement": dict{
						"preferences": array{
							dict{
								"spread": "node.labels.az",
							},
						},
					},
				},
			},
		},
	}

	assert.NilError(t, Validate(config, "3.3"))
}

func TestValidateIsolation(t *testing.T) {
	config := dict{
		"version": "3.5",
		"services": dict{
			"foo": dict{
				"image":     "busybox",
				"isolation": "some-isolation-value",
			},
		},
	}
	assert.NilError(t, Validate(config, "3.5"))
}

func TestValidateRollbackConfig(t *testing.T) {
	config := dict{
		"version": "3.4",
		"services": dict{
			"foo": dict{
				"image": "busybox",
				"deploy": dict{
					"rollback_config": dict{
						"parallelism": 1,
					},
				},
			},
		},
	}

	assert.NilError(t, Validate(config, "3.7"))
}

func TestValidateRollbackConfigWithOrder(t *testing.T) {
	config := dict{
		"version": "3.4",
		"services": dict{
			"foo": dict{
				"image": "busybox",
				"deploy": dict{
					"rollback_config": dict{
						"parallelism": 1,
						"order":       "start-first",
					},
				},
			},
		},
	}

	assert.NilError(t, Validate(config, "3.7"))
}

func TestValidateRollbackConfigWithUpdateConfig(t *testing.T) {
	config := dict{
		"version": "3.4",
		"services": dict{
			"foo": dict{
				"image": "busybox",
				"deploy": dict{
					"update_config": dict{
						"parallelism": 1,
						"order":       "start-first",
					},
					"rollback_config": dict{
						"parallelism": 1,
						"order":       "start-first",
					},
				},
			},
		},
	}

	assert.NilError(t, Validate(config, "3.7"))
}

func TestValidateRollbackConfigWithUpdateConfigFull(t *testing.T) {
	config := dict{
		"version": "3.4",
		"services": dict{
			"foo": dict{
				"image": "busybox",
				"deploy": dict{
					"update_config": dict{
						"parallelism":    1,
						"order":          "start-first",
						"delay":          "10s",
						"failure_action": "pause",
						"monitor":        "10s",
					},
					"rollback_config": dict{
						"parallelism":    1,
						"order":          "start-first",
						"delay":          "10s",
						"failure_action": "pause",
						"monitor":        "10s",
					},
				},
			},
		},
	}

	assert.NilError(t, Validate(config, "3.7"))
}
