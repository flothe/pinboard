package grafic2d

import (
	"container/list"
	//	"fmt"
)

type Animation interface {
	Value(time int) (values []float32)
}

type AnimFrame struct {
	time   int
	values []float32
}

type AnimLinear struct {
	EnableLooping bool
	ValueMult     []float32
	frames        list.List
	maxTime       int
}

func (anim *AnimLinear) AddFrame(time int, values []float32) {
	frame := AnimFrame{time, values}
	var f AnimFrame
	isInserted := false

	for e := anim.frames.Front(); e != nil; e = e.Next() {
		f = e.Value.(AnimFrame)
		if time < f.time {
			anim.frames.InsertBefore(frame, e)
			isInserted = true
			break
		}
	}

	if !isInserted {
		anim.frames.PushBack(frame)
		anim.maxTime = frame.time
	}
}

func (anim *AnimLinear) Value(time int) []float32 {
	if anim.EnableLooping {
		time = time % anim.maxTime
	}
	f1, f2 := anim.getAdjacentFrames(time)

	if f1.time == f2.time {
		return f1.values
	}

	factor := float32(time-f1.time) / float32(f2.time-f1.time)
	n := len(f1.values)
	v := make([]float32, n, n)

	for i := 0; i < n; i++ {
		v[i] = f1.values[i] + factor*(f2.values[i]-f1.values[i])
		if i < len(anim.ValueMult) {
			//fmt.Printf("v[%v]=%v. -->",i,v[i])
			v[i] = v[i] * anim.ValueMult[i]
			//fmt.Printf("%v\n",v[i])
		}
	}
	return v
}

func (anim *AnimLinear) getAdjacentFrames(time int) (*AnimFrame, *AnimFrame) {

	var left, right AnimFrame

	// init the left frame
	e := anim.frames.Front()
	left = e.Value.(AnimFrame)

	for e := e.Next(); e != nil; e = e.Next() {
		right = e.Value.(AnimFrame)
		if time < right.time {
			return &left, &right
		}
		left = right
	}
	return &left, &left
}
