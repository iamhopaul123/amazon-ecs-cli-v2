// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ecs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/copilot-cli/internal/pkg/aws/ecs/mocks"
	"github.com/aws/copilot-cli/internal/pkg/new-sdk-go/ecs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestECS_TaskDefinition(t *testing.T) {
	mockError := errors.New("error")

	testCases := map[string]struct {
		taskDefinitionName string
		mockECSClient      func(m *mocks.Mockapi)

		wantErr     error
		wantTaskDef *TaskDefinition
	}{
		"should return wrapped error given error": {
			taskDefinitionName: "task-def",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
					TaskDefinition: aws.String("task-def"),
				}).Return(nil, mockError)
			},
			wantErr: fmt.Errorf("describe task definition %s: %w", "task-def", mockError),
		},
		"returns task definition given a task definition name": {
			taskDefinitionName: "task-def",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
					TaskDefinition: aws.String("task-def"),
				}).Return(&ecs.DescribeTaskDefinitionOutput{
					TaskDefinition: &ecs.TaskDefinition{
						ContainerDefinitions: []*ecs.ContainerDefinition{
							{
								Environment: []*ecs.KeyValuePair{
									{
										Name:  aws.String("COPILOT_SERVICE_NAME"),
										Value: aws.String("my-app"),
									},
									{
										Name:  aws.String("COPILOT_ENVIRONMENT_NAME"),
										Value: aws.String("prod"),
									},
								},
							},
						},
					},
				}, nil)
			},
			wantTaskDef: &TaskDefinition{
				ContainerDefinitions: []*ecs.ContainerDefinition{
					{
						Environment: []*ecs.KeyValuePair{
							{
								Name:  aws.String("COPILOT_SERVICE_NAME"),
								Value: aws.String("my-app"),
							},
							{
								Name:  aws.String("COPILOT_ENVIRONMENT_NAME"),
								Value: aws.String("prod"),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			service := ECS{
				client: mockECSClient,
			}

			gotTaskDef, gotErr := service.TaskDefinition(tc.taskDefinitionName)

			if gotErr != nil {
				require.Equal(t, tc.wantErr, gotErr)
			} else {
				require.Equal(t, tc.wantTaskDef, gotTaskDef)
			}
		})

	}
}

func TestECS_Service(t *testing.T) {
	testCases := map[string]struct {
		clusterName   string
		serviceName   string
		mockECSClient func(m *mocks.Mockapi)

		wantErr error
		wantSvc *Service
	}{
		"success": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
					Cluster:  aws.String("mockCluster"),
					Services: aws.StringSlice([]string{"mockService"}),
				}).Return(&ecs.DescribeServicesOutput{
					Services: []*ecs.Service{
						{
							ServiceName: aws.String("mockService"),
						},
					},
				}, nil)
			},
			wantSvc: &Service{
				ServiceName: aws.String("mockService"),
			},
		},
		"errors if failed to describe service": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
					Cluster:  aws.String("mockCluster"),
					Services: aws.StringSlice([]string{"mockService"}),
				}).Return(nil, errors.New("some error"))
			},
			wantErr: fmt.Errorf("describe service mockService: some error"),
		},
		"errors if failed to find the service": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
					Cluster:  aws.String("mockCluster"),
					Services: aws.StringSlice([]string{"mockService"}),
				}).Return(&ecs.DescribeServicesOutput{
					Services: []*ecs.Service{
						{
							ServiceName: aws.String("badMockService"),
						},
					},
				}, nil)
			},
			wantErr: fmt.Errorf("cannot find service mockService"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			service := ECS{
				client: mockECSClient,
			}

			gotSvc, gotErr := service.Service(tc.clusterName, tc.serviceName)

			if gotErr != nil {
				require.EqualError(t, tc.wantErr, gotErr.Error())
			} else {
				require.Equal(t, tc.wantSvc, gotSvc)
			}
		})

	}
}

