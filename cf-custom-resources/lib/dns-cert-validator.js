// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
"use strict";

const aws = require("aws-sdk");

const defaultSleep = function (ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
};

// These are used for test purposes only
let defaultResponseURL;
let waiter;
let sleep = defaultSleep;
let random = Math.random;
let maxAttempts = 10;

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
 * Requests a public certificate from AWS Certificate Manager, using DNS validation.
 * The hosted zone ID must refer to a **public** Route53-managed DNS zone that is authoritative
 * for the suffix of the certificate's Common Name (CN).  For example, if the CN is
 * `*.example.com`, the hosted zone ID must point to a Route 53 zone authoritative
 * for `example.com`.
 *
 * @param {string} requestId the CloudFormation request ID
 * @param {string} appName the name of the application
 * @param {string} envName the name of the environment
 * @param {string} domainName the Common Name (CN) field for the requested certificate
 * @param {string} subjectAlternativeNames additional FQDNs to be included in the
 * Subject Alternative Name extension of the requested certificate
 * @param {string} envHostedZoneId the environment Route53 Hosted Zone ID
 * @param {string} rootDnsRole the IAM role ARN that can manage domainName
 * @returns {string} Validated certificate ARN
 */
const requestCertificate = async function (
  requestId,
  appName,
  envName,
  domainName,
  subjectAlternativeNames,
  envHostedZoneId,
  rootDnsRole,
  region
) {
  const crypto = require("crypto");
  const [acm, envRoute53] = clients(region);
  const reqCertResponse = await acm
    .requestCertificate({
      DomainName: `${envName}.${appName}.${domainName}`,
      SubjectAlternativeNames: subjectAlternativeNames,
      IdempotencyToken: crypto
        .createHash("sha256")
        .update(requestId)
        .digest("hex")
        .substr(0, 32),
      ValidationMethod: "DNS",
    })
    .promise();

  let options;
  let attempt;
  for (attempt = 0; attempt < maxAttempts; attempt++) {
    const { Certificate } = await acm
      .describeCertificate({
        CertificateArn: reqCertResponse.CertificateArn,
      })
      .promise();
    options = Certificate.DomainValidationOptions || [];
    var ok = false
    for (const option of options) {
      if (!option.ResourceRecord) {
        ok = false
        break
      }
      ok = true
    }
    if (ok) {
      break;
    }
    // Exponential backoff with jitter based on 200ms base
    // component of backoff fixed to ensure minimum total wait time on
    // slow targets.
    const base = Math.pow(2, attempt);
    await sleep(random() * base * 50 + base * 150);
  }
  if (attempt === maxAttempts) {
    throw new Error(
      `DescribeCertificate did not contain DomainValidationOptions after ${maxAttempts} tries.`
    );
  }

  const appRoute53 = new aws.Route53({
    credentials: new aws.ChainableTemporaryCredentials({
      params: { RoleArn: rootDnsRole },
      masterCredentials: new aws.EnvironmentCredentials("AWS"),
    }),
  });

  for (const option of options) {
    switch (option.DomainName) {
      case `${envName}.${appName}.${domainName}`:
        await validateDomain(
          envRoute53,
          option.ResourceRecord,
          "UPSERT",
          "",
          envHostedZoneId
        );
        break;
      case `${appName}.${domainName}`:
        await validateDomain(
          appRoute53,
          option.ResourceRecord,
          "UPSERT",
          `${appName}.${domainName}`
        );
        break;
      case domainName:
        await validateDomain(
          appRoute53,
          option.ResourceRecord,
          "UPSERT",
          domainName
        );
        break;
    }
  }

  await acm
    .waitFor("certificateValidated", {
      // Wait up to 9 minutes and 30 seconds
      $waiter: {
        delay: 30,
        maxAttempts: 19,
      },
      CertificateArn: reqCertResponse.CertificateArn,
    })
    .promise();

  return reqCertResponse.CertificateArn;
};

const validateDomain = async function (
  route53,
  record,
  action,
  domainName,
  hostedZoneId
) {
  if (!hostedZoneId) {
    const hostedZones = await route53
      .listHostedZonesByName({
        DNSName: domainName,
        MaxItems: "1",
      })
      .promise();

    if (!hostedZones.HostedZones || hostedZones.HostedZones.length == 0) {
      throw new Error(
        `Couldn't find any Hosted Zone with DNS name ${domainName}.`
      );
    }
    hostedZoneId = hostedZones.HostedZones[0].Id.split("/").pop();
  }
  console.log(
    `${action} DNS record into Hosted Zone ${hostedZoneId}: ${record.Name} ${record.Type} ${record.Value}`
  );
  const changeBatch = await updateRecords(
    route53,
    hostedZoneId,
    action,
    record.Name,
    record.Type,
    record.Value
  );
  await waitForRecordChange(route53, changeBatch.ChangeInfo.Id);
};

