# mstdnlambda: An AWS Mastodon Push Notification Gateway

## Purpose
This AWS Lambda function will act as a gateway for processing Mastodon push
notifications. The lambda will process, validate, decrypt and verify all
incoming messages. Those that are valid are then forwarded onto one or more
SNS topics where you can process the messages as desired.

![](https://gitlab.com/ddb_db/mstdnlambda/-/wikis/uploads/eb4fbf48f56d831a3ffa8e19eb3a6437/mstdnlambda-arch__1_.png)

As shown in the diagram, this lambda's job is to deal with the network
communication of receiving the message from Mastodon and dealing with the 
decryption and verification of each message received. When a valid message is
received, the plain JSON structure alone is then posted to an SNS topic, which
then delivers the JSON message to all subscribers of the topic. Any valid SNS
topic subscriber can receive the push notification. What the subcriber does
with the message is up to you. My motivation is to explore interactive
Mastodon bots. So for me, I have another lambda function subscribing to the SNS
topic. The advantage here is that the SNS subscribers don't need to deal with
any of the complexities of processing the web request, decrypting the message
and verifying the JWT token. All of this is handled by this lambda and only
when all of the decryption and verification succeeds does the message get
sent to the SNS topic. The subscriber simply receives the decrypted, already verified JSON object sent by Mastodon and can process it as needed, not
having to deal with nor worry about how to decrypt the incoming message
from Mastodon, if the message is valid, etc.

## Downloads
Downloads are available
[here](https://gitlab.com/ddb_db/mstdnlambda/-/packages/).  The zip files
for `linux/amd64` and `linux/arm64` are purpose built for Lambda and are ready
to go as the code source for your function.

### webpushkeys
The `webpushkeys` utility can be used to generate a suitable set of keys and a
shared secret compatiable with the requirements of the Mastodon push
notification API call. You would use the public key as the argument to the 
Mastodon push API call when registering your push notification. The private key
would be an environment variable configured in the lambda function. The shared
secret is sent to Mastodon as part of the API call and also configured as
another environment variable of the lambda function.

## Setup
The setup as shown in the diagram is rather straightfoward and for low volumes
of Mastodon push notifications, can be run for free* on AWS.

A detailed guide on how to setup this gateway lambda with a second lambda as a
topic subscriber is coming soon. How soon? If you're interested in details on
how to set this up then [let me know](https://mstdn.ca/@ddb_db) as it might
just nudge me to finish the documentation a little quicker. ;)

A very brief overview for anyone just dying to set this up for themselves
before I have time to write a more detailed tutorial:

* [Download](https://gitlab.com/ddb_db/mstdnlambda/-/packages/) the
  `webpushkeys` utility for your local platform, you'll need this
  to generate a set of encryption keys and a shared secret. These are needed
  when registering for push notifications on your Mastodon instance.
* Create an SNS topic on AWS
  * Take note of the topic ARN, you'll need this later.
* Create a lambda function on AWS.
  * Configure the function URL; this is the URL Mastodon will send your push
    notifications to.
  * Configure the function runtime as `Go 1.x`.
  * Download the lastest zip file package, use it as the code source for
    the function.
  * Two environment variables need to be configured for the function:
    * `MSTDN_PRIVATE_KEY`: The private key you generated with `webpushkeys`
    * `MSTDN_SHARED_SECRET`: The shared secret you generated with `webpushkeys`
    * **Please read the security notice below regarding these environment
      variables.**
  * Configure the function's execution role such that it has permissions to
    publish messages to the SNS topic you created above.
  * Register for push notifications from your Mastodon instance.
    [These API docs](https://docs.joinmastodon.org/methods/push/#create)
    should help you.
    * The `p256dh` key parameter is the public key you generated with 
      `webpushkeys`
    * The `auth` key parameter is the shared secret you generated with
      `webpushkeys`
    * The `endpoint` parameter is the lambda function URL **with your topic ARN
      encoded in the path.** See below for more details. But you **MUST** encode the topic ARN in the endpoint URL and use that URL for your
      subscription.
    * Be sure to subscribe to at least one of the notification types as part of
      the API call.
  * Setup one or more subscriptions to the topic. For quick testing, you may
    want to setup an email subscriber just to make sure messages are flowing
    completely through the entire chain. Once you get an email subscriber
    working then you can move onto more interesting ones like lambdas, SQS,
    webhooks, etc.
  * Generate an event on your Mastodon account that you registered for (status
    notification, follow notification, etc.)
  * If all went well, your topic subscriber(s) should have received the 
    message.
    * If not, head over to CloudWatch and start looking at the logs of this
      lambda to ensure that it delivered the message to SNS. If it did then
      debug the SNS fanout, etc.

`*` Though you can run low volumes of notifications easily within the AWS free
tier, you still need to sign up for an AWS account and provide a credit card
number.

## Environment Variable Security
As decribed above, the private key and shared secret are configured as plain
environment variables in your lambda's configuration. **This is definitely not
the most secure approach.** If there are concerns about others with access to
your AWS account seeing these values then you **should NOT deploy** this lambda
to your AWS environment.  These values should not be shared and if compromised
they can be used to impersonate you and possibly intercept push notifications
from Mastodon. I'm assuming and coming from a single user AWS account where
there is only one person with access to the AWS console. In such a setup, there
are no concerns about other untrusted AWS admin users seeing these secrets.

The secure approach here would be to configure these values as secrets in the
AWS Secrets Manager and access them via that secrets vault at runtime using
appropriate IAM access. This implementation is not on my radar but if it's a
concern, open a ticket and it can be explored as an alternate to plain env
variables.

## Endpoint URL
The endpoint URL you pass to Mastodon when registering for push notifications
**MUST** include the URL safe base64 encoded topic ARN(s) you wish to deliver 
your messages to.  If you do not include the encoded topic arns, the messages
will be rejected by the lambda.

The endpoint URL you need to pass to the Mastodon API call will therefore look
like this:

`https://abcdefg123.lambda-url.us-east-1.on.aws/ff012345/`

Everything up to the first slash is just your assigned lambda function URL. The
`ff012345` is the URL safe base64 encoding of your SNS topic ARN. Visit
https://base64encode.org/ and paste in your topic ARN, **check the 'Perform
URL-safe encoding' box then generate the encoded value**.  This is the value
you need to append to the URL. If you wanted to deliver the message to multiple
SNS topics, you can append multiple encoded topic ARNs, just separate them by
slashes on the end of the URL. When you specify multiple topics, all must
publish successfully otherwise an error is returned to Mastodon and it will
retry **all** of them again.

This raises a good point: your subscribers must be able to receive a 
notification more than once. Your handlers must be prepared for this and must
be idempotent.