package sqssendmessage

import (
    "github.com/TIBCOSoftware/flogo-lib/core/activity"
    "github.com/TIBCOSoftware/flogo-lib/logger"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
    "github.com/TIBCOSoftware/flogo-lib/core/data"
)

const (
    ivConnection         = "sqsConnection"
    ivQueueUrl           = "queueUrl"
    ivMessageAttributes  = "MessageAttributes"
    ivDelaySeconds       = "DelaySeconds"
    ivMessageBody        = "MessageBody"
    ovMessageId          = "MessageId"
)

var activityLog = logger.GetLogger("aws-activity-sqssendmessage")

type SQSSendMessageActivity struct {
	metadata *activity.Metadata
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &SQSSendMessageActivity{metadata: metadata}
}

func (a *SQSSendMessageActivity) Metadata() *activity.Metadata {
	return a.metadata
}
func (a *SQSSendMessageActivity) Eval(context activity.Context) (done bool, err error) {
    activityLog.Info("Executing SQS Send Message activity")
    //Read Inputs
    if context.GetInput(ivConnection) == nil {
      return false, activity.NewError("SQS connection is not configured", "SQS-SENDMESSAGE-4001", nil)
    }
    
    if context.GetInput(ivQueueUrl) == nil {
      return false, activity.NewError("SQS Queue URL is not configured", "SQS-SENDMESSAGE-4002", nil)
    }
    
    if context.GetInput(ivMessageBody) == nil {
      return false, activity.NewError("Message body is not configured", "SQS-SENDMESSAGE-4003", nil)
    }


    //Read connection details
    connectionInfo := context.GetInput(ivConnection).(map[string]interface{})
    session, err := session.NewSession(aws.NewConfig().WithRegion(connectionInfo["region"].(string)).WithCredentials(credentials.NewStaticCredentials(connectionInfo["accessKeyId"].(string), connectionInfo["secreteAccessKey"].(string), "")))
    if err != nil {
      return false, activity.NewError(fmt.Sprintf("Failed to connect to AWS due to error:%s. Check credentials configured in the connection:%s.",err.Error(),connectionInfo["name"].(string)), "SQS-SENDMESSAGE-4004", nil)
    }
    //Create SQS service instance
    sqsSvc := sqs.New(session)
    sendMessageInput := &sqs.SendMessageInput{}
    sendMessageInput.QueueUrl = aws.String(context.GetInput(ivQueueUrl).(string))
    sendMessageInput.MessageBody = aws.String(context.GetInput(ivMessageBody).(string))
    
    messageAttributes := context.GetInput(ivMessageAttributes).(*data.ComplexObject)
    if context.GetInput(ivMessageAttributes) != nil {
      //Add message attributes
      messageAttributes := context.GetInput(ivMessageAttributes).(*data.ComplexObject)
      attrs := make(map[string]*sqs.MessageAttributeValue, len(messageAttributes.Value))
      for k, v := range messageAttributes.Value { 
        attrs[k] = &sqs.MessageAttributeValue{
          DataType: aws.String(v["DataType"].(string)),
          StringValue:  aws.String(v["StringValue"].(string)),
        }
      }
      sendMessageInput.MessageAttributes = attrs
    }

    delaySeconds := context.GetInput(ivDelaySeconds)
    if delaySeconds != nil {
      sendMessageInput.DelaySeconds = aws.Int64(delaySeconds.(int64))
    }

    //Send message to SQS
    response, err1 := sqsSvc.SendMessage(sendMessageInput)
    if err1 != nil {
       return false, activity.NewError(fmt.Sprintf("Failed to send message to SQS due to error:%s",err1.Error()), "SQS-SENDMESSAGE-4005", nil)
    }

    //Set Message ID in the output
    context.SetOutput(ovMessageId,*response.MessageId)    
	return true, nil
}