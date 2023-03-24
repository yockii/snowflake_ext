package snowflake

import (
	"errors"
	"sync"
	"time"
)

// 时间序列 41bit，69年；工作机器5bit，31台集群；每毫秒序列14bit，16383个数字
const (
	// defaultWorkerIdBits 工作机器ID，5bit = 31
	defaultWorkerIdBits = uint64(5)
	// defaultSequenceBits 序列，14bit = 16384
	defaultSequenceBits = uint64(14)

	defaultMaxWorkerId = uint64(int64(-1) ^ (int64(-1) << defaultWorkerIdBits)) // 工作机器id最大值，防止溢出
	defaultMaxSequence = uint64(int64(-1) ^ (int64(-1) << defaultSequenceBits)) // 最大序列号，防止溢出

	// defaultBaseEpoch 基准时间，2023-01-01
	defaultBaseEpoch = uint64(1672531200000)
)

type WorkerOption struct {
	BaseEpoch    uint64 // 基准时间戳，默认1672531200000， 2023-01-01
	WorkerIdBits uint64 // 工人id占位数，默认5，31个
	SequenceBits uint64 // 序列占位数，默认14，16383个数字

	// 基准信息处理
	LastStamp uint64 // 上次的时间戳
	Sequence  uint64 // 序列
}

type Worker struct {
	mu        sync.Mutex
	lastStamp uint64 // 上次时间戳
	workerId  uint64 // 工作机器id
	sequence  uint64 // 当前毫秒序列

	// worker限制量
	baseEpoch         uint64 // 基准时间
	maxWorkerId       uint64
	maxSequence       uint64
	timeLeftShift     uint64
	workerIdLeftShift uint64
}

func NewSnowflakeWithConfig(workerId uint64, option ...WorkerOption) (*Worker, error) {
	baseEpoch := defaultBaseEpoch
	maxWorkerId := defaultMaxWorkerId
	maxSequence := defaultMaxSequence
	lastStamp := uint64(0)
	sequence := uint64(0)
	var o = WorkerOption{
		BaseEpoch:    defaultBaseEpoch,
		WorkerIdBits: defaultWorkerIdBits,
		SequenceBits: defaultSequenceBits,
	}
	if len(option) > 0 {
		o = option[0]
		if o.BaseEpoch > 0 {
			baseEpoch = o.BaseEpoch
		}
		if o.WorkerIdBits > 0 {
			maxWorkerId = uint64(int64(-1) ^ (int64(-1) << o.WorkerIdBits))
		}
		if o.SequenceBits > 0 {
			maxSequence = uint64(int64(-1) ^ (int64(-1) << o.SequenceBits))
		}
		lastStamp = o.LastStamp
		sequence = o.Sequence
	}
	if workerId >= maxWorkerId {
		return nil, errors.New("workerId is too big")
	}
	w := &Worker{
		workerId:          workerId,
		baseEpoch:         baseEpoch,
		maxWorkerId:       maxWorkerId,
		maxSequence:       maxSequence,
		timeLeftShift:     o.WorkerIdBits + o.SequenceBits,
		workerIdLeftShift: o.SequenceBits,
		lastStamp:         lastStamp,
		sequence:          sequence,
	}
	return w, nil
}

func NewSnowflake(workerId uint64) (*Worker, error) {
	if workerId >= defaultMaxWorkerId {
		return nil, errors.New("workerId is too big")
	}
	return NewSnowflakeWithConfig(workerId)
}

func (w *Worker) LastStamp() uint64 {
	return w.lastStamp
}
func (w *Worker) Sequence() uint64 {
	return w.sequence
}

func (w *Worker) getMilliSeconds() uint64 {
	return uint64(time.Now().UnixNano() / 1e6)
}

func (w *Worker) NextId() uint64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.nextId()
}

func (w *Worker) nextId() uint64 {
	timestamp := w.getMilliSeconds()
	if timestamp < w.lastStamp {
		//return 0, errors.New("time is moving backwards, waiting until")
		// 处理时间回拨，思路，直接将时间定位到当前时间，并进行下一个序列号给出，直到序列号上限后，进入下一毫秒
		w.sequence = (w.sequence + 1) & w.maxSequence
		if w.sequence == 0 {
			w.lastStamp++ // 进入下一毫秒
		}
		return w.generateId()
	}
	if w.lastStamp == timestamp {
		w.sequence = (w.sequence + 1) & w.maxSequence
		if w.sequence == 0 {
			w.lastStamp++ // 进入下一毫秒
		}
	} else {
		w.sequence = 0
		w.lastStamp = timestamp
	}
	return w.generateId()
}

func (w *Worker) generateId() uint64 {
	id := ((w.lastStamp - w.baseEpoch) << w.timeLeftShift) | (w.workerId << w.workerIdLeftShift) | w.sequence
	return uint64(id)
}
