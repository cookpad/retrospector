import * as cdk from '@aws-cdk/core';
import * as lambda from '@aws-cdk/aws-lambda';
import * as iam from '@aws-cdk/aws-iam';
import * as sns from '@aws-cdk/aws-sns';
import * as sqs from '@aws-cdk/aws-sqs';
import * as events from '@aws-cdk/aws-events';
import * as eventsTargets from '@aws-cdk/aws-events-targets';
import * as dynamodb from '@aws-cdk/aws-dynamodb';

import {
  SqsEventSource,
} from "@aws-cdk/aws-lambda-event-sources";


import * as path from 'path';
import { SqsSubscription } from "@aws-cdk/aws-sns-subscriptions";


interface RetrospectorProps extends cdk.StackProps{
  readonly lambdaRoleARN?: string;
  readonly entityObjectTopicARN?: string;
  readonly slackWebhookURL?: string;
  readonly dynamoCapacity?: number;
};

interface crawler {
  funcName: string;
  interval: cdk.Duration;
}

interface handler {
  funcName: string;
  source: sqs.Queue;
  concurrent: number,
}

export class RetrospectorStack extends cdk.Stack {
  // DynamoDB table
  recordTable: dynamodb.Table;

  // SQS queues
  iocRecordQueue: sqs.Queue;
  iocDetectQueue: sqs.Queue;
  entityRecordQueue: sqs.Queue;
  entityDetectQueue: sqs.Queue;

  // SNS topics
  iocTopic: sns.Topic;
  entityObjectTopic: sns.ITopic;

  // Lambda functions
  crawlers: Array<lambda.Function>;
  handlers: {[key: string]: lambda.Function};

  constructor(scope: cdk.Construct, id: string, props?: RetrospectorProps) {
    super(scope, id, props);

    props = props || {};

    // DynamoDB
    this.recordTable = new dynamodb.Table(this, "recordTable", {
      partitionKey: { name: "pk", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "sk", type: dynamodb.AttributeType.STRING },
      timeToLiveAttribute: "expires_at",
      billingMode: dynamodb.BillingMode.PROVISIONED,
      readCapacity: props.dynamoCapacity || 100,
      writeCapacity: props.dynamoCapacity || 100,
    });

    // SQS
    this.iocRecordQueue = new sqs.Queue(this, 'iocRecordQueue' ,{
      visibilityTimeout: cdk.Duration.seconds(120),
    });
    this.iocDetectQueue = new sqs.Queue(this, 'iocDetectQueue' ,{
      visibilityTimeout: cdk.Duration.seconds(120),
    });
    this.entityRecordQueue = new sqs.Queue(this, 'entityRecordQueue' ,{
      visibilityTimeout: cdk.Duration.seconds(120),
    });
    this.entityDetectQueue = new sqs.Queue(this, 'entityDetectQueue' ,{
      visibilityTimeout: cdk.Duration.seconds(120),
    });

    // SNS
    this.iocTopic = new sns.Topic(this, "iocTopic", {});
    this.iocTopic.addSubscription(new SqsSubscription(this.iocRecordQueue));
    this.iocTopic.addSubscription(new SqsSubscription(this.iocDetectQueue));

    if (props.entityObjectTopicARN !== undefined) {
      this.entityObjectTopic = sns.Topic.fromTopicArn(this, 'entityObjectTopic', props.entityObjectTopicARN);
    } else {
      this.entityObjectTopic = new sns.Topic(this, "entityObjectTopic", {});
    }
    this.entityObjectTopic.addSubscription(new SqsSubscription(this.entityRecordQueue));
    this.entityObjectTopic.addSubscription(new SqsSubscription(this.entityDetectQueue));

    // --------------------------------------
    // Lambda
    const lambdaRole = (props.lambdaRoleARN !== undefined) ? iam.Role.fromRoleArn(this, "LambdaRole", props.lambdaRoleARN, {
      mutable: false,
    }) : undefined;
    const buildPath = lambda.Code.fromAsset(path.join(__dirname, '../build'));

    const baseEnvVars = {
      IOC_TOPIC_ARN: this.iocTopic.topicArn,
      RECORD_TABLE_NAME: this.recordTable.tableName,
      SLACK_WEBHOOK_URL: props.slackWebhookURL || "",
    }

    // Setup crawlers
    const crawlers : Array<crawler> = [{
      funcName: 'crawlURLHouse',
      interval: cdk.Duration.hours(24),
    }]
    crawlers.forEach(crawler => {
      const func = new lambda.Function(this, crawler.funcName, {
        runtime: lambda.Runtime.GO_1_X,
        handler: crawler.funcName,
        code: buildPath,
        role: lambdaRole,
        timeout: cdk.Duration.seconds(300),
        memorySize: 1024,
        environment: baseEnvVars,
        reservedConcurrentExecutions: 1,
      });
      new events.Rule(this, "periodicInvoke" + crawler.funcName, {
        schedule: events.Schedule.rate(crawler.interval),
        targets: [new eventsTargets.LambdaFunction(func)],
      });
      if (lambdaRole == undefined) {
        this.iocTopic.grantPublish(func);
      }
    });

    // Setup recorders and detectors
    const handlers : Array<handler> = [
      {
        funcName: 'iocRecord',
        source: this.iocRecordQueue,
        concurrent: 1,
      },
      {
        funcName: 'iocDetect',
        source: this.iocDetectQueue,
        concurrent: 1,
      },
      {
        funcName: 'entityRecord',
        source: this.entityRecordQueue,
        concurrent: 10,
      },
      {
        funcName: 'entityDetect',
        source: this.entityDetectQueue,
        concurrent: 10,
      },
    ];
    handlers.forEach(handler => {
      const func = new lambda.Function(this, handler.funcName, {
        runtime: lambda.Runtime.GO_1_X,
        handler: handler.funcName,
        code: buildPath,
        role: lambdaRole,
        timeout: cdk.Duration.seconds(300),
        memorySize: 1024,
        environment: baseEnvVars,
        reservedConcurrentExecutions: handler.concurrent,
        events: [new SqsEventSource(handler.source, { batchSize: 10 })],
      });
      this.handlers[handler.funcName] = func;

      if (lambdaRole == undefined) {
        this.recordTable.grantReadWriteData(func);
      }
    })
  }
}
