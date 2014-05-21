package rlog

/*
This file implements the communication facilities between the message generation performed by the
goroutine accessing the logger API and the modules writing the log messages to various sources.
*/

import (
  "container/list"
  "github.com/brsc/rlog/common"
  "log"
  "time"
)

//msgChannels is a linked list of channels. The channels are used to send messages to the modules
var msgChannels *list.List = list.New()

//flushChannels is a linked list of channels. The channels are used to send the flush command to
//the modules
var flushChannels *list.List = list.New()

//getMsgChannel creates a log message channel and registers it.
//Returns: log message channel
func getMsgChannel() <-chan (*common.RlogMsg) {
  c := make(chan *common.RlogMsg, config.ChanCapacity)
  msgChannels.PushBack(c)
  return c
}

//getFlushChannel creates a flush command channel and registers it. A flush channel
//has capacity 1 so even if the flush receiver is currently busy handling a message,
//it gets the flush command. Termination is enforced by waiting only a limited amount
//of time for the module to respond with a success message to the flush.
//Returns: flush message channel
func getFlushChannel() chan (chan (bool)) {
  c := make(chan chan (bool), 1)
  flushChannels.PushBack(c)
  return c
}

//pushToChannels pushes a message to all registered channels.
//Arguments: message to push
func pushToChannels(msg *common.RlogMsg) {

  for e := msgChannels.Front(); e != nil; e = e.Next() {
    //Cycle over all registered channels, perform a type conversion (because of the linked
    //list) and call the helper function to push the log data without blocking
    c, ok := e.Value.(chan (*common.RlogMsg))
    if ok {
      pushToChannelsHelper(c, msg)
    } else {
      log.Panic("[RightLog4Go FATAL] type assertion for msg channel failed\n")
    }
  }
}

//pushToChannelsHelper pushes to a channel without blocking forever. If the channel is full, one element gets
//deleted and the message is pushed again (FIFO ringbuffer channel). The number of retries is limited to three
//to guarantee termination (deleting one element and writing the next element is not atomic).
//Arguments: [c] destination channel. [msg] Message to log
func pushToChannelsHelper(c chan (*common.RlogMsg), msg *common.RlogMsg) {

  success := false
  for retries := 0; retries < 3 && !success; retries++ {
    //Loop until either (a) success (b) #retries exceeded
    select {
    case c <- msg:
      //Send success
      success = true
    default:
      //Send failed, remove one item and retry
      // Do not log send failures using RightLog4Go because it would create a feedback loop
      log.Printf("[RightLog4Go] Log buffer full, delete and retry")
      nonBlockingChanRead(c)
    }
  }
}

//nonBlockingChanRead reads one item from the given channel. nonBlockingChanRead
//shall not block when the channel is empty
//Returns: Element read from channel, nil if channel empty
func nonBlockingChanRead(c <-chan (*common.RlogMsg)) *common.RlogMsg {
  select {
  case ret := <-c:
    return ret
  default:
    return nil
  }
}

//flushHelper sends the flush command and waits for a response from the module. The send channel has buffer
//capacity 1. If the buffer is empty, we place a return buffer in there to trigger the flush. If the buffer is
//full, there is already a pending flush command and we abort. After successfully triggering the flush command,
//we wait for a response or timeout. When timing out, there is no cleanup required as the return channel has
//buffer capacity 1 as well ==> the module can place it response into it without us receiving it. The channel
//will be garbage collected afterwards.
//Arguments: Channel to send flush command
//Returns: true on success, false otherwise
func flushHelper(c chan (chan (bool))) bool {
  responseChan := make(chan (bool), 1)
  select {
  //Phase 1: send flush command including a return channel to module
  case c <- responseChan:
    //Phase 2: wait for module to respond (or time out)
    select {
    case <-responseChan:
      //OK, we are done
      return true
    case <-time.After(time.Second * time.Duration(config.FlushTimeout)):
      log.Printf("[RightLog4Go] flush command ACK timed out\n")
      return false
    }
  default:
    //Flush channel full ==> pending flush?
    log.Printf("[RightLog4Go] Sending flush command to module failed, pending flush?\n")
    return false
  }
}
