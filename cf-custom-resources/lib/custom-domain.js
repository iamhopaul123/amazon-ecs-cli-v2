// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
"use strict";

const aws = require("aws-sdk");

// These are used for test purposes only
let defaultResponseURL;

/**
 * Upload a CloudFormation response object to S3.
 *
 * @param {object} event the Lambda event payload received by the handler function
 * @param {object} context the Lambda context received by the handler function
 * @param {string} responseStatus the response status, either 'SUCCESS' or 'FAILED'
 * @param {string} physicalResourceId CloudFormation physical resource ID
 * @param {object} [responseData] arbitrary response data object
 * @param {string} [reason] reason for failure, if any, to convey to the user
 * @returns {Promise} Promise that is resolved on success, or rejected on connection error or HTTP error response
 */
let report = function (
  event,
  context,
  responseStatus,
  physicalResourceId,
  responseData,
  reason
) {
  return new Promise((resolve, reject) => {
    const https = require("https");
    const { URL } = require("url");

    var responseBody = JSON.stringify({
      Status: responseStatus,
      Reason: reason,
      PhysicalResourceId: physicalResourceId || context.logStreamName,
      StackId: event.StackId,
      RequestId: event.RequestId,
      LogicalResourceId: event.LogicalResourceId,
      Data: responseData,
    });

    const parsedUrl = new URL(event.ResponseURL || defaultResponseURL);
    const options = {
      hostname: parsedUrl.hostname,
      port: 443,
      path: parsedUrl.pathname + parsedUrl.search,
      method: "PUT",
      headers: {
        "Content-Type": "",
        "Content-Length": responseBody.length,
      },
    };

    https
      .request(options)
      .on("error", reject)
      .on("response", (res) => {
        res.resume();
        if (res.statusCode >= 400) {
          reject(new Error(`Error ${res.statusCode}: ${res.statusMessage}`));
        } else {
          resolve();
        }
      })
      .end(responseBody, "utf8");
  });
};

/**
 * @param {string} aliases the custom domain aliases
 * @param {string} lbDNS DNS of the load balancer
 * @param {string} lbHostedZone Hosted Zone of the load balancer
 * @param {string} rootDnsRole the IAM role ARN that can manage domainName
 */
const writeCustomDomainRecord = async function (
  aliases,
  lbDNS,
  lbHostedZone,
  rootDnsRole,
  aliasTypes,
  action
) {
  const envRoute53 = new aws.Route53();
  const appRoute53 = new aws.Route53({
    credentials: new aws.ChainableTemporaryCredentials({
      params: { RoleArn: rootDnsRole },
      masterCredentials: new aws.EnvironmentCredentials("AWS"),
    }),
  });
  const aliasList = getAllAliases(aliases);
  for (const alias of aliasList) {
    const aliasType = getAliasType(aliasTypes, alias);
    switch (aliasType) {
      case aliasTypes.EnvDomainZone:
        await writeARecord(envRoute53, alias, lbDNS, lbHostedZone, aliasType.domain, action);
        break;
      case aliasTypes.AppDomainZone:
        await writeARecord(appRoute53, alias, lbDNS, lbHostedZone, aliasType.domain, action);
        break;
      case aliasTypes.RootDomainZone:
        await writeARecord(appRoute53, alias, lbDNS, lbHostedZone, aliasType.domain, action);
        break;
      default:
    }
  }
};

const writeARecord = async function (
  route53,
  alias,
  lbDNS,
  lbHostedZone,
  domain,
  action
){
  const hostedZones = await route53
    .listHostedZonesByName({
      DNSName: domain,
      MaxItems: "1",
    })
    .promise();

  if (!hostedZones.HostedZones || hostedZones.HostedZones.length == 0) {
    throw new Error(
      `Couldn't find any Hosted Zone with DNS name ${domain}.`
    );
  }
  const hostedZoneId = hostedZones.HostedZones[0].Id.split("/").pop();
  console.log(
    `${action} A record into Hosted Zone ${hostedZoneId}`
  );
  const changeBatch = await updateRecords(
    route53,
    hostedZoneId,
    action,
    alias,
    lbDNS,
    lbHostedZone
  );
  await waitForRecordChange(route53, changeBatch.ChangeInfo.Id);
}

/**
 * Custom domain handler, invoked by Lambda.
 */
exports.handler = async function (event, context) {
  var responseData = {};
  var physicalResourceId;
  const props = event.ResourceProperties;
  const [app, env, domain] = [props.AppName, props.EnvName, props.DomainName]
  var aliasTypes = {
    EnvDomainZone: { regex: `.*${env}.${app}.${domain}`, domain: `${env}.${app}.${domain}` },
    AppDomainZone: { regex: `.*${app}.${domain}`, domain: `${app}.${domain}` },
    RootDomainZone: { regex: `.*${domain}`, domain: `${domain}` },
    OtherDomainZone: { regex: `.*` },
  };

  try {
    switch (event.RequestType) {
      case "Create":
      case "Update":
        await writeCustomDomainRecord(
          props.Aliases,
          props.LoadBalancerDNS,
          props.LoadBalancerHostedZone,
          props.AppDNSRole,
          aliasTypes,
          "UPSERT"
        );
        physicalResourceId = `custom-domain-${event.LogicalResourceId}`;
        break;
      case "Delete":
        await writeCustomDomainRecord(
          props.Aliases,
          props.LoadBalancerDNS,
          props.LoadBalancerHostedZone,
          props.AppDNSRole,
          aliasTypes,
          "DELETE"
        );
        physicalResourceId = event.PhysicalResourceId;
        break;
      default:
        throw new Error(`Unsupported request type ${event.RequestType}`);
    }
    await report(event, context, "SUCCESS", physicalResourceId, responseData);
  } catch (err) {
    console.log(`Caught error ${err}.`);
    await report(
      event,
      context,
      "FAILED",
      physicalResourceId,
      null,
      err.message
    );
  }
};

const getAllAliases = function (aliases) {
  var obj = JSON.parse(aliases || '{}');
  var aliasList = [];
  for (var m in obj) {
    aliasList.push(...obj[m].split(','));
  }
  return [...new Set(aliasList)];
};

const getAliasType = function (
  aliasTypes,
  alias
) {
  switch (true) {
    case new RegExp(aliasTypes.EnvDomainZone.regex).test(alias):
      return aliasTypes.EnvDomainZone;
    case new RegExp(aliasTypes.AppDomainZone.regex).test(alias):
      return aliasTypes.AppDomainZone;
    case new RegExp(aliasTypes.RootDomainZone.regex).test(alias):
      return aliasTypes.RootDomainZone;
    default:
      return aliasTypes.OtherDomainZone;
  }
};

const waitForRecordChange = function (route53, changeId) {
  return route53
    .waitFor("resourceRecordSetsChanged", {
      // Wait up to 5 minutes
      $waiter: {
        delay: 30,
        maxAttempts: 10,
      },
      Id: changeId,
    })
    .promise();
};

const updateRecords = function (
  route53,
  hostedZone,
  action,
  alias,
  lbDNS,
  lbHostedZone
) {
  return route53
    .changeResourceRecordSets({
      ChangeBatch: {
        Changes: [
          {
            Action: action,
            ResourceRecordSet: {
              Name: alias,
              Type: 'A',
              AliasTarget: {
                HostedZoneId: lbHostedZone,
                DNSName: lbDNS,
                EvaluateTargetHealth: true
              },
            },
          },
        ],
      },
      HostedZoneId: hostedZone,
    })
    .promise();
};

/**
 * @private
 */
exports.withDefaultResponseURL = function (url) {
  defaultResponseURL = url;
};
