// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/config"
	"github.com/aws/copilot-cli/internal/pkg/describe/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type envDescriberMocks struct {
	configStoreSvc *mocks.MockConfigStoreSvc
	deployStoreSvc *mocks.MockDeployStoreSvc
	stackDescriber *mocks.MockstackAndResourcesDescriber
}

var wantedResources = []*CfnResource{
	{
		Type:       "AWS::IAM::Role",
		PhysicalID: "testApp-testEnv-CFNExecutionRole",
	},
	{
		Type:       "testApp-testEnv-Cluster",
		PhysicalID: "AWS::ECS::Cluster-jI63pYBWU6BZ",
	},
}

func TestEnvDescriber_Describe(t *testing.T) {
	testApp := "testApp"
	testEnv := &config.Environment{
		App:              "testApp",
		Name:             "testEnv",
		Region:           "us-west-2",
		AccountID:        "123456789012",
		Prod:             false,
		RegistryURL:      "",
		ExecutionRoleARN: "",
		ManagerRoleARN:   "",
	}
	testSvc1 := &config.Service{
		App:  "testApp",
		Name: "testSvc1",
		Type: "load-balanced",
	}
	testSvc2 := &config.Service{
		App:  "testApp",
		Name: "testSvc2",
		Type: "load-balanced",
	}
	testSvc3 := &config.Service{
		App:  "testApp",
		Name: "testSvc3",
		Type: "load-balanced",
	}
	stackTags := []*cloudformation.Tag{
		{
			Key:   aws.String("copilot-application"),
			Value: aws.String("testApp"),
		},
		{
			Key:   aws.String("copilot-environment"),
			Value: aws.String("testEnv"),
		},
	}
	mockResource1 := &cloudformation.StackResource{
		PhysicalResourceId: aws.String("testApp-testEnv-CFNExecutionRole"),
		ResourceType:       aws.String("AWS::IAM::Role"),
	}
	mockResource2 := &cloudformation.StackResource{
		PhysicalResourceId: aws.String("AWS::ECS::Cluster-jI63pYBWU6BZ"),
		ResourceType:       aws.String("testApp-testEnv-Cluster"),
	}
	envSvcs := []*config.Service{testSvc1, testSvc2}
	mockError := errors.New("some error")
	testCases := map[string]struct {
		shouldOutputResources bool

		setupMocks func(mocks envDescriberMocks)

		wantedEnv   *EnvDescription
		wantedSvcs  []*config.Service
		wantedError error
	}{
		"error if fail to list all services": {
			setupMocks: func(m envDescriberMocks) {
				gomock.InOrder(
					m.configStoreSvc.EXPECT().ListServices(testApp).Return(nil, mockError),
				)
			},
			wantedError: fmt.Errorf("list services for app testApp: some error"),
		},
		"error if fail to list deployed services": {
			setupMocks: func(m envDescriberMocks) {
				gomock.InOrder(
					m.configStoreSvc.EXPECT().ListServices(testApp).Return([]*config.Service{
						testSvc1, testSvc2, testSvc3,
					}, nil),
					m.deployStoreSvc.EXPECT().ListDeployedServices(testApp, testEnv.Name).
						Return(nil, mockError),
				)
			},
			wantedError: fmt.Errorf("list deployed services in env testEnv: some error"),
		},
		"error if fail to get env tags": {
			setupMocks: func(m envDescriberMocks) {
				gomock.InOrder(
					m.configStoreSvc.EXPECT().ListServices(testApp).Return([]*config.Service{
						testSvc1, testSvc2, testSvc3,
					}, nil),
					m.deployStoreSvc.EXPECT().ListDeployedServices(testApp, testEnv.Name).
						Return([]string{"testSvc1", "testSvc2"}, nil),
					m.stackDescriber.EXPECT().Stack("testApp-testEnv").Return(nil, mockError),
				)
			},
			wantedError: fmt.Errorf("retrieve environment tags: some error"),
		},
		"error if fail to get env resources": {
			shouldOutputResources: true,
			setupMocks: func(m envDescriberMocks) {
				gomock.InOrder(
					m.configStoreSvc.EXPECT().ListServices(testApp).Return([]*config.Service{
						testSvc1, testSvc2, testSvc3,
					}, nil),
					m.deployStoreSvc.EXPECT().ListDeployedServices(testApp, testEnv.Name).
						Return([]string{"testSvc1", "testSvc2"}, nil),
					m.stackDescriber.EXPECT().Stack("testApp-testEnv").Return(&cloudformation.Stack{
						Tags: stackTags,
					}, nil),
					m.stackDescriber.EXPECT().StackResources("testApp-testEnv").Return(nil, mockError),
				)
			},
			wantedError: fmt.Errorf("retrieve environment resources: some error"),
		},
		"success without resources": {
			shouldOutputResources: false,
			setupMocks: func(m envDescriberMocks) {
				gomock.InOrder(
					m.configStoreSvc.EXPECT().ListServices(testApp).Return([]*config.Service{
						testSvc1, testSvc2, testSvc3,
					}, nil),
					m.deployStoreSvc.EXPECT().ListDeployedServices(testApp, testEnv.Name).
						Return([]string{"testSvc1", "testSvc2"}, nil),
					m.stackDescriber.EXPECT().Stack("testApp-testEnv").Return(&cloudformation.Stack{
						Tags: stackTags,
					}, nil),
				)
			},
			wantedEnv: &EnvDescription{
				Environment: testEnv,
				Services:    envSvcs,
				Tags:        map[string]string{"copilot-application": "testApp", "copilot-environment": "testEnv"},
			},
		},
		"success with resources": {
			shouldOutputResources: true,
			setupMocks: func(m envDescriberMocks) {
				gomock.InOrder(
					m.configStoreSvc.EXPECT().ListServices(testApp).Return([]*config.Service{
						testSvc1, testSvc2, testSvc3,
					}, nil),
					m.deployStoreSvc.EXPECT().ListDeployedServices(testApp, testEnv.Name).
						Return([]string{"testSvc1", "testSvc2"}, nil),
					m.stackDescriber.EXPECT().Stack("testApp-testEnv").Return(&cloudformation.Stack{
						Tags: stackTags,
					}, nil),
					m.stackDescriber.EXPECT().StackResources("testApp-testEnv").Return([]*cloudformation.StackResource{
						mockResource1,
						mockResource2,
					}, nil),
				)
			},
			wantedEnv: &EnvDescription{
				Environment: testEnv,
				Services:    envSvcs,
				Tags:        map[string]string{"copilot-application": "testApp", "copilot-environment": "testEnv"},
				Resources:   wantedResources,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockConfigStoreSvc := mocks.NewMockConfigStoreSvc(ctrl)
			mockDeployStoreSvc := mocks.NewMockDeployStoreSvc(ctrl)
			mockStackDescriber := mocks.NewMockstackAndResourcesDescriber(ctrl)
			mocks := envDescriberMocks{
				configStoreSvc: mockConfigStoreSvc,
				deployStoreSvc: mockDeployStoreSvc,
				stackDescriber: mockStackDescriber,
			}

			tc.setupMocks(mocks)

			d := &EnvDescriber{
				env:             testEnv,
				app:             testApp,
				enableResources: tc.shouldOutputResources,

				configStore:    mockConfigStoreSvc,
				deployStore:    mockDeployStoreSvc,
				stackDescriber: mockStackDescriber,
			}

			// WHEN
			actual, err := d.Describe()

			// THEN
			if tc.wantedError != nil {
				require.EqualError(t, err, tc.wantedError.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.wantedEnv, actual)
			}
		})
	}
}

