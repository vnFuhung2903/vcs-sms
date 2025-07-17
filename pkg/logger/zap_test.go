package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"go.uber.org/zap"
)

type LoggerSuite struct {
	suite.Suite
	tempDir    string
	testLogger *Logger
	logBuffer  *bytes.Buffer
}

func (suite *LoggerSuite) SetupSuite() {
	tempDir, err := os.MkdirTemp("", "zap_test")
	suite.Require().NoError(err)
	suite.tempDir = tempDir
}

func (suite *LoggerSuite) TearDownSuite() {
	os.RemoveAll(suite.tempDir)
}

func (suite *LoggerSuite) SetupTest() {
	once = sync.Once{}
	suite.logBuffer = &bytes.Buffer{}
}

func (suite *LoggerSuite) TearDownTest() {
	if suite.testLogger != nil {
		suite.testLogger.Sync()
	}
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}

func (suite *LoggerSuite) TestLoadLogger() {
	loggerEnv := env.LoggerEnv{
		Level:      "info",
		FilePath:   filepath.Join(suite.tempDir, "test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := LoadLogger(loggerEnv)

	suite.NoError(err)
	suite.NotNil(logger)
	suite.IsType(&Logger{}, logger)
	suite.NotNil(logger.logger)
}

func (suite *LoggerSuite) TestLoadLoggerInvalidLevel() {
	loggerEnv := env.LoggerEnv{
		Level:      "invalid_level",
		FilePath:   filepath.Join(suite.tempDir, "test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := LoadLogger(loggerEnv)

	suite.Error(err)
	suite.Nil(logger)
	suite.Contains(err.Error(), "unrecognized level")
}

func (suite *LoggerSuite) TestInitLogger() {
	loggerEnv := env.LoggerEnv{
		Level:      "warn",
		FilePath:   filepath.Join(suite.tempDir, "init_test.log"),
		MaxSize:    20,
		MaxAge:     14,
		MaxBackups: 5,
	}

	logger, err := initLogger(loggerEnv)

	suite.NoError(err)
	suite.NotNil(logger)
	suite.IsType(&Logger{}, logger)
}

func (suite *LoggerSuite) TestInitLogger_InvalidLogPath() {
	invalidPath := "/root/cannot_write_here.log"

	loggerEnv := env.LoggerEnv{
		Level:      "info",
		FilePath:   invalidPath,
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := initLogger(loggerEnv)

	if err != nil {
		suite.Error(err)
		suite.Nil(logger)
	} else {
		suite.NoError(err)
		suite.NotNil(logger)
		suite.testLogger = logger
	}
}

func (suite *LoggerSuite) TestLogger_Debug() {
	loggerEnv := env.LoggerEnv{
		Level:      "debug",
		FilePath:   filepath.Join(suite.tempDir, "debug_test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := initLogger(loggerEnv)
	suite.Require().NoError(err)
	suite.testLogger = logger

	logger.Debug("Debug message", zap.String("key", "value"))

	err = logger.Sync()
	suite.NoError(err)

	logContent, err := os.ReadFile(loggerEnv.FilePath)
	suite.NoError(err)
	suite.Contains(string(logContent), "Debug message")
	suite.Contains(string(logContent), "debug")
}

func (suite *LoggerSuite) TestLogger_Info() {
	loggerEnv := env.LoggerEnv{
		Level:      "info",
		FilePath:   filepath.Join(suite.tempDir, "info_test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := initLogger(loggerEnv)
	suite.Require().NoError(err)
	suite.testLogger = logger

	logger.Info("Info message", zap.Int("count", 42))

	err = logger.Sync()
	suite.NoError(err)

	logContent, err := os.ReadFile(loggerEnv.FilePath)
	suite.NoError(err)
	suite.Contains(string(logContent), "Info message")
	suite.Contains(string(logContent), "info")
	suite.Contains(string(logContent), "42")
}

func (suite *LoggerSuite) TestLogger_Warn() {
	loggerEnv := env.LoggerEnv{
		Level:      "warn",
		FilePath:   filepath.Join(suite.tempDir, "warn_test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := initLogger(loggerEnv)
	suite.Require().NoError(err)
	suite.testLogger = logger

	logger.Warn("Warning message", zap.Bool("important", true))

	err = logger.Sync()
	suite.NoError(err)

	logContent, err := os.ReadFile(loggerEnv.FilePath)
	suite.NoError(err)
	suite.Contains(string(logContent), "Warning message")
	suite.Contains(string(logContent), "warn")
}

func (suite *LoggerSuite) TestLogger_Error() {
	loggerEnv := env.LoggerEnv{
		Level:      "error",
		FilePath:   filepath.Join(suite.tempDir, "error_test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := initLogger(loggerEnv)
	suite.Require().NoError(err)
	suite.testLogger = logger

	logger.Error("Error message", zap.Error(err))

	err = logger.Sync()
	suite.NoError(err)

	logContent, err := os.ReadFile(loggerEnv.FilePath)
	suite.NoError(err)
	suite.Contains(string(logContent), "Error message")
	suite.Contains(string(logContent), "error")
}

func (suite *LoggerSuite) TestLogger_With() {
	loggerEnv := env.LoggerEnv{
		Level:      "info",
		FilePath:   filepath.Join(suite.tempDir, "with_test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := initLogger(loggerEnv)
	suite.Require().NoError(err)
	suite.testLogger = logger

	childLogger := logger.With(zap.String("module", "test"), zap.Int("version", 1))

	suite.Implements((*ILogger)(nil), childLogger)

	childLogger.Info("Message with context")

	err = childLogger.Sync()
	suite.NoError(err)

	logContent, err := os.ReadFile(loggerEnv.FilePath)
	suite.NoError(err)

	logStr := string(logContent)
	suite.Contains(logStr, "Message with context")
	suite.Contains(logStr, "module")
	suite.Contains(logStr, "test")
	suite.Contains(logStr, "version")
}

func (suite *LoggerSuite) TestLogger_Sync() {
	loggerEnv := env.LoggerEnv{
		Level:      "info",
		FilePath:   filepath.Join(suite.tempDir, "sync_test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := initLogger(loggerEnv)
	suite.Require().NoError(err)
	suite.testLogger = logger

	logger.Info("Sync test message")

	err = logger.Sync()
	suite.NoError(err)

	logContent, err := os.ReadFile(loggerEnv.FilePath)
	suite.NoError(err)
	suite.Contains(string(logContent), "Sync test message")
}

func (suite *LoggerSuite) TestLogger_StacktraceOnError() {
	loggerEnv := env.LoggerEnv{
		Level:      "error",
		FilePath:   filepath.Join(suite.tempDir, "stacktrace_test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	}

	logger, err := initLogger(loggerEnv)
	suite.Require().NoError(err)
	suite.testLogger = logger

	logger.Error("Error with stacktrace")

	err = logger.Sync()
	suite.NoError(err)

	logContent, err := os.ReadFile(loggerEnv.FilePath)
	suite.NoError(err)

	logStr := string(logContent)
	suite.Contains(logStr, "stacktrace")
}