func TestECS_Tasks(t *testing.T) {
	testCases := map[string]struct {
		clusterName   string
		serviceName   string
		mockECSClient func(m *mocks.Mockapi)

		wantErr   error
		wantTasks []*Task
	}{
		"errors if failed to list running tasks": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
				}).Return(nil, errors.New("some error"))
			},
			wantErr: fmt.Errorf("list running tasks: some error"),
		},
		"errors if failed to describe running tasks": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
				}).Return(&ecs.ListTasksOutput{
					NextToken: nil,
					TaskArns:  aws.StringSlice([]string{"mockTaskArn"}),
				}, nil)
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String("mockCluster"),
					Tasks:   aws.StringSlice([]string{"mockTaskArn"}),
				}).Return(nil, errors.New("some error"))
			},
			wantErr: fmt.Errorf("describe running tasks in cluster mockCluster: some error"),
		},
		"success": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
				}).Return(&ecs.ListTasksOutput{
					NextToken: nil,
					TaskArns:  aws.StringSlice([]string{"mockTaskArn"}),
				}, nil)
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String("mockCluster"),
					Tasks:   aws.StringSlice([]string{"mockTaskArn"}),
				}).Return(&ecs.DescribeTasksOutput{
					Tasks: []*ecs.Task{
						{
							TaskArn: aws.String("mockTaskArn"),
						},
					},
				}, nil)
			},
			wantTasks: []*Task{
				{
					TaskArn: aws.String("mockTaskArn"),
				},
			},
		},
		"success with pagination": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
				}).Return(&ecs.ListTasksOutput{
					NextToken: aws.String("mockNextToken"),
					TaskArns:  aws.StringSlice([]string{"mockTaskArn1"}),
				}, nil)
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String("mockCluster"),
					Tasks:   aws.StringSlice([]string{"mockTaskArn1"}),
				}).Return(&ecs.DescribeTasksOutput{
					Tasks: []*ecs.Task{
						{
							TaskArn: aws.String("mockTaskArn1"),
						},
					},
				}, nil)
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
					NextToken:   aws.String("mockNextToken"),
				}).Return(&ecs.ListTasksOutput{
					NextToken: nil,
					TaskArns:  aws.StringSlice([]string{"mockTaskArn2"}),
				}, nil)
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String("mockCluster"),
					Tasks:   aws.StringSlice([]string{"mockTaskArn2"}),
				}).Return(&ecs.DescribeTasksOutput{
					Tasks: []*ecs.Task{
						{
							TaskArn: aws.String("mockTaskArn2"),
						},
					},
				}, nil)
			},
			wantTasks: []*Task{
				{
					TaskArn: aws.String("mockTaskArn1"),
				},
				{
					TaskArn: aws.String("mockTaskArn2"),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			service := ECS{
				client: mockECSClient,
			}

			gotTasks, gotErr := service.ServiceTasks(tc.clusterName, tc.serviceName)

			if gotErr != nil {
				require.EqualError(t, tc.wantErr, gotErr.Error())
			} else {
				require.Equal(t, tc.wantTasks, gotTasks)
			}
		})

	}
}

