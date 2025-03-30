package heartbeat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	apicmd "github.com/wakatime/wakatime-cli/cmd/api"
	offlinecmd "github.com/wakatime/wakatime-cli/cmd/offline"
	paramscmd "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/backoff"
	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/filestats"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/language"
	_ "github.com/wakatime/wakatime-cli/pkg/lexer" // force to load all lexers
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/remote"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// Run executes the heartbeat command.
func Run(ctx context.Context, v *viper.Viper) (int, error) {
	logger := log.Extract(ctx)

	queueFilepath, err := offline.QueueFilepath(ctx, v)
	if err != nil {
		logger.Warnf("failed to load offline queue filepath: %s", err)
	}

	err = SendHeartbeats(ctx, v, queueFilepath)
	if err != nil {
		var errauth api.ErrAuth

		// api.ErrAuth represents an error when parsing api key or timeout.
		// Save heartbeats to offline db when api.ErrAuth as it avoids losing heartbeats.
		if errors.As(err, &errauth) {
			if err := offlinecmd.SaveHeartbeats(ctx, v, nil, queueFilepath); err != nil {
				logger.Errorf("failed to save heartbeats to offline queue: %s", err)
			}

			return errauth.ExitCode(), fmt.Errorf("sending heartbeat(s) failed: %w", errauth)
		}

		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("sending heartbeat(s) failed: %w", errwaka)
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"sending heartbeat(s) failed: %w",
			err,
		)
	}

	logger.Debugln("successfully sent heartbeat(s)")

	return exitcode.Success, nil
}

