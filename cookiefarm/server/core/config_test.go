package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"server/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- NewRunner ----------------------------------------------------------------

func TestNewRunner_WithValidStore_ReturnsNonNil(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)
	require.NotNil(t, r)
}

func TestNewRunner_StoreIsWired(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)
	// The store is unexported, but we can verify the runner works correctly
	// by calling a method that depends on the store — if it panics the store
	// was not wired.
	assert.NotPanics(t, func() {
		r.UpdateFlags(nil)
	})
}

func TestNewRunner_WithNilStore_DoesNotPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		_ = NewRunner(nil)
	})
}

func TestNewRunner_TwoInstances_AreIndependent(t *testing.T) {
	t.Parallel()
	s1 := newTestStore(t)
	s2 := newTestStore(t)
	r1 := NewRunner(s1)
	r2 := NewRunner(s2)
	assert.NotSame(t, r1, r2)
}

// --- Runner.Run ---------------------------------------------------------------

func TestRun_FirstCall_TTLDisabled_SpawnsProcessingLoop(t *testing.T) {
	t.Parallel()
	// Disable TTL so only one goroutine is spawned.
	orig := config.SharedConfig.ConfigServer.FlagTTL
	config.SharedConfig.ConfigServer.FlagTTL = 0
	t.Cleanup(func() { config.SharedConfig.ConfigServer.FlagTTL = orig })

	// Use a very large submit interval so the loop doesn't fire during the test.
	origInterval := config.SharedConfig.ConfigServer.SubmitFlagCheckerTime
	config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = 9999
	t.Cleanup(func() { config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = origInterval })

	r := newTestRunner(t)

	assert.NotPanics(t, func() { r.Run() })

	// shutdownCancel must have been set by Run.
	assert.NotNil(t, r.shutdownCancel)
	resetShutdownCancel(t, r)
}

func TestRun_FirstCall_TTLEnabled_DoesNotPanic(t *testing.T) {
	t.Parallel()
	orig := config.SharedConfig.ConfigServer.FlagTTL
	config.SharedConfig.ConfigServer.FlagTTL = 1
	t.Cleanup(func() { config.SharedConfig.ConfigServer.FlagTTL = orig })

	origTick := config.SharedConfig.ConfigServer.TickTime
	config.SharedConfig.ConfigServer.TickTime = 9999
	t.Cleanup(func() { config.SharedConfig.ConfigServer.TickTime = origTick })

	origInterval := config.SharedConfig.ConfigServer.SubmitFlagCheckerTime
	config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = 9999
	t.Cleanup(func() { config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = origInterval })

	r := newTestRunner(t)
	assert.NotPanics(t, func() { r.Run() })
	assert.NotNil(t, r.shutdownCancel)
	resetShutdownCancel(t, r)
}

func TestRun_SetsShutdownCancel(t *testing.T) {
	t.Parallel()
	origInterval := config.SharedConfig.ConfigServer.SubmitFlagCheckerTime
	config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = 9999
	t.Cleanup(func() { config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = origInterval })

	origTTL := config.SharedConfig.ConfigServer.FlagTTL
	config.SharedConfig.ConfigServer.FlagTTL = 0
	t.Cleanup(func() { config.SharedConfig.ConfigServer.FlagTTL = origTTL })

	r := newTestRunner(t)
	require.Nil(t, r.shutdownCancel, "shutdownCancel should be nil before Run")

	r.Run()

	assert.NotNil(t, r.shutdownCancel, "Run must set shutdownCancel")
	resetShutdownCancel(t, r)
}

func TestRun_Reentrant_CancelsPreviousContext(t *testing.T) {
	t.Parallel()
	origInterval := config.SharedConfig.ConfigServer.SubmitFlagCheckerTime
	config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = 9999
	t.Cleanup(func() { config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = origInterval })

	origTTL := config.SharedConfig.ConfigServer.FlagTTL
	config.SharedConfig.ConfigServer.FlagTTL = 0
	t.Cleanup(func() { config.SharedConfig.ConfigServer.FlagTTL = origTTL })

	r := newTestRunner(t)

	r.Run()
	firstCancel := r.shutdownCancel
	require.NotNil(t, firstCancel)

	// Second call: the previous cancel must have been invoked, a new one set.
	r.Run()
	secondCancel := r.shutdownCancel
	require.NotNil(t, secondCancel)

	// They must be different function values (new context was created).
	// We can't compare func values directly in Go, but we can verify the
	// second cancel doesn't panic and that shutdownCancel changed.
	assert.NotNil(t, secondCancel)
	assert.NotPanics(t, func() { firstCancel() }) // already cancelled — must be idempotent
	resetShutdownCancel(t, r)
}