func TestECS_StopTasks(t *testing.T) {
	mockTasks := []string{"mockTask1", "mockTask2"}
	mockError := errors.New("some error")
	testCases := map[string]struct {
		cluster         string
		stopTasksReason string
		tasks           []string
		mockECSClient   func(m *mocks.Mockapi)

		wantErr error
	}{
		"errors if failed to stop tasks in default cluster": {
			tasks: mockTasks,
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().StopTask(&ecs.StopTaskInput{
					Task: aws.String("mockTask1"),
				}).Return(&ecs.StopTaskOutput{}, nil)
				m.EXPECT().StopTask(&ecs.StopTaskInput{
					Task: aws.String("mockTask2"),
				}).Return(&ecs.StopTaskOutput{}, mockError)
			},
			wantErr: fmt.Errorf("stop task mockTask2: some error"),
		},
		"success": {
			tasks:           mockTasks,
			cluster:         "mockCluster",
			stopTasksReason: "some reason",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().StopTask(&ecs.StopTaskInput{
					Cluster: aws.String("mockCluster"),
					Reason:  aws.String("some reason"),
					Task:    aws.String("mockTask1"),
				}).Return(&ecs.StopTaskOutput{}, nil)
				m.EXPECT().StopTask(&ecs.StopTaskInput{
					Cluster: aws.String("mockCluster"),
					Reason:  aws.String("some reason"),
					Task:    aws.String("mockTask2"),
				}).Return(&ecs.StopTaskOutput{}, nil)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			service := ECS{
				client: mockECSClient,
			}
			var opts []StopTasksOpts
			if tc.cluster != "" {
				opts = append(opts, WithStopTaskCluster(tc.cluster))
			}
			if tc.stopTasksReason != "" {
				opts = append(opts, WithStopTaskReason(tc.stopTasksReason))
			}
			gotErr := service.StopTasks(tc.tasks, opts...)

			if gotErr != nil {
				require.EqualError(t, tc.wantErr, gotErr.Error())
			} else {
				require.NoError(t, tc.wantErr)
			}
		})

	}
}

func TestECS_DefaultCluster(t *testing.T) {
	testCases := map[string]struct {
		mockECSClient func(m *mocks.Mockapi)

		wantedError    error
		wantedClusters string
	}{
		"get default clusters success": {
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().
					DescribeClusters(&ecs.DescribeClustersInput{}).
					Return(&ecs.DescribeClustersOutput{
						Clusters: []*ecs.Cluster{
							{
								ClusterArn:  aws.String("arn:aws:ecs:us-east-1:0123456:cluster/cluster1"),
								ClusterName: aws.String("cluster1"),
							},
							{
								ClusterArn:  aws.String("arn:aws:ecs:us-east-1:0123456:cluster/cluster2"),
								ClusterName: aws.String("cluster2"),
							},
						},
					}, nil)
			},

			wantedClusters: "arn:aws:ecs:us-east-1:0123456:cluster/cluster1",
		},
		"failed to get default clusters": {
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().
					DescribeClusters(&ecs.DescribeClustersInput{}).
					Return(nil, errors.New("error"))
			},
			wantedError: fmt.Errorf("get default cluster: %s", "error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			ecs := ECS{
				client: mockECSClient,
			}
			clusters, err := ecs.DefaultCluster()
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.Equal(t, tc.wantedClusters, clusters)
			}
		})
	}
}

func TestECS_HasDefaultCluster(t *testing.T) {
	testCases := map[string]struct {
		mockECSClient func(m *mocks.Mockapi)

		wantedHasDefaultCluster bool
		wantedErr               error
	}{
		"no default cluster": {
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeClusters(&ecs.DescribeClustersInput{}).
					Return(&ecs.DescribeClustersOutput{
						Clusters: []*ecs.Cluster{},
					}, nil)
			},
			wantedHasDefaultCluster: false,
		},
		"error getting default cluster": {
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeClusters(&ecs.DescribeClustersInput{}).
					Return(nil, errors.New("other error"))
			},
			wantedErr: fmt.Errorf("get default cluster: other error"),
		},
		"has default cluster": {
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeClusters(&ecs.DescribeClustersInput{}).
					Return(&ecs.DescribeClustersOutput{
						Clusters: []*ecs.Cluster{
							{ClusterArn: aws.String("cluster")},
						},
					}, nil)
			},
			wantedHasDefaultCluster: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			ecs := ECS{
				client: mockECSClient,
			}

			hasDefaultCluster, err := ecs.HasDefaultCluster()
			if tc.wantedErr != nil {
				require.EqualError(t, tc.wantedErr, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantedHasDefaultCluster, hasDefaultCluster)
		})
	}
}

