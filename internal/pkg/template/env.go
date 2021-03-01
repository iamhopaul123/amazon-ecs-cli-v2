// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"bytes"
	"fmt"

	"github.com/aws/copilot-cli/internal/pkg/config"
)

const (
	fmtEnvCFTemplatePath    = "environment/versions/cf-%s.yml"
	fmtEnvCFSubTemplatePath = "environment/partials/%s.yml"
)

var (
	// Template names under "environment/partials/".
	envCFSubTemplateNames = []string{
		"cfn-execution-role",
		"custom-resources",
		"custom-resources-role",
		"environment-manager-role",
		"lambdas",
		"vpc-resources",
	}
)

// EnvOpts holds data that can be provided to enable features in an environment stack template.
type EnvOpts struct {
	Version string // The template version to use for the environment. If empty uses the "legacy" template.

	DNSDelegationLambda       string
	ACMValidationLambda       string
	EnableLongARNFormatLambda string
	CustomDomainLambda        string

	ImportVPC *config.ImportVPC
	VPCConfig *config.AdjustVPC
}

// ParseEnv parses an environment's CloudFormation template with the specified data object and returns its content.
func (t *Template) ParseEnv(data *EnvOpts, options ...ParseOption) (*Content, error) {
	tpl, err := t.parse("base", envTemplatePath(data.Version), options...)
	if err != nil {
		return nil, err
	}
	for _, templateName := range envCFSubTemplateNames {
		nestedTpl, err := t.parse(templateName, fmt.Sprintf(fmtEnvCFSubTemplatePath, templateName), options...)
		if err != nil {
			return nil, err
		}
		_, err = tpl.AddParseTree(templateName, nestedTpl.Tree)
		if err != nil {
			return nil, fmt.Errorf("add parse tree of %s to base template: %w", templateName, err)
		}
	}
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("execute environment template with data %v: %w", data, err)
	}
	return &Content{buf}, nil
}

func envTemplatePath(version string) string {
	if version == "" {
		return fmt.Sprintf(fmtEnvCFTemplatePath, "v0.0.0")
	}
	return fmt.Sprintf(fmtEnvCFTemplatePath, version)
}
