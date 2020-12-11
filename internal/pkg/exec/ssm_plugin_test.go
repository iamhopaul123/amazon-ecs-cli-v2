// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/copilot-cli/internal/pkg/exec/mocks"
	"github.com/aws/copilot-cli/internal/pkg/new-sdk-go/ecs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestSSMPlugin_StartSession(t *testing.T) {
	const (
		mockLatestVersion  = "mockLatestVersion"
		mockCurrentVersion = "mockCurrentVersion"
	)
	mockSession := &ecs.Session{
		SessionId:  aws.String("mockSessionID"),
		StreamUrl:  aws.String("mockStreamURL"),
		TokenValue: aws.String("mockTokenValue"),
	}
	var mockRunner *mocks.Mockrunner
	var mockPrompter *mocks.Mockprompter
	mockError := errors.New("some error")
	tests := map[string]struct {
		inLatestVersion  string
		inCurrentVersion string
		inSession        *ecs.Session
		setupMocks       func(controller *gomock.Controller)
		wantedError      error
	}{
		"return error if fail to get the latest version": {
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(mockError)
			},
			wantedError: fmt.Errorf("get ssm plugin latest version: some error"),
		},
		"return error if fail to get the current version": {
			inLatestVersion: mockLatestVersion,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(mockError)
			},
			wantedError: fmt.Errorf("get local ssm plugin version: some error"),
		},
		"return error if fail to prompt to confirm installing the plugin": {
			inLatestVersion: mockLatestVersion,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(errors.New("executable file not found in $PATH"))
				mockPrompter.EXPECT().Confirm(ssmPluginInstallPrompt, ssmPluginInstallPromptHelp).
					Return(false, mockError)
			},
			wantedError: fmt.Errorf("prompt to confirm installing the plugin: some error"),
		},
		"return error if fail to confirm to install": {
			inLatestVersion: mockLatestVersion,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(errors.New("executable file not found in $PATH"))
				mockPrompter.EXPECT().Confirm(ssmPluginInstallPrompt, ssmPluginInstallPromptHelp).
					Return(false, nil)
			},
			wantedError: errSSMPluginInstallCancelled,
		},
		"return error if fail to install binary": {
			inLatestVersion: mockLatestVersion,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(errors.New("executable file not found in $PATH"))
				mockPrompter.EXPECT().Confirm(ssmPluginInstallPrompt, ssmPluginInstallPromptHelp).
					Return(true, nil)
				mockRunner.EXPECT().Run("curl", []string{"https://s3.amazonaws.com/session-manager-downloads/plugin/latest/mac/sessionmanager-bundle.zip",
					"-o", "sessionmanager-bundle.zip"}).
					Return(mockError)
			},
			wantedError: fmt.Errorf("install binary: some error"),
		},
		"return error if fail to prompt to confirm updating binary": {
			inLatestVersion:  mockLatestVersion,
			inCurrentVersion: mockCurrentVersion,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(nil)
				mockPrompter.EXPECT().Confirm(fmt.Sprintf(ssmPluginUpdatePrompt, mockCurrentVersion, mockLatestVersion),
					ssmPluginUpdatePromptHelp).
					Return(false, mockError)
			},
			wantedError: fmt.Errorf("prompt to confirm updating the plugin: some error"),
		},
		"return error if fail to confirm to update binary": {
			inLatestVersion:  mockLatestVersion,
			inCurrentVersion: mockCurrentVersion,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(nil)
				mockPrompter.EXPECT().Confirm(fmt.Sprintf(ssmPluginUpdatePrompt, mockCurrentVersion, mockLatestVersion),
					ssmPluginUpdatePromptHelp).
					Return(false, nil)
			},
			wantedError: errSSMPluginUpdateCancelled,
		},
		"return error if fail to update binary": {
			inLatestVersion:  mockLatestVersion,
			inCurrentVersion: mockCurrentVersion,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(nil)
				mockPrompter.EXPECT().Confirm(fmt.Sprintf(ssmPluginUpdatePrompt, mockCurrentVersion, mockLatestVersion),
					ssmPluginUpdatePromptHelp).
					Return(true, nil)
				mockRunner.EXPECT().Run("curl", []string{"https://s3.amazonaws.com/session-manager-downloads/plugin/latest/mac/sessionmanager-bundle.zip",
					"-o", "sessionmanager-bundle.zip"}).
					Return(mockError)
			},
			wantedError: fmt.Errorf("update binary: some error"),
		},
		"return error if fail to start session": {
			inLatestVersion:  mockLatestVersion,
			inCurrentVersion: mockLatestVersion,
			inSession:        mockSession,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{`{"SessionId":"mockSessionID","StreamUrl":"mockStreamURL","TokenValue":"mockTokenValue"}`, "us-west-2", "StartSession"},
					gomock.Any(), gomock.Any(), gomock.Any()).Return(mockError)
			},
			wantedError: fmt.Errorf("start session: some error"),
		},
		"success with no update and no install": {
			inLatestVersion:  mockLatestVersion,
			inCurrentVersion: mockLatestVersion,
			inSession:        mockSession,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{`{"SessionId":"mockSessionID","StreamUrl":"mockStreamURL","TokenValue":"mockTokenValue"}`, "us-west-2", "StartSession"},
					gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		"success with install": {
			inLatestVersion: mockLatestVersion,
			inSession:       mockSession,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(errors.New("executable file not found in $PATH"))
				mockPrompter.EXPECT().Confirm(ssmPluginInstallPrompt, ssmPluginInstallPromptHelp).
					Return(true, nil)
				mockRunner.EXPECT().Run("curl", []string{"https://s3.amazonaws.com/session-manager-downloads/plugin/latest/mac/sessionmanager-bundle.zip",
					"-o", "sessionmanager-bundle.zip"}).
					Return(nil)
				mockRunner.EXPECT().Run("unzip", []string{"-o", "sessionmanager-bundle.zip"}).
					Return(nil)
				mockRunner.EXPECT().Run("sudo", []string{"./sessionmanager-bundle/install", "-i",
					"/usr/local/sessionmanagerplugin", "-b",
					"/usr/local/bin/session-manager-plugin"}).
					Return(nil)
				mockRunner.EXPECT().Run("rm", []string{"-r", "./sessionmanager-bundle", "sessionmanager-bundle.zip"}).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{`{"SessionId":"mockSessionID","StreamUrl":"mockStreamURL","TokenValue":"mockTokenValue"}`, "us-west-2", "StartSession"},
					gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		"success with update": {
			inLatestVersion:  mockLatestVersion,
			inCurrentVersion: mockCurrentVersion,
			inSession:        mockSession,
			setupMocks: func(controller *gomock.Controller) {
				mockRunner = mocks.NewMockrunner(controller)
				mockPrompter = mocks.NewMockprompter(controller)
				mockRunner.EXPECT().Run("curl", []string{"-s", ssmPluginBinaryLatestVersionURL}, gomock.Any()).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{"--version"}, gomock.Any()).
					Return(nil)
				mockPrompter.EXPECT().Confirm(fmt.Sprintf(ssmPluginUpdatePrompt, mockCurrentVersion, mockLatestVersion),
					ssmPluginUpdatePromptHelp).
					Return(true, nil)
				mockRunner.EXPECT().Run("curl", []string{"https://s3.amazonaws.com/session-manager-downloads/plugin/latest/mac/sessionmanager-bundle.zip",
					"-o", "sessionmanager-bundle.zip"}).
					Return(nil)
				mockRunner.EXPECT().Run("unzip", []string{"-o", "sessionmanager-bundle.zip"}).
					Return(nil)
				mockRunner.EXPECT().Run("sudo", []string{"./sessionmanager-bundle/install", "-i",
					"/usr/local/sessionmanagerplugin", "-b",
					"/usr/local/bin/session-manager-plugin"}).
					Return(nil)
				mockRunner.EXPECT().Run("rm", []string{"-r", "./sessionmanager-bundle", "sessionmanager-bundle.zip"}).
					Return(nil)
				mockRunner.EXPECT().Run(ssmPluginBinaryName, []string{`{"SessionId":"mockSessionID","StreamUrl":"mockStreamURL","TokenValue":"mockTokenValue"}`, "us-west-2", "StartSession"},
					gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tc.setupMocks(ctrl)
			s := SSMPlugin{
				runner:   mockRunner,
				prompter: mockPrompter,
				sess: &session.Session{
					Config: &aws.Config{
						Region: aws.String("us-west-2"),
					},
				},
				currentVersionBuffer: *bytes.NewBufferString(tc.inCurrentVersion),
				latestVersionBuffer:  *bytes.NewBufferString(tc.inLatestVersion),
			}
			err := s.StartSession(tc.inSession)
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