func TestECS_RunTask(t *testing.T) {
	type input struct {
		cluster        string
		count          int
		subnets        []string
		securityGroups []string
		taskFamilyName string
		startedBy      string
	}

	runTaskInput := input{
		cluster:        "my-cluster",
		count:          3,
		subnets:        []string{"subnet-1", "subnet-2"},
		securityGroups: []string{"sg-1", "sg-2"},
		taskFamilyName: "my-task",
		startedBy:      "task",
	}

	testCases := map[string]struct {
		input

		mockECSClient func(m *mocks.Mockapi)

		wantedError error
		wantedTasks []*Task
	}{
		"run task success": {
			input: runTaskInput,
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().RunTask(&ecs.RunTaskInput{
					Cluster:        aws.String("my-cluster"),
					Count:          aws.Int64(3),
					LaunchType:     aws.String(ecs.LaunchTypeFargate),
					StartedBy:      aws.String("task"),
					TaskDefinition: aws.String("my-task"),
					NetworkConfiguration: &ecs.NetworkConfiguration{
						AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
							AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
							Subnets:        aws.StringSlice([]string{"subnet-1", "subnet-2"}),
							SecurityGroups: aws.StringSlice([]string{"sg-1", "sg-2"}),
						},
					},
				}).
					Return(&ecs.RunTaskOutput{
						Tasks: []*ecs.Task{
							{
								TaskArn: aws.String("task-1"),
							},
							{
								TaskArn: aws.String("task-2"),
							},
							{
								TaskArn: aws.String("task-3"),
							},
						},
					}, nil)
				m.EXPECT().WaitUntilTasksRunning(&ecs.DescribeTasksInput{
					Cluster: aws.String("my-cluster"),
					Tasks:   aws.StringSlice([]string{"task-1", "task-2", "task-3"}),
				}).Times(1)
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String("my-cluster"),
					Tasks:   aws.StringSlice([]string{"task-1", "task-2", "task-3"}),
				}).Return(&ecs.DescribeTasksOutput{
					Tasks: []*ecs.Task{
						{
							TaskArn: aws.String("task-1"),
						},
						{
							TaskArn: aws.String("task-2"),
						},
						{
							TaskArn: aws.String("task-3"),
						},
					},
				}, nil).Times(1)
			},

			wantedTasks: []*Task{
				{
					TaskArn: aws.String("task-1"),
				},
				{
					TaskArn: aws.String("task-2"),
				},
				{
					TaskArn: aws.String("task-3"),
				},
			},
		},
		"run task failed": {
			input: runTaskInput,

			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().RunTask(&ecs.RunTaskInput{
					Cluster:        aws.String("my-cluster"),
					Count:          aws.Int64(3),
					LaunchType:     aws.String(ecs.LaunchTypeFargate),
					StartedBy:      aws.String("task"),
					TaskDefinition: aws.String("my-task"),
					NetworkConfiguration: &ecs.NetworkConfiguration{
						AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
							AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
							Subnets:        aws.StringSlice([]string{"subnet-1", "subnet-2"}),
							SecurityGroups: aws.StringSlice([]string{"sg-1", "sg-2"}),
						},
					},
				}).
					Return(&ecs.RunTaskOutput{}, errors.New("error"))
				m.EXPECT().WaitUntilTasksRunning(gomock.Any()).Times(0)
			},
			wantedError: errors.New("run task(s) my-task: error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			ecs := ECS{
				client: mockECSClient,
			}

			tasks, err := ecs.RunTask(RunTaskInput{
				Count:          tc.count,
				Cluster:        tc.cluster,
				TaskFamilyName: tc.taskFamilyName,
				Subnets:        tc.subnets,
				SecurityGroups: tc.securityGroups,
				StartedBy:      tc.startedBy,
			})

			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.Equal(t, tc.wantedTasks, tasks)
			}
		})
	}
}

