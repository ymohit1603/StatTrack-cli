package heartbeat_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/gandarez/go-realpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithFormatting(t *testing.T) {
	opt := heartbeat.WithFormatting()

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		entity, err := filepath.Abs(hh[0].Entity)
		require.NoError(t, err)

		entity, err = realpath.Realpath(entity)
		require.NoError(t, err)

		if runtime.GOOS == "windows" {
			entity = windows.FormatFilePath(entity)
		}

		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity: entity,
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := handle(context.Background(), []heartbeat.Heartbeat{{
		Entity:     "testdata/main.go",
		EntityType: heartbeat.FileType,
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestFormat_NotFileType(t *testing.T) {
	tests := map[string]heartbeat.EntityType{
		"app":    heartbeat.AppType,
		"domain": heartbeat.DomainType,
		"event":  heartbeat.EventType,
		"url":    heartbeat.URLType,
	}

	for name, entityType := range tests {
		t.Run(name, func(t *testing.T) {
			h := heartbeat.Heartbeat{
				Entity:     "/unmodified",
				EntityType: entityType,
			}

			formatted := heartbeat.Format(context.Background(), h)

			assert.Equal(t, "/unmodified", formatted.Entity)
		})
	}
}

func TestFormat_WindowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping because OS is not windows.")
	}

	h := heartbeat.Heartbeat{
		Entity:     `C:\Users\project\main.go`,
		EntityType: heartbeat.FileType,
	}

	formatted := heartbeat.Format(context.Background(), h)

	assert.Equal(t, "C:/Users/project/main.go", formatted.Entity)
}

func TestFormat_NetworkMount(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping because OS is not windows.")
	}

	h := heartbeat.Heartbeat{
		Entity:     `\\192.168.1.1\apilibrary.sl`,
		EntityType: heartbeat.FileType,
	}

	r := heartbeat.Format(context.Background(), h)

	assert.Equal(t, heartbeat.Heartbeat{
		Entity:     `\\192.168.1.1/apilibrary.sl`,
		EntityType: heartbeat.FileType,
	}, r)
}