func TestEnvDescription_JSONString(t *testing.T) {
	testApp := &config.Application{
		Name: "testApp",
		Tags: map[string]string{"key1": "value1", "key2": "value2"},
	}
	testEnv := &config.Environment{
		App:              "testApp",
		Name:             "testEnv",
		Region:           "us-west-2",
		AccountID:        "123456789012",
		Prod:             false,
		RegistryURL:      "",
		ExecutionRoleARN: "",
		ManagerRoleARN:   "",
	}
	testSvc1 := &config.Service{
		App:  "testApp",
		Name: "testSvc1",
		Type: "load-balanced",
	}
	testSvc2 := &config.Service{
		App:  "testApp",
		Name: "testSvc2",
		Type: "load-balanced",
	}
	testSvc3 := &config.Service{
		App:  "testApp",
		Name: "testSvc3",
		Type: "load-balanced",
	}
	allSvcs := []*config.Service{testSvc1, testSvc2, testSvc3}
	wantedContent := "{\"environment\":{\"app\":\"testApp\",\"name\":\"testEnv\",\"region\":\"us-west-2\",\"accountID\":\"123456789012\",\"prod\":false,\"registryURL\":\"\",\"executionRoleARN\":\"\",\"managerRoleARN\":\"\"},\"services\":[{\"app\":\"testApp\",\"name\":\"testSvc1\",\"type\":\"load-balanced\"},{\"app\":\"testApp\",\"name\":\"testSvc2\",\"type\":\"load-balanced\"},{\"app\":\"testApp\",\"name\":\"testSvc3\",\"type\":\"load-balanced\"}],\"tags\":{\"key1\":\"value1\",\"key2\":\"value2\"},\"resources\":[{\"type\":\"AWS::IAM::Role\",\"physicalID\":\"testApp-testEnv-CFNExecutionRole\"},{\"type\":\"testApp-testEnv-Cluster\",\"physicalID\":\"AWS::ECS::Cluster-jI63pYBWU6BZ\"}]}\n"

	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d := &EnvDescription{
		Environment: testEnv,
		Services:    allSvcs,
		Tags:        testApp.Tags,
		Resources:   wantedResources,
	}

	// WHEN
	actual, _ := d.JSONString()

	// THEN
	require.Equal(t, wantedContent, actual)
}