func TestECS_DescribeTasks(t *testing.T) {
	inCluster := "my-cluster"
	inTaskARNs := []string{"task-1", "task-2", "task-3"}
	testCases := map[string]struct {
		mockAPI     func(m *mocks.Mockapi)
		wantedError error
		wantedTasks []*Task
	}{
		"error describing tasks": {
			mockAPI: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String(inCluster),
					Tasks:   aws.StringSlice(inTaskARNs),
				}).Return(nil, errors.New("error describing tasks"))
			},
			wantedError: fmt.Errorf("describe tasks: %w", errors.New("error describing tasks")),
		},
		"successfully described tasks": {
			mockAPI: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String(inCluster),
					Tasks:   aws.StringSlice(inTaskARNs),
				}).Return(&ecs.DescribeTasksOutput{
					Tasks: []*ecs.Task{
						{
							TaskArn: aws.String("task-1"),
						},
						{
							TaskArn: aws.String("task-2"),
						},
						{
							TaskArn: aws.String("task-3"),
						},
					},
				}, nil)
			},
			wantedTasks: []*Task{
				{
					TaskArn: aws.String("task-1"),
				},
				{
					TaskArn: aws.String("task-2"),
				},
				{
					TaskArn: aws.String("task-3"),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := mocks.NewMockapi(ctrl)
			tc.mockAPI(mockAPI)

			ecs := ECS{
				client: mockAPI,
			}

			tasks, err := ecs.DescribeTasks(inCluster, inTaskARNs)
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedTasks, tasks)
			}
		})
	}
}

func TestECS_ExecuteCommand(t *testing.T) {
	mockExecCmdIn := &ecs.ExecuteCommandInput{
		Cluster:     aws.String("mockCluster"),
		Command:     aws.String("mockCommand"),
		Interactive: aws.Bool(true),
		Container:   aws.String("mockContainer"),
		Task:        aws.String("mockTask"),
	}
	mockSess := &ecs.Session{
		SessionId: aws.String("mockSessID"),
	}
	mockErr := errors.New("some error")
	testCases := map[string]struct {
		mockAPI         func(m *mocks.Mockapi)
		mockSessStarter func(m *mocks.MockssmSessionStarter)
		wantedError     error
	}{
		"return error if fail to call ExecuteCommand": {
			mockAPI: func(m *mocks.Mockapi) {
				m.EXPECT().ExecuteCommand(mockExecCmdIn).Return(nil, mockErr)
			},
			mockSessStarter: func(m *mocks.MockssmSessionStarter) {},
			wantedError:     fmt.Errorf("execute command: some error"),
		},
		"return error if fail to start the session": {
			mockAPI: func(m *mocks.Mockapi) {
				m.EXPECT().ExecuteCommand(&ecs.ExecuteCommandInput{
					Cluster:     aws.String("mockCluster"),
					Command:     aws.String("mockCommand"),
					Interactive: aws.Bool(true),
					Container:   aws.String("mockContainer"),
					Task:        aws.String("mockTask"),
				}).Return(&ecs.ExecuteCommandOutput{
					Session: mockSess,
				}, nil)
			},
			mockSessStarter: func(m *mocks.MockssmSessionStarter) {
				m.EXPECT().StartSession(mockSess).Return(mockErr)
			},
			wantedError: fmt.Errorf("start session mockSessID using ssm plugin: some error"),
		},
		"success": {
			mockAPI: func(m *mocks.Mockapi) {
				m.EXPECT().ExecuteCommand(mockExecCmdIn).Return(&ecs.ExecuteCommandOutput{
					Session: mockSess,
				}, nil)
			},
			mockSessStarter: func(m *mocks.MockssmSessionStarter) {
				m.EXPECT().StartSession(mockSess).Return(nil)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := mocks.NewMockapi(ctrl)
			mockSessStarter := mocks.NewMockssmSessionStarter(ctrl)
			tc.mockAPI(mockAPI)
			tc.mockSessStarter(mockSessStarter)

			ecs := ECS{
				client: mockAPI,
				newSessStarter: func() ssmSessionStarter {
					return mockSessStarter
				},
			}

			err := ecs.ExecuteCommand(ExecuteCommandInput{
				Cluster:   "mockCluster",
				Command:   "mockCommand",
				Container: "mockContainer",
				Task:      "mockTask",
			})
			if tc.wantedError != nil {
				require.EqualError(t, err, tc.wantedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