// SendHeartbeats sends a heartbeat to the wakatime api and includes additional
// heartbeats from the offline queue, if available and offline sync is not
// explicitly disabled.
func SendHeartbeats(ctx context.Context, v *viper.Viper, queueFilepath string) error {
	params, err := LoadParams(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	logger := log.Extract(ctx)

	setLogFields(ctx, params)
	logger.Debugf("params: %s", params)

	if RateLimited(RateLimitParams{
		Disabled:   params.Offline.Disabled,
		LastSentAt: params.Offline.LastSentAt,
		Timeout:    params.Offline.RateLimit,
	}) {
		if err = offlinecmd.SaveHeartbeats(ctx, v, nil, queueFilepath); err == nil {
			return nil
		}

		// log offline db error then try to send heartbeats to API so they're not lost
		logger.Errorf("failed to save rate limited heartbeats: %s", err)
	}

	heartbeats := buildHeartbeats(ctx, params)

	var chOfflineSave = make(chan bool)

	// only send at once the maximum amount of `offline.SendLimit`.
	if len(heartbeats) > offline.SendLimit {
		extraHeartbeats := heartbeats[offline.SendLimit:]

		logger.Debugf("save %d extra heartbeat(s) to offline queue", len(extraHeartbeats))

		go func(done chan<- bool) {
			if err := offlinecmd.SaveHeartbeats(ctx, v, extraHeartbeats, queueFilepath); err != nil {
				logger.Errorf("failed to save extra heartbeats to offline queue: %s", err)
			}

			done <- true
		}(chOfflineSave)

		heartbeats = heartbeats[:offline.SendLimit]
	}

	handleOpts := initHandleOptions(params)

	if !params.Offline.Disabled {
		handleOpts = append(handleOpts, offline.WithQueue(queueFilepath))
	}

	handleOpts = append(handleOpts, backoff.WithBackoff(backoff.Config{
		V:        v,
		At:       params.API.BackoffAt,
		Retries:  params.API.BackoffRetries,
		HasProxy: params.API.ProxyURL != "",
	}))

	apiClient, err := apicmd.NewClientWithoutAuth(ctx, params.API)
	if err != nil {
		if !params.Offline.Disabled {
			if err := offlinecmd.SaveHeartbeats(ctx, v, heartbeats, queueFilepath); err != nil {
				logger.Errorf("failed to save heartbeats to offline queue: %s", err)
			}
		}

		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	handle := heartbeat.NewHandle(apiClient, handleOpts...)
	results, err := handle(ctx, heartbeats)

	// wait for offline queue save to finish
	if len(heartbeats) > offline.SendLimit {
		<-chOfflineSave
	}

	if err != nil {
		return err
	}

	for _, result := range results {
		if len(result.Errors) > 0 {
			logger.Warnln(strings.Join(result.Errors, " "))
		}
	}

	if err := ResetRateLimit(ctx, v); err != nil {
		logger.Errorf("failed to reset rate limit: %s", err)
	}

	return nil
}

// LoadParams loads params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(ctx context.Context, v *viper.Viper) (paramscmd.Params, error) {
	if v == nil {
		return paramscmd.Params{}, errors.New("viper instance unset")
	}

	apiParams, err := paramscmd.LoadAPIParams(ctx, v)
	if err != nil {
		return paramscmd.Params{}, fmt.Errorf("failed to load API parameters: %w", err)
	}

	heartbeatParams, err := paramscmd.LoadHeartbeatParams(ctx, v)
	if err != nil {
		return paramscmd.Params{}, fmt.Errorf("failed to load heartbeat params: %s", err)
	}

	return paramscmd.Params{
		API:       apiParams,
		Heartbeat: heartbeatParams,
		Offline:   paramscmd.LoadOfflineParams(ctx, v),
	}, nil
}

// RateLimitParams contains params for the RateLimited function.
type RateLimitParams struct {
	Disabled   bool
	LastSentAt time.Time
	Timeout    time.Duration
}

// RateLimited determines if we should send heartbeats to the API or save to the offline db.
func RateLimited(params RateLimitParams) bool {
	if params.Disabled {
		return false
	}

	if params.Timeout == 0 {
		return false
	}

	if params.LastSentAt.IsZero() {
		return false
	}

	return time.Since(params.LastSentAt) < params.Timeout
}

// ResetRateLimit updates the internal.heartbeats_last_sent_at timestamp.
func ResetRateLimit(ctx context.Context, v *viper.Viper) error {
	w, err := ini.NewWriter(ctx, v, ini.InternalFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %s", err)
	}

	keyValue := map[string]string{
		"heartbeats_last_sent_at": time.Now().Format(ini.DateFormat),
	}

	if err := w.Write(ctx, "internal", keyValue); err != nil {
		return fmt.Errorf("failed to write to internal config file: %s", err)
	}

	return nil
}

func buildHeartbeats(ctx context.Context, params paramscmd.Params) []heartbeat.Heartbeat {
	heartbeats := []heartbeat.Heartbeat{}

	userAgent := heartbeat.UserAgent(ctx, params.API.Plugin)

	heartbeats = append(heartbeats, heartbeat.New(
		params.Heartbeat.Project.BranchAlternate,
		params.Heartbeat.Category,
		params.Heartbeat.CursorPosition,
		params.Heartbeat.Entity,
		params.Heartbeat.EntityType,
		params.Heartbeat.IsUnsavedEntity,
		params.Heartbeat.IsWrite,
		params.Heartbeat.Language,
		params.Heartbeat.LanguageAlternate,
		params.Heartbeat.LineAdditions,
		params.Heartbeat.LineDeletions,
		params.Heartbeat.LineNumber,
		params.Heartbeat.LinesInFile,
		params.Heartbeat.LocalFile,
		params.Heartbeat.Project.Alternate,
		params.Heartbeat.Project.ProjectFromGitRemote,
		params.Heartbeat.Project.Override,
		params.Heartbeat.Sanitize.ProjectPathOverride,
		params.Heartbeat.Time,
		userAgent,
	))

	if len(params.Heartbeat.ExtraHeartbeats) > 0 {
		logger := log.Extract(ctx)
		logger.Debugf("include %d extra heartbeat(s) from stdin", len(params.Heartbeat.ExtraHeartbeats))

		for _, h := range params.Heartbeat.ExtraHeartbeats {
			heartbeats = append(heartbeats, heartbeat.New(
				h.BranchAlternate,
				h.Category,
				h.CursorPosition,
				h.Entity,
				h.EntityType,
				h.IsUnsavedEntity,
				h.IsWrite,
				h.Language,
				h.LanguageAlternate,
				h.LineAdditions,
				h.LineDeletions,
				h.LineNumber,
				h.Lines,
				h.LocalFile,
				h.ProjectAlternate,
				h.ProjectFromGitRemote,
				h.ProjectOverride,
				h.ProjectPathOverride,
				h.Time,
				userAgent,
			))
		}
	}

	return heartbeats
}

func initHandleOptions(params paramscmd.Params) []heartbeat.HandleOption {
	return []heartbeat.HandleOption{
		heartbeat.WithFormatting(),
		heartbeat.WithEntityModifier(),
		filter.WithFiltering(filter.Config{
			Exclude:                    params.Heartbeat.Filter.Exclude,
			Include:                    params.Heartbeat.Filter.Include,
			IncludeOnlyWithProjectFile: params.Heartbeat.Filter.IncludeOnlyWithProjectFile,
		}),
		remote.WithDetection(),
		apikey.WithReplacing(apikey.Config{
			DefaultAPIKey: params.API.Key,
			MapPatterns:   params.API.KeyPatterns,
		}),
		filestats.WithDetection(),
		language.WithDetection(language.Config{
			GuessLanguage: params.Heartbeat.GuessLanguage,
		}),
		deps.WithDetection(deps.Config{
			FilePatterns: params.Heartbeat.Sanitize.HideFileNames,
		}),
		project.WithDetection(project.Config{
			HideProjectNames:     params.Heartbeat.Sanitize.HideProjectNames,
			MapPatterns:          params.Heartbeat.Project.MapPatterns,
			ProjectFromGitRemote: params.Heartbeat.Project.ProjectFromGitRemote,
			Submodule: project.Submodule{
				DisabledPatterns: params.Heartbeat.Project.SubmodulesDisabled,
				MapPatterns:      params.Heartbeat.Project.SubmoduleMapPatterns,
			},
		}),
		project.WithFiltering(project.FilterConfig{
			ExcludeUnknownProject: params.Heartbeat.Filter.ExcludeUnknownProject,
		}),
		heartbeat.WithSanitization(heartbeat.SanitizeConfig{
			BranchPatterns:     params.Heartbeat.Sanitize.HideBranchNames,
			DependencyPatterns: params.Heartbeat.Sanitize.HideDependencies,
			FilePatterns:       params.Heartbeat.Sanitize.HideFileNames,
			HideProjectFolder:  params.Heartbeat.Sanitize.HideProjectFolder,
			ProjectPatterns:    params.Heartbeat.Sanitize.HideProjectNames,
		}),
		remote.WithCleanup(),
		filter.WithLengthValidator(),
	}
}

func setLogFields(ctx context.Context, params paramscmd.Params) {
	log.AddField(ctx, "file", params.Heartbeat.Entity)
	log.AddField(ctx, "time", params.Heartbeat.Time)

	if params.API.Plugin != "" {
		log.AddField(ctx, "plugin", params.API.Plugin)
	}

	if params.Heartbeat.LineNumber != nil {
		log.AddField(ctx, "lineno", params.Heartbeat.LineNumber)
	}

	if params.Heartbeat.IsWrite != nil {
		log.AddField(ctx, "is_write", params.Heartbeat.IsWrite)
	}
}
