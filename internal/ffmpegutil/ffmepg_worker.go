package ffmpegutil

import "log"

type HlsSegmentsParams struct {
	InputFile       string
	StartIndex      int
	TotalSegments   int
	SegmentDuration int
}

type Worker struct {
	CurrentEpisodeID int
	messageChan      chan HlsSegmentsParams
	WorkMap          map[int]bool
}

func NewWorker() *Worker {
	workMap := make(map[int]bool)
	return &Worker{
		messageChan:      make(chan HlsSegmentsParams),
		WorkMap:          workMap,
		CurrentEpisodeID: 0,
	}
}

func (w *Worker) Process(message HlsSegmentsParams) {
	log.Println("message1: ", message)
	if _, ok := w.WorkMap[message.StartIndex]; !ok {
		w.WorkMap[message.StartIndex] = true
		w.messageChan <- message
	}
}
