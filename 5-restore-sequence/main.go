package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Message struct {
	str  string
	wait chan bool
}

func fanIn(inputs ...<-chan Message) <-chan Message {
	c := make(chan Message)
	for i := range inputs {
		input := inputs[i]
		go func() {
			for {
				c <- <-input
			}
		}()
	}
	return c
}

// the boring function return a channel to communicate with it.
func boring(msg string) <-chan Message { // <-chan Message means receives-only channel of Message.
	c := make(chan Message)
	waitForIt := make(chan bool) // share between all messages generated by this function
	go func() {                  // we launch goroutine inside a function.
		for i := 0; ; i++ {
			// Sends a message out,
			// wait for acknowledgement (via the waitForIt channel)
			// before being able to send another message
			c <- Message{
				str:  fmt.Sprintf("%s %d", msg, i),
				wait: waitForIt,
			}
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)

			// every time the goroutine send message.
			// This code waits until the value to be received.
			<-waitForIt
		}

	}()
	return c // return a channel to caller.
}

func main() {
	// merge 2 channels into 1 channel
	c := fanIn(boring("Joe"), boring("Ahn"))

	for i := 0; i < 5; i++ {
		msg1 := <-c // wait to receive message
		fmt.Println(msg1.str)
		msg2 := <-c // this is from another sender and cannot be from the same sender for `msg1` because we have not acknowledged `msg1` via the `wait` channel
		fmt.Println(msg2.str)

		// each go routine have to wait here before we acknowledge their messages
		msg1.wait <- true // main goroutine allows the boring goroutine to send next value to message channel.
		msg2.wait <- true
	}
	fmt.Println("You're both boring. I'm leaving")
}

// main: goroutine                                          boring: goroutine
//    |                                                           |
//    |                                                           |
// wait for receiving msg from channel c                    c <- Message{} // send message
//   <-c                                                          |
//    |                                                           |
//    |                                                     <-waitForIt // wait for wake up signal
// send value to channel                                          |
// hey, boring. You can send next value to me                     |
//   wait <-true                                                  |
///                            REPEAT THE PROCESS