/**
 * Deletes a certificate from AWS Certificate Manager (ACM) by its ARN.
 * If the certificate does not exist, the function will return normally.
 *
 * @param {string} arn The certificate ARN
 */
const deleteCertificate = async function (
  arn,
  appName,
  envName,
  domainName,
  region,
  envHostedZoneId,
  rootDnsRole
) {
  const [acm, envRoute53] = clients(region);
  try {
    console.log(`Waiting for certificate ${arn} to become unused`);

    let inUseByResources;
    let options;

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      const { Certificate } = await acm
        .describeCertificate({
          CertificateArn: arn,
        })
        .promise();

      inUseByResources = Certificate.InUseBy || [];
      options = Certificate.DomainValidationOptions || [];
      var ok = false
      for (const option of options) {
        if (!option.ResourceRecord) {
          ok = false
          break
        }
        ok = true
      }
      if (!ok || inUseByResources.length) {
        // Deleting resources can be quite slow - so just sleep 30 seconds between checks.
        await sleep(30000);
      } else {
        break;
      }
    }

    if (inUseByResources.length) {
      throw new Error(
        `Certificate still in use after checking for ${maxAttempts} attempts.`
      );
    }

    // Fetch the DNS Validation Record and delete it
    const appRoute53 = new aws.Route53({
      credentials: new aws.ChainableTemporaryCredentials({
        params: { RoleArn: rootDnsRole },
        masterCredentials: new aws.EnvironmentCredentials("AWS"),
      }),
    });

    for (const option of options) {
      switch (option.DomainName) {
        case `${envName}.${appName}.${domainName}`:
          await validateDomain(
            envRoute53,
            option.ResourceRecord,
            "DELETE",
            "",
            envHostedZoneId
          );
          break;
        case `${appName}.${domainName}`:
          await validateDomain(
            appRoute53,
            option.ResourceRecord,
            "DELETE",
            `${appName}.${domainName}`
          );
          break;
        case domainName:
          await validateDomain(
            appRoute53,
            option.ResourceRecord,
            "DELETE",
            domainName
          );
          break;
      }
    }

    await acm
      .deleteCertificate({
        CertificateArn: arn,
      })
      .promise();
  } catch (err) {
    if (err.name !== "ResourceNotFoundException") {
      throw err;
    }
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
  recordName,
  recordType,
  recordValue
) {
  return route53
    .changeResourceRecordSets({
      ChangeBatch: {
        Changes: [
          {
            Action: action,
            ResourceRecordSet: {
              Name: recordName,
              Type: recordType,
              TTL: 60,
              ResourceRecords: [
                {
                  Value: recordValue,
                },
              ],
            },
          },
        ],
      },
      HostedZoneId: hostedZone,
    })
    .promise();
};

const clients = function (region) {
  const acm = new aws.ACM({
    region,
  });
  const route53 = new aws.Route53();
  if (waiter) {
    // Used by the test suite, since waiters aren't mockable yet
    route53.waitFor = acm.waitFor = waiter;
  }
  return [acm, route53];
};

/**
 * Main certificate manager handler, invoked by Lambda
 */
exports.certificateRequestHandler = async function (event, context) {
  var responseData = {};
  var physicalResourceId;
  var certificateArn;
  const props = event.ResourceProperties;

  try {
    switch (event.RequestType) {
      case "Create":
      case "Update":
        certificateArn = await requestCertificate(
          event.RequestId,
          props.AppName,
          props.EnvName,
          props.DomainName,
          props.SubjectAlternativeNames,
          props.EnvHostedZoneId,
          props.RootDNSRole,
          props.Region
        );
        responseData.Arn = physicalResourceId = certificateArn;
        break;
      case "Delete":
        physicalResourceId = event.PhysicalResourceId;
        // If the resource didn't create correctly, the physical resource ID won't be the
        // certificate ARN, so don't try to delete it in that case.
        if (physicalResourceId.startsWith("arn:")) {
          await deleteCertificate(
            physicalResourceId,
            props.AppName,
            props.EnvName,
            props.DomainName,
            props.Region,
            props.EnvHostedZoneId,
            props.RootDNSRole
          );
        }
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

/**
 * @private
 */
exports.withDefaultResponseURL = function (url) {
  defaultResponseURL = url;
};

/**
 * @private
 */
exports.withWaiter = function (w) {
  waiter = w;
};

/**
 * @private
 */
exports.withSleep = function (s) {
  sleep = s;
};

/**
 * @private
 */
exports.reset = function () {
  sleep = defaultSleep;
  random = Math.random;
  waiter = undefined;
  maxAttempts = 10;
};

/**
 * @private
 */
exports.withRandom = function (r) {
  random = r;
};

/**
 * @private
 */
exports.withMaxAttempts = function (ma) {
  maxAttempts = ma;
};