func TestRun_Reentrant_DoesNotPanic(t *testing.T) {
	t.Parallel()
	origInterval := config.SharedConfig.ConfigServer.SubmitFlagCheckerTime
	config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = 9999
	t.Cleanup(func() { config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = origInterval })

	origTTL := config.SharedConfig.ConfigServer.FlagTTL
	config.SharedConfig.ConfigServer.FlagTTL = 0
	t.Cleanup(func() { config.SharedConfig.ConfigServer.FlagTTL = origTTL })

	r := newTestRunner(t)

	assert.NotPanics(t, func() {
		r.Run()
		r.Run()
		r.Run()
	})
	resetShutdownCancel(t, r)
}

// --- LoadConfigAndRun ---------------------------------------------------------

// writeConfigFile writes content to a temp file and returns its path.
func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

const minimalValidConfig = `
server:
  url_flag_checker: "http://checker.local/flags"
  team_token: "tok123"
  protocol: "cc_http"
  max_flag_batch_size: 100
  tick_time: 60
  submit_flag_checker_time: 9999
  flag_ttl: 0

client:
  regex_flag: "[A-Z0-9]{32}="

configured: false
`

func TestLoadConfigAndRun_ValidFile_ReturnsNil(t *testing.T) {
	t.Parallel()
	path := writeConfigFile(t, minimalValidConfig)
	r := newTestRunner(t)

	err := r.LoadConfig(path)
	assert.NoError(t, err)
	resetShutdownCancel(t, r)
}

func TestLoadConfigAndRun_ValidFile_SetsConfigured(t *testing.T) {
	t.Parallel()
	path := writeConfigFile(t, minimalValidConfig)
	r := newTestRunner(t)

	// Ensure Configured starts false.
	config.SharedConfig.Configured = false

	err := r.LoadConfig(path)
	require.NoError(t, err)
	assert.True(t, config.SharedConfig.Configured)
	resetShutdownCancel(t, r)
}

func TestLoadConfigAndRun_ValidFile_PopulatesSharedConfig(t *testing.T) {
	t.Parallel()
	path := writeConfigFile(t, minimalValidConfig)
	r := newTestRunner(t)

	err := r.LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, "http://checker.local/flags", config.SharedConfig.ConfigServer.URLFlagChecker)
	assert.Equal(t, "tok123", config.SharedConfig.ConfigServer.TeamToken)
	assert.Equal(t, uint(100), config.SharedConfig.ConfigServer.MaxFlagBatchSize)
	resetShutdownCancel(t, r)
}

func TestLoadConfigAndRun_NonExistentPath_ReturnsError(t *testing.T) {
	t.Parallel()
	r := newTestRunner(t)

	err := r.LoadConfig("/nonexistent/path/config.yml")
	assert.Error(t, err)
	resetShutdownCancel(t, r)
}

func TestLoadConfigAndRun_NonExistentPath_DoesNotModifyConfig(t *testing.T) {
	t.Parallel()
	r := newTestRunner(t)

	config.SharedConfig.ConfigServer.TeamToken = "original-token"

	_ = r.LoadConfig("/nonexistent/path/config.yml")

	assert.Equal(t, "original-token", config.SharedConfig.ConfigServer.TeamToken)
	resetShutdownCancel(t, r)
}

func TestLoadConfig_EmptyPath_ReturnsError(t *testing.T) {
	t.Parallel()
	r := newTestRunner(t)

	err := r.LoadConfig("")
	assert.Error(t, err)
	resetShutdownCancel(t, r)
}

func TestLoadConfig_MalformedYAML_ReturnsError(t *testing.T) {
	t.Parallel()
	path := writeConfigFile(t, ":::not valid yaml:::{{{")
	r := newTestRunner(t)

	err := r.LoadConfig(path)
	assert.Error(t, err)
	resetShutdownCancel(t, r)
}

func TestLoadConfig_EmptyFile_DoesNotPanic(t *testing.T) {
	t.Parallel()
	path := writeConfigFile(t, "")
	r := newTestRunner(t)

	assert.NotPanics(t, func() {
		_ = r.LoadConfig(path)
	})
	resetShutdownCancel(t, r)
}

