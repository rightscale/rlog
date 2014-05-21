/*
These tests cover:
- Channel creation
- Channel multipush: 1 message to multiple channels
- Channel FIFO behavior
- Non blocking channel read
*/
package rlog

import (
  "container/list"
  "github.com/brsc/rlog/common"
  . "launchpad.net/gocheck"
  "strconv"
  "time"
)

//When invoking nonBlockingChanRead, it should never block
func (s *Stateless) TestNonBlockingDelete(t *C) {
  //Create a channel and push 1 item into it
  logItem := &common.RlogMsg{"", "", SeverityError, 0, ""}
  c := make(chan (*common.RlogMsg), 2)
  c <- logItem

  //Channel contains 1 element. Delete 2 to ensure that this method never blocks
  res := nonBlockingChanRead(c)
  if res != logItem {
    t.Fatalf("Item read not equal item pushed")
  }

  res = nonBlockingChanRead(c)
  if res != nil {
    t.Fatalf("Read non nil element from channel, but channel should be empty")
  }
}

//When channel buffer capacity is exceeded, it should behave as a FIFO queue and never block
func (s *Stateless) TestPushToChannelHelper(t *C) {

  //Create message channel with capacity 2 and stuff 5 elements into it
  c := make(chan (*common.RlogMsg), 2)
  for i := 0; i < 5; i++ {
    pushToChannelsHelper(c, &common.RlogMsg{strconv.Itoa(i), "", SeverityError, uint(i), ""})
  }

  //Read back the elements, should receive the last two elements (FIFO)
  item := <-c
  if item.Pc != 3 {
    t.Fatalf("Incorrect FIFO behavior")
  }
  item = <-c
  if item.Pc != 4 {
    t.Fatalf("Incorrect FIFO behavior")
  }
}

//(1) When calling getMsgChannel, it should create a message channel and register it.
//(2) When pushing a message to a set of channels using pushToChannels, it should push
//exactly one message element to each channel.
func (s *Initialized) TestPushToChannels(t *C) {

  //Setup two of our channels for testing
  msgChannels = list.New()
  c1 := getMsgChannel()
  c2 := getMsgChannel()

  logItem := &common.RlogMsg{"", "", SeverityError, 0, ""}
  pushToChannels(logItem)

  //Read back items
  if testPushToChannelsHelper(t, c1) != logItem {
    t.Fatalf("Element in channel is not expected element")
  }
  if testPushToChannelsHelper(t, c2) != logItem {
    t.Fatalf("Element in channel is not expected element")
  }
}

//testPushToChannelsHelper is a helper function for TestPushToChannels that reads two elements from
//the given channel: the first element is expected to be there whereas the second element should be
//not be available. However, the read shall never block.
func testPushToChannelsHelper(t *C, c <-chan (*common.RlogMsg)) *common.RlogMsg {

  //When reading back from the channel, it should return one element
  ret := nonBlockingChanRead(c)
  if ret == nil {
    t.Fatalf("No element in channel, but their should be an element in the channel")
  }

  //... and only one element
  if nonBlockingChanRead(c) != nil {
    t.Fatalf("Read non nil element from channel, but channel should be empty")
  }

  return ret
}

//When invoking the flush command, it should notify all subscribers
func (s *Initialized) TestFlush(t *C) {

  //Setup return channel to check whether the flush worked
  confirm := make(chan (bool), 2)

  //Register two flush listeners
  c1 := getFlushChannel()
  c2 := getFlushChannel()

  //Spawn two goroutines simulating the modules
  simulateModuleAndConfirm(c1, confirm)
  simulateModuleAndConfirm(c2, confirm)

  //Send flush command to each of them
  Flush()

  //Verify we got 1 element
  select {
  case <-confirm:
  case <-time.After(time.Second):
    t.Fatalf("Goroutine did not send confirmation after flush command")
  }

  //...and another one
  select {
  case <-confirm:
  case <-time.After(time.Second):
    t.Fatalf("Goroutine did not send confirmation after flush command")
  }
}

//simulateModuleAndConfirm implements the protocol which is expected to run inside a logger module and
//has in addition a confirmation channel to verify that indeed the flush command has been sent
func simulateModuleAndConfirm(c chan (chan (bool)), confirm chan (bool)) {
  go func(ch chan (chan (bool))) {
    //Block on c until we get something and send response immediately
    ret := <-ch
    ret <- true

    //Put true in confirm channel to notify that we were invoked
    confirm <- true
  }(c)
}

//Test flush helper command algorithm. Run initialized because we depend on the flush timeout.
func (s *Initialized) TestFlushHelper(t *C) {

  //A flush channel and a variable to capture the return value
  var c chan (chan (bool))
  var ret bool

  //Disable flush timeout to speed-up the test case with no receiver
  config.FlushTimeout = 0

  //When sending a flush command with no receiver (e.g module crashed), it should fail but not block forever
  //This includes the following test case: When sending a flush command to a goroutine which receives the
  //command but never responds, it should fail but not block forever
  c = getFlushChannel()
  ret = flushHelper(c)
  if ret {
    t.Fatalf("Flush helper succeeded although there was no receiver")
  }

  //Set flush timeout again to give the goroutine time to respond success to flush command
  config.FlushTimeout = 2

  //When sending a flush command to a correctly behaving goroutine, it should succeed
  c = getFlushChannel()
  go func(ch chan (chan (bool))) {
    //Block on c until we get something and send response immediately
    ret := <-ch
    ret <- true
  }(c)
  ret = flushHelper(c)
  if !ret {
    t.Fatalf("Flush helper did not succeed although it should have")
  }
}
