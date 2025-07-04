package main

import "github.com/golang-collections/collections/queue"

// foldersQ - folders queue
var foldersQ = queue.New()

type qitem struct {
	in, out string
}

func enq(in, out string) {
	foldersQ.Enqueue(&qitem{in: in, out: out})
}

func deq() (in, out string) {
	next := foldersQ.Dequeue().(*qitem)
	in, out = next.in, next.out
	return
}

func clearQ() {
	for foldersQ.Len() > 0 {
		foldersQ.Dequeue()
	}
}
