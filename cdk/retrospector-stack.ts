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

interface CrawlerSettings {
  readonly enableURLHaus?: boolean;
  readonly enableOTX?: boolean;
  readonly secretsARN?: string;
};

interface RetrospectorProps extends cdk.StackProps{
  readonly lambdaRoleARN?: string;
  readonly entityObjectTopicARN?: string;
  readonly slackWebhookURL?: string;
  readonly crawler?: CrawlerSettings;

  readonly dynamoCapacity?: number;
  readonly entityLambdaConcurrency?: number;
  readonly iocLambdaConcurrency?: number;
  readonly sentryDSN?: string;
  readonly sentryEnv?: string;
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

  constructor(scope: cdk.Construct, id: string, retrospectorProps?: RetrospectorProps) {
    super(scope, id, retrospectorProps);
    const props = retrospectorProps || {};
    const crawlerSettings = props.crawler || {};

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
    const queues : {[key: string]: sqs.Queue} = {};
    ['iocRecord', 'iocDetect', 'entityRecord', 'entityDetect'].forEach(queueName => {
      const dlq = new sqs.Queue(this, queueName + 'DLQ');
      queues[queueName] = new sqs.Queue(this, queueName + 'Queue' ,{
        visibilityTimeout: cdk.Duration.seconds(300),
        deadLetterQueue: {
          maxReceiveCount: 3,
          queue: dlq,
        }
      });
    });
    this.iocRecordQueue = queues['iocRecord'];
    this.iocDetectQueue = queues['iocDetect'];
    this.entityRecordQueue = queues['entityRecord'];
    this.entityDetectQueue = queues['entityDetect'];

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

    const rootPath = path.resolve(__dirname, '..');
    const asset = lambda.Code.fromAsset(rootPath, {
      bundling: {
        image: lambda.Runtime.GO_1_X.bundlingDockerImage,
        user: 'root',
        command: ['make', 'asset'],
      },
    });

    const baseEnvVars = {
      IOC_TOPIC_ARN: this.iocTopic.topicArn,
      RECORD_TABLE_NAME: this.recordTable.tableName,
      SLACK_WEBHOOK_URL: props.slackWebhookURL || "",
      SECRETS_ARN: crawlerSettings.secretsARN || "",
      SENTRY_DSN: props.sentryDSN || "",
      SENTRY_ENVIRONMENT: props.sentryEnv || "",
    }

    // Setup crawlers
    const crawlers : Array<crawler> = [];
    if (crawlerSettings.enableURLHaus) {
      crawlers.push({
        funcName: 'crawlURLHaus',
        interval: cdk.Duration.hours(24),
      });
    }
    if (crawlerSettings.enableOTX) {
      crawlers.push({
        funcName: 'crawlOTX',
        interval: cdk.Duration.hours(1),
      });
    }

    crawlers.forEach(crawler => {
      const func = new lambda.Function(this, crawler.funcName, {
        runtime: lambda.Runtime.GO_1_X,
        handler: crawler.funcName,
        code: asset,
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
        concurrent: props.iocLambdaConcurrency || 1,
      },
      {
        funcName: 'iocDetect',
        source: this.iocDetectQueue,
        concurrent: props.iocLambdaConcurrency || 1,
      },
      {
        funcName: 'entityRecord',
        source: this.entityRecordQueue,
        concurrent: props.entityLambdaConcurrency || 10,
      },
      {
        funcName: 'entityDetect',
        source: this.entityDetectQueue,
        concurrent: props.entityLambdaConcurrency || 10,
      },
    ];

    this.handlers = {};

    handlers.forEach(handler => {
      const func = new lambda.Function(this, handler.funcName, {
        runtime: lambda.Runtime.GO_1_X,
        handler: handler.funcName,
        code: asset,
        role: lambdaRole,
        timeout: cdk.Duration.seconds(300),
        memorySize: 1024,
        environment: baseEnvVars,
        reservedConcurrentExecutions: handler.concurrent,
        events: [new SqsEventSource(handler.source, { batchSize: 10 })],
      });
      this.handlers[handler.funcName] = func;

      if (lambdaRole === undefined) {
        this.recordTable.grantReadWriteData(func);
      }
    })
  }
}
