# mstdnlambda: An AWS Mastodon Push Notification Gateway

## Purpose
This AWS Lambda function will act as a gateway for processing Mastodon push
notifications. The lambda will process, validate, decrypt and verify all
incoming messages. Those that are valid are then forwarded onto one or more
SNS topics where you can process the messages as desired.

![](https://gitlab.com/ddb_db/mstdnlambda/-/wikis/uploads/91db85e4c03da8f51f8cc6f477eedd23/mstdnlambda-arch.png)

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

## Setup
The setup as shown in the diagram is rather straightfoward and for low volumes
of Mastodon push notifications, can be run for free* on AWS.

A detailed guide on how to setup this gateway lambda with a second lambda as a
topic subscriber is coming soon. How soon? If you're interested in details on
how to set this up then [let me know](https://mstdn.ca/@ddb_db) as it might
just nudge me to finish the documentation a little quicker. ;)

`*` Though you can run low volumes of notifications easily within the AWS free
tier, you still need to sign up for an AWS account and provide a credit card
number.