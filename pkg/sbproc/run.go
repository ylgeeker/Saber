/**
 * Copyright 2025 Saber authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package sbproc

import "context"

// Run runs the command with stdout/stderr ignored and waits for it to finish.
// Returns the exit code and any error.
func Run(ctx context.Context, cmd string, args ...string) (int, error) {
	child, err := Start(ctx, cmd, args...)
	if err != nil {
		return -1, err
	}

	return child.Wait()
}

// RunWithEnv runs the command with the given environment (merged over os.Environ()),
// ignores stdio, and waits for it to finish.
func RunWithEnv(ctx context.Context, env map[string]string, cmd string, args ...string) (int, error) {
	child, err := StartWithEnv(ctx, env, cmd, args...)
	if err != nil {
		return -1, err
	}

	return child.Wait()
}

// RunWithStdout runs the command and captures stdout; stderr is discarded.
// Returns captured stdout and any error (including non-zero exit).
func RunWithStdout(ctx context.Context, cmd string, args ...string) (stdout []byte, err error) {
	_, out, _, err := StartWithOutput(ctx, cmd, args...)
	return out, err
}

// RunWithStderr runs the command and captures stderr; stdout is discarded.
func RunWithStderr(ctx context.Context, cmd string, args ...string) (stderr []byte, err error) {
	_, _, errOut, err := StartWithOutput(ctx, cmd, args...)
	return errOut, err
}

// RunWithStdoutAndStderr runs the command and captures both stdout and stderr.
func RunWithStdoutAndStderr(ctx context.Context, cmd string, args ...string) (stdout []byte, stderr []byte, err error) {
	_, out, errOut, err := StartWithOutput(ctx, cmd, args...)
	if err != nil {
		return nil, nil, err
	}
	return out, errOut, nil
}
