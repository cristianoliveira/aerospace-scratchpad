package constants

// Define here all environment variables used in the application

const (
	// EnvAeroSpaceScratchpadLogsPath is the environment variable for the AeroSpace marks logs path
	// default: `/tmp/aerospace-scratchpad.log`
	EnvAeroSpaceScratchpadLogsPath string = "AEROSPACE_SCRATCHPAD_LOGS_PATH"

	// EnvAeroSpaceScratchpadLogsLevel is the environment variable for the AeroSpace marks logs level
	// default: `DISABLED`
	EnvAeroSpaceScratchpadLogsLevel string = "AEROSPACE_SCRATCHPAD_LOGS_LEVEL"

	// EnvAeroSpaceSock is the environment variable for the AeroSpace IPC socket path
	EnvAeroSpaceSock string = "AEROSPACESOCK"
)
