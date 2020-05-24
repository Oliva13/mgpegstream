package main

import "C"
import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
	//"unsafe"
)

var boundary = flag.String("boundary", "--BOUNDARY", "boundary marker")

var buffersize = flag.Int("buffersize", 4096, "buffer size")

var data = &threadSafeSlice{
	workers: make([]*worker, 0, 1),
}

type worker struct {
	source chan []byte
	first  bool
	done   bool
}

type threadSafeSlice struct {
	sync.Mutex
	workers []*worker
}

func (s *threadSafeSlice) Len() int {
	s.Lock()
	defer s.Unlock()

	return len(s.workers)
}

func (s *threadSafeSlice) Push(w *worker) {

	s.Lock()
	defer s.Unlock()
	s.workers = append(s.workers, w)
}

func (s *threadSafeSlice) Iter(routine func(*worker) bool) {
	s.Lock()
	defer s.Unlock()

	for i := len(s.workers) - 1; i >= 0; i-- {
		remove := routine(s.workers[i])
		if remove {
			s.workers[i] = nil
			s.workers = append(s.workers[:i], s.workers[i+1:]...)
		}
	}
}

func broadcaster(ch chan []byte) {
	for {
		msg, ok := <-ch
		data.Iter(func(w *worker) bool {
			if w.done || !ok {
				fmt.Fprintf(os.Stderr, "done %p\n", w)
				close(w.source)
				return true
			} else {
				w.source <- msg
				return false
			}
		})

		if !ok {
			break
		}
	}
}

func generator(ch chan []byte) {

	maxSize := 128000
	buffer := make([]byte, maxSize)
	numberStream := 3

	for {
		if Len() == 0 {
			time.Sleep(17 * time.Millisecond)
			continue
		}
		dataSize, err := readFrame(buffer, numberStream)
		if err != nil {
			continue
		}
		sendFrame(buffer[:dataSize], ch)
	}
	close(ch)
}

func readFrame(buf []byte, num int) (int, error) {

	//ptr := (*C.uchar)(unsafe.Pointer(&buf[0]))
	//maxSize := C.uint(len(buf))
	//stream := C.int(num)


	dataSize := 0
	//dataSize := int(C.get_mjpeg(stream, ptr, maxSize))

	if dataSize == 0 || dataSize == -61 {
		return 0, errors.New("no frame")
	}

	return dataSize, nil
}

func sendFrame(frame []byte, ch chan []byte) {
	buffer := new(bytes.Buffer)
	fmt.Fprintf(buffer, "%s\r\n", *boundary)
	fmt.Fprintf(buffer, "Content-Type: image/jpeg\r\n")
	fmt.Fprintf(buffer, "Content-Length: %d\r\n", len(frame))
	buffer.Write([]byte("\r\n"))
	buffer.Write(frame)
	cp := make([]byte, buffer.Len())
	copy(cp, buffer.Bytes())
	ch <- cp
	buffer.Reset()
}

func Broadcast() {
	c := make(chan []byte)
	go broadcaster(c)
	go generator(c)
}

func Len() int {
	return data.Len()
}

func StreamTo(w io.Writer, closed <-chan bool) {

	wk := &worker{
		source: make(chan []byte),
		first:  true,
	}
	//fmt.Fprintf(os.Stderr, "created %p\n", wk)
	data.Push(wk)
loop:
	for {
		select {
		case s, ok := <-wk.source:
			if !ok {
				//fmt.Println("============= break loop")
				break loop
			}
			if !wk.first {
				//fmt.Println("====== wk.first == false")
				w.Write([]byte("\r\n"))
			} else {
				//fmt.Println("wk.first == true => wk.first = false")
				wk.first = false
			}
			w.Write(s)
		case <-closed:
			wk.done = true
		}
	}
}

func WriteStreamOutput(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "multipart/x-mixed-replace;boundary=--BOUNDARY")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if c, ok := w.(http.CloseNotifier); ok {
		StreamTo(w, c.CloseNotify())
	}
}
