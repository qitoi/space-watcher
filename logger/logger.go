/*
 *  Copyright 2021 qitoi
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	info  WriteSyncReopener
	error WriteSyncReopener
}

func New(info, error WriteSyncReopener) *Logger {
	highPriority := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.WarnLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level < zapcore.WarnLevel
	})

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	stdoutSyncer := zapcore.Lock(info)
	stderrSyncer := zapcore.Lock(error)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, stderrSyncer, highPriority),
		zapcore.NewCore(encoder, stdoutSyncer, lowPriority),
	)

	logger := &Logger{
		Logger: zap.New(core),
		info:   info,
		error:  error,
	}

	return logger
}

func (l *Logger) Reopen() error {
	if err := l.info.Reopen(); err != nil {
		return err
	}
	if err := l.error.Reopen(); err != nil {
		return err
	}
	return nil
}