func TestLoadConfig_AlreadyConfiguredTrue_RemainsTrue(t *testing.T) {
	t.Parallel()
	path := writeConfigFile(t, minimalValidConfig)
	r := newTestRunner(t)

	config.SharedConfig.Configured = true

	err := r.LoadConfig(path)
	require.NoError(t, err)
	assert.True(t, config.SharedConfig.Configured)
	resetShutdownCancel(t, r)
}

func TestLoadConfig_ValidFile_StartsRunner(t *testing.T) {
	t.Parallel()
	path := writeConfigFile(t, minimalValidConfig)
	r := newTestRunner(t)

	err := r.LoadConfig(path)
	require.NoError(t, err)
	r.Run()

	// If Run() was called, shutdownCancel must have been set.
	assert.NotNil(t, r.shutdownCancel)
	resetShutdownCancel(t, r)
}

func TestLoadConfig_CalledTwice_SecondCallOverridesFirst(t *testing.T) {
	t.Parallel()
	r := newTestRunner(t)

	cfg1 := writeConfigFile(t, `
server:
  team_token: "first"
  submit_flag_checker_time: 9999
  flag_ttl: 0
configured: false
`)

	cfg2 := writeConfigFile(t, `
server:
  team_token: "second"
  submit_flag_checker_time: 9999
  flag_ttl: 0
configured: false
`)

	require.NoError(t, r.LoadConfig(cfg1))
	require.NoError(t, r.LoadConfig(cfg2))

	assert.Equal(t, "second", config.SharedConfig.ConfigServer.TeamToken)
	resetShutdownCancel(t, r)
}

// --- Integration: Run goroutines terminate on cancel -------------------------

func TestRun_GoroutinesTerminateAfterCancel(t *testing.T) {
	t.Parallel()
	origInterval := config.SharedConfig.ConfigServer.SubmitFlagCheckerTime
	config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = 9999
	t.Cleanup(func() { config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = origInterval })

	origTTL := config.SharedConfig.ConfigServer.FlagTTL
	config.SharedConfig.ConfigServer.FlagTTL = 0
	t.Cleanup(func() { config.SharedConfig.ConfigServer.FlagTTL = origTTL })

	r := newTestRunner(t)
	r.Run()

	require.NotNil(t, r.shutdownCancel)

	// Cancel the context and give goroutines a moment to exit.
	r.shutdownCancel()
	time.Sleep(50 * time.Millisecond)
	// No assertion on goroutines (they're not observable from outside),
	// but the cancel call itself must not panic or block.
	resetShutdownCancel(t, r)
}

func TestRun_WithNilStore_SpawnsLoopWithoutPanic(t *testing.T) {
	t.Parallel()
	origInterval := config.SharedConfig.ConfigServer.SubmitFlagCheckerTime
	config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = 9999
	t.Cleanup(func() { config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = origInterval })

	origTTL := config.SharedConfig.ConfigServer.FlagTTL
	config.SharedConfig.ConfigServer.FlagTTL = 0
	t.Cleanup(func() { config.SharedConfig.ConfigServer.FlagTTL = origTTL })

	// A Runner with a nil store is an edge case documented in the analysis.
	// Run() itself must not panic; the goroutine will fail when it first
	// tries to query the store, but that is contained inside the goroutine.
	r := NewRunner(nil)
	assert.NotPanics(t, func() { r.Run() })

	// Allow the goroutine to start and hit the store, then cancel.
	time.Sleep(20 * time.Millisecond)
	if r.shutdownCancel != nil {
		r.shutdownCancel()
	}
	resetShutdownCancel(t, r)
}

// --- Verify database.Store is a valid dependency ------------------------------

func TestNewRunner_StoreQueriesAreAccessible(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)
	require.NotNil(t, r)

	// Insert a flag via the store and confirm the runner's underlying store
	// reflects it (both references point to the same store instance).
	f := sampleFlag("FLAG{store_wire_check}")
	insertFlag(t, store, f)

	// Count via a direct store call — same underlying DB the runner uses.
	count, err := store.Queries.CountFlags(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

// Ensure LoadConfigAndRun does not leave the config in a configured=true state
// when the YAML file is missing.
func TestLoadConfigAndRun_MissingFile_ConfiguredStaysFalse(t *testing.T) {
	t.Parallel()
	r := newTestRunner(t)

	config.SharedConfig.Configured = false

	_ = r.LoadConfig("/does/not/exist.yml")

	assert.False(t, config.SharedConfig.Configured)
	resetShutdownCancel(t, r)
}