func TestEnvDescription_HumanString(t *testing.T) {
	testApp := &config.Application{
		Name: "testApp",
		Tags: map[string]string{"key1": "value1", "key2": "value2"},
	}
	testEnv := &config.Environment{
		App:              "testApp",
		Name:             "testEnv",
		Region:           "us-west-2",
		AccountID:        "123456789012",
		Prod:             false,
		RegistryURL:      "",
		ExecutionRoleARN: "",
		ManagerRoleARN:   "",
	}
	testSvc1 := &config.Service{
		App:  "testApp",
		Name: "testSvc1",
		Type: "load-balanced",
	}
	testSvc2 := &config.Service{
		App:  "testApp",
		Name: "testSvc2",
		Type: "load-balanced",
	}
	testSvc3 := &config.Service{
		App:  "testApp",
		Name: "testSvc3",
		Type: "load-balanced",
	}
	allSvcs := []*config.Service{testSvc1, testSvc2, testSvc3}

	wantedContent := `About

  Name              testEnv
  Production        false
  Region            us-west-2
  Account ID        123456789012

Services

  Name              Type
  ----              ----
  testSvc1          load-balanced
  testSvc2          load-balanced
  testSvc3          load-balanced

Tags

  Key               Value
  ---               -----
  key1              value1
  key2              value2

Resources

  AWS::IAM::Role           testApp-testEnv-CFNExecutionRole
  testApp-testEnv-Cluster  AWS::ECS::Cluster-jI63pYBWU6BZ
`
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d := &EnvDescription{
		Environment: testEnv,
		Services:    allSvcs,
		Tags:        testApp.Tags,
		Resources:   wantedResources,
	}

	// WHEN
	actual := d.HumanString()

	// THEN
	require.Equal(t, wantedContent, actual)
}
